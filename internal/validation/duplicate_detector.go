package validation

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// FileInfo represents information about a file for duplicate detection
type FileInfo struct {
	Path         string `json:"path"`
	RelativePath string `json:"relative_path"`
	Name         string `json:"name"`
	Extension    string `json:"extension"`
	Type         string `json:"type"` // command, handler, manager, service, etc.
}

// DuplicateDetector handles duplicate detection for Go files in claude-wm-cli
type DuplicateDetector struct {
	projectRoot string
	startTime   time.Time
}

// Result represents the detection result
type Result struct {
	Success      bool     `json:"success"`
	Errors       []string `json:"errors"`
	Warnings     []string `json:"warnings"`
	Duration     int64    `json:"duration_ms"`
	FilesScanned int      `json:"files_scanned"`
}

// NewDuplicateDetector creates a new duplicate detector for claude-wm-cli
func NewDuplicateDetector(projectRoot string) *DuplicateDetector {
	return &DuplicateDetector{
		projectRoot: projectRoot,
		startTime:   time.Now(),
	}
}

// DetectDuplicates performs duplicate detection for Go files
func (dd *DuplicateDetector) DetectDuplicates(filePath string) *Result {
	result := &Result{
		Success:  true,
		Errors:   []string{},
		Warnings: []string{},
	}

	// Only process Go files
	if !strings.HasSuffix(filePath, ".go") {
		result.Duration = time.Since(dd.startTime).Milliseconds()
		return result
	}

	// Skip test files
	if strings.HasSuffix(filePath, "_test.go") {
		result.Duration = time.Since(dd.startTime).Milliseconds()
		return result
	}

	// Get relative path
	relPath, err := filepath.Rel(dd.projectRoot, filePath)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to get relative path: %v", err))
		result.Success = false
		return result
	}

	// Create file info for target file
	targetFile := &FileInfo{
		Path:         filePath,
		RelativePath: relPath,
		Name:         strings.TrimSuffix(filepath.Base(filePath), ".go"),
		Extension:    ".go",
	}

	// Classify the file type
	dd.classifyGoFile(targetFile)

	// Find existing Go files to compare against
	existingFiles, err := dd.findGoFiles()
	if err != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf("Error scanning files: %v", err))
	}

	result.FilesScanned = len(existingFiles)

	// Check for duplicates
	dd.checkForDuplicates(targetFile, existingFiles, result)

	result.Duration = time.Since(dd.startTime).Milliseconds()
	return result
}

// classifyGoFile determines the type of Go file
func (dd *DuplicateDetector) classifyGoFile(fileInfo *FileInfo) {
	path := strings.ToLower(fileInfo.RelativePath)

	// Classify based on path and name
	if strings.Contains(path, "cmd/") {
		fileInfo.Type = "command"
	} else if strings.Contains(path, "internal/") {
		if strings.HasSuffix(fileInfo.Name, "manager") {
			fileInfo.Type = "manager"
		} else if strings.HasSuffix(fileInfo.Name, "service") {
			fileInfo.Type = "service"
		} else if strings.HasSuffix(fileInfo.Name, "handler") {
			fileInfo.Type = "handler"
		} else if strings.HasSuffix(fileInfo.Name, "repository") {
			fileInfo.Type = "repository"
		} else if strings.HasSuffix(fileInfo.Name, "validator") {
			fileInfo.Type = "validator"
		} else if strings.Contains(path, "types") {
			fileInfo.Type = "types"
		} else {
			fileInfo.Type = "internal"
		}
	} else {
		fileInfo.Type = "other"
	}
}

// findGoFiles finds all Go files in the project
func (dd *DuplicateDetector) findGoFiles() ([]*FileInfo, error) {
	var files []*FileInfo

	err := filepath.Walk(dd.projectRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors, don't stop the walk
		}

		// Skip directories and non-Go files
		if info.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}

		// Skip test files
		if strings.HasSuffix(path, "_test.go") {
			return nil
		}

		// Skip vendor and .git directories
		if strings.Contains(path, "/vendor/") || strings.Contains(path, "/.git/") {
			return nil
		}

		relPath, err := filepath.Rel(dd.projectRoot, path)
		if err != nil {
			return nil
		}

		fileInfo := &FileInfo{
			Path:         path,
			RelativePath: relPath,
			Name:         strings.TrimSuffix(filepath.Base(path), ".go"),
			Extension:    ".go",
		}

		dd.classifyGoFile(fileInfo)
		files = append(files, fileInfo)

		return nil
	})

	return files, err
}

// checkForDuplicates checks for various types of duplicates
func (dd *DuplicateDetector) checkForDuplicates(targetFile *FileInfo, existingFiles []*FileInfo, result *Result) {
	switch targetFile.Type {
	case "command":
		dd.checkCommandDuplicates(targetFile, existingFiles, result)
	case "manager", "service", "handler", "repository", "validator":
		dd.checkServiceDuplicates(targetFile, existingFiles, result)
	case "internal":
		dd.checkInternalDuplicates(targetFile, existingFiles, result)
	}
}

// checkCommandDuplicates checks for duplicate command files
func (dd *DuplicateDetector) checkCommandDuplicates(targetFile *FileInfo, existingFiles []*FileInfo, result *Result) {
	for _, file := range existingFiles {
		if file.Path == targetFile.Path {
			continue
		}

		// Check for exact command name match
		if file.Type == "command" && strings.ToLower(file.Name) == strings.ToLower(targetFile.Name) {
			result.Errors = append(result.Errors,
				fmt.Sprintf("Duplicate command detected! Command '%s' already exists at: %s",
					targetFile.Name, file.RelativePath))
			result.Success = false
		}
	}
}

// checkServiceDuplicates checks for duplicate service-type files
func (dd *DuplicateDetector) checkServiceDuplicates(targetFile *FileInfo, existingFiles []*FileInfo, result *Result) {
	targetBaseName := strings.ToLower(targetFile.Name)

	// Remove common suffixes for comparison
	suffixes := []string{"manager", "service", "handler", "repository", "validator"}
	for _, suffix := range suffixes {
		if strings.HasSuffix(targetBaseName, suffix) {
			targetBaseName = strings.TrimSuffix(targetBaseName, suffix)
			break
		}
	}

	for _, file := range existingFiles {
		if file.Path == targetFile.Path {
			continue
		}

		fileBaseName := strings.ToLower(file.Name)
		for _, suffix := range suffixes {
			if strings.HasSuffix(fileBaseName, suffix) {
				fileBaseName = strings.TrimSuffix(fileBaseName, suffix)
				break
			}
		}

		// Check for same base name with different suffix
		if targetBaseName == fileBaseName && targetBaseName != "" {
			if targetFile.Type == file.Type {
				result.Errors = append(result.Errors,
					fmt.Sprintf("Duplicate %s detected! '%s' already exists at: %s",
						targetFile.Type, targetFile.Name, file.RelativePath))
				result.Success = false
			} else {
				result.Warnings = append(result.Warnings,
					fmt.Sprintf("Similar file detected: '%s' (%s) at %s",
						file.Name, file.Type, file.RelativePath))
			}
		}
	}
}

// checkInternalDuplicates checks for duplicate internal files
func (dd *DuplicateDetector) checkInternalDuplicates(targetFile *FileInfo, existingFiles []*FileInfo, result *Result) {
	// Get the package directory
	targetDir := filepath.Dir(targetFile.RelativePath)

	for _, file := range existingFiles {
		if file.Path == targetFile.Path {
			continue
		}

		fileDir := filepath.Dir(file.RelativePath)

		// Check for same name in same package
		if targetDir == fileDir && strings.ToLower(targetFile.Name) == strings.ToLower(file.Name) {
			result.Errors = append(result.Errors,
				fmt.Sprintf("Duplicate file in same package! '%s' already exists at: %s",
					targetFile.Name, file.RelativePath))
			result.Success = false
		}
	}
}

// IsRelevantForDetection checks if a file is relevant for duplicate detection
func (dd *DuplicateDetector) IsRelevantForDetection(filePath string) bool {
	// Only Go files
	if !strings.HasSuffix(filePath, ".go") {
		return false
	}

	// Skip test files
	if strings.HasSuffix(filePath, "_test.go") {
		return false
	}

	// Skip vendor and generated files
	if strings.Contains(filePath, "/vendor/") ||
		strings.Contains(filePath, "/.git/") ||
		strings.Contains(filePath, "/node_modules/") {
		return false
	}

	// Only check cmd/ and internal/ directories
	relPath, err := filepath.Rel(dd.projectRoot, filePath)
	if err != nil {
		return false
	}

	return strings.HasPrefix(relPath, "cmd/") || strings.HasPrefix(relPath, "internal/")
}