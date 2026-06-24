-- V6: Assign categories to Kalshi markets based on ticker prefix patterns.
-- This runs after Kalshi markets are populated via WS lifecycle upsert (no title,
-- so we use the ticker/venue_market_id as the category signal).
--
-- Also re-tries Polymarket markets that didn't match V5 keywords.

-- ── Kalshi: ticker-prefix → category ────────────────────────────────────────
UPDATE markets
SET category = CASE
    -- Crypto intraday (sub-hourly intervals in ticker)
    WHEN venue_market_id ~* '^KX(BTC|ETH|SOL|XRP|DOGE)[^D]*(15M|1H|2H|4H)'
        THEN 'crypto_intraday'

    -- Crypto daily
    WHEN venue_market_id ~* '^KX(BTCD|ETHD|ETHHD|SOLD|DOGE|XRP|BNB|BTC|ETH|SOL)'
        THEN 'crypto_daily'

    -- Sports spread (individual game results — totals, spreads)
    WHEN venue_market_id ~* '^KX(MLB|NBA|NHL|NFLGAME|NFLTOTAL|MLS|NWSL|AFL|EKSTRAKLASA|LIGAMX|KBO|ATPMATCH|WTA|UCLGAME|EPLGAME)'
        THEN 'sports_spread'

    -- Sports team / season / tournament outcomes
    WHEN venue_market_id ~* '^KX(NFLWINS|NFLDIV|NFLCHAMP|PGATOUR|PGATOP|PGAR|PGACHAMP|UFC|NASCAR|MVESPORTS|MVECROSS|MLBWINS|NBACHAMP|NHLCHAMP)'
        THEN 'sports_team'

    -- Politics
    WHEN venue_market_id ~* '^KX(GOV|ELEC|PRES|SEC|SEN|HOUSE|PARTY|VOTE|POLL|CONG|DEMS|REP|MAGA|POTUS|SPOTIFYD)'
        THEN 'politics'

    -- World events
    WHEN venue_market_id ~* '^KX(WAR|NATO|UN|COUNTRY|GEOPO)'
        THEN 'world_events'

    -- Economics / commodities
    WHEN venue_market_id ~* '^KX(GOLDD|GOLD|OIL|GAS|CPI|GDP|FED|RATE|JOB|UNEMPLOY|INFL)'
        THEN 'economics'

    -- Science & tech
    WHEN venue_market_id ~* '^KX(SPACE|SPACEX|FDA|AI|TECH)'
        THEN 'science_tech'

    -- Entertainment
    WHEN venue_market_id ~* '^KX(OSCAR|GRAMMY|EMMY|MOVIE|SHOW|TV)'
        THEN 'entertainment'

    ELSE NULL
END
WHERE venue = 'KALSHI' AND category IS NULL;

-- ── Polymarket: second-pass with broader patterns ────────────────────────────
-- Catches markets missed by V5 (e.g. titles with different phrasing).
UPDATE markets
SET category = CASE
    WHEN title ~* '\y(bitcoin|btc|ethereum|eth|solana|sol|doge|xrp|crypto|defi|nft|stablecoin)\y'
        THEN 'crypto_daily'
    WHEN title ~* '\y(election|president|senate|congress|trump|democrat|republican|vote|ballot|party|governor|white house|supreme court|tariff|nato)\y'
        THEN 'politics'
    WHEN title ~* '\y(super bowl|world series|stanley cup|nba finals|nba champion|mls cup|pga|masters|wimbledon|us open|french open|ufc champion|formula 1|f1 champion|world cup|copa america)\y'
        THEN 'sports_team'
    WHEN title ~* '\y(nfl|nba|mlb|nhl|mls|premier league|champions league|bundesliga|serie a|ufc|mma|nascar|atp|tennis)\y'
        THEN 'sports_team'
    WHEN title ~* '\y(war|ceasefire|invasion|conflict|military|missile|nuclear|ukraine|russia|israel|gaza|taiwan|sanctions|coup)\y'
        THEN 'world_events'
    WHEN title ~* '\y(spacex|nasa|rocket|fda|vaccine|drug approval|ai|artificial intelligence|climate change|renewable|solar)\y'
        THEN 'science_tech'
    WHEN title ~* '\y(inflation|interest rate|federal reserve|unemployment|recession|gdp|cpi|stock market|oil price|gold price)\y'
        THEN 'economics'
    ELSE NULL
END
WHERE venue = 'POLYMARKET' AND category IS NULL;

DO $$
DECLARE
    k_assigned int; k_total int;
    p_assigned int; p_total int;
BEGIN
    SELECT COUNT(*) INTO k_assigned FROM markets WHERE venue = 'KALSHI'     AND category IS NOT NULL;
    SELECT COUNT(*) INTO k_total   FROM markets WHERE venue = 'KALSHI';
    SELECT COUNT(*) INTO p_assigned FROM markets WHERE venue = 'POLYMARKET' AND category IS NOT NULL;
    SELECT COUNT(*) INTO p_total   FROM markets WHERE venue = 'POLYMARKET';
    RAISE NOTICE 'V6: Kalshi %/% categorised, Polymarket %/% categorised', k_assigned, k_total, p_assigned, p_total;
END $$;
