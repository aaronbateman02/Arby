-- How many candidates exist?
SELECT COUNT(*) as total_candidates, COUNT(*) FILTER (WHERE status = 'PENDING') as pending, COUNT(*) FILTER (WHERE status = 'APPROVED') as approved, COUNT(*) FILTER (WHERE status = 'REJECTED') as rejected FROM match_candidates;

-- What do they look like?
SELECT mc.id, ma.venue as venue_a, ma.title as title_a_left80, mb.venue as venue_b, mb.title as title_b_left80, mc.similarity, mc.category, mc.status
FROM match_candidates mc
JOIN markets ma ON ma.id = mc.market_a_id
JOIN markets mb ON mb.id = mc.market_b_id
ORDER BY mc.similarity DESC
LIMIT 30;
