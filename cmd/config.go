package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"claude-wm-cli/internal/config"
	"claude-wm-cli/internal/diff"
	"claude-wm-cli/internal/fsutil"
	"claude-wm-cli/internal/meta"
	"claude-wm-cli/internal/update"
	wmmeta "claude-wm-cli/internal/wm/meta"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage claude-wm configuration",
	Long: `Configuration management for claude-wm using package manager approach.

Available subcommands:
  install  Install initial system configuration to .claude/ and .wm/baseline/
  init     Initialize new configuration workspace
  status   Show configuration differences between upstream, baseline, and local
  update   Update configuration with 3-way merge (use --dry-run to preview)
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

var configStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show configuration differences",
	Long: `Show differences between:
- Upstream (embedded) vs Baseline (.wm/baseline) - changes since installation
- Baseline vs Local (.claude) - your local modifications`,
	RunE:  runConfigStatus,
}

var (
	updateDryRun bool
)

var configUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update configuration with 3-way merge",
	Long: `Update configuration using 3-way merge logic:
- Compares Upstream (embedded) vs Baseline (.wm/baseline) vs Local (.claude)
- Calculates merge actions: keep, apply, preserve_local, conflict, delete
- Use --dry-run to preview changes without applying them`,
	RunE: runConfigUpdate,
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
	configCmd.AddCommand(configStatusCmd)
	configCmd.AddCommand(configUpdateCmd)
	configCmd.AddCommand(configSyncCmd)
	configCmd.AddCommand(configUpgradeCmd)
	configCmd.AddCommand(configShowCmd)

	// Add flags for update command
	configUpdateCmd.Flags().BoolVar(&updateDryRun, "dry-run", false, "Show planned changes without applying them")
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

func runConfigStatus(cmd *cobra.Command, args []string) error {
	projectPath, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	fmt.Println("📊 Configuration Status")
	fmt.Println("======================")

	// Load the three filesystems
	upstream := config.EmbeddedFS
	
	baselinePath := filepath.Join(projectPath, ".wm", "baseline")
	if _, err := os.Stat(baselinePath); os.IsNotExist(err) {
		fmt.Println("❌ Baseline not found - run 'claude-wm-cli config install' first")
		return nil
	}
	baseline := os.DirFS(baselinePath)

	localPath := filepath.Join(projectPath, ".claude")  
	if _, err := os.Stat(localPath); os.IsNotExist(err) {
		fmt.Println("❌ Local configuration not found at .claude/")
		return nil
	}
	local := os.DirFS(localPath)

	// Compare Upstream vs Baseline
	fmt.Println("\n🔄 Upstream vs Baseline (changes since installation):")
	upstreamChanges, err := diff.DiffTrees(upstream, "system", baseline, ".")
	if err != nil {
		return fmt.Errorf("failed to diff upstream vs baseline: %w", err)
	}

	if len(upstreamChanges) == 0 {
		fmt.Println("   ✅ No changes")
	} else {
		for _, change := range upstreamChanges {
			fmt.Printf("   %s %s\n", getChangeSymbol(change.Type), change.Path)
		}
	}

	// Compare Baseline vs Local
	fmt.Println("\n📝 Baseline vs Local (your modifications):")
	localChanges, err := diff.DiffTrees(baseline, ".", local, ".")
	if err != nil {
		return fmt.Errorf("failed to diff baseline vs local: %w", err)
	}

	if len(localChanges) == 0 {
		fmt.Println("   ✅ No modifications")
	} else {
		for _, change := range localChanges {
			fmt.Printf("   %s %s\n", getChangeSymbol(change.Type), change.Path)
		}
	}

	return nil
}

// getChangeSymbol returns a visual symbol for each change type
func getChangeSymbol(changeType diff.ChangeType) string {
	switch changeType {
	case diff.ChangeNew:
		return "+"
	case diff.ChangeDel:
		return "-"
	case diff.ChangeMod:
		return "M"
	default:
		return "?"
	}
}

func runConfigUpdate(cmd *cobra.Command, args []string) error {
	projectPath, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Check if baseline exists
	baselinePath := filepath.Join(projectPath, ".wm", "baseline")
	if _, err := os.Stat(baselinePath); os.IsNotExist(err) {
		return fmt.Errorf("baseline not found at %s - run 'claude-wm-cli config install' first", baselinePath)
	}

	// Check if local configuration exists
	localPath := filepath.Join(projectPath, ".claude")
	if _, err := os.Stat(localPath); os.IsNotExist(err) {
		return fmt.Errorf("local configuration not found at %s", localPath)
	}

	fmt.Println("🔄 Calculating 3-way merge plan...")

	// Load the three filesystems
	upstream := config.EmbeddedFS
	baseline := os.DirFS(baselinePath)
	local := os.DirFS(localPath)

	// Build the update plan
	plan, err := update.BuildPlan(upstream, "system", baseline, ".", local, ".")
	if err != nil {
		return fmt.Errorf("failed to build update plan: %w", err)
	}

	if updateDryRun {
		// Show the plan without applying
		fmt.Println("📋 Update Plan (dry-run)")
		fmt.Println("========================")

		if len(plan.Merge) == 0 {
			fmt.Println("✅ No changes needed")
			return nil
		}

		// Display as JSON for now (can be enhanced with table format later)
		jsonData, err := json.MarshalIndent(plan, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format plan as JSON: %w", err)
		}

		fmt.Println(string(jsonData))
		fmt.Printf("\n💡 Run without --dry-run to apply %d changes\n", len(plan.Merge))
	} else {
		// TODO: Implement actual update logic (not in this task)
		fmt.Println("❌ Actual update not implemented yet - use --dry-run to preview changes")
		return fmt.Errorf("actual update implementation is not yet available")
	}

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