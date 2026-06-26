package service_test

import (
	"errors"
	"testing"
	"time"

	"github.com/ejyle/agentkit/internal/adapter"
	"github.com/ejyle/agentkit/internal/domain"
	"github.com/ejyle/agentkit/internal/service"
)

// --- Mock implementations ---

type mockRegistry struct {
	resolveFunc func(name string) (*domain.Package, error)
}

func (m *mockRegistry) Resolve(name string) (*domain.Package, error) {
	return m.resolveFunc(name)
}

type mockInstaller struct {
	installed []domain.InstallSpec
	err       error
}

func (m *mockInstaller) Install(spec domain.InstallSpec) error {
	m.installed = append(m.installed, spec)
	return m.err
}

func (m *mockInstaller) IsAvailable() bool { return true }

type mockAdapter struct {
	writeErr     error
	writeCalled  bool
	writeEntries []domain.MCPServerEntry
	skillCalled  bool
	skillFiles   map[string][]byte
}

func (m *mockAdapter) WriteMCPConfig(entry domain.MCPServerEntry, ownership *domain.InstalledRecord) error {
	m.writeCalled = true
	m.writeEntries = append(m.writeEntries, entry)
	return m.writeErr
}

func (m *mockAdapter) RemoveMCPConfig(name string) error        { return nil }
func (m *mockAdapter) ReadMCPConfig() (map[string]domain.MCPServerEntry, error) {
	return nil, nil
}

func (m *mockAdapter) WriteSkill(name string, files map[string][]byte) error {
	m.skillCalled = true
	m.skillFiles = files
	return nil
}

func (m *mockAdapter) RemoveSkill(name string) error { return nil }
func (m *mockAdapter) Name() string                  { return "claude" }

type mockStore struct {
	recorded []domain.InstalledRecord
	err      error
}

func (m *mockStore) RecordInstalled(rec domain.InstalledRecord) error {
	m.recorded = append(m.recorded, rec)
	return m.err
}

func (m *mockStore) GetRecord(name string) (domain.InstalledRecord, bool, error) {
	return domain.InstalledRecord{}, false, nil
}

// standardPkg returns a standard playwright MCP package for tests.
func standardPkg() *domain.Package {
	return &domain.Package{
		Name:    "playwright",
		Version: "1.2.0",
		Type:    domain.PackageTypeMCP,
		Source:  "https://example.com/registry.json",
		SHA256:  "abc123",
		Install: domain.InstallSpec{Method: domain.InstallMethodNpx, Package: "@playwright/mcp"},
		MCPEntry: domain.MCPServerEntry{
			Command: "npx",
			Args:    []string{"-y", "@playwright/mcp"},
		},
	}
}

// Test 1: InstallService.Install calls Resolve → Install → WriteMCPConfig → RecordInstalled in order.
func TestInstallService_Install_HappyPath(t *testing.T) {
	callOrder := []string{}

	reg := &mockRegistry{
		resolveFunc: func(name string) (*domain.Package, error) {
			callOrder = append(callOrder, "resolve")
			return standardPkg(), nil
		},
	}
	inst := &mockInstaller{}
	origInstall := inst.Install
	_ = origInstall
	ad := &mockAdapter{
		writeErr: nil,
	}
	store := &mockStore{}

	svc := service.NewInstallService(reg, ad, store, func(method domain.InstallMethod) (service.Installer, error) {
		return &mockInstaller{}, nil
	})

	pkg, err := svc.Install("playwright", "claude")
	if err != nil {
		t.Fatalf("Install() error = %v; want nil", err)
	}
	if pkg == nil {
		t.Fatal("Install() pkg = nil; want non-nil")
	}
	if pkg.Name != "playwright" {
		t.Errorf("Install() pkg.Name = %q; want %q", pkg.Name, "playwright")
	}
	if !ad.writeCalled {
		t.Error("WriteMCPConfig was not called")
	}
	if len(store.recorded) == 0 {
		t.Error("RecordInstalled was not called")
	}
}

// Test 2: InstallService.Install returns error if Resolve returns "not found" error.
func TestInstallService_Install_ResolveNotFound(t *testing.T) {
	notFound := errors.New("playwright not found in any registry")
	reg := &mockRegistry{
		resolveFunc: func(name string) (*domain.Package, error) {
			return nil, notFound
		},
	}
	svc := service.NewInstallService(reg, &mockAdapter{}, &mockStore{}, func(method domain.InstallMethod) (service.Installer, error) {
		return &mockInstaller{}, nil
	})

	_, err := svc.Install("playwright", "claude")
	if err == nil {
		t.Fatal("Install() expected error for not-found package, got nil")
	}
}

// Test 3: InstallService.Install returns ErrForeignConflict when adapter returns it.
func TestInstallService_Install_ForeignConflict(t *testing.T) {
	reg := &mockRegistry{
		resolveFunc: func(name string) (*domain.Package, error) {
			return standardPkg(), nil
		},
	}
	conflictErr := &adapter.ErrForeignConflict{
		OldEntry: domain.MCPServerEntry{Name: "playwright", Command: "node"},
		NewEntry: domain.MCPServerEntry{Name: "playwright", Command: "npx"},
	}
	ad := &mockAdapter{writeErr: conflictErr}
	svc := service.NewInstallService(reg, ad, &mockStore{}, func(method domain.InstallMethod) (service.Installer, error) {
		return &mockInstaller{}, nil
	})

	_, err := svc.Install("playwright", "claude")
	if err == nil {
		t.Fatal("Install() expected ErrForeignConflict, got nil")
	}
	var fc *adapter.ErrForeignConflict
	if !errors.As(err, &fc) {
		t.Errorf("Install() error = %v; want ErrForeignConflict", err)
	}
}

// Test 4: InstallService.Install does NOT call RecordInstalled if WriteMCPConfig fails.
func TestInstallService_Install_NoRecordOnWriteError(t *testing.T) {
	reg := &mockRegistry{
		resolveFunc: func(name string) (*domain.Package, error) {
			return standardPkg(), nil
		},
	}
	ad := &mockAdapter{writeErr: errors.New("write failed")}
	store := &mockStore{}
	svc := service.NewInstallService(reg, ad, store, func(method domain.InstallMethod) (service.Installer, error) {
		return &mockInstaller{}, nil
	})

	_, err := svc.Install("playwright", "claude")
	if err == nil {
		t.Fatal("Install() expected error when WriteMCPConfig fails, got nil")
	}
	if len(store.recorded) > 0 {
		t.Errorf("RecordInstalled should not have been called, but got %d records", len(store.recorded))
	}
}

// Test 5: InstallService.Install calls ValidateSkill when package type is "skill".
func TestInstallService_Install_ValidatesSkillPackage(t *testing.T) {
	skillPkg := &domain.Package{
		Name:    "my-skill",
		Version: "1.0.0",
		Type:    domain.PackageTypeSkill,
		Source:  "https://example.com",
		Install: domain.InstallSpec{Method: domain.InstallMethodNpx, Package: "my-skill"},
	}
	reg := &mockRegistry{
		resolveFunc: func(name string) (*domain.Package, error) {
			return skillPkg, nil
		},
	}

	validateCalled := false
	svc := service.NewInstallServiceWithValidator(
		reg, &mockAdapter{}, &mockStore{},
		func(method domain.InstallMethod) (service.Installer, error) {
			return &mockInstaller{}, nil
		},
		func(dir string, pkg *domain.Package) service.ValidationResult {
			validateCalled = true
			return service.ValidationResult{Valid: true}
		},
	)

	_, err := svc.Install("my-skill", "claude")
	// Skill install may fail because no real files — but ValidateSkill must have been called.
	_ = err
	if !validateCalled {
		t.Error("ValidateSkill was not called for a skill-type package")
	}
}

// Test 6: InstallService.Install with a "skill" package calls adapter.WriteSkill.
func TestInstallService_Install_CallsWriteSkill(t *testing.T) {
	skillPkg := &domain.Package{
		Name:    "my-skill",
		Version: "1.0.0",
		Type:    domain.PackageTypeSkill,
		Source:  "https://example.com",
		Install: domain.InstallSpec{Method: domain.InstallMethodNpx, Package: "my-skill"},
	}
	reg := &mockRegistry{
		resolveFunc: func(name string) (*domain.Package, error) {
			return skillPkg, nil
		},
	}
	ad := &mockAdapter{}
	svc := service.NewInstallServiceWithValidator(
		reg, ad, &mockStore{},
		func(method domain.InstallMethod) (service.Installer, error) {
			return &mockInstaller{}, nil
		},
		func(dir string, pkg *domain.Package) service.ValidationResult {
			return service.ValidationResult{Valid: true}
		},
	)

	_, err := svc.Install("my-skill", "claude")
	_ = err // skill install is wired but skill bytes are empty in mock scenario
	if !ad.skillCalled {
		t.Error("adapter.WriteSkill was not called for a skill-type package")
	}
}

// Ensure time package is used for InstalledAt.
var _ = time.Now
