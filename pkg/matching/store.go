package matching

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/aaronbateman02/Arby/internal/db"
)

type Event struct {
	ID            string     `json:"id"`
	Venue         string     `json:"venue"`
	VenueEventID  string     `json:"venue_event_id"`
	Title         string     `json:"title"`
	Description   string     `json:"description"`
	Category      string     `json:"category"`
	Status        string     `json:"status"`
	CloseTime     *time.Time `json:"close_time,omitempty"`
}

type Market struct {
	ID             string     `json:"id"`
	Venue          string     `json:"venue"`
	VenueMarketID  string     `json:"venue_market_id"`
	EventID        string     `json:"event_id,omitempty"`
	VenueEventID   string     `json:"venue_event_id,omitempty"`
	Title          string     `json:"title"`
	Description    string     `json:"description"`
	Category       string     `json:"category"`
	Subcategory    string     `json:"subcategory"`
	MarketType     string     `json:"market_type"`
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

	CREATE TABLE IF NOT EXISTS events (
		id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		venue           VARCHAR(20) NOT NULL,
		venue_event_id  VARCHAR(255) NOT NULL,
		title           TEXT NOT NULL,
		description     TEXT,
		category        VARCHAR(100),
		status          VARCHAR(20) NOT NULL DEFAULT 'OPEN',
		close_time      TIMESTAMPTZ,
		first_seen_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		last_updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		UNIQUE (venue, venue_event_id)
	);
	CREATE INDEX IF NOT EXISTS idx_events_venue ON events(venue);

	CREATE TABLE IF NOT EXISTS markets (
		id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		venue           VARCHAR(20) NOT NULL,
		venue_market_id VARCHAR(255) NOT NULL,
		event_id        UUID REFERENCES events(id),
		venue_event_id  VARCHAR(255),
		title           TEXT NOT NULL,
		description     TEXT,
		category        VARCHAR(100),
		subcategory     VARCHAR(100),
		market_type     VARCHAR(50),
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
	CREATE INDEX IF NOT EXISTS idx_markets_event ON markets(event_id);
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

func (s *Store) UpsertEvent(ctx context.Context, e Event) (string, error) {
	sql := `
	INSERT INTO events (venue, venue_event_id, title, description, category, status, close_time)
	VALUES ($1, $2, $3, $4, $5, $6, $7)
	ON CONFLICT (venue, venue_event_id) DO UPDATE SET
		title = EXCLUDED.title,
		description = EXCLUDED.description,
		category = EXCLUDED.category,
		status = EXCLUDED.status,
		close_time = EXCLUDED.close_time,
		last_updated_at = NOW()
	RETURNING id`

	var id string
	err := s.pg.P().QueryRow(ctx, sql,
		e.Venue, e.VenueEventID, e.Title, e.Description, e.Category, e.Status, e.CloseTime,
	).Scan(&id)
	if err != nil {
		slog.Error("failed to upsert event", "error", err, "venue", e.Venue, "venue_event_id", e.VenueEventID)
		return "", fmt.Errorf("upsert event: %w", err)
	}
	return id, nil
}

func nullStr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func (s *Store) UpsertMarket(ctx context.Context, m Market) error {
	sql := `
	INSERT INTO markets (venue, venue_market_id, event_id, venue_event_id, title, description, category, subcategory, market_type, structure_type, status, resolution_date)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	ON CONFLICT (venue, venue_market_id) DO UPDATE SET
		event_id = COALESCE(EXCLUDED.event_id, markets.event_id),
		venue_event_id = COALESCE(EXCLUDED.venue_event_id, markets.venue_event_id),
		title = EXCLUDED.title,
		description = EXCLUDED.description,
		category = EXCLUDED.category,
		subcategory = EXCLUDED.subcategory,
		market_type = EXCLUDED.market_type,
		structure_type = EXCLUDED.structure_type,
		status = EXCLUDED.status,
		resolution_date = EXCLUDED.resolution_date,
		last_updated_at = NOW()`

	_, err := s.pg.P().Exec(ctx, sql,
		m.Venue, m.VenueMarketID, nullStr(m.EventID), nullStr(m.VenueEventID),
		m.Title, m.Description, m.Category, m.Subcategory, m.MarketType,
		m.StructureType, m.Status, m.ResolutionDate,
	)
	if err != nil {
		slog.Error("failed to upsert market", "error", err, "venue", m.Venue, "venue_market_id", m.VenueMarketID)
		return fmt.Errorf("upsert market: %w", err)
	}
	return nil
}

func (s *Store) GetUnembeddedMarkets(ctx context.Context, limit int) ([]Market, error) {
	sql := `
	SELECT id, venue, venue_market_id, COALESCE(event_id::text, ''), COALESCE(venue_event_id, ''), title, COALESCE(description, ''), COALESCE(category, ''), COALESCE(subcategory, ''), COALESCE(market_type, ''), structure_type, status, resolution_date
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
			&m.ID, &m.Venue, &m.VenueMarketID, &m.EventID, &m.VenueEventID,
			&m.Title, &m.Description,
			&m.Category, &m.Subcategory, &m.MarketType, &m.StructureType, &m.Status, &m.ResolutionDate,
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
	SELECT id, venue, venue_market_id, COALESCE(event_id::text, ''), COALESCE(venue_event_id, ''), title, COALESCE(description, ''), COALESCE(category, ''), COALESCE(subcategory, ''), COALESCE(market_type, ''), structure_type, status, resolution_date
	FROM markets
	WHERE embedding IS NOT NULL
	  AND status = 'OPEN'
	  AND (resolution_date IS NULL OR resolution_date > NOW())
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
			&m.ID, &m.Venue, &m.VenueMarketID, &m.EventID, &m.VenueEventID,
			&m.Title, &m.Description,
			&m.Category, &m.Subcategory, &m.MarketType, &m.StructureType, &m.Status, &m.ResolutionDate,
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
	tag, err := s.pg.P().Exec(ctx, "UPDATE markets SET embedding = $2::vector WHERE id = $1", id, vecStr)
	if err != nil {
		slog.Error("failed to upsert embedding", "error", err, "id", id)
		return fmt.Errorf("upsert embedding: %w", err)
	}
	if n := tag.RowsAffected(); n == 0 {
		slog.Warn("upsert embedding matched 0 rows", "id", id)
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

func (s *Store) GetPendingCandidatesWithMarkets(ctx context.Context, limit int, minSimilarity float64) ([]CandidateWithMarkets, error) {
	sql := `
	SELECT
		c.id, c.market_a_id, c.market_b_id, c.similarity, c.category, c.status,
		ma.id, ma.venue, ma.venue_market_id, COALESCE(ma.event_id::text, ''), COALESCE(ma.venue_event_id, ''), ma.title, COALESCE(ma.description, ''), COALESCE(ma.category, ''), COALESCE(ma.subcategory, ''), COALESCE(ma.market_type, ''), ma.structure_type, ma.status, ma.resolution_date,
		mb.id, mb.venue, mb.venue_market_id, COALESCE(mb.event_id::text, ''), COALESCE(mb.venue_event_id, ''), mb.title, COALESCE(mb.description, ''), COALESCE(mb.category, ''), COALESCE(mb.subcategory, ''), COALESCE(mb.market_type, ''), mb.structure_type, mb.status, mb.resolution_date
	FROM match_candidates c
	JOIN markets ma ON ma.id = c.market_a_id
	JOIN markets mb ON mb.id = c.market_b_id
	WHERE c.status = 'PENDING'
	  AND c.similarity >= $2
	LIMIT $1`

	rows, err := s.pg.P().Query(ctx, sql, limit, minSimilarity)
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
			&cwm.MarketA.ID, &cwm.MarketA.Venue, &cwm.MarketA.VenueMarketID, &cwm.MarketA.EventID, &cwm.MarketA.VenueEventID, &cwm.MarketA.Title, &cwm.MarketA.Description,
			&cwm.MarketA.Category, &cwm.MarketA.Subcategory, &cwm.MarketA.MarketType, &cwm.MarketA.StructureType, &cwm.MarketA.Status, &cwm.MarketA.ResolutionDate,
			&cwm.MarketB.ID, &cwm.MarketB.Venue, &cwm.MarketB.VenueMarketID, &cwm.MarketB.EventID, &cwm.MarketB.VenueEventID, &cwm.MarketB.Title, &cwm.MarketB.Description,
			&cwm.MarketB.Category, &cwm.MarketB.Subcategory, &cwm.MarketB.MarketType, &cwm.MarketB.StructureType, &cwm.MarketB.Status, &cwm.MarketB.ResolutionDate,
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

func (s *Store) SearchMarkets(ctx context.Context, venue, query string, limit int) ([]Market, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	rows, err := s.pg.P().Query(ctx, `
        SELECT id, venue, venue_market_id, COALESCE(event_id::text, ''), COALESCE(venue_event_id, ''), COALESCE(title,''), COALESCE(description,''), COALESCE(category,'')
        FROM markets
        WHERE venue = $1
          AND (title ILIKE '%' || $2 || '%' OR venue_market_id ILIKE '%' || $2 || '%')
          AND status = 'OPEN'
        ORDER BY
            CASE WHEN title ILIKE $2 || '%' THEN 0 ELSE 1 END,
            last_updated_at DESC
        LIMIT $3
    `, venue, query, limit)
	if err != nil {
		return nil, fmt.Errorf("search markets: %w", err)
	}
	defer rows.Close()

	var result []Market
	for rows.Next() {
		var m Market
		if err := rows.Scan(&m.ID, &m.Venue, &m.VenueMarketID, &m.EventID, &m.VenueEventID, &m.Title, &m.Description, &m.Category); err != nil {
			return nil, fmt.Errorf("search markets scan: %w", err)
		}
		result = append(result, m)
	}
	return result, rows.Err()
}

func (s *Store) GetMatchPairsByMarket(ctx context.Context, marketID string) ([]matchPairResponse, error) {
	rows, err := s.pg.P().Query(ctx, `
        SELECT
            mp.id,
            mp.candidate_id,
            ma.title AS market_a_title,
            mb.title AS market_b_title,
            ma.venue AS venue_a,
            mb.venue AS venue_b,
            COALESCE(mc.category, '') AS category,
            COALESCE(mp.is_same_event, false) AS is_same_event,
            COALESCE(mp.relationship, '') AS relationship,
            COALESCE(mp.confidence, 0) AS confidence,
            COALESCE(mp.reasoning, '') AS reasoning,
            COALESCE(mp.leg_a_model, '') AS leg_a_model,
            COALESCE(mp.leg_b_model, '') AS leg_b_model,
            mp.status
        FROM match_pairs mp
        JOIN match_candidates mc ON mc.id = mp.candidate_id
        JOIN markets ma ON ma.id = mc.market_a_id
        JOIN markets mb ON mb.id = mc.market_b_id
        WHERE mc.market_a_id = $1 OR mc.market_b_id = $1
        ORDER BY mp.reviewed_at DESC, mc.created_at DESC
    `, marketID)
	if err != nil {
		return nil, fmt.Errorf("get pairs by market: %w", err)
	}
	defer rows.Close()

	var result []matchPairResponse
	for rows.Next() {
		var p matchPairResponse
		if err := rows.Scan(
			&p.ID, &p.CandidateID, &p.MarketATitle, &p.MarketBTitle,
			&p.VenueA, &p.VenueB, &p.Category, &p.IsSameEvent,
			&p.Relationship, &p.Confidence, &p.Reasoning,
			&p.LegAModel, &p.LegBModel, &p.Status,
		); err != nil {
			return nil, fmt.Errorf("pairs by market scan: %w", err)
		}
		result = append(result, p)
	}
	return result, rows.Err()
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
	sql := `SELECT id, venue, venue_market_id, COALESCE(event_id::text, ''), COALESCE(venue_event_id, ''), title, description, category, structure_type, status, resolution_date FROM markets WHERE id = $1`
	var m Market
	err := s.pg.P().QueryRow(ctx, sql, id).Scan(
		&m.ID, &m.Venue, &m.VenueMarketID, &m.EventID, &m.VenueEventID, &m.Title, &m.Description,
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

type Stats struct {
	Unembedded           int `json:"unembedded"`
	Embedded             int `json:"embedded"`
	PendingCandidates    int `json:"pending_candidates"`
	ReviewedCandidates   int `json:"reviewed_candidates"`
	PairsPendingApproval int `json:"pairs_pending_approval"`
	PairsApproved        int `json:"pairs_approved"`
	PairsRejected        int `json:"pairs_rejected"`
}

type CategoryCount struct {
	Venue    string `json:"venue"`
	Category string `json:"category"`
	Count    int    `json:"count"`
	Embedded int    `json:"embedded"`
}

type PipelineCounts struct {
	Events  []CategoryCount `json:"events"`
	Markets []CategoryCount `json:"markets"`
}

func (s *Store) GetStats(ctx context.Context) (*Stats, error) {
	var st Stats
	err := s.pg.P().QueryRow(ctx, `
		SELECT
			(SELECT COUNT(*) FROM markets WHERE embedding IS NULL AND status = 'OPEN' AND COALESCE(title, '') <> '' AND description IS NOT NULL AND TRIM(description) <> ''),
			(SELECT COUNT(*) FROM markets WHERE embedding IS NOT NULL AND status = 'OPEN'),
			(SELECT COUNT(*) FROM match_candidates WHERE status = 'PENDING'),
			(SELECT COUNT(*) FROM match_candidates WHERE status = 'REVIEWED'),
			(SELECT COUNT(*) FROM match_pairs WHERE status = 'PENDING_APPROVAL'),
			(SELECT COUNT(*) FROM match_pairs WHERE status = 'APPROVED'),
			(SELECT COUNT(*) FROM match_pairs WHERE status = 'REJECTED')
	`).Scan(
		&st.Unembedded, &st.Embedded, &st.PendingCandidates, &st.ReviewedCandidates,
		&st.PairsPendingApproval, &st.PairsApproved, &st.PairsRejected,
	)
	if err != nil {
		return nil, fmt.Errorf("get stats: %w", err)
	}
	return &st, nil
}

func (s *Store) GetPipelineCounts(ctx context.Context) (*PipelineCounts, error) {
	pc := &PipelineCounts{
		Events:  make([]CategoryCount, 0),
		Markets: make([]CategoryCount, 0),
	}

	rows, err := s.pg.P().Query(ctx,
		`SELECT venue, COALESCE(category, 'Uncategorized'), COUNT(1) FROM events GROUP BY venue, category ORDER BY venue, category`)
	if err != nil {
		return nil, fmt.Errorf("get event counts: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var cc CategoryCount
		if err := rows.Scan(&cc.Venue, &cc.Category, &cc.Count); err != nil {
			return nil, err
		}
		pc.Events = append(pc.Events, cc)
	}

	rows2, err := s.pg.P().Query(ctx,
		`SELECT venue, COALESCE(category, 'Uncategorized'), COUNT(1), COUNT(1) FILTER (WHERE embedding IS NOT NULL) FROM markets WHERE status = 'OPEN' GROUP BY venue, category ORDER BY venue, category`)
	if err != nil {
		return nil, fmt.Errorf("get market counts: %w", err)
	}
	defer rows2.Close()
	for rows2.Next() {
		var cc CategoryCount
		if err := rows2.Scan(&cc.Venue, &cc.Category, &cc.Count, &cc.Embedded); err != nil {
			return nil, err
		}
		pc.Markets = append(pc.Markets, cc)
	}

	return pc, nil
}

type SimilarityPair struct {
	MarketAID  string  `json:"market_a_id"`
	MarketBID  string  `json:"market_b_id"`
	Similarity float64 `json:"similarity"`
	MarketA    Market  `json:"market_a"`
	MarketB    Market  `json:"market_b"`
}

func (s *Store) GetTopSimilarities(ctx context.Context, limit int) ([]SimilarityPair, error) {
	// Use a sampled approach: pick up to 500 Kalshi markets and find their best Polymarket match
	rows, err := s.pg.P().Query(ctx, `
		SELECT k.id, k.embedding FROM markets k
		WHERE k.venue='KALSHI' AND k.embedding IS NOT NULL
		  AND (k.resolution_date IS NULL OR k.resolution_date > NOW())
		ORDER BY k.id
		LIMIT 500`, limit)
	if err != nil {
		return nil, fmt.Errorf("sample kalshi: %w", err)
	}
	defer rows.Close()

	type sample struct {
		id  string
		emb []float64
	}
	var samples []sample
	for rows.Next() {
		var s sample
		if err := rows.Scan(&s.id, &s.emb); err != nil {
			return nil, err
		}
		samples = append(samples, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	var allPairs []SimilarityPair
	for _, s := range samples {
		rows2, err := s.pg.P().Query(ctx, `
			SELECT id, 1 - ($1 <=> embedding) AS sim
			FROM markets
			WHERE venue='POLYMARKET' AND embedding IS NOT NULL
			  AND (resolution_date IS NULL OR resolution_date > NOW())
			ORDER BY $1 <=> embedding
			LIMIT 3`, s.emb)
		if err != nil {
			continue
		}
		for rows2.Next() {
			var pid string
			var sim float64
			if err := rows2.Scan(&pid, &sim); err != nil {
				rows2.Close()
				break
			}
			allPairs = append(allPairs, SimilarityPair{MarketAID: s.id, MarketBID: pid, Similarity: sim})
		}
		rows2.Close()
	}

	// Sort by similarity and take top
	sort.Slice(allPairs, func(i, j int) bool { return allPairs[i].Similarity > allPairs[j].Similarity })
	if len(allPairs) > limit {
		allPairs = allPairs[:limit]
	}

	// Batch-load market details
	idSet := make(map[string]bool)
	for _, p := range allPairs {
		idSet[p.MarketAID] = true
		idSet[p.MarketBID] = true
	}
	ids := make([]string, 0, len(idSet))
	for id := range idSet {
		ids = append(ids, id)
	}
	markets, err := s.batchGetMarkets(ctx, ids)
	if err != nil {
		return nil, err
	}
	for i := range allPairs {
		allPairs[i].MarketA = markets[allPairs[i].MarketAID]
		allPairs[i].MarketB = markets[allPairs[i].MarketBID]
	}

	return allPairs, nil
}

func (s *Store) batchGetMarkets(ctx context.Context, ids []string) (map[string]Market, error) {
	if len(ids) == 0 {
		return map[string]Market{}, nil
	}
	rows, err := s.pg.P().Query(ctx, `
		SELECT id, venue, venue_market_id, COALESCE(event_id::text,''), COALESCE(venue_event_id,''),
		       title, COALESCE(description,''), COALESCE(category,''), COALESCE(subcategory,''),
		       COALESCE(market_type,''), structure_type, status, resolution_date
		FROM markets WHERE id = ANY($1)`, ids)
	if err != nil {
		return nil, fmt.Errorf("batch get markets: %w", err)
	}
	defer rows.Close()

	result := make(map[string]Market, len(ids))
	for rows.Next() {
		var m Market
		if err := rows.Scan(
			&m.ID, &m.Venue, &m.VenueMarketID, &m.EventID, &m.VenueEventID,
			&m.Title, &m.Description,
			&m.Category, &m.Subcategory, &m.MarketType, &m.StructureType, &m.Status, &m.ResolutionDate,
		); err != nil {
			return nil, fmt.Errorf("scan batch market: %w", err)
		}
		result[m.ID] = m
	}
	return result, rows.Err()
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
