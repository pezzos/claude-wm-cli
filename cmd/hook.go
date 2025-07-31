package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"claude-wm-cli/internal/hooks"
)

var hookCmd = &cobra.Command{
	Use:   "hook",
	Short: "Execute internal hooks",
	Long:  "Execute internal validation and formatting hooks for claude-wm-cli",
}

var gitValidationCmd = &cobra.Command{
	Use:   "git-validation",
	Short: "Run git validation hook",
	Run: func(cmd *cobra.Command, args []string) {
		projectRoot, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting current directory: %v\n", err)
			os.Exit(1)
		}

		handler := hooks.NewHookHandler(projectRoot)
		if err := handler.HandleGitValidation(); err != nil {
			fmt.Fprintf(os.Stderr, "Git validation failed: %v\n", err)
			os.Exit(1)
		}
	},
}

var autoFormatCmd = &cobra.Command{
	Use:   "auto-format",
	Short: "Run auto-formatting hook",
	Run: func(cmd *cobra.Command, args []string) {
		projectRoot, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting current directory: %v\n", err)
			os.Exit(1)
		}

		handler := hooks.NewHookHandler(projectRoot)
		if err := handler.HandleAutoFormat(); err != nil {
			fmt.Fprintf(os.Stderr, "Auto-formatting failed: %v\n", err)
			os.Exit(1)
		}
	},
}

var duplicateDetectionCmd = &cobra.Command{
	Use:   "duplicate-detection",
	Short: "Run duplicate detection hook",
	Run: func(cmd *cobra.Command, args []string) {
		projectRoot, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting current directory: %v\n", err)
			os.Exit(1)
		}

		handler := hooks.NewHookHandler(projectRoot)
		if err := handler.HandleDuplicateDetection(); err != nil {
			fmt.Fprintf(os.Stderr, "Duplicate detection failed: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	hookCmd.AddCommand(gitValidationCmd)
	hookCmd.AddCommand(autoFormatCmd)
	hookCmd.AddCommand(duplicateDetectionCmd)
	rootCmd.AddCommand(hookCmd)
}