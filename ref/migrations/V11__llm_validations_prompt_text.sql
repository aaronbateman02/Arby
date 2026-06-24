-- Store the exact user-message prompt sent to the LLM for each validation.
-- This enables auditing of what context the model was given and iterating
-- on prompt quality when decisions are wrong.
ALTER TABLE llm_validations
    ADD COLUMN IF NOT EXISTS prompt_text text;
