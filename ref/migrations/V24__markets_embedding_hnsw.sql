-- ============================================================
-- V24__markets_embedding_hnsw.sql
-- Replace the markets.embedding IVFFlat index with an HNSW index.
--
-- Why:
--   pgvector >= 0.5 ships HNSW which gives both higher recall and lower
--   query cost than IVFFlat at the same recall target. pg_stat_statements
--   showed the cross-venue ANN search accounting for ~75% of remaining
--   postgres CPU on the 555K-row markets table; HNSW is the standard
--   remedy.
--
-- Notes:
--   * idx_markets_embedding_hnsw was built CONCURRENTLY ahead of this
--     migration to avoid blocking writes. This file is recorded so the
--     schema is reproducible from migrations alone.
--   * Defaults (m=16, ef_construction=64) are sufficient for 1024-dim
--     bge-large vectors at the current row count.
--   * Application code uses `SET LOCAL hnsw.ef_search = 40` (pgvector
--     default) in services/matching/host.py.
-- ============================================================

CREATE INDEX IF NOT EXISTS idx_markets_embedding_hnsw
    ON markets USING hnsw (embedding vector_cosine_ops);

DROP INDEX IF EXISTS idx_markets_embedding;
