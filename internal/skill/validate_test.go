package skill_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ejyle/agentkit/internal/domain"
	"github.com/ejyle/agentkit/internal/skill"
)

// makeSkillDir creates a temp directory for skill testing.
func makeSkillDir(t *testing.T) string {
	t.Helper()
	return t.TempDir()
}

// Test 15: ValidateSkill() on a dir with SKILL.md returns Valid=true.
func TestValidateSkill_ValidSkill(t *testing.T) {
	dir := makeSkillDir(t)
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("# My Skill\nThis is the skill."), 0644); err != nil {
		t.Fatal(err)
	}
	pkg := &domain.Package{Type: domain.PackageTypeSkill}
	result := skill.ValidateSkill(dir, pkg)
	if !result.Valid {
		t.Errorf("ValidateSkill() Valid=false; want true. Errors: %v", result.Errors)
	}
}

// Test 16: ValidateSkill() on a dir without SKILL.md returns Valid=false, Error contains "SKILL.md missing".
func TestValidateSkill_MissingSkillMD(t *testing.T) {
	dir := makeSkillDir(t)
	pkg := &domain.Package{Type: domain.PackageTypeSkill}
	result := skill.ValidateSkill(dir, pkg)
	if result.Valid {
		t.Error("ValidateSkill() Valid=true; want false for missing SKILL.md")
	}
	found := false
	for _, e := range result.Errors {
		if strings.Contains(e, "SKILL.md missing") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("ValidateSkill() errors = %v; want one containing %q", result.Errors, "SKILL.md missing")
	}
}

// Test 17: ValidateSkill() on SKILL.md with 501 lines returns Valid=true and Warning contains "exceeds 500 lines".
func TestValidateSkill_ExceedsLineCount(t *testing.T) {
	dir := makeSkillDir(t)
	lines := make([]string, 501)
	for i := range lines {
		lines[i] = "line content"
	}
	content := strings.Join(lines, "\n")
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	pkg := &domain.Package{Type: domain.PackageTypeSkill}
	result := skill.ValidateSkill(dir, pkg)
	if !result.Valid {
		t.Errorf("ValidateSkill() Valid=false; want true (warnings are non-blocking)")
	}
	found := false
	for _, w := range result.Warnings {
		if strings.Contains(w, "exceeds 500 lines") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("ValidateSkill() warnings = %v; want one containing %q", result.Warnings, "exceeds 500 lines")
	}
}

// Test 18: ValidateSkill() on a manifest with references validates that references/*.md files exist.
func TestValidateSkill_ReferencesExist(t *testing.T) {
	dir := makeSkillDir(t)
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("# Skill"), 0644); err != nil {
		t.Fatal(err)
	}
	refsDir := filepath.Join(dir, "references")
	if err := os.MkdirAll(refsDir, 0755); err != nil {
		t.Fatal(err)
	}
	// Create aws.md but NOT gcp.md.
	if err := os.WriteFile(filepath.Join(refsDir, "aws.md"), []byte("# AWS"), 0644); err != nil {
		t.Fatal(err)
	}

	// Package with references field specifying aws and gcp.
	pkg := &domain.Package{
		Type: domain.PackageTypeSkill,
		Install: domain.InstallSpec{
			Args: []string{"aws", "gcp"},
		},
	}
	result := skill.ValidateSkill(dir, pkg)
	// Should be invalid because gcp.md is missing.
	if result.Valid {
		t.Error("ValidateSkill() Valid=true; want false because references/gcp.md is missing")
	}
	found := false
	for _, e := range result.Errors {
		if strings.Contains(e, "gcp") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("ValidateSkill() errors = %v; want one mentioning missing 'gcp' reference", result.Errors)
	}
}
