-- Roll back Belcanto L.2 events and RSVP. Refuses while any RSVP,
-- waitlist entry or spot offer exists: participation history is a
-- record of real people's commitments and never disappears silently.

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM event_rsvps)
       OR EXISTS (SELECT 1 FROM event_waitlist_entries)
       OR EXISTS (SELECT 1 FROM event_spot_offers) THEN
        RAISE EXCEPTION 'cannot roll back events while participation history exists';
    END IF;
END
$$;

DROP TRIGGER IF EXISTS event_spot_offers_capacity_guard ON event_spot_offers;
DROP TRIGGER IF EXISTS event_rsvps_capacity_guard ON event_rsvps;
DROP FUNCTION IF EXISTS enforce_event_capacity();
DROP TABLE IF EXISTS event_spot_offers;
DROP TABLE IF EXISTS event_waitlist_entries;
DROP TABLE IF EXISTS event_rsvps;
DROP TABLE IF EXISTS event_occurrences;
DROP TABLE IF EXISTS event_series;
DROP TABLE IF EXISTS event_categories;
