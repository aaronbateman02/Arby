-- V4: Project todo list
CREATE TABLE todo_items (
    id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    title        TEXT        NOT NULL,
    description  TEXT,
    priority     VARCHAR(10) NOT NULL DEFAULT 'medium'
                             CHECK (priority IN ('high', 'medium', 'low')),
    completed    BOOLEAN     NOT NULL DEFAULT FALSE,
    completed_at TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Seed with known pending work
INSERT INTO todo_items (title, description, priority) VALUES
    ('Review and approve first match pairs',
     'Use the Match Review tab to inspect proposed pairs from the first matching run. Approve the best ones before any strategy fires.',
     'high'),
    ('Enable strategy_active for at least one category',
     'Once a batch of pairs is approved, flip strategy_active ON in the Settings tab for that category to start live arb execution.',
     'high'),
    ('Check first scheduled matching run results',
     'The matcher runs every 6 hours. Check Settings → Matching Runs to see markets_embedded, candidates_found, and pairs_proposed.',
     'high'),
    ('Add nginx resolver directive to avoid stale DNS',
     'Add "resolver 127.0.0.11 valid=5s;" to nginx/nginx.conf upstream blocks so container IP changes do not cause 502s after redeploys.',
     'medium'),
    ('Automate UI build in deploy',
     'Add "npm run build" to the deploy flow (or a docker multi-stage build) so UI changes are live without a manual SSH step.',
     'medium'),
    ('Add HuggingFace token secret',
     'Add HF_TOKEN to ~/PolyBot/secrets/hf_token.txt and wire it into the matching service to avoid rate-limit warnings on model download.',
     'low'),
    ('Reindex IVFFlat once markets table is large',
     'The index was created with little data. Once markets has 1000+ rows run: REINDEX INDEX CONCURRENTLY idx_markets_embedding',
     'low');
