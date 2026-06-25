SELECT venue_market_id, LEFT(title, 80) as title, LEFT(description, 120) as desc, category, subcategory FROM markets ORDER BY RANDOM() LIMIT 10;
