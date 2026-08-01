CREATE TABLE IF NOT EXISTS tenants (
    id          text PRIMARY KEY CHECK (char_length(id) BETWEEN 1 AND 128),
    name        text NOT NULL CHECK (char_length(btrim(name)) BETWEEN 1 AND 200),
    created_at  timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS people (
    id          text PRIMARY KEY,
    tenant_id   text NOT NULL REFERENCES tenants(id),
    full_name   text NOT NULL CHECK (char_length(btrim(full_name)) BETWEEN 1 AND 200),
    created_at  timestamptz NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, id)
);

CREATE TABLE IF NOT EXISTS school_memberships (
    id          text PRIMARY KEY,
    tenant_id   text NOT NULL REFERENCES tenants(id),
    person_id   text NOT NULL,
    created_at  timestamptz NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, id),
    UNIQUE (tenant_id, person_id),
    FOREIGN KEY (tenant_id, person_id) REFERENCES people(tenant_id, id)
);

CREATE TABLE IF NOT EXISTS accounts (
    id          text PRIMARY KEY,
    tenant_id   text NOT NULL REFERENCES tenants(id),
    person_id   text NOT NULL,
    status      text NOT NULL CHECK (status IN ('pending_activation', 'active', 'suspended')),
    activated_at timestamptz,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now(),
    version     bigint NOT NULL DEFAULT 0 CHECK (version >= 0),
	CHECK (
		(status = 'pending_activation' AND activated_at IS NULL)
		OR (status IN ('active', 'suspended') AND activated_at IS NOT NULL)
	),
    UNIQUE (tenant_id, id),
    UNIQUE (tenant_id, person_id),
    FOREIGN KEY (tenant_id, person_id) REFERENCES people(tenant_id, id)
);

CREATE TABLE IF NOT EXISTS account_login_identifiers (
    id              bigserial PRIMARY KEY,
    account_id      text NOT NULL REFERENCES accounts(id),
    tenant_id       text NOT NULL,
    identifier_type text NOT NULL CHECK (identifier_type = 'phone'),
    normalized_value text NOT NULL CHECK (normalized_value ~ '^\+[1-9][0-9]{7,14}$'),
    status          text NOT NULL CHECK (status IN ('reserved', 'confirmed', 'disabled')),
    confirmed_at    timestamptz,
    created_at      timestamptz NOT NULL DEFAULT now(),
    UNIQUE (normalized_value),
    UNIQUE (account_id, identifier_type),
    FOREIGN KEY (tenant_id, account_id) REFERENCES accounts(tenant_id, id)
);

CREATE TABLE IF NOT EXISTS password_credentials (
    account_id      text PRIMARY KEY REFERENCES accounts(id),
    password_hash   text NOT NULL,
    algorithm       text NOT NULL CHECK (algorithm = 'argon2id'),
    created_at      timestamptz NOT NULL,
    updated_at      timestamptz NOT NULL
);

CREATE TABLE IF NOT EXISTS role_grants (
    id              text PRIMARY KEY,
    tenant_id       text NOT NULL REFERENCES tenants(id),
    account_id      text NOT NULL,
    role_type       text NOT NULL CHECK (role_type IN ('Owner', 'Administrator', 'Teacher', 'Student')),
    scope_type      text NOT NULL CHECK (scope_type IN ('tenant', 'student')),
    scope_id        text NOT NULL,
    status          text NOT NULL CHECK (status IN ('active', 'revoked')),
    granted_by      text,
    granted_at      timestamptz NOT NULL,
    revoked_at      timestamptz,
	CHECK (
		(role_type = 'Student' AND scope_type = 'student')
		OR (role_type IN ('Owner', 'Administrator', 'Teacher') AND scope_type = 'tenant' AND scope_id = tenant_id)
	),
	CHECK (
		(status = 'active' AND revoked_at IS NULL)
		OR (status = 'revoked' AND revoked_at IS NOT NULL)
	),
    UNIQUE (tenant_id, id),
    FOREIGN KEY (tenant_id, account_id) REFERENCES accounts(tenant_id, id),
    FOREIGN KEY (tenant_id, granted_by) REFERENCES accounts(tenant_id, id)
);

CREATE UNIQUE INDEX IF NOT EXISTS role_grants_one_active_scope
    ON role_grants (tenant_id, account_id, role_type, scope_type, scope_id)
    WHERE status = 'active';

CREATE UNIQUE INDEX IF NOT EXISTS role_grants_one_active_owner_per_tenant
    ON role_grants (tenant_id)
    WHERE role_type = 'Owner' AND status = 'active';

CREATE TABLE IF NOT EXISTS capability_delegations (
    id                  text PRIMARY KEY,
    tenant_id           text NOT NULL REFERENCES tenants(id),
    grantee_account_id  text NOT NULL,
    granted_by_account_id text NOT NULL,
    bundle              text NOT NULL CHECK (bundle = 'StudentOnboardingManager.v1'),
    status              text NOT NULL CHECK (status IN ('active', 'revoked', 'expired')),
    reason              text NOT NULL CHECK (char_length(btrim(reason)) BETWEEN 1 AND 500),
    granted_at          timestamptz NOT NULL,
    expires_at          timestamptz,
    revoked_at          timestamptz,
    revoked_by_account_id text,
    revocation_reason   text CHECK (revocation_reason IS NULL OR char_length(btrim(revocation_reason)) BETWEEN 1 AND 500),
    version             bigint NOT NULL DEFAULT 0 CHECK (version >= 0),
	CHECK (expires_at IS NULL OR expires_at > granted_at),
	CHECK (
		(status = 'active' AND revoked_at IS NULL AND revoked_by_account_id IS NULL AND revocation_reason IS NULL)
		OR (status = 'revoked' AND revoked_at IS NOT NULL AND revoked_by_account_id IS NOT NULL AND revocation_reason IS NOT NULL)
		OR (status = 'expired' AND expires_at IS NOT NULL AND revoked_at IS NULL AND revoked_by_account_id IS NULL AND revocation_reason IS NULL)
	),
    UNIQUE (tenant_id, id),
    FOREIGN KEY (tenant_id, grantee_account_id) REFERENCES accounts(tenant_id, id),
    FOREIGN KEY (tenant_id, granted_by_account_id) REFERENCES accounts(tenant_id, id),
    FOREIGN KEY (tenant_id, revoked_by_account_id) REFERENCES accounts(tenant_id, id)
);

CREATE UNIQUE INDEX IF NOT EXISTS capability_delegations_one_active_bundle
    ON capability_delegations (tenant_id, grantee_account_id, bundle)
    WHERE status = 'active';

CREATE TABLE IF NOT EXISTS students (
    id                  text PRIMARY KEY,
    tenant_id           text NOT NULL REFERENCES tenants(id),
    person_id           text NOT NULL,
    membership_id       text NOT NULL,
    account_id          text NOT NULL,
    enrollment_reference text NOT NULL CHECK (char_length(btrim(enrollment_reference)) BETWEEN 1 AND 100),
    status              text NOT NULL CHECK (status IN ('active', 'paused', 'graduated', 'archived')),
    locale              text NOT NULL CHECK (char_length(locale) BETWEEN 1 AND 35),
    timezone            text NOT NULL CHECK (char_length(timezone) BETWEEN 1 AND 100),
    adult_confirmed     boolean NOT NULL CHECK (adult_confirmed),
    created_at          timestamptz NOT NULL,
    updated_at          timestamptz NOT NULL,
    version             bigint NOT NULL DEFAULT 0 CHECK (version >= 0),
    UNIQUE (tenant_id, id),
    UNIQUE (tenant_id, enrollment_reference),
    UNIQUE (tenant_id, account_id),
    UNIQUE (tenant_id, id, account_id),
    FOREIGN KEY (tenant_id, person_id) REFERENCES people(tenant_id, id),
    FOREIGN KEY (tenant_id, membership_id) REFERENCES school_memberships(tenant_id, id),
    FOREIGN KEY (tenant_id, account_id) REFERENCES accounts(tenant_id, id)
);

CREATE TABLE IF NOT EXISTS teacher_assignments (
    id                  text PRIMARY KEY,
    tenant_id           text NOT NULL REFERENCES tenants(id),
    student_id          text NOT NULL,
    teacher_account_id  text NOT NULL,
    status              text NOT NULL CHECK (status IN ('active', 'ended')),
    assigned_by_account_id text NOT NULL,
    assigned_at         timestamptz NOT NULL,
    ended_at            timestamptz,
	CHECK (
		(status = 'active' AND ended_at IS NULL)
		OR (status = 'ended' AND ended_at IS NOT NULL)
	),
    UNIQUE (tenant_id, id),
    FOREIGN KEY (tenant_id, student_id) REFERENCES students(tenant_id, id),
    FOREIGN KEY (tenant_id, teacher_account_id) REFERENCES accounts(tenant_id, id),
    FOREIGN KEY (tenant_id, assigned_by_account_id) REFERENCES accounts(tenant_id, id)
);

CREATE UNIQUE INDEX IF NOT EXISTS teacher_assignments_one_active_primary
    ON teacher_assignments (tenant_id, student_id)
    WHERE status = 'active';

CREATE TABLE IF NOT EXISTS first_minute_revisions (
    id                  text PRIMARY KEY,
    tenant_id           text NOT NULL REFERENCES tenants(id),
    student_id          text NOT NULL,
    revision            bigint NOT NULL CHECK (revision > 0),
    what_worked         text NOT NULL CHECK (char_length(btrim(what_worked)) BETWEEN 1 AND 500),
    current_focus       text NOT NULL CHECK (char_length(btrim(current_focus)) BETWEEN 1 AND 500),
    next_step           text NOT NULL CHECK (char_length(btrim(next_step)) BETWEEN 1 AND 500),
    authored_by_account_id text NOT NULL,
    published_at        timestamptz NOT NULL,
    UNIQUE (tenant_id, student_id, revision),
    FOREIGN KEY (tenant_id, student_id) REFERENCES students(tenant_id, id),
    FOREIGN KEY (tenant_id, authored_by_account_id) REFERENCES accounts(tenant_id, id)
);

CREATE TABLE IF NOT EXISTS activation_invitations (
    id                      text PRIMARY KEY,
    tenant_id               text NOT NULL REFERENCES tenants(id),
    account_id              text NOT NULL,
    student_id              text,
    kind                    text NOT NULL CHECK (kind IN ('owner_bootstrap', 'staff_activation', 'student_activation')),
    token_digest            bytea NOT NULL UNIQUE CHECK (octet_length(token_digest) = 32),
    status                  text NOT NULL CHECK (status IN ('issued', 'consumed', 'revoked', 'superseded')),
    issued_by_account_id    text,
    issued_at               timestamptz NOT NULL,
    expires_at              timestamptz NOT NULL,
    consumed_at             timestamptz,
    revoked_at              timestamptz,
    superseded_by_id        text,
    consumed_idempotency_key text CHECK (consumed_idempotency_key IS NULL OR char_length(consumed_idempotency_key) BETWEEN 1 AND 128),
    consumed_payload_fingerprint bytea CHECK (consumed_payload_fingerprint IS NULL OR octet_length(consumed_payload_fingerprint) = 32),
    CHECK ((kind = 'student_activation') = (student_id IS NOT NULL)),
	CHECK (expires_at > issued_at),
	CHECK ((consumed_idempotency_key IS NULL) = (consumed_payload_fingerprint IS NULL)),
	CHECK ((status = 'consumed') = (consumed_at IS NOT NULL)),
	CHECK ((status = 'consumed') = (consumed_idempotency_key IS NOT NULL)),
	CHECK ((status = 'revoked') = (revoked_at IS NOT NULL)),
	CHECK (superseded_by_id IS NULL OR status = 'superseded'),
    CHECK (superseded_by_id IS NULL OR kind = 'student_activation'),
    UNIQUE (tenant_id, id),
    UNIQUE (tenant_id, account_id, student_id, kind, id),
    FOREIGN KEY (tenant_id, account_id) REFERENCES accounts(tenant_id, id),
    FOREIGN KEY (tenant_id, student_id) REFERENCES students(tenant_id, id),
    FOREIGN KEY (tenant_id, student_id, account_id) REFERENCES students(tenant_id, id, account_id),
    FOREIGN KEY (tenant_id, issued_by_account_id) REFERENCES accounts(tenant_id, id),
    FOREIGN KEY (tenant_id, account_id, student_id, kind, superseded_by_id)
        REFERENCES activation_invitations(tenant_id, account_id, student_id, kind, id)
);

CREATE UNIQUE INDEX IF NOT EXISTS activation_invitations_one_issued_student
    ON activation_invitations (tenant_id, student_id)
    WHERE status = 'issued' AND student_id IS NOT NULL;

CREATE TABLE IF NOT EXISTS sessions (
    id                  text PRIMARY KEY,
    family_id           text NOT NULL,
    tenant_id           text NOT NULL REFERENCES tenants(id),
    account_id          text NOT NULL,
    access_digest       bytea NOT NULL UNIQUE CHECK (octet_length(access_digest) = 32),
    refresh_digest      bytea NOT NULL UNIQUE CHECK (octet_length(refresh_digest) = 32),
    status              text NOT NULL CHECK (status IN ('active', 'replaced', 'revoked')),
    access_expires_at   timestamptz NOT NULL,
    refresh_expires_at  timestamptz NOT NULL,
    created_at          timestamptz NOT NULL,
    replaced_by_id      text,
    revoked_at          timestamptz,
	CHECK (access_expires_at > created_at),
	CHECK (refresh_expires_at > access_expires_at),
	CHECK (
		(status IN ('active', 'replaced') AND revoked_at IS NULL)
		OR (status = 'revoked' AND revoked_at IS NOT NULL)
	),
	CHECK (replaced_by_id IS NULL OR status IN ('replaced', 'revoked')),
    UNIQUE (tenant_id, id),
    UNIQUE (tenant_id, account_id, family_id, id),
    FOREIGN KEY (tenant_id, account_id) REFERENCES accounts(tenant_id, id),
    FOREIGN KEY (tenant_id, account_id, family_id, replaced_by_id)
        REFERENCES sessions(tenant_id, account_id, family_id, id)
);

CREATE INDEX IF NOT EXISTS sessions_family_idx ON sessions (tenant_id, account_id, family_id);

CREATE TABLE IF NOT EXISTS idempotency_records (
    tenant_id           text NOT NULL REFERENCES tenants(id),
    actor_account_id    text NOT NULL,
    operation_scope     text NOT NULL CHECK (char_length(operation_scope) BETWEEN 1 AND 100),
    idempotency_key     text NOT NULL CHECK (char_length(idempotency_key) BETWEEN 1 AND 128),
    payload_fingerprint bytea NOT NULL CHECK (octet_length(payload_fingerprint) = 32),
    status              text NOT NULL CHECK (status IN ('processing', 'completed')),
    response_json       jsonb,
    created_at          timestamptz NOT NULL,
    completed_at        timestamptz,
	CHECK (
		(status = 'processing' AND response_json IS NULL AND completed_at IS NULL)
		OR (status = 'completed' AND response_json IS NOT NULL AND completed_at IS NOT NULL)
	),
    PRIMARY KEY (tenant_id, actor_account_id, operation_scope, idempotency_key),
    FOREIGN KEY (tenant_id, actor_account_id) REFERENCES accounts(tenant_id, id)
);

CREATE TABLE IF NOT EXISTS audit_records (
    id                  bigserial PRIMARY KEY,
    tenant_id           text NOT NULL REFERENCES tenants(id),
    actor_account_id    text,
    operator_identifier text CHECK (operator_identifier IS NULL OR char_length(btrim(operator_identifier)) BETWEEN 1 AND 200),
    delegation_id       text,
    action              text NOT NULL,
    target_type         text NOT NULL,
    target_id           text NOT NULL,
    decision            text NOT NULL CHECK (decision IN ('allow', 'deny')),
    reason_code         text,
    correlation_id      text,
    idempotency_key_hash text,
    metadata            jsonb NOT NULL DEFAULT '{}'::jsonb,
    recorded_at         timestamptz NOT NULL,
    FOREIGN KEY (tenant_id, actor_account_id) REFERENCES accounts(tenant_id, id),
    FOREIGN KEY (tenant_id, delegation_id) REFERENCES capability_delegations(tenant_id, id)
);

CREATE TABLE IF NOT EXISTS outbox_events (
    id                  bigserial PRIMARY KEY,
    tenant_id           text NOT NULL REFERENCES tenants(id),
    event_type          text NOT NULL,
    aggregate_type      text NOT NULL,
    aggregate_id        text NOT NULL,
    payload             jsonb NOT NULL,
    recorded_at         timestamptz NOT NULL,
    published_at        timestamptz,
    attempt_count       integer NOT NULL DEFAULT 0
);

CREATE OR REPLACE FUNCTION reject_audit_mutation() RETURNS trigger AS $$
BEGIN
    RAISE EXCEPTION 'audit_records are append-only';
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS audit_records_append_only ON audit_records;
CREATE TRIGGER audit_records_append_only
    BEFORE UPDATE OR DELETE ON audit_records
    FOR EACH ROW EXECUTE FUNCTION reject_audit_mutation();
