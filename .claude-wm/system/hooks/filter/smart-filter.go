package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"claude-hooks-orchestrator/patterns"
)

// SmartFilter represents the smart filtering system
type SmartFilter struct {
	config           FilterConfig
	gitChangedFiles  []string
	toolName         string
	toolInput        map[string]interface{}
	compiledPatterns map[string]*regexp.Regexp
	mu               sync.RWMutex
}

// FilterConfig represents the configuration for smart filtering
type FilterConfig struct {
	FileTypeTriggers map[string][]string `json:"file_type_triggers"`
	HookTriggers     map[string]struct {
		FilePatterns []string `json:"file_patterns"`
		ToolNames    []string `json:"tool_names"`
		AlwaysRun    bool     `json:"always_run"`
		Description  string   `json:"description"`
	} `json:"hook_triggers"`
	OptimizationSettings struct {
		EnableGitDiffAnalysis bool `json:"enable_git_diff_analysis"`
		EnableToolBasedFilter bool `json:"enable_tool_based_filter"`
		EnableFileTypeFilter  bool `json:"enable_file_type_filter"`
		MaxFilesForFullScan   int  `json:"max_files_for_full_scan"`
	} `json:"optimization_settings"`
}

// NewSmartFilter creates a new smart filter instance
func NewSmartFilter(configPath string) (*SmartFilter, error) {
	sf := &SmartFilter{
		compiledPatterns: make(map[string]*regexp.Regexp),
	}

	// Load configuration
	if err := sf.loadConfig(configPath); err != nil {
		return nil, fmt.Errorf("failed to load config: %v", err)
	}

	// Compile regex patterns
	if err := sf.compilePatterns(); err != nil {
		return nil, fmt.Errorf("failed to compile patterns: %v", err)
	}

	return sf, nil
}

// loadConfig loads the filter configuration from file
func (sf *SmartFilter) loadConfig(configPath string) error {
	// Try to load from hook-triggers.json first
	triggerConfigPath := filepath.Join(filepath.Dir(configPath), "hook-triggers.json")
	if _, err := os.Stat(triggerConfigPath); err == nil {
		configData, err := ioutil.ReadFile(triggerConfigPath)
		if err != nil {
			return fmt.Errorf("failed to read trigger config file: %v", err)
		}
		if err := json.Unmarshal(configData, &sf.config); err != nil {
			return fmt.Errorf("failed to parse trigger config: %v", err)
		}
	} else {
		// Fallback to parallel-groups.json for basic file type triggers
		configData, err := ioutil.ReadFile(configPath)
		if err != nil {
			return fmt.Errorf("failed to read config file: %v", err)
		}
		
		var parallelConfig struct {
			FileTypeTriggers map[string][]string `json:"file_type_triggers"`
		}
		if err := json.Unmarshal(configData, &parallelConfig); err != nil {
			return fmt.Errorf("failed to parse parallel config: %v", err)
		}
		
		sf.config.FileTypeTriggers = parallelConfig.FileTypeTriggers
		
		// Set default optimization settings
		sf.config.OptimizationSettings.EnableGitDiffAnalysis = true
		sf.config.OptimizationSettings.EnableToolBasedFilter = true
		sf.config.OptimizationSettings.EnableFileTypeFilter = true
		sf.config.OptimizationSettings.MaxFilesForFullScan = 100
	}

	return nil
}

// compilePatterns compiles all regex patterns for performance
func (sf *SmartFilter) compilePatterns() error {
	sf.mu.Lock()
	defer sf.mu.Unlock()

	// Compile file type patterns
	for pattern := range sf.config.FileTypeTriggers {
		if !strings.Contains(pattern, "*") && !strings.Contains(pattern, "|") {
			continue // Skip non-pattern entries like "Bash", "Write", etc.
		}
		
		// Convert glob pattern to regex
		regexPattern := sf.globToRegex(pattern)
		compiled, err := patterns.CompilePattern(regexPattern)
		if err != nil {
			return fmt.Errorf("failed to compile pattern %s: %v", pattern, err)
		}
		sf.compiledPatterns[pattern] = compiled
	}

	// Compile hook-specific patterns
	for hookName, trigger := range sf.config.HookTriggers {
		for _, pattern := range trigger.FilePatterns {
			if _, exists := sf.compiledPatterns[pattern]; !exists {
				regexPattern := sf.globToRegex(pattern)
				compiled, err := patterns.CompilePattern(regexPattern)
				if err != nil {
					return fmt.Errorf("failed to compile pattern %s for hook %s: %v", pattern, hookName, err)
				}
				sf.compiledPatterns[pattern] = compiled
			}
		}
	}

	return nil
}

// globToRegex converts glob patterns to regex
func (sf *SmartFilter) globToRegex(pattern string) string {
	// Handle multiple patterns separated by |
	if strings.Contains(pattern, "|") {
		patterns := strings.Split(pattern, "|")
		var regexParts []string
		for _, p := range patterns {
			regexParts = append(regexParts, sf.singleGlobToRegex(strings.TrimSpace(p)))
		}
		return "(" + strings.Join(regexParts, "|") + ")"
	}
	
	return sf.singleGlobToRegex(pattern)
}

// singleGlobToRegex converts a single glob pattern to regex
func (sf *SmartFilter) singleGlobToRegex(pattern string) string {
	pattern = strings.ReplaceAll(pattern, ".", "\\.")
	pattern = strings.ReplaceAll(pattern, "*", ".*")
	pattern = strings.ReplaceAll(pattern, "?", ".")
	return "^" + pattern + "$"
}

// FilterHooks filters hooks based on changed files and tool context
func (sf *SmartFilter) FilterHooks(toolName string, toolInput map[string]interface{}, hookGroups map[string][]string) (map[string][]string, error) {
	sf.toolName = toolName
	sf.toolInput = toolInput
	
	// Get changed files if git diff analysis is enabled
	if sf.config.OptimizationSettings.EnableGitDiffAnalysis {
		if err := sf.analyzeGitChanges(); err != nil {
			fmt.Printf("⚠️  Git diff analysis failed: %v\n", err)
			// Continue without git analysis
		}
	}

	filteredGroups := make(map[string][]string)
	
	for groupName, hooks := range hookGroups {
		var filteredHooks []string
		
		for _, hook := range hooks {
			if sf.shouldRunHook(hook) {
				filteredHooks = append(filteredHooks, hook)
			}
		}
		
		if len(filteredHooks) > 0 {
			filteredGroups[groupName] = filteredHooks
		}
	}

	// Log filtering results
	sf.logFilteringResults(hookGroups, filteredGroups)

	return filteredGroups, nil
}

// shouldRunHook determines if a hook should run based on filtering criteria
func (sf *SmartFilter) shouldRunHook(hookName string) bool {
	// Check if hook should always run
	if trigger, exists := sf.config.HookTriggers[hookName]; exists && trigger.AlwaysRun {
		return true
	}

	// Check tool-based triggers
	if sf.config.OptimizationSettings.EnableToolBasedFilter {
		if sf.matchesToolTrigger(hookName) {
			return true
		}
	}

	// Check file-based triggers
	if sf.config.OptimizationSettings.EnableFileTypeFilter {
		if sf.matchesFileTrigger(hookName) {
			return true
		}
	}

	// Check hook-specific triggers
	if trigger, exists := sf.config.HookTriggers[hookName]; exists {
		if sf.matchesHookTrigger(hookName, trigger) {
			return true
		}
	}

	// Default to running conservative hooks if no specific criteria match
	return sf.isConservativeHook(hookName)
}

// matchesToolTrigger checks if hook matches the current tool
func (sf *SmartFilter) matchesToolTrigger(hookName string) bool {
	// Check file type triggers for tool names
	if hooks, exists := sf.config.FileTypeTriggers[sf.toolName]; exists {
		for _, hook := range hooks {
			if hook == hookName {
				return true
			}
		}
	}

	// Check hook-specific tool triggers
	if trigger, exists := sf.config.HookTriggers[hookName]; exists {
		for _, toolName := range trigger.ToolNames {
			if toolName == sf.toolName {
				return true
			}
		}
	}

	return false
}

// matchesFileTrigger checks if hook matches changed files
func (sf *SmartFilter) matchesFileTrigger(hookName string) bool {
	// If no git changes detected, be conservative and run most hooks
	if len(sf.gitChangedFiles) == 0 {
		return sf.isConservativeHook(hookName)
	}

	// If too many files changed, run full scan
	if len(sf.gitChangedFiles) > sf.config.OptimizationSettings.MaxFilesForFullScan {
		return true
	}

	// Check file type triggers
	for pattern, hooks := range sf.config.FileTypeTriggers {
		if sf.containsHook(hooks, hookName) {
			if sf.matchesFilePattern(pattern) {
				return true
			}
		}
	}

	return false
}

// matchesHookTrigger checks if hook matches its specific triggers
func (sf *SmartFilter) matchesHookTrigger(hookName string, trigger struct {
	FilePatterns []string `json:"file_patterns"`
	ToolNames    []string `json:"tool_names"`
	AlwaysRun    bool     `json:"always_run"`
	Description  string   `json:"description"`
}) bool {
	// Check file patterns
	for _, pattern := range trigger.FilePatterns {
		if sf.matchesFilePattern(pattern) {
			return true
		}
	}

	// Check tool names
	for _, toolName := range trigger.ToolNames {
		if toolName == sf.toolName {
			return true
		}
	}

	return false
}

// matchesFilePattern checks if any changed file matches the pattern
func (sf *SmartFilter) matchesFilePattern(pattern string) bool {
	sf.mu.RLock()
	compiled, exists := sf.compiledPatterns[pattern]
	sf.mu.RUnlock()

	if !exists {
		// Fallback to string matching for non-regex patterns
		return sf.matchesStringPattern(pattern)
	}

	for _, file := range sf.gitChangedFiles {
		if compiled.MatchString(file) {
			return true
		}
	}

	return false
}

// matchesStringPattern handles non-regex pattern matching
func (sf *SmartFilter) matchesStringPattern(pattern string) bool {
	// Handle special patterns
	switch pattern {
	case "Bash":
		return sf.toolName == "Bash"
	case "Write", "Edit", "MultiEdit":
		return sf.toolName == pattern
	case "Write|Edit|MultiEdit":
		return sf.toolName == "Write" || sf.toolName == "Edit" || sf.toolName == "MultiEdit"
	case "PostToolUse":
		return sf.toolName == "PostToolUse"
	case "*":
		return true
	}

	// Handle file extension patterns
	if strings.HasPrefix(pattern, "*.") {
		ext := strings.TrimPrefix(pattern, "*")
		for _, file := range sf.gitChangedFiles {
			if strings.HasSuffix(file, ext) {
				return true
			}
		}
	}

	return false
}

// containsHook checks if hook is in the hooks slice
func (sf *SmartFilter) containsHook(hooks []string, hookName string) bool {
	for _, hook := range hooks {
		if hook == hookName {
			return true
		}
	}
	return false
}

// isConservativeHook determines if hook should run when no git changes are detected
func (sf *SmartFilter) isConservativeHook(hookName string) bool {
	// Always run critical hooks
	conservativeHooks := []string{
		"security-validator",
		"git-validator-optimized",
		"timestamp-validator.py",
		"pre-commit-validator.py",
	}

	for _, conservative := range conservativeHooks {
		if conservative == hookName {
			return true
		}
	}

	return false
}

// analyzeGitChanges analyzes git diff to determine changed files
func (sf *SmartFilter) analyzeGitChanges() error {
	// Try multiple git commands to get changed files
	commands := [][]string{
		{"git", "diff", "--name-only", "HEAD"},
		{"git", "diff", "--cached", "--name-only"},
		{"git", "status", "--porcelain"},
	}

	for _, cmdArgs := range commands {
		cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
		output, err := cmd.CombinedOutput()
		if err == nil && len(output) > 0 {
			if cmdArgs[1] == "status" {
				sf.parseGitStatus(string(output))
			} else {
				files := strings.Split(strings.TrimSpace(string(output)), "\n")
				for _, file := range files {
					if file != "" {
						sf.gitChangedFiles = append(sf.gitChangedFiles, file)
					}
				}
			}
			return nil
		}
	}

	// If no git commands work, assume no changes (not an error)
	fmt.Fprintf(os.Stderr, "No git changes detected or not in git repository\n")
	return nil
}

// parseGitStatus parses git status porcelain output
func (sf *SmartFilter) parseGitStatus(output string) {
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if len(line) > 3 {
			// Skip the status indicators (first 3 characters)
			filename := line[3:]
			sf.gitChangedFiles = append(sf.gitChangedFiles, filename)
		}
	}
}

// logFilteringResults logs the results of filtering
func (sf *SmartFilter) logFilteringResults(original, filtered map[string][]string) {
	originalCount := 0
	filteredCount := 0
	
	for _, hooks := range original {
		originalCount += len(hooks)
	}
	
	for _, hooks := range filtered {
		filteredCount += len(hooks)
	}
	
	if originalCount > filteredCount {
		saved := originalCount - filteredCount
		percentage := (float64(saved) / float64(originalCount)) * 100
		fmt.Fprintf(os.Stderr, "Smart filter: %d/%d hooks filtered out (%.1f%% reduction)\n", 
			saved, originalCount, percentage)
		
		if len(sf.gitChangedFiles) > 0 {
			fmt.Fprintf(os.Stderr, "Changed files: %v\n", sf.gitChangedFiles)
		}
	}
}

// GetChangedFiles returns the list of changed files
func (sf *SmartFilter) GetChangedFiles() []string {
	return sf.gitChangedFiles
}

// GetFilterStats returns filtering statistics
func (sf *SmartFilter) GetFilterStats() map[string]interface{} {
	return map[string]interface{}{
		"changed_files_count": len(sf.gitChangedFiles),
		"changed_files":       sf.gitChangedFiles,
		"tool_name":           sf.toolName,
		"git_analysis_enabled": sf.config.OptimizationSettings.EnableGitDiffAnalysis,
		"file_filter_enabled":  sf.config.OptimizationSettings.EnableFileTypeFilter,
		"tool_filter_enabled":  sf.config.OptimizationSettings.EnableToolBasedFilter,
	}
}

func main() {
	if len(os.Args) < 4 {
		fmt.Printf("Usage: %s <config-path> <tool-name> <hooks-json>\n", os.Args[0])
		os.Exit(1)
	}

	configPath := os.Args[1]
	toolName := os.Args[2]
	hooksJSON := os.Args[3]

	// Parse hooks from JSON
	var hookGroups map[string][]string
	if err := json.Unmarshal([]byte(hooksJSON), &hookGroups); err != nil {
		fmt.Printf("Error parsing hooks JSON: %v\n", err)
		os.Exit(1)
	}

	// Create smart filter
	filter, err := NewSmartFilter(configPath)
	if err != nil {
		fmt.Printf("Error creating smart filter: %v\n", err)
		os.Exit(1)
	}

	// Read tool input from stdin if available
	var toolInput map[string]interface{}
	if stat, _ := os.Stdin.Stat(); (stat.Mode() & os.ModeCharDevice) == 0 {
		input, err := ioutil.ReadAll(os.Stdin)
		if err == nil && len(input) > 0 {
			json.Unmarshal(input, &toolInput)
		}
	}

	// Filter hooks
	filteredGroups, err := filter.FilterHooks(toolName, toolInput, hookGroups)
	if err != nil {
		fmt.Printf("Error filtering hooks: %v\n", err)
		os.Exit(1)
	}

	// Output filtered hooks as JSON
	output, err := json.Marshal(filteredGroups)
	if err != nil {
		fmt.Printf("Error marshaling filtered hooks: %v\n", err)
		os.Exit(1)
	}

	fmt.Print(string(output))
}