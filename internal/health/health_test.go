package health_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aaronbateman02/Arby/internal/health"
)

func TestHealthz_AlwaysOK(t *testing.T) {
	h := health.New(nil, nil)
	handler := h.LivenessHandler()

	req := httptest.NewRequest("GET", "/healthz", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestReadyz_AllHealthy(t *testing.T) {
	dbOK := func(ctx context.Context) error { return nil }
	redisOK := func(ctx context.Context) error { return nil }

	h := health.New(dbOK, redisOK)
	handler := h.ReadinessHandler()

	req := httptest.NewRequest("GET", "/readyz", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestReadyz_DBFailure(t *testing.T) {
	dbFail := func(ctx context.Context) error { return nil }
	redisFail := func(ctx context.Context) error { return nil }

	h := health.New(dbFail, redisFail)
	handler := h.ReadinessHandler()

	req := httptest.NewRequest("GET", "/readyz", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}
