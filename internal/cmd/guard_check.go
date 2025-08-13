package cmd

import (
	"fmt"
	"strings"

	"claude-wm-cli/internal/gitdiff"
	"claude-wm-cli/internal/mode"
)

// RunGuardCheck implements the guard check logic for SELF mode restrictions
func RunGuardCheck(workingDir string) error {
	// Determine if we're in SELF mode
	env := mode.Env{Root: workingDir}
	isSelfMode := mode.Self(env)

	// In non-SELF mode, no restrictions apply
	if !isSelfMode {
		// Silent success - no restrictions in user projects
		return nil
	}

	// We're in SELF mode - check for violations
	return checkSelfModeViolations(workingDir)
}

// checkSelfModeViolations checks for writing violations in SELF mode
func checkSelfModeViolations(workingDir string) error {
	// Get list of changed file paths
	changedPaths, err := gitdiff.ChangedPaths(workingDir)
	if err != nil {
		// This shouldn't happen due to gitdiff's graceful degradation,
		// but handle it just in case
		fmt.Printf("Warning: Failed to get changed paths: %v\n", err)
		return nil // Non-fatal - return success
	}

	// Check each changed file for violations
	var violations []string
	for _, path := range changedPaths {
		if isSelfModeViolation(path) {
			violations = append(violations, path)
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