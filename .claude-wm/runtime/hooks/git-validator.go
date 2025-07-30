package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
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

// GitValidator combines git comprehensive validation and README validation
type GitValidator struct {
	repo         *git.Repository
	workTree     *git.Worktree
	repoRoot     string
	currentDir   string
	errors       []string
	warnings     []string
	cacheEnabled bool
	startTime    time.Time
}

// Forbidden files patterns from the original Python validator
var FORBIDDEN_FILES = map[string][]string{
	"private_keys": {
		`.*\.pem$`, `.*\.key$`, `.*private.*key.*`, `id_rsa.*`, `id_dsa.*`, `id_ecdsa.*`, `id_ed25519.*`,
	},
	"env_files": {
		`^\.env$`, `^\.env\.[^.]+$`, `.*\.env\.(?!example|sample|template).*$`,
	},
	"credentials": {
		`.*credentials.*\.(json|yml|yaml)$`, `.*service[-_]?account.*\.json$`, 
		`.*secrets?\.(json|yml|yaml|txt)$`, `.*password.*\.(txt|json|yml|yaml)$`,
	},
	"test_scripts": {
		`.*test[-_]?script.*\.(sh|bash|py|js)$`, `.*scratch.*\.(py|js|ts|sh)$`, 
		`.*temp[-_]?test.*`, `.*debug[-_]?script.*`,
	},
	"backups": {
		`.*\.backup$`, `.*\.bak$`, `.*\.old$`, `.*~$`, `.*\.(orig|save)$`,
	},
	"archives": {
		`.*\.(zip|tar|tar\.gz|tgz|rar|7z)$`,
	},
	"large_files": {
		`.*\.(mp4|avi|mov|mkv|wmv)$`, `.*\.(psd|ai|sketch|fig)$`, `.*\.(exe|dmg|pkg|deb|rpm)$`,
	},
}

var WARNING_FILES = map[string][]string{
	"configs": {`config\.(json|yml|yaml)$`, `settings\.(json|yml|yaml)$`},
	"data":    {`.*\.(csv|xlsx|xls)$`, `.*\.sql$`, `.*dump.*`},
	"logs":    {`.*\.log$`, `debug\.txt$`, `error\.txt$`},
}

// README analysis patterns
var SIGNIFICANT_PATTERNS = []struct {
	Pattern     string
	FileType    string
	Sections    []string
	Suggestion  string
}{
	{`/api/`, "API endpoint", []string{"API Documentation", "API Reference", "Endpoints"}, 
		"Document new API endpoints, request/response formats, and authentication"},
	{`config\.`, "configuration", []string{"Configuration", "Environment Variables", "Setup"}, 
		"Update configuration instructions and environment variable documentation"},
	{`package\.json|requirements\.txt|go\.mod`, "dependencies", []string{"Installation", "Requirements", "Dependencies"}, 
		"Update installation instructions and dependency requirements"},
	{`\.sh$|/scripts/`, "scripts", []string{"Scripts", "Usage", "Commands"}, 
		"Document new scripts, their purpose, and usage examples"},
	{`/hooks/.*\.py$`, "hooks", []string{"Hooks Overview", "Configuration", "Features"}, 
		"Update hooks documentation with new hooks and their purposes"},
	{`/components/`, "components", []string{"Components", "UI Documentation", "Features"}, 
		"Document new UI components and their usage"},
}

// NewGitValidator creates a new Git validator instance
func NewGitValidator() (*GitValidator, error) {
	gv := &GitValidator{
		errors:       make([]string, 0),
		warnings:     make([]string, 0),
		cacheEnabled: os.Getenv("CACHE_ENABLED") == "true",
		startTime:    time.Now(),
	}

	// Get current working directory
	var err error
	gv.currentDir, err = os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current directory: %v", err)
	}

	// Open git repository
	gv.repo, err = git.PlainOpen(gv.currentDir)
	if err != nil {
		// Try opening from parent directories
		dir := gv.currentDir
		for dir != "/" {
			dir = filepath.Dir(dir)
			gv.repo, err = git.PlainOpen(dir)
			if err == nil {
				gv.repoRoot = dir
				break
			}
		}
		if err != nil {
			return nil, fmt.Errorf("not in a git repository: %v", err)
		}
	} else {
		gv.repoRoot = gv.currentDir
	}

	// Get worktree
	gv.workTree, err = gv.repo.Worktree()
	if err != nil {
		return nil, fmt.Errorf("failed to get worktree: %v", err)
	}

	return gv, nil
}

// ValidateRepositoryContext validates git repository context and location
func (gv *GitValidator) ValidateRepositoryContext() bool {
	// Check if we're at repository root
	if gv.repoRoot != gv.currentDir {
		gv.warnings = append(gv.warnings, 
			fmt.Sprintf("Not at repository root. Root: %s, Current: %s", gv.repoRoot, gv.currentDir))
	}

	// Check repository health
	_, err := gv.repo.Head()
	if err != nil {
		gv.errors = append(gv.errors, fmt.Sprintf("Repository head error: %v", err))
		return false
	}

	return true
}

// GitOperationsBatch represents a batch of git operations with caching
type GitOperationsBatch struct {
	validator    *GitValidator
	stagedFiles  []string
	gitStatus    git.Status
	headCommit   *object.Commit
	branch       string
	cacheKey     string
	computed     bool
}

// NewGitOperationsBatch creates a new batch of git operations
func (gv *GitValidator) NewGitOperationsBatch() *GitOperationsBatch {
	// Create cache key based on current git state
	head, _ := gv.repo.Head()
	headHash := ""
	if head != nil {
		headHash = head.Hash().String()[:8]
	}
	
	cacheKey := fmt.Sprintf("git_batch_%s_%d", headHash, time.Now().Unix()/30) // 30-second cache window
	
	return &GitOperationsBatch{
		validator: gv,
		cacheKey:  cacheKey,
		computed:  false,
	}
}

// ComputeAll computes all git operations in one batch
func (batch *GitOperationsBatch) ComputeAll() error {
	if batch.computed {
		return nil
	}

	// Check cache first
	if batch.validator.cacheEnabled {
		if cached := batch.getCachedBatch(); cached != nil {
			*batch = *cached
			batch.computed = true
			log.Printf("🎯 Cache HIT: Git operations batch")
			return nil
		}
	}

	log.Printf("💫 Cache MISS: Computing git operations batch")
	
	// Batch all git operations together
	var err error
	
	// Get git status (includes staged files info)
	batch.gitStatus, err = batch.validator.workTree.Status()
	if err != nil {
		return fmt.Errorf("failed to get git status: %v", err)
	}

	// Extract staged files from status
	for file, fileStatus := range batch.gitStatus {
		if fileStatus.Staging != git.Unmodified {
			batch.stagedFiles = append(batch.stagedFiles, file)
		}
	}

	// Get head commit
	head, err := batch.validator.repo.Head()
	if err == nil {
		batch.headCommit, _ = batch.validator.repo.CommitObject(head.Hash())
		batch.branch = head.Name().Short()
	}

	batch.computed = true

	// Cache the batch results
	if batch.validator.cacheEnabled {
		batch.cacheBatch()
	}

	return nil
}

// GetStagedFiles returns list of staged files using batched operations
func (gv *GitValidator) GetStagedFiles() ([]string, error) {
	batch := gv.NewGitOperationsBatch()
	if err := batch.ComputeAll(); err != nil {
		return nil, err
	}
	return batch.stagedFiles, nil
}

// GetGitStatus returns git status using batched operations
func (gv *GitValidator) GetGitStatus() (git.Status, error) {
	batch := gv.NewGitOperationsBatch()
	if err := batch.ComputeAll(); err != nil {
		return nil, err
	}
	return batch.gitStatus, nil
}

// GetCurrentBranch returns current branch using batched operations
func (gv *GitValidator) GetCurrentBranch() (string, error) {
	batch := gv.NewGitOperationsBatch()
	if err := batch.ComputeAll(); err != nil {
		return "", err
	}
	return batch.branch, nil
}

// GetHeadCommit returns head commit using batched operations
func (gv *GitValidator) GetHeadCommit() (*object.Commit, error) {
	batch := gv.NewGitOperationsBatch()
	if err := batch.ComputeAll(); err != nil {
		return nil, err
	}
	return batch.headCommit, nil
}

// ValidateStagedFiles validates staged files for security and size
func (gv *GitValidator) ValidateStagedFiles() bool {
	stagedFiles, err := gv.GetStagedFiles()
	if err != nil {
		gv.errors = append(gv.errors, fmt.Sprintf("Failed to get staged files: %v", err))
		return false
	}

	if len(stagedFiles) == 0 {
		return true
	}

	// Check for forbidden files
	forbiddenPatterns := []string{
		`^\.git/`, `^\.claude/`, `\.log$`, `node_modules/`, `^\.env$`, `\.DS_Store$`,
	}

	var forbiddenFiles []string
	for _, filePath := range stagedFiles {
		for _, pattern := range forbiddenPatterns {
			matched, err := regexp.MatchString(pattern, filePath)
			if err == nil && matched {
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

	// Check file sizes
	var largeFiles []struct {
		path string
		size int64
	}
	for _, filePath := range stagedFiles {
		fullPath := filepath.Join(gv.repoRoot, filePath)
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
		gv.warnings = append(gv.warnings, "Large files detected (>10MB):")
		for _, file := range largeFiles {
			gv.warnings = append(gv.warnings, 
				fmt.Sprintf("  - %s (%.1fMB)", file.path, float64(file.size)/(1024*1024)))
		}
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
func (gv *GitValidator) ValidateGitignore() bool {
	gitignorePath := filepath.Join(gv.repoRoot, ".gitignore")
	
	if _, err := os.Stat(gitignorePath); os.IsNotExist(err) {
		gv.errors = append(gv.errors, "No .gitignore file found")
		gv.errors = append(gv.errors, "Create .gitignore with basic security patterns")
		return false
	}

	// Parse existing .gitignore
	content, err := ioutil.ReadFile(gitignorePath)
	if err != nil {
		gv.errors = append(gv.errors, fmt.Sprintf("Failed to read .gitignore: %v", err))
		return false
	}

	var gitignorePatterns []string
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			gitignorePatterns = append(gitignorePatterns, line)
		}
	}

	// Check for missing critical patterns
	criticalPatterns := []string{".git/", ".claude/", "node_modules/", ".env"}
	var missingPatterns []string

	for _, pattern := range criticalPatterns {
		found := false
		for _, gitignorePattern := range gitignorePatterns {
			if strings.Contains(gitignorePattern, pattern) || 
			   strings.HasPrefix(gitignorePattern, strings.TrimSuffix(pattern, "/")) {
				found = true
				break
			}
		}
		if !found {
			missingPatterns = append(missingPatterns, pattern)
		}
	}

	if len(missingPatterns) > 0 {
		gv.warnings = append(gv.warnings, "Missing critical .gitignore patterns:")
		for _, pattern := range missingPatterns {
			gv.warnings = append(gv.warnings, fmt.Sprintf("  - %s", pattern))
		}
	}

	return true
}

// ValidateCommitMessage validates commit message format and content
func (gv *GitValidator) ValidateCommitMessage(message string) bool {
	if message == "" {
		return true // Skip if no message provided
	}

	// Check for Co-Authored-By (blocked per project rules)
	if strings.Contains(message, "Co-Authored-By") || strings.Contains(strings.ToLower(message), "co-authored-by") {
		gv.errors = append(gv.errors, "Co-authored commits are not allowed per project rules")
	}

	// Check for Claude signature (should be removed)
	if strings.Contains(message, "🤖 Generated with [Claude Code]") || strings.Contains(message, "🤖 Generated with Claude") {
		gv.errors = append(gv.errors, "Remove Claude signature from commit messages")
	}

	// Extract main message (first line)
	lines := strings.Split(strings.TrimSpace(message), "\n")
	if len(lines) == 0 {
		gv.errors = append(gv.errors, "Empty commit message")
		return false
	}

	mainMessage := strings.TrimSpace(lines[0])

	// Check message length
	if len(mainMessage) > 72 {
		gv.warnings = append(gv.warnings, 
			fmt.Sprintf("First line should be ≤72 characters (current: %d)", len(mainMessage)))
	} else if len(mainMessage) < 10 {
		gv.errors = append(gv.errors, "Commit message too short (minimum 10 characters)")
	}

	// Check for conventional commit format
	conventionalPattern := `^(feat|fix|docs|style|refactor|test|chore|perf|ci|build|revert)(\(.+\))?: .+`
	if matched, _ := regexp.MatchString(conventionalPattern, mainMessage); matched {
		lowercasePattern := `^[a-z]+(\(.+\))?: [a-z]`
		if matched, _ := regexp.MatchString(lowercasePattern, mainMessage); !matched {
			gv.warnings = append(gv.warnings, "Conventional commits should start with lowercase after type")
		}
	} else {
		if len(mainMessage) > 0 && !isUppercase(rune(mainMessage[0])) {
			gv.warnings = append(gv.warnings, "Commit message should start with capital letter")
		}
	}

	// Check for imperative mood
	pastTensePatterns := []string{
		`\b(added|deleted|changed|fixed|updated|removed|created|modified)\b`,
		`\b(implemented|refactored|improved|optimized)\b`,
	}
	for _, pattern := range pastTensePatterns {
		if matched, _ := regexp.MatchString("(?i)"+pattern, mainMessage); matched {
			gv.warnings = append(gv.warnings, "Use imperative mood (e.g., 'Add' not 'Added')")
			break
		}
	}

	// Check body formatting
	if len(lines) > 1 {
		if len(lines) > 1 && strings.TrimSpace(lines[1]) != "" {
			gv.errors = append(gv.errors, "Add blank line after commit message summary")
		}

		for i, line := range lines[2:] {
			if len(line) > 72 && !strings.HasPrefix(line, "http") {
				gv.warnings = append(gv.warnings, fmt.Sprintf("Line %d exceeds 72 characters", i+3))
			}
		}
	}

	return true
}

// ValidateForbiddenFiles checks for forbidden files in commit
func (gv *GitValidator) ValidateForbiddenFiles(files []string) bool {
	type Issue struct {
		File     string
		Category string
		Severity string
	}

	var issues []Issue

	for _, filePath := range files {
		fullPath := filepath.Join(gv.repoRoot, filePath)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			continue
		}

		// Check against forbidden patterns
		for category, patterns := range FORBIDDEN_FILES {
			for _, pattern := range patterns {
				if matched, _ := regexp.MatchString("(?i)"+pattern, filePath); matched {
					issues = append(issues, Issue{
						File:     filePath,
						Category: category,
						Severity: "high",
					})
					goto nextFile
				}
			}
		}

		// Check against warning patterns
		for category, patterns := range WARNING_FILES {
			for _, pattern := range patterns {
				if matched, _ := regexp.MatchString("(?i)"+pattern, filePath); matched {
					// Don't add if already in issues
					found := false
					for _, issue := range issues {
						if issue.File == filePath {
							found = true
							break
						}
					}
					if !found {
						issues = append(issues, Issue{
							File:     filePath,
							Category: category,
							Severity: "medium",
						})
					}
					break
				}
			}
		}
		nextFile:
	}

	// Process issues
	var highSeverity, mediumSeverity []Issue
	for _, issue := range issues {
		if issue.Severity == "high" {
			highSeverity = append(highSeverity, issue)
		} else {
			mediumSeverity = append(mediumSeverity, issue)
		}
	}

	if len(highSeverity) > 0 {
		gv.errors = append(gv.errors, "Forbidden files detected:")
		for _, issue := range highSeverity {
			gv.errors = append(gv.errors, 
				fmt.Sprintf("  - %s (%s)", issue.File, strings.ReplaceAll(issue.Category, "_", " ")))
		}
	}

	if len(mediumSeverity) > 0 {
		gv.warnings = append(gv.warnings, "Warning files detected:")
		for _, issue := range mediumSeverity {
			gv.warnings = append(gv.warnings, 
				fmt.Sprintf("  - %s (%s)", issue.File, strings.ReplaceAll(issue.Category, "_", " ")))
		}
	}

	return len(highSeverity) == 0
}

// AnalyzeREADMEChanges analyzes changes and suggests README updates
func (gv *GitValidator) AnalyzeREADMEChanges(files []string) {
	if len(files) == 0 {
		return
	}

	// Find README files
	readmeFiles := gv.findREADMEFiles()
	if len(readmeFiles) == 0 {
		gv.warnings = append(gv.warnings, "No README file found!")
		gv.warnings = append(gv.warnings, "Consider creating a README.md to document your project")
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

	// Analyze changes for suggestions
	suggestions := gv.analyzeChanges(files)
	newFiles := gv.checkForNewFiles(files)

	if (len(suggestions) > 0 || len(newFiles) > 0) && !readmeUpdated {
		gv.warnings = append(gv.warnings, "README Update Reminder:")

		if len(newFiles) > 0 {
			gv.warnings = append(gv.warnings, "New files added that may need documentation:")
			for i, file := range newFiles {
				if i >= 5 { // Limit output
					break
				}
				gv.warnings = append(gv.warnings, fmt.Sprintf("  • %s (%s)", file.File, file.Type))
			}
		}

		if len(suggestions) > 0 {
			gv.warnings = append(gv.warnings, "Based on your changes, consider updating these README sections:")
			for _, suggestion := range suggestions {
				gv.warnings = append(gv.warnings, fmt.Sprintf("  %s changes detected:", strings.ToUpper(suggestion.Type)))
				fileList := strings.Join(suggestion.Files[:min(3, len(suggestion.Files))], ", ")
				gv.warnings = append(gv.warnings, fmt.Sprintf("  Files: %s", fileList))
				if len(suggestion.Files) > 3 {
					gv.warnings = append(gv.warnings, fmt.Sprintf("         ... and %d more", len(suggestion.Files)-3))
				}
				gv.warnings = append(gv.warnings, fmt.Sprintf("  💡 %s", suggestion.Suggestion))
				sectionList := strings.Join(suggestion.Sections[:min(2, len(suggestion.Sections))], ", ")
				gv.warnings = append(gv.warnings, fmt.Sprintf("  📍 Suggested sections: %s", sectionList))
			}
		}

		gv.warnings = append(gv.warnings, "✅ Proceeding with commit - don't forget to update docs later!")
	}
}

// Helper methods for README analysis
type READMESuggestion struct {
	Type       string
	Files      []string
	Sections   []string
	Suggestion string
}

type NewFile struct {
	File string
	Type string
}

func (gv *GitValidator) findREADMEFiles() []string {
	var readmeFiles []string
	patterns := []string{"README.md", "README.rst", "README.txt", "readme.md", "Readme.md"}

	for _, pattern := range patterns {
		path := filepath.Join(gv.repoRoot, pattern)
		if _, err := os.Stat(path); err == nil {
			readmeFiles = append(readmeFiles, pattern)
		}
	}

	// Also check docs folder
	docsPatterns := []string{"docs/README.md", "documentation/README.md"}
	for _, pattern := range docsPatterns {
		path := filepath.Join(gv.repoRoot, pattern)
		if _, err := os.Stat(path); err == nil {
			readmeFiles = append(readmeFiles, pattern)
		}
	}

	return readmeFiles
}

func (gv *GitValidator) analyzeChanges(files []string) []READMESuggestion {
	var suggestions []READMESuggestion
	
	// Categorize changes using separate maps for each category
	categoryFiles := map[string][]string{
		"api":           {},
		"configuration": {},
		"dependencies":  {},
		"scripts":       {},
		"hooks":         {},
		"components":    {},
		"features":      {},
	}

	// Define category metadata
	categoryMetadata := map[string]struct {
		sections   []string
		suggestion string
	}{
		"api":           {[]string{"API Documentation", "API Reference", "Endpoints", "Routes"}, "Document new API endpoints, request/response formats, and authentication"},
		"configuration": {[]string{"Configuration", "Environment Variables", "Setup", "Installation"}, "Update configuration instructions and environment variable documentation"},
		"dependencies":  {[]string{"Installation", "Requirements", "Dependencies", "Prerequisites"}, "Update installation instructions and dependency requirements"},
		"scripts":       {[]string{"Scripts", "Usage", "Commands", "Development"}, "Document new scripts, their purpose, and usage examples"},
		"hooks":         {[]string{"Hooks Overview", "Configuration", "Features"}, "Update hooks documentation with new hooks and their purposes"},
		"components":    {[]string{"Components", "UI Documentation", "Features"}, "Document new UI components and their usage"},
		"features":      {[]string{"Features", "What's New", "Functionality"}, "Add documentation for new features and capabilities"},
	}

	for _, filePath := range files {
		fileLower := strings.ToLower(filePath)

		// Categorize file
		if strings.Contains(filePath, "/api/") || strings.Contains(filePath, "route.") || 
		   strings.Contains(filePath, "controller.") || strings.Contains(filePath, "endpoint") {
			categoryFiles["api"] = append(categoryFiles["api"], filePath)
		} else if strings.Contains(fileLower, "config.") || strings.Contains(fileLower, "settings.") || 
				  strings.Contains(fileLower, ".env.example") {
			categoryFiles["configuration"] = append(categoryFiles["configuration"], filePath)
		} else if filePath == "package.json" || filePath == "requirements.txt" || filePath == "Gemfile" || 
				  filePath == "go.mod" || filePath == "Cargo.toml" || filePath == "pom.xml" {
			categoryFiles["dependencies"] = append(categoryFiles["dependencies"], filePath)
		} else if strings.HasSuffix(filePath, ".sh") || strings.HasSuffix(filePath, ".bash") || 
				  strings.Contains(filePath, "scripts/") {
			categoryFiles["scripts"] = append(categoryFiles["scripts"], filePath)
		} else if strings.HasSuffix(filePath, ".py") && strings.Contains(fileLower, "hook") {
			categoryFiles["hooks"] = append(categoryFiles["hooks"], filePath)
		} else if strings.Contains(filePath, "/components/") || strings.Contains(filePath, "/pages/") || 
				  strings.Contains(filePath, "/views/") {
			categoryFiles["components"] = append(categoryFiles["components"], filePath)
		} else if strings.Contains(fileLower, "feature") || strings.Contains(fileLower, "service") || 
				  strings.Contains(fileLower, "util") || strings.Contains(fileLower, "helper") || 
				  strings.Contains(fileLower, "lib") {
			categoryFiles["features"] = append(categoryFiles["features"], filePath)
		}
	}

	// Generate suggestions for categories with files
	for catType, files := range categoryFiles {
		if len(files) > 0 {
			metadata := categoryMetadata[catType]
			suggestions = append(suggestions, READMESuggestion{
				Type:       catType,
				Files:      files,
				Sections:   metadata.sections,
				Suggestion: metadata.suggestion,
			})
		}
	}

	return suggestions
}

func (gv *GitValidator) checkForNewFiles(files []string) []NewFile {
	var newFiles []NewFile

	for _, filePath := range files {
		for _, pattern := range SIGNIFICANT_PATTERNS {
			if matched, _ := regexp.MatchString(pattern.Pattern, filePath); matched {
				// Check if it's a new file (simplified heuristic)
				fullPath := filepath.Join(gv.repoRoot, filePath)
				if _, err := os.Stat(fullPath); err == nil {
					// Simple way to check if file is new - check git status
					// This is a simplified version, could be enhanced
					newFiles = append(newFiles, NewFile{
						File: filePath,
						Type: pattern.FileType,
					})
				}
				break
			}
		}
	}

	return newFiles
}

// ExtractCommitMessageFromCommand extracts commit message from git commit command
func (gv *GitValidator) ExtractCommitMessageFromCommand(command string) string {
	patterns := []string{
		`-m\s+"([^"]+)"`,   // -m "message"
		`-m\s+'([^']+)'`,   // -m 'message'
		`-m\s+([^\s]+)`,    // -m message
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

// RunFullValidation runs comprehensive git validation
func (gv *GitValidator) RunFullValidation(toolName string, toolInput map[string]interface{}) bool {
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

				// Get staged files for forbidden check and README analysis
				stagedFiles, err := gv.GetStagedFiles()
				if err == nil {
					gv.ValidateForbiddenFiles(stagedFiles)
					gv.AnalyzeREADMEChanges(stagedFiles)
				}
			} else if strings.Contains(command, "git add") {
				// Git add validation
				gv.ValidateStagedFiles()
				gv.ValidateGitignore()
			}
		}
	} else if toolName == "Write" {
		// Check if creating potentially sensitive files
		if filePath, ok := toolInput["file_path"].(string); ok {
			gv.ValidateForbiddenFiles([]string{filePath})
		}
	}

	// Return true if no errors (warnings are okay)
	return len(gv.errors) == 0
}

// PrintResults prints validation results
func (gv *GitValidator) PrintResults() {
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

// Cache integration methods for batched operations
func (batch *GitOperationsBatch) getCachedBatch() *GitOperationsBatch {
	if !batch.validator.cacheEnabled {
		return nil
	}

	cacheScript := os.Getenv("CACHE_INTEGRATION_SCRIPT")
	if cacheScript == "" {
		return nil
	}

	cmd := exec.Command(cacheScript, "get", batch.cacheKey)
	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	var cachedBatch GitOperationsBatch
	if err := json.Unmarshal(output, &cachedBatch); err != nil {
		return nil
	}

	// Restore the validator reference (not serialized)
	cachedBatch.validator = batch.validator
	return &cachedBatch
}

func (batch *GitOperationsBatch) cacheBatch() {
	if !batch.validator.cacheEnabled {
		return
	}

	cacheScript := os.Getenv("CACHE_INTEGRATION_SCRIPT")
	if cacheScript == "" {
		return
	}

	// Create a serializable version of the batch (without validator reference)
	serializable := struct {
		StagedFiles []string `json:"staged_files"`
		Branch      string   `json:"branch"`
		CacheKey    string   `json:"cache_key"`
		Computed    bool     `json:"computed"`
	}{
		StagedFiles: batch.stagedFiles,
		Branch:      batch.branch,
		CacheKey:    batch.cacheKey,
		Computed:    batch.computed,
	}

	batchJSON, err := json.Marshal(serializable)
	if err != nil {
		return
	}

	cmd := exec.Command(cacheScript, "set", batch.cacheKey, string(batchJSON))
	cmd.Run() // Don't block on cache errors
}

// Enhanced cache integration with shared cache system
func (gv *GitValidator) integrateWithSharedCache() {
	if !gv.cacheEnabled {
		return
	}

	// Use the shared cache binary directly
	cacheDir := filepath.Join(filepath.Dir(gv.currentDir), "hooks", "cache")
	cacheBin := filepath.Join(cacheDir, "shared-cache")
	
	// Check if shared cache binary exists
	if _, err := os.Stat(cacheBin); err == nil {
		// Use shared cache for git info
		cmd := exec.Command(cacheBin, "get-git-info")
		if _, err := cmd.Output(); err == nil {
			log.Printf("🎯 Shared cache integration: git info available")
		}
	}
}

// InvalidateGitCaches invalidates all git-related caches
func (gv *GitValidator) InvalidateGitCaches() {
	if !gv.cacheEnabled {
		return
	}

	cacheScript := os.Getenv("CACHE_INTEGRATION_SCRIPT")
	if cacheScript == "" {
		return
	}

	// Invalidate git cache
	cmd := exec.Command(cacheScript, "invalidate-git")
	cmd.Run()

	// Also invalidate our batch caches
	cmd = exec.Command(cacheScript, "invalidate-pattern", "git_batch_*")
	cmd.Run()

	log.Printf("🗑️  Git caches invalidated")
}

// Legacy cache methods for backwards compatibility
func (gv *GitValidator) getCachedStagedFiles() []string {
	batch := gv.NewGitOperationsBatch()
	if cached := batch.getCachedBatch(); cached != nil {
		return cached.stagedFiles
	}
	return nil
}

func (gv *GitValidator) cacheStagedFiles(files []string) {
	// This is now handled by the batch caching system
	// Keep for backwards compatibility but functionality moved to batching
}

// Utility functions
func isUppercase(r rune) bool {
	return r >= 'A' && r <= 'Z'
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
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
	validator, err := NewGitValidator()
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