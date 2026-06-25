package discovery

import (
	"context"
	"encoding/json"
	"time"
)

type Outcome struct {
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

type Market struct {
	Venue       string          `json:"venue"`
	MarketID    string          `json:"market_id"`
	Ticker      string          `json:"ticker"`
	Title       string          `json:"title"`
	Description string          `json:"description"`
	Series      string          `json:"series"`
	Category    string          `json:"category"`
	Outcomes    []Outcome       `json:"outcomes"`
	OpenTime    time.Time       `json:"open_time"`
	CloseTime   time.Time       `json:"close_time"`
	Extra       json.RawMessage `json:"extra,omitempty"`
}

type DiscoveryClient interface {
	FetchMarkets(ctx context.Context) ([]Market, error)
	Venue() string
}
