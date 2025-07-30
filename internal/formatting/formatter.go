package formatting

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Formatter handles code formatting for claude-wm-cli
type Formatter struct {
	projectRoot string
}

// NewFormatter creates a new formatter instance
func NewFormatter(projectRoot string) *Formatter {
	return &Formatter{
		projectRoot: projectRoot,
	}
}

// FormatAll formats all supported files in the project
func (f *Formatter) FormatAll() error {
	fmt.Println("Running auto-format...")

	// Format Go files
	if err := f.formatGo(); err != nil {
		return fmt.Errorf("go formatting failed: %v", err)
	}

	// Format JSON files (claude-wm-cli specific)
	if err := f.formatJSON(); err != nil {
		return fmt.Errorf("JSON formatting failed: %v", err)
	}

	fmt.Println("Auto-formatting completed ✓")
	return nil
}

// formatGo formats Go files using go fmt and goimports
func (f *Formatter) formatGo() error {
	// Check if this is a Go project
	goModPath := filepath.Join(f.projectRoot, "go.mod")
	if _, err := os.Stat(goModPath); os.IsNotExist(err) {
		return nil // Not a Go project
	}

	fmt.Println("Formatting Go files...")

	// Run go fmt
	cmd := exec.Command("go", "fmt", "./...")
	cmd.Dir = f.projectRoot
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go fmt failed: %v", err)
	}

	// Try to run goimports if available
	if f.hasCommand("goimports") {
		fmt.Println("Running goimports...")
		cmd = exec.Command("goimports", "-w", ".")
		cmd.Dir = f.projectRoot
		if err := cmd.Run(); err != nil {
			// Don't fail if goimports fails, just warn
			fmt.Printf("Warning: goimports failed: %v\n", err)
		}
	}

	return nil
}

// formatJSON formats JSON files with proper indentation
func (f *Formatter) formatJSON() error {
	fmt.Println("Formatting JSON files...")

	// JSON files to format in claude-wm-cli
	jsonPaths := []string{
		"docs/1-project/epics.json",
		"docs/2-current-epic/current-epic.json",
		"docs/2-current-epic/stories.json",
		"docs/2-current-epic/current-story.json",
		"docs/3-current-task/current-task.json",
		"docs/3-current-task/iterations.json",
		"docs/3-current-task/metrics.json",
	}

	for _, relPath := range jsonPaths {
		fullPath := filepath.Join(f.projectRoot, relPath)
		if err := f.formatJSONFile(fullPath); err != nil {
			// Don't fail for missing files, just continue
			continue
		}
	}

	return nil
}

// formatJSONFile formats a single JSON file
func (f *Formatter) formatJSONFile(filePath string) error {
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return err
	}

	// Read file
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	// Parse JSON
	var data interface{}
	if err := json.Unmarshal(content, &data); err != nil {
		return fmt.Errorf("invalid JSON in %s: %v", filePath, err)
	}

	// Format with proper indentation
	formatted, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	// Write back to file
	return os.WriteFile(filePath, formatted, 0644)
}

// hasCommand checks if a command is available in PATH
func (f *Formatter) hasCommand(command string) bool {
	_, err := exec.LookPath(command)
	return err == nil
}

// FormatFile formats a specific file based on its extension
func (f *Formatter) FormatFile(filePath string) error {
	ext := strings.ToLower(filepath.Ext(filePath))

	switch ext {
	case ".go":
		return f.formatGoFile(filePath)
	case ".json":
		return f.formatJSONFile(filePath)
	default:
		// No formatting needed for this file type
		return nil
	}
}

// formatGoFile formats a single Go file
func (f *Formatter) formatGoFile(filePath string) error {
	cmd := exec.Command("go", "fmt", filePath)
	cmd.Dir = f.projectRoot
	return cmd.Run()
}