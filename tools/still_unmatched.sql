SELECT LEFT(title, 100) as title_preview FROM markets WHERE venue = 'KALSHI' AND (market_type IS NULL OR market_type = '') LIMIT 10;
SELECT LEFT(title, 100) as title_preview FROM markets WHERE venue = 'POLYMARKET' AND (market_type IS NULL OR market_type = '') LIMIT 10;
