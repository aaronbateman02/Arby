package matching

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"strings"
	"time"
)

type annResult struct {
	ID         string
	Venue      string
	Similarity float64
}

type CandidateDiscoverer struct {
	store               *Store
	interval            time.Duration
	similarityThreshold float64
	annLimit            int
}

func NewCandidateDiscoverer(store *Store, interval time.Duration, similarityThreshold float64, annLimit int) *CandidateDiscoverer {
	return &CandidateDiscoverer{
		store:               store,
		interval:            interval,
		similarityThreshold: similarityThreshold,
		annLimit:            annLimit,
	}
}

func (d *CandidateDiscoverer) Run(ctx context.Context) {
	slog.Info("candidate discoverer started", "interval", d.interval, "threshold", d.similarityThreshold)
	if err := d.RunOnce(ctx); err != nil {
		slog.Error("candidate discoverer run once", "error", err)
	}
	ticker := time.NewTicker(d.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if err := d.RunOnce(ctx); err != nil {
				slog.Error("candidate discoverer", "error", err)
			}
		case <-ctx.Done():
			slog.Info("candidate discoverer stopped")
			return
		}
	}
}

func (d *CandidateDiscoverer) RunOnce(ctx context.Context) error {
	threshold, err := d.store.GetSimilarityThreshold(ctx)
	if err != nil {
		threshold = d.similarityThreshold
	}

	markets, err := d.store.GetEmbeddedMarkets(ctx, 10000)
	if err != nil {
		return fmt.Errorf("get embedded markets: %w", err)
	}

	var kalshi, poly []Market
	for _, m := range markets {
		switch strings.ToUpper(m.Venue) {
		case "KALSHI":
			kalshi = append(kalshi, m)
		case "POLYMARKET":
			poly = append(poly, m)
		}
	}

	if len(kalshi) == 0 || len(poly) == 0 {
		slog.Info("candidate discoverer skipping: one or both venue groups empty",
			"kalshi", len(kalshi), "polymarket", len(poly))
		return nil
	}

	var inserted int
	polyByID := indexByID(poly)

	venueB := "Polymarket"
	for _, a := range kalshi {
		results, err := d.queryANN(ctx, a.ID, venueB, a.Category, a.Subcategory, a.MarketType, d.annLimit)
		if err != nil {
			slog.Error("ann query failed", "error", err, "market_id", a.ID, "venue", venueB)
			continue
		}
		for _, r := range results {
			b, ok := polyByID[r.ID]
			if !ok {
				continue
			}
			comp := compositeSimilarity(a, b, r.Similarity)
			if comp >= threshold {
				_ = d.store.InsertCandidate(ctx, Candidate{
					MarketAID:  a.ID,
					MarketBID:  r.ID,
					Similarity: math.Round(comp*10000) / 10000,
					Category:   b.Category,
					Status:     "PENDING",
				})
				inserted++
			}
		}
	}

	kalshiByID := indexByID(kalshi)
	venueB = "Kalshi"
	for _, a := range poly {
		results, err := d.queryANN(ctx, a.ID, venueB, a.Category, a.Subcategory, a.MarketType, d.annLimit)
		if err != nil {
			slog.Error("ann query failed", "error", err, "market_id", a.ID, "venue", venueB)
			continue
		}
		for _, r := range results {
			b, ok := kalshiByID[r.ID]
			if !ok {
				continue
			}
			comp := compositeSimilarity(a, b, r.Similarity)
			if comp >= threshold {
				_ = d.store.InsertCandidate(ctx, Candidate{
					MarketAID:  a.ID,
					MarketBID:  r.ID,
					Similarity: math.Round(comp*10000) / 10000,
					Category:   b.Category,
					Status:     "PENDING",
				})
				inserted++
			}
		}
	}

	slog.Info("candidate discoverer pass complete",
		"kalshi", len(kalshi), "polymarket", len(poly), "inserted", inserted)
	return nil
}

func indexByID(markets []Market) map[string]Market {
	m := make(map[string]Market, len(markets))
	for i := range markets {
		m[markets[i].ID] = markets[i]
	}
	return m
}

func compositeSimilarity(a, b Market, embeddingSim float64) float64 {
	return embeddingSim
}

func (d *CandidateDiscoverer) queryANN(ctx context.Context, marketID, venue, category, subcategory, marketType string, limit int) ([]annResult, error) {
	sql := `
	SELECT m2.id, m2.venue, 1 - (m1.embedding <=> m2.embedding) AS similarity
	FROM markets m1, markets m2
	WHERE m1.id = $1
	  AND m2.venue = $2
	  AND m2.embedding IS NOT NULL
	  AND m2.id != $1
	  AND ($3 = '' OR m2.category = $3)
	  AND ($4 = '' OR m2.subcategory = $4)
	  AND ($5 = '' OR m2.market_type = $5)
	ORDER BY m1.embedding <=> m2.embedding
	LIMIT $6`

	rows, err := d.store.pg.P().Query(ctx, sql, marketID, venue, category, subcategory, marketType, limit)
	if err != nil {
		return nil, fmt.Errorf("ann query: %w", err)
	}
	defer rows.Close()

	var results []annResult
	for rows.Next() {
		var r annResult
		if err := rows.Scan(&r.ID, &r.Venue, &r.Similarity); err != nil {
			return nil, fmt.Errorf("scan ann result: %w", err)
		}
		results = append(results, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration: %w", err)
	}
	return results, nil
}
