package migrations

import (
	"context"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const initialVersion int64 = 1
const migrationLockKey = "belcanto:schema-migrations:v1"

const ledgerDDL = `
CREATE TABLE IF NOT EXISTS schema_migrations (
    version      bigint PRIMARY KEY CHECK (version > 0),
    checksum     text NOT NULL CHECK (checksum ~ '^[0-9a-f]{64}$'),
    description text NOT NULL CHECK (char_length(description) BETWEEN 1 AND 200),
    applied_at   timestamptz NOT NULL
)`

//go:embed 000001_initial.up.sql
var initialUp string

//go:embed 000001_initial.down.sql
var initialDown string

//go:embed 000002_internal_scheduling.up.sql
var internalSchedulingUp string

//go:embed 000002_internal_scheduling.down.sql
var internalSchedulingDown string

//go:embed 000003_session_security.up.sql
var sessionSecurityUp string

//go:embed 000003_session_security.down.sql
var sessionSecurityDown string

//go:embed 000004_contact_twofa.up.sql
var contactTwofaUp string

//go:embed 000004_contact_twofa.down.sql
var contactTwofaDown string

//go:embed 000005_policies_privacy.up.sql
var policiesPrivacyUp string

//go:embed 000005_policies_privacy.down.sql
var policiesPrivacyDown string

//go:embed 000006_rooms_core_lessons.up.sql
var roomsCoreLessonsUp string

//go:embed 000006_rooms_core_lessons.down.sql
var roomsCoreLessonsDown string

//go:embed 000007_events_rsvp.up.sql
var eventsRsvpUp string

//go:embed 000007_events_rsvp.down.sql
var eventsRsvpDown string

//go:embed 000008_reschedule_requests.up.sql
var rescheduleRequestsUp string

//go:embed 000008_reschedule_requests.down.sql
var rescheduleRequestsDown string

//go:embed 000009_journal_progress.up.sql
var journalProgressUp string

//go:embed 000009_journal_progress.down.sql
var journalProgressDown string

//go:embed 000010_homework_practice.up.sql
var homeworkPracticeUp string

//go:embed 000010_homework_practice.down.sql
var homeworkPracticeDown string

//go:embed 000011_attendance.up.sql
var attendanceUp string

//go:embed 000011_attendance.down.sql
var attendanceDown string

type migration struct {
	version     int64
	description string
	up          string
	down        string
}

func registeredMigrations() []migration {
	return []migration{
		{version: initialVersion, description: "Belcanto B.0 initial schema", up: initialUp, down: initialDown},
		{version: 2, description: "Belcanto L.1 internal scheduling", up: internalSchedulingUp, down: internalSchedulingDown},
		{version: 3, description: "Belcanto P.1 session security", up: sessionSecurityUp, down: sessionSecurityDown},
		{version: 4, description: "Belcanto P.1 contacts and two-factor authentication", up: contactTwofaUp, down: contactTwofaDown},
		{version: 5, description: "Belcanto P.1 policies, privacy and data rights", up: policiesPrivacyUp, down: policiesPrivacyDown},
		{version: 6, description: "Belcanto L.2 rooms and core lessons", up: roomsCoreLessonsUp, down: roomsCoreLessonsDown},
		{version: 7, description: "Belcanto L.2 events and RSVP", up: eventsRsvpUp, down: eventsRsvpDown},
		{version: 8, description: "Belcanto L.2 reschedule requests", up: rescheduleRequestsUp, down: rescheduleRequestsDown},
		{version: 9, description: "Belcanto L.3 journals and progress", up: journalProgressUp, down: journalProgressDown},
		{version: 10, description: "Belcanto L.3 homework and practice", up: homeworkPracticeUp, down: homeworkPracticeDown},
		{version: 11, description: "Belcanto L.4 lesson attendance", up: attendanceUp, down: attendanceDown},
	}
}

func Up(ctx context.Context, pool *pgxpool.Pool) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin migration transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1, 0))`, migrationLockKey); err != nil {
		return fmt.Errorf("lock migration ledger: %w", err)
	}
	if _, err := tx.Exec(ctx, ledgerDDL); err != nil {
		return fmt.Errorf("create migration ledger: %w", err)
	}
	if _, err := tx.Exec(ctx, `SELECT version, checksum, description, applied_at FROM schema_migrations LIMIT 0`); err != nil {
		return fmt.Errorf("migration ledger schema_migrations has an incompatible shape (does DATABASE_URL point at a database owned by another application?): %w", err)
	}
	for _, candidate := range registeredMigrations() {
		digest := sha256.Sum256([]byte(candidate.up))
		wantChecksum := hex.EncodeToString(digest[:])
		var storedChecksum string
		err = tx.QueryRow(ctx, `
			SELECT checksum FROM schema_migrations WHERE version = $1
		`, candidate.version).Scan(&storedChecksum)
		if err == nil {
			if storedChecksum != wantChecksum {
				return fmt.Errorf("migration %d checksum drift: database=%s binary=%s", candidate.version, storedChecksum, wantChecksum)
			}
			continue
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("read migration %d ledger: %w", candidate.version, err)
		}
		if _, err := tx.Exec(ctx, candidate.up); err != nil {
			return fmt.Errorf("apply migration %d: %w", candidate.version, err)
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO schema_migrations (version, checksum, description, applied_at)
			VALUES ($1, $2, $3, transaction_timestamp())
		`, candidate.version, wantChecksum, candidate.description); err != nil {
			return fmt.Errorf("record migration %d: %w", candidate.version, err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit initial migration: %w", err)
	}
	return nil
}

func Down(ctx context.Context, pool *pgxpool.Pool) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin migration rollback: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1, 0))`, migrationLockKey); err != nil {
		return fmt.Errorf("lock migration ledger: %w", err)
	}
	migrations := registeredMigrations()
	for index := len(migrations) - 1; index >= 0; index-- {
		if _, err := tx.Exec(ctx, migrations[index].down); err != nil {
			return fmt.Errorf("revert migration %d: %w", migrations[index].version, err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit migration rollback: %w", err)
	}
	return nil
}
