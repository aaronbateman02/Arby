package redis_test

import (
	"context"
	"testing"

	"github.com/aaronbateman02/Arby/internal/redis"
)

func TestConnect_InvalidAddr(t *testing.T) {
	client, err := redis.Connect(context.Background(), "")
	if err == nil {
		client.Close()
		t.Fatal("expected error for empty addr")
	}
}

func TestHealthCheck_NotConnected(t *testing.T) {
	client := &redis.Client{}
	err := client.HealthCheck(context.Background())
	if err == nil {
		t.Fatal("expected error when client is nil")
	}
}
