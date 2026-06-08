// Package skill provides utilities for validating agentkit skill packages.
package skill

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ejyle/agentkit/internal/domain"
)

// ValidationResult holds the outcome of a skill validation run.
type ValidationResult struct {
	// Valid reports whether the skill is structurally valid enough to be installed.
	// Warnings are non-blocking; Errors set Valid=false.
	Valid    bool
	Warnings []string
	Errors   []string
}

// ValidateSkill validates the skill directory against the SKL-01, SKL-02, SKL-03 rules:
//  1. SKILL.md must exist.
//  2. SKILL.md must not exceed 500 lines (warning, non-blocking).
//  3. Each name in manifest.Install.Args is treated as a required skill reference;
//     references/<name>.md must exist in dir.
//
// Install proceeds when Valid=true (warnings allowed). Errors set Valid=false.
func ValidateSkill(dir string, manifest *domain.Package) ValidationResult {
	result := ValidationResult{Valid: true}

	skillMD := filepath.Join(dir, "SKILL.md")
	f, err := os.Open(skillMD)
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, "SKILL.md missing: "+err.Error())
		return result // no point checking references if SKILL.md is absent
	}
	defer f.Close()

	// Count lines (SKL-02).
	lineCount := 0
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lineCount++
	}
	if scanner.Err() != nil {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("error reading SKILL.md: %v", scanner.Err()))
		return result
	}
	if lineCount > 500 {
		result.Warnings = append(result.Warnings,
			fmt.Sprintf("SKILL.md exceeds 500 lines (%d lines); consider splitting into references/", lineCount))
	}

	// Validate references/ files (SKL-03).
	// manifest.Install.Args carries the list of required reference names.
	if manifest != nil {
		for _, ref := range manifest.Install.Args {
			refPath := filepath.Join(dir, "references", ref+".md")
			if _, err := os.Stat(refPath); os.IsNotExist(err) {
				result.Valid = false
				result.Errors = append(result.Errors,
					fmt.Sprintf("missing reference file: references/%s.md", ref))
			}
		}
	}

	return result
}
