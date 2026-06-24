-- Track when a manually-rejected pair has been "prompt-reviewed"
-- (i.e. the rejection reason was incorporated into a prompt update).
-- NULL  = not yet reviewed for prompts (shows up in the Prompt Review tab).
-- SET   = reviewed at this timestamp; won't appear in the unreviewed list.

ALTER TABLE match_pairs
    ADD COLUMN IF NOT EXISTS prompt_reviewed_at TIMESTAMPTZ;
