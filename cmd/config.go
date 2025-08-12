package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"claude-wm-cli/internal/config"
	"claude-wm-cli/internal/fsutil"
	"claude-wm-cli/internal/meta"
	wmmeta "claude-wm-cli/internal/wm/meta"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage claude-wm configuration",
	Long: `Configuration management for claude-wm using package manager approach.

Available subcommands:
  install  Install initial system configuration to .claude/ and .wm/baseline/
  init     Initialize new configuration workspace
  sync     Regenerate runtime configuration from templates and overrides
  upgrade  Update system templates (preserves user customizations)
  edit     Edit user configuration files
  show     Show effective runtime configuration`,
}

var configInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install initial system configuration",
	Long:  `Install initial system configuration to .claude/ and .wm/baseline/ directories from embedded templates`,
	RunE:  runConfigInstall,
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
	configCmd.AddCommand(configInstallCmd)
	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configSyncCmd)
	configCmd.AddCommand(configUpgradeCmd)
	configCmd.AddCommand(configShowCmd)
}

func runConfigInstall(cmd *cobra.Command, args []string) error {
	projectPath, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Check if already installed
	metaPath := filepath.Join(projectPath, ".wm", "meta.json")
	if _, err := os.Stat(metaPath); err == nil {
		return fmt.Errorf("configuration already installed (found %s)", metaPath)
	}

	fmt.Println("📦 Installing system configuration...")

	// Copy system configuration to .claude/
	claudePath := filepath.Join(projectPath, ".claude")
	fmt.Printf("   → Copying to %s\n", claudePath)
	if err := fsutil.CopyTreeFS(config.EmbeddedFS, "system", claudePath); err != nil {
		return fmt.Errorf("failed to copy configuration to .claude: %w", err)
	}

	// Copy system configuration to .wm/baseline/
	baselinePath := filepath.Join(projectPath, ".wm", "baseline")
	fmt.Printf("   → Copying to %s\n", baselinePath)
	if err := fsutil.CopyTreeFS(config.EmbeddedFS, "system", baselinePath); err != nil {
		return fmt.Errorf("failed to copy configuration to .wm/baseline: %w", err)
	}

	// Create .wm/meta.json
	fmt.Printf("   → Creating %s\n", metaPath)
	metaData := wmmeta.Default("claude-wm-cli", meta.Version)
	if err := wmmeta.Save(metaPath, metaData); err != nil {
		return fmt.Errorf("failed to create meta.json: %w", err)
	}

	fmt.Println("✅ System configuration installed successfully!")
	fmt.Println("")
	fmt.Printf("📁 Configuration installed to:\n")
	fmt.Printf("   %s        - System configuration\n", claudePath)
	fmt.Printf("   %s   - Baseline backup\n", baselinePath)
	fmt.Printf("   %s      - Installation metadata\n", metaPath)
	fmt.Println("")
	fmt.Println("💡 Next step: Run 'claude-wm-cli config init' to set up workspace")

	return nil
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