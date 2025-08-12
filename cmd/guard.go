package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"claude-wm-cli/internal/git"
	"claude-wm-cli/internal/mode"
)

const hookTemplateContent = `#!/bin/sh
# Claude WM CLI Pre-commit Hook
# 
# This hook runs 'claude-wm-cli guard check' before each commit.
# It will block commits that violate SELF mode restrictions.
#
# Installation: claude-wm-cli guard install-hook
# Manual installation: copy this file to .git/hooks/pre-commit and chmod +x

set -e

# Find git repository root
REPO_ROOT=$(git rev-parse --show-toplevel 2>/dev/null || echo "")
CLAUDE_WM_CLI=""

# Strategy 1: Look for claude-wm-cli in repo root
if [ -n "$REPO_ROOT" ] && [ -x "$REPO_ROOT/claude-wm-cli" ]; then
    CLAUDE_WM_CLI="$REPO_ROOT/claude-wm-cli"
    echo "Using local claude-wm-cli: $CLAUDE_WM_CLI"
elif [ -n "$REPO_ROOT" ] && [ -x "$REPO_ROOT/build/claude-wm-cli" ]; then
    CLAUDE_WM_CLI="$REPO_ROOT/build/claude-wm-cli"
    echo "Using built claude-wm-cli: $CLAUDE_WM_CLI"
# Strategy 2: Look in PATH
elif command -v claude-wm-cli >/dev/null 2>&1; then
    CLAUDE_WM_CLI="claude-wm-cli"
    echo "Using claude-wm-cli from PATH"
# Strategy 3: Try to build it
elif [ -n "$REPO_ROOT" ] && [ -f "$REPO_ROOT/Makefile" ]; then
    echo "claude-wm-cli not found, attempting to build..."
    cd "$REPO_ROOT"
    if make build >/dev/null 2>&1; then
        if [ -x "$REPO_ROOT/build/claude-wm-cli" ]; then
            CLAUDE_WM_CLI="$REPO_ROOT/build/claude-wm-cli"
            echo "Built and using: $CLAUDE_WM_CLI"
        fi
    fi
fi

# Final check
if [ -z "$CLAUDE_WM_CLI" ]; then
    echo "Error: claude-wm-cli not found" >&2
    echo "Tried:" >&2
    echo "  1. $REPO_ROOT/claude-wm-cli" >&2
    echo "  2. $REPO_ROOT/build/claude-wm-cli" >&2
    echo "  3. PATH lookup" >&2
    echo "  4. Building with make" >&2
    echo "" >&2
    echo "Please ensure claude-wm-cli is available or the build system works." >&2
    exit 1
fi

# Run guard check
echo "Running pre-commit guard check..."

# Execute guard check and capture result
if "$CLAUDE_WM_CLI" guard check; then
    echo "✅ Guard check passed - commit allowed"
    exit 0
else
    echo "" >&2
    echo "❌ Commit blocked by guard check" >&2
    echo "" >&2
    echo "The guard check detected violations that prevent this commit." >&2
    echo "Please fix the issues above and try committing again." >&2
    echo "" >&2
    echo "To bypass this check temporarily, use:" >&2
    echo "  git commit --no-verify" >&2
    echo "" >&2
    exit 1
fi
`

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

var (
	installHookYes bool
)

var guardInstallHookCmd = &cobra.Command{
	Use:   "install-hook",
	Short: "Install Git pre-commit hook",
	Long: `Install a Git pre-commit hook that runs 'guard check'.

The hook will be installed at .git/hooks/pre-commit and will automatically
run guard check before each commit. If guard check fails, the commit will
be blocked.

The hook can be bypassed with 'git commit --no-verify' if needed.`,
	RunE: runGuardInstallHook,
}

func init() {
	rootCmd.AddCommand(guardCmd)
	guardCmd.AddCommand(guardCheckCmd)
	guardCmd.AddCommand(guardInstallHookCmd)
	
	// Add flags
	guardInstallHookCmd.Flags().BoolVarP(&installHookYes, "yes", "y", false, "Skip confirmation prompt")
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

// runGuardInstallHook installs the pre-commit hook
func runGuardInstallHook(cmd *cobra.Command, args []string) error {
	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Check if .git/hooks exists
	hooksDir := filepath.Join(cwd, ".git", "hooks")
	if _, err := os.Stat(hooksDir); os.IsNotExist(err) {
		return fmt.Errorf("not a Git repository (no .git/hooks directory found)")
	}

	hookPath := filepath.Join(hooksDir, "pre-commit")

	// Check if hook already exists
	if _, err := os.Stat(hookPath); err == nil {
		if !installHookYes {
			// Ask for confirmation
			fmt.Printf("A pre-commit hook already exists. Install guard hook anyway? [y/N]: ")
			reader := bufio.NewReader(os.Stdin)
			response, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("failed to read input: %w", err)
			}
			response = strings.TrimSpace(strings.ToLower(response))
			if response != "y" && response != "yes" {
				fmt.Println("Installation cancelled.")
				return nil
			}
		}

		// Create backup
		backupPath := hookPath + ".bak"
		if err := backupFile(hookPath, backupPath); err != nil {
			return fmt.Errorf("failed to backup existing hook: %w", err)
		}
		fmt.Printf("✅ Existing hook backed up to %s\n", backupPath)
	}

	// Copy template and set permissions
	if err := installHookTemplate(hookPath); err != nil {
		return fmt.Errorf("failed to install hook: %w", err)
	}

	fmt.Printf("✅ Pre-commit hook installed successfully at %s\n", hookPath)
	fmt.Println("📋 The hook will run 'claude-wm-cli guard check' before each commit.")
	fmt.Println("💡 To bypass the hook temporarily, use: git commit --no-verify")

	return nil
}

// installHookTemplate copies the embedded hook template to the target path
func installHookTemplate(targetPath string) error {
	// Write embedded template to target path
	if err := os.WriteFile(targetPath, []byte(hookTemplateContent), 0755); err != nil {
		return fmt.Errorf("failed to write hook file: %w", err)
	}

	return nil
}

// backupFile creates a backup copy of a file
func backupFile(src, dst string) error {
	// Read source file
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	// Get source file permissions
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	// Write backup with same permissions
	return os.WriteFile(dst, data, srcInfo.Mode())
}