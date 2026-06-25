SELECT m1.id, m1.venue_market_id, LEFT(m1.title, 80) as title_a,
       m2.id, m2.venue_market_id, LEFT(m2.title, 80) as title_b,
       ROUND((1 - (m1.embedding <=> m2.embedding))::numeric, 6) as sim
FROM markets m1, markets m2
WHERE m1.id = '03fa57fc-ed29-4cf7-9e2e-d0eff234285d'
  AND m2.venue = 'POLYMARKET'
  AND m2.embedding IS NOT NULL
  AND m2.id != m1.id
  AND (1 - (m1.embedding <=> m2.embedding)) IS NAN
ORDER BY sim NULLS FIRST
LIMIT 5;
