package config

import (
	"os"
	"path/filepath"
)

// GetConfigManager returns a configuration manager for the current directory
func GetConfigManager() (*Manager, error) {
	projectPath, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	return NewManager(projectPath), nil
}

// GetRuntimeConfigPath returns the path to a runtime configuration file
func GetRuntimeConfigPath(relativePath string) (string, error) {
	manager, err := GetConfigManager()
	if err != nil {
		return "", err
	}
	return manager.GetRuntimePath(relativePath), nil
}

// GetUserConfigPath returns the path to a user configuration file
func GetUserConfigPath(relativePath string) (string, error) {
	manager, err := GetConfigManager()
	if err != nil {
		return "", err
	}
	return manager.GetUserPath(relativePath), nil
}

// IsConfigInitialized checks if the package manager structure exists
func IsConfigInitialized(projectPath string) bool {
	manager := NewManager(projectPath)
	_, err := os.Stat(manager.RuntimePath)
	return err == nil
}

// EnsureConfigInitialized initializes config if not already done
func EnsureConfigInitialized(projectPath string) error {
	if IsConfigInitialized(projectPath) {
		return nil
	}

	manager := NewManager(projectPath)
	
	// Initialize directory structure
	if err := manager.Initialize(); err != nil {
		return err
	}

	// Install default templates
	if err := manager.InstallSystemTemplates(); err != nil {
		return err
	}

	// Migrate from legacy structure if it exists
	legacyPath := filepath.Join(projectPath, ".claude-wm", ".claude")
	if err := manager.MigrateFromLegacy(legacyPath); err != nil {
		return err
	}

	// Generate runtime configuration
	return manager.Sync()
}