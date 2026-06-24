-- ============================================================
-- 8. matching_runs stage tracking
-- Keep trigger within the existing SCHEDULED/MANUAL constraint and
-- store the logical pipeline stage separately so embed/review runs
-- can be tracked and guarded independently.
-- ============================================================
ALTER TABLE matching_runs
    ADD COLUMN IF NOT EXISTS stage VARCHAR(20) NOT NULL DEFAULT 'REVIEW';

ALTER TABLE matching_runs
    DROP CONSTRAINT IF EXISTS matching_runs_stage_check;

ALTER TABLE matching_runs
    ADD CONSTRAINT matching_runs_stage_check
    CHECK (stage IN ('EMBED', 'REVIEW'));

UPDATE matching_runs
SET stage = CASE
    WHEN stage IN ('EMBED', 'REVIEW') THEN stage
    WHEN trigger ILIKE '%EMBED%' THEN 'EMBED'
    ELSE 'REVIEW'
END;
