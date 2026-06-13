package installer_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/ejyle/agentkit/internal/domain"
	"github.com/ejyle/agentkit/internal/installer"
)

// TestGitHubReleaseInstaller_ExtractToSkillPath verifies the end-to-end extraction
// flow: tarball served by an httptest server is extracted to a simulated SkillInstallPath,
// confirming that SKILL.md and references/ are present with correct contents.
//
// This test satisfies ROADMAP Phase 3 success criterion 4:
// "installed aws skill directory contains SKILL.md and references/".
func TestGitHubReleaseInstaller_ExtractToSkillPath(t *testing.T) {
	// Build test tarball with prefix matching archive structure for v0.1.0 of ejyle/agentkit.
	tarball := buildTestTarball("agentkit-0.1.0/skills/aws/", map[string]string{
		"SKILL.md":              "name: aws\n",
		"references/ec2.md":     "# EC2\n",
		"references/s3.md":      "# S3\n",
	})

	// Start httptest TLS server.
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(tarball)
	}))
	defer ts.Close()

	// Use TLS-aware client with URL redirection to test server.
	tlsClient := redirectingClient(ts)

	// Create installer with injected disk cache dir to avoid writing to real UserCacheDir.
	inst := installer.NewGitHubReleaseInstallerWithClient(tlsClient, "v0.1.0").
		WithDiskCacheDir(t.TempDir())

	// Simulates SkillInstallPath(target="claude", name="aws") output.
	// We use a temp dir to avoid writing to real ~/.claude/skills/aws/.
	destDir := filepath.Join(t.TempDir(), "claude", "skills", "aws")

	// Call Install with domain.InstallSpec.SkillDir set (as service.Install() would do).
	err := inst.Install(domain.InstallSpec{
		Method:  domain.InstallMethodGitHubRelease,
		Repo:    "ejyle/agentkit",
		Path:    "skills/aws",
		Package: "aws",
		SkillDir: destDir,
	})

	// Assertion 1: no error.
	if err != nil {
		t.Fatalf("Install() unexpected error: %v", err)
	}

	// Assertion 2: SKILL.md exists.
	skillMD := filepath.Join(destDir, "SKILL.md")
	if _, statErr := os.Stat(skillMD); statErr != nil {
		t.Fatalf("SKILL.md not found at %s: %v", skillMD, statErr)
	}

	// Assertion 3: SKILL.md has correct content.
	data, err := os.ReadFile(skillMD)
	if err != nil {
		t.Fatalf("reading SKILL.md: %v", err)
	}
	if string(data) != "name: aws\n" {
		t.Errorf("SKILL.md content = %q; want %q", string(data), "name: aws\n")
	}

	// Assertion 4: references/ is a directory.
	refsDir := filepath.Join(destDir, "references")
	info, err := os.Stat(refsDir)
	if err != nil {
		t.Fatalf("references/ directory not found at %s: %v", refsDir, err)
	}
	if !info.IsDir() {
		t.Errorf("references/ is not a directory at %s", refsDir)
	}

	// Assertion 5: references/ec2.md exists.
	ec2MD := filepath.Join(destDir, "references", "ec2.md")
	if _, statErr := os.Stat(ec2MD); statErr != nil {
		t.Fatalf("references/ec2.md not found at %s: %v", ec2MD, statErr)
	}

	// Assertion 6: references/s3.md exists.
	s3MD := filepath.Join(destDir, "references", "s3.md")
	if _, statErr := os.Stat(s3MD); statErr != nil {
		t.Fatalf("references/s3.md not found at %s: %v", s3MD, statErr)
	}
}
