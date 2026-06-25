SELECT LEFT(venue_market_id, 45) as ticker, LEFT(title, 60) as title, embedding IS NOT NULL as has_emb
FROM markets WHERE venue = 'KALSHI' ORDER BY venue_market_id;
