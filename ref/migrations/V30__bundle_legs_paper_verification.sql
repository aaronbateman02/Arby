-- Per-leg paper-trade fill verification.
--
-- In paper mode the execution service simulates fills using the leg's
-- preflight ask. To know whether a real order WOULD have actually filled at
-- that price we make a second REST call to the venue immediately after the
-- simulated fill and compare the live best-ask to what we paid:
--
--   live_ask <= paper_fill_price → 'VERIFIED'     (the price level still
--                                                  exists; we'd have filled)
--   live_ask >  paper_fill_price → 'PRICE_MOVED'  (price ran past us between
--                                                  the two API calls)
--   any error fetching live ask  → 'CHECK_FAILED' (venue blip; inconclusive)
--
-- A bundle is considered "would have filled" only when EVERY leg is
-- 'VERIFIED'. The aggregate is computed at query time by the reporting
-- service rather than stored on bundles to keep this migration additive.

ALTER TABLE bundle_legs
    ADD COLUMN IF NOT EXISTS paper_verification_status     VARCHAR(20),
    ADD COLUMN IF NOT EXISTS paper_verification_ask        NUMERIC(8,4),
    ADD COLUMN IF NOT EXISTS paper_verification_checked_at TIMESTAMPTZ;

-- Partial index — only the checked rows matter; un-checked rows are
-- non-paper fills or pre-V30 bundles.
CREATE INDEX IF NOT EXISTS idx_bundle_legs_paper_verif_status
    ON bundle_legs(paper_verification_status)
    WHERE paper_verification_status IS NOT NULL;
