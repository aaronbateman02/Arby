-- Allow RETROACTIVE_GEMINI as a valid trigger value for matching_runs.
ALTER TABLE matching_runs
    DROP CONSTRAINT matching_runs_trigger_check;
ALTER TABLE matching_runs
    ADD CONSTRAINT matching_runs_trigger_check
        CHECK (trigger = ANY (ARRAY['SCHEDULED', 'MANUAL', 'RETROACTIVE_GEMINI']));
