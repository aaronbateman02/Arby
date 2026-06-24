-- ============================================================
-- PolyBot — Trigram indexes for fast Market Search
-- Migration: V21
-- Date: 2026-05-29
--
-- The Market Search UI runs `ILIKE '%q%'` across several text
-- columns on the `markets` table. With ~300k rows this performs
-- a full sequential scan + regex per row, taking many seconds
-- and frequently tripping the nginx 60s timeout (502 Bad Gateway).
--
-- pg_trgm + GIN trigram indexes allow ILIKE '%...%' to use the
-- index, dropping latency to tens of milliseconds.
-- ============================================================

-- Indexes are built non-concurrently so they participate in
-- Flyway's per-migration transaction. This briefly blocks writes
-- to the markets table (~30-90s on ~300k rows), which is
-- acceptable for a one-time migration during a deploy window.

CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE INDEX IF NOT EXISTS idx_markets_title_trgm
    ON markets USING gin (title gin_trgm_ops);

CREATE INDEX IF NOT EXISTS idx_markets_venue_market_id_trgm
    ON markets USING gin (venue_market_id gin_trgm_ops);

CREATE INDEX IF NOT EXISTS idx_markets_slug_trgm
    ON markets USING gin (slug gin_trgm_ops)
    WHERE slug IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_markets_event_slug_trgm
    ON markets USING gin (event_slug gin_trgm_ops)
    WHERE event_slug IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_markets_description_trgm
    ON markets USING gin (description gin_trgm_ops)
    WHERE description IS NOT NULL;
