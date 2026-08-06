// Package media is the production storage adapter boundary for uploaded
// bytes. The database keeps only opaque storage keys — never URLs and
// never content — so access always flows through short-lived signed
// links served by the API.
package media

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Storage writes and reads media bytes by an opaque server-generated key.
// WriteAt must be idempotent for a replayed chunk (same offset, same
// data) so an interrupted upload can resume safely.
type Storage interface {
	WriteAt(ctx context.Context, key string, offset int64, data []byte) error
	Read(ctx context.Context, key string) ([]byte, error)
}

// MemoryStorage keeps bytes in process; the default for tests and for
// wiring that has not configured a durable adapter.
type MemoryStorage struct {
	mu      sync.Mutex
	objects map[string][]byte
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{objects: make(map[string][]byte)}
}

func (s *MemoryStorage) WriteAt(_ context.Context, key string, offset int64, data []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	current := s.objects[key]
	needed := offset + int64(len(data))
	if int64(len(current)) < needed {
		grown := make([]byte, needed)
		copy(grown, current)
		current = grown
	}
	copy(current[offset:], data)
	s.objects[key] = current
	return nil
}

func (s *MemoryStorage) Read(_ context.Context, key string) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	object, ok := s.objects[key]
	if !ok {
		return nil, fmt.Errorf("media object %q is not stored", key)
	}
	result := make([]byte, len(object))
	copy(result, object)
	return result, nil
}

// FSStorage stores each object as one file under a root directory.
type FSStorage struct {
	root string
}

func NewFSStorage(root string) (*FSStorage, error) {
	if root == "" {
		return nil, fmt.Errorf("media storage root is required")
	}
	if err := os.MkdirAll(root, 0o750); err != nil {
		return nil, fmt.Errorf("prepare media storage root: %w", err)
	}
	return &FSStorage{root: root}, nil
}

func (s *FSStorage) path(key string) (string, error) {
	if key == "" || strings.Contains(key, "..") || filepath.IsAbs(key) {
		return "", fmt.Errorf("invalid media storage key")
	}
	return filepath.Join(s.root, filepath.FromSlash(key)), nil
}

func (s *FSStorage) WriteAt(_ context.Context, key string, offset int64, data []byte) error {
	target, err := s.path(key)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o750); err != nil {
		return fmt.Errorf("prepare media directory: %w", err)
	}
	file, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, 0o640)
	if err != nil {
		return fmt.Errorf("open media object: %w", err)
	}
	defer file.Close()
	if _, err := file.WriteAt(data, offset); err != nil {
		return fmt.Errorf("write media chunk: %w", err)
	}
	return nil
}

func (s *FSStorage) Read(_ context.Context, key string) ([]byte, error) {
	target, err := s.path(key)
	if err != nil {
		return nil, err
	}
	content, err := os.ReadFile(target)
	if err != nil {
		return nil, fmt.Errorf("read media object: %w", err)
	}
	return content, nil
}
