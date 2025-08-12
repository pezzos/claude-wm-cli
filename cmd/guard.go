package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"claude-wm-cli/internal/git"
	"claude-wm-cli/internal/mode"
)

var guardCmd = &cobra.Command{
	Use:   "guard",
	Short: "Guard against unwanted modifications",
	Long: `Guard commands help prevent unwanted modifications based on the current mode.

In SELF mode (claude-wm-cli project), certain directories are protected:
- .claude/ modifications are restricted (use 'config install' or 'config update')
- internal/config/system/ and source code modifications are allowed

In non-SELF mode (user projects), no restrictions apply.`,
}

var guardCheckCmd = &cobra.Command{
	Use:   "check",
	Short: "Check current changes against writing restrictions",
	Long: `Check Git working tree changes against SELF mode writing restrictions.

This command examines staged and unstaged changes in the Git working tree
and validates them against the current mode's writing restrictions.

Exit codes:
  0 - No violations found or not in SELF mode
  1 - Violations found (in SELF mode)
  2 - Warning conditions (Git not available, not a Git repo)`,
	RunE: runGuardCheck,
}

func init() {
	rootCmd.AddCommand(guardCmd)
	guardCmd.AddCommand(guardCheckCmd)
}

// runGuardCheck implements the guard check logic
func runGuardCheck(cmd *cobra.Command, args []string) error {
	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Determine if we're in SELF mode
	env := mode.Env{Root: cwd}
	isSelfMode := mode.Self(env)

	// In non-SELF mode, no restrictions apply
	if !isSelfMode {
		// Silent success - no restrictions in user projects
		return nil
	}

	// We're in SELF mode - check for violations
	return checkSelfModeViolations(cwd)
}

// checkSelfModeViolations checks for writing violations in SELF mode
func checkSelfModeViolations(workingDir string) error {
	// Create Git repository wrapper
	repo := git.NewRepository(workingDir, nil)

	// Check if this is a Git repository
	if !repo.IsRepository() {
		fmt.Printf("Warning: Not a Git repository - guard check skipped\n")
		return nil // Non-fatal - return exit code 0
	}

	// Get current Git status
	status, err := repo.GetStatus()
	if err != nil {
		// Handle Git command failures gracefully
		fmt.Printf("Warning: Failed to get Git status: %v\n", err)
		return nil // Non-fatal - return exit code 0
	}

	// Check each changed file for violations
	var violations []string
	for _, fileStatus := range status.Files {
		// Check if file path violates SELF mode restrictions
		if isSelfModeViolation(fileStatus.Path) {
			violations = append(violations, fileStatus.Path)
		}
	}

	// Report violations if found
	if len(violations) > 0 {
		fmt.Printf("❌ SELF mode violation: modifications to .claude/ detected:\n\n")
		for _, path := range violations {
			fmt.Printf("   %s\n", path)
		}
		fmt.Printf("\nIn SELF mode, .claude/ modifications are restricted to prevent\n")
		fmt.Printf("accidental changes to the project's configuration structure.\n\n")
		fmt.Printf("💡 To make configuration changes:\n")
		fmt.Printf("   • Use 'claude-wm-cli config install' for initial setup\n")
		fmt.Printf("   • Use 'claude-wm-cli config update' for configuration updates\n\n")
		fmt.Printf("Reason for SELF mode: %s\n", mode.Reason(mode.Env{Root: workingDir}))
		
		return fmt.Errorf("guard check failed: %d violation(s) found", len(violations))
	}

	// Silent success - no violations found
	return nil
}

// isSelfModeViolation checks if a file path violates SELF mode restrictions
func isSelfModeViolation(path string) bool {
	// Normalize path separators to forward slashes for consistent checking
	normalizedPath := strings.ReplaceAll(path, "\\", "/")
	
	// SELF mode restriction: no modifications to .claude/
	if strings.HasPrefix(normalizedPath, ".claude/") {
		return true
	}

	// Future: Add more SELF mode restrictions here if needed
	// For now, only .claude/ is restricted

	return false
}