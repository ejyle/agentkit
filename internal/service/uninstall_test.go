package service_test

import (
	"errors"
	"testing"

	"github.com/ejyle/agentkit/internal/domain"
	"github.com/ejyle/agentkit/internal/service"
)

// --- mock implementations ---

type mockUninstallAdapter struct {
	removeMCPConfigCalled bool
	removeMCPConfigName   string
	removeMCPConfigErr    error
	removeSkillCalled     bool
	removeSkillName       string
	removeSkillErr        error
}

func (m *mockUninstallAdapter) RemoveMCPConfig(name string) error {
	m.removeMCPConfigCalled = true
	m.removeMCPConfigName = name
	return m.removeMCPConfigErr
}

func (m *mockUninstallAdapter) RemoveSkill(name string) error {
	m.removeSkillCalled = true
	m.removeSkillName = name
	return m.removeSkillErr
}

type mockUninstallStore struct {
	records          map[string]domain.InstalledRecord
	removeRecordName string
	removeRecordErr  error
	removeCalled     bool
}

func (m *mockUninstallStore) GetRecord(name string) (domain.InstalledRecord, bool, error) {
	rec, ok := m.records[name]
	return rec, ok, nil
}

func (m *mockUninstallStore) RemoveRecord(name string) error {
	m.removeCalled = true
	m.removeRecordName = name
	return m.removeRecordErr
}

// --- tests ---

// Test 1: UninstallService.Uninstall calls adapter.RemoveMCPConfig with correct name.
func TestUninstallService_CallsRemoveMCPConfig(t *testing.T) {
	ad := &mockUninstallAdapter{}
	store := &mockUninstallStore{
		records: map[string]domain.InstalledRecord{
			"playwright": {Name: "playwright", Type: domain.PackageTypeMCP},
		},
	}

	svc := service.NewUninstallService(ad, store)
	if err := svc.Uninstall("playwright"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !ad.removeMCPConfigCalled {
		t.Error("expected RemoveMCPConfig to be called")
	}
	if ad.removeMCPConfigName != "playwright" {
		t.Errorf("expected RemoveMCPConfig called with %q, got %q", "playwright", ad.removeMCPConfigName)
	}
}

// Test 2: UninstallService.Uninstall calls adapter.RemoveSkill for skill type.
func TestUninstallService_CallsRemoveSkillForSkillType(t *testing.T) {
	ad := &mockUninstallAdapter{}
	store := &mockUninstallStore{
		records: map[string]domain.InstalledRecord{
			"gsd": {Name: "gsd", Type: domain.PackageTypeSkill},
		},
	}

	svc := service.NewUninstallService(ad, store)
	if err := svc.Uninstall("gsd"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !ad.removeSkillCalled {
		t.Error("expected RemoveSkill to be called for skill type")
	}
	if ad.removeSkillName != "gsd" {
		t.Errorf("expected RemoveSkill called with %q, got %q", "gsd", ad.removeSkillName)
	}
}

// Test 3: UninstallService.Uninstall calls store.RemoveRecord after adapter calls succeed.
func TestUninstallService_CallsRemoveRecord(t *testing.T) {
	ad := &mockUninstallAdapter{}
	store := &mockUninstallStore{
		records: map[string]domain.InstalledRecord{
			"playwright": {Name: "playwright", Type: domain.PackageTypeMCP},
		},
	}

	svc := service.NewUninstallService(ad, store)
	if err := svc.Uninstall("playwright"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !store.removeCalled {
		t.Error("expected RemoveRecord to be called")
	}
	if store.removeRecordName != "playwright" {
		t.Errorf("expected RemoveRecord called with %q, got %q", "playwright", store.removeRecordName)
	}
}

// Test 4: UninstallService.Uninstall returns ErrNotInstalled when package not in store.
func TestUninstallService_ReturnsErrNotInstalledWhenAbsent(t *testing.T) {
	ad := &mockUninstallAdapter{}
	store := &mockUninstallStore{
		records: map[string]domain.InstalledRecord{},
	}

	svc := service.NewUninstallService(ad, store)
	err := svc.Uninstall("playwright")

	if !errors.Is(err, service.ErrNotInstalled) {
		t.Errorf("expected ErrNotInstalled, got %v", err)
	}
}

// Test 5: UninstallService.Uninstall does NOT call RemoveRecord if RemoveMCPConfig fails.
func TestUninstallService_DoesNotRemoveRecordOnAdapterError(t *testing.T) {
	adapterErr := errors.New("adapter error")
	ad := &mockUninstallAdapter{
		removeMCPConfigErr: adapterErr,
	}
	store := &mockUninstallStore{
		records: map[string]domain.InstalledRecord{
			"playwright": {Name: "playwright", Type: domain.PackageTypeMCP},
		},
	}

	svc := service.NewUninstallService(ad, store)
	err := svc.Uninstall("playwright")

	if err == nil {
		t.Fatal("expected error but got nil")
	}
	if store.removeCalled {
		t.Error("RemoveRecord should NOT have been called after RemoveMCPConfig failure")
	}
}

// Test 6: UninstallService.Uninstall calls both RemoveMCPConfig and RemoveSkill when type is "skill".
func TestUninstallService_CallsBothRemovesForSkillType(t *testing.T) {
	ad := &mockUninstallAdapter{}
	store := &mockUninstallStore{
		records: map[string]domain.InstalledRecord{
			"gsd": {Name: "gsd", Type: domain.PackageTypeSkill},
		},
	}

	svc := service.NewUninstallService(ad, store)
	if err := svc.Uninstall("gsd"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !ad.removeMCPConfigCalled {
		t.Error("expected RemoveMCPConfig to be called")
	}
	if !ad.removeSkillCalled {
		t.Error("expected RemoveSkill to be called")
	}
}
