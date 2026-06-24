-- capital_reservations.bundle_id had an FK → bundles(id), but the risk service
-- inserts the reservation BEFORE the bundle row exists (circular dependency).
-- The column is still used as a lookup key; only the constraint is removed.
ALTER TABLE capital_reservations
    DROP CONSTRAINT IF EXISTS fk_capital_reservations_bundle;
