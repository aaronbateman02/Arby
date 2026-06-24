package ingestion_test

import (
	"testing"

	"github.com/aaronbateman02/Arby/pkg/ingestion"
)

func TestCache_SetAndGet(t *testing.T) {
	c := ingestion.NewPriceCache()
	c.Set("KALSHI", "market-1", 0.52, 0.54)

	snap, ok := c.Get("KALSHI", "market-1")
	if !ok {
		t.Fatal("expected to find market-1")
	}
	if snap.Bid != 0.52 {
		t.Fatalf("expected bid 0.52, got %f", snap.Bid)
	}
	if snap.Ask != 0.54 {
		t.Fatalf("expected ask 0.54, got %f", snap.Ask)
	}
	if snap.Venue != "KALSHI" {
		t.Fatalf("expected venue KALSHI, got %s", snap.Venue)
	}
}

func TestCache_MissingMarket(t *testing.T) {
	c := ingestion.NewPriceCache()
	_, ok := c.Get("KALSHI", "nonexistent")
	if ok {
		t.Fatal("expected false for missing market")
	}
}

func TestCache_Overwrite(t *testing.T) {
	c := ingestion.NewPriceCache()
	c.Set("KALSHI", "m1", 0.50, 0.55)
	c.Set("KALSHI", "m1", 0.51, 0.56)

	snap, ok := c.Get("KALSHI", "m1")
	if !ok {
		t.Fatal("expected to find m1")
	}
	if snap.Bid != 0.51 {
		t.Fatalf("expected bid 0.51, got %f", snap.Bid)
	}
	if snap.Ask != 0.56 {
		t.Fatalf("expected ask 0.56, got %f", snap.Ask)
	}
}

func TestCache_ConcurrentSafety(t *testing.T) {
	c := ingestion.NewPriceCache()
	done := make(chan bool)

	go func() {
		for i := 0; i < 100; i++ {
			c.Set("KALSHI", "m1", float64(i)/100, float64(i+1)/100)
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 100; i++ {
			c.Get("KALSHI", "m1")
		}
		done <- true
	}()

	<-done
	<-done
}
