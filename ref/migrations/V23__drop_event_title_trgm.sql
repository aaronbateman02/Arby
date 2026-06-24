-- ============================================================
-- PolyBot — Drop event_title trigram index
-- Migration: V23
-- Date: 2026-05-29
--
-- event_title is shared across many markets in the same event,
-- so common terms (e.g. "bitcoin") match tens of thousands of
-- rows and the planner falls back to a sequential scan anyway.
-- Removing this column from Market Search and dropping the
-- unused index keeps writes cheap.
-- ============================================================

DROP INDEX IF EXISTS idx_markets_event_title_trgm;
