package main

import (
	"context"
	"encoding/json"
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
	"github.com/aaronbateman02/Arby/pkg/matching"
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

	eventBus := bus.New(100000)
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

	matchingStore := matching.NewStore(pg)
	if err := matchingStore.CreateTables(ctx); err != nil {
		slog.Error("matching tables", "error", err)
		os.Exit(1)
	}
	slog.Info("matching store initialized")

	go func() {
		marketCh, err := eventBus.Subscribe("MarketDiscovered")
		if err != nil {
			slog.Error("subscribe market discovered", "error", err)
			return
		}
		slog.Info("subscribed to MarketDiscovered events")

		for {
			select {
			case msg := <-marketCh:
				var discMarket discovery.Market
				if err := json.Unmarshal(msg.Payload, &discMarket); err != nil {
					slog.Error("unmarshal discovered market", "error", err)
					continue
				}

				desc := discMarket.Description
				if desc == "" {
					desc = discMarket.Title
					if discMarket.Series != "" {
						desc = discMarket.Series + " - " + discMarket.Title
					}
				}

				var resDate *time.Time
				if !discMarket.CloseTime.IsZero() {
					resDate = &discMarket.CloseTime
				}

				m := matching.Market{
					Venue:          discMarket.Venue,
					VenueMarketID:  discMarket.MarketID,
					Title:          discMarket.Title,
					Description:    desc,
					Category:       discMarket.Category,
					StructureType:  "BINARY",
					Status:         "OPEN",
					ResolutionDate: resDate,
				}

				if m.Category == "" {
					m.Category = matching.CategorizeMarket(m.Venue, m.VenueMarketID, m.Title)
				}
				m.Subcategory = matching.SubcategorizeMarket(m.Venue, m.VenueMarketID, m.Title)
				m.MarketType = matching.CategorizeMarketType(m.Title)

				if err := matchingStore.UpsertMarket(ctx, m); err != nil {
					slog.Error("upsert discovered market", "error", err, "market_id", discMarket.MarketID)
					continue
				}
				slog.Debug("market upserted", "venue", discMarket.Venue, "id", discMarket.MarketID, "title", discMarket.Title)

			case <-ctx.Done():
				slog.Info("market subscriber stopped")
				return
			}
		}
	}()

	h := health.New(
		func(ctx context.Context) error { return pg.HealthCheck(ctx) },
		func(ctx context.Context) error { return rdb.HealthCheck(ctx) },
	)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", h.LivenessHandler())
	mux.HandleFunc("GET /readyz", h.ReadinessHandler())
	mux.Handle("GET /metrics", met.Handler())

	matchingHandler := matching.NewHandler(matchingStore)
	matchingHandler.WireRoutes(mux)

	discoverer := matching.NewCandidateDiscoverer(
		matchingStore,
		5*time.Minute,
		0.80,
		20,
	)
	go discoverer.Run(ctx)
	slog.Info("candidate discoverer started")

	reviewer := matching.NewReviewer(
		matchingStore,
		5*time.Minute,
		cfg.OpenRouterAPIKey,
		cfg.OpenRouterBaseURL,
		"openai/gpt-4o",
		"openai/gpt-4o-mini",
		40,
		0.90,
	)
	go reviewer.Run(ctx)
	slog.Info("matching reviewer started")

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
