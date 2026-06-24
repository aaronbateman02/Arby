package matching

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/aaronbateman02/Arby/internal/db"
)

type Market struct {
	ID             string     `json:"id"`
	Venue          string     `json:"venue"`
	VenueMarketID  string     `json:"venue_market_id"`
	Title          string     `json:"title"`
	Description    string     `json:"description"`
	Category       string     `json:"category"`
	StructureType  string     `json:"structure_type"`
	Status         string     `json:"status"`
	ResolutionDate *time.Time `json:"resolution_date,omitempty"`
}

type Candidate struct {
	ID         string  `json:"id"`
	MarketAID  string  `json:"market_a_id"`
	MarketBID  string  `json:"market_b_id"`
	Similarity float64 `json:"similarity"`
	Category   string  `json:"category"`
	Status     string  `json:"status"`
}

type CandidateWithMarkets struct {
	Candidate
	MarketA Market
	MarketB Market
}

type MatchPair struct {
	ID           string  `json:"id"`
	CandidateID  string  `json:"candidate_id"`
	IsSameEvent  bool    `json:"is_same_event"`
	Relationship string  `json:"relationship"`
	Confidence   float64 `json:"confidence"`
	Reasoning    string  `json:"reasoning"`
	LegAModel    string  `json:"leg_a_model"`
	LegBModel    string  `json:"leg_b_model"`
	Status       string  `json:"status"`
}

type Store struct {
	pg *db.Pool
}

func NewStore(pg *db.Pool) *Store {
	return &Store{pg: pg}
}

func (s *Store) CreateTables(ctx context.Context) error {
	sql := `
	CREATE EXTENSION IF NOT EXISTS "pgcrypto";
	CREATE EXTENSION IF NOT EXISTS "vector";

	CREATE TABLE IF NOT EXISTS markets (
		id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		venue           VARCHAR(20) NOT NULL,
		venue_market_id VARCHAR(255) NOT NULL,
		title           TEXT NOT NULL,
		description     TEXT,
		category        VARCHAR(100),
		structure_type  VARCHAR(20) NOT NULL DEFAULT 'BINARY',
		status          VARCHAR(20) NOT NULL DEFAULT 'OPEN',
		resolution_date TIMESTAMPTZ,
		embedding       VECTOR(1024),
		embedding_model VARCHAR(100),
		embedding_updated_at TIMESTAMPTZ,
		first_seen_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		last_updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		UNIQUE (venue, venue_market_id)
	);
	CREATE INDEX IF NOT EXISTS idx_markets_venue ON markets(venue);
	CREATE INDEX IF NOT EXISTS idx_markets_embedding ON markets USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);

	CREATE TABLE IF NOT EXISTS match_candidates (
		id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		market_a_id UUID NOT NULL REFERENCES markets(id),
		market_b_id UUID NOT NULL REFERENCES markets(id),
		similarity  DOUBLE PRECISION NOT NULL,
		category    VARCHAR(100),
		status      VARCHAR(20) NOT NULL DEFAULT 'PENDING',
		created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		UNIQUE (market_a_id, market_b_id)
	);

	CREATE TABLE IF NOT EXISTS match_pairs (
		id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		candidate_id    UUID NOT NULL REFERENCES match_candidates(id) ON DELETE CASCADE,
		is_same_event   BOOLEAN,
		relationship    VARCHAR(20),
		confidence      DOUBLE PRECISION,
		reasoning       TEXT,
		leg_a_model     VARCHAR(100),
		leg_b_model     VARCHAR(100),
		reviewed_at     TIMESTAMPTZ,
		approved_by     VARCHAR(100),
		status          VARCHAR(20) NOT NULL DEFAULT 'PENDING_APPROVAL'
	);

	CREATE TABLE IF NOT EXISTS matching_config (
		key   VARCHAR(100) PRIMARY KEY,
		value TEXT NOT NULL
	);
	INSERT INTO matching_config (key, value) VALUES ('similarity_threshold', '0.80')
	ON CONFLICT (key) DO NOTHING;`

	_, err := s.pg.P().Exec(ctx, sql)
	if err != nil {
		slog.Error("failed to create tables", "error", err)
		return fmt.Errorf("create tables: %w", err)
	}
	return nil
}

func (s *Store) UpsertMarket(ctx context.Context, m Market) error {
	sql := `
	INSERT INTO markets (venue, venue_market_id, title, description, category, structure_type, status, resolution_date)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	ON CONFLICT (venue, venue_market_id) DO UPDATE SET
		title = EXCLUDED.title,
		description = EXCLUDED.description,
		category = EXCLUDED.category,
		structure_type = EXCLUDED.structure_type,
		status = EXCLUDED.status,
		resolution_date = EXCLUDED.resolution_date,
		last_updated_at = NOW()`

	_, err := s.pg.P().Exec(ctx, sql,
		m.Venue, m.VenueMarketID, m.Title, m.Description,
		m.Category, m.StructureType, m.Status, m.ResolutionDate,
	)
	if err != nil {
		slog.Error("failed to upsert market", "error", err, "venue", m.Venue, "venue_market_id", m.VenueMarketID)
		return fmt.Errorf("upsert market: %w", err)
	}
	return nil
}

func (s *Store) GetUnembeddedMarkets(ctx context.Context, limit int) ([]Market, error) {
	sql := `
	SELECT id, venue, venue_market_id, title, description, category, structure_type, status, resolution_date
	FROM markets
	WHERE embedding IS NULL AND title != '' AND description IS NOT NULL AND description != ''
	ORDER BY last_updated_at ASC
	LIMIT $1`

	rows, err := s.pg.P().Query(ctx, sql, limit)
	if err != nil {
		slog.Error("failed to query unembedded markets", "error", err)
		return nil, fmt.Errorf("get unembedded markets: %w", err)
	}
	defer rows.Close()

	var markets []Market
	for rows.Next() {
		var m Market
		if err := rows.Scan(
			&m.ID, &m.Venue, &m.VenueMarketID, &m.Title, &m.Description,
			&m.Category, &m.StructureType, &m.Status, &m.ResolutionDate,
		); err != nil {
			slog.Error("failed to scan market row", "error", err)
			return nil, fmt.Errorf("scan market: %w", err)
		}
		markets = append(markets, m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration: %w", err)
	}
	return markets, nil
}

func (s *Store) GetEmbeddedMarkets(ctx context.Context, limit int) ([]Market, error) {
	sql := `
	SELECT id, venue, venue_market_id, title, description, category, structure_type, status, resolution_date
	FROM markets
	WHERE embedding IS NOT NULL
	ORDER BY last_updated_at DESC
	LIMIT $1`

	rows, err := s.pg.P().Query(ctx, sql, limit)
	if err != nil {
		slog.Error("failed to query embedded markets", "error", err)
		return nil, fmt.Errorf("get embedded markets: %w", err)
	}
	defer rows.Close()

	var markets []Market
	for rows.Next() {
		var m Market
		if err := rows.Scan(
			&m.ID, &m.Venue, &m.VenueMarketID, &m.Title, &m.Description,
			&m.Category, &m.StructureType, &m.Status, &m.ResolutionDate,
		); err != nil {
			slog.Error("failed to scan market row", "error", err)
			return nil, fmt.Errorf("scan market: %w", err)
		}
		markets = append(markets, m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration: %w", err)
	}
	return markets, nil
}

func (s *Store) UpsertEmbedding(ctx context.Context, id string, vector []float64) error {
	vecStr := joinFloats(vector, ",")
	_, err := s.pg.P().Exec(ctx, "UPDATE markets SET embedding = $2::vector WHERE id = $1", id, vecStr)
	if err != nil {
		slog.Error("failed to upsert embedding", "error", err, "id", id)
		return fmt.Errorf("upsert embedding: %w", err)
	}
	return nil
}

func (s *Store) InsertCandidate(ctx context.Context, c Candidate) error {
	sql := `
	INSERT INTO match_candidates (market_a_id, market_b_id, similarity, category, status)
	VALUES ($1, $2, $3, $4, $5)
	ON CONFLICT (market_a_id, market_b_id) DO NOTHING`

	_, err := s.pg.P().Exec(ctx, sql, c.MarketAID, c.MarketBID, c.Similarity, c.Category, c.Status)
	if err != nil {
		slog.Error("failed to insert candidate", "error", err)
		return fmt.Errorf("insert candidate: %w", err)
	}
	return nil
}

func (s *Store) GetPendingCandidates(ctx context.Context, limit int) ([]Candidate, error) {
	sql := `
	SELECT id, market_a_id, market_b_id, similarity, category, status
	FROM match_candidates
	WHERE status = 'PENDING'
	LIMIT $1`

	rows, err := s.pg.P().Query(ctx, sql, limit)
	if err != nil {
		slog.Error("failed to query pending candidates", "error", err)
		return nil, fmt.Errorf("get pending candidates: %w", err)
	}
	defer rows.Close()

	var candidates []Candidate
	for rows.Next() {
		var c Candidate
		if err := rows.Scan(&c.ID, &c.MarketAID, &c.MarketBID, &c.Similarity, &c.Category, &c.Status); err != nil {
			slog.Error("failed to scan candidate row", "error", err)
			return nil, fmt.Errorf("scan candidate: %w", err)
		}
		candidates = append(candidates, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration: %w", err)
	}
	return candidates, nil
}

func (s *Store) GetPendingCandidatesWithMarkets(ctx context.Context, limit int) ([]CandidateWithMarkets, error) {
	sql := `
	SELECT
		c.id, c.market_a_id, c.market_b_id, c.similarity, c.category, c.status,
		ma.id, ma.venue, ma.venue_market_id, ma.title, ma.description, ma.category, ma.structure_type, ma.status, ma.resolution_date,
		mb.id, mb.venue, mb.venue_market_id, mb.title, mb.description, mb.category, mb.structure_type, mb.status, mb.resolution_date
	FROM match_candidates c
	JOIN markets ma ON ma.id = c.market_a_id
	JOIN markets mb ON mb.id = c.market_b_id
	WHERE c.status = 'PENDING'
	LIMIT $1`

	rows, err := s.pg.P().Query(ctx, sql, limit)
	if err != nil {
		slog.Error("failed to query pending candidates with markets", "error", err)
		return nil, fmt.Errorf("get pending candidates with markets: %w", err)
	}
	defer rows.Close()

	var results []CandidateWithMarkets
	for rows.Next() {
		var cwm CandidateWithMarkets
		if err := rows.Scan(
			&cwm.ID, &cwm.MarketAID, &cwm.MarketBID, &cwm.Similarity, &cwm.Category, &cwm.Status,
			&cwm.MarketA.ID, &cwm.MarketA.Venue, &cwm.MarketA.VenueMarketID, &cwm.MarketA.Title, &cwm.MarketA.Description,
			&cwm.MarketA.Category, &cwm.MarketA.StructureType, &cwm.MarketA.Status, &cwm.MarketA.ResolutionDate,
			&cwm.MarketB.ID, &cwm.MarketB.Venue, &cwm.MarketB.VenueMarketID, &cwm.MarketB.Title, &cwm.MarketB.Description,
			&cwm.MarketB.Category, &cwm.MarketB.StructureType, &cwm.MarketB.Status, &cwm.MarketB.ResolutionDate,
		); err != nil {
			slog.Error("failed to scan candidate with markets row", "error", err)
			return nil, fmt.Errorf("scan candidate with markets: %w", err)
		}
		results = append(results, cwm)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration: %w", err)
	}
	return results, nil
}

func (s *Store) UpdateCandidateStatus(ctx context.Context, id, status string) error {
	_, err := s.pg.P().Exec(ctx, "UPDATE match_candidates SET status = $2 WHERE id = $1", id, status)
	if err != nil {
		slog.Error("failed to update candidate status", "error", err, "id", id, "status", status)
		return fmt.Errorf("update candidate status: %w", err)
	}
	return nil
}

func (s *Store) InsertMatchPair(ctx context.Context, p MatchPair) error {
	sql := `
	INSERT INTO match_pairs (candidate_id, is_same_event, relationship, confidence, reasoning, leg_a_model, leg_b_model, status)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := s.pg.P().Exec(ctx, sql,
		p.CandidateID, p.IsSameEvent, p.Relationship, p.Confidence,
		p.Reasoning, p.LegAModel, p.LegBModel, p.Status,
	)
	if err != nil {
		slog.Error("failed to insert match pair", "error", err)
		return fmt.Errorf("insert match pair: %w", err)
	}
	return nil
}

func (s *Store) GetMatchPairs(ctx context.Context, status string) ([]MatchPair, error) {
	var sql string
	var args []any

	if status != "" {
		sql = `SELECT id, candidate_id, is_same_event, relationship, confidence, reasoning, leg_a_model, leg_b_model, status FROM match_pairs WHERE status = $1`
		args = append(args, status)
	} else {
		sql = `SELECT id, candidate_id, is_same_event, relationship, confidence, reasoning, leg_a_model, leg_b_model, status FROM match_pairs`
	}

	rows, err := s.pg.P().Query(ctx, sql, args...)
	if err != nil {
		slog.Error("failed to query match pairs", "error", err)
		return nil, fmt.Errorf("get match pairs: %w", err)
	}
	defer rows.Close()

	var pairs []MatchPair
	for rows.Next() {
		var p MatchPair
		if err := rows.Scan(
			&p.ID, &p.CandidateID, &p.IsSameEvent, &p.Relationship,
			&p.Confidence, &p.Reasoning, &p.LegAModel, &p.LegBModel, &p.Status,
		); err != nil {
			slog.Error("failed to scan match pair row", "error", err)
			return nil, fmt.Errorf("scan match pair: %w", err)
		}
		pairs = append(pairs, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration: %w", err)
	}
	return pairs, nil
}

func (s *Store) UpdateMatchPairStatus(ctx context.Context, id, status string) error {
	_, err := s.pg.P().Exec(ctx, "UPDATE match_pairs SET status = $2 WHERE id = $1", id, status)
	if err != nil {
		slog.Error("failed to update match pair status", "error", err, "id", id, "status", status)
		return fmt.Errorf("update match pair status: %w", err)
	}
	return nil
}

func (s *Store) GetMatchPairsWithDetails(ctx context.Context, status string) ([]matchPairResponse, error) {
	var sql string
	var args []any

	if status != "" {
		sql = `
		SELECT mp.id, mp.candidate_id, mp.is_same_event, mp.relationship, mp.confidence, mp.reasoning, mp.leg_a_model, mp.leg_b_model, mp.status,
		       ma.title, ma.venue, mb.title, mb.venue, c.category
		FROM match_pairs mp
		JOIN match_candidates c ON c.id = mp.candidate_id
		JOIN markets ma ON ma.id = c.market_a_id
		JOIN markets mb ON mb.id = c.market_b_id
		WHERE mp.status = $1`
		args = append(args, status)
	} else {
		sql = `
		SELECT mp.id, mp.candidate_id, mp.is_same_event, mp.relationship, mp.confidence, mp.reasoning, mp.leg_a_model, mp.leg_b_model, mp.status,
		       ma.title, ma.venue, mb.title, mb.venue, c.category
		FROM match_pairs mp
		JOIN match_candidates c ON c.id = mp.candidate_id
		JOIN markets ma ON ma.id = c.market_a_id
		JOIN markets mb ON mb.id = c.market_b_id`
	}

	rows, err := s.pg.P().Query(ctx, sql, args...)
	if err != nil {
		slog.Error("failed to query match pairs with details", "error", err)
		return nil, fmt.Errorf("get match pairs with details: %w", err)
	}
	defer rows.Close()

	var pairs []matchPairResponse
	for rows.Next() {
		var p matchPairResponse
		if err := rows.Scan(
			&p.ID, &p.CandidateID, &p.IsSameEvent, &p.Relationship,
			&p.Confidence, &p.Reasoning, &p.LegAModel, &p.LegBModel, &p.Status,
			&p.MarketATitle, &p.VenueA, &p.MarketBTitle, &p.VenueB, &p.Category,
		); err != nil {
			slog.Error("failed to scan match pair with details row", "error", err)
			return nil, fmt.Errorf("scan match pair with details: %w", err)
		}
		pairs = append(pairs, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration: %w", err)
	}
	return pairs, nil
}

func (s *Store) GetCandidateByID(ctx context.Context, id string) (*Candidate, error) {
	sql := `SELECT id, market_a_id, market_b_id, similarity, category, status FROM match_candidates WHERE id = $1`
	var c Candidate
	err := s.pg.P().QueryRow(ctx, sql, id).Scan(
		&c.ID, &c.MarketAID, &c.MarketBID, &c.Similarity, &c.Category, &c.Status,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		slog.Error("failed to query candidate by id", "error", err, "id", id)
		return nil, fmt.Errorf("get candidate by id: %w", err)
	}
	return &c, nil
}

func (s *Store) GetMarketByID(ctx context.Context, id string) (*Market, error) {
	sql := `SELECT id, venue, venue_market_id, title, description, category, structure_type, status, resolution_date FROM markets WHERE id = $1`
	var m Market
	err := s.pg.P().QueryRow(ctx, sql, id).Scan(
		&m.ID, &m.Venue, &m.VenueMarketID, &m.Title, &m.Description,
		&m.Category, &m.StructureType, &m.Status, &m.ResolutionDate,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		slog.Error("failed to query market by id", "error", err, "id", id)
		return nil, fmt.Errorf("get market by id: %w", err)
	}
	return &m, nil
}

func (s *Store) GetSimilarityThreshold(ctx context.Context) (float64, error) {
	var val string
	err := s.pg.P().QueryRow(ctx,
		`SELECT value FROM matching_config WHERE key = 'similarity_threshold'`,
	).Scan(&val)
	if err != nil {
		return 0.80, fmt.Errorf("get similarity threshold: %w", err)
	}
	f, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return 0.80, nil
	}
	return f, nil
}

func (s *Store) SetSimilarityThreshold(ctx context.Context, val float64) error {
	_, err := s.pg.P().Exec(ctx,
		`INSERT INTO matching_config (key, value) VALUES ('similarity_threshold', $1)
		 ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value`,
		fmt.Sprintf("%.4f", val),
	)
	if err != nil {
		return fmt.Errorf("set similarity threshold: %w", err)
	}
	return nil
}

func joinFloats(v []float64, sep string) string {
	if len(v) == 0 {
		return "[]"
	}
	parts := make([]string, len(v))
	for i, f := range v {
		parts[i] = fmt.Sprint(f)
	}
	return "[" + strings.Join(parts, sep) + "]"
}
