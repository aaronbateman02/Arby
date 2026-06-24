-- ============================================================
-- PolyBot — Candidates Queued
-- Migration: 007
-- Changes:
--   1. Add candidates_queued to matching_runs — number of candidates
--      that passed pre-filters and were sent to the LLM for review.
--      Set before asyncio.gather so a RUNNING run shows live progress.
-- ============================================================

ALTER TABLE matching_runs ADD COLUMN IF NOT EXISTS candidates_queued INTEGER;
