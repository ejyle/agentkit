// Package service provides the core install orchestration for agentkit.
package service

import (
	"fmt"
	"os"
	"time"

	"github.com/ejyle/agentkit/internal/config"
	"github.com/ejyle/agentkit/internal/domain"
	"github.com/ejyle/agentkit/internal/skill"
)

// Resolver resolves a package by name from the registry.
type Resolver interface {
	Resolve(name string) (*domain.Package, error)
}

// Installer runs the MCP or binary install step for a package.
type Installer interface {
	Install(spec domain.InstallSpec) error
	IsAvailable() bool
}

// AdapterWriter writes MCP config and skill files for the target assistant.
type AdapterWriter interface {
	WriteMCPConfig(entry domain.MCPServerEntry, ownership *domain.InstalledRecord) error
	WriteSkill(name string, files map[string][]byte) error
	Name() string
}

// Recorder records an install to the per-assistant state file.
type Recorder interface {
	RecordInstalled(rec domain.InstalledRecord) error
	GetRecord(name string) (domain.InstalledRecord, bool, error)
}

// InstallerFactory creates an Installer for the given install method.
type InstallerFactory func(method domain.InstallMethod) (Installer, error)

// ValidationResult mirrors skill.ValidationResult to avoid coupling callers to the skill package.
type ValidationResult struct {
	Valid    bool
	Warnings []string
	Errors   []string
}

// SkillValidator validates a skill directory against a package manifest.
type SkillValidator func(dir string, pkg *domain.Package) ValidationResult

// defaultSkillValidator wraps the real skill.ValidateSkill function.
func defaultSkillValidator(dir string, pkg *domain.Package) ValidationResult {
	r := skill.ValidateSkill(dir, pkg)
	return ValidationResult{Valid: r.Valid, Warnings: r.Warnings, Errors: r.Errors}
}

// InstallService orchestrates the full install flow:
//
//	Resolve → Run installer → Validate (if skill) → Write config → Record
type InstallService struct {
	registry  Resolver
	adapter   AdapterWriter
	store     Recorder
	newInst   InstallerFactory
	validator SkillValidator
}

// NewInstallService creates an InstallService with the default skill validator.
func NewInstallService(
	reg Resolver,
	ad AdapterWriter,
	store Recorder,
	factory InstallerFactory,
) *InstallService {
	return &InstallService{
		registry:  reg,
		adapter:   ad,
		store:     store,
		newInst:   factory,
		validator: defaultSkillValidator,
	}
}

// NewInstallServiceWithValidator creates an InstallService with an injected skill validator.
// Used in tests to avoid filesystem access.
func NewInstallServiceWithValidator(
	reg Resolver,
	ad AdapterWriter,
	store Recorder,
	factory InstallerFactory,
	validator SkillValidator,
) *InstallService {
	return &InstallService{
		registry:  reg,
		adapter:   ad,
		store:     store,
		newInst:   factory,
		validator: validator,
	}
}

// Install executes the 9-step install flow (per plan spec):
//
//  1. Resolve package from registry
//  2. Create installer for install method
//  3. Run installer
//  4. If skill: validate, emit warnings to stderr, return error on validation errors
//  5. Build MCPServerEntry
//  6. Write MCP config via adapter (returns ErrForeignConflict on ownership conflict)
//  7. Build InstalledRecord
//  8. Record to installed.json
//  9. Return package
func (s *InstallService) Install(name, target string) (*domain.Package, error) {
	// Step 1: Resolve package.
	pkg, err := s.registry.Resolve(name)
	if err != nil {
		return nil, err
	}

	// Step 2: Create installer.
	inst, err := s.newInst(pkg.Install.Method)
	if err != nil {
		return nil, err
	}

	// Step 3: Populate SkillDir for filesystem-extraction install methods.
	if pkg.Install.Method == domain.InstallMethodGitHubRelease ||
		pkg.Install.Method == domain.InstallMethodGitHubDefaultBranch {
		skillDir, err := config.SkillInstallPath(target, name)
		if err != nil {
			return nil, fmt.Errorf("resolving skill install path for %q: %w", name, err)
		}
		pkg.Install.SkillDir = skillDir
	}

	// Step 3a: Substitute $TARGET in Args for custom packages that are target-aware.
	pkg.Install = substituteTarget(pkg.Install, target)

	// Step 3b: Run installer.
	if err := inst.Install(pkg.Install); err != nil {
		return nil, fmt.Errorf("install failed for %q: %w", name, err)
	}

	// Step 4: Skill validation (non-blocking warnings, blocking errors).
	if pkg.Type == domain.PackageTypeSkill {
		// For filesystem-extraction methods, validate the extracted directory.
		// For other methods (mock/unit tests), dir is empty — validator handles gracefully.
		validationDir := ""
		if pkg.Install.Method == domain.InstallMethodGitHubRelease ||
			pkg.Install.Method == domain.InstallMethodGitHubDefaultBranch {
			validationDir = pkg.Install.SkillDir
		}
		result := s.validator(validationDir, pkg)
		for _, w := range result.Warnings {
			fmt.Fprintf(os.Stderr, "warning: %s\n", w)
		}
		if !result.Valid {
			for _, e := range result.Errors {
				return nil, fmt.Errorf("validating installed skill %q: %s", name, e)
			}
		}

		// Step 6b for skills: call WriteSkill with SKILL.md bytes.
		// For filesystem-extraction methods the installer already placed all files — skip
		// WriteSkill to avoid overwriting the extracted content.
		if pkg.Install.Method != domain.InstallMethodGitHubRelease &&
			pkg.Install.Method != domain.InstallMethodGitHubDefaultBranch {
			if err := s.adapter.WriteSkill(name, map[string][]byte{
				"SKILL.md": []byte(""),
			}); err != nil {
				return nil, fmt.Errorf("writing skill files for %q: %w", name, err)
			}
		}
	}

	// Step 5: Build MCPServerEntry.
	entry := domain.MCPServerEntry{
		Name:    name,
		Command: pkg.MCPEntry.Command,
		Args:    pkg.MCPEntry.Args,
		Env:     pkg.MCPEntry.Env,
	}

	// Step 6: Write MCP config — may return ErrForeignConflict.
	if pkg.Type != domain.PackageTypeSkill {
		if err := s.adapter.WriteMCPConfig(entry, nil); err != nil {
			return nil, err
		}
	}

	// Step 7–8: Record install.
	rec := domain.InstalledRecord{
		Name:        name,
		Version:     pkg.Version,
		Type:        pkg.Type,
		InstallPath: "mcpServers." + name,
		InstalledAt: time.Now().UTC(),
		SourceURL:   pkg.Source,
		Checksum:    "sha256:" + pkg.SHA256,
	}
	if err := s.store.RecordInstalled(rec); err != nil {
		return nil, fmt.Errorf("recording install for %q: %w", name, err)
	}

	// Step 9: Return package.
	return pkg, nil
}

// substituteTarget replaces the literal string "$TARGET" in spec.Args with the
// target-appropriate flag so packages like gsd can pass --claude/--gemini/--codex
// to their installers without hardcoding one runtime.
func substituteTarget(spec domain.InstallSpec, target string) domain.InstallSpec {
	flag := targetFlag(target)
	args := make([]string, len(spec.Args))
	for i, a := range spec.Args {
		if a == "--$TARGET" {
			args[i] = flag
		} else {
			args[i] = a
		}
	}
	spec.Args = args
	return spec
}

// targetFlag maps an agentkit target name to the appropriate runtime flag
// for external installers (e.g. @opengsd/gsd-core@latest).
func targetFlag(target string) string {
	switch target {
	case "claude":
		return "--claude"
	case "gemini":
		return "--gemini"
	case "codex":
		return "--codex"
	case "opencode":
		return "--opencode"
	case "copilot-cli", "copilot-vscode":
		return "--copilot"
	default:
		return "--claude"
	}
}

