package config_test

import (
	"testing"

	"github.com/aaronbateman02/Arby/internal/config"
)

func TestConfig_ValidMinimal(t *testing.T) {
	cfg := config.Config{
		PostgresDSN: "postgres://user:pass@localhost:5432/arby",
		RedisAddr:   "localhost:6379",
		ListenAddr:  ":8086",
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestConfig_MissingPostgresDSN(t *testing.T) {
	cfg := config.Config{
		RedisAddr:  "localhost:6379",
		ListenAddr: ":8086",
	}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected validation error for missing PostgresDSN")
	}
}
