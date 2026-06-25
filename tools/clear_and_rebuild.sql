-- Clear all KALSHI markets and related match candidates/pairs
DELETE FROM match_pairs USING match_candidates
  WHERE match_pairs.candidate_id = match_candidates.id
  AND (match_candidates.market_a_id IN (SELECT id FROM markets WHERE venue = 'KALSHI')
    OR match_candidates.market_b_id IN (SELECT id FROM markets WHERE venue = 'KALSHI'));

DELETE FROM match_candidates
  WHERE market_a_id IN (SELECT id FROM markets WHERE venue = 'KALSHI')
    OR market_b_id IN (SELECT id FROM markets WHERE venue = 'KALSHI');

DELETE FROM markets WHERE venue = 'KALSHI';
