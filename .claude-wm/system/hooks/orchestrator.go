package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
	
	"./profiler"
)

// Configuration structures
type ParallelConfig struct {
	Version             string `json:"version"`
	Description         string `json:"description"`
	MaxConcurrentGroups int    `json:"max_concurrent_groups"`
	TimeoutSeconds      int    `json:"timeout_seconds"`
}

type HookGroup struct {
	Description       string   `json:"description"`
	Parallel          bool     `json:"parallel"`
	BackgroundEligible bool    `json:"background_eligible,omitempty"`
	Hooks             []string `json:"hooks"`
	MaxConcurrent     int      `json:"max_concurrent"`
	Priority          int      `json:"priority"`
}

type HookConfig struct {
	ParallelizationConfig ParallelConfig            `json:"parallelization_config"`
	HookGroups           map[string]HookGroup      `json:"hook_groups"`
	HookDependencies     map[string][]string       `json:"hook_dependencies"`
	FileTypeTriggers     map[string][]string       `json:"file_type_triggers"`
}

type HookResult struct {
	HookName     string        `json:"hook_name"`
	Success      bool          `json:"success"`
	ExitCode     int           `json:"exit_code"`
	Output       string        `json:"output"`
	Error        string        `json:"error,omitempty"`
	Duration     time.Duration `json:"duration"`
	GroupName    string        `json:"group_name"`
}

type ExecutionStats struct {
	TotalHooks       int           `json:"total_hooks"`
	SuccessfulHooks  int           `json:"successful_hooks"`
	FailedHooks      int           `json:"failed_hooks"`
	TotalDuration    time.Duration `json:"total_duration"`
	ParallelGroups   int           `json:"parallel_groups"`
	SequentialGroups int           `json:"sequential_groups"`
}

type HookOrchestrator struct {
	config     HookConfig
	hooksDir   string
	results    []HookResult
	stats      ExecutionStats
	mu         sync.Mutex
}

// NewHookOrchestrator creates a new orchestrator instance
func NewHookOrchestrator(configPath, hooksDir string) (*HookOrchestrator, error) {
	orchestrator := &HookOrchestrator{
		hooksDir: hooksDir,
		results:  make([]HookResult, 0),
	}

	// Load configuration
	configData, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	if err := json.Unmarshal(configData, &orchestrator.config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %v", err)
	}

	return orchestrator, nil
}

// ExecuteHooks executes all hooks according to the parallel configuration
func (ho *HookOrchestrator) ExecuteHooks(ctx context.Context, hookMatcher string, toolInput map[string]interface{}) error {
	startTime := time.Now()
	
	fmt.Printf("🚀 Hook Orchestrator starting - Parallel execution enabled\n")
	fmt.Printf("📊 Configuration: %d groups, %ds timeout\n", 
		ho.config.ParallelizationConfig.MaxConcurrentGroups,
		ho.config.ParallelizationConfig.TimeoutSeconds)
	
	// Initialize shared cache for this execution
	ho.initializeSharedCache()

	// Filter hooks based on matcher and file changes
	relevantGroups := ho.filterRelevantGroups(hookMatcher, toolInput)
	
	if len(relevantGroups) == 0 {
		fmt.Printf("✅ No relevant hooks to execute for matcher: %s\n", hookMatcher)
		return nil
	}

	// Sort groups by priority
	sortedGroups := ho.sortGroupsByPriority(relevantGroups)

	// Execute groups
	for _, groupName := range sortedGroups {
		group := ho.config.HookGroups[groupName]
		
		if group.Parallel {
			ho.stats.ParallelGroups++
			err := ho.executeParallelGroup(ctx, groupName, group, toolInput)
			if err != nil {
				return fmt.Errorf("parallel group %s failed: %v", groupName, err)
			}
		} else {
			ho.stats.SequentialGroups++
			err := ho.executeSequentialGroup(ctx, groupName, group, toolInput)
			if err != nil {
				return fmt.Errorf("sequential group %s failed: %v", groupName, err)
			}
		}
	}

	// Calculate final stats
	ho.stats.TotalDuration = time.Since(startTime)
	ho.stats.TotalHooks = len(ho.results)
	
	for _, result := range ho.results {
		if result.Success {
			ho.stats.SuccessfulHooks++
		} else {
			ho.stats.FailedHooks++
		}
	}

	// Print execution summary
	ho.printExecutionSummary()

	// Return error if any critical hooks failed
	if ho.stats.FailedHooks > 0 {
		return fmt.Errorf("%d hooks failed execution", ho.stats.FailedHooks)
	}

	return nil
}

// executeParallelGroup executes hooks in a group concurrently
func (ho *HookOrchestrator) executeParallelGroup(ctx context.Context, groupName string, group HookGroup, toolInput map[string]interface{}) error {
	fmt.Printf("⚡ Executing parallel group: %s (%d hooks)\n", groupName, len(group.Hooks))
	
	ctx, cancel := context.WithTimeout(ctx, time.Duration(ho.config.ParallelizationConfig.TimeoutSeconds)*time.Second)
	defer cancel()

	// Check if group is background eligible
	if group.BackgroundEligible {
		return ho.executeBackgroundGroup(ctx, groupName, group, toolInput)
	}

	// Create semaphore for concurrency control
	semaphore := make(chan struct{}, group.MaxConcurrent)
	var wg sync.WaitGroup
	
	// Execute hooks concurrently
	for _, hookName := range group.Hooks {
		wg.Add(1)
		go func(hook string) {
			defer wg.Done()
			
			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			
			result := ho.executeHookWithCache(ctx, hook, groupName, toolInput)
			
			ho.mu.Lock()
			ho.results = append(ho.results, result)
			ho.mu.Unlock()
			
			// Print result immediately
			ho.printHookResult(result)
		}(hookName)
	}
	
	wg.Wait()
	return nil
}

// executeSequentialGroup executes hooks in a group sequentially
func (ho *HookOrchestrator) executeSequentialGroup(ctx context.Context, groupName string, group HookGroup, toolInput map[string]interface{}) error {
	fmt.Printf("🔄 Executing sequential group: %s (%d hooks)\n", groupName, len(group.Hooks))
	
	for _, hookName := range group.Hooks {
		result := ho.executeHookWithCache(ctx, hookName, groupName, toolInput)
		
		ho.mu.Lock()
		ho.results = append(ho.results, result)
		ho.mu.Unlock()
		
		ho.printHookResult(result)
		
		// Invalidate cache if needed
		ho.invalidateCacheAfterExecution(hookName, toolInput)
		
		// Stop on first failure for sequential groups
		if !result.Success {
			return fmt.Errorf("hook %s failed in sequential group", hookName)
		}
	}
	
	return nil
}

// executeHook executes a single hook
func (ho *HookOrchestrator) executeHook(ctx context.Context, hookName, groupName string, toolInput map[string]interface{}) HookResult {
	startTime := time.Now()
	
	hookPath := filepath.Join(ho.hooksDir, hookName)
	
	// Create command based on file extension
	var cmd *exec.Cmd
	if strings.HasSuffix(hookName, ".py") {
		cmd = exec.CommandContext(ctx, "python3", hookPath)
	} else if strings.HasSuffix(hookName, ".sh") {
		cmd = exec.CommandContext(ctx, "bash", hookPath)
	} else {
		cmd = exec.CommandContext(ctx, hookPath)
	}
	
	// Set environment variables
	cmd.Env = append(os.Environ(),
		"HOOK_GROUP="+groupName,
		"PARALLEL_MODE=true",
	)
	
	// Pass tool input as JSON via stdin in the format expected by Go hooks
	// Always provide a JSON input, even if empty
	toolInputMap := make(map[string]interface{})
	toolName := ""
	
	if toolInput != nil {
		if tool, ok := toolInput["tool"]; ok {
			toolName = tool.(string)
		}
		if args, ok := toolInput["args"]; ok {
			if argsArray, ok := args.([]interface{}); ok && len(argsArray) > 0 {
				// Convert array to map for Go hooks
				if len(argsArray) > 0 {
					toolInputMap["command"] = argsArray[0]
				}
				if len(argsArray) > 1 {
					toolInputMap["args"] = argsArray[1:]
				}
			}
		}
	}
	
	wrappedInput := map[string]interface{}{
		"tool_name": toolName,
		"tool_input": toolInputMap,
	}
	inputJSON, _ := json.Marshal(wrappedInput)
	cmd.Stdin = strings.NewReader(string(inputJSON))
	
	// Execute command
	output, err := cmd.CombinedOutput()
	
	result := HookResult{
		HookName:  hookName,
		Success:   err == nil,
		ExitCode:  cmd.ProcessState.ExitCode(),
		Output:    string(output),
		Duration:  time.Since(startTime),
		GroupName: groupName,
	}
	
	if err != nil {
		result.Error = err.Error()
	}
	
	return result
}

// executeBackgroundGroup enqueues hooks for background execution
func (ho *HookOrchestrator) executeBackgroundGroup(ctx context.Context, groupName string, group HookGroup, toolInput map[string]interface{}) error {
	fmt.Printf("🔄 Enqueuing background group: %s (%d hooks)\n", groupName, len(group.Hooks))
	
	// Prepare tool input as JSON string
	var argsJSON string
	if toolInput != nil {
		inputBytes, err := json.Marshal(toolInput)
		if err != nil {
			fmt.Printf("⚠️  Failed to marshal tool input for background group: %v\n", err)
			argsJSON = "{}"
		} else {
			argsJSON = string(inputBytes)
		}
	} else {
		argsJSON = "{}"
	}

	// Enqueue each hook in the background
	for _, hookName := range group.Hooks {
		priority := 5 // Normal priority for background hooks
		
		// Adjust priority based on hook type
		if strings.Contains(hookName, "log-") {
			priority = 6 // Lower priority for logging
		} else if strings.Contains(hookName, "error-") {
			priority = 4 // Higher priority for error reporting
		}

		if err := ho.enqueueBackgroundHook(hookName, argsJSON, priority); err != nil {
			fmt.Printf("⚠️  Failed to enqueue background hook %s: %v\n", hookName, err)
			// Create a placeholder result for failed enqueue
			result := HookResult{
				HookName:  hookName,
				Success:   false,
				ExitCode:  1,
				Output:    "",
				Error:     fmt.Sprintf("Failed to enqueue: %v", err),
				Duration:  0,
				GroupName: groupName,
			}
			
			ho.mu.Lock()
			ho.results = append(ho.results, result)
			ho.mu.Unlock()
		} else {
			// Create a placeholder result for successful enqueue
			result := HookResult{
				HookName:  hookName,
				Success:   true,
				ExitCode:  0,
				Output:    "Enqueued for background execution",
				Error:     "",
				Duration:  time.Millisecond, // Minimal time for enqueue
				GroupName: groupName,
			}
			
			ho.mu.Lock()
			ho.results = append(ho.results, result)
			ho.mu.Unlock()
			
			fmt.Printf("🔄 %s enqueued for background execution (priority: %d)\n", hookName, priority)
		}
	}

	return nil
}

// enqueueBackgroundHook enqueues a single hook for background execution
func (ho *HookOrchestrator) enqueueBackgroundHook(hookName, args string, priority int) error {
	// Use the enqueue-hook.sh script to add job to background queue
	cmd := exec.Command(filepath.Join(ho.hooksDir, "enqueue-hook.sh"), hookName, args, fmt.Sprintf("%d", priority))
	cmd.Env = append(os.Environ(), "BACKGROUND_ENQUEUE=true")
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("enqueue command failed: %v, output: %s", err, output)
	}
	
	return nil
}

// initializeSharedCache sets up the shared cache for hook execution
func (ho *HookOrchestrator) initializeSharedCache() {
	// Ensure cache binary is available
	cacheBin := filepath.Join(ho.hooksDir, "shared-cache")
	if _, err := os.Stat(cacheBin); os.IsNotExist(err) {
		// Try to compile cache binary
		cacheDir := filepath.Join(ho.hooksDir, "cache")
		if _, err := os.Stat(filepath.Join(cacheDir, "shared-cache.go")); err == nil {
			cmd := exec.Command("go", "build", "-o", cacheBin, "shared-cache.go")
			cmd.Dir = cacheDir
			if err := cmd.Run(); err != nil {
				fmt.Printf("⚠️  Failed to compile shared cache: %v\n", err)
				return
			}
			fmt.Printf("🔨 Compiled shared cache binary\n")
		}
	}
	
	// Warm up cache with common operations
	ho.warmupCache()
}

// warmupCache pre-populates cache with frequently used data
func (ho *HookOrchestrator) warmupCache() {
	cacheScript := filepath.Join(ho.hooksDir, "cache-integration.sh")
	
	// Warm up git cache
	cmd := exec.Command(cacheScript, "warm-git")
	if err := cmd.Run(); err == nil {
		fmt.Printf("🔥 Git cache warmed\n")
	}
	
	// Warm up file cache for current directory
	cmd = exec.Command(cacheScript, "warm-files", ".")
	cmd.Dir = ho.hooksDir
	if err := cmd.Run(); err == nil {
		fmt.Printf("🔥 File cache warmed\n")
	}
}

// executeHookWithCache executes a hook with cache integration
func (ho *HookOrchestrator) executeHookWithCache(ctx context.Context, hookName, groupName string, toolInput map[string]interface{}) HookResult {
	// Start profiling
	prof := profiler.GetProfiler()
	timer := prof.StartTimer(hookName)
	
	startTime := time.Now()
	
	hookPath := filepath.Join(ho.hooksDir, hookName)
	
	// Create command based on file extension
	var cmd *exec.Cmd
	if strings.HasSuffix(hookName, ".py") {
		cmd = exec.CommandContext(ctx, "python3", hookPath)
	} else if strings.HasSuffix(hookName, ".sh") {
		cmd = exec.CommandContext(ctx, "bash", hookPath)
	} else {
		cmd = exec.CommandContext(ctx, hookPath)
	}
	
	// Set environment variables including cache integration
	cmd.Env = append(os.Environ(),
		"HOOK_GROUP="+groupName,
		"PARALLEL_MODE=true",
		"CACHE_ENABLED=true",
		"CACHE_INTEGRATION_SCRIPT="+filepath.Join(ho.hooksDir, "cache-integration.sh"),
	)
	
	// Pass tool input as JSON via stdin
	if toolInput != nil {
		inputJSON, _ := json.Marshal(toolInput)
		cmd.Stdin = strings.NewReader(string(inputJSON))
	}
	
	// Execute command
	output, err := cmd.CombinedOutput()
	
	result := HookResult{
		HookName:  hookName,
		Success:   err == nil,
		ExitCode:  cmd.ProcessState.ExitCode(),
		Output:    string(output),
		Duration:  time.Since(startTime),
		GroupName: groupName,
	}
	
	if err != nil {
		result.Error = err.Error()
	}
	
	// Record profiling metrics
	status := "success"
	if !result.Success {
		status = "error"
	}
	
	// Add custom metrics
	timer.SetCustomMetric("group_name", groupName)
	timer.SetCustomMetric("execution_mode", "cached")
	if toolInput != nil {
		if tool, ok := toolInput["tool"]; ok {
			timer.SetCustomMetric("tool_name", tool)
		}
	}
	
	// Stop profiling
	if err != nil {
		timer.Stop(status, result.Error)
	} else {
		timer.Stop(status)
	}
	
	// Log cache performance for analysis
	ho.logCachePerformance(hookName, result.Duration)
	
	return result
}

// logCachePerformance logs cache performance metrics
func (ho *HookOrchestrator) logCachePerformance(hookName string, duration time.Duration) {
	// This could be expanded to track cache hit rates, performance improvements, etc.
	if duration < 100*time.Millisecond {
		fmt.Printf("⚡ %s executed quickly (%dms) - likely cache benefit\n", hookName, duration.Milliseconds())
	}
}

// invalidateCacheAfterExecution invalidates relevant cache entries after hook execution
func (ho *HookOrchestrator) invalidateCacheAfterExecution(hookName string, toolInput map[string]interface{}) {
	cacheScript := filepath.Join(ho.hooksDir, "cache-integration.sh")
	
	// Invalidate git cache if this was a git-related hook
	if strings.Contains(hookName, "git") || strings.Contains(hookName, "commit") {
		cmd := exec.Command(cacheScript, "invalidate-git")
		cmd.Run() // Don't block on cache invalidation errors
	}
	
	// Invalidate file cache if this hook modifies files
	if strings.Contains(hookName, "write") || strings.Contains(hookName, "edit") {
		// Try to determine which files were modified from tool input
		if toolInput != nil {
			if filePath, ok := toolInput["file_path"].(string); ok {
				cmd := exec.Command(cacheScript, "invalidate-file", filePath)
				cmd.Run()
			}
		}
	}
}

// filterRelevantGroups filters groups based on matcher and file changes using smart filter
func (ho *HookOrchestrator) filterRelevantGroups(hookMatcher string, toolInput map[string]interface{}) map[string]HookGroup {
	relevantGroups := make(map[string]HookGroup)
	
	// Use smart filter to determine which hooks should run
	filteredGroups, err := ho.runSmartFilter(hookMatcher, toolInput)
	if err != nil {
		fmt.Printf("⚠️  Smart filter failed: %v, falling back to all groups\n", err)
		// Fallback to all groups if smart filter fails
		for name, group := range ho.config.HookGroups {
			relevantGroups[name] = group
		}
		return relevantGroups
	}
	
	// Build relevant groups based on filtered hooks
	for groupName, group := range ho.config.HookGroups {
		if filteredHooks, exists := filteredGroups[groupName]; exists && len(filteredHooks) > 0 {
			// Create a new group with only the filtered hooks
			filteredGroup := group
			filteredGroup.Hooks = filteredHooks
			relevantGroups[groupName] = filteredGroup
		}
	}
	
	return relevantGroups
}

// runSmartFilter executes the smart filter to determine which hooks should run
func (ho *HookOrchestrator) runSmartFilter(hookMatcher string, toolInput map[string]interface{}) (map[string][]string, error) {
	// Prepare tool name from matcher or tool input
	toolName := hookMatcher
	if toolName == "" && toolInput != nil {
		if tool, ok := toolInput["tool"]; ok {
			toolName = tool.(string)
		}
	}
	if toolName == "" {
		toolName = "unknown"
	}
	
	// Convert hook groups to the format expected by smart filter
	hookGroups := make(map[string][]string)
	for groupName, group := range ho.config.HookGroups {
		hookGroups[groupName] = group.Hooks
	}
	
	// Marshal hook groups to JSON
	hooksJSON, err := json.Marshal(hookGroups)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal hook groups: %v", err)
	}
	
	// Prepare smart filter command
	smartFilterPath := filepath.Join(ho.hooksDir, "filter", "smart-filter")
	configPath := filepath.Join(ho.hooksDir, "config", "parallel-groups.json")
	
	// Ensure smart filter is compiled
	if err := ho.ensureSmartFilterCompiled(); err != nil {
		return nil, fmt.Errorf("failed to compile smart filter: %v", err)
	}
	
	// Execute smart filter
	cmd := exec.Command(smartFilterPath, configPath, toolName, string(hooksJSON))
	
	// Pass tool input via stdin
	if toolInput != nil {
		inputJSON, _ := json.Marshal(toolInput)
		cmd.Stdin = strings.NewReader(string(inputJSON))
	}
	
	// Set environment variables
	cmd.Env = append(os.Environ(), "SMART_FILTER_MODE=true")
	
	// Capture stdout and stderr separately
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("smart filter execution failed: %v, stderr: %s", err, stderr.String())
	}
	
	output := stdout.Bytes()
	
	// Parse smart filter output
	var filteredGroups map[string][]string
	if err := json.Unmarshal(output, &filteredGroups); err != nil {
		return nil, fmt.Errorf("failed to parse smart filter output: %v", err)
	}
	
	return filteredGroups, nil
}

// ensureSmartFilterCompiled ensures the smart filter binary is compiled
func (ho *HookOrchestrator) ensureSmartFilterCompiled() error {
	smartFilterPath := filepath.Join(ho.hooksDir, "filter", "smart-filter")
	
	// Check if binary exists
	if _, err := os.Stat(smartFilterPath); err == nil {
		return nil
	}
	
	// Try to compile smart filter
	filterDir := filepath.Join(ho.hooksDir, "filter")
	if _, err := os.Stat(filepath.Join(filterDir, "smart-filter.go")); err == nil {
		fmt.Printf("🔨 Compiling smart filter...\n")
		cmd := exec.Command("go", "build", "-o", smartFilterPath, "smart-filter.go")
		cmd.Dir = filterDir
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to compile smart filter: %v", err)
		}
		fmt.Printf("✅ Smart filter compiled successfully\n")
	} else {
		return fmt.Errorf("smart filter source not found: %v", err)
	}
	
	return nil
}

// sortGroupsByPriority sorts groups by their priority
func (ho *HookOrchestrator) sortGroupsByPriority(groups map[string]HookGroup) []string {
	type groupPriority struct {
		name     string
		priority int
	}
	
	var sorted []groupPriority
	for name, group := range groups {
		sorted = append(sorted, groupPriority{name, group.Priority})
	}
	
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].priority < sorted[j].priority
	})
	
	var result []string
	for _, gp := range sorted {
		result = append(result, gp.name)
	}
	
	return result
}

// printHookResult prints the result of a single hook execution
func (ho *HookOrchestrator) printHookResult(result HookResult) {
	status := "✅"
	if !result.Success {
		status = "❌"
	}
	
	fmt.Printf("%s %s (%s) - %dms [%s]\n", 
		status, 
		result.HookName, 
		result.GroupName,
		result.Duration.Milliseconds(),
		func() string {
			if result.Success {
				return "SUCCESS"
			}
			return fmt.Sprintf("FAILED: %d", result.ExitCode)
		}())
}

// printExecutionSummary prints the final execution summary
func (ho *HookOrchestrator) printExecutionSummary() {
	fmt.Printf("\n📊 Execution Summary:\n")
	fmt.Printf("   Total hooks: %d\n", ho.stats.TotalHooks)
	fmt.Printf("   Successful: %d\n", ho.stats.SuccessfulHooks)
	fmt.Printf("   Failed: %d\n", ho.stats.FailedHooks)
	fmt.Printf("   Total time: %dms\n", ho.stats.TotalDuration.Milliseconds())
	fmt.Printf("   Parallel groups: %d\n", ho.stats.ParallelGroups)
	fmt.Printf("   Sequential groups: %d\n", ho.stats.SequentialGroups)
	
	// Calculate average time per hook
	if ho.stats.TotalHooks > 0 {
		avgTime := ho.stats.TotalDuration.Milliseconds() / int64(ho.stats.TotalHooks)
		fmt.Printf("   Average time per hook: %dms\n", avgTime)
	}
	
	// Print profiling summary
	fmt.Printf("\n📈 Performance Profiling:\n")
	prof := profiler.GetProfiler()
	slowest := prof.GetTopSlowestHooks(3)
	if len(slowest) > 0 {
		fmt.Printf("   Top 3 slowest hooks:\n")
		for i, hook := range slowest {
			fmt.Printf("     %d. %s: %.2fms avg (runs: %d)\n", 
				i+1, hook.HookName, hook.AverageTimeMs, hook.TotalExecutions)
		}
	}
}

// GetResults returns the execution results
func (ho *HookOrchestrator) GetResults() []HookResult {
	return ho.results
}

// GetStats returns the execution statistics
func (ho *HookOrchestrator) GetStats() ExecutionStats {
	return ho.stats
}

func main() {
	if len(os.Args) < 3 {
		fmt.Printf("Usage: %s <config-path> <hooks-dir> [matcher] [tool-input-json]\n", os.Args[0])
		os.Exit(1)
	}
	
	configPath := os.Args[1]
	hooksDir := os.Args[2]
	
	hookMatcher := ""
	if len(os.Args) > 3 {
		hookMatcher = os.Args[3]
	}
	
	var toolInput map[string]interface{}
	if len(os.Args) > 4 {
		if err := json.Unmarshal([]byte(os.Args[4]), &toolInput); err != nil {
			fmt.Printf("Error parsing tool input JSON: %v\n", err)
			os.Exit(1)
		}
	}
	
	// Create orchestrator
	orchestrator, err := NewHookOrchestrator(configPath, hooksDir)
	if err != nil {
		fmt.Printf("Error creating orchestrator: %v\n", err)
		os.Exit(1)
	}
	
	// Execute hooks
	ctx := context.Background()
	if err := orchestrator.ExecuteHooks(ctx, hookMatcher, toolInput); err != nil {
		fmt.Printf("Error executing hooks: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Printf("🎉 All hooks completed successfully!\n")
}