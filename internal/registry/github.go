package registry

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/ejyle/agentkit/internal/config"
	"github.com/ejyle/agentkit/internal/domain"
	retryablehttp "github.com/hashicorp/go-retryablehttp"
)

// GitHubManifestRegistry fetches a registry.json manifest from a GitHub raw URL.
// It caches manifests on disk using ETag-based conditional requests (REG-06).
// All HTTP calls use go-retryablehttp with RetryMax=3 for resilience.
// Timeouts: 3s dial, 10s response header (T-02-01, architecture constraints).
type GitHubManifestRegistry struct {
	name        string
	manifestURL string
	cachePath   string
	client      *http.Client
}

// NewGitHubManifestRegistry creates a registry using the standard cache path derived
// from ManifestCachePath(name). Uses production timeouts: 3s dial, 10s response.
func NewGitHubManifestRegistry(name, manifestURL string) *GitHubManifestRegistry {
	cachePath, _ := config.ManifestCachePath(sanitizeID(name))
	return newRegistry(name, manifestURL, cachePath, 3*time.Second, 10*time.Second)
}

// NewGitHubManifestRegistryWithCache creates a registry with an injected cache path (for testing).
func NewGitHubManifestRegistryWithCache(name, manifestURL, cachePath string) *GitHubManifestRegistry {
	return newRegistry(name, manifestURL, cachePath, 3*time.Second, 10*time.Second)
}

// NewGitHubManifestRegistryWithCacheAndTimeout creates a registry with injected cache path
// and custom timeouts (for timeout tests that need short timeouts).
func NewGitHubManifestRegistryWithCacheAndTimeout(name, manifestURL, cachePath string,
	dialTimeout, responseTimeout time.Duration) *GitHubManifestRegistry {
	return newRegistry(name, manifestURL, cachePath, dialTimeout, responseTimeout)
}

func newRegistry(name, manifestURL, cachePath string, dialTimeout, responseTimeout time.Duration) *GitHubManifestRegistry {
	rc := retryablehttp.NewClient()
	rc.Logger = nil // suppress default log output in tests

	// When using non-production (short) timeouts, disable retries so timeout tests
	// complete quickly without multiplying the timeout by RetryMax.
	isTestTimeout := responseTimeout < time.Second
	if isTestTimeout {
		rc.RetryMax = 0
	} else {
		rc.RetryMax = 3
	}

	// Override the transport with explicit timeouts.
	// ResponseHeaderTimeout covers "time to first byte" from the server.
	// The total request timeout wraps both dial and response.
	rc.HTTPClient = &http.Client{
		Timeout: dialTimeout + responseTimeout,
		Transport: &http.Transport{
			ResponseHeaderTimeout: responseTimeout,
		},
	}

	return &GitHubManifestRegistry{
		name:        name,
		manifestURL: manifestURL,
		cachePath:   cachePath,
		client:      rc.StandardClient(),
	}
}

// Name returns the registry identifier.
func (r *GitHubManifestRegistry) Name() string { return r.name }

// Resolve finds a package by exact name (case-insensitive).
// Returns (*Package, nil) on hit, (nil, nil) when not found,
// or (nil, error) when the registry is unreachable and no cache exists.
func (r *GitHubManifestRegistry) Resolve(name string) (*domain.Package, error) {
	manifest, err := r.fetch()
	if err != nil {
		return nil, err
	}
	needle := strings.ToLower(name)
	for _, p := range manifest.Packages {
		if strings.ToLower(p.Name) == needle {
			pkg := p
			return &pkg, nil
		}
	}
	return nil, nil
}

// Search returns all packages from the manifest that match query.
// An empty query returns all packages.
func (r *GitHubManifestRegistry) Search(query string) ([]domain.Package, error) {
	manifest, err := r.fetch()
	if err != nil {
		return nil, err
	}
	if query == "" {
		return manifest.Packages, nil
	}
	needle := strings.ToLower(query)
	var results []domain.Package
	for _, p := range manifest.Packages {
		if strings.Contains(strings.ToLower(p.Name), needle) ||
			strings.Contains(strings.ToLower(p.Description), needle) {
			results = append(results, p)
		}
	}
	return results, nil
}

// fetch retrieves the manifest, using ETag cache when available (REG-06).
//
// Flow:
//  1. Load cache from disk (empty if absent).
//  2. Send GET with If-None-Match header if ETag is known.
//  3. On 304: return cached manifest unchanged.
//  4. On 200: parse body, save cache with new ETag, return parsed manifest.
//  5. On network error with cache: log warning, return stale cache.
//  6. On network error without cache: return "registry unreachable" error.
func (r *GitHubManifestRegistry) fetch() (domain.Manifest, error) {
	cached, err := loadCache(r.cachePath)
	if err != nil {
		// Cache read failure is non-fatal; proceed without cache.
		log.Printf("agentkit: cache read error for %s: %v", r.name, err)
		cached = CachedManifest{}
	}

	req, err := http.NewRequest(http.MethodGet, r.manifestURL, nil)
	if err != nil {
		return r.fallback(cached, fmt.Errorf("registry unreachable: %w", err))
	}
	if cached.ETag != "" {
		req.Header.Set("If-None-Match", cached.ETag)
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return r.fallback(cached, fmt.Errorf("registry unreachable and no cache: %w", err))
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotModified {
		// Cache is still fresh — return it without re-parsing.
		return cached.Manifest, nil
	}

	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("registry %s returned HTTP %d", r.name, resp.StatusCode)
		return r.fallback(cached, fmt.Errorf("registry unreachable and no cache: %w", err))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return r.fallback(cached, fmt.Errorf("registry unreachable and no cache: %w", err))
	}

	var manifest domain.Manifest
	if err := json.Unmarshal(body, &manifest); err != nil {
		return r.fallback(cached, fmt.Errorf("registry unreachable and no cache: %w", err))
	}

	// Persist updated cache with new ETag.
	newCache := CachedManifest{
		ETag:      resp.Header.Get("ETag"),
		FetchedAt: time.Now().UTC(),
		Manifest:  manifest,
	}
	if err := saveCache(r.cachePath, newCache); err != nil {
		log.Printf("agentkit: cache write error for %s: %v", r.name, err)
	}

	return manifest, nil
}

// fallback returns the stale cache when one exists, or propagates the original error.
func (r *GitHubManifestRegistry) fallback(cached CachedManifest, origErr error) (domain.Manifest, error) {
	if len(cached.Manifest.Packages) > 0 {
		log.Printf("agentkit: %s unreachable, using stale cache from %s",
			r.name, cached.FetchedAt.Format(time.RFC3339))
		return cached.Manifest, nil
	}
	return domain.Manifest{}, origErr
}

// sanitizeID allows only [a-zA-Z0-9_-] in a registry ID used as a path component (T-02-05).
func sanitizeID(id string) string {
	out := make([]byte, 0, len(id))
	for i := 0; i < len(id); i++ {
		c := id[i]
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') || c == '_' || c == '-' {
			out = append(out, c)
		}
	}
	return string(out)
}
