package installer_test

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/ejyle/agentkit/internal/domain"
	"github.com/ejyle/agentkit/internal/installer"
)

// buildTestTarball creates an in-memory .tar.gz with entries under the given prefix.
// files maps relative paths (relative to prefix) to file contents.
func buildTestTarball(prefix string, files map[string]string) []byte {
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)

	// Write prefix directory entry.
	_ = tw.WriteHeader(&tar.Header{
		Typeflag: tar.TypeDir,
		Name:     prefix,
		Mode:     0755,
	})

	for rel, content := range files {
		data := []byte(content)
		// Write parent directory entry if the path contains a slash.
		if idx := strings.LastIndex(rel, "/"); idx > 0 {
			dirEntry := prefix + rel[:idx+1]
			_ = tw.WriteHeader(&tar.Header{
				Typeflag: tar.TypeDir,
				Name:     dirEntry,
				Mode:     0755,
			})
		}
		_ = tw.WriteHeader(&tar.Header{
			Typeflag: tar.TypeReg,
			Name:     prefix + rel,
			Size:     int64(len(data)),
			Mode:     0644,
		})
		_, _ = tw.Write(data)
	}

	_ = tw.Close()
	_ = gw.Close()
	return buf.Bytes()
}

// redirectingClient returns an *http.Client that forwards all requests to the
// given httptest.Server regardless of the target URL.
func redirectingClient(ts *httptest.Server) *http.Client {
	base := ts.Client()
	origTransport := base.Transport
	base.Transport = &redirectTransport{
		base:   origTransport,
		target: ts.URL,
	}
	return base
}

type redirectTransport struct {
	base   http.RoundTripper
	target string
}

func (rt *redirectTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	newReq := req.Clone(req.Context())
	newReq.URL.Scheme = "https"
	newReq.URL.Host = strings.TrimPrefix(rt.target, "https://")
	return rt.base.RoundTrip(newReq)
}

// TestGitHubReleaseInstaller_Extract verifies that Install() extracts files from a
// valid tarball to spec.SkillDir, stripping the archive prefix.
func TestGitHubReleaseInstaller_Extract(t *testing.T) {
	tarball := buildTestTarball("agentkit-0.1.0/skills/aws/", map[string]string{
		"SKILL.md":          "name: aws\n",
		"references/ec2.md": "# EC2\n",
	})

	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(tarball)
	}))
	defer ts.Close()

	destDir := t.TempDir()
	inst := installer.NewGitHubReleaseInstallerWithClient(redirectingClient(ts), "v0.1.0").
		WithDiskCacheDir(t.TempDir())

	err := inst.Install(domain.InstallSpec{
		Method:   domain.InstallMethodGitHubRelease,
		Repo:     "ejyle/agentkit",
		Path:     "skills/aws",
		SkillDir: destDir,
	})
	if err != nil {
		t.Fatalf("Install() unexpected error: %v", err)
	}

	// Verify extracted files.
	assertFileContent(t, destDir+"/SKILL.md", "name: aws\n")
	assertFileContent(t, destDir+"/references/ec2.md", "# EC2\n")
}

// TestGitHubReleaseInstaller_PathTraversalRejected verifies that a tarball with
// a path traversal entry causes Install() to return an error containing "path traversal".
func TestGitHubReleaseInstaller_PathTraversalRejected(t *testing.T) {
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)

	prefix := "agentkit-0.1.0/skills/aws/"
	_ = tw.WriteHeader(&tar.Header{Typeflag: tar.TypeDir, Name: prefix, Mode: 0755})

	malicious := []byte("root:x:0:0:root:/root:/bin/bash\n")
	_ = tw.WriteHeader(&tar.Header{
		Typeflag: tar.TypeReg,
		Name:     prefix + "../../etc/passwd",
		Size:     int64(len(malicious)),
		Mode:     0644,
	})
	_, _ = tw.Write(malicious)
	_ = tw.Close()
	_ = gw.Close()

	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(buf.Bytes())
	}))
	defer ts.Close()

	destDir := t.TempDir()
	inst := installer.NewGitHubReleaseInstallerWithClient(redirectingClient(ts), "v0.1.0").
		WithDiskCacheDir(t.TempDir())

	err := inst.Install(domain.InstallSpec{
		Method:   domain.InstallMethodGitHubRelease,
		Repo:     "ejyle/agentkit",
		Path:     "skills/aws",
		SkillDir: destDir,
	})
	if err == nil {
		t.Fatal("Install() expected path traversal error, got nil")
	}
	if !strings.Contains(err.Error(), "path traversal") {
		t.Errorf("Install() error = %q; want error containing \"path traversal\"", err.Error())
	}
}

// TestGitHubReleaseInstaller_CacheHit verifies that calling Install() twice with
// the same repo@version triggers only one HTTP request (in-process cache hit).
func TestGitHubReleaseInstaller_CacheHit(t *testing.T) {
	tarball := buildTestTarball("agentkit-0.1.0/skills/aws/", map[string]string{
		"SKILL.md": "name: aws\n",
	})

	var callCount int64
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(&callCount, 1)
		w.WriteHeader(http.StatusOK)
		w.Write(tarball)
	}))
	defer ts.Close()

	inst := installer.NewGitHubReleaseInstallerWithClient(redirectingClient(ts), "v0.1.0").
		WithDiskCacheDir(t.TempDir())

	spec := domain.InstallSpec{
		Method:   domain.InstallMethodGitHubRelease,
		Repo:     "ejyle/agentkit",
		Path:     "skills/aws",
		SkillDir: t.TempDir(),
	}

	if err := inst.Install(spec); err != nil {
		t.Fatalf("Install() first call error: %v", err)
	}

	// Second call — fresh destDir, same repo@version — should use in-process cache.
	spec.SkillDir = t.TempDir()
	if err := inst.Install(spec); err != nil {
		t.Fatalf("Install() second call error: %v", err)
	}

	if n := atomic.LoadInt64(&callCount); n != 1 {
		t.Errorf("HTTP handler called %d times; want 1 (cache hit on second call)", n)
	}
}

// TestGitHubReleaseInstaller_NotFound verifies that a 404 response returns
// installer.ErrGitHubReleaseNotFound.
func TestGitHubReleaseInstaller_NotFound(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	inst := installer.NewGitHubReleaseInstallerWithClient(redirectingClient(ts), "v0.1.0").
		WithDiskCacheDir(t.TempDir())

	err := inst.Install(domain.InstallSpec{
		Method:   domain.InstallMethodGitHubRelease,
		Repo:     "ejyle/agentkit",
		Path:     "skills/aws",
		SkillDir: t.TempDir(),
	})
	if err == nil {
		t.Fatal("Install() expected ErrGitHubReleaseNotFound for 404, got nil")
	}
	if err != installer.ErrGitHubReleaseNotFound {
		t.Errorf("Install() error = %v; want ErrGitHubReleaseNotFound", err)
	}
}

// assertFileContent checks that a file exists and has the expected content.
func assertFileContent(t *testing.T, path, want string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("expected file %s not found: %v", path, err)
	}
	if string(data) != want {
		t.Errorf("file %s content = %q; want %q", path, string(data), want)
	}
}
