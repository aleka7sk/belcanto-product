package media

import (
	"bytes"
	"context"
	"testing"
)

func TestFSStorageResumableWriteAndRead(t *testing.T) {
	storage, err := NewFSStorage(t.TempDir())
	if err != nil {
		t.Fatalf("new fs storage: %v", err)
	}
	ctx := context.Background()
	payload := bytes.Repeat([]byte{0x42}, 96)
	if err := storage.WriteAt(ctx, "tenant_a/med_1", 0, payload[:64]); err != nil {
		t.Fatalf("first chunk: %v", err)
	}
	// A replayed chunk lands on the same bytes.
	if err := storage.WriteAt(ctx, "tenant_a/med_1", 0, payload[:64]); err != nil {
		t.Fatalf("replayed chunk: %v", err)
	}
	if err := storage.WriteAt(ctx, "tenant_a/med_1", 64, payload[64:]); err != nil {
		t.Fatalf("second chunk: %v", err)
	}
	content, err := storage.Read(ctx, "tenant_a/med_1")
	if err != nil || !bytes.Equal(content, payload) {
		t.Fatalf("read back = %d bytes, %v", len(content), err)
	}
}

func TestFSStorageRejectsTraversalKeys(t *testing.T) {
	storage, err := NewFSStorage(t.TempDir())
	if err != nil {
		t.Fatalf("new fs storage: %v", err)
	}
	ctx := context.Background()
	for _, key := range []string{"", "../escape", "/absolute", "tenant/../../escape"} {
		if err := storage.WriteAt(ctx, key, 0, []byte{1}); err == nil {
			t.Fatalf("key %q must be rejected", key)
		}
		if _, err := storage.Read(ctx, key); err == nil {
			t.Fatalf("read of key %q must be rejected", key)
		}
	}
}
