package discovery

import (
	"context"
	"log/slog"
	"time"

	"github.com/aaronbateman02/Arby/internal/bus"
)

type Scanner struct {
	clients  []DiscoveryClient
	eventBus *bus.Bus
	interval time.Duration
}

func NewScanner(eventBus *bus.Bus, interval time.Duration, clients ...DiscoveryClient) *Scanner {
	return &Scanner{
		clients:  clients,
		eventBus: eventBus,
		interval: interval,
	}
}

func (s *Scanner) Run(ctx context.Context) {
	slog.Info("discovery scanner started", "interval", s.interval)

	if err := s.scanAll(ctx); err != nil {
		slog.Error("initial discovery scan", "error", err)
	}

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Info("discovery scanner stopped")
			return
		case <-ticker.C:
			if err := s.scanAll(ctx); err != nil {
				slog.Error("discovery scan cycle", "error", err)
			}
		}
	}
}

func (s *Scanner) scanAll(ctx context.Context) error {
	for _, client := range s.clients {
		markets, err := client.FetchMarkets(ctx)
		if err != nil {
			slog.Error("discovery fetch", "venue", client.Venue(), "error", err)
			continue
		}
		slog.Info("discovery fetched markets", "venue", client.Venue(), "count", len(markets))

		for _, m := range markets {
			if err := s.eventBus.PublishTyped("MarketDiscovered", m); err != nil {
				slog.Error("discovery publish event", "market", m.MarketID, "error", err)
			}
		}
	}
	return nil
}
