package validation

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// JSONValidator provides validation functionality for JSON files
type JSONValidator struct {
	hookScript string
}

// NewJSONValidator creates a new JSON validator instance
func NewJSONValidator() *JSONValidator {
	return &JSONValidator{
		hookScript: ".claude/hooks/post-write-json-validator-simple.sh",
	}
}

// ValidateJSONFile validates a single JSON file using the validation hook
func (v *JSONValidator) ValidateJSONFile(filePath string) error {
	if !v.fileExists(filePath) {
		return nil // Skip validation if file doesn't exist
	}

	// Get absolute path to the hook script
	hookPath := filepath.Join(".", v.hookScript)
	if !v.fileExists(hookPath) {
		return fmt.Errorf("validation hook not found: %s", hookPath)
	}

	// Make sure hook is executable
	if err := os.Chmod(hookPath, 0755); err != nil {
		return fmt.Errorf("failed to make hook executable: %w", err)
	}

	// Run the validation hook
	cmd := exec.Command("bash", hookPath, filePath)
	output, err := cmd.CombinedOutput()
	
	if err != nil {
		return fmt.Errorf("JSON validation failed for %s: %s", filePath, string(output))
	}

	return nil
}

// ValidateAllProjectJSONs validates all JSON files in standard project locations
func (v *JSONValidator) ValidateAllProjectJSONs() error {
	jsonFiles := []string{
		"docs/1-project/epics.json",
		"docs/2-current-epic/current-epic.json",
		"docs/2-current-epic/stories.json",
		"docs/3-current-task/current-task.json",
		"docs/3-current-task/iterations.json",
		"docs/project/metrics.json",
	}

	var errors []string

	for _, file := range jsonFiles {
		if err := v.ValidateJSONFile(file); err != nil {
			errors = append(errors, err.Error())
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("JSON validation failed:\n%s", strings.Join(errors, "\n"))
	}

	return nil
}

// ValidateSpecificJSON validates JSON files specific to a command type
func (v *JSONValidator) ValidateSpecificJSON(jsonType string) error {
	var filePath string

	switch jsonType {
	case "epics":
		filePath = "docs/1-project/epics.json"
	case "stories":
		filePath = "docs/2-current-epic/stories.json"
	case "current-epic":
		filePath = "docs/2-current-epic/current-epic.json"
	case "current-story":
		filePath = "docs/2-current-epic/current-story.json"
	case "current-task":
		filePath = "docs/3-current-task/current-task.json"
	case "iterations":
		filePath = "docs/3-current-task/iterations.json"
	case "metrics":
		filePath = "docs/project/metrics.json"
	default:
		return fmt.Errorf("unknown JSON type: %s", jsonType)
	}

	return v.ValidateJSONFile(filePath)
}

// fileExists checks if a file exists
func (v *JSONValidator) fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

// ValidateOnStartup performs startup validation of all JSON files
func ValidateOnStartup() error {
	validator := NewJSONValidator()
	return validator.ValidateAllProjectJSONs()
}