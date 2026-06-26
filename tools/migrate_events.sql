-- Migration: match old PolyBot events schema
ALTER TABLE events ADD COLUMN IF NOT EXISTS mutually_exclusive BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE events ADD COLUMN IF NOT EXISTS description TEXT;
CREATE INDEX IF NOT EXISTS idx_events_category ON events(category);

ALTER TABLE markets ADD COLUMN IF NOT EXISTS event_title TEXT;
