-- ============================================================
-- V14: Events table, market_prices table, normalised markets.
--
-- Summary of changes:
--   • Add relationship + similarity_score to match_pairs (matching
--     service already inserts these; ADD IF NOT EXISTS is safe).
--   • Create events table (venue-agnostic event/series container).
--   • Wipe all transient market/matching data (clean slate).
--   • Drop legs, leg_prices, leg_mappings (replaced by token IDs
--     stored directly on markets + new market_prices table).
--   • Add event_id, venue_event_id, yes_token_id, no_token_id to
--     markets.
--   • Rebuild bundle_legs: drop leg_id FK, add market_id + leg_direction.
--   • Create market_prices (hot-write bid/ask keyed by market).
-- ============================================================

-- ============================================================
-- 1. Patch match_pairs: add relationship + similarity_score
--    (matching service inserts these; ADD IF NOT EXISTS is safe
--    in case they were added manually on the running DB).
-- ============================================================
ALTER TABLE match_pairs
    ADD COLUMN IF NOT EXISTS relationship     VARCHAR(30)
        CHECK (relationship IN ('EQUIVALENT', 'INVERSE', 'SUBSET', 'MUTUALLY_EXCLUSIVE')),
    ADD COLUMN IF NOT EXISTS similarity_score NUMERIC(6,4);

-- ============================================================
-- 2. Create events table.
-- ============================================================
CREATE TABLE IF NOT EXISTS events (
    id                  UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    venue               VARCHAR(20) NOT NULL REFERENCES venues(id),
    venue_event_id      VARCHAR(255) NOT NULL,
    title               TEXT        NOT NULL,
    description         TEXT,
    category            VARCHAR(100),
    status              VARCHAR(20) NOT NULL DEFAULT 'OPEN'
                        CHECK (status IN ('OPEN', 'CLOSED', 'RESOLVED')),
    mutually_exclusive  BOOLEAN     NOT NULL DEFAULT FALSE,
    close_time          TIMESTAMPTZ,
    first_seen_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (venue, venue_event_id)
);

CREATE INDEX IF NOT EXISTS idx_events_venue    ON events(venue);
CREATE INDEX IF NOT EXISTS idx_events_category ON events(category);
CREATE INDEX IF NOT EXISTS idx_events_status   ON events(status);

-- ============================================================
-- 3. Wipe all transient market/matching data.
--    CASCADE truncates: legs, leg_prices, bundle_legs,
--    match_pairs, match_candidates, leg_mappings, llm_validations.
-- ============================================================
TRUNCATE markets CASCADE;
TRUNCATE matching_runs CASCADE;

-- ============================================================
-- 4. Drop bundle_legs.leg_id FK before dropping legs.
-- ============================================================
ALTER TABLE bundle_legs
    DROP COLUMN IF EXISTS leg_id;

-- ============================================================
-- 5. Drop obsolete tables (safe now that bundle_legs FK is gone).
-- ============================================================
DROP TABLE IF EXISTS leg_prices;
DROP TABLE IF EXISTS leg_mappings;
DROP TABLE IF EXISTS legs;

-- ============================================================
-- 6. Add new columns to markets.
--    event_title already exists from V13 — kept as denorm cache.
-- ============================================================
ALTER TABLE markets
    ADD COLUMN IF NOT EXISTS event_id       UUID REFERENCES events(id),
    ADD COLUMN IF NOT EXISTS venue_event_id VARCHAR(255),
    ADD COLUMN IF NOT EXISTS yes_token_id   VARCHAR(255),
    ADD COLUMN IF NOT EXISTS no_token_id    VARCHAR(255);

CREATE INDEX IF NOT EXISTS idx_markets_event_id ON markets(event_id);

-- ============================================================
-- 7. Rebuild bundle_legs: add market_id + leg_direction in place
--    of the removed leg_id column.
-- ============================================================
ALTER TABLE bundle_legs
    ADD COLUMN IF NOT EXISTS market_id    UUID REFERENCES markets(id),
    ADD COLUMN IF NOT EXISTS leg_direction VARCHAR(3) NOT NULL DEFAULT 'YES'
        CHECK (leg_direction IN ('YES', 'NO'));

-- ============================================================
-- 8. Create market_prices (hot-write bid/ask keyed by market).
-- ============================================================
CREATE TABLE IF NOT EXISTS market_prices (
    market_id   UUID        PRIMARY KEY REFERENCES markets(id) ON DELETE CASCADE,
    yes_ask     NUMERIC(8,6),
    yes_bid     NUMERIC(8,6),
    no_ask      NUMERIC(8,6),
    no_bid      NUMERIC(8,6),
    last_price  NUMERIC(8,6),
    volume_24h  NUMERIC(14,2),
    recorded_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_market_prices_recorded ON market_prices(recorded_at DESC);
