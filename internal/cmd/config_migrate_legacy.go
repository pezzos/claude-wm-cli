package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"claude-wm-cli/internal/diff"
	"claude-wm-cli/internal/fsutil"
	wmmeta "claude-wm-cli/internal/wm/meta"
	"github.com/spf13/cobra"
)

// Flags for the migrate-legacy command
var (
	migrateDryRun bool
	migrateArchive bool
)

// MigrationReport represents the result of a migration analysis
type MigrationReport struct {
	LegacyPath    string                   `json:"legacy_path"`
	TargetPath    string                   `json:"target_path"`
	Actions       []MigrationAction        `json:"actions"`
	Summary       MigrationSummary         `json:"summary"`
	Conflicts     []string                 `json:"conflicts,omitempty"`
	AnalyzedFiles map[string]FileAnalysis  `json:"analyzed_files"`
}

// MigrationAction represents a single migration action
type MigrationAction struct {
	Type        string `json:"type"`        // "copy", "convert", "ignore", "archive"
	SourcePath  string `json:"source_path"`
	TargetPath  string `json:"target_path,omitempty"`
	Description string `json:"description"`
	Status      string `json:"status"` // "planned", "completed", "failed", "skipped"
}

// MigrationSummary provides a summary of the migration
type MigrationSummary struct {
	FilesAnalyzed int `json:"files_analyzed"`
	FilesToCopy   int `json:"files_to_copy"`
	FilesToIgnore int `json:"files_to_ignore"`
	Conflicts     int `json:"conflicts"`
	EstimatedSize int64 `json:"estimated_size_bytes"`
}

// FileAnalysis contains analysis results for a specific file
type FileAnalysis struct {
	Path        string `json:"path"`
	Type        string `json:"type"`        // "config", "template", "hook", "backup", "cache", "unknown"
	Category    string `json:"category"`    // "system", "user", "runtime", "meta", "other"
	Size        int64  `json:"size"`
	Modified    time.Time `json:"modified"`
	Disposition string `json:"disposition"` // "migrate", "ignore", "convert"
	Reason      string `json:"reason"`
}

// ConfigMigrateLegacyCmd migrates from .claude-wm to .wm structure
var ConfigMigrateLegacyCmd = &cobra.Command{
	Use:   "migrate-legacy",
	Short: "Migrate from legacy .claude-wm to new .wm structure",
	Long: `Migrate from the legacy .claude-wm directory structure to the new .wm structure.

This command analyzes your existing .claude-wm directory and selectively migrates
relevant content to the new .wm workspace structure:

- System configuration baseline → .wm/baseline/
- User customizations → .wm/meta.json and user configs
- Hooks and templates → appropriate .wm/ locations
- Runtime files → ignored (will be regenerated)
- Cache and temporary files → ignored

The migration preserves your customizations while moving to the new structure.

Examples:
  # Analyze migration plan
  claude-wm config migrate-legacy

  # Preview migration without applying
  claude-wm config migrate-legacy --dry-run

  # Apply migration and archive old directory
  claude-wm config migrate-legacy --archive`,
	RunE: runConfigMigrateLegacy,
}

func init() {
	ConfigMigrateLegacyCmd.Flags().BoolVar(&migrateDryRun, "dry-run", false, "Show migration plan without applying changes")
	ConfigMigrateLegacyCmd.Flags().BoolVar(&migrateArchive, "archive", false, "Rename .claude-wm to .claude-wm.bak after successful migration")
}

// runConfigMigrateLegacy implements the legacy migration logic
func runConfigMigrateLegacy(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	legacyPath := filepath.Join(cwd, ".claude-wm")
	targetPath := filepath.Join(cwd, ".wm")

	fmt.Printf("🔍 Legacy Configuration Migration Analysis\n")
	fmt.Printf("═══════════════════════════════════════════\n\n")

	// Step 1: Check if legacy directory exists
	if _, err := os.Stat(legacyPath); os.IsNotExist(err) {
		fmt.Printf("✅ No legacy .claude-wm directory found.\n")
		fmt.Printf("💡 Your project is already using the current structure.\n")
		return nil
	}

	fmt.Printf("📁 Legacy directory detected: %s\n", legacyPath)

	// Step 2: Check if target directory exists and is not empty
	if err := validateTargetDirectory(targetPath); err != nil {
		return err
	}

	// Step 3: Analyze legacy structure
	fmt.Printf("🔍 Analyzing legacy structure...\n")
	report, err := analyzeLegacyStructure(legacyPath, targetPath)
	if err != nil {
		return fmt.Errorf("failed to analyze legacy structure: %w", err)
	}

	// Step 4: Display analysis results
	if err := displayMigrationReport(report, migrateDryRun); err != nil {
		return fmt.Errorf("failed to display migration report: %w", err)
	}

	// Step 5: Check for conflicts
	if len(report.Conflicts) > 0 {
		fmt.Printf("\n⚠️  Migration Conflicts Detected:\n")
		for _, conflict := range report.Conflicts {
			fmt.Printf("   • %s\n", conflict)
		}
		fmt.Printf("\n💡 Please resolve conflicts before proceeding.\n")
		return fmt.Errorf("migration cannot proceed due to conflicts")
	}

	// Step 6: If dry-run, stop here
	if migrateDryRun {
		fmt.Printf("\n💡 This was a dry run. Use --dry-run=false to apply the migration.\n")
		return nil
	}

	// Step 7: Confirm migration if not in dry-run mode
	if !confirmMigration(report) {
		fmt.Printf("Migration cancelled.\n")
		return nil
	}

	// Step 8: Apply migration
	fmt.Printf("\n🚀 Applying Migration...\n")
	fmt.Printf("═══════════════════════\n\n")

	if err := applyMigration(report); err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	// Step 9: Archive legacy directory if requested
	if migrateArchive {
		fmt.Printf("📦 Archiving legacy directory...\n")
		archivePath := legacyPath + ".bak"
		if err := os.Rename(legacyPath, archivePath); err != nil {
			fmt.Printf("⚠️  Warning: Failed to archive legacy directory: %v\n", err)
		} else {
			fmt.Printf("   ✓ Archived to: %s\n", archivePath)
		}
	}

	// Step 10: Success message
	fmt.Printf("\n🎉 Migration Completed Successfully!\n\n")
	fmt.Printf("📋 Migration Summary:\n")
	fmt.Printf("   • %d files migrated\n", report.Summary.FilesToCopy)
	fmt.Printf("   • %d files ignored\n", report.Summary.FilesToIgnore)
	fmt.Printf("   • Legacy structure: %s\n", getArchiveStatus(legacyPath, migrateArchive))
	fmt.Printf("   • New structure: %s\n", targetPath)

	fmt.Printf("\n💡 Next Steps:\n")
	fmt.Printf("   • Review the migrated configuration\n")
	fmt.Printf("   • Test your setup with: claude-wm config status\n")
	fmt.Printf("   • Remove legacy backup when satisfied: rm -rf %s.bak\n", legacyPath)

	return nil
}

// validateTargetDirectory checks if the target directory is valid for migration
func validateTargetDirectory(targetPath string) error {
	if stat, err := os.Stat(targetPath); err == nil {
		if stat.IsDir() {
			// Check if directory is empty
			entries, err := os.ReadDir(targetPath)
			if err != nil {
				return fmt.Errorf("failed to read target directory: %w", err)
			}
			if len(entries) > 0 {
				return fmt.Errorf("target directory %s already exists and is not empty\nPlease remove it or use a different location", targetPath)
			}
			fmt.Printf("📁 Target directory exists but is empty: %s\n", targetPath)
		} else {
			return fmt.Errorf("target path %s exists but is not a directory", targetPath)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to check target directory: %w", err)
	}

	return nil
}

// analyzeLegacyStructure analyzes the legacy directory structure
func analyzeLegacyStructure(legacyPath, targetPath string) (*MigrationReport, error) {
	report := &MigrationReport{
		LegacyPath:    legacyPath,
		TargetPath:    targetPath,
		Actions:       []MigrationAction{},
		AnalyzedFiles: make(map[string]FileAnalysis),
		Conflicts:     []string{},
	}

	err := filepath.Walk(legacyPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip the root directory
		if path == legacyPath {
			return nil
		}

		// Calculate relative path
		relPath, err := filepath.Rel(legacyPath, path)
		if err != nil {
			return err
		}

		// Skip directories in file analysis (but continue walking)
		if info.IsDir() {
			return nil
		}

		// Analyze the file
		analysis := analyzeFile(relPath, info)
		report.AnalyzedFiles[relPath] = analysis
		report.Summary.FilesAnalyzed++
		report.Summary.EstimatedSize += info.Size()

		// Determine migration action based on analysis
		action := determineMigrationAction(relPath, analysis, targetPath)
		report.Actions = append(report.Actions, action)

		// Update summary counters
		switch action.Type {
		case "copy", "convert":
			report.Summary.FilesToCopy++
		case "ignore":
			report.Summary.FilesToIgnore++
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return report, nil
}

// analyzeFile analyzes a single file and determines its type and category
func analyzeFile(relPath string, info os.FileInfo) FileAnalysis {
	analysis := FileAnalysis{
		Path:     relPath,
		Size:     info.Size(),
		Modified: info.ModTime(),
	}

	// Determine file type based on extension and path
	ext := strings.ToLower(filepath.Ext(relPath))
	dir := strings.ToLower(filepath.Dir(relPath))

	switch {
	case ext == ".json":
		analysis.Type = "config"
	case ext == ".md":
		analysis.Type = "template"
	case ext == ".sh":
		analysis.Type = "hook"
	case strings.Contains(relPath, "backup"):
		analysis.Type = "backup"
	case strings.Contains(relPath, "cache"):
		analysis.Type = "cache"
	default:
		analysis.Type = "unknown"
	}

	// Determine category based on directory structure
	switch {
	case strings.HasPrefix(dir, "system"):
		analysis.Category = "system"
	case strings.HasPrefix(dir, "user"):
		analysis.Category = "user"
	case strings.HasPrefix(dir, "runtime"):
		analysis.Category = "runtime"
	case relPath == "meta.json" || relPath == "config.json":
		analysis.Category = "meta"
	default:
		analysis.Category = "other"
	}

	// Determine disposition
	switch analysis.Category {
	case "system":
		analysis.Disposition = "migrate"
		analysis.Reason = "System configuration should be preserved as baseline"
	case "user":
		analysis.Disposition = "migrate"
		analysis.Reason = "User customizations should be preserved"
	case "meta":
		analysis.Disposition = "convert"
		analysis.Reason = "Metadata needs format conversion"
	case "runtime":
		analysis.Disposition = "ignore"
		analysis.Reason = "Runtime files will be regenerated"
	default:
		if analysis.Type == "cache" || analysis.Type == "backup" {
			analysis.Disposition = "ignore"
			analysis.Reason = "Cache/backup files not needed in new structure"
		} else {
			analysis.Disposition = "migrate"
			analysis.Reason = "Preserve unknown files for safety"
		}
	}

	return analysis
}

// determineMigrationAction determines what action to take for a specific file
func determineMigrationAction(relPath string, analysis FileAnalysis, targetPath string) MigrationAction {
	action := MigrationAction{
		SourcePath: relPath,
		Status:     "planned",
	}

	switch analysis.Disposition {
	case "migrate":
		action.Type = "copy"
		action.TargetPath = mapLegacyPathToNew(relPath, analysis)
		action.Description = fmt.Sprintf("Copy %s to new structure", analysis.Type)

	case "convert":
		action.Type = "convert"
		action.TargetPath = mapLegacyPathToNew(relPath, analysis)
		action.Description = fmt.Sprintf("Convert %s to new format", analysis.Type)

	case "ignore":
		action.Type = "ignore"
		action.Description = fmt.Sprintf("Skip %s (%s)", analysis.Type, analysis.Reason)

	default:
		action.Type = "ignore"
		action.Description = "Unknown file type, skipping for safety"
	}

	return action
}

// mapLegacyPathToNew maps legacy paths to new structure paths
func mapLegacyPathToNew(relPath string, analysis FileAnalysis) string {
	switch analysis.Category {
	case "system":
		// Map system files to baseline
		return strings.Replace(relPath, "system/", "baseline/", 1)
	case "user":
		// Keep user files in similar structure but maybe adapt
		return relPath
	case "meta":
		// Meta files go to root of .wm
		if filepath.Base(relPath) == "meta.json" {
			return "meta.json"
		}
		return relPath
	default:
		// Other files keep relative structure
		return relPath
	}
}

// displayMigrationReport displays the migration analysis report
func displayMigrationReport(report *MigrationReport, isDryRun bool) error {
	fmt.Printf("\n📊 Migration Analysis Report\n")
	fmt.Printf("═══════════════════════════\n\n")

	// Display summary
	fmt.Printf("📋 Summary:\n")
	fmt.Printf("   • Files analyzed: %d\n", report.Summary.FilesAnalyzed)
	fmt.Printf("   • Files to migrate: %d\n", report.Summary.FilesToCopy)
	fmt.Printf("   • Files to ignore: %d\n", report.Summary.FilesToIgnore)
	fmt.Printf("   • Estimated size: %.2f MB\n", float64(report.Summary.EstimatedSize)/(1024*1024))

	if len(report.Actions) > 0 {
		fmt.Printf("\n📁 Planned Actions:\n")
		
		// Group actions by type
		actionsByType := make(map[string][]MigrationAction)
		for _, action := range report.Actions {
			actionsByType[action.Type] = append(actionsByType[action.Type], action)
		}

		// Display each group
		for actionType, actions := range actionsByType {
			if len(actions) == 0 {
				continue
			}

			icon := getActionIcon(actionType)
			fmt.Printf("\n   %s %s (%d files):\n", icon, strings.Title(actionType), len(actions))

			// Show up to 10 files per type, then summarize
			shown := 0
			for _, action := range actions {
				if shown < 10 {
					if action.TargetPath != "" {
						fmt.Printf("      %s → %s\n", action.SourcePath, action.TargetPath)
					} else {
						fmt.Printf("      %s\n", action.SourcePath)
					}
					shown++
				}
			}
			if len(actions) > 10 {
				fmt.Printf("      ... and %d more files\n", len(actions)-10)
			}
		}
	}

	return nil
}

// getActionIcon returns an icon for the action type
func getActionIcon(actionType string) string {
	switch actionType {
	case "copy":
		return "📄"
	case "convert":
		return "🔄"
	case "ignore":
		return "⏭️"
	case "archive":
		return "📦"
	default:
		return "❓"
	}
}

// confirmMigration asks user to confirm the migration
func confirmMigration(report *MigrationReport) bool {
	fmt.Printf("\n❓ Confirm Migration\n")
	fmt.Printf("═══════════════════\n\n")
	fmt.Printf("This will migrate %d files from .claude-wm to .wm structure.\n", report.Summary.FilesToCopy)
	fmt.Printf("Continue with migration? [y/N]: ")

	var response string
	fmt.Scanln(&response)
	response = strings.ToLower(strings.TrimSpace(response))

	return response == "y" || response == "yes"
}

// applyMigration applies the migration plan
func applyMigration(report *MigrationReport) error {
	// Create target directory structure
	if err := fsutil.EnsureDir(report.TargetPath); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	// Apply actions
	for i, action := range report.Actions {
		if action.Type == "ignore" {
			continue
		}

		sourcePath := filepath.Join(report.LegacyPath, action.SourcePath)
		targetPath := filepath.Join(report.TargetPath, action.TargetPath)

		var err error
		switch action.Type {
		case "copy":
			err = fsutil.CopyFileWithDir(sourcePath, targetPath)
		case "convert":
			err = fsutil.CopyFileWithDir(sourcePath, targetPath)
		default:
			continue
		}

		if err != nil {
			fmt.Printf("   ❌ Failed: %s (%v)\n", action.SourcePath, err)
			report.Actions[i].Status = "failed"
		} else {
			fmt.Printf("   ✓ %s %s\n", getActionIcon(action.Type), action.SourcePath)
			report.Actions[i].Status = "completed"
		}
	}

	// Create meta.json in new structure
	metaPath := filepath.Join(report.TargetPath, "meta.json")
	metaData := wmmeta.Default("claude-wm-cli", "migrated")
	if err := wmmeta.Save(metaPath, metaData); err != nil {
		return fmt.Errorf("failed to create meta.json: %w", err)
	}

	return nil
}


// getArchiveStatus returns a human-readable status of the archive operation
func getArchiveStatus(legacyPath string, archived bool) string {
	if archived {
		if _, err := os.Stat(legacyPath + ".bak"); err == nil {
			return "archived to " + legacyPath + ".bak"
		} else {
			return "archive failed, original preserved"
		}
	}
	return "preserved at " + legacyPath
}