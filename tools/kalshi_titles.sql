SELECT LEFT(title, 80) as title_preview, COUNT(*) as cnt
FROM markets WHERE venue = 'KALSHI' AND venue_market_id LIKE 'KXMVESPORT%'
GROUP BY LEFT(title, 80) ORDER BY cnt DESC LIMIT 20;
