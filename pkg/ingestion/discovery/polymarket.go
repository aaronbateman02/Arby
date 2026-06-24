package discovery

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type PolymarketClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewPolymarketClient() *PolymarketClient {
	return &PolymarketClient{
		baseURL:    "https://clob.polymarket.com",
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *PolymarketClient) Venue() string { return "POLYMARKET" }

func (c *PolymarketClient) FetchMarkets(ctx context.Context) ([]Market, error) {
	url := fmt.Sprintf("%s/markets?limit=1000&closed=false", c.baseURL)

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

	var raw []polymarketMarket
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("polymarket decode: %w", err)
	}

	var result []Market
	for _, pm := range raw {
		result = append(result, c.normalize(pm))
	}

	return result, nil
}

type polymarketMarket struct {
	ConditionID string    `json:"condition_id"`
	Question    string    `json:"question"`
	Slug        string    `json:"slug"`
	Category    string    `json:"category"`
	EndDate     string    `json:"end_date"`
	Outcomes    []string  `json:"outcomes"`
	Prices      []float64 `json:"prices"`
}

func (c *PolymarketClient) normalize(pm polymarketMarket) Market {
	var closeTime time.Time
	if pm.EndDate != "" {
		closeTime, _ = time.Parse(time.RFC3339, pm.EndDate)
	}

	ticker := pm.Slug
	if ticker == "" {
		ticker = pm.ConditionID
	}

	outcomes := make([]Outcome, len(pm.Outcomes))
	for i, name := range pm.Outcomes {
		price := 0.0
		if i < len(pm.Prices) {
			price = pm.Prices[i]
		}
		outcomes[i] = Outcome{Name: name, Price: price}
	}

	return Market{
		Venue:     "POLYMARKET",
		MarketID:  pm.ConditionID,
		Ticker:    ticker,
		Title:     pm.Question,
		Category:  pm.Category,
		Outcomes:  outcomes,
		CloseTime: closeTime,
	}
}
