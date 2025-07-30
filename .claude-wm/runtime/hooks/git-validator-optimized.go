package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"claude-hooks-orchestrator/patterns"
)

// ToolInput represents the input from Claude Code
type ToolInput struct {
	ToolName  string                 `json:"tool_name"`
	ToolInput map[string]interface{} `json:"tool_input"`
}

// ValidationResult represents the output of validation
type ValidationResult struct {
	Success  bool     `json:"success"`
	Errors   []string `json:"errors"`
	Warnings []string `json:"warnings"`
	Duration int64    `json:"duration_ms"`
}

// OptimizedGitValidator uses subprocess calls for better performance comparison
type OptimizedGitValidator struct {
	repoRoot     string
	currentDir   string
	errors       []string
	warnings     []string
	startTime    time.Time
}

// NewOptimizedGitValidator creates a new optimized Git validator instance
func NewOptimizedGitValidator() (*OptimizedGitValidator, error) {
	gv := &OptimizedGitValidator{
		errors:    make([]string, 0),
		warnings:  make([]string, 0),
		startTime: time.Now(),
	}

	// Get current working directory
	var err error
	gv.currentDir, err = os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current directory: %v", err)
	}

	// Check if we're in a git repository using subprocess (like Python version)
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("not in a git repository: %v", err)
	}
	
	gv.repoRoot = strings.TrimSpace(string(output))
	return gv, nil
}

// GetStagedFiles returns list of staged files using subprocess
func (gv *OptimizedGitValidator) GetStagedFiles() ([]string, error) {
	cmd := exec.Command("git", "diff", "--cached", "--name-only")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	files := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(files) == 1 && files[0] == "" {
		return []string{}, nil
	}
	
	return files, nil
}

// ValidateRepositoryContext validates git repository context
func (gv *OptimizedGitValidator) ValidateRepositoryContext() bool {
	// Check if we're at repository root
	if gv.repoRoot != gv.currentDir {
		gv.warnings = append(gv.warnings, 
			fmt.Sprintf("Not at repository root. Root: %s, Current: %s", gv.repoRoot, gv.currentDir))
	}

	// Quick repository health check
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	if err := cmd.Run(); err != nil {
		gv.errors = append(gv.errors, fmt.Sprintf("Repository health check failed: %v", err))
		return false
	}

	return true
}

// ValidateStagedFiles validates staged files for security and size
func (gv *OptimizedGitValidator) ValidateStagedFiles() bool {
	stagedFiles, err := gv.GetStagedFiles()
	if err != nil {
		gv.errors = append(gv.errors, fmt.Sprintf("Failed to get staged files: %v", err))
		return false
	}

	if len(stagedFiles) == 0 {
		return true
	}

	// Check for forbidden files (simplified patterns for performance)
	forbiddenPatterns := []string{
		`^\.git/`, `^\.claude/`, `\.log$`, `node_modules/`, `^\.env$`, `\.DS_Store$`,
	}

	var forbiddenFiles []string
	for _, filePath := range stagedFiles {
		for _, pattern := range forbiddenPatterns {
			if matched, _ := regexp.MatchString(pattern, filePath); matched {
				forbiddenFiles = append(forbiddenFiles, filePath)
				break
			}
		}
	}

	if len(forbiddenFiles) > 0 {
		gv.errors = append(gv.errors, "Forbidden files detected in staging:")
		for _, file := range forbiddenFiles {
			gv.errors = append(gv.errors, fmt.Sprintf("  - %s", file))
		}
		gv.errors = append(gv.errors, "Use 'git reset HEAD <file>' to unstage")
		return false
	}

	// Basic file size check (only for existing files)
	var largeFiles []string
	for _, filePath := range stagedFiles {
		fullPath := filepath.Join(gv.repoRoot, filePath)
		if info, err := os.Stat(fullPath); err == nil {
			if info.Size() > 10*1024*1024 { // 10MB
				largeFiles = append(largeFiles, filePath)
			}
		}
	}

	if len(largeFiles) > 0 {
		gv.warnings = append(gv.warnings, fmt.Sprintf("Large files detected (>10MB): %s", strings.Join(largeFiles, ", ")))
		gv.warnings = append(gv.warnings, "Consider Git LFS for large files")
	}

	// Check number of staged files
	if len(stagedFiles) > 50 {
		gv.warnings = append(gv.warnings, 
			fmt.Sprintf("Many staged files (%d). Consider more atomic commits", len(stagedFiles)))
	}

	return true
}

// ValidateGitignore validates .gitignore file existence and content
func (gv *OptimizedGitValidator) ValidateGitignore() bool {
	gitignorePath := filepath.Join(gv.repoRoot, ".gitignore")
	
	if _, err := os.Stat(gitignorePath); os.IsNotExist(err) {
		gv.errors = append(gv.errors, "No .gitignore file found")
		gv.errors = append(gv.errors, "Create .gitignore with basic security patterns")
		return false
	}

	// Basic .gitignore validation (simplified for performance)
	content, err := ioutil.ReadFile(gitignorePath)
	if err != nil {
		gv.errors = append(gv.errors, fmt.Sprintf("Failed to read .gitignore: %v", err))
		return false
	}

	gitignoreContent := string(content)
	criticalPatterns := []string{".env", "node_modules/", ".DS_Store"}
	
	for _, pattern := range criticalPatterns {
		if !strings.Contains(gitignoreContent, pattern) {
			gv.warnings = append(gv.warnings, fmt.Sprintf("Missing critical .gitignore pattern: %s", pattern))
		}
	}

	return true
}

// ValidateCommitMessage validates commit message format (simplified)
func (gv *OptimizedGitValidator) ValidateCommitMessage(message string) bool {
	if message == "" {
		return true // Skip if no message provided
	}

	// Quick validation checks
	if strings.Contains(message, "Co-Authored-By") {
		gv.errors = append(gv.errors, "Co-authored commits are not allowed per project rules")
	}

	if strings.Contains(message, "🤖 Generated with [Claude Code]") {
		gv.errors = append(gv.errors, "Remove Claude signature from commit messages")
	}

	// Quick length check
	firstLine := strings.Split(message, "\n")[0]
	if len(firstLine) < 10 {
		gv.errors = append(gv.errors, "Commit message too short (minimum 10 characters)")
	} else if len(firstLine) > 72 {
		gv.warnings = append(gv.warnings, "First line should be ≤72 characters")
	}

	return true
}

// AnalyzeREADMEChanges provides simple README analysis
func (gv *OptimizedGitValidator) AnalyzeREADMEChanges(files []string) {
	if len(files) == 0 {
		return
	}

	// Check if README exists
	readmePath := filepath.Join(gv.repoRoot, "README.md")
	if _, err := os.Stat(readmePath); os.IsNotExist(err) {
		gv.warnings = append(gv.warnings, "No README.md found - consider creating one")
		return
	}

	// Check if README is being updated
	readmeUpdated := false
	for _, file := range files {
		if strings.Contains(strings.ToLower(file), "readme") {
			readmeUpdated = true
			break
		}
	}

	// Simple heuristic: if significant files changed and README not updated, warn
	significantFiles := 0
	for _, file := range files {
		if strings.Contains(file, "package.json") || 
		   strings.Contains(file, "/api/") || 
		   strings.HasSuffix(file, ".sh") ||
		   strings.Contains(file, "config") {
			significantFiles++
		}
	}

	if significantFiles > 0 && !readmeUpdated {
		gv.warnings = append(gv.warnings, "Consider updating README for significant changes")
	}
}

// ExtractCommitMessageFromCommand extracts commit message from git commit command
func (gv *OptimizedGitValidator) ExtractCommitMessageFromCommand(command string) string {
	// Simple regex patterns for common commit message formats
	commitPatterns := []string{
		`-m\s+"([^"]+)"`,   // -m "message"
		`-m\s+'([^']+)'`,   // -m 'message'
		`-m\s+([^\s]+)`,    // -m message
	}

	for _, pattern := range commitPatterns {
		re := patterns.MustCompilePattern(pattern)
		matches := re.FindStringSubmatch(command)
		if len(matches) > 1 {
			return matches[1]
		}
	}

	return ""
}

// RunFullValidation runs comprehensive git validation
func (gv *OptimizedGitValidator) RunFullValidation(toolName string, toolInput map[string]interface{}) bool {
	// Always validate repository context
	if !gv.ValidateRepositoryContext() {
		return false
	}

	// Validate based on tool and command
	if toolName == "Bash" {
		if command, ok := toolInput["command"].(string); ok {
			// Git commit validation
			if strings.Contains(command, "git commit") {
				// Skip amend with no-edit
				if strings.Contains(command, "--amend") && strings.Contains(command, "--no-edit") {
					return true
				}

				// Validate commit message
				commitMessage := gv.ExtractCommitMessageFromCommand(command)
				if commitMessage != "" {
					gv.ValidateCommitMessage(commitMessage)
				}

				// Validate staged files
				gv.ValidateStagedFiles()

				// Analyze README changes
				if stagedFiles, err := gv.GetStagedFiles(); err == nil {
					gv.AnalyzeREADMEChanges(stagedFiles)
				}
			} else if strings.Contains(command, "git add") {
				// Git add validation
				gv.ValidateStagedFiles()
				gv.ValidateGitignore()
			}
		}
	}

	// Return true if no errors (warnings are okay)
	return len(gv.errors) == 0
}

// PrintResults prints validation results
func (gv *OptimizedGitValidator) PrintResults() {
	if len(gv.errors) > 0 {
		fmt.Fprintf(os.Stderr, "\n🚨 Git Validation Errors:\n")
		for _, error := range gv.errors {
			fmt.Fprintf(os.Stderr, "❌ %s\n", error)
		}
	}

	if len(gv.warnings) > 0 {
		fmt.Fprintf(os.Stderr, "\n⚠️  Git Validation Warnings:\n")
		for _, warning := range gv.warnings {
			fmt.Fprintf(os.Stderr, "⚠️  %s\n", warning)
		}
	}

	if len(gv.errors) > 0 {
		fmt.Fprintf(os.Stderr, "\n❌ Git operation blocked due to validation errors\n")
		fmt.Fprintf(os.Stderr, "Please fix the errors above and try again.\n")
	} else if len(gv.warnings) > 0 {
		fmt.Fprintf(os.Stderr, "\nProceeding with warnings...\n")
	}
}

func main() {
	// Read input from stdin
	inputBytes, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
		os.Exit(1)
	}

	var input ToolInput
	if err := json.Unmarshal(inputBytes, &input); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing JSON input: %v\n", err)
		os.Exit(1)
	}

	// Initialize validator
	validator, err := NewOptimizedGitValidator()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing git validator: %v\n", err)
		os.Exit(1)
	}

	// Run validation
	success := validator.RunFullValidation(input.ToolName, input.ToolInput)

	// Print results
	validator.PrintResults()

	// Create result
	result := ValidationResult{
		Success:  success,
		Errors:   validator.errors,
		Warnings: validator.warnings,
		Duration: time.Since(validator.startTime).Milliseconds(),
	}

	// Output result as JSON for potential consumption by orchestrator
	resultJSON, _ := json.Marshal(result)
	fmt.Printf("VALIDATION_RESULT: %s\n", resultJSON)

	// Exit with appropriate code
	if success {
		os.Exit(0)
	} else {
		os.Exit(2)
	}
}