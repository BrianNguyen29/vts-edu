// Package storage defines the storage provider seam and a local-disk
// implementation for development and E2E tests. Production adapters
// (Supabase/S3-compatible) can be plugged in behind the same interface.
package storage

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Provider abstracts object storage operations.
type Provider interface {
	// Store persists the object under a generated key and returns that key.
	Store(ctx context.Context, r io.Reader, size int64, contentType string) (key string, err error)
	// Retrieve returns a ReadCloser for the object identified by key.
	Retrieve(ctx context.Context, key string) (io.ReadCloser, error)
	// Delete removes the object identified by key.
	Delete(ctx context.Context, key string) error
}

// LocalProvider stores objects on the local filesystem.
// It is intended for local development and E2E tests only.
type LocalProvider struct {
	BaseDir string
}

// NewLocalProvider creates a local storage provider rooted at baseDir.
func NewLocalProvider(baseDir string) (*LocalProvider, error) {
	if baseDir == "" {
		baseDir = "/tmp/vts-edu-resources"
	}
	abs, err := filepath.Abs(baseDir)
	if err != nil {
		return nil, fmt.Errorf("resolve storage path: %w", err)
	}
	if err := os.MkdirAll(abs, 0o750); err != nil {
		return nil, fmt.Errorf("create storage dir: %w", err)
	}
	return &LocalProvider{BaseDir: abs}, nil
}

// Store writes the object to a new random key under BaseDir.
func (p *LocalProvider) Store(ctx context.Context, r io.Reader, size int64, contentType string) (string, error) {
	key, err := generateKey()
	if err != nil {
		return "", fmt.Errorf("generate storage key: %w", err)
	}
	path := p.objectPath(key)

	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return "", fmt.Errorf("create object dir: %w", err)
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0o640)
	if err != nil {
		return "", fmt.Errorf("create object file: %w", err)
	}

	written, err := io.Copy(f, r)
	if closeErr := f.Close(); closeErr != nil && err == nil {
		err = closeErr
	}
	if err != nil {
		_ = os.Remove(path)
		return "", fmt.Errorf("write object: %w", err)
	}
	if written != size {
		_ = os.Remove(path)
		return "", fmt.Errorf("size mismatch: expected %d, wrote %d", size, written)
	}

	return key, nil
}

// Retrieve opens the object identified by key.
func (p *LocalProvider) Retrieve(ctx context.Context, key string) (io.ReadCloser, error) {
	if !isSafeKey(key) {
		return nil, fmt.Errorf("invalid storage key")
	}
	f, err := os.Open(p.objectPath(key))
	if err != nil {
		return nil, fmt.Errorf("open object: %w", err)
	}
	return f, nil
}

// Delete removes the object identified by key.
func (p *LocalProvider) Delete(ctx context.Context, key string) error {
	if !isSafeKey(key) {
		return fmt.Errorf("invalid storage key")
	}
	if err := os.Remove(p.objectPath(key)); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("delete object: %w", err)
	}
	return nil
}

func (p *LocalProvider) objectPath(key string) string {
	// Spread objects across subdirectories for filesystem efficiency.
	if len(key) >= 4 {
		return filepath.Join(p.BaseDir, key[:2], key[2:4], key)
	}
	return filepath.Join(p.BaseDir, key)
}

// isSafeKey rejects keys that could escape the storage root.
func isSafeKey(key string) bool {
	if key == "" {
		return false
	}
	for _, r := range key {
		if (r >= 'a' && r <= 'f') || (r >= '0' && r <= '9') {
			continue
		}
		return false
	}
	return true
}

func generateKey() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(b[:]), nil
}
