package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"claude-hooks-orchestrator/patterns"
)

// ToolInput represents the input structure for Claude Code tools
type ToolInput struct {
	ToolName  string                 `json:"tool_name"`
	ToolInput map[string]interface{} `json:"tool_input"`
}

// EnvVariable represents a parsed environment variable
type EnvVariable struct {
	Value    string `json:"value"`
	Line     int    `json:"line"`
	HasValue bool   `json:"has_value"`
}

// EnvFiles represents found environment files
type EnvFiles struct {
	Env     []string `json:"env"`
	Example []string `json:"example"`
}

// Issue represents a validation issue
type Issue struct {
	Type       string `json:"type"`
	Variable   string `json:"variable"`
	Line       int    `json:"line,omitempty"`
	Suggestion string `json:"suggestion,omitempty"`
	Severity   string `json:"severity"`
}

// FileComparison represents the comparison result between env files
type FileComparison struct {
	EnvFile     string  `json:"env_file"`
	ExampleFile string  `json:"example_file"`
	Issues      []Issue `json:"issues"`
}

// EnvSyncValidator manages environment file validation
type EnvSyncValidator struct {
	projectRoot     string
	envVarPatterns  []*regexp.Regexp
	secretPatterns  []*regexp.Regexp
	fileExtensions  []string
	cacheMutex      sync.RWMutex
	parsedFiles     map[string]map[string]EnvVariable
	initOnce        sync.Once
}

// Global validator instance
var validator *EnvSyncValidator

// Initialize the validator with compiled patterns
func (v *EnvSyncValidator) init() {
	// Use pre-compiled patterns from the patterns package
	envPatterns := patterns.GetEnvPatterns()
	v.envVarPatterns = make([]*regexp.Regexp, 0, len(envPatterns))
	for _, pattern := range envPatterns {
		v.envVarPatterns = append(v.envVarPatterns, pattern)
	}

	// Use pre-compiled security patterns
	securityPatterns := patterns.GetSecurityPatterns()
	v.secretPatterns = make([]*regexp.Regexp, 0, len(securityPatterns))
	for _, pattern := range securityPatterns {
		v.secretPatterns = append(v.secretPatterns, pattern)
	}

	v.fileExtensions = []string{".js", ".ts", ".py", ".rb", ".php", ".go"}
	v.parsedFiles = make(map[string]map[string]EnvVariable)
}

// parseEnvFile parses an environment file and extracts variables
func (v *EnvSyncValidator) parseEnvFile(filePath string) map[string]EnvVariable {
	v.cacheMutex.Lock()
	defer v.cacheMutex.Unlock()

	// Check cache first
	if cached, exists := v.parsedFiles[filePath]; exists {
		return cached
	}

	variables := make(map[string]EnvVariable)

	file, err := os.Open(filePath)
	if err != nil {
		v.parsedFiles[filePath] = variables
		return variables
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 0

	// Use pre-compiled KEY=VALUE pattern
	keyValueRegex := patterns.GetPatterns().EnvKeyValue

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Match KEY=value pattern
		matches := keyValueRegex.FindStringSubmatch(line)
		if len(matches) == 3 {
			key := matches[1]
			value := strings.TrimSpace(matches[2])

			// Remove quotes if present
			if len(value) >= 2 {
				if (value[0] == '"' && value[len(value)-1] == '"') ||
					(value[0] == '\'' && value[len(value)-1] == '\'') {
					value = value[1 : len(value)-1]
				}
			}

			variables[key] = EnvVariable{
				Value:    value,
				Line:     lineNum,
				HasValue: value != "",
			}
		}
	}

	// Cache the result
	v.parsedFiles[filePath] = variables
	return variables
}

// findEnvFiles finds all .env and .env.example files in the project
func (v *EnvSyncValidator) findEnvFiles() EnvFiles {
	envFiles := EnvFiles{
		Env:     make([]string, 0),
		Example: make([]string, 0),
	}

	// Common patterns
	envPatterns := []string{".env", ".env.local", ".env.development", ".env.production", ".env.test"}
	examplePatterns := []string{".env.example", ".env.sample", ".env.template", ".env.example.local"}

	for _, pattern := range envPatterns {
		path := filepath.Join(v.projectRoot, pattern)
		if _, err := os.Stat(path); err == nil {
			envFiles.Env = append(envFiles.Env, path)
		}
	}

	for _, pattern := range examplePatterns {
		path := filepath.Join(v.projectRoot, pattern)
		if _, err := os.Stat(path); err == nil {
			envFiles.Example = append(envFiles.Example, path)
		}
	}

	return envFiles
}

// suggestSafeValue suggests a safe example value for a given key
func (v *EnvSyncValidator) suggestSafeValue(key, actualValue string) string {
	keyLower := strings.ToLower(key)

	// Database URLs
	dbKeywords := []string{"database", "db", "postgres", "mysql", "mongo", "redis"}
	for _, db := range dbKeywords {
		if strings.Contains(keyLower, db) {
			if strings.Contains(keyLower, "url") || strings.Contains(keyLower, "uri") {
				switch {
				case strings.Contains(keyLower, "postgres"):
					return "postgresql://user:password@localhost:5432/dbname"
				case strings.Contains(keyLower, "mysql"):
					return "mysql://user:password@localhost:3306/dbname"
				case strings.Contains(keyLower, "mongo"):
					return "mongodb://localhost:27017/dbname"
				case strings.Contains(keyLower, "redis"):
					return "redis://localhost:6379"
				default:
					return "protocol://user:password@host:port/database"
				}
			}
		}
	}

	// API keys and secrets
	secretKeywords := []string{"key", "secret", "token", "password", "pwd"}
	for _, secret := range secretKeywords {
		if strings.Contains(keyLower, secret) {
			switch {
			case strings.Contains(keyLower, "api"):
				return "your-api-key-here"
			case strings.Contains(keyLower, "secret"):
				return "your-secret-here"
			case strings.Contains(keyLower, "token"):
				return "your-token-here"
			case strings.Contains(keyLower, "password") || strings.Contains(keyLower, "pwd"):
				return "your-password-here"
			default:
				return "your-value-here"
			}
		}
	}

	// URLs and endpoints
	urlKeywords := []string{"url", "uri", "endpoint", "host"}
	for _, url := range urlKeywords {
		if strings.Contains(keyLower, url) {
			switch {
			case strings.Contains(keyLower, "api"):
				return "https://api.example.com"
			case strings.Contains(keyLower, "webhook"):
				return "https://example.com/webhook"
			default:
				return "https://example.com"
			}
		}
	}

	// Ports
	if strings.Contains(keyLower, "port") {
		switch {
		case strings.Contains(keyLower, "db") || strings.Contains(keyLower, "database"):
			return "5432"
		case strings.Contains(keyLower, "redis"):
			return "6379"
		default:
			return "3000"
		}
	}

	// Email
	if strings.Contains(keyLower, "email") || strings.Contains(keyLower, "mail") {
		if strings.Contains(keyLower, "from") || strings.Contains(keyLower, "sender") {
			return "noreply@example.com"
		}
		return "user@example.com"
	}

	// Booleans
	actualLower := strings.ToLower(actualValue)
	if actualLower == "true" || actualLower == "false" || actualLower == "1" || actualLower == "0" || actualLower == "yes" || actualLower == "no" {
		return "false"
	}

	// Numbers
	if _, err := strconv.Atoi(actualValue); err == nil {
		return "0"
	}

	// Default
	return "your-value-here"
}

// compareEnvFiles compares env and example files and finds discrepancies
func (v *EnvSyncValidator) compareEnvFiles(envVars, exampleVars map[string]EnvVariable) []Issue {
	var issues []Issue

	// Find variables in .env but not in .env.example
	for envVar := range envVars {
		if _, exists := exampleVars[envVar]; !exists {
			actualValue := envVars[envVar].Value
			suggestedValue := v.suggestSafeValue(envVar, actualValue)
			issues = append(issues, Issue{
				Type:       "missing_in_example",
				Variable:   envVar,
				Suggestion: fmt.Sprintf("%s=%s", envVar, suggestedValue),
				Severity:   "high",
			})
		}
	}

	// Find variables in .env.example but not in .env (informational)
	for exampleVar := range exampleVars {
		if _, exists := envVars[exampleVar]; !exists {
			issues = append(issues, Issue{
				Type:     "missing_in_env",
				Variable: exampleVar,
				Severity: "info",
			})
		}
	}

	// Check for exposed sensitive values in .env.example
	for varName, data := range exampleVars {
		value := data.Value
		if len(value) > 10 {
			// Check if it looks like a real secret
			for _, pattern := range v.secretPatterns {
				if pattern.MatchString(value) {
					issues = append(issues, Issue{
						Type:     "exposed_secret",
						Variable: varName,
						Line:     data.Line,
						Severity: "high",
					})
					break
				}
			}
		}
	}

	return issues
}

// extractEnvVarsFromContent extracts environment variable references from code
func (v *EnvSyncValidator) extractEnvVarsFromContent(content string) map[string]bool {
	envVars := make(map[string]bool)

	for _, pattern := range v.envVarPatterns {
		if pattern == nil {
			continue
		}
		matches := pattern.FindAllStringSubmatch(content, -1)
		for _, match := range matches {
			if len(match) >= 2 {
				envVars[match[1]] = true
			}
		}
	}

	return envVars
}

// processFileOperations handles file operations (Write, Edit, MultiEdit)
func (v *EnvSyncValidator) processFileOperations(toolName string, toolInput map[string]interface{}) {
	filePath, _ := toolInput["file_path"].(string)
	if filePath == "" {
		return
	}

	// Check if editing .env file
	if strings.Contains(filePath, ".env") && !strings.Contains(filePath, ".example") && !strings.Contains(filePath, ".sample") && !strings.Contains(filePath, ".template") {
		envFiles := v.findEnvFiles()

		if len(envFiles.Example) == 0 {
			fmt.Fprintf(os.Stderr, "\n⚠️  No .env.example file found!\n")
			fmt.Fprintf(os.Stderr, "   Create .env.example with safe placeholder values\n")
			fmt.Fprintf(os.Stderr, "   This helps other developers understand required variables\n")
		} else {
			fmt.Fprintf(os.Stderr, "\n📝 Remember to update .env.example if adding new variables\n")
			fmt.Fprintf(os.Stderr, "   Use safe placeholder values, never real secrets\n")
		}
		return
	}

	// Check for env var usage in code
	isCodeFile := false
	for _, ext := range v.fileExtensions {
		if strings.HasSuffix(filePath, ext) {
			isCodeFile = true
			break
		}
	}

	if !isCodeFile {
		return
	}

	var content string
	switch toolName {
	case "Write":
		content, _ = toolInput["content"].(string)
	case "Edit":
		content, _ = toolInput["new_string"].(string)
	case "MultiEdit":
		if edits, ok := toolInput["edits"].([]interface{}); ok {
			var parts []string
			for _, edit := range edits {
				if editMap, ok := edit.(map[string]interface{}); ok {
					if newString, ok := editMap["new_string"].(string); ok {
						parts = append(parts, newString)
					}
				}
			}
			content = strings.Join(parts, "\n")
		}
	}

	if content != "" {
		usedVars := v.extractEnvVarsFromContent(content)
		if len(usedVars) > 0 {
			// Check if these vars are documented
			envFiles := v.findEnvFiles()
			if len(envFiles.Example) > 0 {
				exampleVars := v.parseEnvFile(envFiles.Example[0])
				var missing []string
				for varName := range usedVars {
					if _, exists := exampleVars[varName]; !exists {
						missing = append(missing, varName)
					}
				}

				if len(missing) > 0 {
					fmt.Fprintf(os.Stderr, "\n📋 New environment variables detected: %s\n", strings.Join(missing, ", "))
					fmt.Fprintf(os.Stderr, "   Remember to add them to .env.example with placeholder values\n")
				}
			}
		}
	}
}

// processGitCommit handles git commit validation
func (v *EnvSyncValidator) processGitCommit() {
	// Check if any .env files were modified
	cmd := exec.Command("git", "diff", "--cached", "--name-only")
	cmd.Dir = v.projectRoot
	output, err := cmd.Output()
	if err != nil {
		return
	}

	files := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(files) == 0 || files[0] == "" {
		return
	}

	envModified := false
	for _, file := range files {
		if strings.Contains(file, ".env") && !strings.Contains(file, ".example") && !strings.Contains(file, ".sample") {
			envModified = true
			break
		}
	}

	if !envModified {
		return
	}

	// Find and compare env files
	envFiles := v.findEnvFiles()

	if len(envFiles.Example) == 0 {
		fmt.Fprintf(os.Stderr, "\n❌ Environment Configuration Error:\n")
		fmt.Fprintf(os.Stderr, "   No .env.example file found!\n")
		fmt.Fprintf(os.Stderr, "   Create .env.example to document required variables\n")
		os.Exit(2)
	}

	// Parse files concurrently
	var allIssues []FileComparison
	var wg sync.WaitGroup
	issuesChan := make(chan FileComparison, len(envFiles.Env)*len(envFiles.Example))

	for _, envFile := range envFiles.Env {
		for _, exampleFile := range envFiles.Example {
			wg.Add(1)
			go func(env, example string) {
				defer wg.Done()
				envVars := v.parseEnvFile(env)
				exampleVars := v.parseEnvFile(example)
				issues := v.compareEnvFiles(envVars, exampleVars)

				if len(issues) > 0 {
					issuesChan <- FileComparison{
						EnvFile:     env,
						ExampleFile: example,
						Issues:      issues,
					}
				}
			}(envFile, exampleFile)
		}
	}

	go func() {
		wg.Wait()
		close(issuesChan)
	}()

	for comparison := range issuesChan {
		allIssues = append(allIssues, comparison)
	}

	if len(allIssues) > 0 {
		fmt.Fprintf(os.Stderr, "\n🔧 Environment File Sync Issues:\n\n")

		blocking := false
		for _, fileData := range allIssues {
			var highIssues, infoIssues []Issue
			for _, issue := range fileData.Issues {
				if issue.Severity == "high" {
					highIssues = append(highIssues, issue)
				} else if issue.Severity == "info" {
					infoIssues = append(infoIssues, issue)
				}
			}

			if len(highIssues) > 0 {
				blocking = true
				fmt.Fprintf(os.Stderr, "❌ %s is missing variables:\n", fileData.ExampleFile)
				for _, issue := range highIssues {
					switch issue.Type {
					case "missing_in_example":
						fmt.Fprintf(os.Stderr, "   Add: %s\n", issue.Suggestion)
					case "exposed_secret":
						fmt.Fprintf(os.Stderr, "   Line %d: %s contains a real secret!\n", issue.Line, issue.Variable)
						fmt.Fprintf(os.Stderr, "   Replace with a placeholder value\n")
					}
				}
			}

			if len(infoIssues) > 0 && !blocking {
				var missingVars []string
				for _, issue := range infoIssues {
					if issue.Type == "missing_in_env" {
						missingVars = append(missingVars, issue.Variable)
					}
				}
				if len(missingVars) > 0 {
					fmt.Fprintf(os.Stderr, "\n💡 Optional: These variables are in .env.example but not .env:\n")
					fmt.Fprintf(os.Stderr, "   %s\n", strings.Join(missingVars, ", "))
				}
			}
		}

		if blocking {
			fmt.Fprintf(os.Stderr, "\n📝 Best Practices:\n")
			fmt.Fprintf(os.Stderr, "   • Keep .env.example updated with all required variables\n")
			fmt.Fprintf(os.Stderr, "   • Use descriptive placeholder values\n")
			fmt.Fprintf(os.Stderr, "   • Never commit real secrets to .env.example\n")
			fmt.Fprintf(os.Stderr, "   • Document variable purposes with comments\n")
			os.Exit(2)
		}
	}
}

// main function
func main() {
	start := time.Now()

	// Initialize validator
	validator = &EnvSyncValidator{
		projectRoot: ".",
	}
	if cwd, err := os.Getwd(); err == nil {
		validator.projectRoot = cwd
	}

	validator.initOnce.Do(func() {
		validator.init()
	})

	// Read input from stdin
	var input ToolInput
	if err := json.NewDecoder(os.Stdin).Decode(&input); err != nil {
		fmt.Fprintf(os.Stderr, "Error: Invalid JSON input: %v\n", err)
		os.Exit(1)
	}

	// Process based on tool type
	switch input.ToolName {
	case "Write", "Edit", "MultiEdit":
		validator.processFileOperations(input.ToolName, input.ToolInput)
	case "Bash":
		if command, ok := input.ToolInput["command"].(string); ok {
			if strings.Contains(command, "git commit") {
				validator.processGitCommit()
			}
		}
	}

	// Log performance info to stderr
	duration := time.Since(start)
	fmt.Fprintf(os.Stderr, "Env sync validator completed in %v\n", duration)
}