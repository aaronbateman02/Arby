package discovery

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync"
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

type kalshiMarket struct {
	Ticker          string    `json:"ticker"`
	Title           string    `json:"title"`
	Sector          string    `json:"sector"`
	Series          string    `json:"series"`
	OpenTime        string    `json:"open_time"`
	CloseTime       string    `json:"close_time"`
	YesBid          float64   `json:"yes_bid"`
	YesAsk          float64   `json:"yes_ask"`
	NoBid           float64   `json:"no_bid"`
	NoAsk           float64   `json:"no_ask"`
	EventTicker     string    `json:"event_ticker"`
	MveSelectedLegs []mveLeg  `json:"mve_selected_legs"`
	MarketType      string    `json:"market_type"`
	RulesPrimary    string    `json:"rules_primary"`
	YesSubTitle     string    `json:"yes_sub_title"`
	NoSubTitle      string    `json:"no_sub_title"`
}

type kalshiEvent struct {
	EventTicker string `json:"event_ticker"`
	Title       string `json:"title"`
	SubTitle    string `json:"sub_title"`
	Category    string `json:"category"`
	SeriesTicker string `json:"series_ticker"`
}

type eventResponse struct {
	Event   kalshiEvent    `json:"event"`
	Markets []kalshiMarket `json:"markets"`
}

type mveLeg struct {
	EventTicker  string `json:"event_ticker"`
	MarketTicker string `json:"market_ticker"`
	Side         string `json:"side"`
}

func (c *KalshiClient) normalize(km kalshiMarket) Market {
	return c.normalizeWithEvent(km, "")
}

func (c *KalshiClient) normalizeWithEvent(km kalshiMarket, eventTitle string) Market {
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

	// Only set event fields when this is a real event market (not a bundle)
	venueEventID := ""
	if eventTitle != "" {
		venueEventID = km.EventTicker
	}

	return Market{
		Venue:        "KALSHI",
		MarketID:     km.Ticker,
		Ticker:       km.Ticker,
		Title:        km.Title,
		Description:  strings.TrimSpace(km.RulesPrimary),
		Series:       series,
		Category:     km.Sector,
		VenueEventID: venueEventID,
		EventTitle:   eventTitle,
		Outcomes: []Outcome{
			{Name: "Yes", Price: km.YesBid},
			{Name: "No", Price: km.NoBid},
		},
		OpenTime:  openTime,
		CloseTime: closeTime,
	}
}

type marketsResponse struct {
	Markets []kalshiMarket `json:"markets"`
	Cursor  string         `json:"cursor"`
}

func (c *KalshiClient) fetchPage(ctx context.Context, url string) (*marketsResponse, error) {
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

	var result marketsResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("kalshi decode: %w", err)
	}
	return &result, nil
}

func (c *KalshiClient) FetchMarkets(ctx context.Context) ([]Market, error) {
	slog.Info("kalshi fetch starting")
	start := time.Now()

	// Phase 1: fetch bundles (for price data / reference)
	bundles, err := c.fetchBundles(ctx)
	if err != nil {
		return nil, err
	}
	slog.Info("kalshi bundles fetched", "count", len(bundles), "elapsed", time.Since(start).String())

	// Phase 2: paginate through /events list to discover ALL individual events
	eventTickers, err := c.fetchEventTickers(ctx)
	if err != nil {
		slog.Warn("kalshi event list fetch failed, continuing with bundles only", "error", err)
	}
	slog.Info("kalshi events discovered", "count", len(eventTickers), "elapsed", time.Since(start).String())

	// Phase 3: fetch individual markets for each discovered event
	eventMarkets, err := c.fetchEventMarkets(ctx, eventTickers)
	if err != nil {
		return nil, err
	}
	slog.Info("kalshi event markets fetched", "count", len(eventMarkets), "elapsed", time.Since(start).String())

	// Event markets first so they get priority in the event bus channel (bundles are lower quality)
	all := append(eventMarkets, bundles...)
	slog.Info("kalshi fetch complete", "total", len(all), "bundles", len(bundles), "event_markets", len(eventMarkets), "elapsed", time.Since(start).String())
	return all, nil
}

func (c *KalshiClient) fetchBundles(ctx context.Context) ([]Market, error) {
	var all []Market
	cursor := ""

	for {
		url := fmt.Sprintf("%s/markets?limit=1000&status=open", c.baseURL)
		if cursor != "" {
			url += "&cursor=" + cursor
		}

		result, err := c.fetchPage(ctx, url)
		if err != nil {
			return nil, err
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

type eventsListResponse struct {
	Events []kalshiEvent `json:"events"`
	Cursor string        `json:"cursor"`
}

func (c *KalshiClient) fetchEventTickers(ctx context.Context) (map[string]bool, error) {
	tickers := make(map[string]bool)
	cursor := ""
	limit := 500
	pages := 0

	for {
		url := fmt.Sprintf("%s/events?limit=%d&status=open", c.baseURL, limit)
		if cursor != "" {
			url += "&cursor=" + cursor
		}

		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return nil, err
		}
		c.signRequest(req)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			slog.Warn("kalshi events list request failed", "error", err)
			return tickers, nil // return what we have so far
		}
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, err
		}

		slog.Info("kalshi events list page", "page", pages, "status", resp.StatusCode, "body_len", len(body))

		var list eventsListResponse
		if err := json.Unmarshal(body, &list); err != nil {
			preview := string(body)
			if len(preview) > 300 {
				preview = preview[:300]
			}
			slog.Warn("kalshi events list decode failed", "error", err, "body_preview", preview)
			return tickers, nil
		}

		for _, e := range list.Events {
			tickers[e.EventTicker] = true
		}
		pages++

		if list.Cursor == "" || len(list.Events) < limit {
			break
		}
		cursor = list.Cursor
	}

	slog.Info("kalshi events list complete", "pages", pages, "tickers", len(tickers))
	return tickers, nil
}

func (c *KalshiClient) fetchEventMarkets(ctx context.Context, eventTickers map[string]bool) ([]Market, error) {
	if len(eventTickers) == 0 {
		return nil, nil
	}

	type result struct {
		markets []Market
		err     error
	}

	results := make(chan result, len(eventTickers))
	var wg sync.WaitGroup
	sem := make(chan struct{}, 20) // 20 concurrent workers

	for et := range eventTickers {
		et := et
		wg.Add(1)
		go func() {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			markets, err := c.fetchSingleEvent(ctx, et)
			if err != nil {
				slog.Warn("kalshi event markets fetch failed", "event_ticker", et, "error", err)
				results <- result{err: err}
				return
			}
			results <- result{markets: markets}
		}()
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	var all []Market
	seen := make(map[string]bool)
	for r := range results {
		if r.err != nil {
			continue
		}
		for _, m := range r.markets {
			if !seen[m.MarketID] {
				seen[m.MarketID] = true
				all = append(all, m)
			}
		}
	}

	return all, nil
}

func (c *KalshiClient) fetchSingleEvent(ctx context.Context, eventTicker string) ([]Market, error) {
	url := fmt.Sprintf("%s/events/%s", c.baseURL, eventTicker)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("kalshi event request: %w", err)
	}
	c.signRequest(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("kalshi event do: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("kalshi event read: %w", err)
	}

	var result eventResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("kalshi event decode: %w", err)
	}

	eventTitle := strings.TrimSpace(result.Event.Title)
	var markets []Market
	for _, km := range result.Markets {
		markets = append(markets, c.normalizeWithEvent(km, eventTitle))
	}
	return markets, nil
}

func (c *KalshiClient) signRequest(req *http.Request) {
	if c.keyID != "" {
		req.Header.Set("Kalshi-Key-Id", c.keyID)
	}
}
