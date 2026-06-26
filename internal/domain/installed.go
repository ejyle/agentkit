package domain

import "time"

// InstalledRecord tracks a single package installed by agentkit (per D-11).
// JSON field names are snake_case to match the installed.json schema.
type InstalledRecord struct {
	Name        string      `json:"name"`
	Version     string      `json:"version"`
	Type        PackageType `json:"type"`
	InstallPath string      `json:"install_path"`
	InstalledAt time.Time   `json:"installed_at"`
	SourceURL   string      `json:"source_url"`
	Checksum    string      `json:"checksum"`
}

// InstalledState is the top-level structure of an installed.json file.
// Packages maps package name to its installed record.
type InstalledState struct {
	Packages map[string]InstalledRecord `json:"packages"`
}
