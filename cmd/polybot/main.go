package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aaronbateman02/Arby/internal/auth"
	"github.com/aaronbateman02/Arby/internal/bus"
	"github.com/aaronbateman02/Arby/internal/config"
	"github.com/aaronbateman02/Arby/internal/db"
	"github.com/aaronbateman02/Arby/internal/health"
	"github.com/aaronbateman02/Arby/internal/logging"
	"github.com/aaronbateman02/Arby/internal/metrics"
	arbyredis "github.com/aaronbateman02/Arby/internal/redis"
	"github.com/aaronbateman02/Arby/pkg/ingestion"
	"github.com/aaronbateman02/Arby/pkg/ingestion/discovery"
	"github.com/aaronbateman02/Arby/pkg/ingestion/pricing"
)

func main() {
	cfg, err := config.LoadFromEnv()
	if err != nil {
		slog.Error("config", "error", err)
		os.Exit(1)
	}

	logger, err := logging.New(cfg.LogLevel)
	if err != nil {
		slog.Error("logging", "error", err)
		os.Exit(1)
	}
	slog.SetDefault(logger)
	slog.Info("starting arby", "listen", cfg.ListenAddr)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pg, err := db.Connect(ctx, cfg.PostgresDSN)
	if err != nil {
		slog.Error("postgres", "error", err)
		os.Exit(1)
	}
	defer pg.Close()
	slog.Info("postgres connected")

	rdb, err := arbyredis.Connect(ctx, cfg.RedisAddr)
	if err != nil {
		slog.Error("redis", "error", err)
		os.Exit(1)
	}
	defer rdb.Close()
	slog.Info("redis connected")

	met := metrics.New()
	slog.Info("metrics initialized")

	eventBus := bus.New(1000)
	slog.Info("event bus initialized")

	authenticator, err := auth.New(cfg.JWTSecretKey)
	if err != nil {
		slog.Error("auth", "error", err)
		os.Exit(1)
	}
	slog.Info("auth initialized")

	priceCache := ingestion.NewPriceCache()
	slog.Info("price cache initialized")

	discKalshi := discovery.NewKalshiClient(cfg.KalshiKeyID, cfg.KalshiPrivateKeyPEM)
	discPoly := discovery.NewPolymarketClient()
	discScanner := discovery.NewScanner(eventBus, cfg.Ingestion.DiscoveryInterval, discKalshi, discPoly)
	go discScanner.Run(ctx)
	slog.Info("discovery scanner started")

	priceKalshi := pricing.NewKalshiClient(cfg.KalshiKeyID, cfg.KalshiPrivateKeyPEM)
	pricePoly := pricing.NewPolymarketClient()
	pricingMgr := pricing.NewManager(priceCache, priceKalshi, pricePoly)
	go pricingMgr.Run(ctx)
	slog.Info("pricing manager started")

	h := health.New(
		func(ctx context.Context) error { return pg.HealthCheck(ctx) },
		func(ctx context.Context) error { return rdb.HealthCheck(ctx) },
	)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", h.LivenessHandler())
	mux.HandleFunc("GET /readyz", h.ReadinessHandler())
	mux.Handle("GET /metrics", met.Handler())

	srv := &http.Server{
		Addr:    cfg.ListenAddr,
		Handler: mux,
	}

	go func() {
		slog.Info("http server listening", "addr", cfg.ListenAddr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("http server", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	slog.Info("shutdown signal received", "signal", sig)

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("http server shutdown", "error", err)
	}

	cancel()
	pg.Close()
	rdb.Close()
	slog.Info("shutdown complete")

	_ = eventBus
	_ = authenticator
	_ = priceCache
}
