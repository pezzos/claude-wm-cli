package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// TestEnvironment represents a test git repository
type TestEnvironment struct {
	TempDir    string
	RepoDir    string
	Repo       *git.Repository
	WorkTree   *git.Worktree
	Validator  *GitValidator
	CleanupFunc func()
}

// SetupTestRepo creates a test git repository for testing
func SetupTestRepo(t *testing.T) *TestEnvironment {
	// Create temporary directory
	tempDir, err := ioutil.TempDir("", "git-validator-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	repoDir := filepath.Join(tempDir, "test-repo")
	err = os.MkdirAll(repoDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create repo dir: %v", err)
	}

	// Initialize git repository
	repo, err := git.PlainInit(repoDir, false)
	if err != nil {
		t.Fatalf("Failed to init git repo: %v", err)
	}

	// Set up basic git config
	cfg, err := repo.Config()
	if err != nil {
		t.Fatalf("Failed to get repo config: %v", err)
	}
	
	cfg.User.Name = "Test User"
	cfg.User.Email = "test@example.com"
	repo.Storer.SetConfig(cfg)

	workTree, err := repo.Worktree()
	if err != nil {
		t.Fatalf("Failed to get worktree: %v", err)
	}

	// Change to repo directory
	originalDir, _ := os.Getwd()
	os.Chdir(repoDir)

	// Create .gitignore file
	gitignoreContent := `.env
node_modules/
*.log
.DS_Store
`
	err = ioutil.WriteFile(filepath.Join(repoDir, ".gitignore"), []byte(gitignoreContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create .gitignore: %v", err)
	}

	// Add and commit .gitignore
	workTree.Add(".gitignore")
	
	// Create initial commit
	commit, err := workTree.Commit("Initial commit", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Test User",
			Email: "test@example.com",
			When:  time.Now(),
		},
	})
	if err != nil {
		t.Fatalf("Failed to create initial commit: %v", err)
	}

	// Verify commit was created
	_, err = repo.CommitObject(commit)
	if err != nil {
		t.Fatalf("Failed to verify initial commit: %v", err)
	}

	// Create validator
	validator, err := NewGitValidator()
	if err != nil {
		t.Fatalf("Failed to create git validator: %v", err)
	}

	cleanup := func() {
		os.Chdir(originalDir)
		os.RemoveAll(tempDir)
	}

	return &TestEnvironment{
		TempDir:    tempDir,
		RepoDir:    repoDir,
		Repo:       repo,
		WorkTree:   workTree,
		Validator:  validator,
		CleanupFunc: cleanup,
	}
}

// Test git validator initialization
func TestGitValidatorInitialization(t *testing.T) {
	env := SetupTestRepo(t)
	defer env.CleanupFunc()

	// Test successful initialization
	if env.Validator == nil {
		t.Fatal("Validator should be initialized")
	}

	if env.Validator.repo == nil {
		t.Fatal("Repository should be set")
	}

	if env.Validator.workTree == nil {
		t.Fatal("WorkTree should be set")
	}

	if env.Validator.repoRoot == "" {
		t.Fatal("Repository root should be set")
	}
}

// Test repository context validation
func TestValidateRepositoryContext(t *testing.T) {
	env := SetupTestRepo(t)
	defer env.CleanupFunc()

	result := env.Validator.ValidateRepositoryContext()
	
	if !result {
		t.Fatal("Repository context validation should pass")
	}

	if len(env.Validator.errors) > 0 {
		t.Errorf("Should not have errors: %v", env.Validator.errors)
	}
}

// Test staged files detection
func TestGetStagedFiles(t *testing.T) {
	env := SetupTestRepo(t)
	defer env.CleanupFunc()

	// Initially should have no staged files
	stagedFiles, err := env.Validator.GetStagedFiles()
	if err != nil {
		t.Fatalf("Failed to get staged files: %v", err)
	}

	if len(stagedFiles) != 0 {
		t.Errorf("Expected no staged files, got %d", len(stagedFiles))
	}

	// Create and stage a file
	testFile := filepath.Join(env.RepoDir, "test.txt")
	err = ioutil.WriteFile(testFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	_, err = env.WorkTree.Add("test.txt")
	if err != nil {
		t.Fatalf("Failed to stage test file: %v", err)
	}

	// Now should have one staged file
	stagedFiles, err = env.Validator.GetStagedFiles()
	if err != nil {
		t.Fatalf("Failed to get staged files after staging: %v", err)
	}

	if len(stagedFiles) != 1 {
		t.Errorf("Expected 1 staged file, got %d", len(stagedFiles))
	}

	if stagedFiles[0] != "test.txt" {
		t.Errorf("Expected 'test.txt', got '%s'", stagedFiles[0])
	}
}

// Test forbidden files validation
func TestValidateForbiddenFiles(t *testing.T) {
	env := SetupTestRepo(t)
	defer env.CleanupFunc()

	testCases := []struct {
		filename     string
		shouldError  bool
		shouldWarn   bool
		description  string
	}{
		{"test.txt", false, false, "normal file"},
		{".env", true, false, "environment file"},
		{"secret.key", true, false, "private key"},
		{"config.json", false, true, "config file (warning)"},
		{"test.log", false, true, "log file (warning)"},
		{"node_modules/package.json", true, false, "node_modules file"},
		{"credentials.json", true, false, "credentials file"},
		{"backup.bak", true, false, "backup file"},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			// Reset validator state
			env.Validator.errors = []string{}
			env.Validator.warnings = []string{}

			// Test validation
			result := env.Validator.ValidateForbiddenFiles([]string{tc.filename})

			if tc.shouldError && result {
				t.Errorf("Expected validation to fail for %s, but it passed", tc.filename)
			}

			if !tc.shouldError && !result {
				t.Errorf("Expected validation to pass for %s, but it failed", tc.filename)
			}

			if tc.shouldError && len(env.Validator.errors) == 0 {
				t.Errorf("Expected errors for %s, but got none", tc.filename)
			}

			if tc.shouldWarn && len(env.Validator.warnings) == 0 {
				t.Errorf("Expected warnings for %s, but got none", tc.filename)
			}
		})
	}
}

// Test commit message validation
func TestValidateCommitMessage(t *testing.T) {
	env := SetupTestRepo(t)
	defer env.CleanupFunc()

	testCases := []struct {
		message     string
		shouldError bool
		shouldWarn  bool
		description string
	}{
		{"Add new feature", false, false, "valid commit message"},
		{"Fix bug in validation", false, false, "valid fix message"},
		{"", false, false, "empty message (skipped)"},
		{"x", true, false, "too short message"},
		{"This is a very long commit message that exceeds the 72 character limit for the first line", false, true, "too long first line"},
		{"feat: add new feature", false, false, "conventional commit"},
		{"Feat: Add new feature", false, true, "conventional commit wrong case"},
		{"Added new feature", false, true, "past tense (should be imperative)"},
		{"🤖 Generated with [Claude Code]", true, false, "Claude signature"},
		{"Co-Authored-By: Test <test@example.com>", true, false, "co-authored commit"},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			// Reset validator state
			env.Validator.errors = []string{}
			env.Validator.warnings = []string{}

			// Test validation
			result := env.Validator.ValidateCommitMessage(tc.message)

			if tc.shouldError && len(env.Validator.errors) == 0 {
				t.Errorf("Expected errors for message '%s', but got none", tc.message)
			}

			if tc.shouldWarn && len(env.Validator.warnings) == 0 {
				t.Errorf("Expected warnings for message '%s', but got none", tc.message)
			}

			if !tc.shouldError && !tc.shouldWarn && (len(env.Validator.errors) > 0 || len(env.Validator.warnings) > 0) {
				t.Errorf("Expected no issues for message '%s', but got errors: %v, warnings: %v", 
					tc.message, env.Validator.errors, env.Validator.warnings)
			}

			_ = result // We don't check the return value as it's not the primary indicator
		})
	}
}

// Test .gitignore validation
func TestValidateGitignore(t *testing.T) {
	env := SetupTestRepo(t)
	defer env.CleanupFunc()

	// Should pass with existing .gitignore
	result := env.Validator.ValidateGitignore()
	if !result {
		t.Error("Gitignore validation should pass with valid .gitignore")
	}

	// Test missing .gitignore
	gitignorePath := filepath.Join(env.RepoDir, ".gitignore")
	os.Remove(gitignorePath)

	env.Validator.errors = []string{}
	result = env.Validator.ValidateGitignore()
	if result {
		t.Error("Gitignore validation should fail without .gitignore")
	}

	if len(env.Validator.errors) == 0 {
		t.Error("Should have errors when .gitignore is missing")
	}
}

// Test README analysis
func TestAnalyzeREADMEChanges(t *testing.T) {
	env := SetupTestRepo(t)
	defer env.CleanupFunc()

	// Create a README file
	readmeContent := `# Test Project

This is a test project.

## Features

- Feature 1
- Feature 2
`
	readmePath := filepath.Join(env.RepoDir, "README.md")
	err := ioutil.WriteFile(readmePath, []byte(readmeContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create README: %v", err)
	}

	testCases := []struct {
		files       []string
		shouldWarn  bool
		description string
	}{
		{[]string{"test.txt"}, false, "normal file change"},
		{[]string{"package.json"}, true, "dependency change should suggest README update"},
		{[]string{"api/users.js"}, true, "API change should suggest README update"},
		{[]string{"scripts/deploy.sh"}, true, "script change should suggest README update"},
		{[]string{"README.md", "test.txt"}, false, "README is being updated"},
		{[]string{"hooks/new-hook.py"}, true, "new hook should suggest README update"},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			// Reset validator state
			env.Validator.warnings = []string{}

			// Test analysis
			env.Validator.AnalyzeREADMEChanges(tc.files)

			hasREADMEWarning := false
			for _, warning := range env.Validator.warnings {
				if strings.Contains(warning, "README") {
					hasREADMEWarning = true
					break
				}
			}

			if tc.shouldWarn && !hasREADMEWarning {
				t.Errorf("Expected README warning for files %v, but got none", tc.files)
			}

			if !tc.shouldWarn && hasREADMEWarning {
				t.Errorf("Did not expect README warning for files %v, but got one", tc.files)
			}
		})
	}
}

// Test batched git operations
func TestGitOperationsBatch(t *testing.T) {
	env := SetupTestRepo(t)
	defer env.CleanupFunc()

	// Create and stage some files
	testFiles := []string{"file1.txt", "file2.txt", "file3.txt"}
	for _, filename := range testFiles {
		filePath := filepath.Join(env.RepoDir, filename)
		err := ioutil.WriteFile(filePath, []byte("test content"), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}
		env.WorkTree.Add(filename)
	}

	// Test batch operations
	batch := env.Validator.NewGitOperationsBatch()
	if batch == nil {
		t.Fatal("Batch should be created")
	}

	err := batch.ComputeAll()
	if err != nil {
		t.Fatalf("Failed to compute batch: %v", err)
	}

	if len(batch.stagedFiles) != len(testFiles) {
		t.Errorf("Expected %d staged files, got %d", len(testFiles), len(batch.stagedFiles))
	}

	// Test that subsequent calls use cached results
	batch2 := env.Validator.NewGitOperationsBatch()
	err = batch2.ComputeAll()
	if err != nil {
		t.Fatalf("Failed to compute second batch: %v", err)
	}

	// Verify batch is computed
	if !batch2.computed {
		t.Error("Second batch should be marked as computed")
	}
}

// Test full validation workflow
func TestFullValidationWorkflow(t *testing.T) {
	env := SetupTestRepo(t)
	defer env.CleanupFunc()

	// Test git commit validation
	toolInput := map[string]interface{}{
		"command": `git commit -m "Add new feature"`,
	}

	result := env.Validator.RunFullValidation("Bash", toolInput)
	if !result {
		t.Errorf("Validation should pass for valid commit. Errors: %v", env.Validator.errors)
	}

	// Test with forbidden file
	env.Validator.errors = []string{}
	env.Validator.warnings = []string{}

	// Create and stage a forbidden file
	envFile := filepath.Join(env.RepoDir, ".env")
	err := ioutil.WriteFile(envFile, []byte("SECRET=test"), 0644)
	if err != nil {
		t.Fatalf("Failed to create .env file: %v", err)
	}
	env.WorkTree.Add(".env")

	result = env.Validator.RunFullValidation("Bash", toolInput)
	if result {
		t.Error("Validation should fail with forbidden file staged")
	}

	if len(env.Validator.errors) == 0 {
		t.Error("Should have errors when forbidden file is staged")
	}
}

// Test commit message extraction
func TestExtractCommitMessageFromCommand(t *testing.T) {
	env := SetupTestRepo(t)
	defer env.CleanupFunc()

	testCases := []struct {
		command  string
		expected string
	}{
		{`git commit -m "Add new feature"`, "Add new feature"},
		{`git commit -m 'Fix bug in parser'`, "Fix bug in parser"},
		{`git commit --message="Update documentation"`, "Update documentation"},
		{`git commit --message='Refactor code'`, "Refactor code"},
		{`git commit -m simple`, "simple"},
		{`git add file.txt`, ""},
		{`git commit`, ""},
	}

	for _, tc := range testCases {
		result := env.Validator.ExtractCommitMessageFromCommand(tc.command)
		if result != tc.expected {
			t.Errorf("For command '%s', expected '%s', got '%s'", tc.command, tc.expected, result)
		}
	}
}

// Test performance with large number of files
func TestPerformanceWithManyFiles(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	env := SetupTestRepo(t)
	defer env.CleanupFunc()

	// Create many files
	numFiles := 100
	for i := 0; i < numFiles; i++ {
		filename := fmt.Sprintf("file_%03d.txt", i)
		filePath := filepath.Join(env.RepoDir, filename)
		err := ioutil.WriteFile(filePath, []byte("test content"), 0644)
		if err != nil {
			t.Fatalf("Failed to create file %s: %v", filename, err)
		}
		env.WorkTree.Add(filename)
	}

	// Time the validation
	start := time.Now()
	
	toolInput := map[string]interface{}{
		"command": `git commit -m "Add many files"`,
	}

	result := env.Validator.RunFullValidation("Bash", toolInput)
	
	duration := time.Since(start)
	
	t.Logf("Validation of %d files took %v", numFiles, duration)
	
	if !result {
		t.Errorf("Validation should pass. Errors: %v", env.Validator.errors)
	}

	// Performance target: should complete within 1 second for 100 files
	if duration > time.Second {
		t.Errorf("Performance target missed: took %v for %d files", duration, numFiles)
	}
}

// Test JSON output format
func TestJSONOutput(t *testing.T) {
	env := SetupTestRepo(t)
	defer env.CleanupFunc()

	// Run validation and capture output
	var buf bytes.Buffer
	
	// Temporarily redirect stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Create input
	input := ToolInput{
		ToolName: "Bash",
		ToolInput: map[string]interface{}{
			"command": `git commit -m "Test commit"`,
		},
	}

	inputJSON, _ := json.Marshal(input)
	
	// Simulate main function logic
	var testInput ToolInput
	json.Unmarshal(inputJSON, &testInput)
	
	validator, _ := NewGitValidator()
	success := validator.RunFullValidation(testInput.ToolName, testInput.ToolInput)
	
	result := ValidationResult{
		Success:  success,
		Errors:   validator.errors,
		Warnings: validator.warnings,
		Duration: time.Since(validator.startTime).Milliseconds(),
	}

	resultJSON, _ := json.Marshal(result)
	fmt.Printf("VALIDATION_RESULT: %s\n", resultJSON)

	// Restore stdout and read output
	w.Close()
	os.Stdout = oldStdout
	
	output, _ := ioutil.ReadAll(r)
	buf.Write(output)

	outputStr := buf.String()
	
	if !strings.Contains(outputStr, "VALIDATION_RESULT:") {
		t.Error("Output should contain VALIDATION_RESULT")
	}

	// Extract and parse JSON
	lines := strings.Split(outputStr, "\n")
	var jsonLine string
	for _, line := range lines {
		if strings.HasPrefix(line, "VALIDATION_RESULT:") {
			jsonLine = strings.TrimPrefix(line, "VALIDATION_RESULT: ")
			break
		}
	}

	if jsonLine == "" {
		t.Fatal("Could not find JSON result in output")
	}

	var parsedResult ValidationResult
	err := json.Unmarshal([]byte(jsonLine), &parsedResult)
	if err != nil {
		t.Fatalf("Failed to parse JSON result: %v", err)
	}

	if !parsedResult.Success {
		t.Error("Result should indicate success")
	}

	if parsedResult.Duration < 0 {
		t.Error("Duration should be non-negative")
	}
}

// Test cache integration
func TestCacheIntegration(t *testing.T) {
	env := SetupTestRepo(t)
	defer env.CleanupFunc()

	// Enable cache for testing
	os.Setenv("CACHE_ENABLED", "true")
	defer os.Unsetenv("CACHE_ENABLED")

	// Create a mock cache script
	cacheScript := filepath.Join(env.TempDir, "cache-integration.sh")
	cacheContent := `#!/bin/bash
# Mock cache script for testing
case "$1" in
    "get")
        echo "[]"
        ;;
    "set")
        # Do nothing
        ;;
    "invalidate-git")
        # Do nothing
        ;;
    *)
        exit 1
        ;;
esac
`
	err := ioutil.WriteFile(cacheScript, []byte(cacheContent), 0755)
	if err != nil {
		t.Fatalf("Failed to create cache script: %v", err)
	}

	os.Setenv("CACHE_INTEGRATION_SCRIPT", cacheScript)
	defer os.Unsetenv("CACHE_INTEGRATION_SCRIPT")

	// Test cache integration
	env.Validator.cacheEnabled = true
	
	// Test cache methods don't crash
	env.Validator.InvalidateGitCaches()
	env.Validator.integrateWithSharedCache()

	// Test batch caching
	batch := env.Validator.NewGitOperationsBatch()
	err = batch.ComputeAll()
	if err != nil {
		t.Fatalf("Failed to compute batch with cache enabled: %v", err)
	}
}

// Benchmark for performance comparison
func BenchmarkGitValidation(b *testing.B) {
	env := SetupTestRepo(&testing.T{})
	defer env.CleanupFunc()

	// Create test files
	for i := 0; i < 10; i++ {
		filename := fmt.Sprintf("bench_file_%d.txt", i)
		filePath := filepath.Join(env.RepoDir, filename)
		ioutil.WriteFile(filePath, []byte("benchmark content"), 0644)
		env.WorkTree.Add(filename)
	}

	toolInput := map[string]interface{}{
		"command": `git commit -m "Benchmark commit"`,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Reset validator state
		validator, _ := NewGitValidator()
		validator.RunFullValidation("Bash", toolInput)
	}
}