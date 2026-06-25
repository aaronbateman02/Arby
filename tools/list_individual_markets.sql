SELECT id, venue_market_id, title, LEFT(description, 60) as desc_short, category, subcategory, market_type
FROM markets
WHERE venue = 'KALSHI' AND venue_market_id NOT LIKE 'KXMV%'
ORDER BY venue_market_id;
