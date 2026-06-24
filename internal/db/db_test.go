package db_test

import (
	"context"
	"testing"

	"github.com/aaronbateman02/Arby/internal/db"
)

func TestConnect_InvalidDSN(t *testing.T) {
	pool, err := db.Connect(context.Background(), "invalid-dsn")
	if err == nil {
		pool.Close()
		t.Fatal("expected error for invalid DSN")
	}
}

func TestHealthCheck_NotConnected(t *testing.T) {
	pool := &db.Pool{}
	err := pool.HealthCheck(context.Background())
	if err == nil {
		t.Fatal("expected error when pool is nil")
	}
}
