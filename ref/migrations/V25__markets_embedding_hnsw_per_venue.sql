-- V25: Replace global HNSW with per-venue partial HNSW indexes.
--
-- Why: the matching ANN query filters by venue, status, resolution_date and
-- a similarity threshold. With a single global HNSW over all 555k markets,
-- the planner ends up preferring a venue btree scan (~25k POLY rows) and
-- brute-forcing the cosine distance for every row (1.2s per query). By
-- partitioning HNSW per-venue, the index serves ~10-25k rows each and the
-- planner can use it directly for ORDER BY embedding <=> $vec LIMIT k.
--
-- Side benefits:
--   * Total HNSW pages on disk shrink (two small indexes < one giant one).
--   * Writes that change embedding only update the matching venue's index.
--   * Status filter is left out of the predicate on purpose so that markets
--     closing/reopening do not require an index rebuild; we post-filter.
--
-- Note: indexes must be created CONCURRENTLY to avoid blocking writes. This
-- file is not transactional; run statements individually.

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_markets_emb_hnsw_poly
    ON markets USING hnsw (embedding vector_cosine_ops)
    WHERE venue = 'POLYMARKET' AND embedding IS NOT NULL;

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_markets_emb_hnsw_kalshi
    ON markets USING hnsw (embedding vector_cosine_ops)
    WHERE venue = 'KALSHI' AND embedding IS NOT NULL;

-- Drop the old global HNSW once partials are valid. Done in a separate file
-- (V25b) so it can be run after CONCURRENTLY builds complete and after we
-- verify the partials are being picked up by EXPLAIN.
