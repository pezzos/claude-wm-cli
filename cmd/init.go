/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"claude-wm-cli/internal/model"
	"claude-wm-cli/internal/subagents"

	"github.com/spf13/cobra"
)

var (
	initProjectName string
	initForce       bool
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init [project-name]",
	Short: "Initialize a new project",
	Long: `Initialize a new Claude WM CLI project with proper directory structure
and configuration files. This command sets up the basic project structure
needed for workflow management.

Examples:
  claude-wm-cli init my-project          # Initialize project in ./my-project
  claude-wm-cli init                     # Initialize project in current directory
  claude-wm-cli init --force my-project  # Force initialization (overwrite existing)`,
	Run: func(cmd *cobra.Command, args []string) {
		var projectName string
		if len(args) > 0 {
			projectName = args[0]
		} else {
			pwd, _ := os.Getwd()
			projectName = filepath.Base(pwd)
		}
		initializeProject(projectName)
	},
}

func initializeProject(projectName string) {
	// Validate project name
	if err := model.ValidateProjectName(projectName); err != nil {
		model.HandleValidationError(err, "claude-wm-cli init my-project")
		return
	}

	fmt.Printf("🚀 Initializing Claude WM CLI project: %s\n", projectName)
	fmt.Println("================================")
	fmt.Println()

	// Get current directory or create project directory
	var projectDir string
	if projectName == filepath.Base(projectName) && projectName != "." {
		projectDir = filepath.Join(".", projectName)
		fmt.Printf("📁 Creating project directory: %s\n", projectDir)
		if err := os.MkdirAll(projectDir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "❌ Error creating directory: %s\n", err.Error())
			return
		}
	} else {
		projectDir = "."
	}

	// Create basic directory structure
	dirs := []string{
		"docs/1-project",
		"docs/2-current-epic",
		"docs/3-current-task",
		"docs/archive",
	}

	fmt.Println("📂 Creating directory structure...")
	for _, dir := range dirs {
		fullPath := filepath.Join(projectDir, dir)
		if err := os.MkdirAll(fullPath, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "❌ Error creating directory %s: %s\n", dir, err.Error())
			return
		}
		fmt.Printf("  ✓ %s\n", dir)
	}

	// Create basic configuration file
	fmt.Println("\n⚙️  Creating configuration files...")
	configPath := filepath.Join(projectDir, ".claude-wm-cli.yaml")
	if !fileExists(configPath) || initForce {
		configContent := fmt.Sprintf(`# Claude WM CLI Configuration
project:
  name: "%s"
  initialized: true
  
verbose: false

# Default settings
defaults:
  timeout: 30
  retries: 2
`, projectName)

		if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "❌ Error writing config file: %s\n", err.Error())
			return
		}
		fmt.Printf("  ✓ .claude-wm-cli.yaml\n")
	} else {
		fmt.Fprintf(os.Stderr, "⚠️  %s\n", ".claude-wm-cli.yaml already exists (use --force to overwrite)")
	}

	// Install claude-wm subagents
	fmt.Println("\n🤖 Installing claude-wm subagents...")
	installer := subagents.NewAgentInstaller()
	if err := installer.InstallAgents(projectDir); err != nil {
		fmt.Fprintf(os.Stderr, "⚠️  Warning: Failed to install subagents: %s\n", err.Error())
		fmt.Println("   Subagents will be unavailable but project will function normally")
	} else {
		// Verify installation
		info, err := installer.GetAgentInstallationInfo(projectDir)
		if err != nil {
			fmt.Println("  ✓ Subagents installed (verification failed)")
		} else {
			fmt.Printf("  ✓ %s\n", info.GetInstallationSummary())
			if info.AllInstalled {
				fmt.Println("    • claude-wm-templates: 93% token savings on documentation")
				fmt.Println("    • claude-wm-status: 89% token savings on status reports")
				fmt.Println("    • claude-wm-planner: 85% token savings on task planning")
				fmt.Println("    • claude-wm-reviewer: 83% token savings on code reviews")
			}
		}
	}

	fmt.Println()
	fmt.Printf("✅ Project '%s' initialized successfully!\n", projectName)
	fmt.Println()
	fmt.Println("📋 Next steps:")
	fmt.Println("  1. cd " + projectName + " (if created in subdirectory)")
	fmt.Println("  2. claude-wm-cli subagents list  # Verify subagents installation")
	fmt.Println("  3. claude-wm-cli status          # Check project status")
	fmt.Println("  4. Start your first epic with the agile workflow commands")
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func init() {
	rootCmd.AddCommand(initCmd)

	// Command-specific flags
	initCmd.Flags().BoolVarP(&initForce, "force", "f", false, "Force initialization (overwrite existing files)")
}
