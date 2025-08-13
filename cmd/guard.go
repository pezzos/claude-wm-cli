package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	guardcmd "claude-wm-cli/internal/cmd"
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

	// Use the implementation from internal/cmd package
	return guardcmd.RunGuardCheck(cwd)
}

// runGuardInstallHook installs the pre-commit hook
func runGuardInstallHook(cmd *cobra.Command, args []string) error {
	// Use the implementation from internal/cmd package
	return guardcmd.RunGuardInstallHook(cmd, args)
}

