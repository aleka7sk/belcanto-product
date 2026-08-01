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
	digest := sha256.Sum256([]byte(initialUp))
	wantChecksum := hex.EncodeToString(digest[:])
	var storedChecksum string
	err = tx.QueryRow(ctx, `
		SELECT checksum FROM schema_migrations WHERE version = $1
	`, initialVersion).Scan(&storedChecksum)
	if err == nil {
		if storedChecksum != wantChecksum {
			return fmt.Errorf("migration %d checksum drift: database=%s binary=%s", initialVersion, storedChecksum, wantChecksum)
		}
		return tx.Commit(ctx)
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("read migration ledger: %w", err)
	}
	if _, err := tx.Exec(ctx, initialUp); err != nil {
		return fmt.Errorf("apply initial migration: %w", err)
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO schema_migrations (version, checksum, description, applied_at)
		VALUES ($1, $2, 'Belcanto B.0 initial schema', transaction_timestamp())
	`, initialVersion, wantChecksum); err != nil {
		return fmt.Errorf("record initial migration: %w", err)
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
	if _, err := tx.Exec(ctx, initialDown); err != nil {
		return fmt.Errorf("revert initial migration: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit migration rollback: %w", err)
	}
	return nil
}
