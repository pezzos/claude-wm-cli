package git

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// ValidationResult represents the output of git validation
type ValidationResult struct {
	Success  bool     `json:"success"`
	Errors   []string `json:"errors"`
	Warnings []string `json:"warnings"`
	Duration int64    `json:"duration_ms"`
}

// Validator provides Git validation functionality for claude-wm-cli
type Validator struct {
	repo       *git.Repository
	workTree   *git.Worktree
	repoRoot   string
	currentDir string
	errors     []string
	warnings   []string
	startTime  time.Time
}

// Forbidden files patterns specific to claude-wm-cli
var forbiddenPatterns = []string{
	`^\.git/`,           // Git internal files
	`^\.claude-wm/`,     // Claude WM internal files
	`\.log$`,            // Log files
	`^\.env$`,           // Environment files
	`\.DS_Store$`,       // macOS system files
	`.*\.backup$`,       // Backup files
	`.*\.bak$`,          // Backup files
	`.*\.tmp$`,          // Temporary files
	`.*~$`,              // Editor backup files
}

// Warning files patterns
var warningPatterns = []string{
	`config\.(json|yml|yaml)$`,
	`settings\.(json|yml|yaml)$`,
	`.*\.sql$`,
	`debug\.txt$`,
	`error\.txt$`,
}

// NewValidator creates a new Git validator instance
func NewValidator() (*Validator, error) {
	v := &Validator{
		errors:    make([]string, 0),
		warnings:  make([]string, 0),
		startTime: time.Now(),
	}

	// Get current working directory
	var err error
	v.currentDir, err = os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current directory: %v", err)
	}

	// Open git repository
	v.repo, err = git.PlainOpen(v.currentDir)
	if err != nil {
		// Try opening from parent directories
		dir := v.currentDir
		for dir != "/" {
			dir = filepath.Dir(dir)
			v.repo, err = git.PlainOpen(dir)
			if err == nil {
				v.repoRoot = dir
				break
			}
		}
		if err != nil {
			return nil, fmt.Errorf("not in a git repository: %v", err)
		}
	} else {
		v.repoRoot = v.currentDir
	}

	// Get worktree
	v.workTree, err = v.repo.Worktree()
	if err != nil {
		return nil, fmt.Errorf("failed to get worktree: %v", err)
	}

	return v, nil
}

// ValidateRepositoryContext validates git repository context and status
func (v *Validator) ValidateRepositoryContext() bool {
	// Check if we're at repository root
	if v.repoRoot != v.currentDir {
		v.warnings = append(v.warnings,
			fmt.Sprintf("Not at repository root. Root: %s, Current: %s", v.repoRoot, v.currentDir))
	}

	// Check repository health
	_, err := v.repo.Head()
	if err != nil {
		v.errors = append(v.errors, fmt.Sprintf("Repository head error: %v", err))
		return false
	}

	// Check git status is clean for sensitive operations
	status, err := v.workTree.Status()
	if err != nil {
		v.warnings = append(v.warnings, fmt.Sprintf("Could not get git status: %v", err))
		return true
	}

	if !status.IsClean() {
		modifiedFiles := 0
		for _, fileStatus := range status {
			if fileStatus.Worktree != git.Unmodified {
				modifiedFiles++
			}
		}
		if modifiedFiles > 0 {
			v.warnings = append(v.warnings,
				fmt.Sprintf("Working directory not clean: %d modified files", modifiedFiles))
		}
	}

	return true
}

// ValidateStagedFiles validates staged files for forbidden patterns and size
func (v *Validator) ValidateStagedFiles() bool {
	status, err := v.workTree.Status()
	if err != nil {
		v.errors = append(v.errors, fmt.Sprintf("Failed to get git status: %v", err))
		return false
	}

	var stagedFiles []string
	for file, fileStatus := range status {
		if fileStatus.Staging != git.Unmodified {
			stagedFiles = append(stagedFiles, file)
		}
	}

	if len(stagedFiles) == 0 {
		return true
	}

	// Check for forbidden files
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
		v.errors = append(v.errors, "Forbidden files detected in staging:")
		for _, file := range forbiddenFiles {
			v.errors = append(v.errors, fmt.Sprintf("  - %s", file))
		}
		v.errors = append(v.errors, "Use 'git reset HEAD <file>' to unstage")
		return false
	}

	// Check for warning files
	var warningFiles []string
	for _, filePath := range stagedFiles {
		for _, pattern := range warningPatterns {
			if matched, _ := regexp.MatchString(pattern, filePath); matched {
				warningFiles = append(warningFiles, filePath)
				break
			}
		}
	}

	if len(warningFiles) > 0 {
		v.warnings = append(v.warnings, "Warning files detected:")
		for _, file := range warningFiles {
			v.warnings = append(v.warnings, fmt.Sprintf("  - %s", file))
		}
	}

	// Check file sizes
	var largeFiles []struct {
		path string
		size int64
	}
	for _, filePath := range stagedFiles {
		fullPath := filepath.Join(v.repoRoot, filePath)
		if info, err := os.Stat(fullPath); err == nil {
			size := info.Size()
			if size > 10*1024*1024 { // 10MB
				largeFiles = append(largeFiles, struct {
					path string
					size int64
				}{filePath, size})
			}
		}
	}

	if len(largeFiles) > 0 {
		v.warnings = append(v.warnings, "Large files detected (>10MB):")
		for _, file := range largeFiles {
			v.warnings = append(v.warnings,
				fmt.Sprintf("  - %s (%.1fMB)", file.path, float64(file.size)/(1024*1024)))
		}
		v.warnings = append(v.warnings, "Consider Git LFS for large files")
	}

	// Check claude-wm-cli specific JSON files
	v.validateClaudeWMFiles(stagedFiles)

	return true
}

// validateClaudeWMFiles validates claude-wm-cli specific JSON files
func (v *Validator) validateClaudeWMFiles(files []string) {
	for _, file := range files {
		if strings.HasSuffix(file, ".json") && strings.Contains(file, "docs/") {
			if strings.Contains(file, "epics.json") ||
				strings.Contains(file, "stories.json") ||
				strings.Contains(file, "current-task.json") ||
				strings.Contains(file, "current-epic.json") ||
				strings.Contains(file, "current-story.json") {
				v.validateJSONStructure(file)
			}
		}
	}
}

// validateJSONStructure validates JSON file structure
func (v *Validator) validateJSONStructure(file string) {
	fullPath := filepath.Join(v.repoRoot, file)
	content, err := os.ReadFile(fullPath)
	if err != nil {
		v.warnings = append(v.warnings, fmt.Sprintf("Could not read %s for validation", file))
		return
	}

	var data interface{}
	if err := json.Unmarshal(content, &data); err != nil {
		v.errors = append(v.errors, fmt.Sprintf("Invalid JSON in %s: %v", file, err))
	}
}

// ValidateCommitMessage validates commit message format
func (v *Validator) ValidateCommitMessage(message string) bool {
	if message == "" {
		return true
	}

	// Block Co-authored commits and Claude signatures
	if strings.Contains(message, "Co-Authored-By") || strings.Contains(strings.ToLower(message), "co-authored-by") {
		v.errors = append(v.errors, "Co-authored commits are not allowed per project rules")
	}

	if strings.Contains(message, "🤖 Generated with [Claude Code]") || strings.Contains(message, "🤖 Generated with Claude") {
		v.errors = append(v.errors, "Remove Claude signature from commit messages")
	}

	// Extract main message
	lines := strings.Split(strings.TrimSpace(message), "\n")
	if len(lines) == 0 {
		v.errors = append(v.errors, "Empty commit message")
		return false
	}

	mainMessage := strings.TrimSpace(lines[0])

	// Check message length
	if len(mainMessage) > 72 {
		v.warnings = append(v.warnings,
			fmt.Sprintf("First line should be ≤72 characters (current: %d)", len(mainMessage)))
	} else if len(mainMessage) < 10 {
		v.errors = append(v.errors, "Commit message too short (minimum 10 characters)")
	}

	// Check conventional commit format
	conventionalPattern := `^(feat|fix|docs|style|refactor|test|chore|perf|ci|build|revert)(\(.+\))?: .+`
	if matched, _ := regexp.MatchString(conventionalPattern, mainMessage); matched {
		lowercasePattern := `^[a-z]+(\(.+\))?: [a-z]`
		if matched, _ := regexp.MatchString(lowercasePattern, mainMessage); !matched {
			v.warnings = append(v.warnings, "Conventional commits should start with lowercase after type")
		}
	} else {
		if len(mainMessage) > 0 && mainMessage[0] >= 'a' && mainMessage[0] <= 'z' {
			v.warnings = append(v.warnings, "Commit message should start with capital letter")
		}
	}

	return true
}

// ExtractCommitMessageFromCommand extracts commit message from git commit command
func (v *Validator) ExtractCommitMessageFromCommand(command string) string {
	patterns := []string{
		`-m\s+"([^"]+)"`,      // -m "message"
		`-m\s+'([^']+)'`,      // -m 'message'
		`-m\s+([^\s]+)`,       // -m message
		`--message="([^"]+)"`, // --message="message"
		`--message='([^']+)'`, // --message='message'
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(command)
		if len(matches) > 1 {
			return matches[1]
		}
	}

	return ""
}

// ValidateTool validates based on tool and command context
func (v *Validator) ValidateTool(toolName string, toolInput map[string]interface{}) bool {
	// Always validate repository context
	if !v.ValidateRepositoryContext() {
		return false
	}

	if toolName == "Bash" {
		if command, ok := toolInput["command"].(string); ok {
			// Git commit validation
			if strings.Contains(command, "git commit") {
				// Skip amend with no-edit
				if strings.Contains(command, "--amend") && strings.Contains(command, "--no-edit") {
					return true
				}

				// Validate commit message
				commitMessage := v.ExtractCommitMessageFromCommand(command)
				if commitMessage != "" {
					v.ValidateCommitMessage(commitMessage)
				}

				// Validate staged files
				v.ValidateStagedFiles()
			} else if strings.Contains(command, "git add") {
				// Git add validation
				v.ValidateStagedFiles()
			}
		}
	} else if toolName == "Write" {
		// Check if creating potentially sensitive files
		if filePath, ok := toolInput["file_path"].(string); ok {
			relPath, _ := filepath.Rel(v.repoRoot, filePath)
			for _, pattern := range forbiddenPatterns {
				if matched, _ := regexp.MatchString(pattern, relPath); matched {
					v.errors = append(v.errors, fmt.Sprintf("Forbidden file creation: %s", relPath))
					break
				}
			}
		}
	}

	return len(v.errors) == 0
}

// GetResult returns the validation result
func (v *Validator) GetResult() ValidationResult {
	return ValidationResult{
		Success:  len(v.errors) == 0,
		Errors:   v.errors,
		Warnings: v.warnings,
		Duration: time.Since(v.startTime).Milliseconds(),
	}
}

// PrintResults prints validation results to stderr
func (v *Validator) PrintResults() {
	if len(v.errors) > 0 {
		fmt.Fprintf(os.Stderr, "\n🚨 Git Validation Errors:\n")
		for _, error := range v.errors {
			fmt.Fprintf(os.Stderr, "❌ %s\n", error)
		}
	}

	if len(v.warnings) > 0 {
		fmt.Fprintf(os.Stderr, "\n⚠️  Git Validation Warnings:\n")
		for _, warning := range v.warnings {
			fmt.Fprintf(os.Stderr, "⚠️  %s\n", warning)
		}
	}

	if len(v.errors) > 0 {
		fmt.Fprintf(os.Stderr, "\n❌ Git operation blocked due to validation errors\n")
		fmt.Fprintf(os.Stderr, "Please fix the errors above and try again.\n")
	} else if len(v.warnings) > 0 {
		fmt.Fprintf(os.Stderr, "\nProceeding with warnings...\n")
	}
}