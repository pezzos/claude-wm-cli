package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"claude-wm-cli/internal/config"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage claude-wm configuration",
	Long: `Configuration management for claude-wm using package manager approach.

Available subcommands:
  init     Initialize new configuration workspace
  sync     Regenerate runtime configuration from templates and overrides
  upgrade  Update system templates (preserves user customizations)
  edit     Edit user configuration files
  show     Show effective runtime configuration`,
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize configuration workspace",
	Long:  `Initialize the .claude-wm workspace with package manager structure`,
	RunE:  runConfigInit,
}

var configSyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Regenerate runtime configuration",
	Long:  `Merge system templates and user overrides to generate runtime configuration`,
	RunE:  runConfigSync,
}

var configUpgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Update system templates",
	Long:  `Update system templates while preserving user customizations`,
	RunE:  runConfigUpgrade,
}

var configShowCmd = &cobra.Command{
	Use:   "show [file]",
	Short: "Show effective configuration",
	Long:  `Show the effective runtime configuration or a specific file`,
	RunE:  runConfigShow,
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configSyncCmd)
	configCmd.AddCommand(configUpgradeCmd)
	configCmd.AddCommand(configShowCmd)
}

func runConfigInit(cmd *cobra.Command, args []string) error {
	projectPath, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	manager := config.NewManager(projectPath)

	fmt.Println("🚀 Initializing claude-wm configuration workspace...")

	// Initialize directory structure
	if err := manager.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize workspace: %w", err)
	}

	// Install default templates
	if err := manager.InstallSystemTemplates(); err != nil {
		return fmt.Errorf("failed to install system templates: %w", err)
	}

	// Migrate from legacy structure if it exists
	legacyPath := filepath.Join(projectPath, ".claude-wm", ".claude")
	if err := manager.MigrateFromLegacy(legacyPath); err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	// Generate initial runtime configuration
	if err := manager.Sync(); err != nil {
		return fmt.Errorf("failed to generate runtime configuration: %w", err)
	}

	fmt.Println("✅ Configuration workspace initialized successfully!")
	fmt.Println("")
	fmt.Println("📁 Structure created:")
	fmt.Printf("   %s/system/    - System templates (read-only)\n", manager.WorkspaceRoot)
	fmt.Printf("   %s/user/      - Your customizations\n", manager.WorkspaceRoot)
	fmt.Printf("   %s/runtime/   - Effective configuration\n", manager.WorkspaceRoot)
	fmt.Println("")
	fmt.Println("💡 Next steps:")
	fmt.Println("   - Edit user configurations: claude-wm config edit")
	fmt.Println("   - View effective config: claude-wm config show")

	return nil
}

func runConfigSync(cmd *cobra.Command, args []string) error {
	projectPath, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	manager := config.NewManager(projectPath)

	fmt.Println("🔄 Syncing configuration...")

	if err := manager.Sync(); err != nil {
		return fmt.Errorf("sync failed: %w", err)
	}

	fmt.Println("✅ Configuration synced successfully!")
	return nil
}

func runConfigUpgrade(cmd *cobra.Command, args []string) error {
	projectPath, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	manager := config.NewManager(projectPath)

	fmt.Println("⬆️  Upgrading system templates...")

	// Reinstall system templates (this updates defaults without touching user files)
	if err := manager.InstallSystemTemplates(); err != nil {
		return fmt.Errorf("failed to upgrade system templates: %w", err)
	}

	// Regenerate runtime configuration
	if err := manager.Sync(); err != nil {
		return fmt.Errorf("failed to sync after upgrade: %w", err)
	}

	fmt.Println("✅ System templates upgraded successfully!")
	fmt.Println("💡 Your user customizations have been preserved")
	return nil
}

func runConfigShow(cmd *cobra.Command, args []string) error {
	projectPath, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	manager := config.NewManager(projectPath)

	if len(args) == 0 {
		// Show overview
		fmt.Println("📋 Configuration Overview:")
		fmt.Println("")
		
		// Show directory status
		showDirStatus("System", manager.SystemPath)
		showDirStatus("User", manager.UserPath)
		showDirStatus("Runtime", manager.RuntimePath)
		
		return nil
	}

	// Show specific file
	fileName := args[0]
	runtimeFile := manager.GetRuntimePath(fileName)
	
	if _, err := os.Stat(runtimeFile); os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s", fileName)
	}

	data, err := os.ReadFile(runtimeFile)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	fmt.Printf("📄 %s (runtime):\n", fileName)
	fmt.Println(string(data))
	
	return nil
}

func showDirStatus(name, path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		fmt.Printf("   %s: ❌ Not found\n", name)
		return
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		fmt.Printf("   %s: ❌ Error reading\n", name)
		return
	}

	fmt.Printf("   %s: ✅ %d items\n", name, len(entries))
}