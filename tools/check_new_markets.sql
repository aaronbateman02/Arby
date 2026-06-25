-- Check for individual event markets with proper titles
SELECT COUNT(*) as total_kalshi FROM markets WHERE venue = 'KALSHI';

SELECT LEFT(venue_market_id, 30) as ticker_prefix, LEFT(title, 60) as title, category, subcategory, market_type
FROM markets
WHERE venue = 'KALSHI'
  AND venue_market_id NOT LIKE 'KXMV%'
LIMIT 30;

-- Check for the specific baseball market
SELECT id, venue_market_id, title, description, category, subcategory, market_type
FROM markets
WHERE venue_market_id LIKE 'KXMLBGAME%'
LIMIT 10;
