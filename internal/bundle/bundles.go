// Package bundle provides the bundle manifest loader and resolver for preset package groups.
package bundle

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

//go:embed bundles.json
var bundlesData []byte

// BundleDef is a named set of packages that can be installed together.
type BundleDef struct {
	Packages []string `json:"packages"`
}

// BundleManifest holds all defined bundles, keyed by bundle name.
type BundleManifest struct {
	Bundles map[string]BundleDef `json:"bundles"`
}

// LoadBundles parses and returns the embedded bundle manifest.
// Returns an error if the embedded JSON is malformed.
func LoadBundles() (*BundleManifest, error) {
	var m BundleManifest
	if err := json.Unmarshal(bundlesData, &m); err != nil {
		return nil, fmt.Errorf("parsing bundles.json: %w", err)
	}
	return &m, nil
}

// Resolve looks up a bundle by name and returns its package list.
// Returns a descriptive error listing available bundles if the name is not found.
func (m *BundleManifest) Resolve(name string) ([]string, error) {
	b, ok := m.Bundles[name]
	if !ok {
		return nil, fmt.Errorf("bundle %q not found; available: cloud, dev, context", name)
	}
	return b.Packages, nil
}
