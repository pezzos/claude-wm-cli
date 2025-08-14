package cmd

import (
	"fmt"
	"io"
	"io/fs"
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

	// Generate .claude/settings.json if not exists (using canonical settings.json)
	settingsPath := filepath.Join(claudePath, "settings.json")
	if _, err := os.Stat(settingsPath); os.IsNotExist(err) {
		fmt.Printf("   → Generating %s\n", settingsPath)
		if err := copyEmbedFileToLocal(config.EmbeddedFS, "system/settings.json", settingsPath); err != nil {
			return fmt.Errorf("failed to copy canonical settings.json: %w", err)
		}
	} else {
		fmt.Printf("   ✓ %s already exists (skipping)\n", settingsPath)
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

// copyEmbedFileToLocal copies a single file from embedded FS to local file system
func copyEmbedFileToLocal(src fs.FS, srcPath, dstPath string) error {
	srcFile, err := src.Open(srcPath)
	if err != nil {
		return fmt.Errorf("failed to open source file %s: %w", srcPath, err)
	}
	defer srcFile.Close()

	if err := os.MkdirAll(filepath.Dir(dstPath), 0o755); err != nil {
		return fmt.Errorf("failed to create destination directory %s: %w", filepath.Dir(dstPath), err)
	}

	dstFile, err := os.OpenFile(dstPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return fmt.Errorf("failed to create destination file %s: %w", dstPath, err)
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("failed to copy content from %s to %s: %w", srcPath, dstPath, err)
	}

	return nil
}
