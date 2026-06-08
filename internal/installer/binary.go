package installer

import (
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/ejyle/agentkit/internal/config"
	"github.com/ejyle/agentkit/internal/domain"
)

// BinaryInstaller downloads a pre-built binary, verifies its SHA256 checksum, and
// places it in AgentBinPath() with chmod +x. Uses HTTPS-only enforcement (T-03-03)
// and SHA256 verification (T-03-02).
type BinaryInstaller struct {
	client  *http.Client
	binPath string // empty means use config.AgentBinPath() at runtime
}

// NewBinaryInstaller returns a BinaryInstaller using the default http.Client and AgentBinPath.
func NewBinaryInstaller() *BinaryInstaller {
	return &BinaryInstaller{client: http.DefaultClient}
}

// NewBinaryInstallerWithBinDir returns a BinaryInstaller with an injected bin directory and HTTP client.
// Used in tests to avoid writing to the real AgentBinPath and to use httptest TLS clients.
func NewBinaryInstallerWithBinDir(client *http.Client, binDir string) *BinaryInstaller {
	return &BinaryInstaller{client: client, binPath: binDir}
}

// Method returns InstallMethodBinary.
func (b *BinaryInstaller) Method() domain.InstallMethod {
	return domain.InstallMethodBinary
}

// IsAvailable always returns true — binary download has no runtime prerequisite.
func (b *BinaryInstaller) IsAvailable() bool {
	return true
}

// Install downloads the binary at spec.URL, verifies its SHA256, and writes it to
// AgentBinPath()/<spec.Package> with permissions 0755.
//
// spec.Args[0] is expected to hold the expected checksum in the form "sha256:<hex>".
// Returns ErrInsecureURL if spec.URL is not https://.
// Returns ErrChecksumMismatch if the downloaded file's hash differs from the expected hash.
func (b *BinaryInstaller) Install(spec domain.InstallSpec) error {
	// Security: reject non-HTTPS URLs (T-03-03).
	u, err := url.Parse(spec.URL)
	if err != nil || u.Scheme != "https" {
		return ErrInsecureURL
	}

	// Download to memory.
	resp, err := b.client.Get(spec.URL)
	if err != nil {
		return fmt.Errorf("binary download failed: %w", err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading download body: %w", err)
	}

	// Verify SHA256 checksum (T-03-02).
	if len(spec.Args) > 0 && spec.Args[0] != "" {
		expected := spec.Args[0]
		sum := sha256.Sum256(data)
		actual := fmt.Sprintf("sha256:%x", sum)
		if actual != expected {
			return ErrChecksumMismatch
		}
	}

	// Resolve output directory.
	binDir := b.binPath
	if binDir == "" {
		binDir, err = config.AgentBinPath()
		if err != nil {
			return fmt.Errorf("resolving bin path: %w", err)
		}
	}
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return fmt.Errorf("creating bin directory: %w", err)
	}

	// Write to a temp file in the same directory, then rename (atomic).
	outPath := filepath.Join(binDir, spec.Package)
	tmpFile, err := os.CreateTemp(binDir, spec.Package+".*.tmp")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	if _, err := tmpFile.Write(data); err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("writing binary: %w", err)
	}
	tmpFile.Close()

	// Set executable bit before rename so the final path is immediately executable.
	if err := os.Chmod(tmpPath, 0755); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("chmod +x: %w", err)
	}
	if err := os.Rename(tmpPath, outPath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("renaming binary: %w", err)
	}
	return nil
}

// IsErrChecksumMismatch reports whether err equals ErrChecksumMismatch.
func IsErrChecksumMismatch(err error) bool {
	return err != nil && err.Error() == ErrChecksumMismatch.Error()
}

// IsErrInsecureURL reports whether err equals ErrInsecureURL.
func IsErrInsecureURL(err error) bool {
	return err != nil && err.Error() == ErrInsecureURL.Error()
}
