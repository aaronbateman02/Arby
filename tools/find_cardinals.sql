-- Search for Cardinals vs Diamondbacks markets
SELECT venue, venue_market_id, LEFT(title, 120) as title, category, subcategory, market_type, status
FROM markets
WHERE (title ILIKE '%cardinal%' OR title ILIKE '%diamondback%' OR title ILIKE '%d-backs%' OR title ILIKE '%ari%' OR venue_market_id ILIKE '%STL%' OR venue_market_id ILIKE '%ARI%')
AND (title ILIKE '%baseball%' OR title ILIKE '%mlb%' OR title ILIKE '%st.%' OR title ILIKE '%cardinal%')
ORDER BY venue, last_updated_at DESC
LIMIT 20;
