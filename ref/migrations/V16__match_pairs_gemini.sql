-- V16: Add Gemini second-opinion columns to match_pairs
-- These store Round 2 (Gemini) data from the two-round approval system.
ALTER TABLE match_pairs
    ADD COLUMN IF NOT EXISTS gemini_reasoning  TEXT,
    ADD COLUMN IF NOT EXISTS gemini_confidence NUMERIC,
    ADD COLUMN IF NOT EXISTS gemini_model      VARCHAR(100);
