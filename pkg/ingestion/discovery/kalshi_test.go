package discovery_test

import (
	"testing"

	"github.com/aaronbateman02/Arby/pkg/ingestion/discovery"
)

func TestKalshiClient_Venue(t *testing.T) {
	c := discovery.NewKalshiClient("", "")
	if c.Venue() != "KALSHI" {
		t.Fatalf("expected KALSHI, got %s", c.Venue())
	}
}
