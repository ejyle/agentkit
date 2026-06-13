package config_test

import (
	"path/filepath"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/ejyle/agentkit/internal/config"
	"github.com/ejyle/agentkit/internal/domain"
)

func makeRecord(name string) domain.InstalledRecord {
	return domain.InstalledRecord{
		Name:        name,
		Version:     "1.0.0",
		Type:        domain.PackageTypeMCP,
		InstallPath: "mcpServers." + name,
		InstalledAt: time.Now().UTC().Truncate(time.Second),
		SourceURL:   "https://example.com/registry.json",
		Checksum:    "sha256:abc123",
	}
}

// Test 1: RecordInstalled writes and re-reads the same record.
func TestRecordInstalled_WriteAndRead(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "installed.json")
	store := config.NewConfigStoreWithPath("claude", path)

	rec := makeRecord("playwright")
	if err := store.RecordInstalled(rec); err != nil {
		t.Fatalf("RecordInstalled: %v", err)
	}

	got, ok, err := store.GetRecord("playwright")
	if err != nil {
		t.Fatalf("GetRecord: %v", err)
	}
	if !ok {
		t.Fatal("expected record to be present")
	}
	if got.Name != rec.Name || got.Version != rec.Version || got.Type != rec.Type {
		t.Errorf("record mismatch: got %+v, want %+v", got, rec)
	}
}

// Test 2: RecordInstalled auto-creates a non-existent directory.
func TestRecordInstalled_AutoCreatesDir(t *testing.T) {
	dir := t.TempDir()
	// Use a nested path that does not yet exist.
	path := filepath.Join(dir, "nested", "deep", "installed.json")
	store := config.NewConfigStoreWithPath("claude", path)

	if err := store.RecordInstalled(makeRecord("playwright")); err != nil {
		t.Fatalf("RecordInstalled should auto-create dir, got: %v", err)
	}
}

// Test 3: RemoveRecord removes the entry; ListInstalled does not contain it.
func TestRemoveRecord(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "installed.json")
	store := config.NewConfigStoreWithPath("claude", path)

	if err := store.RecordInstalled(makeRecord("playwright")); err != nil {
		t.Fatalf("RecordInstalled: %v", err)
	}
	if err := store.RemoveRecord("playwright"); err != nil {
		t.Fatalf("RemoveRecord: %v", err)
	}

	list, err := store.ListInstalled()
	if err != nil {
		t.Fatalf("ListInstalled: %v", err)
	}
	for _, r := range list {
		if r.Name == "playwright" {
			t.Error("expected playwright to be removed from list")
		}
	}
}

// Test 4: ListInstalled on empty/missing file returns empty slice, no error.
func TestListInstalled_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "installed.json")
	store := config.NewConfigStoreWithPath("claude", path)

	list, err := store.ListInstalled()
	if err != nil {
		t.Fatalf("ListInstalled on missing file should not error: %v", err)
	}
	if len(list) != 0 {
		t.Errorf("expected empty slice, got %d entries", len(list))
	}
}

// Test 5: RecordInstalled is idempotent — second call with same name updates, no duplicates.
func TestRecordInstalled_Idempotent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "installed.json")
	store := config.NewConfigStoreWithPath("claude", path)

	rec := makeRecord("playwright")
	if err := store.RecordInstalled(rec); err != nil {
		t.Fatalf("first RecordInstalled: %v", err)
	}
	rec.Version = "2.0.0"
	if err := store.RecordInstalled(rec); err != nil {
		t.Fatalf("second RecordInstalled: %v", err)
	}

	list, err := store.ListInstalled()
	if err != nil {
		t.Fatalf("ListInstalled: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("expected 1 entry, got %d", len(list))
	}
	if list[0].Version != "2.0.0" {
		t.Errorf("expected updated version 2.0.0, got %s", list[0].Version)
	}
}

// Test 6: GetRecord returns record when present, (empty, false) when absent.
func TestGetRecord(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "installed.json")
	store := config.NewConfigStoreWithPath("claude", path)

	// Missing record returns false, no error.
	_, ok, err := store.GetRecord("playwright")
	if err != nil {
		t.Fatalf("GetRecord on missing: %v", err)
	}
	if ok {
		t.Fatal("expected ok=false for missing record")
	}

	// After recording, returns true.
	if err := store.RecordInstalled(makeRecord("playwright")); err != nil {
		t.Fatalf("RecordInstalled: %v", err)
	}
	_, ok, err = store.GetRecord("playwright")
	if err != nil {
		t.Fatalf("GetRecord after record: %v", err)
	}
	if !ok {
		t.Fatal("expected ok=true after recording")
	}
}

// Test 7: Concurrent RecordInstalled calls produce valid JSON (no corruption).
func TestRecordInstalled_Concurrent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "installed.json")
	store := config.NewConfigStoreWithPath("claude", path)

	const n = 20
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(idx int) {
			defer wg.Done()
			name := "pkg-" + string(rune('a'+idx%26))
			_ = store.RecordInstalled(makeRecord(name))
		}(i)
	}
	wg.Wait()

	// File must be valid JSON — ListInstalled must not error.
	list, err := store.ListInstalled()
	if err != nil {
		t.Fatalf("ListInstalled after concurrent writes: %v", err)
	}
	// Results should be sorted by name.
	names := make([]string, len(list))
	for i, r := range list {
		names[i] = r.Name
	}
	if !sort.StringsAreSorted(names) {
		t.Errorf("ListInstalled should return sorted results, got: %v", names)
	}
}
