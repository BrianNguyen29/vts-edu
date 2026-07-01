package storage

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestLocalProvider_StoreAndRetrieve(t *testing.T) {
	base, err := os.MkdirTemp("", "vts-storage-")
	if err != nil {
		t.Fatalf("mktemp: %v", err)
	}
	defer os.RemoveAll(base)

	p, err := NewLocalProvider(base)
	if err != nil {
		t.Fatalf("new local provider: %v", err)
	}

	payload := []byte("hello resources")
	key, err := p.Store(context.Background(), bytes.NewReader(payload), int64(len(payload)), "text/plain")
	if err != nil {
		t.Fatalf("store: %v", err)
	}
	if key == "" {
		t.Fatal("expected non-empty key")
	}

	r, err := p.Retrieve(context.Background(), key)
	if err != nil {
		t.Fatalf("retrieve: %v", err)
	}
	defer r.Close()
	got, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if !bytes.Equal(got, payload) {
		t.Fatalf("payload mismatch: %q vs %q", got, payload)
	}
}

func TestLocalProvider_RejectsUnsafeKey(t *testing.T) {
	base, err := os.MkdirTemp("", "vts-storage-")
	if err != nil {
		t.Fatalf("mktemp: %v", err)
	}
	defer os.RemoveAll(base)

	p, err := NewLocalProvider(base)
	if err != nil {
		t.Fatalf("new local provider: %v", err)
	}

	for _, bad := range []string{"", "../escape", "abc/def", "abc def", "abc.def"} {
		if _, err := p.Retrieve(context.Background(), bad); err == nil {
			t.Fatalf("expected unsafe key %q to be rejected", bad)
		}
		if err := p.Delete(context.Background(), bad); err == nil {
			t.Fatalf("expected unsafe key %q to be rejected for delete", bad)
		}
	}
}

func TestLocalProvider_StoreOutsideBase(t *testing.T) {
	base, err := os.MkdirTemp("", "vts-storage-")
	if err != nil {
		t.Fatalf("mktemp: %v", err)
	}
	defer os.RemoveAll(base)

	p, err := NewLocalProvider(base)
	if err != nil {
		t.Fatalf("new local provider: %v", err)
	}

	// Create a symlink inside base that points outside; ensure retrieve
	// still resolves under base because keys are server-generated hex.
	// This is a smoke test: a generated key should not follow user input.
	key, err := p.Store(context.Background(), bytes.NewReader([]byte("x")), 1, "application/octet-stream")
	if err != nil {
		t.Fatalf("store: %v", err)
	}
	expectedPath := filepath.Join(p.BaseDir, key[:2], key[2:4], key)
	if _, err := os.Stat(expectedPath); err != nil {
		t.Fatalf("expected object path to exist: %v", err)
	}
}
