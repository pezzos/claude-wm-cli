package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"claude-hooks-orchestrator/patterns"
)

type QualityRules struct {
	StyleConsistency struct {
		Enabled     bool     `json:"enabled"`
		Severity    string   `json:"severity"`
		FileTypes   []string `json:"file_types"`
		Indentation struct {
			Type string `json:"type"`
			Size int    `json:"size"`
		} `json:"indentation"`
		LineLength struct {
			Max int `json:"max"`
		} `json:"line_length"`
		Naming struct {
			CamelCase  []string `json:"camelCase"`
			SnakeCase  []string `json:"snake_case"`
			KebabCase  []string `json:"kebab_case"`
			PascalCase []string `json:"PascalCase"`
		} `json:"naming"`
		Patterns []struct {
			Name        string `json:"name"`
			Pattern     string `json:"pattern"`
			Message     string `json:"message"`
			Severity    string `json:"severity"`
			Replacement string `json:"replacement,omitempty"`
		} `json:"patterns"`
	} `json:"style_consistency"`
	MockCodeDetection struct {
		Enabled   bool     `json:"enabled"`
		Severity  string   `json:"severity"`
		FileTypes []string `json:"file_types"`
		Patterns  []struct {
			Name     string `json:"name"`
			Pattern  string `json:"pattern"`
			Message  string `json:"message"`
			Severity string `json:"severity"`
		} `json:"patterns"`
	} `json:"mock_code_detection"`
	TimestampValidation struct {
		Enabled   bool     `json:"enabled"`
		Severity  string   `json:"severity"`
		FileTypes []string `json:"file_types"`
		Patterns  []struct {
			Name     string `json:"name"`
			Pattern  string `json:"pattern"`
			Message  string `json:"message"`
			Severity string `json:"severity"`
		} `json:"patterns"`
	} `json:"timestamp_validation"`
}

type Issue struct {
	File     string `json:"file"`
	Line     int    `json:"line"`
	Column   int    `json:"column"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
	Rule     string `json:"rule"`
	Type     string `json:"type"`
}

type ValidationResult struct {
	Success bool    `json:"success"`
	Issues  []Issue `json:"issues"`
	Summary struct {
		Total    int `json:"total"`
		Critical int `json:"critical"`
		High     int `json:"high"`
		Medium   int `json:"medium"`
		Low      int `json:"low"`
	} `json:"summary"`
	Timestamp time.Time `json:"timestamp"`
}

var compiledPatterns = make(map[string]*regexp.Regexp)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <file_path> [--config=path]\n", os.Args[0])
		os.Exit(1)
	}

	filePath := os.Args[1]
	configPath := "/Users/a.pezzotta/.claude/hooks/rules/quality-rules.json"

	for _, arg := range os.Args[2:] {
		if strings.HasPrefix(arg, "--config=") {
			configPath = strings.TrimPrefix(arg, "--config=")
		}
	}

	result := validateFile(filePath, configPath)
	
	output, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling result: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(output))

	if !result.Success {
		os.Exit(1)
	}
}

func validateFile(filePath, configPath string) ValidationResult {
	result := ValidationResult{
		Success:   true,
		Issues:    []Issue{},
		Timestamp: time.Now(),
	}

	rules, err := loadQualityRules(configPath)
	if err != nil {
		result.Success = false
		result.Issues = append(result.Issues, Issue{
			File:     filePath,
			Severity: "critical",
			Message:  fmt.Sprintf("Failed to load quality rules: %v", err),
			Rule:     "config_error",
			Type:     "system",
		})
		return result
	}

	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		result.Success = false
		result.Issues = append(result.Issues, Issue{
			File:     filePath,
			Severity: "critical",
			Message:  fmt.Sprintf("Failed to read file: %v", err),
			Rule:     "file_error",
			Type:     "system",
		})
		return result
	}

	fileExt := strings.ToLower(filepath.Ext(filePath))
	contentStr := string(content)

	// Early exit: Skip if file is empty
	if len(contentStr) == 0 {
		return result
	}

	// Early exit: Skip if file is too large (> 1MB for performance)
	if len(contentStr) > 1024*1024 {
		return result
	}

	// Early exit: Skip if file is binary or generated
	if isBinaryOrGenerated(contentStr) {
		return result
	}

	// Style Consistency Validation
	if rules.StyleConsistency.Enabled && containsFileType(rules.StyleConsistency.FileTypes, fileExt) {
		// Early exit: Skip if no style-relevant content
		if !hasStyleRelevantContent(contentStr) {
			// Skip style validation
		} else {
			styleIssues := validateStyleConsistency(filePath, contentStr, rules.StyleConsistency)
			result.Issues = append(result.Issues, styleIssues...)
		}
	}

	// Mock Code Detection
	if rules.MockCodeDetection.Enabled && containsFileType(rules.MockCodeDetection.FileTypes, fileExt) {
		// Early exit: Skip if no mock-related patterns
		if !hasMockRelatedContent(contentStr) {
			// Skip mock detection
		} else {
			mockIssues := validateMockCode(filePath, contentStr, rules.MockCodeDetection)
			result.Issues = append(result.Issues, mockIssues...)
		}
	}

	// Timestamp Validation
	if rules.TimestampValidation.Enabled && containsFileType(rules.TimestampValidation.FileTypes, fileExt) {
		// Early exit: Skip if no timestamp patterns
		if !hasTimestampContent(contentStr) {
			// Skip timestamp validation
		} else {
			timestampIssues := validateTimestamps(filePath, contentStr, rules.TimestampValidation)
			result.Issues = append(result.Issues, timestampIssues...)
		}
	}

	// Calculate summary
	for _, issue := range result.Issues {
		result.Summary.Total++
		switch issue.Severity {
		case "critical":
			result.Summary.Critical++
			result.Success = false
		case "high":
			result.Summary.High++
		case "medium":
			result.Summary.Medium++
		case "low":
			result.Summary.Low++
		}
	}

	return result
}

func loadQualityRules(configPath string) (*QualityRules, error) {
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var rules QualityRules
	err = json.Unmarshal(data, &rules)
	if err != nil {
		return nil, err
	}

	return &rules, nil
}

func containsFileType(fileTypes []string, ext string) bool {
	for _, ft := range fileTypes {
		if ft == ext || ft == "*" {
			return true
		}
	}
	return false
}

func validateStyleConsistency(filePath, content string, rules struct {
	Enabled     bool     `json:"enabled"`
	Severity    string   `json:"severity"`
	FileTypes   []string `json:"file_types"`
	Indentation struct {
		Type string `json:"type"`
		Size int    `json:"size"`
	} `json:"indentation"`
	LineLength struct {
		Max int `json:"max"`
	} `json:"line_length"`
	Naming struct {
		CamelCase  []string `json:"camelCase"`
		SnakeCase  []string `json:"snake_case"`
		KebabCase  []string `json:"kebab_case"`
		PascalCase []string `json:"PascalCase"`
	} `json:"naming"`
	Patterns []struct {
		Name        string `json:"name"`
		Pattern     string `json:"pattern"`
		Message     string `json:"message"`
		Severity    string `json:"severity"`
		Replacement string `json:"replacement,omitempty"`
	} `json:"patterns"`
}) []Issue {
	var issues []Issue
	lines := strings.Split(content, "\n")

	// Check indentation
	if rules.Indentation.Type != "" {
		issues = append(issues, validateIndentation(filePath, lines, rules.Indentation, rules.Severity)...)
	}

	// Check line length
	if rules.LineLength.Max > 0 {
		issues = append(issues, validateLineLength(filePath, lines, rules.LineLength.Max, rules.Severity)...)
	}

	// Check custom patterns
	for _, pattern := range rules.Patterns {
		patternIssues := validatePattern(filePath, content, pattern.Pattern, pattern.Message, pattern.Severity, "style_consistency", pattern.Name)
		issues = append(issues, patternIssues...)
	}

	return issues
}

func validateMockCode(filePath, content string, rules struct {
	Enabled   bool     `json:"enabled"`
	Severity  string   `json:"severity"`
	FileTypes []string `json:"file_types"`
	Patterns  []struct {
		Name     string `json:"name"`
		Pattern  string `json:"pattern"`
		Message  string `json:"message"`
		Severity string `json:"severity"`
	} `json:"patterns"`
}) []Issue {
	var issues []Issue

	for _, pattern := range rules.Patterns {
		severity := pattern.Severity
		if severity == "" {
			severity = rules.Severity
		}
		patternIssues := validatePattern(filePath, content, pattern.Pattern, pattern.Message, severity, "mock_code_detection", pattern.Name)
		issues = append(issues, patternIssues...)
	}

	return issues
}

func validateTimestamps(filePath, content string, rules struct {
	Enabled   bool     `json:"enabled"`
	Severity  string   `json:"severity"`
	FileTypes []string `json:"file_types"`
	Patterns  []struct {
		Name     string `json:"name"`
		Pattern  string `json:"pattern"`
		Message  string `json:"message"`
		Severity string `json:"severity"`
	} `json:"patterns"`
}) []Issue {
	var issues []Issue

	for _, pattern := range rules.Patterns {
		severity := pattern.Severity
		if severity == "" {
			severity = rules.Severity
		}
		patternIssues := validatePattern(filePath, content, pattern.Pattern, pattern.Message, severity, "timestamp_validation", pattern.Name)
		issues = append(issues, patternIssues...)
	}

	return issues
}

func validateIndentation(filePath string, lines []string, indentConfig struct {
	Type string `json:"type"`
	Size int    `json:"size"`
}, severity string) []Issue {
	var issues []Issue

	for lineNum, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue // Skip empty lines
		}

		leadingWhitespace := line[:len(line)-len(strings.TrimLeft(line, " \t"))]
		
		if indentConfig.Type == "spaces" {
			if strings.Contains(leadingWhitespace, "\t") {
				issues = append(issues, Issue{
					File:     filePath,
					Line:     lineNum + 1,
					Column:   1,
					Severity: severity,
					Message:  fmt.Sprintf("Found tabs in indentation, expected %d spaces", indentConfig.Size),
					Rule:     "indentation_consistency",
					Type:     "style_consistency",
				})
			}
		} else if indentConfig.Type == "tabs" {
			if strings.Contains(leadingWhitespace, " ") {
				issues = append(issues, Issue{
					File:     filePath,
					Line:     lineNum + 1,
					Column:   1,
					Severity: severity,
					Message:  "Found spaces in indentation, expected tabs",
					Rule:     "indentation_consistency",
					Type:     "style_consistency",
				})
			}
		}
	}

	return issues
}

func validateLineLength(filePath string, lines []string, maxLength int, severity string) []Issue {
	var issues []Issue

	for lineNum, line := range lines {
		if len(line) > maxLength {
			issues = append(issues, Issue{
				File:     filePath,
				Line:     lineNum + 1,
				Column:   maxLength + 1,
				Severity: severity,
				Message:  fmt.Sprintf("Line too long (%d characters), maximum is %d", len(line), maxLength),
				Rule:     "line_length",
				Type:     "style_consistency",
			})
		}
	}

	return issues
}

func validatePattern(filePath, content, pattern, message, severity, ruleType, ruleName string) []Issue {
	var issues []Issue

	regex, err := getCompiledPattern(pattern)
	if err != nil {
		issues = append(issues, Issue{
			File:     filePath,
			Severity: "critical",
			Message:  fmt.Sprintf("Invalid regex pattern '%s': %v", pattern, err),
			Rule:     "pattern_error",
			Type:     ruleType,
		})
		return issues
	}

	lines := strings.Split(content, "\n")
	for lineNum, line := range lines {
		matches := regex.FindAllStringIndex(line, -1)
		for _, match := range matches {
			issues = append(issues, Issue{
				File:     filePath,
				Line:     lineNum + 1,
				Column:   match[0] + 1,
				Severity: severity,
				Message:  message,
				Rule:     ruleName,
				Type:     ruleType,
			})
		}
	}

	return issues
}

// Early exit helper functions
func isBinaryOrGenerated(content string) bool {
	// Check for null bytes (binary files)
	if strings.Contains(content, "\x00") {
		return true
	}

	// Check for generated file markers
	generatedMarkers := []string{
		"// Code generated by",
		"/* Generated by",
		"# Generated by",
		"# This file is auto-generated",
		"// This file is generated",
		"// DO NOT EDIT",
		"/* DO NOT EDIT */",
		"# DO NOT EDIT",
		"// Code generated",
		"// AUTO-GENERATED",
		"/**\n * AUTO-GENERATED",
		"// GENERATED CODE",
		"# GENERATED CODE",
		"// WARNING: This file is auto-generated",
		"// WARNING: Generated code",
		"// This is a generated file",
		"# This is a generated file",
		"// protoc-gen-go",
		"// swagger:model",
		"// swagger:operation",
		"// <auto-generated />",
		"// <auto-generated/>",
		"# <auto-generated />",
		"# <auto-generated/>",
	}

	// Check first few lines for generated markers
	lines := strings.Split(content, "\n")
	checkLines := 10
	if len(lines) < checkLines {
		checkLines = len(lines)
	}

	for i := 0; i < checkLines; i++ {
		line := strings.TrimSpace(lines[i])
		for _, marker := range generatedMarkers {
			if strings.Contains(line, marker) {
				return true
			}
		}
	}

	return false
}

func hasStyleRelevantContent(content string) bool {
	// Quick check for style-relevant patterns
	stylePatterns := []string{
		"function", "class", "const", "let", "var", "if", "for", "while", "switch",
		"def", "import", "from", "export", "interface", "type", "enum", "struct",
		"public", "private", "protected", "static", "async", "await", "return",
		"{", "}", "(", ")", "[", "]", ";", ":", "=", "=>", "->", "||", "&&",
		"//", "/*", "*/", "#", "<!--", "-->", "\"", "'", "`",
	}

	for _, pattern := range stylePatterns {
		if strings.Contains(content, pattern) {
			return true
		}
	}

	return false
}

func hasMockRelatedContent(content string) bool {
	// Check for mock-related patterns
	mockPatterns := []string{
		"mock", "Mock", "MOCK", "fake", "Fake", "FAKE", "stub", "Stub", "STUB",
		"test", "Test", "TEST", "TODO", "FIXME", "XXX", "HACK", "NOTE",
		"placeholder", "Placeholder", "PLACEHOLDER", "example", "Example", "EXAMPLE",
		"lorem", "Lorem", "LOREM", "ipsum", "Ipsum", "IPSUM", "dummy", "Dummy", "DUMMY",
		"sample", "Sample", "SAMPLE", "template", "Template", "TEMPLATE",
		"console.log", "print(", "printf(", "println(", "debug", "Debug", "DEBUG",
	}

	for _, pattern := range mockPatterns {
		if strings.Contains(content, pattern) {
			return true
		}
	}

	return false
}

func hasTimestampContent(content string) bool {
	// Check for timestamp patterns
	timestampPatterns := []string{
		"Date", "date", "TIME", "time", "timestamp", "Timestamp", "TIMESTAMP",
		"created", "Created", "CREATED", "updated", "Updated", "UPDATED",
		"modified", "Modified", "MODIFIED", "lastModified", "LastModified",
		"createdAt", "CreatedAt", "updatedAt", "UpdatedAt", "modifiedAt", "ModifiedAt",
		"2023", "2024", "2025", "2022", "2021", "2020", "2019", "2018", "2017",
		"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec",
		"January", "February", "March", "April", "May", "June", "July", "August", "September", "October", "November", "December",
		"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun",
		"Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday",
		":", "-", "/", "T", "Z", "UTC", "GMT", "EST", "PST", "PDT", "CET",
		"now()", "Now()", "NOW()", "today", "Today", "TODAY", "yesterday", "Yesterday", "YESTERDAY",
		"tomorrow", "Tomorrow", "TOMORROW", "current_time", "Current_Time", "CURRENT_TIME",
	}

	for _, pattern := range timestampPatterns {
		if strings.Contains(content, pattern) {
			return true
		}
	}

	return false
}

func getCompiledPattern(pattern string) (*regexp.Regexp, error) {
	if compiled, exists := compiledPatterns[pattern]; exists {
		return compiled, nil
	}

	compiled, err := patterns.CompilePattern(pattern)
	if err != nil {
		return nil, err
	}

	compiledPatterns[pattern] = compiled
	return compiled, nil
}