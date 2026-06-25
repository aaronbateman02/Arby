SELECT LEFT(venue_market_id, 30) as ticker_prefix, LEFT(title, 60) as title, category, subcategory, market_type
FROM markets
WHERE venue = 'KALSHI' AND venue_market_id NOT LIKE 'KXMV%'
LIMIT 30;
