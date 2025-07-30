/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	"claude-wm-cli/internal/model"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Version information - will be set at build time
var (
	Version   = "dev"
	GitCommit = "unknown"
	BuildTime = "unknown"
)

// Global configuration variables
var (
	cfgFile   string
	verbose   bool
	debugMode bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "claude-wm-cli",
	Short: "Window Management CLI Tool with AI Integration",
	Long: `Claude WM CLI is a window management command-line tool that integrates with
Claude AI to provide intelligent window management and command execution capabilities.

This tool helps you manage your desktop windows efficiently and execute complex
commands through an AI-powered interface with robust timeout and retry mechanisms.

CORE FEATURES:
  • Project workflow management with atomic JSON state
  • Robust Claude command execution with 30s timeout
  • Context-aware interactive navigation
  • Cross-platform compatibility (Windows, macOS, Linux)
  • Comprehensive error handling and validation

WORKFLOW:
  1. Initialize a new project: claude-wm-cli init my-project
  2. Check project status:    claude-wm-cli status
  3. Execute Claude commands: claude-wm-cli execute "claude --help"
  4. Use verbose mode:        claude-wm-cli --verbose [command]

EXAMPLES:
  claude-wm-cli init my-project                    # Initialize new project
  claude-wm-cli status                             # Show current state
  claude-wm-cli execute "claude -p '/help'"       # Execute Claude with prompt
  claude-wm-cli execute --timeout 60 "claude build"  # Custom timeout
  claude-wm-cli --config ./custom.yaml status     # Use custom config
  claude-wm-cli --verbose execute "claude test"   # Verbose output

CONFIGURATION:
  Default config file: ~/.claude-wm-cli.yaml or ./.claude-wm-cli.yaml
  Environment variables: CLAUDE_WM_* (e.g., CLAUDE_WM_VERBOSE=true)`,
	Version: Version,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global persistent flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.claude-wm-cli.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().BoolVar(&debugMode, "debug", false, "debug output - shows all commands executed including Claude calls")

	// Bind flags to viper
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))
}

// initConfig reads in config file and ENV variables.
func initConfig() {
	// Validate config file if specified
	if cfgFile != "" {
		if err := model.ValidateConfigFile(cfgFile); err != nil {
			model.HandleValidationError(err, "")
			return
		}
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		if err != nil {
			cliErr := model.NewInternalError("failed to get user home directory").
				WithCause(err).
				WithSuggestions([]string{"Specify a config file explicitly with --config"})
			model.HandleValidationError(cliErr, "")
			return
		}

		// Search config in home directory with name ".claude-wm-cli" (without extension).
		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName(".claude-wm-cli")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		// Only show error if config file was explicitly specified
		if cfgFile != "" {
			cliErr := model.NewFileSystemError("read", cfgFile, err).
				WithSuggestions([]string{"Check that the config file exists and is valid YAML/JSON"})
			model.HandleValidationError(cliErr, "")
			return
		}
		// If no explicit config file, it's okay if default doesn't exist
	} else if verbose {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
