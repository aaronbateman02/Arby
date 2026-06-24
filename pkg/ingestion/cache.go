package ingestion

import (
	"sync"
	"time"
)

type PriceSnapshot struct {
	Venue     string
	MarketID  string
	Bid       float64
	Ask       float64
	UpdatedAt time.Time
}

type PriceCache struct {
	mu     sync.RWMutex
	prices map[string]PriceSnapshot
}

func NewPriceCache() *PriceCache {
	return &PriceCache{
		prices: make(map[string]PriceSnapshot),
	}
}

func (c *PriceCache) Get(venue, marketID string) (PriceSnapshot, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	snap, ok := c.prices[key(venue, marketID)]
	return snap, ok
}

func (c *PriceCache) Set(venue, marketID string, bid, ask float64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.prices[key(venue, marketID)] = PriceSnapshot{
		Venue:     venue,
		MarketID:  marketID,
		Bid:       bid,
		Ask:       ask,
		UpdatedAt: time.Now(),
	}
}

func key(venue, marketID string) string {
	return venue + ":" + marketID
}
