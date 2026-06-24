-- Indexes to support NOT EXISTS pair-deduplication in matching review stage.
-- These allow the candidate SQL query to skip already-validated/existing pairs
-- at the DB level rather than loading them into Python for per-row filtering.

CREATE INDEX IF NOT EXISTS idx_llm_validations_pair
    ON llm_validations(market_a_id, market_b_id);

CREATE INDEX IF NOT EXISTS idx_llm_validations_pair_rev
    ON llm_validations(market_b_id, market_a_id);

CREATE INDEX IF NOT EXISTS idx_match_pairs_ab
    ON match_pairs(market_a_id, market_b_id);

CREATE INDEX IF NOT EXISTS idx_match_pairs_ba
    ON match_pairs(market_b_id, market_a_id);
