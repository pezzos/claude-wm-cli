package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"claude-wm-cli/internal/fsutil"
	"github.com/spf13/cobra"
)

var (
	resetSandbox bool
)

// DevSandboxCmd creates a testing sandbox from Upstream system files
var DevSandboxCmd = &cobra.Command{
	Use:   "sandbox",
	Short: "Create and manage testing sandbox from Upstream system files",
	Long: `Create and manage a testing sandbox in .wm/sandbox/claude/ from Upstream system files.

This command group provides sandbox management functionality including creation,
diff analysis, and selective upstreaming of changes.

The sandbox is particularly useful in SELF mode where .claude/ modifications
are restricted. All experimentation should be done within the sandbox.

Sandbox location: .wm/sandbox/claude/
Source: internal/config/system/

Subcommands:
  (none)    Create or reset the sandbox
  diff      Compare sandbox with source and selectively upstream changes`,
	RunE: runDevSandbox,
}

func init() {
	DevSandboxCmd.Flags().BoolVar(&resetSandbox, "reset", false, "Reset existing sandbox (removes existing directory)")
	
	// Add subcommands
	DevSandboxCmd.AddCommand(DevSandboxDiffCmd)
}

// runDevSandbox implements the dev sandbox logic
func runDevSandbox(cmd *cobra.Command, args []string) error {
	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Define paths
	sandboxPath := filepath.Join(cwd, ".wm", "sandbox", "claude")
	systemPath := filepath.Join(cwd, "internal", "config", "system")

	// Check if source system directory exists
	if _, err := os.Stat(systemPath); os.IsNotExist(err) {
		return fmt.Errorf("Upstream system directory not found: %s", systemPath)
	}

	// Check if sandbox already exists
	if _, err := os.Stat(sandboxPath); err == nil {
		if !resetSandbox {
			// Ask for confirmation to reset
			fmt.Printf("Sandbox already exists at: %s\n", sandboxPath)
			fmt.Printf("Do you want to reset it? This will remove all existing sandbox content. [y/N]: ")
			
			reader := bufio.NewReader(os.Stdin)
			response, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("failed to read input: %w", err)
			}
			
			response = strings.TrimSpace(strings.ToLower(response))
			if response != "y" && response != "yes" {
				fmt.Println("Operation cancelled. Use --reset flag to skip confirmation.")
				fmt.Println("\n💡 Try 'claude-wm dev sandbox diff' to see what has changed in your sandbox.")
				return nil
			}
		}

		// Remove existing sandbox
		fmt.Printf("🗑️  Removing existing sandbox at %s\n", sandboxPath)
		if err := os.RemoveAll(sandboxPath); err != nil {
			return fmt.Errorf("failed to remove existing sandbox: %w", err)
		}
	}

	// Create sandbox directory structure
	if err := fsutil.EnsureDir(sandboxPath); err != nil {
		return fmt.Errorf("failed to create sandbox directory: %w", err)
	}

	// Copy Upstream system files to sandbox
	fmt.Printf("📁 Creating sandbox from Upstream system files...\n")
	fmt.Printf("   Source: %s\n", systemPath)
	fmt.Printf("   Target: %s\n", sandboxPath)

	// Use OS filesystem for copying (not embedded FS)
	if err := copySystemToSandbox(systemPath, sandboxPath); err != nil {
		return fmt.Errorf("failed to copy system files to sandbox: %w", err)
	}

	// Display success message and usage instructions
	fmt.Printf("\n✅ Sandbox created successfully!\n\n")
	fmt.Printf("📋 Sandbox Details:\n")
	fmt.Printf("   Location: %s\n", sandboxPath)
	fmt.Printf("   Source:   %s\n", systemPath)
	fmt.Printf("\n💡 Usage Guidelines:\n")
	fmt.Printf("   • All experimentation should be done in the sandbox directory\n")
	fmt.Printf("   • The sandbox is isolated from your main .claude/ configuration\n")
	fmt.Printf("   • Use this environment to test new configurations safely\n")
	fmt.Printf("   • Changes in the sandbox do not affect your working environment\n")
	fmt.Printf("\n🚀 Next Steps:\n")
	fmt.Printf("   • Navigate to the sandbox: cd %s\n", sandboxPath)
	fmt.Printf("   • Make your experimental changes there\n")
	fmt.Printf("   • Test your modifications in isolation\n")
	fmt.Printf("   • Use 'claude-wm dev sandbox diff' to compare with source\n")
	fmt.Printf("   • Use 'claude-wm dev sandbox diff --apply' to upstream changes\n")

	return nil
}

// copySystemToSandbox copies files from the system directory to sandbox
func copySystemToSandbox(sourcePath, targetPath string) error {
	return filepath.Walk(sourcePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Calculate relative path
		relPath, err := filepath.Rel(sourcePath, path)
		if err != nil {
			return fmt.Errorf("failed to calculate relative path: %w", err)
		}

		// Calculate target path
		targetFilePath := filepath.Join(targetPath, relPath)

		if info.IsDir() {
			// Create directory
			return fsutil.EnsureDir(targetFilePath)
		}

		// Copy file
		return copyFile(path, targetFilePath)
	})
}

// copyFile copies a single file from source to destination
func copyFile(sourcePath, targetPath string) error {
	// Ensure target directory exists
	if err := fsutil.EnsureDir(filepath.Dir(targetPath)); err != nil {
		return err
	}

	// Read source file
	sourceData, err := os.ReadFile(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to read source file %s: %w", sourcePath, err)
	}

	// Get source file permissions
	sourceInfo, err := os.Stat(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to get source file info %s: %w", sourcePath, err)
	}

	// Write to target with same permissions
	if err := os.WriteFile(targetPath, sourceData, sourceInfo.Mode()); err != nil {
		return fmt.Errorf("failed to write target file %s: %w", targetPath, err)
	}

	return nil
}