-- Migration: add events table and event columns to markets
-- Run against the existing arby database

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Create events table
CREATE TABLE IF NOT EXISTS events (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    venue           VARCHAR(20) NOT NULL,
    venue_event_id  VARCHAR(255) NOT NULL,
    title           TEXT NOT NULL,
    description     TEXT,
    category        VARCHAR(100),
    status          VARCHAR(20) NOT NULL DEFAULT 'OPEN',
    close_time      TIMESTAMPTZ,
    first_seen_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (venue, venue_event_id)
);
CREATE INDEX IF NOT EXISTS idx_events_venue ON events(venue);

-- Add event columns to markets
ALTER TABLE markets ADD COLUMN IF NOT EXISTS event_id UUID REFERENCES events(id);
ALTER TABLE markets ADD COLUMN IF NOT EXISTS venue_event_id VARCHAR(255);
CREATE INDEX IF NOT EXISTS idx_markets_event ON markets(event_id);
