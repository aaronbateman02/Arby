SELECT COUNT(*) as total_markets, COUNT(*) FILTER (WHERE market_type IS NULL) as no_type FROM markets;
SELECT venue, market_type, COUNT(*) as cnt FROM markets WHERE market_type IS NOT NULL AND market_type != '' GROUP BY venue, market_type ORDER BY venue, cnt DESC;
SELECT venue, COUNT(*) as cnt FROM markets WHERE market_type IS NULL OR market_type = '' GROUP BY venue;
