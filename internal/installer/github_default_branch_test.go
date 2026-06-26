package installer_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/ejyle/agentkit/internal/domain"
	"github.com/ejyle/agentkit/internal/installer"
)

// TestGitHubDefaultBranch_ExtractRoot verifies that path "." extracts the entire
// repo root, stripping the archive prefix (e.g. "azure-skills-main/").
func TestGitHubDefaultBranch_ExtractRoot(t *testing.T) {
	tarball := buildTestTarball("azure-skills-main/", map[string]string{
		"SKILL.md":          "name: azure-skills\n",
		"references/aks.md": "# AKS\n",
	})

	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(tarball)
	}))
	defer ts.Close()

	destDir := t.TempDir()
	inst := installer.NewGitHubDefaultBranchInstallerWithClient(redirectingClient(ts)).
		WithDiskCacheDir(t.TempDir())

	err := inst.Install(domain.InstallSpec{
		Method:   domain.InstallMethodGitHubDefaultBranch,
		Repo:     "microsoft/azure-skills",
		Path:     ".",
		SkillDir: destDir,
	})
	if err != nil {
		t.Fatalf("Install() unexpected error: %v", err)
	}

	assertFileContent(t, destDir+"/SKILL.md", "name: azure-skills\n")
	assertFileContent(t, destDir+"/references/aks.md", "# AKS\n")
}

// TestGitHubDefaultBranch_ExtractSubdir verifies that a non-root path extracts
// only the matching subdirectory.
func TestGitHubDefaultBranch_ExtractSubdir(t *testing.T) {
	tarball := buildTestTarball("myrepo-main/", map[string]string{
		"skills/aws/SKILL.md":  "name: aws\n",
		"skills/aws/README.md": "# AWS\n",
		"other/file.txt":       "should not be extracted\n",
	})

	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(tarball)
	}))
	defer ts.Close()

	destDir := t.TempDir()
	inst := installer.NewGitHubDefaultBranchInstallerWithClient(redirectingClient(ts)).
		WithDiskCacheDir(t.TempDir())

	err := inst.Install(domain.InstallSpec{
		Method:   domain.InstallMethodGitHubDefaultBranch,
		Repo:     "example/myrepo",
		Path:     "skills/aws",
		SkillDir: destDir,
	})
	if err != nil {
		t.Fatalf("Install() unexpected error: %v", err)
	}

	assertFileContent(t, destDir+"/SKILL.md", "name: aws\n")
	assertFileContent(t, destDir+"/README.md", "# AWS\n")

	if _, err := os.Stat(destDir + "/other/file.txt"); err == nil {
		t.Error("Install() extracted file outside requested path; want it excluded")
	}
}

// TestGitHubDefaultBranch_CustomBranch verifies that spec.Args[0] overrides the
// default branch, changing both the URL and the archive prefix used for extraction.
func TestGitHubDefaultBranch_CustomBranch(t *testing.T) {
	tarball := buildTestTarball("myrepo-develop/", map[string]string{
		"SKILL.md": "branch: develop\n",
	})

	var capturedURL string
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedURL = r.URL.String()
		w.WriteHeader(http.StatusOK)
		w.Write(tarball)
	}))
	defer ts.Close()

	destDir := t.TempDir()
	inst := installer.NewGitHubDefaultBranchInstallerWithClient(redirectingClient(ts)).
		WithDiskCacheDir(t.TempDir())

	err := inst.Install(domain.InstallSpec{
		Method:   domain.InstallMethodGitHubDefaultBranch,
		Repo:     "example/myrepo",
		Path:     ".",
		Args:     []string{"develop"},
		SkillDir: destDir,
	})
	if err != nil {
		t.Fatalf("Install() unexpected error: %v", err)
	}

	if !strings.Contains(capturedURL, "develop") {
		t.Errorf("expected URL to contain branch %q, got %q", "develop", capturedURL)
	}
	assertFileContent(t, destDir+"/SKILL.md", "branch: develop\n")
}

// TestGitHubDefaultBranch_NotFound verifies that a 404 returns ErrGitHubDefaultBranchNotFound.
func TestGitHubDefaultBranch_NotFound(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	inst := installer.NewGitHubDefaultBranchInstallerWithClient(redirectingClient(ts)).
		WithDiskCacheDir(t.TempDir())

	err := inst.Install(domain.InstallSpec{
		Method:   domain.InstallMethodGitHubDefaultBranch,
		Repo:     "example/does-not-exist",
		Path:     ".",
		SkillDir: t.TempDir(),
	})
	if err != installer.ErrGitHubDefaultBranchNotFound {
		t.Errorf("Install() error = %v; want ErrGitHubDefaultBranchNotFound", err)
	}
}

// TestGitHubDefaultBranch_MissingRepo verifies that an empty Repo returns an error.
func TestGitHubDefaultBranch_MissingRepo(t *testing.T) {
	inst := installer.NewGitHubDefaultBranchInstaller()
	err := inst.Install(domain.InstallSpec{
		Method:   domain.InstallMethodGitHubDefaultBranch,
		Repo:     "",
		Path:     ".",
		SkillDir: t.TempDir(),
	})
	if err == nil {
		t.Fatal("Install() expected error for empty Repo, got nil")
	}
}

// TestGitHubDefaultBranch_MissingSkillDir verifies that an empty SkillDir returns an error.
func TestGitHubDefaultBranch_MissingSkillDir(t *testing.T) {
	inst := installer.NewGitHubDefaultBranchInstaller()
	err := inst.Install(domain.InstallSpec{
		Method:   domain.InstallMethodGitHubDefaultBranch,
		Repo:     "microsoft/azure-skills",
		Path:     ".",
		SkillDir: "",
	})
	if err == nil {
		t.Fatal("Install() expected error for empty SkillDir, got nil")
	}
}

// TestGitHubDefaultBranch_NoCacheAcrossInstances verifies that two separate installer
// instances (simulating two `agentkit update` runs) each make a fresh HTTP request.
func TestGitHubDefaultBranch_NoCacheAcrossInstances(t *testing.T) {
	tarball := buildTestTarball("myrepo-main/", map[string]string{
		"SKILL.md": "name: myrepo\n",
	})

	var callCount int64
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(&callCount, 1)
		w.WriteHeader(http.StatusOK)
		w.Write(tarball)
	}))
	defer ts.Close()

	spec := domain.InstallSpec{
		Method: domain.InstallMethodGitHubDefaultBranch,
		Repo:   "example/myrepo",
		Path:   ".",
	}

	inst1 := installer.NewGitHubDefaultBranchInstallerWithClient(redirectingClient(ts)).
		WithDiskCacheDir(t.TempDir())
	spec.SkillDir = t.TempDir()
	if err := inst1.Install(spec); err != nil {
		t.Fatalf("first Install() error: %v", err)
	}

	inst2 := installer.NewGitHubDefaultBranchInstallerWithClient(redirectingClient(ts)).
		WithDiskCacheDir(t.TempDir())
	spec.SkillDir = t.TempDir()
	if err := inst2.Install(spec); err != nil {
		t.Fatalf("second Install() error: %v", err)
	}

	if n := atomic.LoadInt64(&callCount); n != 2 {
		t.Errorf("HTTP handler called %d times across two instances; want 2", n)
	}
}
