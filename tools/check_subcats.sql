SELECT venue, COALESCE(subcategory, 'none') as subcat, COUNT(*) as cnt
FROM markets WHERE subcategory IS NOT NULL AND subcategory != ''
GROUP BY venue, subcategory ORDER BY venue, cnt DESC;
SELECT venue, COUNT(*) as cnt FROM markets WHERE subcategory IS NULL OR subcategory = '' GROUP BY venue;
