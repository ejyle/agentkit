package installer

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/ejyle/agentkit/internal/config"
	"github.com/ejyle/agentkit/internal/domain"
	"github.com/google/renameio/v2"
)

// Version is the current agentkit binary version.
// Override via ldflags: -X github.com/ejyle/agentkit/internal/installer.Version=v0.1.0
var Version = "dev"

// GitHubReleaseInstaller fetches the GitHub source-archive tarball for the current binary
// version, extracts a named subdirectory to spec.SkillDir, and caches the tarball
// in-process (and on disk) to avoid redundant downloads during bundle installs.
type GitHubReleaseInstaller struct {
	client       *http.Client
	version      string
	cache        sync.Map // key: "repo@version", value: []byte
	diskCacheDir string   // if set, overrides UserCacheDir for on-disk cache (used in tests)
}

// NewGitHubReleaseInstaller returns a GitHubReleaseInstaller using http.DefaultClient
// and the package-level Version var.
func NewGitHubReleaseInstaller() *GitHubReleaseInstaller {
	return &GitHubReleaseInstaller{
		client:  http.DefaultClient,
		version: Version,
	}
}

// NewGitHubReleaseInstallerWithClient returns a GitHubReleaseInstaller with an injected
// HTTP client and version string. Used in tests to avoid hitting real GitHub.
func NewGitHubReleaseInstallerWithClient(client *http.Client, version string) *GitHubReleaseInstaller {
	return &GitHubReleaseInstaller{
		client:  client,
		version: version,
	}
}

// WithDiskCacheDir returns a copy of the installer with the given disk cache directory.
// Used in tests to avoid writing to the real UserCacheDir.
func (g *GitHubReleaseInstaller) WithDiskCacheDir(dir string) *GitHubReleaseInstaller {
	return &GitHubReleaseInstaller{
		client:       g.client,
		version:      g.version,
		diskCacheDir: dir,
	}
}

// Method returns InstallMethodGitHubRelease.
func (g *GitHubReleaseInstaller) Method() domain.InstallMethod {
	return domain.InstallMethodGitHubRelease
}

// IsAvailable always returns true — GitHub release download has no runtime prerequisite.
func (g *GitHubReleaseInstaller) IsAvailable() bool {
	return true
}

// Install fetches the release tarball for spec.Repo at g.version, extracts spec.Path
// subdirectory to spec.SkillDir, using in-process and on-disk caching to avoid
// redundant downloads.
//
// Returns ErrGitHubReleaseNotFound if the tarball URL returns HTTP 404.
// Returns a descriptive error if spec.Repo, spec.Path, or spec.SkillDir are empty.
func (g *GitHubReleaseInstaller) Install(spec domain.InstallSpec) error {
	// Validate required fields.
	if spec.Repo == "" {
		return fmt.Errorf("github-release: spec.Repo must not be empty")
	}
	if spec.Path == "" {
		return fmt.Errorf("github-release: spec.Path must not be empty")
	}
	if spec.SkillDir == "" {
		return fmt.Errorf("github-release: SkillDir not set; service.Install() must populate it")
	}

	// Resolve version.
	version := g.version
	if version == "" {
		version = "dev"
	}

	// Build tarball URL (HTTPS-only, T-03-02).
	tarURL := "https://github.com/" + spec.Repo + "/archive/refs/tags/" + version + ".tar.gz"

	// Cache key.
	cacheKey := spec.Repo + "@" + version

	// Fetch tarball bytes (in-process cache → disk cache → download).
	data, err := g.fetchTarball(tarURL, cacheKey, spec.Repo, version)
	if err != nil {
		return err
	}

	// Compute archive prefix: <repoName>-<version-without-v>/
	repoName := spec.Repo
	if idx := strings.LastIndex(repoName, "/"); idx >= 0 {
		repoName = repoName[idx+1:]
	}
	versionNoV := strings.TrimPrefix(version, "v")
	archivePrefix := repoName + "-" + versionNoV + "/"
	fullPrefix := archivePrefix + strings.TrimSuffix(spec.Path, "/") + "/"

	destDir := spec.SkillDir

	// Extract matching entries.
	return extractSubdir(data, fullPrefix, destDir)
}

// resolveDiskCachePath returns the on-disk tarball cache path, using diskCacheDir
// if set (for tests), otherwise falling back to config.TarballCachePath.
func (g *GitHubReleaseInstaller) resolveDiskCachePath(repo, version string) (string, error) {
	if g.diskCacheDir != "" {
		slug := strings.ReplaceAll(repo, "/", "-")
		return filepath.Join(g.diskCacheDir, slug, version, "tarball.tar.gz"), nil
	}
	return config.TarballCachePath(repo, version)
}

// fetchTarball retrieves tarball bytes, using in-process cache first, then disk cache,
// then network download. Returns ErrGitHubReleaseNotFound on HTTP 404.
func (g *GitHubReleaseInstaller) fetchTarball(tarURL, cacheKey, repo, version string) ([]byte, error) {
	// In-process cache hit.
	if cached, ok := g.cache.Load(cacheKey); ok {
		return cached.([]byte), nil
	}

	// Disk cache check.
	diskPath, diskPathErr := g.resolveDiskCachePath(repo, version)
	if diskPathErr == nil {
		if diskData, readErr := os.ReadFile(diskPath); readErr == nil {
			g.cache.Store(cacheKey, diskData)
			return diskData, nil
		}
	}

	// Network download.
	resp, err := g.client.Get(tarURL)
	if err != nil {
		return nil, fmt.Errorf("github-release: download failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrGitHubReleaseNotFound
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github-release: unexpected HTTP status %d for %s", resp.StatusCode, tarURL)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("github-release: reading response body: %w", err)
	}

	// Write to disk cache atomically.
	if diskPathErr == nil && diskPath != "" {
		if mkErr := os.MkdirAll(filepath.Dir(diskPath), 0755); mkErr == nil {
			_ = renameio.WriteFile(diskPath, data, 0644)
		}
	}

	// Populate in-process cache.
	g.cache.Store(cacheKey, data)

	return data, nil
}

// extractSubdir extracts entries matching fullPrefix from the gzip-compressed tar data
// into destDir, stripping the fullPrefix from each path.
// Path traversal entries are rejected (T-03-01).
func extractSubdir(data []byte, fullPrefix, destDir string) error {
	gr, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("github-release: decompressing tarball: %w", err)
	}
	defer gr.Close()

	// Ensure destDir has a trailing separator for boundary checks.
	destDirClean := filepath.Clean(destDir)
	destDirPrefix := destDirClean + string(filepath.Separator)

	tr := tar.NewReader(gr)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("github-release: reading tar entry: %w", err)
		}

		if !strings.HasPrefix(hdr.Name, fullPrefix) {
			continue
		}

		rel := strings.TrimPrefix(hdr.Name, fullPrefix)
		if rel == "" {
			// Root directory entry — ensure destDir exists.
			if mkErr := os.MkdirAll(destDirClean, 0755); mkErr != nil {
				return fmt.Errorf("github-release: creating dest dir: %w", mkErr)
			}
			continue
		}

		// Path traversal guard (T-03-01).
		if strings.Contains(rel, "..") {
			return fmt.Errorf("github-release: path traversal rejected: %q", hdr.Name)
		}
		resolved := filepath.Join(destDirClean, rel)
		if resolved != destDirClean && !strings.HasPrefix(resolved, destDirPrefix) {
			return fmt.Errorf("github-release: path traversal rejected: %q", hdr.Name)
		}

		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(resolved, 0755); err != nil {
				return fmt.Errorf("github-release: creating directory %q: %w", resolved, err)
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(resolved), 0755); err != nil {
				return fmt.Errorf("github-release: creating parent dir for %q: %w", resolved, err)
			}
			mode := os.FileMode(hdr.Mode) & 0755
			f, err := os.OpenFile(resolved, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
			if err != nil {
				return fmt.Errorf("github-release: creating file %q: %w", resolved, err)
			}
			if _, err := io.Copy(f, tr); err != nil {
				f.Close()
				return fmt.Errorf("github-release: writing file %q: %w", resolved, err)
			}
			f.Close()
		}
	}

	return nil
}
