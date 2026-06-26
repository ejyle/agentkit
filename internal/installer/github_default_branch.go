package installer

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/ejyle/agentkit/internal/domain"
	"github.com/ejyle/agentkit/internal/fileutil"
)

const defaultBranch = "main"

// GitHubDefaultBranchInstaller fetches the default-branch source archive of an external
// GitHub repository and extracts spec.Path into spec.SkillDir.
//
// Unlike GitHubReleaseInstaller, the download URL is not tied to the agentkit binary
// version — it always fetches the current HEAD of the branch, so `agentkit update`
// naturally pulls the latest upstream skill content.
//
// Branch selection: defaults to "main". Override by setting spec.Args[0].
// Path "." extracts the entire repo root (all files, stripped of the archive prefix).
type GitHubDefaultBranchInstaller struct {
	client       *http.Client
	cache        sync.Map // key: "repo@branch", value: []byte
	diskCacheDir string   // overrides UserCacheDir in tests
}

// NewGitHubDefaultBranchInstaller returns a GitHubDefaultBranchInstaller using http.DefaultClient.
func NewGitHubDefaultBranchInstaller() *GitHubDefaultBranchInstaller {
	return &GitHubDefaultBranchInstaller{client: http.DefaultClient}
}

// NewGitHubDefaultBranchInstallerWithClient returns an installer with an injected HTTP client.
// Used in tests to avoid hitting real GitHub.
func NewGitHubDefaultBranchInstallerWithClient(client *http.Client) *GitHubDefaultBranchInstaller {
	return &GitHubDefaultBranchInstaller{client: client}
}

// WithDiskCacheDir returns a copy of the installer with the given disk cache directory.
func (g *GitHubDefaultBranchInstaller) WithDiskCacheDir(dir string) *GitHubDefaultBranchInstaller {
	return &GitHubDefaultBranchInstaller{
		client:       g.client,
		diskCacheDir: dir,
	}
}

// Method returns InstallMethodGitHubDefaultBranch.
func (g *GitHubDefaultBranchInstaller) Method() domain.InstallMethod {
	return domain.InstallMethodGitHubDefaultBranch
}

// IsAvailable always returns true — no runtime prerequisite beyond network access.
func (g *GitHubDefaultBranchInstaller) IsAvailable() bool { return true }

// Install fetches the default-branch tarball for spec.Repo, extracts spec.Path into
// spec.SkillDir. spec.Path may be "." or "" to extract the entire repo root.
//
// The branch is taken from spec.Args[0] when set; otherwise defaults to "main".
//
// Returns ErrGitHubDefaultBranchNotFound on HTTP 404.
func (g *GitHubDefaultBranchInstaller) Install(spec domain.InstallSpec) error {
	if spec.Repo == "" {
		return fmt.Errorf("github-default-branch: spec.Repo must not be empty")
	}
	if spec.SkillDir == "" {
		return fmt.Errorf("github-default-branch: SkillDir not set; service.Install() must populate it")
	}

	branch := defaultBranch
	if len(spec.Args) > 0 && spec.Args[0] != "" {
		branch = spec.Args[0]
	}

	tarURL := "https://github.com/" + spec.Repo + "/archive/refs/heads/" + branch + ".tar.gz"
	cacheKey := spec.Repo + "@" + branch

	data, err := g.fetchTarball(tarURL, cacheKey, spec.Repo, branch)
	if err != nil {
		return err
	}

	// Archive prefix: <repoName>-<branch>/
	repoName := spec.Repo
	if idx := strings.LastIndex(repoName, "/"); idx >= 0 {
		repoName = repoName[idx+1:]
	}
	archivePrefix := repoName + "-" + branch + "/"

	// "." or "" means extract the entire repo root — use archivePrefix as the full prefix.
	path := strings.TrimSuffix(spec.Path, "/")
	var fullPrefix string
	if path == "" || path == "." {
		fullPrefix = archivePrefix
	} else {
		fullPrefix = archivePrefix + path + "/"
	}

	return extractSubdir(data, fullPrefix, spec.SkillDir)
}

func (g *GitHubDefaultBranchInstaller) resolveDiskCachePath(repo, branch string) (string, error) {
	if g.diskCacheDir != "" {
		slug := strings.ReplaceAll(repo, "/", "-")
		return filepath.Join(g.diskCacheDir, slug, branch, "tarball.tar.gz"), nil
	}
	// Reuse the same cache layout as github-release but keyed by branch name.
	slug := strings.ReplaceAll(repo, "/", "-")
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(cacheDir, "agentkit", "tarballs", slug, branch, "tarball.tar.gz"), nil
}

func (g *GitHubDefaultBranchInstaller) fetchTarball(tarURL, cacheKey, repo, branch string) ([]byte, error) {
	// In-process cache — intentionally NOT used for default-branch installs on update
	// (the caller clears the installer per invocation, so each `agentkit update` gets
	// a fresh instance and always hits the network).
	if cached, ok := g.cache.Load(cacheKey); ok {
		return cached.([]byte), nil
	}

	// For update flows the disk cache is also bypassed because the installer is
	// re-created per command invocation and the cache key changes on re-run if we
	// include a timestamp. For now we do NOT persist the tarball to disk — this
	// ensures `agentkit update` always fetches the latest content.

	resp, err := g.client.Get(tarURL)
	if err != nil {
		return nil, fmt.Errorf("github-default-branch: download failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrGitHubDefaultBranchNotFound
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github-default-branch: unexpected HTTP status %d for %s", resp.StatusCode, tarURL)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("github-default-branch: reading response body: %w", err)
	}

	// Write to disk cache for same-session bundle installs (avoids re-downloading the
	// same repo multiple times in one `agentkit install` run).
	diskPath, diskPathErr := g.resolveDiskCachePath(repo, branch)
	if diskPathErr == nil && diskPath != "" {
		if mkErr := os.MkdirAll(filepath.Dir(diskPath), 0755); mkErr == nil {
			_ = fileutil.WriteFile(diskPath, data, 0644)
		}
	}

	g.cache.Store(cacheKey, data)
	return data, nil
}
