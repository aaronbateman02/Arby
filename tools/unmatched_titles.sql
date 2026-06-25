SELECT LEFT(title, 120) as title_preview, venue_market_id FROM markets WHERE venue = 'KALSHI' AND market_type IS NULL LIMIT 20;
