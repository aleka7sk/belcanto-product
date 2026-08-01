package postgres

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestAdvisoryLockKeyIsPostgreSQLTextSafeAndTupleFramed(t *testing.T) {
	key := advisoryLockKey("activation", "tenant\x00with-separator", "student:value")
	if !utf8.ValidString(key) || strings.ContainsRune(key, '\x00') {
		t.Fatalf("advisory key is not PostgreSQL text safe: %q", key)
	}
	if key != advisoryLockKey("activation", "tenant\x00with-separator", "student:value") {
		t.Fatal("advisory key is not deterministic")
	}
	if advisoryLockKey("activation", "ab", "c") == advisoryLockKey("activation", "a", "bc") {
		t.Fatal("length-framed tuples collided")
	}
	if advisoryLockKey("activation", "tenant", "subject") == advisoryLockKey("staff-bootstrap", "tenant", "subject") {
		t.Fatal("lock namespaces collided")
	}
}
