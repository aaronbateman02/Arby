-- Add `execution.max_positions_per_bundle` to every existing strategy config
-- by inserting a NEW config_version row per strategy whose current config is
-- missing the key. Existing versions are immutable (audit log).
-- Default is 1 — strategies will not open a second simultaneous bundle on
-- the same match_pair until the first is exited (state = ABORTED or
-- actual_payout IS NOT NULL).
--
-- A value <0 means "unlimited". The loader (strategies/strategy-arb/store.go)
-- normalises a missing/zero value to 1 as well, so this migration is
-- belt-and-suspenders and gives operators a default to edit via the UI.

INSERT INTO strategy_configs (strategy_id, config_version, config, created_by)
SELECT  scc.strategy_id,
        scc.config_version + 1,
        jsonb_set(
            scc.config,
            '{execution,max_positions_per_bundle}',
            to_jsonb(1),
            true
        ),
        'migration:V29'
FROM    strategy_current_configs scc
WHERE   NOT (scc.config -> 'execution' ? 'max_positions_per_bundle');
