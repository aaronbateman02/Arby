-- ============================================================
-- V17: Arb Strategy Config v2
--
-- Adds execution fields required by the cross-market arb scanner:
--   max_position_usd        – max total spend per bundle ($8 default)
--   max_days_to_resolution  – skip pairs expiring beyond this window (90 days)
--   price_staleness_secs    – reject prices older than this threshold (30 sec)
--
-- Also corrects min_roi from the placeholder 2% to the target 5%.
-- Inserts as config_version=2 so the audit history is preserved.
-- ============================================================

INSERT INTO strategy_configs (strategy_id, config_version, config, created_by)
SELECT
    'arb-cross-market-v1',
    COALESCE((SELECT MAX(config_version) FROM strategy_configs WHERE strategy_id = 'arb-cross-market-v1'), 0) + 1,
    jsonb_build_object(
        'strategy_id',   'arb-cross-market-v1',
        'display_name',  'Cross-Market Arbitrage v1',
        'version',       '2.0.0',
        'enabled',       true,
        'paper_mode',    true,
        'execution',     jsonb_build_object(
            'min_roi',                  0.05,
            'max_position_usd',         8.00,
            'max_days_to_resolution',   90,
            'price_staleness_secs',     30,
            'opportunity_ttl_ms',       15000,
            'partial_fill_policy',      'UNWIND_ON_PARTIAL'
        ),
        'selling',       jsonb_build_object(
            'early_exit_eligible', true
        ),
        'risk',          jsonb_build_object(
            'circuit_breaker', jsonb_build_object(
                'metric',        'consecutive_losses',
                'threshold',     3,
                'comparison',    'greater_than',
                'alert_message', 'Arb strategy has 3 consecutive losses. Strategy paused — manual review required.'
            )
        ),
        'sub_matcher_filters', '[]'::jsonb,
        'custom_settings',     '{}'::jsonb
    ),
    'migration'
WHERE EXISTS (
    SELECT 1 FROM strategy_risk_state WHERE strategy_id = 'arb-cross-market-v1'
);
