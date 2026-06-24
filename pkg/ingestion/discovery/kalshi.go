package discovery

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type KalshiClient struct {
	baseURL    string
	keyID      string
	keyPEM     string
	httpClient *http.Client
}

func NewKalshiClient(keyID, keyPEM string) *KalshiClient {
	return &KalshiClient{
		baseURL:    "https://api.elections.kalshi.com/trade-api/v2",
		keyID:      keyID,
		keyPEM:     keyPEM,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *KalshiClient) Venue() string { return "KALSHI" }

func (c *KalshiClient) FetchMarkets(ctx context.Context) ([]Market, error) {
	var all []Market
	cursor := ""

	for {
		url := fmt.Sprintf("%s/markets?limit=1000&status=open", c.baseURL)
		if cursor != "" {
			url += "&cursor=" + cursor
		}

		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return nil, fmt.Errorf("kalshi request: %w", err)
		}
		c.signRequest(req)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("kalshi do: %w", err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("kalshi read: %w", err)
		}

		var result struct {
			Markets []kalshiMarket `json:"markets"`
			Cursor  string         `json:"cursor"`
		}
		if err := json.Unmarshal(body, &result); err != nil {
			return nil, fmt.Errorf("kalshi decode: %w", err)
		}

		for _, km := range result.Markets {
			all = append(all, c.normalize(km))
		}

		if result.Cursor == "" || len(result.Markets) == 0 {
			break
		}
		cursor = result.Cursor
	}

	return all, nil
}

type kalshiMarket struct {
	Ticker    string  `json:"ticker"`
	Title     string  `json:"title"`
	Sector    string  `json:"sector"`
	Series    string  `json:"series"`
	OpenTime  string  `json:"open_time"`
	CloseTime string  `json:"close_time"`
	YesBid    float64 `json:"yes_bid"`
	YesAsk    float64 `json:"yes_ask"`
	NoBid     float64 `json:"no_bid"`
	NoAsk     float64 `json:"no_ask"`
}

func (c *KalshiClient) normalize(km kalshiMarket) Market {
	var openTime, closeTime time.Time
	if km.OpenTime != "" {
		openTime, _ = time.Parse(time.RFC3339, km.OpenTime)
	}
	if km.CloseTime != "" {
		closeTime, _ = time.Parse(time.RFC3339, km.CloseTime)
	}

	tickerParts := strings.Split(km.Ticker, "-")
	series := ""
	if len(tickerParts) > 0 {
		series = strings.ToLower(tickerParts[0])
	}

	return Market{
		Venue:    "KALSHI",
		MarketID: km.Ticker,
		Ticker:   km.Ticker,
		Title:    km.Title,
		Series:   series,
		Category: km.Sector,
		Outcomes: []Outcome{
			{Name: "Yes", Price: km.YesBid},
			{Name: "No", Price: km.NoBid},
		},
		OpenTime:  openTime,
		CloseTime: closeTime,
	}
}

func (c *KalshiClient) signRequest(req *http.Request) {
	if c.keyID != "" {
		req.Header.Set("Kalshi-Key-Id", c.keyID)
	}
}
