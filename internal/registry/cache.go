package registry

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/ejyle/agentkit/internal/domain"
	"github.com/ejyle/agentkit/internal/fileutil"
)

// CachedManifest is the on-disk representation of a cached registry manifest.
type CachedManifest struct {
	ETag      string          `json:"etag"`
	FetchedAt time.Time       `json:"fetched_at"`
	Manifest  domain.Manifest `json:"manifest"`
}

// loadCache reads a CachedManifest from path.
// If the file does not exist, returns an empty CachedManifest with no error.
func loadCache(path string) (CachedManifest, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return CachedManifest{}, nil
	}
	if err != nil {
		return CachedManifest{}, err
	}
	var c CachedManifest
	if err := json.Unmarshal(data, &c); err != nil {
		return CachedManifest{}, err
	}
	return c, nil
}

// saveCache writes a CachedManifest to path using an atomic rename write (T-02-04).
func saveCache(path string, c CachedManifest) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return fileutil.WriteFile(path, data, 0644)
}
