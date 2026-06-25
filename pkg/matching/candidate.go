package matching

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"strings"
	"time"
	"unicode"
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

// compositeSimilarity combines embedding similarity, keyword overlap, category/subcategory/market_type
// alignment, and date proximity into a single score between 0 and 1.
func compositeSimilarity(a, b Market, embeddingSim float64) float64 {
	kw := keywordJaccard(a.Title+" "+a.Description, b.Title+" "+b.Description)
	cat := boolToFloat(a.Category != "" && a.Category == b.Category)
	sub := boolToFloat(a.Subcategory != "" && a.Subcategory == b.Subcategory)
	mt := boolToFloat(a.MarketType != "" && a.MarketType == b.MarketType)
	dt := dateProximityScore(a.ResolutionDate, b.ResolutionDate)

	return embeddingSim*0.25 + kw*0.25 + cat*0.15 + sub*0.10 + mt*0.10 + dt*0.15
}

func boolToFloat(v bool) float64 {
	if v {
		return 1.0
	}
	return 0.0
}

// keywordJaccari computes Jaccard similarity of cleaned word tokens.
func keywordJaccard(s1, s2 string) float64 {
	t1 := tokenize(s1)
	t2 := tokenize(s2)
	if len(t1) == 0 && len(t2) == 0 {
		return 0.0
	}
	set1 := make(map[string]struct{}, len(t1))
	for _, w := range t1 {
		set1[w] = struct{}{}
	}
	intersection := 0
	for _, w := range t2 {
		if _, ok := set1[w]; ok {
			intersection++
			delete(set1, w)
		}
	}
	union := len(t1) + len(t2) - intersection
	if union == 0 {
		return 0.0
	}
	return float64(intersection) / float64(union)
}

// tokenize splits text into lowercase word tokens, dropping punctuation and short words.
func tokenize(s string) []string {
	var tokens []string
	var buf []rune
	for _, r := range strings.ToLower(s) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			buf = append(buf, r)
		} else {
			if len(buf) >= 2 {
				tokens = append(tokens, string(buf))
			}
			buf = buf[:0]
		}
	}
	if len(buf) >= 2 {
		tokens = append(tokens, string(buf))
	}
	return tokens
}

// dateProximityScore returns 1.0 if same day, decaying to 0 at 30 days apart.
func dateProximityScore(da, db *time.Time) float64 {
	if da == nil || db == nil {
		return 0.0
	}
	diff := da.Sub(*db)
	if diff < 0 {
		diff = -diff
	}
	days := diff.Hours() / 24
	if days >= 30 {
		return 0.0
	}
	return 1.0 - days/30.0
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
