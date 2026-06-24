-- V5: Keyword-based category assignment for existing Polymarket markets.
-- Kalshi markets get categories at ingest time (CategoryFromTicker in Go).
-- This migration covers the 179k+ Polymarket markets already in the DB.
--
-- Categories that are pair_enabled (the ones matching cares about):
--   politics, science_tech, sports_spread, sports_team, world_events

UPDATE markets
SET category = CASE
    -- Crypto
    WHEN title ~* '\y(bitcoin|btc|ethereum|eth|solana|sol|dogecoin|doge|ripple|xrp|binance|bnb|crypto|defi|nft|web3|altcoin|stablecoin|usdc|usdt)\y'
        THEN 'crypto_daily'

    -- Politics
    WHEN title ~* '\y(trump|biden|harris|democrat|republican|gop|congress|senate|house of representatives|president|presidency|election|ballot|vote|governor|mayor|cabinet|inauguration|impeach|potus|vp|vice president|midterm|primary|poll|approval rating|white house|supreme court|roe|abortion|gun control|immigration|border|tariff|nato|g7|g20|un security|sanctions|executive order)\y'
        THEN 'politics'

    -- Sports spread (game-level: totals, spreads, individual game results)
    WHEN title ~* '\y(will .{1,40} (beat|cover|score more|win by|total|over|under|first half|second half|halftime|inning|quarter|period|set|match))\y'
         AND title ~* '\y(nfl|nba|mlb|nhl|mls|ncaa|college football|college basketball|epl|premier league|champions league|la liga|bundesliga|serie a|ligue 1|ufc|mma|tennis|atp|wta|pga tour)\y'
        THEN 'sports_spread'

    -- Sports team / tournament / season outcomes
    WHEN title ~* '\y(nfl|nba|mlb|nhl|mls|premier league|champions league|la liga|bundesliga|serie a|super bowl|world series|stanley cup|nba finals|nba champion|mls cup|ncaa|college football playoff|pga|masters|us open|wimbledon|french open|australian open|ufc|mma|nascar|formula 1|f1|olympics|world cup|copa america|euro 2024|euro 2025|euro 2026)\y'
        THEN 'sports_team'

    -- World events / geopolitics
    WHEN title ~* '\y(war|ceasefire|invasion|conflict|troops|military|nato|un resolution|sanctions|nuclear|missile|drone|attack|coup|protest|revolution|treaty|peace deal|ukraine|russia|israel|gaza|taiwan|china|north korea|iran|middle east|africa|europe|asia|latin america|refugee|humanitarian)\y'
        THEN 'world_events'

    -- Science & tech
    WHEN title ~* '\y(spacex|nasa|rocket|launch|satellite|mars|moon|asteroid|fda|drug approval|vaccine|clinical trial|cancer|alzheimer|ai|artificial intelligence|gpt|llm|chatgpt|openai|anthropic|google deepmind|autonomous|self-driving|quantum|fusion|climate change|carbon|emissions|renewable|solar|wind energy)\y'
        THEN 'science_tech'

    -- Economics
    WHEN title ~* '\y(gdp|cpi|inflation|interest rate|federal reserve|fed rate|unemployment|jobs report|recession|stock market|s&p 500|nasdaq|dow jones|oil price|gold price|commodities|trade deficit|budget|debt ceiling|ipo)\y'
        THEN 'economics'

    -- Entertainment
    WHEN title ~* '\y(oscar|grammy|emmy|golden globe|box office|movie|film|tv show|series|celebrity|award)\y'
        THEN 'entertainment'

    ELSE NULL
END
WHERE venue = 'POLYMARKET' AND category IS NULL;

-- Log how many got assigned
DO $$
DECLARE
    assigned int;
    total int;
BEGIN
    SELECT COUNT(*) INTO assigned FROM markets WHERE venue = 'POLYMARKET' AND category IS NOT NULL;
    SELECT COUNT(*) INTO total   FROM markets WHERE venue = 'POLYMARKET';
    RAISE NOTICE 'Polymarket category assignment: %/% markets categorised', assigned, total;
END $$;
