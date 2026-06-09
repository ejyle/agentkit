package bundle_test

import (
	"testing"

	"github.com/ejyle/agentkit/internal/bundle"
)

func TestLoadBundles_OK(t *testing.T) {
	m, err := bundle.LoadBundles()
	if err != nil {
		t.Fatalf("LoadBundles() returned error: %v", err)
	}
	if m == nil {
		t.Fatal("LoadBundles() returned nil manifest")
	}
	if len(m.Bundles) != 3 {
		t.Errorf("expected 3 bundles, got %d", len(m.Bundles))
	}
}

func TestResolveCloud(t *testing.T) {
	m, err := bundle.LoadBundles()
	if err != nil {
		t.Fatalf("LoadBundles() returned error: %v", err)
	}
	pkgs, err := m.Resolve("cloud")
	if err != nil {
		t.Fatalf("Resolve(\"cloud\") returned error: %v", err)
	}
	want := []string{"aws", "gcp", "azure"}
	if len(pkgs) != len(want) {
		t.Fatalf("Resolve(\"cloud\") returned %d packages, want %d", len(pkgs), len(want))
	}
	for i, w := range want {
		if pkgs[i] != w {
			t.Errorf("Resolve(\"cloud\")[%d] = %q, want %q", i, pkgs[i], w)
		}
	}
}

func TestResolveUnknown(t *testing.T) {
	m, err := bundle.LoadBundles()
	if err != nil {
		t.Fatalf("LoadBundles() returned error: %v", err)
	}
	_, err = m.Resolve("foo")
	if err == nil {
		t.Fatal("Resolve(\"foo\") expected error, got nil")
	}
}
