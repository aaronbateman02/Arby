-- V33: Add subcategory column and backfill with league/domain extraction.
-- Mirrors pkg/matching/categorize.go SubcategorizeMarket function.

ALTER TABLE markets ADD COLUMN IF NOT EXISTS subcategory VARCHAR(100);

-- Kalshi: subcategory from ticker prefix
UPDATE markets
SET subcategory = CASE
    WHEN venue_market_id ~* '^KXNFL' THEN 'nfl'
    WHEN venue_market_id ~* '^KXNBA' THEN 'nba'
    WHEN venue_market_id ~* '^KXMLB' THEN 'mlb'
    WHEN venue_market_id ~* '^KXNHL' THEN 'nhl'
    WHEN venue_market_id ~* '^KXMLS|^KXNWSL' THEN 'mls'
    WHEN venue_market_id ~* '^KXUCLGAME|^KXEPLGAME' THEN 'soccer'
    WHEN venue_market_id ~* '^KXAFL' THEN 'afl'
    WHEN venue_market_id ~* '^KXPGATOUR|^KXPGATOP|^KXPGAR|^KXPGACHAMP' THEN 'golf'
    WHEN venue_market_id ~* '^KXUFC' THEN 'ufc'
    WHEN venue_market_id ~* '^KXNASCAR' THEN 'nascar'
    WHEN venue_market_id ~* '^KXATP|^KXWTA' THEN 'tennis'
    WHEN venue_market_id ~* '^KX(GOV|ELEC|PRES|SEC|SEN|HOUSE|PARTY|VOTE|POLL|CONG|DEMS|REP|MAGA|POTUS)' THEN 'us_politics'
    WHEN venue_market_id ~* '^KX(WAR|NATO|UN|COUNTRY|GEOPO)' THEN 'geopolitics'
    WHEN venue_market_id ~* '^KX(BTC|ETH|SOL|XRP|DOGE|BNB)' THEN 'crypto'
    WHEN venue_market_id ~* '^KX(GOLDD|GOLD|OIL|GAS|CPI|GDP|FED|RATE|JOB|UNEMPLOY|INFL)' THEN 'macro'
    WHEN venue_market_id ~* '^KX(SPACE|SPACEX|FDA|AI|TECH)' THEN 'tech'
    WHEN venue_market_id ~* '^KX(OSCAR|GRAMMY|EMMY|MOVIE|SHOW|TV)' THEN 'culture'
    ELSE NULL
END
WHERE venue = 'KALSHI' AND (subcategory IS NULL OR subcategory = '');

-- Polymarket: subcategory from title keywords
UPDATE markets
SET subcategory = CASE
    WHEN title ~* '\y(nfl|super bowl)\y' THEN 'nfl'
    WHEN title ~* '\y(nba|nba finals|nba champion)\y' THEN 'nba'
    WHEN title ~* '\y(mlb|world series)\y' THEN 'mlb'
    WHEN title ~* '\y(nhl|stanley cup)\y' THEN 'nhl'
    WHEN title ~* '\y(mls|mls cup|premier league|champions league|la liga|bundesliga|serie a|ligue 1)\y' THEN 'soccer'
    WHEN title ~* '\y(ufc|mma)\y' THEN 'ufc'
    WHEN title ~* '\y(tennis|atp|wta|wimbledon|us open|french open|australian open)\y' THEN 'tennis'
    WHEN title ~* '\y(golf|pga|masters|pga tour)\y' THEN 'golf'
    WHEN title ~* '\y(nascar|formula 1|f1)\y' THEN 'motorsports'
    WHEN title ~* '\y(trump|biden|harris|election|president|presidency|congress|senate|governor|vote|ballot|democrat|republican)\y' THEN 'us_politics'
    WHEN title ~* '\y(ukraine|russia|israel|gaza|taiwan|china|north korea|iran|war|ceasefire|conflict|sanctions|nato)\y' THEN 'geopolitics'
    WHEN title ~* '\y(bitcoin|btc|ethereum|eth|solana|crypto|defi)\y' THEN 'crypto'
    WHEN title ~* '\y(inflation|interest rate|federal reserve|gdp|cpi|recession|stock market)\y' THEN 'macro'
    WHEN title ~* '\y(spacex|nasa|ai|artificial intelligence|fda|vaccine|climate change)\y' THEN 'tech'
    WHEN title ~* '\y(oscar|grammy|emmy|golden globe|box office|movie)\y' THEN 'culture'
    ELSE NULL
END
WHERE venue = 'POLYMARKET' AND (subcategory IS NULL OR subcategory = '');
