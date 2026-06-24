package pricing

import (
	"context"
	"log/slog"

	"github.com/aaronbateman02/Arby/pkg/ingestion"
)

type Manager struct {
	cache   *ingestion.PriceCache
	clients []PricingClient
}

func NewManager(cache *ingestion.PriceCache, clients ...PricingClient) *Manager {
	return &Manager{
		cache:   cache,
		clients: clients,
	}
}

func (m *Manager) Run(ctx context.Context) {
	slog.Info("pricing manager started", "clients", len(m.clients))

	for _, client := range m.clients {
		go m.runClient(ctx, client)
	}

	<-ctx.Done()
	slog.Info("pricing manager stopped")

	for _, client := range m.clients {
		if err := client.Close(); err != nil {
			slog.Error("pricing client close", "error", err)
		}
	}
}

func (m *Manager) runClient(ctx context.Context, client PricingClient) {
	updates := client.Updates()
	for {
		select {
		case <-ctx.Done():
			return
		case tick, ok := <-updates:
			if !ok {
				return
			}
			m.cache.Set(tick.Venue, tick.MarketID, tick.Bid, tick.Ask)
		}
	}
}

func (m *Manager) SubscribeAll(ctx context.Context, marketIDs []string) error {
	for _, client := range m.clients {
		if err := client.Subscribe(ctx, marketIDs); err != nil {
			slog.Error("subscribe all", "error", err)
			return err
		}
	}
	return nil
}
