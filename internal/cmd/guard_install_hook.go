package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
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

var (
	installHookYes bool
)

// GuardInstallHookCmd creates the guard install-hook subcommand
var GuardInstallHookCmd = &cobra.Command{
	Use:   "install-hook",
	Short: "Install Git pre-commit hook",
	Long: `Install a Git pre-commit hook that runs 'guard check'.

The hook will be installed at .git/hooks/pre-commit and will automatically
run guard check before each commit. If guard check fails, the commit will
be blocked.

The hook can be bypassed with 'git commit --no-verify' if needed.`,
	RunE: RunGuardInstallHook,
}

func init() {
	GuardInstallHookCmd.Flags().BoolVarP(&installHookYes, "yes", "y", false, "Skip confirmation prompt")
}

// RunGuardInstallHook installs the pre-commit hook (exported for use in cmd package)
func RunGuardInstallHook(cmd *cobra.Command, args []string) error {
	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Check if .git/hooks exists (Git repository detection)
	hooksDir := filepath.Join(cwd, ".git", "hooks")
	if _, err := os.Stat(hooksDir); os.IsNotExist(err) {
		return fmt.Errorf("not a Git repository (no .git/hooks directory found)")
	}

	// Define exact hook path: .git/hooks/pre-commit
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

		// Create backup in the same directory (.git/hooks/pre-commit.bak)
		backupPath := hookPath + ".bak"
		if err := backupFile(hookPath, backupPath); err != nil {
			return fmt.Errorf("failed to backup existing hook: %w", err)
		}
		fmt.Printf("✅ Existing hook backed up to %s\n", backupPath)
	}

	// Copy template to .git/hooks/pre-commit and set permissions
	if err := installHookTemplate(hookPath); err != nil {
		return fmt.Errorf("failed to install hook: %w", err)
	}

	fmt.Printf("✅ Pre-commit hook installed successfully at %s\n", hookPath)
	fmt.Println("📋 The hook will run 'claude-wm-cli guard check' before each commit.")
	fmt.Println("💡 To bypass the hook temporarily, use: git commit --no-verify")

	return nil
}

// installHookTemplate copies the embedded hook template to the target path with executable permissions
func installHookTemplate(targetPath string) error {
	// Write embedded template to target path
	if err := os.WriteFile(targetPath, []byte(hookTemplateContent), 0644); err != nil {
		return fmt.Errorf("failed to write hook file: %w", err)
	}

	// Explicitly set executable permissions (0755) to override umask
	if err := os.Chmod(targetPath, 0755); err != nil {
		return fmt.Errorf("failed to set executable permissions: %w", err)
	}

	return nil
}

// backupFile creates a backup copy of a file in the same directory
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