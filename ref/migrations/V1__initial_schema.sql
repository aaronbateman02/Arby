-- ============================================================
-- PolyBot — Initial Schema
-- Migration: 001
-- Date: 2026-05-15
-- ============================================================

-- Prerequisites
CREATE EXTENSION IF NOT EXISTS "pgcrypto";    -- gen_random_uuid()
CREATE EXTENSION IF NOT EXISTS "vector";      -- pgvector for embeddings

-- ============================================================
-- VENUES
-- Static reference table. Seeded in 002_seed_data.sql.
-- ============================================================
CREATE TABLE venues (
    id              VARCHAR(20)     PRIMARY KEY,          -- 'KALSHI', 'POLYMARKET'
    display_name    VARCHAR(100)    NOT NULL,
    base_url        VARCHAR(255)    NOT NULL,
    fee_rate        NUMERIC(6,4)    NOT NULL DEFAULT 0,   -- fraction e.g. 0.02 = 2%
    min_order_shares INTEGER        NOT NULL DEFAULT 1,
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

-- ============================================================
-- MARKETS
-- One row per market per venue. Updated on each snapshot event.
-- ============================================================
CREATE TABLE markets (
    id                  UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    venue               VARCHAR(20) NOT NULL REFERENCES venues(id),
    venue_market_id     VARCHAR(255) NOT NULL,
    title               TEXT        NOT NULL,
    description         TEXT,
    category            VARCHAR(100),
    structure_type      VARCHAR(20) NOT NULL CHECK (structure_type IN ('BINARY', 'MULTI_CHOICE')),
    status              VARCHAR(20) NOT NULL CHECK (status IN ('OPEN', 'CLOSED', 'RESOLVED')),
    resolution_date     TIMESTAMPTZ,
    -- pgvector: text-embedding-3-small = 1536 dimensions
    embedding           VECTOR(1536),
    embedding_model     VARCHAR(100),                     -- model used to generate embedding
    embedding_updated_at TIMESTAMPTZ,
    first_seen_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (venue, venue_market_id)
);

CREATE INDEX idx_markets_venue        ON markets(venue);
CREATE INDEX idx_markets_status       ON markets(status);
CREATE INDEX idx_markets_category     ON markets(category);
-- IVFFlat index for approximate nearest-neighbour embedding search
CREATE INDEX idx_markets_embedding    ON markets USING ivfflat (embedding vector_cosine_ops)
    WITH (lists = 100);

-- ============================================================
-- LEGS
-- One row per outcome/leg within a market.
-- ============================================================
CREATE TABLE legs (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    market_id       UUID        NOT NULL REFERENCES markets(id) ON DELETE CASCADE,
    venue_leg_id    VARCHAR(255) NOT NULL,
    label           TEXT        NOT NULL,
    direction       VARCHAR(20) NOT NULL CHECK (direction IN ('YES', 'NO', 'OPTION')),
    is_winning      BOOLEAN,                              -- NULL until market resolves
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (market_id, venue_leg_id)
);

CREATE INDEX idx_legs_market_id ON legs(market_id);

-- ============================================================
-- LEG PRICES
-- Latest known ask/bid per leg. Updated on every tick event.
-- Also mirrored in Redis for hot-path reads.
-- ============================================================
CREATE TABLE leg_prices (
    leg_id          UUID        PRIMARY KEY REFERENCES legs(id) ON DELETE CASCADE,
    ask             NUMERIC(8,4) NOT NULL,
    bid             NUMERIC(8,4) NOT NULL,
    volume_24h      NUMERIC(14,2),
    recorded_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================
-- MATCH PAIRS
-- Proposed/approved/rejected market pairs from the Matching Engine.
-- ============================================================
CREATE TABLE match_pairs (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    sub_matcher_id  VARCHAR(100) NOT NULL,
    status          VARCHAR(20) NOT NULL CHECK (status IN ('PENDING', 'APPROVED', 'REJECTED'))
                    DEFAULT 'PENDING',
    confidence      NUMERIC(4,3) NOT NULL CHECK (confidence BETWEEN 0 AND 1),
    reasoning       TEXT        NOT NULL,               -- LLM explanation verbatim
    market_a_id     UUID        NOT NULL REFERENCES markets(id),
    market_b_id     UUID        NOT NULL REFERENCES markets(id),
    reviewed_at     TIMESTAMPTZ,
    reviewed_by     VARCHAR(100),                       -- 'AUTO' or operator id
    rejection_reason TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (market_a_id <> market_b_id)
);

CREATE INDEX idx_match_pairs_status         ON match_pairs(status);
CREATE INDEX idx_match_pairs_sub_matcher    ON match_pairs(sub_matcher_id);
CREATE INDEX idx_match_pairs_market_a       ON match_pairs(market_a_id);
CREATE INDEX idx_match_pairs_market_b       ON match_pairs(market_b_id);

-- ============================================================
-- LEG MAPPINGS
-- How individual legs map across a matched pair.
-- ============================================================
CREATE TABLE leg_mappings (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    match_pair_id   UUID        NOT NULL REFERENCES match_pairs(id) ON DELETE CASCADE,
    leg_a_id        UUID        NOT NULL REFERENCES legs(id),
    leg_b_id        UUID        NOT NULL REFERENCES legs(id),
    relationship    VARCHAR(30) NOT NULL
                    CHECK (relationship IN ('EQUIVALENT', 'INVERSE', 'SUBSET', 'MUTUALLY_EXCLUSIVE')),
    confidence      NUMERIC(4,3) NOT NULL CHECK (confidence BETWEEN 0 AND 1),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_leg_mappings_pair  ON leg_mappings(match_pair_id);
CREATE INDEX idx_leg_mappings_leg_a ON leg_mappings(leg_a_id);
CREATE INDEX idx_leg_mappings_leg_b ON leg_mappings(leg_b_id);

-- ============================================================
-- STRATEGY CONFIGS (versioned)
-- Current config = highest config_version per strategy_id.
-- ============================================================
CREATE TABLE strategy_configs (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    strategy_id     VARCHAR(100) NOT NULL,
    config_version  INTEGER     NOT NULL,
    config          JSONB       NOT NULL,               -- full StrategyConfig JSON
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by      VARCHAR(100),
    UNIQUE (strategy_id, config_version)
);

CREATE INDEX idx_strategy_configs_strategy_id ON strategy_configs(strategy_id);

-- Convenience view: current config per strategy
CREATE VIEW strategy_current_configs AS
    SELECT DISTINCT ON (strategy_id)
        strategy_id,
        config_version,
        config,
        created_at
    FROM strategy_configs
    ORDER BY strategy_id, config_version DESC;

-- ============================================================
-- STRATEGY RISK STATE
-- Active/paused state per strategy. Owned by Risk Service.
-- ============================================================
CREATE TABLE strategy_risk_state (
    strategy_id     VARCHAR(100) PRIMARY KEY,
    status          VARCHAR(20) NOT NULL CHECK (status IN ('ACTIVE', 'PAUSED')) DEFAULT 'ACTIVE',
    paused_at       TIMESTAMPTZ,
    paused_reason   TEXT,
    -- latest metric value that triggered evaluation
    last_metric_value NUMERIC(12,6),
    last_evaluated_at TIMESTAMPTZ,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================
-- STRATEGY METRIC HISTORY
-- Rolling window values tracked by Risk Service.
-- ============================================================
CREATE TABLE strategy_metric_snapshots (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    strategy_id     VARCHAR(100) NOT NULL,
    metric          VARCHAR(50) NOT NULL,
    value           NUMERIC(12,6) NOT NULL,
    recorded_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_metric_snapshots_strategy ON strategy_metric_snapshots(strategy_id, recorded_at DESC);

-- ============================================================
-- CAPITAL RESERVATIONS
-- Tracks capital reserved per bundle to prevent double-spend.
-- ============================================================
CREATE TABLE capital_reservations (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    bundle_id       UUID        NOT NULL,               -- FK added after bundles table
    amount          NUMERIC(12,4) NOT NULL,
    status          VARCHAR(20) NOT NULL
                    CHECK (status IN ('ACTIVE', 'RELEASED', 'CONVERTED')) DEFAULT 'ACTIVE',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    released_at     TIMESTAMPTZ
);

CREATE INDEX idx_capital_reservations_status ON capital_reservations(status);

-- ============================================================
-- BUNDLES
-- One row per opportunity bundle from creation through resolution.
-- ============================================================
CREATE TABLE bundles (
    id                      UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    opportunity_id          UUID        NOT NULL,       -- ID from strategy's opportunity event
    strategy_id             VARCHAR(100) NOT NULL,
    match_pair_id           UUID        REFERENCES match_pairs(id),
    state                   VARCHAR(30) NOT NULL
                            CHECK (state IN (
                                'PENDING', 'EXECUTING', 'PENDING_COMPLETION',
                                'COMPLETE', 'UNWINDING', 'ABORTED'
                            )) DEFAULT 'PENDING',
    paper                   BOOLEAN     NOT NULL DEFAULT FALSE,
    early_exit_eligible     BOOLEAN     NOT NULL DEFAULT TRUE,
    min_roi                 NUMERIC(6,4) NOT NULL,
    partial_fill_policy     VARCHAR(30) NOT NULL
                            CHECK (partial_fill_policy IN ('UNWIND_ON_PARTIAL', 'HOLD_AND_GTC')),
    gtc_ttl_ms              INTEGER,
    gtc_ttl_expiry_action   VARCHAR(20) CHECK (gtc_ttl_expiry_action IN ('UNWIND', 'HOLD')),
    -- financials (populated as legs fill)
    estimated_cost          NUMERIC(12,4),              -- at time of ROI gate pass
    actual_cost             NUMERIC(12,4),              -- sum of fill prices
    actual_fees             NUMERIC(12,4),
    actual_payout           NUMERIC(12,4),              -- set on resolution
    net_roi                 NUMERIC(8,4),               -- (payout - cost - fees) / cost
    capital_reservation_id  UUID        REFERENCES capital_reservations(id),
    -- timestamps
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at            TIMESTAMPTZ,
    aborted_at              TIMESTAMPTZ,
    aborted_reason          TEXT
);

CREATE INDEX idx_bundles_strategy_id    ON bundles(strategy_id);
CREATE INDEX idx_bundles_state          ON bundles(state);
CREATE INDEX idx_bundles_paper          ON bundles(paper);
CREATE INDEX idx_bundles_match_pair_id  ON bundles(match_pair_id);
CREATE INDEX idx_bundles_created_at     ON bundles(created_at DESC);

-- Now we can add the FK from capital_reservations
ALTER TABLE capital_reservations
    ADD CONSTRAINT fk_capital_reservations_bundle
    FOREIGN KEY (bundle_id) REFERENCES bundles(id);

-- ============================================================
-- BUNDLE LEGS
-- The individual legs that make up a bundle.
-- ============================================================
CREATE TABLE bundle_legs (
    id                  UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    bundle_id           UUID        NOT NULL REFERENCES bundles(id) ON DELETE CASCADE,
    leg_id              UUID        NOT NULL REFERENCES legs(id),
    direction           VARCHAR(10) NOT NULL DEFAULT 'BUY',
    shares              INTEGER     NOT NULL CHECK (shares > 0),
    target_max_ask      NUMERIC(8,4),                   -- optional per-leg price ceiling
    fill_price          NUMERIC(8,4),
    fees_paid           NUMERIC(8,4),
    fill_timestamp      TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_bundle_legs_bundle_id ON bundle_legs(bundle_id);
CREATE INDEX idx_bundle_legs_leg_id    ON bundle_legs(leg_id);

-- ============================================================
-- ORDERS
-- One row per order placed on a venue.
-- ============================================================
CREATE TABLE orders (
    id                  UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    bundle_id           UUID        NOT NULL REFERENCES bundles(id),
    bundle_leg_id       UUID        NOT NULL REFERENCES bundle_legs(id),
    venue               VARCHAR(20) NOT NULL REFERENCES venues(id),
    venue_order_id      VARCHAR(255),                   -- null until placed
    order_type          VARCHAR(10) NOT NULL CHECK (order_type IN ('MARKET', 'LIMIT', 'GTC')),
    shares              INTEGER     NOT NULL CHECK (shares > 0),
    ask_at_placement    NUMERIC(8,4),
    shares_filled       INTEGER     NOT NULL DEFAULT 0,
    fill_price          NUMERIC(8,4),
    fees_paid           NUMERIC(8,4),
    status              VARCHAR(25) NOT NULL
                        CHECK (status IN (
                            'PENDING', 'PLACED', 'PARTIALLY_FILLED',
                            'FILLED', 'CANCELLED', 'FAILED'
                        )) DEFAULT 'PENDING',
    failure_reason      TEXT,
    placed_at           TIMESTAMPTZ,
    filled_at           TIMESTAMPTZ,
    cancelled_at        TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_orders_bundle_id       ON orders(bundle_id);
CREATE INDEX idx_orders_venue_order_id  ON orders(venue_order_id) WHERE venue_order_id IS NOT NULL;
CREATE INDEX idx_orders_status          ON orders(status);

-- ============================================================
-- BUNDLE STATE HISTORY (audit log)
-- Every state transition for every bundle. Never deleted.
-- ============================================================
CREATE TABLE bundle_state_history (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    bundle_id       UUID        NOT NULL REFERENCES bundles(id),
    previous_state  VARCHAR(30),
    new_state       VARCHAR(30) NOT NULL,
    reason          TEXT,
    source          VARCHAR(50) NOT NULL DEFAULT 'runtime'
                    CHECK (source IN ('runtime', 'crash_recovery', 'operator')),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_bundle_state_history_bundle ON bundle_state_history(bundle_id, created_at DESC);

-- ============================================================
-- EXECUTION METRICS
-- Latency timestamps for the pipeline from signal to fill.
-- Used by UI execution latency panel.
-- ============================================================
CREATE TABLE execution_metrics (
    id                      UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    bundle_id               UUID        NOT NULL REFERENCES bundles(id),
    strategy_id             VARCHAR(100) NOT NULL,
    -- pipeline stage timestamps
    opportunity_emitted_at  TIMESTAMPTZ,   -- when strategy published candidate
    opportunity_received_at TIMESTAMPTZ,   -- when Opportunity Service received it
    roi_gate_passed_at      TIMESTAMPTZ,   -- when ROI threshold was met
    capital_reserved_at     TIMESTAMPTZ,   -- when Risk Service approved reservation
    first_order_placed_at   TIMESTAMPTZ,   -- when first venue order was submitted
    last_order_filled_at    TIMESTAMPTZ,   -- when all legs confirmed filled
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_exec_metrics_strategy ON execution_metrics(strategy_id, created_at DESC);

-- ============================================================
-- ALERTS (in-app alert log)
-- ============================================================
CREATE TABLE alerts (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    severity        VARCHAR(10) NOT NULL CHECK (severity IN ('INFO', 'WARN', 'ERROR', 'CRITICAL')),
    source_service  VARCHAR(100) NOT NULL,
    message         TEXT        NOT NULL,
    metadata        JSONB,
    acknowledged    BOOLEAN     NOT NULL DEFAULT FALSE,
    acknowledged_at TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_alerts_severity       ON alerts(severity);
CREATE INDEX idx_alerts_acknowledged   ON alerts(acknowledged);
CREATE INDEX idx_alerts_created_at     ON alerts(created_at DESC);

-- ============================================================
-- PLATFORM CONFIG
-- Global settings managed by operator via UI.
-- ============================================================
CREATE TABLE platform_config (
    key             VARCHAR(100) PRIMARY KEY,
    value           JSONB        NOT NULL,
    description     TEXT,
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_by      VARCHAR(100)
);

-- ============================================================
-- SUB-MATCHER CONFIG
-- Per-sub-matcher settings including the human review toggle.
-- ============================================================
CREATE TABLE sub_matcher_configs (
    sub_matcher_id          VARCHAR(100) PRIMARY KEY,
    display_name            VARCHAR(200) NOT NULL,
    enabled                 BOOLEAN      NOT NULL DEFAULT TRUE,
    human_review_required   BOOLEAN      NOT NULL DEFAULT TRUE,   -- toggle off when accuracy proven
    auto_approve_threshold  NUMERIC(4,3) NOT NULL DEFAULT 0.95    -- min confidence for auto-approve
                            CHECK (auto_approve_threshold BETWEEN 0 AND 1),
    similarity_threshold    NUMERIC(4,3) NOT NULL DEFAULT 0.75    -- min cosine sim for candidate pairs
                            CHECK (similarity_threshold BETWEEN 0 AND 1),
    created_at              TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
