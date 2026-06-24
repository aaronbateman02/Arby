-- ============================================================
-- PolyBot — Tune Market Search trigram indexes
-- Migration: V22
-- Date: 2026-05-29
--
-- Drop the description trigram index: broad queries like "bitcoin"
-- match ~80k rows on description, so the planner ignores the GIN
-- index and falls back to a sequential scan anyway. Keeping the
-- index just slows down every market write.
--
-- Add event_title to the indexed search columns.
-- ============================================================

DROP INDEX IF EXISTS idx_markets_description_trgm;

CREATE INDEX IF NOT EXISTS idx_markets_event_title_trgm
    ON markets USING gin (event_title gin_trgm_ops)
    WHERE event_title IS NOT NULL;
