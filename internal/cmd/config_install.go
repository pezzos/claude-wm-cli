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

// ConfigInstallCmd installs the embedded system configuration to .claude/ and .wm/baseline/.
var ConfigInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install initial system configuration",
	Long:  `Install initial system configuration to .claude/ and .wm/baseline/ directories from embedded templates`,
	RunE:  runConfigInstall,
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
	if err := fsutil.CopyTreeFromEmbed(config.EmbeddedFS, "system", claudePath); err != nil {
		return fmt.Errorf("failed to copy configuration to .claude: %w", err)
	}

	// Copy system configuration to .wm/baseline/
	baselinePath := filepath.Join(projectPath, ".wm", "baseline")
	fmt.Printf("   → Copying to %s\n", baselinePath)
	if err := fsutil.CopyTreeFromEmbed(config.EmbeddedFS, "system", baselinePath); err != nil {
		return fmt.Errorf("failed to copy configuration to .wm/baseline: %w", err)
	}

	// Create .wm/meta.json
	fmt.Printf("   → Creating %s\n", metaPath)
	metaData := wmmeta.Default("claude-wm-cli", meta.Version)
	if err := wmmeta.Save(metaPath, metaData); err != nil {
		return fmt.Errorf("failed to create meta.json: %w", err)
	}

	// Note: settings.json and all other files are already copied by CopyTreeFromEmbed above

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

