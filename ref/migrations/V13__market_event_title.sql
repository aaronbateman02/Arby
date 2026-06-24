-- V13: Add event_title column to markets table.
-- Stores the Kalshi parent event/award name (e.g. "World Cup: Silver Ball Winner")
-- separately from the market subtitle (e.g. "Erling Haaland").
-- Polymarket markets leave this NULL.
ALTER TABLE markets ADD COLUMN IF NOT EXISTS event_title TEXT;
