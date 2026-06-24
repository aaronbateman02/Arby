-- ============================================================
-- PolyBot — Matching v2
-- Migration: 003
-- Changes:
--   1. category_configs  — per-category ingest/pair/strategy toggles
--   2. matching_runs     — audit log for each scheduled/manual run
--   3. llm_validations   — per-pair LLM call record for cost tracking
--   4. markets.embedding — resize from VECTOR(1536) to VECTOR(1024)
--                          (bge-large-en-v1.5 produces 1024-dim vectors)
-- ============================================================

-- ============================================================
-- 1. CATEGORY CONFIGS
-- ============================================================
CREATE TABLE category_configs (
    category        VARCHAR(100)    PRIMARY KEY,
    display_name    VARCHAR(150)    NOT NULL,
    ingest_enabled  BOOLEAN         NOT NULL DEFAULT TRUE,
    pair_enabled    BOOLEAN         NOT NULL DEFAULT TRUE,
    strategy_active BOOLEAN         NOT NULL DEFAULT FALSE,
    notes           TEXT,
    updated_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

-- Seed with the agreed category matrix
INSERT INTO category_configs (category, display_name, ingest_enabled, pair_enabled, strategy_active, notes) VALUES
    ('politics',         'Politics / Elections',          TRUE,  TRUE,  FALSE, NULL),
    ('world_events',     'World Events',                  TRUE,  TRUE,  FALSE, NULL),
    ('sports_team',      'Sports — Team Win/Loss',        TRUE,  TRUE,  FALSE, NULL),
    ('sports_spread',    'Sports — Point Spreads',        TRUE,  TRUE,  FALSE, 'Numeric spread value must match exactly'),
    ('sports_player',    'Sports — Player Stats',         TRUE,  FALSE, FALSE, 'Manual pair review only'),
    ('economics',        'Economics',                     TRUE,  FALSE, FALSE, 'Manual pair review only'),
    ('crypto_daily',     'Crypto Prices (daily)',         TRUE,  FALSE, FALSE, 'No pairing — directional only'),
    ('crypto_intraday',  'Crypto / Financial Intra-day',  FALSE, FALSE, FALSE, 'Excluded entirely'),
    ('climate',          'Climate / Weather',             TRUE,  FALSE, FALSE, NULL),
    ('entertainment',    'Entertainment',                 TRUE,  FALSE, FALSE, 'Manual pair review only'),
    ('science_tech',     'Science / Tech Milestones',     TRUE,  TRUE,  FALSE, NULL)
ON CONFLICT (category) DO NOTHING;


-- ============================================================
-- 2. MATCHING RUNS
-- ============================================================
CREATE TABLE matching_runs (
    id                  UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    trigger             VARCHAR(20) NOT NULL CHECK (trigger IN ('SCHEDULED', 'MANUAL')),
    started_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    finished_at         TIMESTAMPTZ,
    status              VARCHAR(20) NOT NULL
                        CHECK (status IN ('RUNNING', 'COMPLETED', 'FAILED'))
                        DEFAULT 'RUNNING',
    markets_embedded    INTEGER,
    candidates_found    INTEGER,
    pairs_proposed      INTEGER,
    error_msg           TEXT
);

CREATE INDEX idx_matching_runs_started ON matching_runs(started_at DESC);


-- ============================================================
-- 3. LLM VALIDATIONS
-- Per-pair GPT call record — for cost tracking and debugging.
-- ============================================================
CREATE TABLE llm_validations (
    id                  UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    run_id              UUID        REFERENCES matching_runs(id) ON DELETE SET NULL,
    market_a_id         UUID        NOT NULL REFERENCES markets(id),
    market_b_id         UUID        NOT NULL REFERENCES markets(id),
    similarity_score    NUMERIC(6,4),
    model               VARCHAR(100) NOT NULL,
    prompt_tokens       INTEGER,
    completion_tokens   INTEGER,
    result_json         JSONB,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_llm_validations_run    ON llm_validations(run_id);
CREATE INDEX idx_llm_validations_pair   ON llm_validations(market_a_id, market_b_id);


-- ============================================================
-- 4. RESIZE markets.embedding: 1536 → 1024
--    The IVFFlat index must be dropped first.
-- ============================================================
DROP INDEX IF EXISTS idx_markets_embedding;

-- Wipe any existing embeddings — they were generated with a different
-- model (text-embedding-3-small, 1536-dim) and are incompatible.
UPDATE markets SET
    embedding             = NULL,
    embedding_model       = NULL,
    embedding_updated_at  = NULL;

-- Change column type to 1024 dimensions (bge-large-en-v1.5).
ALTER TABLE markets ALTER COLUMN embedding TYPE VECTOR(1024);

-- Recreate the IVFFlat index for approximate nearest-neighbour search.
-- lists=100 is appropriate for up to ~1M rows; tune upward if needed.
CREATE INDEX idx_markets_embedding ON markets
    USING ivfflat (embedding vector_cosine_ops)
    WITH (lists = 100);
