package hooks

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"time"
)

// HookResult represents the result of a hook execution
type HookResult struct {
	HookName string        `json:"hook_name"`
	Success  bool          `json:"success"`
	ExitCode int           `json:"exit_code"`
	Output   string        `json:"output"`
	Error    string        `json:"error,omitempty"`
	Duration time.Duration `json:"duration"`
}

// ExecutionStats tracks execution statistics
type ExecutionStats struct {
	TotalHooks      int           `json:"total_hooks"`
	SuccessfulHooks int           `json:"successful_hooks"`
	FailedHooks     int           `json:"failed_hooks"`
	TotalDuration   time.Duration `json:"total_duration"`
}

// Orchestrator manages parallel hook execution for claude-wm-cli
type Orchestrator struct {
	results []HookResult
	stats   ExecutionStats
	mu      sync.Mutex
	timeout time.Duration
}

// NewOrchestrator creates a new hook orchestrator
func NewOrchestrator() *Orchestrator {
	return &Orchestrator{
		results: make([]HookResult, 0),
		timeout: 30 * time.Second, // 30 second timeout for hooks
	}
}

// ExecuteHooks executes multiple hooks in parallel
func (o *Orchestrator) ExecuteHooks(ctx context.Context, hooks []string, toolInput map[string]interface{}) error {
	startTime := time.Now()

	fmt.Printf("🚀 Executing %d hooks in parallel\n", len(hooks))

	// Channel for results
	resultsChan := make(chan HookResult, len(hooks))

	// Execute hooks in parallel
	var wg sync.WaitGroup
	for _, hookPath := range hooks {
		wg.Add(1)
		go func(path string) {
			defer wg.Done()
			result := o.executeHook(ctx, path, toolInput)
			resultsChan <- result
		}(hookPath)
	}

	// Wait for all hooks to complete
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Collect results
	for result := range resultsChan {
		o.mu.Lock()
		o.results = append(o.results, result)
		if result.Success {
			o.stats.SuccessfulHooks++
		} else {
			o.stats.FailedHooks++
		}
		o.mu.Unlock()
	}

	o.stats.TotalHooks = len(hooks)
	o.stats.TotalDuration = time.Since(startTime)

	// Print summary
	o.printSummary()

	return nil
}

// executeHook executes a single hook with timeout
func (o *Orchestrator) executeHook(ctx context.Context, hookPath string, toolInput map[string]interface{}) HookResult {
	startTime := time.Now()

	result := HookResult{
		HookName: hookPath,
		Success:  false,
	}

	// Create context with timeout
	hookCtx, cancel := context.WithTimeout(ctx, o.timeout)
	defer cancel()

	// Prepare command
	var cmd *exec.Cmd
	if hookPath[len(hookPath)-3:] == ".go" {
		// Go hook - compile and run
		cmd = exec.CommandContext(hookCtx, "go", "run", hookPath)
	} else {
		// Shell hook
		cmd = exec.CommandContext(hookCtx, "bash", hookPath)
	}

	// Set up environment
	cmd.Env = os.Environ()

	// Execute command
	output, err := cmd.CombinedOutput()
	result.Output = string(output)
	result.Duration = time.Since(startTime)

	if err != nil {
		result.Error = err.Error()
		if exitError, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitError.ExitCode()
		} else {
			result.ExitCode = 1
		}
	} else {
		result.Success = true
		result.ExitCode = 0
	}

	return result
}

// GetResults returns all hook execution results
func (o *Orchestrator) GetResults() []HookResult {
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.results
}

// GetStats returns execution statistics
func (o *Orchestrator) GetStats() ExecutionStats {
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.stats
}

// printSummary prints execution summary
func (o *Orchestrator) printSummary() {
	o.mu.Lock()
	defer o.mu.Unlock()

	fmt.Printf("\n📊 Hook Execution Summary:\n")
	fmt.Printf("  Total: %d hooks\n", o.stats.TotalHooks)
	fmt.Printf("  Success: %d\n", o.stats.SuccessfulHooks)
	fmt.Printf("  Failed: %d\n", o.stats.FailedHooks)
	fmt.Printf("  Duration: %v\n", o.stats.TotalDuration)

	// Show failed hooks
	if o.stats.FailedHooks > 0 {
		fmt.Printf("\n❌ Failed hooks:\n")
		for _, result := range o.results {
			if !result.Success {
				fmt.Printf("  - %s (exit code: %d)\n", result.HookName, result.ExitCode)
				if result.Error != "" {
					fmt.Printf("    Error: %s\n", result.Error)
				}
			}
		}
	}
}

// HasFailures returns true if any hooks failed
func (o *Orchestrator) HasFailures() bool {
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.stats.FailedHooks > 0
}