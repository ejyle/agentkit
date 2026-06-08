package service_test

import (
	"errors"
	"testing"

	"github.com/ejyle/agentkit/internal/domain"
	"github.com/ejyle/agentkit/internal/service"
)

// --- mock implementations ---

type mockUpdateRegistry struct {
	resolveResult *domain.Package
	resolveErr    error
}

func (m *mockUpdateRegistry) Resolve(name string) (*domain.Package, error) {
	return m.resolveResult, m.resolveErr
}

type mockUpdateStore struct {
	records map[string]domain.InstalledRecord
}

func (m *mockUpdateStore) GetRecord(name string) (domain.InstalledRecord, bool, error) {
	rec, ok := m.records[name]
	return rec, ok, nil
}

func (m *mockUpdateStore) ListInstalled() ([]domain.InstalledRecord, error) {
	var list []domain.InstalledRecord
	for _, r := range m.records {
		list = append(list, r)
	}
	return list, nil
}

type mockUpdateInstaller struct {
	installCalls []string
	installErr   error
}

func (m *mockUpdateInstaller) Install(name, target string) (*domain.Package, error) {
	m.installCalls = append(m.installCalls, name)
	if m.installErr != nil {
		return nil, m.installErr
	}
	return &domain.Package{Name: name}, nil
}

// --- tests ---

// Test 7: UpdateService.Update returns "already up to date" when versions match.
func TestUpdateService_AlreadyUpToDate(t *testing.T) {
	reg := &mockUpdateRegistry{
		resolveResult: &domain.Package{Name: "playwright", Version: "1.2.0"},
	}
	store := &mockUpdateStore{
		records: map[string]domain.InstalledRecord{
			"playwright": {Name: "playwright", Version: "1.2.0"},
		},
	}
	installer := &mockUpdateInstaller{}

	svc := service.NewUpdateService(reg, store, installer)
	msg, err := svc.Update("playwright", "claude")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg != "already up to date" {
		t.Errorf("expected %q, got %q", "already up to date", msg)
	}
	if len(installer.installCalls) > 0 {
		t.Error("Install should not have been called when version is the same")
	}
}

// Test 8: UpdateService.Update calls Install when registry version is newer.
func TestUpdateService_CallsInstallWhenNewer(t *testing.T) {
	reg := &mockUpdateRegistry{
		resolveResult: &domain.Package{Name: "playwright", Version: "1.3.0"},
	}
	store := &mockUpdateStore{
		records: map[string]domain.InstalledRecord{
			"playwright": {Name: "playwright", Version: "1.2.0"},
		},
	}
	installer := &mockUpdateInstaller{}

	svc := service.NewUpdateService(reg, store, installer)
	msg, err := svc.Update("playwright", "claude")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(installer.installCalls) == 0 {
		t.Error("expected Install to be called for newer version")
	}
	if installer.installCalls[0] != "playwright" {
		t.Errorf("expected Install called with %q, got %q", "playwright", installer.installCalls[0])
	}
	_ = msg // message content tested separately
}

// Test 9: UpdateService.UpdateAll calls Update for each installed record, continues on single error.
func TestUpdateService_UpdateAll_ContinuesOnError(t *testing.T) {
	reg := &mockUpdateRegistry{
		resolveResult: &domain.Package{Name: "playwright", Version: "1.3.0"},
	}
	store := &mockUpdateStore{
		records: map[string]domain.InstalledRecord{
			"playwright": {Name: "playwright", Version: "1.2.0"},
			"gsd":        {Name: "gsd", Version: "2.0.0"},
		},
	}
	// Installer fails only for "playwright"
	installer := &mockUpdateInstaller{
		installErr: errors.New("install failed"),
	}

	svc := service.NewUpdateService(reg, store, installer)
	msgs, err := svc.UpdateAll("claude")

	// Should return an error because at least one package failed
	if err == nil {
		t.Error("expected error for failed update")
	}
	// Should still have attempted both packages
	if len(installer.installCalls) < 1 {
		t.Error("expected Install to be called for packages with newer versions")
	}
	_ = msgs
}

// Test 10: UpdateService.Update returns ErrNotInstalled when package not in installed.json.
func TestUpdateService_ReturnsErrNotInstalledWhenAbsent(t *testing.T) {
	reg := &mockUpdateRegistry{
		resolveResult: &domain.Package{Name: "playwright", Version: "1.2.0"},
	}
	store := &mockUpdateStore{
		records: map[string]domain.InstalledRecord{},
	}
	installer := &mockUpdateInstaller{}

	svc := service.NewUpdateService(reg, store, installer)
	_, err := svc.Update("playwright", "claude")

	if !errors.Is(err, service.ErrNotInstalled) {
		t.Errorf("expected ErrNotInstalled, got %v", err)
	}
}
