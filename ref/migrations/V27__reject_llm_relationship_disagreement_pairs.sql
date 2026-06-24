-- V27: Reject match_pairs where the GPT and Gemini reviewers disagreed on
-- relationship (EQUIVALENT vs INVERSE). Prior to V27, the host stored GPT's
-- relationship even when Gemini disagreed, producing non-hedged "arb" bundles
-- (notably for H2H sports markets, where Kalshi has one binary market per team
-- and Polymarket has a single head-to-head market). The matching service is
-- now patched to reject such pairs up-front; this migration backfills the
-- already-APPROVED ones to REJECTED.

WITH per_pair AS (
    SELECT
        market_a_id,
        market_b_id,
        max(CASE WHEN model LIKE 'gpt%'    THEN result_json->>'relationship' END) AS gpt_rel,
        max(CASE WHEN model LIKE 'gemini%' THEN result_json->>'relationship' END) AS gem_rel
    FROM llm_validations
    GROUP BY market_a_id, market_b_id
),
bad AS (
    SELECT mp.id
    FROM match_pairs mp
    JOIN per_pair pp
      ON pp.market_a_id = mp.market_a_id
     AND pp.market_b_id = mp.market_b_id
    WHERE pp.gpt_rel IS NOT NULL
      AND pp.gem_rel IS NOT NULL
      AND pp.gpt_rel <> pp.gem_rel
      AND mp.status IN ('APPROVED', 'PENDING')
)
UPDATE match_pairs mp
   SET status           = 'REJECTED',
       rejection_reason = 'llm_relationship_disagreement_backfill',
       updated_at       = NOW()
  FROM bad
 WHERE mp.id = bad.id;
