package pricing_test

import (
	"context"
	"testing"
	"time"

	"github.com/aaronbateman02/Arby/pkg/ingestion"
	"github.com/aaronbateman02/Arby/pkg/ingestion/pricing"
)

type mockPricingClient struct {
	updates chan pricing.PriceTick
}

func newMockPricingClient() *mockPricingClient {
	return &mockPricingClient{updates: make(chan pricing.PriceTick, 100)}
}

func (m *mockPricingClient) Subscribe(ctx context.Context, ids []string) error { return nil }
func (m *mockPricingClient) Unsubscribe(ids []string) error                   { return nil }
func (m *mockPricingClient) Close() error                                     { return nil }
func (m *mockPricingClient) Updates() <-chan pricing.PriceTick                { return m.updates }

func TestManager_PriceCacheUpdates(t *testing.T) {
	cache := ingestion.NewPriceCache()
	mock := newMockPricingClient()

	mgr := pricing.NewManager(cache, mock)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go mgr.Run(ctx)

	mock.updates <- pricing.PriceTick{
		Venue:    "KALSHI",
		MarketID: "test-1",
		Bid:      0.50,
		Ask:      0.55,
	}

	time.Sleep(50 * time.Millisecond)

	snap, ok := cache.Get("KALSHI", "test-1")
	if !ok {
		t.Fatal("expected price in cache")
	}
	if snap.Bid != 0.50 {
		t.Fatalf("expected bid 0.50, got %f", snap.Bid)
	}
}
