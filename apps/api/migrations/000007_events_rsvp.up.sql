-- Belcanto L.2 events and RSVP (Figma Page 24 event catalog, Page 29
-- admin operations; HOF DEC-001, DEC-003, DEC-101). Events never mix
-- with core lessons (DEC-001). RSVP binds to one occurrence, never the
-- series (DEC-003). Categories are data, not an enum. A seat is held by
-- a confirmed RSVP or by one pending spot offer; the capacity invariant
-- counts both, and triggers close the race a plain CHECK cannot see.
-- DEC-101 (waitlist TTL) is open: the offer expiry is always supplied by
-- the service from configuration, never assumed here.

CREATE TABLE event_categories (
    id         text PRIMARY KEY CHECK (char_length(id) BETWEEN 1 AND 128),
    tenant_id  text NOT NULL REFERENCES tenants(id),
    name       text NOT NULL CHECK (char_length(btrim(name)) BETWEEN 1 AND 100),
    status     text NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'archived')),
    created_at timestamptz NOT NULL,
    UNIQUE (tenant_id, id),
    UNIQUE (tenant_id, name)
);

CREATE TABLE event_series (
    id                    text PRIMARY KEY CHECK (char_length(id) BETWEEN 1 AND 128),
    tenant_id             text NOT NULL REFERENCES tenants(id),
    category_id           text NOT NULL,
    title                 text NOT NULL CHECK (char_length(btrim(title)) BETWEEN 1 AND 200),
    description           text CHECK (description IS NULL OR char_length(description) BETWEEN 1 AND 2000),
    host_account_id       text NOT NULL,
    room_id               text,
    capacity              integer NOT NULL CHECK (capacity BETWEEN 1 AND 500),
    weekday               smallint NOT NULL CHECK (weekday BETWEEN 0 AND 6),
    start_minutes         smallint NOT NULL CHECK (start_minutes BETWEEN 0 AND 1439),
    duration_minutes      integer NOT NULL CHECK (duration_minutes BETWEEN 1 AND 1440),
    effective_from        date NOT NULL,
    effective_until       date CHECK (effective_until IS NULL OR effective_until >= effective_from),
    status                text NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'paused', 'ended')),
    version               bigint NOT NULL DEFAULT 0 CHECK (version >= 0),
    created_by_account_id text NOT NULL,
    created_at            timestamptz NOT NULL,
    updated_at            timestamptz NOT NULL,
    UNIQUE (tenant_id, id),
    FOREIGN KEY (tenant_id, category_id) REFERENCES event_categories(tenant_id, id),
    FOREIGN KEY (tenant_id, host_account_id) REFERENCES accounts(tenant_id, id),
    FOREIGN KEY (tenant_id, room_id) REFERENCES rooms(tenant_id, id),
    FOREIGN KEY (tenant_id, created_by_account_id) REFERENCES accounts(tenant_id, id)
);

CREATE TABLE event_occurrences (
    id                    text PRIMARY KEY CHECK (char_length(id) BETWEEN 1 AND 128),
    tenant_id             text NOT NULL REFERENCES tenants(id),
    series_id             text,
    category_id           text NOT NULL,
    title                 text NOT NULL CHECK (char_length(btrim(title)) BETWEEN 1 AND 200),
    description           text CHECK (description IS NULL OR char_length(description) BETWEEN 1 AND 2000),
    starts_at             timestamptz NOT NULL,
    duration_minutes      integer NOT NULL CHECK (duration_minutes BETWEEN 1 AND 1440),
    host_account_id       text NOT NULL,
    room_id               text,
    capacity              integer NOT NULL CHECK (capacity BETWEEN 1 AND 500),
    status                text NOT NULL DEFAULT 'scheduled' CHECK (status IN ('scheduled', 'completed', 'cancelled')),
    version               bigint NOT NULL DEFAULT 0 CHECK (version >= 0),
    created_by_account_id text NOT NULL,
    created_at            timestamptz NOT NULL,
    updated_at            timestamptz NOT NULL,
    UNIQUE (tenant_id, id),
    FOREIGN KEY (tenant_id, series_id) REFERENCES event_series(tenant_id, id),
    FOREIGN KEY (tenant_id, category_id) REFERENCES event_categories(tenant_id, id),
    FOREIGN KEY (tenant_id, host_account_id) REFERENCES accounts(tenant_id, id),
    FOREIGN KEY (tenant_id, room_id) REFERENCES rooms(tenant_id, id),
    FOREIGN KEY (tenant_id, created_by_account_id) REFERENCES accounts(tenant_id, id)
);

CREATE INDEX event_occurrences_tenant_starts_idx
    ON event_occurrences (tenant_id, starts_at, id);

CREATE INDEX event_occurrences_series_idx
    ON event_occurrences (tenant_id, series_id, starts_at)
    WHERE series_id IS NOT NULL;

CREATE TABLE event_rsvps (
    tenant_id     text NOT NULL,
    occurrence_id text NOT NULL,
    student_id    text NOT NULL,
    status        text NOT NULL CHECK (status IN ('confirmed', 'cancelled')),
    confirmed_at  timestamptz,
    cancelled_at  timestamptz,
    updated_at    timestamptz NOT NULL,
    PRIMARY KEY (tenant_id, occurrence_id, student_id),
    CHECK (
        (status = 'confirmed' AND confirmed_at IS NOT NULL)
        OR (status = 'cancelled' AND cancelled_at IS NOT NULL)
    ),
    FOREIGN KEY (tenant_id, occurrence_id) REFERENCES event_occurrences(tenant_id, id),
    FOREIGN KEY (tenant_id, student_id) REFERENCES students(tenant_id, id)
);

CREATE INDEX event_rsvps_student_idx
    ON event_rsvps (tenant_id, student_id, occurrence_id);

CREATE TABLE event_waitlist_entries (
    tenant_id     text NOT NULL,
    occurrence_id text NOT NULL,
    student_id    text NOT NULL,
    position      integer NOT NULL CHECK (position > 0),
    joined_at     timestamptz NOT NULL,
    PRIMARY KEY (tenant_id, occurrence_id, student_id),
    UNIQUE (tenant_id, occurrence_id, position) DEFERRABLE INITIALLY DEFERRED,
    FOREIGN KEY (tenant_id, occurrence_id) REFERENCES event_occurrences(tenant_id, id),
    FOREIGN KEY (tenant_id, student_id) REFERENCES students(tenant_id, id)
);

CREATE TABLE event_spot_offers (
    id            text PRIMARY KEY CHECK (char_length(id) BETWEEN 1 AND 128),
    tenant_id     text NOT NULL REFERENCES tenants(id),
    occurrence_id text NOT NULL,
    student_id    text NOT NULL,
    status        text NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'confirmed', 'declined', 'expired')),
    offered_at    timestamptz NOT NULL,
    expires_at    timestamptz NOT NULL CHECK (expires_at > offered_at),
    resolved_at   timestamptz,
    CHECK (
        (status = 'pending' AND resolved_at IS NULL)
        OR (status <> 'pending' AND resolved_at IS NOT NULL)
    ),
    UNIQUE (tenant_id, id),
    FOREIGN KEY (tenant_id, occurrence_id) REFERENCES event_occurrences(tenant_id, id),
    FOREIGN KEY (tenant_id, student_id) REFERENCES students(tenant_id, id)
);

-- One pending offer per occurrence: the seat passes down the waitlist
-- one head at a time, deterministically.
CREATE UNIQUE INDEX event_spot_offers_single_pending_idx
    ON event_spot_offers (tenant_id, occurrence_id)
    WHERE status = 'pending';

CREATE INDEX event_spot_offers_student_idx
    ON event_spot_offers (tenant_id, student_id)
    WHERE status = 'pending';

-- Capacity invariant: confirmed RSVPs plus pending offers never exceed
-- capacity. The occurrence row lock serializes concurrent claims.
CREATE FUNCTION enforce_event_capacity() RETURNS trigger AS $$
DECLARE
    seat_capacity integer;
    held_seats    integer;
BEGIN
    SELECT capacity INTO seat_capacity
    FROM event_occurrences
    WHERE tenant_id = NEW.tenant_id AND id = NEW.occurrence_id
    FOR UPDATE;
    IF seat_capacity IS NULL THEN
        RAISE EXCEPTION 'event occurrence % was not found', NEW.occurrence_id;
    END IF;
    SELECT (
        SELECT count(*) FROM event_rsvps
        WHERE tenant_id = NEW.tenant_id AND occurrence_id = NEW.occurrence_id
          AND status = 'confirmed'
    ) + (
        SELECT count(*) FROM event_spot_offers
        WHERE tenant_id = NEW.tenant_id AND occurrence_id = NEW.occurrence_id
          AND status = 'pending'
    ) INTO held_seats;
    IF held_seats > seat_capacity THEN
        RAISE EXCEPTION 'event occurrence % is over capacity', NEW.occurrence_id;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE CONSTRAINT TRIGGER event_rsvps_capacity_guard
    AFTER INSERT OR UPDATE ON event_rsvps
    FOR EACH ROW WHEN (NEW.status = 'confirmed')
    EXECUTE FUNCTION enforce_event_capacity();

CREATE CONSTRAINT TRIGGER event_spot_offers_capacity_guard
    AFTER INSERT OR UPDATE ON event_spot_offers
    FOR EACH ROW WHEN (NEW.status = 'pending')
    EXECUTE FUNCTION enforce_event_capacity();
