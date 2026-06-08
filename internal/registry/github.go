package registry

import (
	"net/http"
	"time"

	"github.com/ejyle/agentkit/internal/config"
	"github.com/ejyle/agentkit/internal/domain"
)

// GitHubManifestRegistry fetches a registry.json manifest from a GitHub raw URL.
// It caches manifests on disk using ETag-based conditional requests (REG-06).
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
// and custom timeouts (for timeout tests).
func NewGitHubManifestRegistryWithCacheAndTimeout(name, manifestURL, cachePath string,
	dialTimeout, responseTimeout time.Duration) *GitHubManifestRegistry {
	return newRegistry(name, manifestURL, cachePath, dialTimeout, responseTimeout)
}

func newRegistry(name, manifestURL, cachePath string, _, responseTimeout time.Duration) *GitHubManifestRegistry {
	// Use a simple http.Client with a total request timeout. The timeout covers
	// both dial and response (sufficient for tests; production uses go-retryablehttp).
	client := &http.Client{Timeout: responseTimeout}
	return &GitHubManifestRegistry{
		name:        name,
		manifestURL: manifestURL,
		cachePath:   cachePath,
		client:      client,
	}
}

// Name returns the registry identifier.
func (r *GitHubManifestRegistry) Name() string { return r.name }

// Resolve finds a package by exact name (case-insensitive).
func (r *GitHubManifestRegistry) Resolve(name string) (*domain.Package, error) {
	panic("not implemented")
}

// Search returns all packages that match query. Empty query returns all packages.
func (r *GitHubManifestRegistry) Search(query string) ([]domain.Package, error) {
	panic("not implemented")
}

// fetch retrieves the manifest, using ETag cache when available.
func (r *GitHubManifestRegistry) fetch() (domain.Manifest, error) {
	panic("not implemented")
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
