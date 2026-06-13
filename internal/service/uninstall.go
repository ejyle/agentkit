// Package service provides the core install/uninstall/update orchestration for agentkit.
package service

import (
	"errors"

	"github.com/ejyle/agentkit/internal/domain"
)

// ErrNotInstalled is returned when an operation targets a package that is not
// recorded in installed.json for the target assistant.
var ErrNotInstalled = errors.New("package not installed")

// uninstallAdapter is the local interface for adapter methods needed by UninstallService.
// Defined locally to enable injection of mocks in tests without importing the adapter package.
type uninstallAdapter interface {
	RemoveMCPConfig(name string) error
	RemoveSkill(name string) error
}

// uninstallStore is the local interface for config store methods needed by UninstallService.
type uninstallStore interface {
	GetRecord(name string) (domain.InstalledRecord, bool, error)
	RemoveRecord(name string) error
}

// UninstallService removes a package from the assistant config and the installed.json record.
// It implements D-09: non-destructive merge — only the named mcpServers key is removed;
// all other config keys and entries remain untouched.
type UninstallService struct {
	adapter uninstallAdapter
	store   uninstallStore
}

// NewUninstallService constructs an UninstallService.
// ad and store are typically *adapter.ClaudeCodeAdapter and *config.ConfigStore respectively,
// but any implementation of the local interfaces works (enabling unit test injection).
func NewUninstallService(ad uninstallAdapter, store uninstallStore) *UninstallService {
	return &UninstallService{adapter: ad, store: store}
}

// Uninstall removes the named package from the assistant config and from installed.json.
//
// Flow (D-09):
//  1. Look up the record in installed.json — return ErrNotInstalled if absent.
//  2. Call adapter.RemoveMCPConfig — return the error and halt if it fails.
//  3. For skill-type packages: call adapter.RemoveSkill — return error if it fails.
//  4. Call store.RemoveRecord to remove the tracking entry.
func (s *UninstallService) Uninstall(name string) error {
	rec, found, err := s.store.GetRecord(name)
	if err != nil {
		return err
	}
	if !found {
		return ErrNotInstalled
	}

	// Step 2: remove MCP config entry (D-09 non-destructive merge via adapter).
	if err := s.adapter.RemoveMCPConfig(name); err != nil {
		return err
	}

	// Step 3: for skill packages, remove the skill directory as well.
	if rec.Type == domain.PackageTypeSkill {
		if err := s.adapter.RemoveSkill(name); err != nil {
			return err
		}
	}

	// Step 4: remove the installed.json record only after adapter calls succeed.
	return s.store.RemoveRecord(name)
}
