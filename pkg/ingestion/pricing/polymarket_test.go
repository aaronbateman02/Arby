package pricing_test

import (
	"testing"

	"github.com/aaronbateman02/Arby/pkg/ingestion/pricing"
)

func TestPolymarketClient_CloseNotConnected(t *testing.T) {
	c := pricing.NewPolymarketClient()
	if err := c.Close(); err != nil {
		t.Fatalf("expected no error closing unconnected client, got: %v", err)
	}
}
