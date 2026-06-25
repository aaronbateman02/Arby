SELECT venue, structure_type, COUNT(*) as cnt
FROM markets WHERE structure_type IS NOT NULL AND structure_type != ''
GROUP BY venue, structure_type ORDER BY venue, cnt DESC;
