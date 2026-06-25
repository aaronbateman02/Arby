SELECT LEFT(venue_market_id, 10) as prefix, COUNT(*) as cnt
FROM markets WHERE venue = 'KALSHI'
GROUP BY prefix ORDER BY cnt DESC LIMIT 30;
