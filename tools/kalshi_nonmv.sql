SELECT LEFT(venue_market_id, 4) as prefix, COUNT(*) as cnt
FROM markets WHERE venue = 'KALSHI' AND venue_market_id NOT LIKE 'KXMV%'
GROUP BY prefix ORDER BY cnt DESC;
