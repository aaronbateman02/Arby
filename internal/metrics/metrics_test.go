package metrics_test

import (
	"testing"

	"github.com/aaronbateman02/Arby/internal/metrics"
)

func TestRegisterAndGet(t *testing.T) {
	m := metrics.New()
	if m == nil {
		t.Fatal("expected non-nil metrics")
	}

	httpRequests := m.Counter("http_requests_total", "Total HTTP requests", "method", "path")
	if httpRequests == nil {
		t.Fatal("expected non-nil counter")
	}
}
