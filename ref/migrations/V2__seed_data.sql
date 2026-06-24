-- ============================================================
-- PolyBot — Seed Data
-- Migration: 002
-- Date: 2026-05-15
-- ============================================================

-- ============================================================
-- VENUES
-- ============================================================
INSERT INTO venues (id, display_name, base_url, fee_rate, min_order_shares) VALUES
    ('KALSHI',      'Kalshi',      'https://trading-api.kalshi.com/trade-api/v2', 0.07, 1),
    ('POLYMARKET',  'Polymarket',  'https://clob.polymarket.com',                 0.02, 1)
ON CONFLICT (id) DO NOTHING;

-- Note: fee_rate values above are placeholders.
-- Kalshi charges up to 7% on winnings; Polymarket charges ~2% on winnings.
-- Actual fee structures must be confirmed against current API documentation
-- before live trading. The Ingestion Service will update these on market snapshots
-- when per-market fee data is available.

-- ============================================================
-- PLATFORM CONFIG
-- Global settings. All monetary values in USD.
-- ============================================================
INSERT INTO platform_config (key, value, description) VALUES
    (
        'selling_service.min_exit_roi',
        '0.01',
        'Minimum net ROI (fraction) required to trigger an early exit on an eligible bundle. e.g. 0.01 = 1%.'
    ),
    (
        'global_kill_switch',
        'false',
        'When true, all strategy containers are instructed to pause immediately. Set via UI only.'
    ),
    (
        'execution_service.paper_mode',
        'true',
        'When true, the Execution Service simulates fills at current market prices and does not place real orders.'
    ),
    (
        'selling_service.paper_mode',
        'true',
        'When true, the Selling Service logs exit opportunities but does not place real sell orders.'
    )
ON CONFLICT (key) DO NOTHING;

-- ============================================================
-- SUB-MATCHER CONFIGS
-- Initial sub-matchers for v1. All start with human_review_required = true.
-- ============================================================
INSERT INTO sub_matcher_configs (sub_matcher_id, display_name, enabled, human_review_required, auto_approve_threshold, similarity_threshold) VALUES
    (
        'sports-binary',
        'Sports — Binary Outcomes',
        TRUE,
        TRUE,    -- human review required until accuracy is proven
        0.95,
        0.75
    ),
    (
        'politics-binary',
        'Politics — Binary Outcomes',
        FALSE,   -- disabled until sports-binary is validated
        TRUE,
        0.95,
        0.75
    ),
    (
        'politics-multi',
        'Politics — Multi-Candidate',
        FALSE,   -- disabled until sports-binary is validated
        TRUE,
        0.92,    -- slightly lower threshold for complex multi-leg matches
        0.72
    )
ON CONFLICT (sub_matcher_id) DO NOTHING;

-- ============================================================
-- STRATEGY RISK STATE (initial active state for known strategies)
-- ============================================================
INSERT INTO strategy_risk_state (strategy_id, status) VALUES
    ('arb-cross-market-v1', 'ACTIVE')
ON CONFLICT (strategy_id) DO NOTHING;

-- ============================================================
-- STRATEGY CONFIGS (v1 — arb-cross-market, paper mode, minimum sizes)
-- ============================================================
INSERT INTO strategy_configs (strategy_id, config_version, config, created_by) VALUES
    (
        'arb-cross-market-v1',
        1,
        '{
            "strategy_id": "arb-cross-market-v1",
            "version": "1.0.0",
            "display_name": "Cross-Market Arbitrage v1",
            "enabled": true,
            "paper_mode": true,
            "sub_matcher_filters": [],
            "execution": {
                "shares_per_leg": 1,
                "min_roi": 0.02,
                "opportunity_ttl_ms": 15000,
                "partial_fill_policy": "UNWIND_ON_PARTIAL"
            },
            "selling": {
                "early_exit_eligible": true
            },
            "risk": {
                "circuit_breaker": {
                    "metric": "consecutive_losses",
                    "threshold": 3,
                    "comparison": "greater_than",
                    "alert_message": "Arb strategy has 3 consecutive losses. Strategy paused — manual review required."
                }
            },
            "custom_settings": {}
        }',
        'system'
    )
ON CONFLICT (strategy_id, config_version) DO NOTHING;
