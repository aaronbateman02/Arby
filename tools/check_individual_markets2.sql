SELECT COUNT(*) FROM markets WHERE venue = 'KALSHI';
SELECT venue_market_id, title FROM markets WHERE venue = 'KALSHI' AND venue_market_id NOT LIKE 'KXMV%' LIMIT 20;
