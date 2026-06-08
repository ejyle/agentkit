package installer_test

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/ejyle/agentkit/internal/domain"
	"github.com/ejyle/agentkit/internal/installer"
)

// Test 4: BinaryInstaller.Install(spec) downloads from spec.URL, writes to AgentBinPath()/name, chmod +x.
func TestBinaryInstaller_Install_Success(t *testing.T) {
	content := []byte("fake binary content")
	hash := sha256.Sum256(content)
	checksum := fmt.Sprintf("sha256:%x", hash)

	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(content)
	}))
	defer ts.Close()

	binDir := t.TempDir()
	inst := installer.NewBinaryInstallerWithBinDir(ts.Client(), binDir)

	err := inst.Install(domain.InstallSpec{
		Method:  domain.InstallMethodBinary,
		Package: "mytool",
		URL:     ts.URL,
		Args:    []string{checksum},
	})
	if err != nil {
		t.Fatalf("Install() unexpected error: %v", err)
	}

	outPath := filepath.Join(binDir, "mytool")
	info, err := os.Stat(outPath)
	if err != nil {
		t.Fatalf("output file not found at %s: %v", outPath, err)
	}
	// Check that file is executable (chmod +x sets 0755).
	if info.Mode()&0111 == 0 {
		t.Errorf("output file %s is not executable, mode=%v", outPath, info.Mode())
	}
}

// Test 5: BinaryInstaller.Install() returns error when SHA256 of downloaded file does not match spec SHA256.
func TestBinaryInstaller_Install_ChecksumMismatch(t *testing.T) {
	content := []byte("fake binary content")
	wrongChecksum := "sha256:0000000000000000000000000000000000000000000000000000000000000000"

	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(content)
	}))
	defer ts.Close()

	binDir := t.TempDir()
	inst := installer.NewBinaryInstallerWithBinDir(ts.Client(), binDir)

	err := inst.Install(domain.InstallSpec{
		Method:  domain.InstallMethodBinary,
		Package: "mytool",
		URL:     ts.URL,
		Args:    []string{wrongChecksum},
	})
	if err == nil {
		t.Fatal("Install() expected checksum mismatch error, got nil")
	}
	if !installer.IsErrChecksumMismatch(err) {
		t.Errorf("Install() error = %v; want ErrChecksumMismatch", err)
	}
}

// Test 6: BinaryInstaller.Install() returns error when spec.URL scheme is not "https".
func TestBinaryInstaller_Install_InsecureURL(t *testing.T) {
	binDir := t.TempDir()
	inst := installer.NewBinaryInstallerWithBinDir(http.DefaultClient, binDir)

	err := inst.Install(domain.InstallSpec{
		Method:  domain.InstallMethodBinary,
		Package: "mytool",
		URL:     "http://example.com/binary",
		Args:    []string{"sha256:abc"},
	})
	if err == nil {
		t.Fatal("Install() expected ErrInsecureURL for http:// URL, got nil")
	}
	if !installer.IsErrInsecureURL(err) {
		t.Errorf("Install() error = %v; want ErrInsecureURL", err)
	}
}
