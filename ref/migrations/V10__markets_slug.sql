-- Store venue-specific URL slugs so the UI can link directly to each market.
-- Polymarket markets have a human-readable slug (e.g. "will-x-win-nba-mvp").
-- Kalshi markets use their ticker as the URL key; we leave slug NULL for Kalshi.
ALTER TABLE markets ADD COLUMN IF NOT EXISTS slug varchar(512);
CREATE INDEX IF NOT EXISTS idx_markets_slug ON markets(slug) WHERE slug IS NOT NULL;
