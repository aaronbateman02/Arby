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
	url := fmt.Sprintf("%s/markets?limit=1000", c.baseURL)

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

	var wrapper struct {
		Data []polymarketMarket `json:"data"`
	}
	if err := json.Unmarshal(body, &wrapper); err != nil {
		return nil, fmt.Errorf("polymarket decode: %w", err)
	}

	var result []Market
	for _, pm := range wrapper.Data {
		result = append(result, c.normalize(pm))
	}

	return result, nil
}

type polymarketToken struct {
	TokenID string  `json:"token_id"`
	Outcome string  `json:"outcome"`
	Price   float64 `json:"price"`
}

type polymarketMarket struct {
	ConditionID string           `json:"condition_id"`
	Question    string           `json:"question"`
	MarketSlug  string           `json:"market_slug"`
	EndDateISO  string           `json:"end_date_iso"`
	Tokens      []polymarketToken `json:"tokens"`
}

func (c *PolymarketClient) normalize(pm polymarketMarket) Market {
	var closeTime time.Time
	if pm.EndDateISO != "" {
		closeTime, _ = time.Parse(time.RFC3339, pm.EndDateISO)
	}

	ticker := pm.MarketSlug
	if ticker == "" {
		ticker = pm.ConditionID
	}

	outcomes := make([]Outcome, len(pm.Tokens))
	for i, t := range pm.Tokens {
		outcomes[i] = Outcome{Name: t.Outcome, Price: t.Price}
	}

	return Market{
		Venue:     "POLYMARKET",
		MarketID:  pm.ConditionID,
		Ticker:    ticker,
		Title:     pm.Question,
		Outcomes:  outcomes,
		CloseTime: closeTime,
	}
}
