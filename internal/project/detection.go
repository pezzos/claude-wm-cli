package project

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// ProjectStatus represents the initialization status of a Claude WM project
type ProjectStatus int

const (
	// NotInitialized indicates no project structure exists
	NotInitialized ProjectStatus = iota
	// Partial indicates incomplete project structure or missing files
	Partial
	// Complete indicates a fully initialized project
	Complete
)

// String returns a string representation of the project status
func (ps ProjectStatus) String() string {
	switch ps {
	case NotInitialized:
		return "not_initialized"
	case Partial:
		return "partial"
	case Complete:
		return "complete"
	default:
		return "unknown"
	}
}

// ProjectDetectionResult contains the results of project initialization detection
type ProjectDetectionResult struct {
	Status       ProjectStatus `json:"status"`
	Issues       []string      `json:"issues,omitempty"`
	MissingFiles []string      `json:"missing_files,omitempty"`
	RootPath     string        `json:"root_path"`
	HasStructure bool          `json:"has_structure"`
	HasFiles     bool          `json:"has_files"`
}

// RequiredDirectories defines the expected project directory structure
var RequiredDirectories = []string{
	"docs/1-project",
	"docs/2-current-epic",
	"docs/3-current-task",
}

// RequiredFiles defines the critical files that must exist for a complete project
var RequiredFiles = []string{
	"docs/1-project/epics.json",
}

// OptionalFiles defines files that may exist in different project states
var OptionalFiles = []string{
	"docs/2-current-epic/current-epic.json",
	"docs/2-current-epic/stories.json",
}

// DetectProjectInitialization analyzes the current directory to determine
// if it contains a properly initialized Claude WM project
func DetectProjectInitialization(rootPath string) (*ProjectDetectionResult, error) {
	result := &ProjectDetectionResult{
		RootPath: rootPath,
		Issues:   []string{},
		MissingFiles: []string{},
	}

	// Check if root path exists and is accessible
	if _, err := os.Stat(rootPath); err != nil {
		if os.IsNotExist(err) {
			result.Issues = append(result.Issues, fmt.Sprintf("Root path does not exist: %s", rootPath))
			result.Status = NotInitialized
			return result, nil
		}
		return nil, fmt.Errorf("failed to access root path %s: %w", rootPath, err)
	}

	// Check directory structure
	hasStructure, structureIssues := checkDocsStructure(rootPath)
	result.HasStructure = hasStructure
	result.Issues = append(result.Issues, structureIssues...)

	// Check required files
	hasFiles, missingFiles := validateRequiredFiles(rootPath)
	result.HasFiles = hasFiles
	result.MissingFiles = append(result.MissingFiles, missingFiles...)

	// Check optional files and add informational messages
	optionalFileStatus := checkOptionalFiles(rootPath)
	if len(optionalFileStatus) > 0 {
		result.Issues = append(result.Issues, optionalFileStatus...)
	}

	// Determine overall status
	result.Status = determineInitializationStatus(hasStructure, hasFiles)

	return result, nil
}

// checkDocsStructure verifies that the required directory structure exists
func checkDocsStructure(rootPath string) (bool, []string) {
	issues := []string{}
	missingDirs := 0

	// Check if docs/ directory exists
	docsPath := filepath.Join(rootPath, "docs")
	if _, err := os.Stat(docsPath); err != nil {
		if os.IsNotExist(err) {
			issues = append(issues, "docs/ directory does not exist")
			return false, issues
		}
		issues = append(issues, fmt.Sprintf("Cannot access docs/ directory: %v", err))
		return false, issues
	}

	// Check each required subdirectory
	for _, dir := range RequiredDirectories {
		fullPath := filepath.Join(rootPath, dir)
		if _, err := os.Stat(fullPath); err != nil {
			if os.IsNotExist(err) {
				issues = append(issues, fmt.Sprintf("Missing directory: %s", dir))
				missingDirs++
			} else {
				issues = append(issues, fmt.Sprintf("Cannot access directory %s: %v", dir, err))
				missingDirs++
			}
		}
	}

	// Structure is considered present if at least half the directories exist
	hasStructure := missingDirs <= len(RequiredDirectories)/2
	
	return hasStructure, issues
}

// validateRequiredFiles checks that critical project files exist and are readable
func validateRequiredFiles(rootPath string) (bool, []string) {
	missingFiles := []string{}

	for _, file := range RequiredFiles {
		fullPath := filepath.Join(rootPath, file)
		
		// Check if file exists
		if _, err := os.Stat(fullPath); err != nil {
			if os.IsNotExist(err) {
				missingFiles = append(missingFiles, file)
				continue
			}
			missingFiles = append(missingFiles, fmt.Sprintf("%s (access error: %v)", file, err))
			continue
		}

		// Check if file is readable and valid JSON
		if err := validateJSONFile(fullPath); err != nil {
			missingFiles = append(missingFiles, fmt.Sprintf("%s (invalid JSON: %v)", file, err))
		}
	}

	hasFiles := len(missingFiles) == 0
	return hasFiles, missingFiles
}

// checkOptionalFiles provides status information about optional files
func checkOptionalFiles(rootPath string) []string {
	status := []string{}

	// Check for current epic state files
	currentEpicPath := filepath.Join(rootPath, "docs/2-current-epic/current-epic.json")
	storiesPath := filepath.Join(rootPath, "docs/2-current-epic/stories.json")

	hasCurrentEpic := fileExists(currentEpicPath)
	hasStories := fileExists(storiesPath)

	if !hasCurrentEpic && !hasStories {
		status = append(status, "No current epic or stories file found (project may be in initial state)")
	} else if hasCurrentEpic && !hasStories {
		status = append(status, "Current epic defined but no stories file found")
	} else if !hasCurrentEpic && hasStories {
		status = append(status, "Stories file found but no current epic defined")
	}

	return status
}

// validateJSONFile checks if a file contains valid JSON
func validateJSONFile(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("cannot read file: %w", err)
	}

	var jsonData interface{}
	if err := json.Unmarshal(data, &jsonData); err != nil {
		return fmt.Errorf("invalid JSON format: %w", err)
	}

	return nil
}

// fileExists checks if a file exists and is readable
func fileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return err == nil
}

// determineInitializationStatus analyzes the structure and file status to determine overall project status
func determineInitializationStatus(hasStructure, hasFiles bool) ProjectStatus {
	if !hasStructure {
		return NotInitialized
	}

	if hasStructure && hasFiles {
		return Complete
	}

	// Has structure but missing files = partial initialization
	return Partial
}

// GetRequiredFilesForCompletion returns a list of files needed to complete project initialization
func GetRequiredFilesForCompletion(rootPath string) []string {
	result, err := DetectProjectInitialization(rootPath)
	if err != nil || result.Status == Complete {
		return []string{}
	}

	missing := []string{}
	missing = append(missing, result.MissingFiles...)

	// If we have no structure, also recommend creating directories
	if !result.HasStructure {
		for _, dir := range RequiredDirectories {
			missing = append(missing, dir+"/")
		}
	}

	return missing
}