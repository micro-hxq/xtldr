package history

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAppendAndList(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "history.jsonl")
	store := &Store{path: path, limit: 200}

	if err := store.Append("list files", "/tmp/project"); err != nil {
		t.Fatalf("append failed: %v", err)
	}
	if err := store.Append("show git status", "/tmp/project"); err != nil {
		t.Fatalf("append failed: %v", err)
	}

	sessions, err := store.List()
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if len(sessions) != 2 {
		t.Fatalf("expected 2 sessions, got %d", len(sessions))
	}
	if sessions[0].Request != "show git status" {
		t.Fatalf("expected latest session first, got %q", sessions[0].Request)
	}
}

func TestSearch(t *testing.T) {
	sessions := []Session{
		{Request: "find large files", WorkingDir: "/a"},
		{Request: "git status", WorkingDir: "/repo"},
	}
	filtered := Search(sessions, "git")
	if len(filtered) != 1 || filtered[0].Request != "git status" {
		t.Fatalf("unexpected search result: %+v", filtered)
	}
}

func TestNewDefaultStoreFromEnv(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "custom-history.jsonl")
	t.Setenv(historyFileEnvKey, path)

	store, err := NewDefaultStore()
	if err != nil {
		t.Fatalf("new store failed: %v", err)
	}
	if store.path != path {
		t.Fatalf("expected path %q, got %q", path, store.path)
	}
}

func TestAppendDeduplicatesLatest(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "history.jsonl")
	store := &Store{path: path, limit: 200}

	if err := store.Append("git status", "/repo"); err != nil {
		t.Fatalf("append failed: %v", err)
	}
	if err := store.Append("ls -la", "/repo"); err != nil {
		t.Fatalf("append failed: %v", err)
	}
	if err := store.Append("git status", "/repo"); err != nil {
		t.Fatalf("append failed: %v", err)
	}

	sessions, err := store.List()
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if len(sessions) != 2 {
		t.Fatalf("expected 2 unique sessions, got %d", len(sessions))
	}
	if sessions[0].Request != "git status" {
		t.Fatalf("expected deduplicated request moved to top, got %q", sessions[0].Request)
	}
}

func TestListMissingFile(t *testing.T) {
	store := &Store{path: filepath.Join(t.TempDir(), "missing.jsonl"), limit: 200}
	sessions, err := store.List()
	if err != nil {
		t.Fatalf("expected nil error for missing file, got %v", err)
	}
	if len(sessions) != 0 {
		t.Fatalf("expected empty sessions, got %d", len(sessions))
	}
}

func TestAppendIgnoresEmptyRequest(t *testing.T) {
	store := &Store{path: filepath.Join(t.TempDir(), "history.jsonl"), limit: 200}
	if err := store.Append("   ", "/tmp"); err != nil {
		t.Fatalf("append failed: %v", err)
	}
	if _, err := os.Stat(store.path); !os.IsNotExist(err) {
		t.Fatalf("expected history file not created for empty request")
	}
}
