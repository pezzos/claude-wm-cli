package meta

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type Meta struct {
	Product          string `json:"product"`
	InstalledVersion string `json:"installed_version"`
	InstalledAt      string `json:"installed_at"`
	Schema           int    `json:"schema"`
}

// Load reads and parses a meta.json file from the given path
func Load(path string) (*Meta, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read meta file %s: %w", path, err)
	}

	var meta Meta
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, fmt.Errorf("failed to parse meta file %s: %w", path, err)
	}

	return &meta, nil
}

// Save writes a Meta struct to a JSON file at the given path, creating parent directories if needed
func Save(path string, m *Meta) error {
	// Create parent directories if they don't exist
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create directories for %s: %w", path, err)
	}

	// Marshal the meta struct to JSON
	data, err := json.Marshal(m)
	if err != nil {
		return fmt.Errorf("failed to marshal meta: %w", err)
	}

	// Write to file with 644 permissions
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write meta file %s: %w", path, err)
	}

	return nil
}

// Default creates a new Meta struct with the given product and version,
// setting InstalledAt to the current UTC time and Schema to 1
func Default(product, version string) *Meta {
	return &Meta{
		Product:          product,
		InstalledVersion: version,
		InstalledAt:      time.Now().UTC().Format(time.RFC3339),
		Schema:           1,
	}
}