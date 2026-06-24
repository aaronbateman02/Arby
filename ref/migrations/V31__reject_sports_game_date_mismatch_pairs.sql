-- Reject APPROVED sports game-level pairs whose venue close dates differ by
-- more than 1 calendar day. These were a major source of false-positive arbs
-- (same teams, different scheduled games).

UPDATE match_pairs mp
SET    status = 'REJECTED',
       rejection_reason = 'guardrail_date_mismatch_backfill',
       reviewed_at = NOW(),
       reviewed_by = 'AUTO',
       updated_at = NOW()
FROM   markets ma,
       markets mb
WHERE  ma.id = mp.market_a_id
  AND  mb.id = mp.market_b_id
  AND  ma.category = 'sports'
  AND  mp.status = 'APPROVED'
  AND  ma.resolution_date IS NOT NULL
  AND  mb.resolution_date IS NOT NULL
  AND  ABS(EXTRACT(EPOCH FROM (ma.resolution_date - mb.resolution_date)) / 86400) > 1
  AND  NOT (
           ma.title ~* '\m(division|champion|championship|title|winner|pennant|cup|trophy|mvp|rookie of the year|cy young|most valuable|season|regular season|playoffs?|finals?|conference)\M'
        OR mb.title ~* '\m(division|champion|championship|title|winner|pennant|cup|trophy|mvp|rookie of the year|cy young|most valuable|season|regular season|playoffs?|finals?|conference)\M'
       );
