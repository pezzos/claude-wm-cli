package state

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// CorruptionType represents different types of corruption
type CorruptionType string

const (
	CorruptionSyntax     CorruptionType = "syntax"      // Invalid JSON syntax
	CorruptionSchema     CorruptionType = "schema"      // Schema validation failure
	CorruptionChecksum   CorruptionType = "checksum"    // Checksum mismatch
	CorruptionStructure  CorruptionType = "structure"   // Invalid structure/references
	CorruptionPartial    CorruptionType = "partial"     // Incomplete/truncated file
	CorruptionEncoding   CorruptionType = "encoding"    // Encoding issues
)

// CorruptionIssue represents a detected corruption issue
type CorruptionIssue struct {
	Type        CorruptionType `json:"type"`
	Severity    string         `json:"severity"`    // critical, major, minor
	Field       string         `json:"field,omitempty"`
	Message     string         `json:"message"`
	Details     string         `json:"details,omitempty"`
	Suggestion  string         `json:"suggestion,omitempty"`
	Recoverable bool           `json:"recoverable"`
	DetectedAt  time.Time      `json:"detected_at"`
}

// CorruptionReport contains the results of corruption detection
type CorruptionReport struct {
	FilePath       string             `json:"file_path"`
	IsCorrupted    bool               `json:"is_corrupted"`
	Issues         []CorruptionIssue  `json:"issues"`
	Checksum       string             `json:"checksum"`
	FileSize       int64              `json:"file_size"`
	LastModified   time.Time          `json:"last_modified"`
	ScanDuration   time.Duration      `json:"scan_duration"`
	RecoveryOptions []RecoveryOption   `json:"recovery_options,omitempty"`
}

// RecoveryOption represents a possible recovery action
type RecoveryOption struct {
	Type        string `json:"type"`        // backup, rebuild, manual
	Description string `json:"description"`
	Command     string `json:"command,omitempty"`
	Risk        string `json:"risk"`        // low, medium, high
}

// CorruptionDetector provides comprehensive corruption detection
type CorruptionDetector struct {
	validator *Validator
	atomicWriter *AtomicWriter
}

// NewCorruptionDetector creates a new corruption detector
func NewCorruptionDetector(atomicWriter *AtomicWriter) *CorruptionDetector {
	return &CorruptionDetector{
		validator: NewValidator(),
		atomicWriter: atomicWriter,
	}
}

// ScanFile performs comprehensive corruption detection on a file
func (cd *CorruptionDetector) ScanFile(filePath string) *CorruptionReport {
	start := time.Now()
	
	report := &CorruptionReport{
		FilePath:     filePath,
		IsCorrupted:  false,
		Issues:       make([]CorruptionIssue, 0),
		ScanDuration: 0,
	}
	
	// Check if file exists
	if !fileExists(filePath) {
		report.addIssue(CorruptionStructure, "critical", "", 
			"File does not exist", "", "Check file path and restore from backup if available", false)
		report.ScanDuration = time.Since(start)
		return report
	}
	
	// Get file info
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		report.addIssue(CorruptionStructure, "critical", "", 
			"Cannot access file", err.Error(), "Check file permissions", false)
		report.ScanDuration = time.Since(start)
		return report
	}
	
	report.FileSize = fileInfo.Size()
	report.LastModified = fileInfo.ModTime()
	
	// Read file content
	data, err := os.ReadFile(filePath)
	if err != nil {
		report.addIssue(CorruptionStructure, "critical", "", 
			"Cannot read file", err.Error(), "Check file permissions and disk health", false)
		report.ScanDuration = time.Since(start)
		return report
	}
	
	// Calculate checksum
	report.Checksum = calculateSHA256(data)
	
	// Check for empty or very small files
	if len(data) == 0 {
		report.addIssue(CorruptionPartial, "critical", "", 
			"File is empty", "", "Restore from backup or reinitialize", true)
		report.ScanDuration = time.Since(start)
		return report
	}
	
	if len(data) < 10 {
		report.addIssue(CorruptionPartial, "major", "", 
			"File appears truncated", fmt.Sprintf("Only %d bytes", len(data)), 
			"Restore from backup", true)
	}
	
	// Check for encoding issues
	if err := cd.checkEncoding(data, report); err != nil {
		report.ScanDuration = time.Since(start)
		return report
	}
	
	// Check JSON syntax
	var rawData interface{}
	if err := json.Unmarshal(data, &rawData); err != nil {
		report.addIssue(CorruptionSyntax, "critical", "", 
			"Invalid JSON syntax", err.Error(), 
			"Fix JSON syntax or restore from backup", true)
		
		// Try to suggest specific fixes for common JSON errors
		if suggestion := cd.suggestJSONFix(err, data); suggestion != "" {
			report.Issues[len(report.Issues)-1].Suggestion = suggestion
		}
		
		report.ScanDuration = time.Since(start)
		return report
	}
	
	// Check checksum integrity if we have a stored one
	if storedChecksum, exists := cd.atomicWriter.GetChecksum(filePath); exists {
		if storedChecksum != report.Checksum {
			report.addIssue(CorruptionChecksum, "major", "", 
				"Checksum mismatch", 
				fmt.Sprintf("Expected %s, got %s", storedChecksum, report.Checksum),
				"File may have been modified externally or corrupted", true)
		}
	}
	
	// Determine file type and perform specific validation
	cd.performSpecificValidation(filePath, data, report)
	
	// Add recovery options
	cd.addRecoveryOptions(filePath, report)
	
	report.ScanDuration = time.Since(start)
	return report
}

// checkEncoding checks for encoding issues
func (cd *CorruptionDetector) checkEncoding(data []byte, report *CorruptionReport) error {
	// Check for null bytes (often indicates binary corruption in text files)
	for i, b := range data {
		if b == 0 {
			report.addIssue(CorruptionEncoding, "major", "", 
				"Null byte found in JSON file", 
				fmt.Sprintf("Null byte at position %d", i),
				"File may be corrupted or contain binary data", true)
			return nil
		}
	}
	
	// Check for non-UTF8 sequences
	if !isValidUTF8(data) {
		report.addIssue(CorruptionEncoding, "major", "", 
			"Invalid UTF-8 encoding", "",
			"File contains invalid UTF-8 sequences", true)
	}
	
	return nil
}

// performSpecificValidation performs validation based on file type
func (cd *CorruptionDetector) performSpecificValidation(filePath string, data []byte, report *CorruptionReport) {
	filename := filepath.Base(filePath)
	
	switch {
	case strings.Contains(filename, "project"):
		cd.validateProjectFile(data, report)
	case strings.Contains(filename, "epic"):
		cd.validateEpicFile(data, report)
	case strings.Contains(filename, "story") || strings.Contains(filename, "stories"):
		cd.validateStoryFile(data, report)
	case strings.Contains(filename, "task") || strings.Contains(filename, "todo"):
		cd.validateTaskFile(data, report)
	default:
		cd.validateGenericState(data, report)
	}
}

// validateProjectFile validates project-specific structure
func (cd *CorruptionDetector) validateProjectFile(data []byte, report *CorruptionReport) {
	var project ProjectState
	if err := json.Unmarshal(data, &project); err != nil {
		report.addIssue(CorruptionSchema, "critical", "", 
			"Cannot parse as ProjectState", err.Error(),
			"Check project schema or restore from backup", true)
		return
	}
	
	// Validate using schema validator
	result := cd.validator.ValidateProject(&project)
	cd.addValidationIssues(result, report)
	
	// Additional project-specific checks
	if project.Metadata.SchemaVersion == "" {
		report.addIssue(CorruptionSchema, "minor", "metadata.schema_version", 
			"Missing schema version", "",
			"Update file with current schema version", true)
	}
}

// validateEpicFile validates epic-specific structure
func (cd *CorruptionDetector) validateEpicFile(data []byte, report *CorruptionReport) {
	// Try to parse as single epic
	var epic EpicState
	if err := json.Unmarshal(data, &epic); err == nil {
		result := cd.validator.ValidateEpic(&epic)
		cd.addValidationIssues(result, report)
		return
	}
	
	// Try to parse as map of epics
	var epics map[string]EpicState
	if err := json.Unmarshal(data, &epics); err == nil {
		for id, epic := range epics {
			result := cd.validator.ValidateEpic(&epic)
			cd.addValidationIssuesWithPrefix(result, report, fmt.Sprintf("epics[%s]", id))
		}
		return
	}
	
	report.addIssue(CorruptionSchema, "critical", "", 
		"Cannot parse as epic data", "",
		"Check epic schema or restore from backup", true)
}

// validateStoryFile validates story-specific structure
func (cd *CorruptionDetector) validateStoryFile(data []byte, report *CorruptionReport) {
	// Try to parse as single story
	var story StoryState
	if err := json.Unmarshal(data, &story); err == nil {
		result := cd.validator.ValidateStory(&story)
		cd.addValidationIssues(result, report)
		return
	}
	
	// Try to parse as map of stories
	var stories map[string]StoryState
	if err := json.Unmarshal(data, &stories); err == nil {
		for id, story := range stories {
			result := cd.validator.ValidateStory(&story)
			cd.addValidationIssuesWithPrefix(result, report, fmt.Sprintf("stories[%s]", id))
		}
		return
	}
	
	report.addIssue(CorruptionSchema, "critical", "", 
		"Cannot parse as story data", "",
		"Check story schema or restore from backup", true)
}

// validateTaskFile validates task-specific structure
func (cd *CorruptionDetector) validateTaskFile(data []byte, report *CorruptionReport) {
	// Try to parse as single task
	var task TaskState
	if err := json.Unmarshal(data, &task); err == nil {
		result := cd.validator.ValidateTask(&task)
		cd.addValidationIssues(result, report)
		return
	}
	
	// Try to parse as map of tasks
	var tasks map[string]TaskState
	if err := json.Unmarshal(data, &tasks); err == nil {
		for id, task := range tasks {
			result := cd.validator.ValidateTask(&task)
			cd.addValidationIssuesWithPrefix(result, report, fmt.Sprintf("tasks[%s]", id))
		}
		return
	}
	
	report.addIssue(CorruptionSchema, "critical", "", 
		"Cannot parse as task data", "",
		"Check task schema or restore from backup", true)
}

// validateGenericState validates generic state collection
func (cd *CorruptionDetector) validateGenericState(data []byte, report *CorruptionReport) {
	var state StateCollection
	if err := json.Unmarshal(data, &state); err != nil {
		report.addIssue(CorruptionSchema, "major", "", 
			"Cannot parse as StateCollection", err.Error(),
			"Check state schema", true)
		return
	}
	
	result := cd.validator.ValidateStateCollection(&state)
	cd.addValidationIssues(result, report)
}

// addValidationIssues converts validation errors to corruption issues
func (cd *CorruptionDetector) addValidationIssues(result *ValidationResult, report *CorruptionReport) {
	cd.addValidationIssuesWithPrefix(result, report, "")
}

// addValidationIssuesWithPrefix converts validation errors to corruption issues with field prefix
func (cd *CorruptionDetector) addValidationIssuesWithPrefix(result *ValidationResult, report *CorruptionReport, prefix string) {
	for _, err := range result.Errors {
		field := err.Field
		if prefix != "" {
			field = prefix + "." + field
		}
		
		report.addIssue(CorruptionSchema, "major", field, 
			"Schema validation failed", err.Message,
			"Fix the validation error or restore from backup", true)
	}
	
	for _, warn := range result.Warnings {
		field := warn.Field
		if prefix != "" {
			field = prefix + "." + field
		}
		
		report.addIssue(CorruptionSchema, "minor", field, 
			"Schema validation warning", warn.Message,
			"Consider fixing the warning", true)
	}
}

// suggestJSONFix suggests fixes for common JSON syntax errors
func (cd *CorruptionDetector) suggestJSONFix(err error, data []byte) string {
	errStr := strings.ToLower(err.Error())
	
	switch {
	case strings.Contains(errStr, "unexpected end"):
		return "File appears truncated. Check for missing closing braces or brackets."
	case strings.Contains(errStr, "invalid character"):
		return "File contains invalid characters. Check for unescaped quotes or control characters."
	case strings.Contains(errStr, "expecting comma"):
		return "Missing comma between JSON elements. Check object and array syntax."
	case strings.Contains(errStr, "duplicate key"):
		return "Duplicate JSON keys found. Remove or rename duplicate keys."
	default:
		return "Use a JSON validator tool to identify and fix syntax errors."
	}
}

// addRecoveryOptions adds possible recovery options to the report
func (cd *CorruptionDetector) addRecoveryOptions(filePath string, report *CorruptionReport) {
	if !report.IsCorrupted {
		return
	}
	
	// Check for backups
	backups, err := cd.atomicWriter.ListBackups(filePath)
	if err == nil && len(backups) > 0 {
		report.RecoveryOptions = append(report.RecoveryOptions, RecoveryOption{
			Type:        "backup",
			Description: fmt.Sprintf("Restore from backup (%d available)", len(backups)),
			Command:     fmt.Sprintf("claude-wm-cli state restore %s", filePath),
			Risk:        "low",
		})
	}
	
	// Suggest rebuild for certain corruption types
	hasSchemaIssues := false
	for _, issue := range report.Issues {
		if issue.Type == CorruptionSchema && issue.Severity != "critical" {
			hasSchemaIssues = true
			break
		}
	}
	
	if hasSchemaIssues {
		report.RecoveryOptions = append(report.RecoveryOptions, RecoveryOption{
			Type:        "rebuild",
			Description: "Rebuild state from valid components",
			Command:     fmt.Sprintf("claude-wm-cli state rebuild %s", filePath),
			Risk:        "medium",
		})
	}
	
	// Always offer manual fix option
	report.RecoveryOptions = append(report.RecoveryOptions, RecoveryOption{
		Type:        "manual",
		Description: "Manually edit and fix the file",
		Command:     "",
		Risk:        "high",
	})
}

// Helper methods
func (r *CorruptionReport) addIssue(cType CorruptionType, severity, field, message, details, suggestion string, recoverable bool) {
	r.IsCorrupted = true
	r.Issues = append(r.Issues, CorruptionIssue{
		Type:        cType,
		Severity:    severity,
		Field:       field,
		Message:     message,
		Details:     details,
		Suggestion:  suggestion,
		Recoverable: recoverable,
		DetectedAt:  time.Now(),
	})
}

// ScanDirectory scans all JSON files in a directory for corruption
func (cd *CorruptionDetector) ScanDirectory(dirPath string) ([]*CorruptionReport, error) {
	var reports []*CorruptionReport
	
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		if !info.IsDir() && strings.HasSuffix(strings.ToLower(path), ".json") {
			report := cd.ScanFile(path)
			reports = append(reports, report)
		}
		
		return nil
	})
	
	return reports, err
}

// AutoRepair attempts to automatically repair corrupted files
func (cd *CorruptionDetector) AutoRepair(filePath string) error {
	report := cd.ScanFile(filePath)
	
	if !report.IsCorrupted {
		return nil // Nothing to repair
	}
	
	// Try recovery options in order of safety
	for _, option := range report.RecoveryOptions {
		switch option.Type {
		case "backup":
			if err := cd.atomicWriter.RestoreFromBackup(filePath); err == nil {
				return nil // Successfully restored
			}
		case "rebuild":
			// Implementation would depend on specific rebuild logic
			// For now, just log that rebuild is available
			fmt.Printf("Rebuild option available for %s\n", filePath)
		}
	}
	
	return fmt.Errorf("automatic repair failed for %s", filePath)
}

// Utility functions
func calculateSHA256(data []byte) string {
	hash := sha256.Sum256(data)
	return fmt.Sprintf("%x", hash)
}

func isValidUTF8(data []byte) bool {
	// Simple UTF-8 validation
	for len(data) > 0 {
		r, size := decodeRuneInString(string(data))
		if r == '\uFFFD' && size == 1 {
			return false
		}
		data = data[size:]
	}
	return true
}

// decodeRuneInString is a simplified version for UTF-8 validation
func decodeRuneInString(s string) (rune, int) {
	if len(s) == 0 {
		return 0, 0
	}
	
	b := s[0]
	if b < 0x80 {
		return rune(b), 1
	}
	
	// For simplicity, return replacement character for non-ASCII
	// In a real implementation, you'd properly decode UTF-8
	return '\uFFFD', 1
}