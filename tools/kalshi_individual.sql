SELECT COUNT(*) as cnt FROM markets WHERE venue = 'KALSHI' AND venue_market_id NOT LIKE 'KXMV%' AND embedding IS NOT NULL;
SELECT id, venue_market_id, LEFT(title, 100) as title FROM markets WHERE venue = 'KALSHI' AND venue_market_id NOT LIKE 'KXMV%' AND embedding IS NOT NULL ORDER BY venue_market_id;
