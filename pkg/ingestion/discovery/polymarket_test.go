package discovery_test

import (
	"testing"

	"github.com/aaronbateman02/Arby/pkg/ingestion/discovery"
)

func TestPolymarketClient_Venue(t *testing.T) {
	c := discovery.NewPolymarketClient()
	if c.Venue() != "POLYMARKET" {
		t.Fatalf("expected POLYMARKET, got %s", c.Venue())
	}
}
