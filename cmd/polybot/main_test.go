package main

import (
	"context"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/aaronbateman02/Arby/internal/db"
)

func TestMainStartup(t *testing.T) {
	dsn := getEnvOrDefault("TEST_POSTGRES_DSN", "")
	if dsn == "" {
		t.Skip("TEST_POSTGRES_DSN not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pg, err := db.Connect(ctx, dsn)
	if err != nil {
		t.Fatalf("db connect: %v", err)
	}
	defer pg.Close()

	if err := pg.HealthCheck(ctx); err != nil {
		t.Fatalf("db health: %v", err)
	}
}

func TestHealthEndpoint(t *testing.T) {
	resp, err := http.Get("http://localhost:8087/healthz")
	if err != nil {
		t.Skipf("server not running: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func getEnvOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
