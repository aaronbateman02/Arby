-- V32: Assign categories to existing uncategorized markets.
-- Mirrors pkg/matching/categorize.go logic so Go and SQL stay aligned.

-- Kalshi: ticker prefix patterns
UPDATE markets
SET category = CASE
    WHEN venue_market_id ~* '^KX(BTC|ETH|SOL|XRP|DOGE)[^D]*(15M|1H|2H|4H)' THEN 'crypto_intraday'
    WHEN venue_market_id ~* '^KX(BTCD|ETHD|ETHHD|SOLD|DOGE|XRP|BNB|BTC|ETH|SOL)' THEN 'crypto_daily'
    WHEN venue_market_id ~* '^KX(MLB|NBA|NHL|NFLGAME|NFLTOTAL|MLS|NWSL|AFL|EKSTRAKLASA|LIGAMX|KBO|ATPMATCH|WTA|UCLGAME|EPLGAME)' THEN 'sports_spread'
    WHEN venue_market_id ~* '^KX(NFLWINS|NFLDIV|NFLCHAMP|PGATOUR|PGATOP|PGAR|PGACHAMP|UFC|NASCAR|MVESPORTS|MVECROSS|MLBWINS|NBACHAMP|NHLCHAMP)' THEN 'sports_team'
    WHEN venue_market_id ~* '^KX(GOV|ELEC|PRES|SEC|SEN|HOUSE|PARTY|VOTE|POLL|CONG|DEMS|REP|MAGA|POTUS|SPOTIFYD)' THEN 'politics'
    WHEN venue_market_id ~* '^KX(WAR|NATO|UN|COUNTRY|GEOPO)' THEN 'world_events'
    WHEN venue_market_id ~* '^KX(GOLDD|GOLD|OIL|GAS|CPI|GDP|FED|RATE|JOB|UNEMPLOY|INFL)' THEN 'economics'
    WHEN venue_market_id ~* '^KX(SPACE|SPACEX|FDA|AI|TECH)' THEN 'science_tech'
    WHEN venue_market_id ~* '^KX(OSCAR|GRAMMY|EMMY|MOVIE|SHOW|TV)' THEN 'entertainment'
    ELSE 'uncategorized'
END
WHERE venue = 'KALSHI' AND (category IS NULL OR category = '');

-- Polymarket: title keyword patterns
UPDATE markets
SET category = CASE
    WHEN title ~* '\y(bitcoin|btc|ethereum|eth|solana|sol|dogecoin|doge|ripple|xrp|binance|bnb|crypto|defi|nft|web3|altcoin|stablecoin|usdc|usdt)\y' THEN 'crypto_daily'
    WHEN title ~* '\y(trump|biden|harris|democrat|republican|gop|congress|senate|house of representatives|president|presidency|election|ballot|vote|governor|mayor|cabinet|inauguration|impeach|potus|vp|vice president|midterm|primary|poll|approval rating|white house|supreme court|roe|abortion|gun control|immigration|border|tariff|nato|g7|g20|un security|sanctions|executive order)\y' THEN 'politics'
    WHEN title ~* '\y(nfl|nba|mlb|nhl|mls|ncaa|college football|college basketball|epl|premier league|champions league|la liga|bundesliga|serie a|ligue 1|ufc|mma|tennis|atp|wta|pga tour)\y' THEN 'sports_team'
    WHEN title ~* '\y(super bowl|world series|stanley cup|nba finals|nba champion|mls cup|pga|masters|us open|wimbledon|french open|australian open|ufc champion|formula 1|f1 champion|world cup|copa america|euro 2024|euro 2025|euro 2026|olympics|nascar)\y' THEN 'sports_team'
    WHEN title ~* '\y(war|ceasefire|invasion|conflict|troops|military|nato|un resolution|sanctions|nuclear|missile|drone|attack|coup|protest|revolution|treaty|peace deal|ukraine|russia|israel|gaza|taiwan|china|north korea|iran|middle east|africa|europe|asia|latin america|refugee|humanitarian)\y' THEN 'world_events'
    WHEN title ~* '\y(spacex|nasa|rocket|launch|satellite|mars|moon|asteroid|fda|drug approval|vaccine|clinical trial|cancer|alzheimer|ai|artificial intelligence|gpt|llm|chatgpt|openai|anthropic|google deepmind|autonomous|self-driving|quantum|fusion|climate change|carbon|emissions|renewable|solar|wind energy)\y' THEN 'science_tech'
    WHEN title ~* '\y(gdp|cpi|inflation|interest rate|federal reserve|fed rate|unemployment|jobs report|recession|stock market|s&p 500|nasdaq|dow jones|oil price|gold price|commodities|trade deficit|budget|debt ceiling|ipo)\y' THEN 'economics'
    WHEN title ~* '\y(oscar|grammy|emmy|golden globe|box office|movie|film|tv show|series|celebrity|award)\y' THEN 'entertainment'
    WHEN title ~* '\y(climate|weather|hurricane|tornado|flood|drought|temperature|record heat|cold spell|blizzard)\y' THEN 'climate'
    ELSE 'uncategorized'
END
WHERE venue = 'POLYMARKET' AND (category IS NULL OR category = '');
