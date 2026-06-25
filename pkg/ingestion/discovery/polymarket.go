package discovery

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

type PolymarketClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewPolymarketClient() *PolymarketClient {
	return &PolymarketClient{
		baseURL:    "https://gamma-api.polymarket.com",
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *PolymarketClient) Venue() string { return "POLYMARKET" }

func (c *PolymarketClient) FetchMarkets(ctx context.Context) ([]Market, error) {
	slog.Info("polymarket fetch starting")
	start := time.Now()

	events, err := c.fetchEvents(ctx)
	if err != nil {
		return nil, err
	}

	var all []Market
	for _, e := range events {
		for _, m := range e.Markets {
			if !m.Active {
				continue
			}
			all = append(all, c.normalize(e, m))
		}
	}

	slog.Info("polymarket fetch complete", "events", len(events), "markets", len(all), "elapsed", time.Since(start).String())
	return all, nil
}

type pmEvent struct {
	ID          string        `json:"id"`
	Slug        string        `json:"slug"`
	Title       string        `json:"title"`
	Description string        `json:"description"`
	StartDate   string        `json:"startDate"`
	EndDate     string        `json:"endDate"`
	Active      bool          `json:"active"`
	Closed      bool          `json:"closed"`
	Markets     []pmMarket    `json:"markets"`
	Tags        []pmTag       `json:"tags"`
}

type pmTag struct {
	ID    json.Number `json:"id"`
	Label string      `json:"label"`
	Slug  string      `json:"slug"`
}

type pmMarket struct {
	ID          string `json:"id"`
	Question    string `json:"question"`
	ConditionID string `json:"conditionId"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	Outcomes    string `json:"outcomes"`
	EndDateISO  string `json:"endDateIso"`
	Active      bool   `json:"active"`
	Closed      bool   `json:"closed"`
}

func (c *PolymarketClient) fetchEvents(ctx context.Context) ([]pmEvent, error) {
	var all []pmEvent
	offset := 0
	limit := 500

	for {
		url := fmt.Sprintf("%s/events?active=true&closed=false&limit=%d&offset=%d&order=id&ascending=true",
			c.baseURL, limit, offset)

		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return nil, fmt.Errorf("polymarket request: %w", err)
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("polymarket do: %w", err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("polymarket read: %w", err)
		}

		var events []pmEvent
		if err := json.Unmarshal(body, &events); err != nil {
			return nil, fmt.Errorf("polymarket decode: %w", err)
		}

		all = append(all, events...)

		if len(events) < limit {
			break
		}
		offset += limit
	}

	return all, nil
}

func (c *PolymarketClient) normalize(e pmEvent, m pmMarket) Market {
	var closeTime time.Time
	if m.EndDateISO != "" {
		closeTime, _ = time.Parse("2006-01-02", m.EndDateISO)
	}

	ticker := m.Slug
	if ticker == "" {
		ticker = m.ConditionID
	}

	var outcomes []Outcome
	var outcomeNames []string
	if m.Outcomes != "" {
		if err := json.Unmarshal([]byte(m.Outcomes), &outcomeNames); err == nil {
			for _, name := range outcomeNames {
				outcomes = append(outcomes, Outcome{Name: name})
			}
		}
	}
	if len(outcomes) == 0 {
		outcomes = []Outcome{{Name: "Yes"}, {Name: "No"}}
	}

	return Market{
		Venue:        "POLYMARKET",
		MarketID:     m.ConditionID,
		Ticker:       ticker,
		Title:        m.Question,
		Description:  m.Description,
		VenueEventID: e.ID,
		EventTitle:   e.Title,
		Outcomes:     outcomes,
		CloseTime:    closeTime,
	}
}
