package registry

import (
	"fmt"
	"strings"

	"github.com/ejyle/agentkit/internal/domain"
)

// builtinPackages lists packages that ship with agentkit regardless of external registries.
// Use $TARGET in Args to inject the agentkit --target value at install time.
var builtinPackages = []domain.Package{
	{
		Name:        "gsd",
		Version:     "latest",
		Description: "GSD (Get Shit Done) — AI workflow system: plan, execute, and ship development phases",
		Type:        domain.PackageTypeSkill,
		Install: domain.InstallSpec{
			Method:  domain.InstallMethodCustom,
			Package: "npx",
			Args:    []string{"@opengsd/gsd-core@latest", "--$TARGET", "--global"},
		},
	},
}

// BuiltinRegistry serves the hardcoded built-in package list.
// It is always queried last, after external registries.
type BuiltinRegistry struct{}

// Name returns "builtin".
func (b *BuiltinRegistry) Name() string { return "builtin" }

// Resolve returns the named package from builtinPackages, or an error if not found.
func (b *BuiltinRegistry) Resolve(name string) (*domain.Package, error) {
	lower := strings.ToLower(name)
	for _, p := range builtinPackages {
		if strings.ToLower(p.Name) == lower {
			cp := p
			return &cp, nil
		}
	}
	return nil, fmt.Errorf("%q not found in builtin registry", name)
}

// Search returns builtin packages whose name or description contains the query.
func (b *BuiltinRegistry) Search(query string) ([]domain.Package, error) {
	lower := strings.ToLower(query)
	var out []domain.Package
	for _, p := range builtinPackages {
		if strings.Contains(strings.ToLower(p.Name), lower) ||
			strings.Contains(strings.ToLower(p.Description), lower) {
			out = append(out, p)
		}
	}
	return out, nil
}
