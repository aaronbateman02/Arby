package pricing

import (
	"context"
	"time"
)

type PriceTick struct {
	Venue     string    `json:"venue"`
	MarketID  string    `json:"market_id"`
	Bid       float64   `json:"bid"`
	Ask       float64   `json:"ask"`
	Timestamp time.Time `json:"timestamp"`
}

type PricingClient interface {
	Subscribe(ctx context.Context, marketIDs []string) error
	Unsubscribe(marketIDs []string) error
	Close() error
	Updates() <-chan PriceTick
}
