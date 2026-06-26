package service

import (
	"fmt"

	"github.com/ejyle/agentkit/internal/domain"
)

// updateRegistry is the local interface for registry resolution used by UpdateService.
type updateRegistry interface {
	Resolve(name string) (*domain.Package, error)
}

// updateStore is the local interface for config store methods needed by UpdateService.
type updateStore interface {
	GetRecord(name string) (domain.InstalledRecord, bool, error)
	ListInstalled() ([]domain.InstalledRecord, error)
}

// updateInstaller is the local interface for re-running an install (D-08 auto-overwrite).
type updateInstaller interface {
	Install(name, target string) (*domain.Package, error)
}

// UpdateService resolves the latest version from the registry and re-installs packages
// when a newer version is available. It delegates to InstallService for the actual install
// which satisfies D-08: ownership-confirmed auto-overwrite.
type UpdateService struct {
	registry  updateRegistry
	store     updateStore
	installer updateInstaller
}

// NewUpdateService constructs an UpdateService.
// reg, store, and installSvc accept any compatible implementation, enabling unit test injection.
func NewUpdateService(reg updateRegistry, store updateStore, installSvc updateInstaller) *UpdateService {
	return &UpdateService{
		registry:  reg,
		store:     store,
		installer: installSvc,
	}
}

// Update checks if a newer version of the named package is available in the registry.
// If the installed version matches the latest, it returns "already up to date".
// If a newer version is available, it calls installer.Install to upgrade (D-08).
//
// Returns a human-readable message and any error.
func (s *UpdateService) Update(name, target string) (string, error) {
	// Step 1: look up installed record.
	rec, found, err := s.store.GetRecord(name)
	if err != nil {
		return "", err
	}
	if !found {
		return "", ErrNotInstalled
	}

	// Step 2: resolve latest version from registry.
	latest, err := s.registry.Resolve(name)
	if err != nil {
		return "", err
	}

	// Step 3: if already at latest, no action needed.
	if latest.Version == rec.Version {
		return "already up to date", nil
	}

	// Step 4: re-run install (D-08 auto-overwrite — InstallService handles ownership check).
	if _, err := s.installer.Install(name, target); err != nil {
		return "", err
	}

	// Step 5: return a descriptive upgrade message.
	return fmt.Sprintf("updated %s: %s → %s", name, rec.Version, latest.Version), nil
}

// UpdateAll calls Update for every package recorded in installed.json for the target.
// It continues updating remaining packages even if one fails.
// Returns all result messages and the first non-nil error encountered.
func (s *UpdateService) UpdateAll(target string) ([]string, error) {
	records, err := s.store.ListInstalled()
	if err != nil {
		return nil, err
	}

	var msgs []string
	var firstErr error
	for _, rec := range records {
		msg, updateErr := s.Update(rec.Name, target)
		if updateErr != nil {
			if firstErr == nil {
				firstErr = updateErr
			}
			continue
		}
		msgs = append(msgs, msg)
	}
	return msgs, firstErr
}
