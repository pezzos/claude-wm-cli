package config

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Manager handles the package manager style configuration system
type Manager struct {
	WorkspaceRoot string // .claude-wm root directory
	SystemPath    string // system/ - templates (read-only)
	UserPath      string // user/ - user overrides
	RuntimePath   string // runtime/ - effective config (generated)
}

// NewManager creates a new configuration manager
func NewManager(projectPath string) *Manager {
	workspaceRoot := filepath.Join(projectPath, ".claude-wm")
	return &Manager{
		WorkspaceRoot: workspaceRoot,
		SystemPath:    filepath.Join(workspaceRoot, "system"),
		UserPath:      filepath.Join(workspaceRoot, "user"),
		RuntimePath:   filepath.Join(workspaceRoot, "runtime"),
	}
}

// Initialize creates the package manager directory structure
func (m *Manager) Initialize() error {
	// Create base directories
	dirs := []string{
		m.SystemPath,
		m.UserPath,
		m.RuntimePath,
		filepath.Join(m.SystemPath, "hooks"),
		filepath.Join(m.SystemPath, "commands"),
		filepath.Join(m.SystemPath, "templates"),
		filepath.Join(m.UserPath, "hooks"),
		filepath.Join(m.UserPath, "commands"),
		filepath.Join(m.RuntimePath, "hooks"),
		filepath.Join(m.RuntimePath, "commands"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}

// InstallSystemTemplates installs default templates to system directory
func (m *Manager) InstallSystemTemplates() error {
	// Get the path to the embedded system templates (in the claude-wm-cli repo)
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}
	
	// Try to find the system templates relative to the executable
	execDir := filepath.Dir(execPath)
	embeddedSystemPath := filepath.Join(execDir, ".claude-wm", "system")
	
	// If not found relative to executable, try relative to working directory (for development)
	if _, err := os.Stat(embeddedSystemPath); os.IsNotExist(err) {
		cwd, _ := os.Getwd()
		embeddedSystemPath = filepath.Join(cwd, ".claude-wm", "system")
	}
	
	// Check if embedded system templates exist
	if _, err := os.Stat(embeddedSystemPath); os.IsNotExist(err) {
		// Fallback to basic templates if no embedded system found
		return m.installBasicTemplates()
	}

	// Copy embedded system templates to user's system directory
	if err := copyDir(embeddedSystemPath, m.SystemPath); err != nil {
		return fmt.Errorf("failed to copy system templates: %w", err)
	}

	return nil
}

// installBasicTemplates creates minimal templates when embedded system is not available
func (m *Manager) installBasicTemplates() error {
	// Create default settings.json template
	defaultSettings := map[string]interface{}{
		"version": "1.0.0",
		"hooks": map[string]interface{}{
			"PreToolUse":  []interface{}{},
			"PostToolUse": []interface{}{},
		},
		"permissions": map[string]interface{}{
			"allowed_tools": []string{"*"},
		},
		"env": map[string]string{},
	}

	settingsPath := filepath.Join(m.SystemPath, "settings.json.template")
	settingsData, err := json.MarshalIndent(defaultSettings, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal default settings: %w", err)
	}

	if err := os.WriteFile(settingsPath, settingsData, 0644); err != nil {
		return fmt.Errorf("failed to write settings template: %w", err)
	}

	// Create minimal template files
	templates := map[string]string{
		"commands/templates/current-task.json": `{
  "id": "",
  "title": "",
  "description": "",
  "status": "pending",
  "priority": "medium",
  "created": "",
  "updated": ""
}`,
		"commands/templates/iterations.json": `{
  "iteration": 1,
  "tasks": [],
  "status": "active",
  "created": "",
  "updated": ""
}`,
		"commands/templates/TEST.md": `# Test Template

This is a test template for claude-wm-cli.
`,
	}

	for relPath, content := range templates {
		fullPath := filepath.Join(m.SystemPath, relPath)
		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create template directory %s: %w", dir, err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write template %s: %w", relPath, err)
		}
	}

	return nil
}

// MigrateFromLegacy migrates files from the old .claude-wm/.claude structure
func (m *Manager) MigrateFromLegacy(legacyPath string) error {
	if _, err := os.Stat(legacyPath); os.IsNotExist(err) {
		return nil // Nothing to migrate
	}

	// Migrate settings.json if it exists
	legacySettings := filepath.Join(legacyPath, "settings.json")
	if _, err := os.Stat(legacySettings); err == nil {
		userSettings := filepath.Join(m.UserPath, "settings.json")
		if err := copyFile(legacySettings, userSettings); err != nil {
			return fmt.Errorf("failed to migrate settings.json: %w", err)
		}
	}

	// Migrate commands directory
	legacyCommands := filepath.Join(legacyPath, "commands")
	if _, err := os.Stat(legacyCommands); err == nil {
		systemCommands := filepath.Join(m.SystemPath, "commands")
		if err := copyDir(legacyCommands, systemCommands); err != nil {
			return fmt.Errorf("failed to migrate commands: %w", err)
		}
	}

	// Migrate hooks directory
	legacyHooks := filepath.Join(legacyPath, "hooks")
	if _, err := os.Stat(legacyHooks); err == nil {
		systemHooks := filepath.Join(m.SystemPath, "hooks")
		if err := copyDir(legacyHooks, systemHooks); err != nil {
			return fmt.Errorf("failed to migrate hooks: %w", err)
		}
	}

	return nil
}

// Sync generates the runtime configuration by merging system and user configs
func (m *Manager) Sync() error {
	// Merge settings
	if err := m.mergeSettings(); err != nil {
		return fmt.Errorf("failed to merge settings: %w", err)
	}

	// Merge commands
	if err := m.mergeDirectory("commands"); err != nil {
		return fmt.Errorf("failed to merge commands: %w", err)
	}

	// Merge hooks
	if err := m.mergeDirectory("hooks"); err != nil {
		return fmt.Errorf("failed to merge hooks: %w", err)
	}

	// Sync runtime configuration to .claude/ directory for Claude Code
	if err := m.syncToClaudeDir(); err != nil {
		return fmt.Errorf("failed to sync to .claude directory: %w", err)
	}

	return nil
}

// mergeSettings merges system template and user overrides
func (m *Manager) mergeSettings() error {
	// Load system template
	systemSettings := filepath.Join(m.SystemPath, "settings.json.template")
	var config map[string]interface{}

	if data, err := os.ReadFile(systemSettings); err == nil {
		if err := json.Unmarshal(data, &config); err != nil {
			return fmt.Errorf("failed to parse system settings: %w", err)
		}
	} else {
		config = make(map[string]interface{})
	}

	// Apply user overrides
	userSettings := filepath.Join(m.UserPath, "settings.json")
	if data, err := os.ReadFile(userSettings); err == nil {
		var userConfig map[string]interface{}
		if err := json.Unmarshal(data, &userConfig); err != nil {
			return fmt.Errorf("failed to parse user settings: %w", err)
		}
		// Deep merge user config into system config
		mergeMap(config, userConfig)
	}

	// Write runtime settings
	runtimeSettings := filepath.Join(m.RuntimePath, "settings.json")
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal runtime settings: %w", err)
	}

	return os.WriteFile(runtimeSettings, data, 0644)
}

// mergeDirectory merges system and user directories into runtime
func (m *Manager) mergeDirectory(dirName string) error {
	systemDir := filepath.Join(m.SystemPath, dirName)
	userDir := filepath.Join(m.UserPath, dirName)
	runtimeDir := filepath.Join(m.RuntimePath, dirName)

	// Clean runtime directory
	if err := os.RemoveAll(runtimeDir); err != nil {
		return fmt.Errorf("failed to clean runtime directory: %w", err)
	}
	if err := os.MkdirAll(runtimeDir, 0755); err != nil {
		return fmt.Errorf("failed to create runtime directory: %w", err)
	}

	// Copy system files first
	if _, err := os.Stat(systemDir); err == nil {
		if err := copyDir(systemDir, runtimeDir); err != nil {
			return fmt.Errorf("failed to copy system directory: %w", err)
		}
	}

	// Overlay user files (they override system files with same names)
	if _, err := os.Stat(userDir); err == nil {
		if err := copyDir(userDir, runtimeDir); err != nil {
			return fmt.Errorf("failed to overlay user directory: %w", err)
		}
	}

	return nil
}

// GetRuntimePath returns the path to a file in the runtime configuration
func (m *Manager) GetRuntimePath(relativePath string) string {
	return filepath.Join(m.RuntimePath, relativePath)
}

// GetUserPath returns the path to a file in the user configuration
func (m *Manager) GetUserPath(relativePath string) string {
	return filepath.Join(m.UserPath, relativePath)
}

// Utility functions

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	
	// Create destination directory if needed
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}
	
	return os.WriteFile(dst, data, 0644)
}

func copyDir(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Calculate relative path
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dst, relPath)

		if d.IsDir() {
			return os.MkdirAll(dstPath, 0755)
		}

		return copyFile(path, dstPath)
	})
}

// syncToClaudeDir copies the runtime configuration to .claude/ directory
func (m *Manager) syncToClaudeDir() error {
	// Get project root (parent of .claude-wm)
	projectRoot := filepath.Dir(m.WorkspaceRoot)
	claudeDir := filepath.Join(projectRoot, ".claude")

	// Clean and recreate .claude directory
	if err := os.RemoveAll(claudeDir); err != nil {
		return fmt.Errorf("failed to clean .claude directory: %w", err)
	}
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		return fmt.Errorf("failed to create .claude directory: %w", err)
	}

	// Copy runtime configuration to .claude with path corrections
	if err := m.copyDirWithPathCorrection(m.RuntimePath, claudeDir); err != nil {
		return fmt.Errorf("failed to copy runtime to .claude: %w", err)
	}

	return nil
}

// copyDirWithPathCorrection copies a directory while correcting path references in text files
func (m *Manager) copyDirWithPathCorrection(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Calculate relative path
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dst, relPath)

		if d.IsDir() {
			return os.MkdirAll(dstPath, 0755)
		}

		// For text files (.md, .sh, .json), correct path references
		ext := strings.ToLower(filepath.Ext(path))
		if ext == ".md" || ext == ".sh" || ext == ".json" {
			return m.copyFileWithPathCorrection(path, dstPath)
		}

		// For other files, copy directly
		return copyFile(path, dstPath)
	})
}

// copyFileWithPathCorrection copies a file while correcting path references
func (m *Manager) copyFileWithPathCorrection(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	// Replace .claude-wm/.claude/ references with .claude/
	content := string(data)
	content = strings.ReplaceAll(content, ".claude-wm/.claude/", ".claude/")

	// Create destination directory if needed
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	return os.WriteFile(dst, []byte(content), 0644)
}

func mergeMap(dst, src map[string]interface{}) {
	for key, value := range src {
		if srcMap, ok := value.(map[string]interface{}); ok {
			if dstMap, ok := dst[key].(map[string]interface{}); ok {
				mergeMap(dstMap, srcMap)
			} else {
				dst[key] = srcMap
			}
		} else {
			dst[key] = value
		}
	}
}