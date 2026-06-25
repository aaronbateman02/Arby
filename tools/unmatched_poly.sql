SELECT LEFT(title, 120) as title_preview FROM markets WHERE venue = 'POLYMARKET' AND market_type IS NULL ORDER BY RANDOM() LIMIT 20;
