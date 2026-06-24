-- Materialised ANN candidate cache populated by the matching service.
-- Replaces the expensive CROSS JOIN LATERAL pgvector scan in the reporting API.
CREATE TABLE IF NOT EXISTS match_candidates (
    kalshi_id      uuid    NOT NULL,
    polymarket_id  uuid    NOT NULL,
    similarity     real    NOT NULL,
    category       varchar(64),
    last_seen_at   timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (kalshi_id, polymarket_id)
);

CREATE INDEX IF NOT EXISTS idx_match_candidates_sim
    ON match_candidates (similarity DESC);

CREATE INDEX IF NOT EXISTS idx_match_candidates_cat_sim
    ON match_candidates (category, similarity DESC);
