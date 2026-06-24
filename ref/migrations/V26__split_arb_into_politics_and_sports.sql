-- ============================================================
-- V26: Split the cross-market arb strategy into Politics and Sports
--
-- The single `arb-cross-market-v1` strategy historically scanned every
-- approved match pair regardless of category. Politics markets and sports
-- markets have meaningfully different pricing dynamics, liquidity profiles,
-- and resolution mechanics across Kalshi vs Polymarket — keeping them in one
-- code path made future tuning risky.
--
-- This migration:
--   1. Inserts two new strategies — one per market vertical — copying the
--      current arb-cross-market-v1 execution params and adding a top-level
--      `categories` array that the Go strategy uses to filter match pairs.
--   2. Pauses the legacy arb-cross-market-v1 so its risk state stays in DB
--      (preserving historical bundle FKs) but no container should run it.
--
-- Idempotent: re-running this migration is a no-op (uses ON CONFLICT and
-- existence guards).
-- ============================================================

-- ─── 1. Risk-state rows for the two new strategies ──────────────────────────
INSERT INTO strategy_risk_state (strategy_id, status)
VALUES
    ('arb-cross-market-politics-v1', 'ACTIVE'),
    ('arb-cross-market-sports-v1',   'ACTIVE')
ON CONFLICT (strategy_id) DO NOTHING;

-- ─── 2. Politics config (clones execution params from arb-cross-market-v1) ──
INSERT INTO strategy_configs (strategy_id, config_version, config, created_by)
SELECT
    'arb-cross-market-politics-v1',
    COALESCE(
        (SELECT MAX(config_version) FROM strategy_configs WHERE strategy_id = 'arb-cross-market-politics-v1'),
        0
    ) + 1,
    jsonb_build_object(
        'strategy_id',  'arb-cross-market-politics-v1',
        'display_name', 'Cross-Market Arbitrage — Politics',
        'version',      '1.0.0',
        'enabled',      true,
        'paper_mode',   true,
        'categories',   jsonb_build_array('politics'),
        'execution',    jsonb_build_object(
            'min_roi',                  0.05,
            'max_position_usd',         8.00,
            'max_days_to_resolution',   90,
            'price_staleness_secs',     30,
            'opportunity_ttl_ms',       15000,
            'partial_fill_policy',      'UNWIND_ON_PARTIAL'
        ),
        'selling',      jsonb_build_object(
            'early_exit_eligible', true
        ),
        'risk',         jsonb_build_object(
            'circuit_breaker', jsonb_build_object(
                'metric',        'consecutive_losses',
                'threshold',     3,
                'comparison',    'greater_than',
                'alert_message', 'Politics arb strategy has 3 consecutive losses. Strategy paused — manual review required.'
            )
        ),
        'sub_matcher_filters', '[]'::jsonb,
        'custom_settings',     '{}'::jsonb
    ),
    'migration'
WHERE NOT EXISTS (
    SELECT 1 FROM strategy_configs WHERE strategy_id = 'arb-cross-market-politics-v1'
);

-- ─── 3. Sports config ──────────────────────────────────────────────────────
INSERT INTO strategy_configs (strategy_id, config_version, config, created_by)
SELECT
    'arb-cross-market-sports-v1',
    COALESCE(
        (SELECT MAX(config_version) FROM strategy_configs WHERE strategy_id = 'arb-cross-market-sports-v1'),
        0
    ) + 1,
    jsonb_build_object(
        'strategy_id',  'arb-cross-market-sports-v1',
        'display_name', 'Cross-Market Arbitrage — Sports',
        'version',      '1.0.0',
        'enabled',      true,
        'paper_mode',   true,
        'categories',   jsonb_build_array('sports'),
        'execution',    jsonb_build_object(
            'min_roi',                  0.05,
            'max_position_usd',         8.00,
            'max_days_to_resolution',   90,
            'price_staleness_secs',     30,
            'opportunity_ttl_ms',       15000,
            'partial_fill_policy',      'UNWIND_ON_PARTIAL'
        ),
        'selling',      jsonb_build_object(
            'early_exit_eligible', true
        ),
        'risk',         jsonb_build_object(
            'circuit_breaker', jsonb_build_object(
                'metric',        'consecutive_losses',
                'threshold',     3,
                'comparison',    'greater_than',
                'alert_message', 'Sports arb strategy has 3 consecutive losses. Strategy paused — manual review required.'
            )
        ),
        'sub_matcher_filters', '[]'::jsonb,
        'custom_settings',     '{}'::jsonb
    ),
    'migration'
WHERE NOT EXISTS (
    SELECT 1 FROM strategy_configs WHERE strategy_id = 'arb-cross-market-sports-v1'
);

-- ─── 4. Retire the legacy single-strategy row ──────────────────────────────
-- Pause (don't delete) so historical bundles.strategy_id values remain
-- resolvable in reporting queries / dashboards.
UPDATE strategy_risk_state
   SET status        = 'PAUSED',
       paused_at     = COALESCE(paused_at, NOW()),
       paused_reason = COALESCE(
           paused_reason,
           'Superseded by arb-cross-market-politics-v1 and arb-cross-market-sports-v1 (V26).'
       ),
       updated_at    = NOW()
 WHERE strategy_id = 'arb-cross-market-v1'
   AND status      = 'ACTIVE';

-- Also disable the legacy config so any orphaned container that boots
-- against it exits immediately instead of silently double-scanning pairs.
INSERT INTO strategy_configs (strategy_id, config_version, config, created_by)
SELECT
    'arb-cross-market-v1',
    (SELECT MAX(config_version) FROM strategy_configs WHERE strategy_id = 'arb-cross-market-v1') + 1,
    jsonb_set(
        (SELECT config FROM strategy_current_configs WHERE strategy_id = 'arb-cross-market-v1'),
        '{enabled}',
        'false'::jsonb,
        true
    ),
    'migration'
WHERE EXISTS (
    SELECT 1 FROM strategy_current_configs
     WHERE strategy_id = 'arb-cross-market-v1' AND (config->>'enabled')::boolean = true
);
