package cmd

import (
	"claude-wm-cli/internal/cmd"
	
	"github.com/spf13/cobra"
)

// devCmd represents the dev command
var devCmd = &cobra.Command{
	Use:   "dev",
	Short: "Development utilities and testing tools",
	Long: `Development utilities and testing tools for claude-wm-cli.

This command group provides development-focused utilities for testing,
debugging, and experimenting with the claude-wm-cli system.

Available subcommands:
  sandbox      Create a testing sandbox from Upstream system files
  import-local Import changes from .claude/ to internal/config/system/`,
}

func init() {
	rootCmd.AddCommand(devCmd)
	
	// Add dev subcommands
	devCmd.AddCommand(cmd.DevSandboxCmd)
	devCmd.AddCommand(cmd.DevImportLocalCmd)
}