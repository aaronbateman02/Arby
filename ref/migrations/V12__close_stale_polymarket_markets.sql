-- V12: Close stale POLYMARKET markets that have had no price activity for 7+ days.
-- No-op if leg_prices table does not exist in this environment.
DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'leg_prices') THEN
    CREATE INDEX IF NOT EXISTS idx_leg_prices_recorded_at ON leg_prices (recorded_at);

    CREATE TEMP TABLE active_market_ids AS
    SELECT DISTINCT l.market_id
    FROM   legs l
    JOIN   leg_prices lp ON lp.leg_id = l.id
    WHERE  lp.recorded_at > NOW() - INTERVAL '7 days';

    UPDATE markets
    SET    status = 'CLOSED'
    WHERE  venue  = 'POLYMARKET'
      AND  status = 'OPEN'
      AND  id NOT IN (SELECT market_id FROM active_market_ids);

    DROP TABLE active_market_ids;
  END IF;
END $$;
