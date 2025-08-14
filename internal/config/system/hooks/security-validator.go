//go:build ignore

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

	"claude-hooks-orchestrator/patterns"
)

// SecurityPatterns represents the compiled security patterns
type SecurityPatterns struct {
	Secrets           map[string]PatternCategory `json:"secrets"`
	APIDecSecurity    PatternCategory            `json:"api_security"`
	DatabaseSchema    map[string]PatternCategory `json:"database_schema"`
	AllowedPatterns   AllowedPatterns            `json:"allowed_patterns"`
	FileExtensions    FileExtensions             `json:"file_extensions"`
	GitignoreRequired GitignorePatterns          `json:"gitignore_required"`
}

type PatternCategory struct {
	Patterns    []string `json:"patterns"`
	Severity    string   `json:"severity"`
	Description string   `json:"description"`
}

type AllowedPatterns struct {
	FalsePositives    []string `json:"false_positives"`
	TestIndicators    []string `json:"test_indicators"`
	CommentIndicators []string `json:"comment_indicators"`
}

type FileExtensions struct {
	SkipBinary    []string `json:"skip_binary"`
	SkipGenerated []string `json:"skip_generated"`
	ScanCode      []string `json:"scan_code"`
	ScanConfig    []string `json:"scan_config"`
	ScanDatabase  []string `json:"scan_database"`
}

type GitignorePatterns struct {
	Patterns        []string            `json:"patterns"`
	BroaderPatterns map[string][]string `json:"broader_patterns"`
}

// Issue represents a security issue found
type Issue struct {
	Type       string `json:"type"`
	Category   string `json:"category"`
	File       string `json:"file,omitempty"`
	Line       int    `json:"line,omitempty"`
	Message    string `json:"message"`
	Severity   string `json:"severity"`
	Pattern    string `json:"pattern,omitempty"`
	Value      string `json:"value,omitempty"`
	Suggestion string `json:"suggestion,omitempty"`
}

// SecurityValidator handles all security validation
type SecurityValidator struct {
	patterns   *SecurityPatterns
	compiled   map[string]*regexp.Regexp
	hooksDir   string
	workingDir string
}

// NewSecurityValidator creates a new security validator
func NewSecurityValidator(hooksDir string) (*SecurityValidator, error) {
	validator := &SecurityValidator{
		hooksDir:   hooksDir,
		compiled:   make(map[string]*regexp.Regexp),
		workingDir: hooksDir,
	}

	// Load patterns
	patternsFile := filepath.Join(hooksDir, "patterns", "security-patterns.json")
	data, err := ioutil.ReadFile(patternsFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load security patterns: %v", err)
	}

	if err := json.Unmarshal(data, &validator.patterns); err != nil {
		return nil, fmt.Errorf("failed to parse security patterns: %v", err)
	}

	// Pre-compile regex patterns for performance
	validator.compilePatterns()

	return validator, nil
}

// compilePatterns pre-compiles all regex patterns
func (sv *SecurityValidator) compilePatterns() {
	// Compile secret patterns
	for category, patternData := range sv.patterns.Secrets {
		for i, pattern := range patternData.Patterns {
			key := fmt.Sprintf("secrets_%s_%d", category, i)
			if compiled, err := regexp.Compile(pattern); err == nil {
				sv.compiled[key] = compiled
			}
		}
	}

	// Compile API security patterns
	for i, pattern := range sv.patterns.APIDecSecurity.Patterns {
		key := fmt.Sprintf("api_auth_%d", i)
		if compiled, err := regexp.Compile(pattern); err == nil {
			sv.compiled[key] = compiled
		}
	}

	// Compile database patterns
	for category, patternData := range sv.patterns.DatabaseSchema {
		for i, pattern := range patternData.Patterns {
			key := fmt.Sprintf("db_%s_%d", category, i)
			if compiled, err := regexp.Compile(pattern); err == nil {
				sv.compiled[key] = compiled
			}
		}
	}

	// Compile allowed patterns
	for i, pattern := range sv.patterns.AllowedPatterns.FalsePositives {
		key := fmt.Sprintf("allowed_%d", i)
		if compiled, err := regexp.Compile(pattern); err == nil {
			sv.compiled[key] = compiled
		}
	}
}

// ValidateGitCommit validates files being committed
func (sv *SecurityValidator) ValidateGitCommit() ([]Issue, error) {
	var issues []Issue

	// Get staged files
	cmd := exec.Command("git", "diff", "--cached", "--name-only")
	output, err := cmd.Output()
	if err != nil {
		return issues, fmt.Errorf("failed to get staged files: %v", err)
	}

	files := strings.Fields(string(output))
	if len(files) == 0 {
		return issues, nil
	}

	// Check for .env files being committed
	envIssues := sv.checkEnvFilesCommit(files)
	issues = append(issues, envIssues...)

	// Check .gitignore
	gitignoreIssues := sv.checkGitignore()
	issues = append(issues, gitignoreIssues...)

	// Scan each file
	for _, file := range files {
		if sv.shouldSkipFile(file) {
			continue
		}

		fileIssues, err := sv.scanFile(file)
		if err != nil {
			continue // Skip files that can't be read
		}
		issues = append(issues, fileIssues...)
	}

	return issues, nil
}

// ValidateFileContent validates content being written/edited
func (sv *SecurityValidator) ValidateFileContent(filePath, content string) []Issue {
	var issues []Issue

	// Warn about .env files
	if strings.Contains(filePath, ".env") && !strings.HasSuffix(filePath, ".example") {
		issues = append(issues, Issue{
			Type:     "env_file_edit",
			Category: "warning",
			File:     filePath,
			Message:  "Editing .env file - ensure it's in .gitignore",
			Severity: "medium",
		})
	}

	// Scan content for secrets
	secretIssues := sv.scanContentForSecrets(content, filePath)
	issues = append(issues, secretIssues...)

	// Check for API endpoints
	apiIssues := sv.scanContentForAPIEndpoints(content, filePath)
	issues = append(issues, apiIssues...)

	// Check for database schema changes
	dbIssues := sv.scanContentForDatabaseChanges(content, filePath)
	issues = append(issues, dbIssues...)

	return issues
}

// scanFile scans a file for security issues
func (sv *SecurityValidator) scanFile(filePath string) ([]Issue, error) {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// Skip binary files
	if sv.isBinaryContent(content) {
		return nil, nil
	}

	return sv.ValidateFileContent(filePath, string(content)), nil
}

// scanContentForSecrets scans content for potential secrets
func (sv *SecurityValidator) scanContentForSecrets(content, filePath string) []Issue {
	var issues []Issue

	for category, patternData := range sv.patterns.Secrets {
		for i := range patternData.Patterns {
			key := fmt.Sprintf("secrets_%s_%d", category, i)
			compiled, exists := sv.compiled[key]
			if !exists {
				continue
			}

			matches := compiled.FindAllStringIndex(content, -1)
			for _, match := range matches {
				if sv.isAllowedContext(content, match[0], match[1]) {
					continue
				}

				lineNum := sv.getLineNumber(content, match[0])
				secretValue := content[match[0]:match[1]]
				redacted := sv.redactSecret(secretValue)

				issues = append(issues, Issue{
					Type:     "secret_detected",
					Category: category,
					File:     filePath,
					Line:     lineNum,
					Message:  fmt.Sprintf("Potential %s detected", strings.ReplaceAll(category, "_", " ")),
					Severity: patternData.Severity,
					Pattern:  sv.truncatePattern(patternData.Patterns[i]),
					Value:    redacted,
				})
			}
		}
	}

	return issues
}

// scanContentForAPIEndpoints scans for API endpoint definitions
func (sv *SecurityValidator) scanContentForAPIEndpoints(content, filePath string) []Issue {
	var issues []Issue

	// Skip if not an API file
	if !sv.isAPIFile(filePath) {
		return issues
	}

	endpoints := sv.extractAPIEndpoints(content, filePath)
	for _, endpoint := range endpoints {
		// Check security for internal endpoints
		securityIssues := sv.checkAPIEndpointSecurity(content, endpoint)
		issues = append(issues, securityIssues...)
	}

	return issues
}

// scanContentForDatabaseChanges scans for database schema changes
func (sv *SecurityValidator) scanContentForDatabaseChanges(content, filePath string) []Issue {
	var issues []Issue

	// Check if it's a database file
	if !sv.isDatabaseFile(filePath) {
		return issues
	}

	// Check for new table creation
	if strings.Contains(filePath, "schema.prisma") {
		issues = append(issues, sv.checkPrismaChanges(content, filePath)...)
	} else if strings.Contains(filePath, ".sql") {
		issues = append(issues, sv.checkSQLChanges(content)...)
	}

	return issues
}

// Helper methods

func (sv *SecurityValidator) shouldSkipFile(filePath string) bool {
	// Skip binary files
	for _, ext := range sv.patterns.FileExtensions.SkipBinary {
		if strings.HasSuffix(filePath, ext) {
			return true
		}
	}

	// Skip generated files
	for _, ext := range sv.patterns.FileExtensions.SkipGenerated {
		if strings.HasSuffix(filePath, ext) {
			return true
		}
	}

	return false
}

func (sv *SecurityValidator) isBinaryContent(content []byte) bool {
	// Simple binary detection
	for i := 0; i < len(content) && i < 512; i++ {
		if content[i] == 0 {
			return true
		}
	}
	return false
}

func (sv *SecurityValidator) isAllowedContext(content string, start, end int) bool {
	// Get surrounding context
	contextStart := start - 100
	if contextStart < 0 {
		contextStart = 0
	}
	contextEnd := end + 100
	if contextEnd > len(content) {
		contextEnd = len(content)
	}
	context := content[contextStart:contextEnd]

	// Check against allowed patterns
	for i := range sv.patterns.AllowedPatterns.FalsePositives {
		key := fmt.Sprintf("allowed_%d", i)
		if compiled, exists := sv.compiled[key]; exists {
			if compiled.MatchString(context) {
				return true
			}
		}
	}

	// Check if it's in a comment
	lineStart := strings.LastIndex(content[:start], "\n")
	if lineStart == -1 {
		lineStart = 0
	}
	lineContent := content[lineStart:start]
	for _, indicator := range sv.patterns.AllowedPatterns.CommentIndicators {
		if strings.Contains(lineContent, indicator) {
			return true
		}
	}

	// Check if it's in test context
	for _, indicator := range sv.patterns.AllowedPatterns.TestIndicators {
		if strings.Contains(context, indicator) {
			return true
		}
	}

	return false
}

func (sv *SecurityValidator) getLineNumber(content string, position int) int {
	return strings.Count(content[:position], "\n") + 1
}

func (sv *SecurityValidator) redactSecret(secret string) string {
	if len(secret) > 20 {
		return secret[:10] + "...[REDACTED]"
	}
	return "[REDACTED]"
}

func (sv *SecurityValidator) truncatePattern(pattern string) string {
	if len(pattern) > 30 {
		return pattern[:30] + "..."
	}
	return pattern
}

func (sv *SecurityValidator) isAPIFile(filePath string) bool {
	apiPatterns := []string{"/api/", "route.ts", "route.js", "controller.", ".route."}
	for _, pattern := range apiPatterns {
		if strings.Contains(filePath, pattern) {
			return true
		}
	}
	return false
}

func (sv *SecurityValidator) isDatabaseFile(filePath string) bool {
	dbPatterns := []string{"schema.prisma", ".sql", "migration", "database", "db"}
	fileLower := strings.ToLower(filePath)
	for _, pattern := range dbPatterns {
		if strings.Contains(fileLower, pattern) {
			return true
		}
	}
	return false
}

func (sv *SecurityValidator) extractAPIEndpoints(content, filePath string) []map[string]string {
	var endpoints []map[string]string

	// Next.js App Router pattern
	if strings.Contains(filePath, "route.ts") || strings.Contains(filePath, "route.js") {
		methodPattern := patterns.GetPatterns().APINextJS
		methods := methodPattern.FindAllStringSubmatch(content, -1)

		routePattern := patterns.GetPatterns().APIRoute
		routeMatch := routePattern.FindStringSubmatch(filePath)

		if routeMatch != nil && len(methods) > 0 {
			route := strings.Replace(routeMatch[1], "/route.ts", "", 1)
			route = strings.Replace(route, "/route.js", "", 1)
			for _, method := range methods {
				endpoints = append(endpoints, map[string]string{
					"path":   route,
					"method": method[1],
					"file":   filePath,
				})
			}
		}
	}

	// Express/FastAPI style routes
	routePatterns := []*regexp.Regexp{
		patterns.GetPatterns().APIExpress,
		patterns.GetPatterns().APIExpress, // Reuse for app pattern
		patterns.GetPatterns().APIFastAPI,
	}

	for _, pattern := range routePatterns {
		matches := pattern.FindAllStringSubmatch(content, -1)
		for _, match := range matches {
			if len(match) >= 3 {
				endpoints = append(endpoints, map[string]string{
					"path":   match[2],
					"method": strings.ToUpper(match[1]),
					"file":   filePath,
				})
			}
		}
	}

	return endpoints
}

func (sv *SecurityValidator) checkAPIEndpointSecurity(content string, endpoint map[string]string) []Issue {
	var issues []Issue
	path := endpoint["path"]

	// Check if it's an internal API
	internalPatterns := []string{"/internal/", "/admin/", "/system/", "/private/"}
	isInternal := false
	for _, pattern := range internalPatterns {
		if strings.Contains(path, pattern) {
			isInternal = true
			break
		}
	}

	if !isInternal {
		return issues
	}

	// Check for authentication
	authPatterns := sv.patterns.APIDecSecurity.Patterns
	hasAuth := false
	for _, authPattern := range authPatterns {
		if matched, _ := regexp.MatchString(authPattern, content); matched {
			hasAuth = true
			break
		}
	}

	if !hasAuth {
		issues = append(issues, Issue{
			Type:       "missing_auth",
			Category:   "api_security",
			File:       endpoint["file"],
			Message:    fmt.Sprintf("Internal API endpoint %s %s lacks authentication", endpoint["method"], path),
			Severity:   "high",
			Suggestion: "Add API key validation or authentication middleware",
		})
	}

	return issues
}

func (sv *SecurityValidator) checkPrismaChanges(content, filePath string) []Issue {
	var issues []Issue

	// Find new models
	modelPattern := patterns.GetPatterns().DBModel
	newModels := modelPattern.FindAllStringSubmatch(content, -1)

	// Try to read existing file to compare
	existingModels := make(map[string]bool)
	if existingContent, err := ioutil.ReadFile(filePath); err == nil {
		existingMatches := modelPattern.FindAllStringSubmatch(string(existingContent), -1)
		for _, match := range existingMatches {
			if len(match) > 1 {
				existingModels[match[1]] = true
			}
		}
	}

	// Check for newly added models
	for _, match := range newModels {
		if len(match) > 1 {
			modelName := match[1]
			if !existingModels[modelName] {
				issues = append(issues, Issue{
					Type:       "new_table",
					Category:   "database_schema",
					File:       filePath,
					Message:    fmt.Sprintf("New database model detected: %s", modelName),
					Severity:   "high",
					Suggestion: "Consider extending existing tables instead of creating new ones",
				})
			}
		}
	}

	return issues
}

func (sv *SecurityValidator) checkSQLChanges(content string) []Issue {
	var issues []Issue

	createTablePattern := patterns.GetPatterns().DBCreateTable
	tables := createTablePattern.FindAllStringSubmatch(content, -1)

	if len(tables) > 0 {
		var tableNames []string
		for _, table := range tables {
			if len(table) > 1 {
				tableNames = append(tableNames, table[1])
			}
		}
		issues = append(issues, Issue{
			Type:       "new_table",
			Category:   "database_schema",
			Message:    fmt.Sprintf("New tables detected in SQL: %s", strings.Join(tableNames, ", ")),
			Severity:   "high",
			Suggestion: "Consider extending existing tables instead of creating new ones",
		})
	}

	return issues
}

func (sv *SecurityValidator) checkEnvFilesCommit(files []string) []Issue {
	var issues []Issue
	var envFiles []string

	for _, file := range files {
		if strings.Contains(file, ".env") && !strings.HasSuffix(file, ".example") && !strings.HasSuffix(file, ".sample") {
			envFiles = append(envFiles, file)
		}
	}

	if len(envFiles) > 0 {
		issues = append(issues, Issue{
			Type:       "env_file_commit",
			Category:   "critical",
			Message:    fmt.Sprintf("Attempting to commit .env files: %s", strings.Join(envFiles, ", ")),
			Severity:   "critical",
			Suggestion: "Add these files to .gitignore immediately",
		})
	}

	return issues
}

func (sv *SecurityValidator) checkGitignore() []Issue {
	var issues []Issue
	gitignorePath := ".gitignore"

	if _, err := os.Stat(gitignorePath); err != nil {
		// .gitignore doesn't exist
		issues = append(issues, Issue{
			Type:       "missing_gitignore",
			Category:   "configuration",
			Message:    ".gitignore file not found",
			Severity:   "medium",
			Suggestion: "Create .gitignore to protect sensitive files",
		})
		return issues
	}

	content, err := ioutil.ReadFile(gitignorePath)
	if err != nil {
		return issues
	}

	gitignoreContent := string(content)
	var missing []string

	for _, pattern := range sv.patterns.GitignoreRequired.Patterns {
		if !strings.Contains(gitignoreContent, pattern) {
			// Check if a broader pattern covers it
			covered := false
			for broaderPattern, coveredPatterns := range sv.patterns.GitignoreRequired.BroaderPatterns {
				if strings.Contains(gitignoreContent, broaderPattern) {
					for _, coveredPattern := range coveredPatterns {
						if coveredPattern == pattern {
							covered = true
							break
						}
					}
				}
				if covered {
					break
				}
			}
			if !covered {
				missing = append(missing, pattern)
			}
		}
	}

	if len(missing) > 0 {
		issues = append(issues, Issue{
			Type:       "gitignore_missing",
			Category:   "configuration",
			Message:    fmt.Sprintf("Consider adding to .gitignore: %s", strings.Join(missing, ", ")),
			Severity:   "medium",
			Suggestion: "Add these patterns to prevent committing sensitive files",
		})
	}

	return issues
}

// Output formatting

func (sv *SecurityValidator) PrintIssues(issues []Issue, context string) {
	if len(issues) == 0 {
		return
	}

	fmt.Fprintf(os.Stderr, "\n🔒 Security Validation Results (%s):\n\n", context)

	// Group issues by severity
	critical := []Issue{}
	high := []Issue{}
	medium := []Issue{}

	for _, issue := range issues {
		switch issue.Severity {
		case "critical":
			critical = append(critical, issue)
		case "high":
			high = append(high, issue)
		case "medium":
			medium = append(medium, issue)
		}
	}

	blocking := len(critical) > 0 || len(high) > 0

	// Print critical issues
	if len(critical) > 0 {
		fmt.Fprintf(os.Stderr, "🚨 CRITICAL ISSUES (blocking):\n")
		for _, issue := range critical {
			sv.printIssue(issue)
		}
	}

	// Print high severity issues
	if len(high) > 0 {
		fmt.Fprintf(os.Stderr, "❌ HIGH SEVERITY ISSUES (blocking):\n")
		for _, issue := range high {
			sv.printIssue(issue)
		}
	}

	// Print medium severity issues
	if len(medium) > 0 {
		fmt.Fprintf(os.Stderr, "⚠️  MEDIUM SEVERITY ISSUES (warnings):\n")
		for _, issue := range medium {
			sv.printIssue(issue)
		}
	}

	// Print security best practices
	fmt.Fprintf(os.Stderr, "\n💡 Security Best Practices:\n")
	fmt.Fprintf(os.Stderr, "   • Use environment variables for sensitive data\n")
	fmt.Fprintf(os.Stderr, "   • Never commit .env files (add to .gitignore)\n")
	fmt.Fprintf(os.Stderr, "   • Add authentication to internal APIs\n")
	fmt.Fprintf(os.Stderr, "   • Extend existing tables instead of creating new ones\n")
	fmt.Fprintf(os.Stderr, "   • Use secret management services in production\n")

	if blocking {
		fmt.Fprintf(os.Stderr, "\n❌ Validation failed due to security issues\n")
		os.Exit(2)
	}
}

func (sv *SecurityValidator) printIssue(issue Issue) {
	if issue.File != "" && issue.Line > 0 {
		fmt.Fprintf(os.Stderr, "   📄 %s:%d\n", issue.File, issue.Line)
	} else if issue.File != "" {
		fmt.Fprintf(os.Stderr, "   📄 %s\n", issue.File)
	}

	fmt.Fprintf(os.Stderr, "      %s\n", issue.Message)

	if issue.Value != "" {
		fmt.Fprintf(os.Stderr, "      Found: %s\n", issue.Value)
	}

	if issue.Suggestion != "" {
		fmt.Fprintf(os.Stderr, "      💡 %s\n", issue.Suggestion)
	}

	fmt.Fprintf(os.Stderr, "\n")
}

// isSecurityRelevantFile checks if a file type is relevant for security validation
func (sv *SecurityValidator) isSecurityRelevantFile(filePath string) bool {
	if filePath == "" {
		return false
	}

	ext := strings.ToLower(filepath.Ext(filePath))
	filename := strings.ToLower(filepath.Base(filePath))

	// Security-relevant file extensions
	securityExtensions := []string{
		".py", ".js", ".ts", ".tsx", ".jsx", ".go", ".java", ".php", ".rb", ".cs", ".cpp", ".c", ".h",
		".env", ".yml", ".yaml", ".json", ".xml", ".toml", ".cfg", ".ini", ".conf", ".config",
		".sql", ".prisma", ".graphql", ".proto", ".sh", ".bash", ".zsh", ".ps1", ".cmd", ".bat",
		".tf", ".tfvars", ".hcl", ".dockerfile", ".dockerignore", ".gitignore", ".gitattributes",
		".pem", ".key", ".crt", ".cert", ".p12", ".pfx", ".jks", ".keystore", ".properties",
	}

	// Security-relevant filenames
	securityFilenames := []string{
		"dockerfile", "makefile", "cmakelists.txt", "package.json", "requirements.txt",
		"pipfile", "composer.json", "pom.xml", "build.gradle", "cargo.toml", "go.mod",
		"secrets", "config", "settings", "database", "schema", "migration", "seed",
		"auth", "login", "password", "token", "key", "cert", "ssl", "tls", "security",
	}

	// Check extension
	for _, secExt := range securityExtensions {
		if ext == secExt {
			return true
		}
	}

	// Check filename contains security-relevant keywords
	for _, secName := range securityFilenames {
		if strings.Contains(filename, secName) {
			return true
		}
	}

	// Check if path contains security-relevant directories
	lowerPath := strings.ToLower(filePath)
	securityDirs := []string{
		"auth", "security", "config", "settings", "secrets", "keys", "certs", "ssl",
		"database", "db", "migration", "seed", "sql", "prisma", "graphql", "api",
		"middleware", "guard", "filter", "interceptor", "validation", "sanitization",
	}

	for _, secDir := range securityDirs {
		if strings.Contains(lowerPath, "/"+secDir+"/") || strings.Contains(lowerPath, "\\"+secDir+"\\") {
			return true
		}
	}

	return false
}

// CLI interface and main function

type ToolInput struct {
	ToolName  string                 `json:"tool_name"`
	ToolInput map[string]interface{} `json:"tool_input"`
}

func main() {
	hooksDir := "/Users/a.pezzotta/.claude/hooks"
	if len(os.Args) > 1 {
		hooksDir = os.Args[1]
	}

	validator, err := NewSecurityValidator(hooksDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing security validator: %v\n", err)
		os.Exit(1)
	}

	// Read input from stdin
	var input ToolInput
	decoder := json.NewDecoder(os.Stdin)
	if err := decoder.Decode(&input); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
		os.Exit(1)
	}

	var issues []Issue

	switch input.ToolName {
	case "Bash":
		command, ok := input.ToolInput["command"].(string)
		if ok && (strings.Contains(command, "git commit") || strings.Contains(command, "git add")) {
			issues, err = validator.ValidateGitCommit()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error validating git commit: %v\n", err)
				os.Exit(1)
			}
			validator.PrintIssues(issues, "Git Commit")
		}

	case "Write", "Edit", "MultiEdit":
		filePath, _ := input.ToolInput["file_path"].(string)

		// Early exit: Skip if no file path
		if filePath == "" {
			os.Exit(0)
		}

		// Early exit: Skip if not a security-relevant file type
		if !validator.isSecurityRelevantFile(filePath) {
			os.Exit(0)
		}

		var content string

		switch input.ToolName {
		case "Write":
			content, _ = input.ToolInput["content"].(string)
		case "Edit":
			content, _ = input.ToolInput["new_string"].(string)
		case "MultiEdit":
			edits, ok := input.ToolInput["edits"].([]interface{})
			if ok {
				var contentParts []string
				for _, edit := range edits {
					if editMap, ok := edit.(map[string]interface{}); ok {
						if newString, ok := editMap["new_string"].(string); ok {
							contentParts = append(contentParts, newString)
						}
					}
				}
				content = strings.Join(contentParts, "\n")
			}
		}

		// Early exit: Skip if no content to validate
		if content == "" {
			os.Exit(0)
		}

		issues = validator.ValidateFileContent(filePath, content)
		validator.PrintIssues(issues, "File Edit")
	}

	os.Exit(0)
}
