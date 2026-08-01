package migrations

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestMigrationLedgerRepeatConcurrencyAndDrift(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	pool := isolatedMigrationPool(t, ctx, databaseURL)

	const callers = 6
	var wait sync.WaitGroup
	errorsByCaller := make(chan error, callers)
	wait.Add(callers)
	for index := 0; index < callers; index++ {
		go func() {
			defer wait.Done()
			errorsByCaller <- Up(ctx, pool)
		}()
	}
	wait.Wait()
	close(errorsByCaller)
	for err := range errorsByCaller {
		if err != nil {
			t.Fatalf("concurrent migration: %v", err)
		}
	}
	if err := Up(ctx, pool); err != nil {
		t.Fatalf("repeat migration: %v", err)
	}
	var ledgerCount int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM schema_migrations WHERE version = $1`, initialVersion).Scan(&ledgerCount); err != nil {
		t.Fatalf("count migration ledger: %v", err)
	}
	if ledgerCount != 1 {
		t.Fatalf("migration ledger rows = %d, want 1", ledgerCount)
	}
	if _, err := pool.Exec(ctx, `UPDATE schema_migrations SET checksum = $2 WHERE version = $1`, initialVersion, strings.Repeat("0", 64)); err != nil {
		t.Fatalf("corrupt migration checksum for drift test: %v", err)
	}
	if err := Up(ctx, pool); err == nil || !strings.Contains(err.Error(), "checksum drift") {
		t.Fatalf("migration checksum drift error = %v", err)
	}
}

func isolatedMigrationPool(t *testing.T, ctx context.Context, databaseURL string) *pgxpool.Pool {
	t.Helper()
	adminPool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("open migration admin pool: %v", err)
	}
	if err := adminPool.Ping(ctx); err != nil {
		adminPool.Close()
		t.Fatalf("ping migration database: %v", err)
	}
	random := make([]byte, 12)
	if _, err := rand.Read(random); err != nil {
		adminPool.Close()
		t.Fatalf("generate migration schema: %v", err)
	}
	schema := "migration_test_" + hex.EncodeToString(random)
	identifier := pgx.Identifier{schema}.Sanitize()
	if _, err := adminPool.Exec(ctx, "CREATE SCHEMA "+identifier); err != nil {
		adminPool.Close()
		t.Fatalf("create migration schema: %v", err)
	}
	configuration, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		_, _ = adminPool.Exec(ctx, "DROP SCHEMA "+identifier+" CASCADE")
		adminPool.Close()
		t.Fatalf("parse migration database URL: %v", err)
	}
	configuration.ConnConfig.RuntimeParams["search_path"] = schema
	pool, err := pgxpool.NewWithConfig(ctx, configuration)
	if err != nil {
		_, _ = adminPool.Exec(ctx, "DROP SCHEMA "+identifier+" CASCADE")
		adminPool.Close()
		t.Fatalf("open migration test pool: %v", err)
	}
	t.Cleanup(func() {
		pool.Close()
		cleanupContext, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		_, _ = adminPool.Exec(cleanupContext, "DROP SCHEMA "+identifier+" CASCADE")
		adminPool.Close()
	})
	return pool
}
