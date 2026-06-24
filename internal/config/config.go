package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/go-playground/validator/v10"
)

type Config struct {
	PostgresDSN        string `validate:"required"`
	RedisAddr          string `validate:"required"`
	ListenAddr         string `validate:"required"`
	LogLevel           string `validate:"omitempty,oneof=debug info warn error"`
	OpenRouterAPIKey   string
	OpenRouterBaseURL  string
	KalshiKeyID        string
	KalshiPrivateKeyPEM string
	PolymarketProxyURL string
	JWTSecretKey       string
}

func (c *Config) Validate() error {
	v := validator.New()
	return v.Struct(c)
}

func LoadFromEnv() (*Config, error) {
	cfg := &Config{
		PostgresDSN:        readSecretOr(os.Getenv("POSTGRES_DSN_FILE")),
		RedisAddr:          envOrDefault("REDIS_ADDR", "localhost:6379"),
		ListenAddr:         envOrDefault("LISTEN_ADDR", ":8086"),
		LogLevel:           envOrDefault("LOG_LEVEL", "info"),
		OpenRouterAPIKey:   readSecretOr(os.Getenv("OPENROUTER_API_KEY_FILE")),
		OpenRouterBaseURL:  envOrDefault("OPENROUTER_BASE_URL", "https://openrouter.ai/api/v1"),
		KalshiKeyID:        readSecretOr(os.Getenv("KALSHI_KEY_ID_FILE")),
		KalshiPrivateKeyPEM: readSecretOr(os.Getenv("KALSHI_PRIVATE_KEY_FILE")),
		PolymarketProxyURL: os.Getenv("POLYMARKET_PROXY_URL"),
		JWTSecretKey:       readSecretOr(os.Getenv("JWT_SECRET_KEY_FILE")),
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

func readSecretOr(path string) string {
	if path == "" {
		return ""
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
