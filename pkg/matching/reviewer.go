package matching

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

type Reviewer struct {
	store               *Store
	interval            time.Duration
	openRouterKey       string
	openRouterBaseURL   string
	legAModel           string
	legBModel           string
	batchSize           int
	confidenceThreshold float64
}

type ReviewResult struct {
	PairIndex    int     `json:"pair_index"`
	IsSameEvent  bool    `json:"is_same_event"`
	Relationship string  `json:"relationship"`
	Confidence   float64 `json:"confidence"`
	Reasoning    string  `json:"reasoning"`
}

type openRouterMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openRouterRequest struct {
	Model          string              `json:"model"`
	Messages       []openRouterMessage `json:"messages"`
	ResponseFormat *responseFormat     `json:"response_format,omitempty"`
}

type responseFormat struct {
	Type       string          `json:"type"`
	JSONSchema *jsonSchemaSpec `json:"json_schema,omitempty"`
}

type jsonSchemaSpec struct {
	Name   string         `json:"name"`
	Schema map[string]any `json:"schema"`
}

type openRouterChoice struct {
	Message openRouterMessage `json:"message"`
}

type openRouterResponse struct {
	Choices []openRouterChoice `json:"choices"`
}

type batchReviewResponse struct {
	Reviews []ReviewResult `json:"reviews"`
}

func NewReviewer(store *Store, interval time.Duration, openRouterKey, openRouterBaseURL, legAModel, legBModel string, batchSize int, confidenceThreshold float64) *Reviewer {
	return &Reviewer{
		store:               store,
		interval:            interval,
		openRouterKey:       openRouterKey,
		openRouterBaseURL:   openRouterBaseURL,
		legAModel:           legAModel,
		legBModel:           legBModel,
		batchSize:           batchSize,
		confidenceThreshold: confidenceThreshold,
	}
}

func (r *Reviewer) Run(ctx context.Context) {
	slog.Info("reviewer started", "interval", r.interval, "batch_size", r.batchSize)
	if err := r.RunOnce(ctx); err != nil {
		slog.Error("reviewer run once", "error", err)
	}
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if err := r.RunOnce(ctx); err != nil {
				slog.Error("reviewer", "error", err)
			}
		case <-ctx.Done():
			slog.Info("reviewer stopped")
			return
		}
	}
}

func (r *Reviewer) RunOnce(ctx context.Context) error {
	candidates, err := r.store.GetPendingCandidatesWithMarkets(ctx, r.batchSize*4)
	if err != nil {
		return fmt.Errorf("get pending candidates: %w", err)
	}
	if len(candidates) == 0 {
		slog.Info("no pending candidates to review")
		return nil
	}

	slog.Info("reviewing candidates", "count", len(candidates))

	var batches [][]CandidateWithMarkets
	for i := 0; i < len(candidates); i += r.batchSize {
		end := i + r.batchSize
		if end > len(candidates) {
			end = len(candidates)
		}
		batches = append(batches, candidates[i:end])
	}

	totalReviewed := 0
	totalConfirmed := 0

	for _, batch := range batches {
		results, err := r.reviewBatch(ctx, batch)
		if err != nil {
			slog.Error("batch review failed", "error", err, "batch_size", len(batch))
			continue
		}

		for _, result := range results {
			if result.PairIndex < 0 || result.PairIndex >= len(batch) {
				slog.Warn("review result pair_index out of range", "pair_index", result.PairIndex)
				continue
			}
			c := batch[result.PairIndex]

			legBModel := ""
			if result.IsSameEvent {
				legBResult, err := r.secondOpinion(ctx, c, result)
				if err != nil {
					slog.Error("second opinion failed", "error", err, "candidate_id", c.ID)
				} else if legBResult != nil {
					slog.Info("second opinion received",
						"candidate_id", c.ID,
						"is_same_event", legBResult.IsSameEvent,
						"relationship", legBResult.Relationship,
						"confidence", legBResult.Confidence,
					)
				}
				legBModel = r.legBModel
			}

			mp := MatchPair{
				CandidateID:  c.ID,
				IsSameEvent:  result.IsSameEvent,
				Relationship: result.Relationship,
				Confidence:   result.Confidence,
				Reasoning:    result.Reasoning,
				LegAModel:    r.legAModel,
				LegBModel:    legBModel,
				Status:       "PENDING_APPROVAL",
			}
			if err := r.store.InsertMatchPair(ctx, mp); err != nil {
				slog.Error("failed to insert match pair", "error", err, "candidate_id", c.ID)
				continue
			}

			if err := r.store.UpdateCandidateStatus(ctx, c.ID, "REVIEWED"); err != nil {
				slog.Error("failed to update candidate status", "error", err, "candidate_id", c.ID)
			}

			totalReviewed++
			if result.IsSameEvent {
				totalConfirmed++
			}
		}
	}

	slog.Info("review pass complete", "reviewed", totalReviewed, "confirmed", totalConfirmed)
	return nil
}

func (r *Reviewer) reviewBatch(ctx context.Context, batch []CandidateWithMarkets) ([]ReviewResult, error) {
	if len(batch) == 0 {
		return nil, fmt.Errorf("empty batch")
	}

	systemPrompt := `You are a financial market matching expert. Your task is to determine whether two prediction market contracts resolve based on the same real-world event.

For each pair of markets, analyze:
1. Do both markets reference the same underlying event or outcome?
2. What is the relationship between them (EQUIVALENT, INVERSE, or UNRELATED)?
3. How confident are you in your assessment?

EQUIVALENT means both resolve to YES/NO based on the same condition.
INVERSE means one market resolves YES when the other resolves NO.
UNRELATED means they reference different events or conditions.`

	var userPrompt string
	for i, c := range batch {
		descA := truncateStr(c.MarketA.Description, 200)
		descB := truncateStr(c.MarketB.Description, 200)

		userPrompt += fmt.Sprintf(`Pair %d:
Market A (%s): "%s"
Description: %s

Market B (%s): "%s"
Description: %s

Similarity: %.3f | Category: %s
---
`, i, c.MarketA.Venue, c.MarketA.Title, descA,
			c.MarketB.Venue, c.MarketB.Title, descB,
			c.Similarity, c.Category)
	}

	userPrompt += fmt.Sprintf("\nProvide your analysis for all %d pairs as a JSON object with a \"reviews\" array. Each review must include: pair_index, is_same_event, relationship, confidence (0-1), reasoning.", len(batch))

	body := openRouterRequest{
		Model: r.legAModel,
		Messages: []openRouterMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		ResponseFormat: &responseFormat{
			Type: "json_schema",
			JSONSchema: &jsonSchemaSpec{
				Name: "BatchReview",
				Schema: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"reviews": map[string]any{
							"type": "array",
							"items": map[string]any{
								"type": "object",
								"properties": map[string]any{
									"pair_index":    map[string]any{"type": "integer"},
									"is_same_event": map[string]any{"type": "boolean"},
									"relationship":  map[string]any{"type": "string", "enum": []string{"EQUIVALENT", "INVERSE", "UNRELATED"}},
									"confidence":    map[string]any{"type": "number", "minimum": 0, "maximum": 1},
									"reasoning":     map[string]any{"type": "string"},
									"potential_ambiguity": map[string]any{"type": "string"},
								},
								"required": []string{"pair_index", "is_same_event", "relationship", "confidence", "reasoning"},
							},
						},
					},
					"required": []string{"reviews"},
				},
			},
		},
	}

	reqBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, r.openRouterBaseURL+"/chat/completions", bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+r.openRouterKey)
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 60 * time.Second}
	httpResp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer httpResp.Body.Close()

	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("openrouter api error: status=%d body=%s", httpResp.StatusCode, string(respBody))
	}

	var orResp openRouterResponse
	if err := json.Unmarshal(respBody, &orResp); err != nil {
		return nil, fmt.Errorf("unmarshal openrouter response: %w", err)
	}

	if len(orResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in openrouter response")
	}

	content := orResp.Choices[0].Message.Content

	var batchResp batchReviewResponse
	if err := json.Unmarshal([]byte(content), &batchResp); err != nil {
		return nil, fmt.Errorf("unmarshal batch review content: %w", err)
	}

	return batchResp.Reviews, nil
}

func (r *Reviewer) secondOpinion(ctx context.Context, c CandidateWithMarkets, legAResult ReviewResult) (*ReviewResult, error) {
	systemPrompt := `You are a financial market matching expert. Your task is to determine whether two prediction market contracts resolve based on the same real-world event.

For each pair of markets, analyze:
1. Do both markets reference the same underlying event or outcome?
2. What is the relationship between them (EQUIVALENT, INVERSE, or UNRELATED)?
3. How confident are you in your assessment?

EQUIVALENT means both resolve to YES/NO based on the same condition.
INVERSE means one market resolves YES when the other resolves NO.
UNRELATED means they reference different events or conditions.`

	descA := truncateStr(c.MarketA.Description, 200)
	descB := truncateStr(c.MarketB.Description, 200)

	userPrompt := fmt.Sprintf(`Pair 0:
Market A (%s): "%s"
Description: %s

Market B (%s): "%s"
Description: %s

Similarity: %.3f | Category: %s
---
Provide your analysis for 1 pair as a JSON object with a "reviews" array. Each review must include: pair_index, is_same_event, relationship, confidence (0-1), reasoning.`,
		c.MarketA.Venue, c.MarketA.Title, descA,
		c.MarketB.Venue, c.MarketB.Title, descB,
		c.Similarity, c.Category)

	body := openRouterRequest{
		Model: r.legBModel,
		Messages: []openRouterMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		ResponseFormat: &responseFormat{
			Type: "json_schema",
			JSONSchema: &jsonSchemaSpec{
				Name: "BatchReview",
				Schema: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"reviews": map[string]any{
							"type": "array",
							"items": map[string]any{
								"type": "object",
								"properties": map[string]any{
									"pair_index":    map[string]any{"type": "integer"},
									"is_same_event": map[string]any{"type": "boolean"},
									"relationship":  map[string]any{"type": "string", "enum": []string{"EQUIVALENT", "INVERSE", "UNRELATED"}},
									"confidence":    map[string]any{"type": "number", "minimum": 0, "maximum": 1},
									"reasoning":     map[string]any{"type": "string"},
								},
								"required": []string{"pair_index", "is_same_event", "relationship", "confidence", "reasoning"},
							},
						},
					},
					"required": []string{"reviews"},
				},
			},
		},
	}

	reqBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, r.openRouterBaseURL+"/chat/completions", bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+r.openRouterKey)
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 60 * time.Second}
	httpResp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer httpResp.Body.Close()

	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("openrouter api error: status=%d body=%s", httpResp.StatusCode, string(respBody))
	}

	var orResp openRouterResponse
	if err := json.Unmarshal(respBody, &orResp); err != nil {
		return nil, fmt.Errorf("unmarshal openrouter response: %w", err)
	}

	if len(orResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in openrouter response")
	}

	content := orResp.Choices[0].Message.Content

	var batchResp batchReviewResponse
	if err := json.Unmarshal([]byte(content), &batchResp); err != nil {
		return nil, fmt.Errorf("unmarshal batch review content: %w", err)
	}

	if len(batchResp.Reviews) == 0 {
		return nil, fmt.Errorf("no reviews in second opinion response")
	}

	return &batchResp.Reviews[0], nil
}

func truncateStr(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}
