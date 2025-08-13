package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// TestLevel represents a testing level in the L0-L3 protocol
type TestLevel struct {
	Level       string
	Name        string
	Description string
	Commands    []string
	Timeout     time.Duration
}

// TestResult represents the result of running a test level
type TestResult struct {
	Level   string
	Success bool
	Output  string
	Error   string
	Duration time.Duration
}

// TestRunner orchestrates the complete test suite
type TestRunner struct {
	levels []TestLevel
	results []TestResult
	verbose bool
}

// NewTestRunner creates a new test runner with default configuration
func NewTestRunner() *TestRunner {
	return &TestRunner{
		levels: []TestLevel{
			{
				Level:       "L0",
				Name:        "Smoke Tests",
				Description: "Basic functionality validation",
				Commands:    []string{"make", "test-smoke"},
				Timeout:     30 * time.Second,
			},
			{
				Level:       "L1",
				Name:        "Unit Tests",
				Description: "Component testing",
				Commands:    []string{"make", "test-unit"},
				Timeout:     2 * time.Minute,
			},
			{
				Level:       "L2",
				Name:        "Integration Tests",
				Description: "Component interaction testing",
				Commands:    []string{"make", "test-integration"},
				Timeout:     5 * time.Minute,
			},
			{
				Level:       "L3",
				Name:        "Guard/Hook Tests",
				Description: "Guard and hook validation",
				Commands:    []string{"make", "test-guard"},
				Timeout:     3 * time.Minute,
			},
			{
				Level:       "L4",
				Name:        "System Tests",
				Description: "End-to-end system validation",
				Commands:    []string{"make", "test-system"},
				Timeout:     10 * time.Minute,
			},
		},
		verbose: false,
	}
}

// Run executes the complete test suite
func (tr *TestRunner) Run() error {
	fmt.Println("🚀 Claude WM CLI Test Suite Runner")
	fmt.Println("═══════════════════════════════════════")
	fmt.Println()

	// Generate manifest first
	fmt.Println("📋 Generating system manifest...")
	if err := tr.runCommand([]string{"make", "manifest"}, 30*time.Second); err != nil {
		fmt.Printf("❌ Failed to generate manifest: %v\n", err)
		return err
	}
	fmt.Println("✅ Manifest generated successfully")
	fmt.Println()

	startTime := time.Now()
	
	// Run each test level
	for _, level := range tr.levels {
		result := tr.runTestLevel(level)
		tr.results = append(tr.results, result)
		
		if !result.Success {
			fmt.Println()
			fmt.Printf("❌ Test suite failed at %s level\n", level.Level)
			tr.printSummary(false)
			return fmt.Errorf("tests failed at %s level", level.Level)
		}
	}

	totalDuration := time.Since(startTime)
	fmt.Println()
	fmt.Printf("🎉 All tests completed successfully in %v\n", totalDuration.Round(time.Second))
	tr.printSummary(true)
	
	return nil
}

// runTestLevel executes a single test level
func (tr *TestRunner) runTestLevel(level TestLevel) TestResult {
	fmt.Printf("🧪 Running %s: %s\n", level.Level, level.Name)
	fmt.Printf("   %s\n", level.Description)
	
	startTime := time.Now()
	
	err := tr.runCommand(level.Commands, level.Timeout)
	duration := time.Since(startTime)
	
	result := TestResult{
		Level:    level.Level,
		Success:  err == nil,
		Duration: duration,
	}
	
	if err != nil {
		result.Error = err.Error()
		fmt.Printf("   ❌ Failed in %v: %s\n", duration.Round(time.Millisecond), err.Error())
	} else {
		fmt.Printf("   ✅ Passed in %v\n", duration.Round(time.Millisecond))
	}
	
	return result
}

// runCommand executes a command with timeout
func (tr *TestRunner) runCommand(args []string, timeout time.Duration) error {
	if len(args) == 0 {
		return fmt.Errorf("no command specified")
	}
	
	cmd := exec.Command(args[0], args[1:]...)
	
	if tr.verbose {
		fmt.Printf("   → Running: %s\n", strings.Join(args, " "))
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	
	// Start the command
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start command: %w", err)
	}
	
	// Wait for completion with timeout
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()
	
	select {
	case err := <-done:
		return err
	case <-time.After(timeout):
		// Kill the process on timeout
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
		return fmt.Errorf("command timed out after %v", timeout)
	}
}

// printSummary prints a summary of all test results
func (tr *TestRunner) printSummary(allPassed bool) {
	fmt.Println()
	fmt.Println("📊 Test Suite Summary")
	fmt.Println("══════════════════════")
	
	maxLevelWidth := 0
	maxNameWidth := 0
	
	for _, result := range tr.results {
		if len(result.Level) > maxLevelWidth {
			maxLevelWidth = len(result.Level)
		}
		
		for _, level := range tr.levels {
			if level.Level == result.Level && len(level.Name) > maxNameWidth {
				maxNameWidth = len(level.Name)
			}
		}
	}
	
	for _, result := range tr.results {
		var levelName string
		for _, level := range tr.levels {
			if level.Level == result.Level {
				levelName = level.Name
				break
			}
		}
		
		status := "❌"
		if result.Success {
			status = "✅"
		}
		
		fmt.Printf("%-*s %-*s %s (%v)\n", 
			maxLevelWidth, result.Level,
			maxNameWidth, levelName,
			status, 
			result.Duration.Round(time.Millisecond))
	}
	
	fmt.Println()
	
	if allPassed {
		fmt.Println("🎊 All test levels passed successfully!")
	} else {
		fmt.Println("💥 Some tests failed - see details above")
	}
	
	// Coverage suggestion
	if allPassed {
		fmt.Println()
		fmt.Println("💡 Next steps:")
		fmt.Println("   • Generate coverage report: make coverage-html")
		fmt.Println("   • Run performance benchmarks: go test -bench=./...")
		fmt.Println("   • Check code quality: make lint")
	}
}

// SetVerbose enables or disables verbose output
func (tr *TestRunner) SetVerbose(verbose bool) {
	tr.verbose = verbose
}

// GetResults returns the test results
func (tr *TestRunner) GetResults() []TestResult {
	return tr.results
}

// main is the entry point for the test runner
func main() {
	runner := NewTestRunner()
	
	// Check for verbose flag
	for _, arg := range os.Args[1:] {
		switch arg {
		case "-v", "--verbose":
			runner.SetVerbose(true)
		case "-h", "--help":
			printHelp()
			os.Exit(0)
		}
	}
	
	if err := runner.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Test runner failed: %v\n", err)
		os.Exit(1)
	}
}

// printHelp prints usage information
func printHelp() {
	fmt.Println("Claude WM CLI Test Suite Runner")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  go run ./internal/testrunner/main.go [flags]")
	fmt.Println("  make test-runner")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  -v, --verbose    Enable verbose output")
	fmt.Println("  -h, --help       Show this help message")
	fmt.Println()
	fmt.Println("Test Levels:")
	fmt.Println("  L0: Smoke Tests       - Basic functionality (< 30s)")
	fmt.Println("  L1: Unit Tests        - Component testing (< 2m)")
	fmt.Println("  L2: Integration Tests - Component interaction (< 5m)")
	fmt.Println("  L3: Guard/Hook Tests  - Validation systems (< 3m)")
	fmt.Println("  L4: System Tests      - End-to-end testing (< 10m)")
	fmt.Println()
	fmt.Println("The runner executes tests sequentially and stops on first failure.")
	fmt.Println("Use 'make test-all' for direct Make-based execution.")
}