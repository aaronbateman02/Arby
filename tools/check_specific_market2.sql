SELECT id, venue_market_id, title, description, category, subcategory, market_type
FROM markets
WHERE venue_market_id LIKE 'KXMLBGAME-26JUN251945AZSTL%'
   OR (title ILIKE '%arizona%' AND title ILIKE '%st.%20louis%' AND venue = 'KALSHI');
