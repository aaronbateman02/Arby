SELECT
  CASE WHEN venue_market_id LIKE 'KXMV%' THEN 'multi_venue_bundle'
       WHEN venue_market_id LIKE 'KX%' THEN 'individual_event'
       ELSE 'other'
  END as market_kind,
  COUNT(*) as cnt
FROM markets WHERE venue = 'KALSHI'
GROUP BY 1;

SELECT LEFT(venue_market_id, 20) as ticker_prefix, COUNT(*) as cnt
FROM markets WHERE venue = 'KALSHI' AND venue_market_id NOT LIKE 'KXMV%'
GROUP BY ticker_prefix ORDER BY cnt DESC;

-- Any single-game MLB markets?
SELECT venue_market_id, LEFT(title, 100) FROM markets WHERE venue = 'KALSHI' AND venue_market_id LIKE 'KXMLB%' LIMIT 10;

-- Check Polymarket structure
SELECT venue_market_id, LEFT(title, 100) FROM markets WHERE venue = 'POLYMARKET' AND title ILIKE '%cardinal%' LIMIT 10;
