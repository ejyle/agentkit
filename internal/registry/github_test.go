package registry_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ejyle/agentkit/internal/domain"
	"github.com/ejyle/agentkit/internal/registry"
)

// sampleManifest returns a test manifest with two packages.
func sampleManifest() domain.Manifest {
	return domain.Manifest{
		Packages: []domain.Package{
			{
				Name:        "playwright",
				Version:     "1.2.0",
				Description: "Browser automation and E2E testing MCP server",
				Type:        domain.PackageTypeMCP,
				Source:      "github.com/microsoft/playwright-mcp",
			},
			{
				Name:        "play-sounds",
				Version:     "0.1.0",
				Description: "Play audio files skill",
				Type:        domain.PackageTypeSkill,
				Source:      "github.com/example/play-sounds",
			},
		},
	}
}

func marshalManifest(t *testing.T, m domain.Manifest) []byte {
	t.Helper()
	data, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("marshal manifest: %v", err)
	}
	return data
}

// Test 1: GitHubManifestRegistry fetches and parses manifest correctly.
func TestGitHubManifestRegistry_FetchesManifest(t *testing.T) {
	manifest := sampleManifest()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(marshalManifest(t, manifest))
	}))
	defer srv.Close()

	cacheDir := t.TempDir()
	reg := registry.NewGitHubManifestRegistryWithCache("test-reg", srv.URL+"/registry.json",
		filepath.Join(cacheDir, "manifest.json"))

	pkgs, err := reg.Search("")
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(pkgs) != len(manifest.Packages) {
		t.Errorf("expected %d packages, got %d", len(manifest.Packages), len(pkgs))
	}
}

// Test 2: 304 Not Modified returns cached manifest without parsing body.
func TestGitHubManifestRegistry_ETag304(t *testing.T) {
	manifest := sampleManifest()
	var callCount atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount.Add(1)
		if r.Header.Get("If-None-Match") == `"etag-v1"` {
			w.WriteHeader(http.StatusNotModified)
			return
		}
		w.Header().Set("ETag", `"etag-v1"`)
		w.Header().Set("Content-Type", "application/json")
		w.Write(marshalManifest(t, manifest))
	}))
	defer srv.Close()

	cacheDir := t.TempDir()
	cachePath := filepath.Join(cacheDir, "manifest.json")
	reg := registry.NewGitHubManifestRegistryWithCache("test-reg", srv.URL+"/registry.json", cachePath)

	// First call — fetches and caches.
	if _, err := reg.Search(""); err != nil {
		t.Fatalf("first Search: %v", err)
	}

	// Second call — should send If-None-Match and accept 304.
	pkgs, err := reg.Search("")
	if err != nil {
		t.Fatalf("second Search: %v", err)
	}
	if len(pkgs) != len(manifest.Packages) {
		t.Errorf("expected %d packages from cache after 304, got %d", len(manifest.Packages), len(pkgs))
	}
	if callCount.Load() != 2 {
		t.Errorf("expected 2 HTTP calls, got %d", callCount.Load())
	}
}

// Test 3: Network failure with cache — Resolve() returns cached result, no error.
func TestGitHubManifestRegistry_NetworkFailureFallsBackToCache(t *testing.T) {
	manifest := sampleManifest()

	// Start a server, fetch once to populate cache, then shut it down.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(marshalManifest(t, manifest))
	}))

	cacheDir := t.TempDir()
	cachePath := filepath.Join(cacheDir, "manifest.json")
	reg := registry.NewGitHubManifestRegistryWithCache("test-reg", srv.URL+"/registry.json", cachePath)

	// Populate cache.
	if _, err := reg.Search(""); err != nil {
		t.Fatalf("initial Search: %v", err)
	}
	srv.Close()

	// After server shutdown, should fall back to stale cache.
	pkg, err := reg.Resolve("playwright")
	if err != nil {
		t.Fatalf("expected stale cache fallback, got error: %v", err)
	}
	if pkg == nil || pkg.Name != "playwright" {
		t.Errorf("expected playwright package from cache, got %v", pkg)
	}
}

// Test 4: Network failure with no cache — Resolve() returns error containing "registry unreachable".
func TestGitHubManifestRegistry_NetworkFailureNoCache(t *testing.T) {
	// Use a URL that will never connect.
	reg := registry.NewGitHubManifestRegistryWithCache("test-reg",
		"http://127.0.0.1:1", // nothing listening here
		filepath.Join(t.TempDir(), "manifest.json"))

	_, err := reg.Resolve("playwright")
	if err == nil {
		t.Fatal("expected error when network fails and no cache exists")
	}
	if !containsSubstring(err.Error(), "registry unreachable") {
		t.Errorf("error should contain 'registry unreachable', got: %v", err)
	}
}

// Test 5: RegistryManager.Resolve returns correct Package from first matching registry.
func TestRegistryManager_Resolve(t *testing.T) {
	manifest := sampleManifest()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(marshalManifest(t, manifest))
	}))
	defer srv.Close()

	cacheDir := t.TempDir()
	reg := registry.NewGitHubManifestRegistryWithCache("test-reg", srv.URL+"/registry.json",
		filepath.Join(cacheDir, "manifest.json"))

	mgr := registry.NewRegistryManagerWithRegistries(reg)
	pkg, err := mgr.Resolve("playwright")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if pkg == nil || pkg.Name != "playwright" {
		t.Errorf("expected playwright, got %v", pkg)
	}
}

// Test 6: RegistryManager.Search returns results ranked: exact name first, then fuzzy, then description.
func TestRegistryManager_SearchRanking(t *testing.T) {
	manifest := sampleManifest()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(marshalManifest(t, manifest))
	}))
	defer srv.Close()

	cacheDir := t.TempDir()
	reg := registry.NewGitHubManifestRegistryWithCache("test-reg", srv.URL+"/registry.json",
		filepath.Join(cacheDir, "manifest.json"))
	mgr := registry.NewRegistryManagerWithRegistries(reg)

	results, err := mgr.Search("playwright")
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected at least one result")
	}
	// Exact name match must come first.
	if results[0].Package.Name != "playwright" {
		t.Errorf("expected playwright first (exact match), got %s", results[0].Package.Name)
	}
	// Exact match score must be 100.
	if results[0].Score != 100 {
		t.Errorf("expected exact match score 100, got %d", results[0].Score)
	}
}

// Test 7: Search ranking is deterministic — same query always returns same order.
func TestRegistryManager_SearchDeterministic(t *testing.T) {
	manifest := sampleManifest()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(marshalManifest(t, manifest))
	}))
	defer srv.Close()

	cacheDir := t.TempDir()
	reg := registry.NewGitHubManifestRegistryWithCache("test-reg", srv.URL+"/registry.json",
		filepath.Join(cacheDir, "manifest.json"))
	mgr := registry.NewRegistryManagerWithRegistries(reg)

	run1, err := mgr.Search("play")
	if err != nil {
		t.Fatalf("first Search: %v", err)
	}
	run2, err := mgr.Search("play")
	if err != nil {
		t.Fatalf("second Search: %v", err)
	}
	if len(run1) != len(run2) {
		t.Fatalf("run1 has %d, run2 has %d", len(run1), len(run2))
	}
	for i := range run1 {
		if run1[i].Package.Name != run2[i].Package.Name {
			t.Errorf("result[%d] differs: %s vs %s", i, run1[i].Package.Name, run2[i].Package.Name)
		}
	}
}

// Test 8: HTTP client enforces a short timeout (tested via a deliberately slow server).
func TestGitHubManifestRegistry_Timeout(t *testing.T) {
	slow := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Sleep longer than the expected response timeout.
		time.Sleep(15 * time.Second)
	}))
	defer slow.Close()

	reg := registry.NewGitHubManifestRegistryWithCacheAndTimeout(
		"test-reg",
		slow.URL+"/registry.json",
		filepath.Join(t.TempDir(), "manifest.json"),
		200*time.Millisecond, // override for tests
		200*time.Millisecond,
	)

	start := time.Now()
	_, err := reg.Resolve("anything")
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected timeout error")
	}
	if elapsed > 5*time.Second {
		t.Errorf("expected timeout well under 5s, took %v", elapsed)
	}
}

func containsSubstring(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 ||
		func() bool {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
			return false
		}())
}
