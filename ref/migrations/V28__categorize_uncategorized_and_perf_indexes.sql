-- V28: Performance — stop re-scanning unmatched Kalshi markets and add
-- partial indexes that make the matching service's per-cycle queries cheap.
--
-- Context: the matching service runs a regex CASE sweep on every embedding
-- cycle to assign categories to Kalshi markets auto-registered from tick
-- events. With ~52K rows that match no regex branch, the sweep was rewriting
-- NULL→NULL every 5 minutes and pegging Postgres. We now stamp them
-- 'uncategorized' so the WHERE category IS NULL filter excludes them on
-- subsequent runs, and provide a view to surface them for manual review and
-- regex-rule expansion.
--
-- Also adds partial indexes used by the embedding-need scan and the auto-
-- categorize sweep, both of which were doing seq scans on the 530K-row
-- markets table.

-- 1) Backfill: anything currently NULL on Kalshi gets the same regex
--    classification the runtime sweep uses, with ELSE 'uncategorized'.
UPDATE markets
SET category = CASE
    WHEN venue_market_id ~* '^KX(BTC|ETH|SOL|XRP|DOGE)[^D]*(15M|1H|2H|4H)' THEN 'crypto_intraday'
    WHEN venue_market_id ~* '^KX(BTCD|ETHD|ETHHD|SOLD|DOGE|XRP|BNB|BTC|ETH|SOL)' THEN 'crypto_daily'
    WHEN venue_market_id ~* '^KX(MLB|NBA|NHL|NFLGAME|NFLTOTAL|MLS|NWSL|AFL|EKSTRAKLASA|LIGAMX|KBO|ATPMATCH|WTA|UCLGAME|EPLGAME)' THEN 'sports'
    WHEN venue_market_id ~* '^KX(NFLWINS|NFLDIV|NFLCHAMP|PGATOUR|PGATOP|PGAR|PGACHAMP|UFC|NASCAR|MVESPORTS|MVECROSS|MLBWINS|NBACHAMP|NHLCHAMP)' THEN 'sports'
    WHEN venue_market_id ~* '^KX(GOV|ELEC|PRES|SEC|SEN|HOUSE|PARTY|VOTE|POLL|CONG|DEMS|REP|MAGA|POTUS)' THEN 'politics'
    WHEN venue_market_id ~* '^KX(WAR|NATO|GEOPO)' THEN 'world_events'
    WHEN venue_market_id ~* '^KX(GOLDD|GOLD|OIL|GAS|CPI|GDP|FED|RATE|UNEMPLOY|INFL)' THEN 'economics'
    WHEN venue_market_id ~* '^KX(SPACE|SPACEX|FDA|AI|TECH)' THEN 'science_tech'
    ELSE 'uncategorized'
END
WHERE venue = 'KALSHI' AND category IS NULL;

-- 2) Review view: surfaces uncategorized markets so we can spot common
--    ticker prefixes and add new regex branches in
--    services/matching/host.py and this migration's runtime equivalent.
--    Sample query:
--      SELECT prefix, n FROM v_uncategorized_kalshi_prefixes ORDER BY n DESC LIMIT 30;
CREATE OR REPLACE VIEW v_uncategorized_kalshi_prefixes AS
SELECT
    -- Strip the leading "KX" and take everything up to the first digit or hyphen
    -- so we get a clean grouping by "ticker family" (e.g. KXAIART2026 -> AIART).
    regexp_replace(venue_market_id, '^KX', '')                          AS body,
    regexp_replace(regexp_replace(venue_market_id, '^KX', ''),
                   '[0-9\-].*$', '')                                    AS prefix,
    COUNT(*)                                                            AS n,
    MIN(title)                                                          AS example_title,
    MIN(venue_market_id)                                                AS example_ticker
FROM markets
WHERE venue = 'KALSHI' AND category = 'uncategorized'
GROUP BY 1, 2;

CREATE OR REPLACE VIEW v_uncategorized_markets AS
SELECT id, venue, venue_market_id, title, description, resolution_date, status
FROM markets
WHERE category = 'uncategorized'
ORDER BY venue, venue_market_id;

-- 3) Partial index for the auto-categorize sweep: now matches only the
--    (small + shrinking) set of NULL-category Kalshi rows.
CREATE INDEX IF NOT EXISTS idx_markets_kalshi_null_category
    ON markets (venue_market_id)
    WHERE venue = 'KALSHI' AND category IS NULL;

-- 4) Partial index for the matching service's "needs embedding" scan
--    (_fetch_markets_needing_embedding). Mirrors that query's stable
--    predicates; the volatile resolution_date filter is left to a recheck.
CREATE INDEX IF NOT EXISTS idx_markets_needs_embedding
    ON markets (category, venue_market_id)
    WHERE embedding IS NULL
      AND status = 'OPEN'
      AND category IS NOT NULL
      AND category <> 'uncategorized';

-- Refresh planner stats so the new indexes actually get picked up.
ANALYZE markets;
