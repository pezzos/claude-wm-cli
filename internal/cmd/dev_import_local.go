package cmd

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"claude-wm-cli/internal/diff"
	"claude-wm-cli/internal/fsutil"

	"github.com/spf13/cobra"
)

var (
	importLocalOnly   []string
	importLocalDryRun bool
	importLocalApply  bool
)

// DevImportLocalCmd imports local changes from .claude/ to internal/config/system/
var DevImportLocalCmd = &cobra.Command{
	Use:   "import-local",
	Short: "Import changes from .claude/ to internal/config/system/",
	Long: `Import selective changes from .claude/ to internal/config/system/ with glob filtering.
	
This command supports selective copying with glob patterns and "apply by default" behavior:
- By default, changes are applied unless --dry-run is specified
- Use --only with glob patterns to filter specific files/directories
- Use --dry-run to preview changes without applying them
- Use --apply to explicitly apply changes (optional, default behavior)

Examples:
  dev import-local                           # Import all changes (apply by default)
  dev import-local --only "agents/**"        # Import only agent files
  dev import-local --only "hooks/**" --only "agents/**"  # Import hooks and agents
  dev import-local --dry-run                # Preview all changes without applying
  dev import-local --only "agents/**" --dry-run  # Preview only agent changes`,
	RunE: runDevImportLocal,
}

func init() {
	DevImportLocalCmd.Flags().StringArrayVar(&importLocalOnly, "only", []string{}, "Glob pattern to filter files (can be repeated)")
	DevImportLocalCmd.Flags().BoolVar(&importLocalDryRun, "dry-run", false, "Show planned changes without applying them")
	DevImportLocalCmd.Flags().BoolVar(&importLocalApply, "apply", false, "Apply changes (default behavior, optional flag)")
}

func runDevImportLocal(cmd *cobra.Command, args []string) error {
	projectPath, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Source and destination paths
	sourcePath := filepath.Join(projectPath, ".claude")
	destPath := filepath.Join(projectPath, "internal", "config", "system")

	// Verify source exists
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		return fmt.Errorf("source directory not found: %s", sourcePath)
	}

	// Verify destination exists
	if _, err := os.Stat(destPath); os.IsNotExist(err) {
		return fmt.Errorf("destination directory not found: %s", destPath)
	}

	fmt.Println("🔄 Analyzing local changes for import...")

	// Load source and destination filesystems
	sourceFS := os.DirFS(sourcePath)
	destFS := os.DirFS(destPath)

	// Calculate differences
	changes, err := diff.DiffTrees(sourceFS, ".", destFS, ".")
	if err != nil {
		return fmt.Errorf("failed to calculate differences: %w", err)
	}

	// Apply glob filters if specified
	if len(importLocalOnly) > 0 {
		changes = filterChangesByGlobs(changes, importLocalOnly)
	}

	// Separate into actionable changes (new and modified files only)
	var newFiles, modFiles []diff.Change
	for _, change := range changes {
		switch change.Type {
		case diff.ChangeNew:
			newFiles = append(newFiles, change)
		case diff.ChangeMod:
			modFiles = append(modFiles, change)
		// Skip deleted files - we don't delete from upstream in this version
		}
	}

	totalChanges := len(newFiles) + len(modFiles)

	if totalChanges == 0 {
		fmt.Println("✅ No changes to import")
		return nil
	}

	// Display plan
	fmt.Printf("📋 Import Plan\n")
	fmt.Printf("==============\n")
	fmt.Printf("New files: %d\n", len(newFiles))
	fmt.Printf("Modified files: %d\n", len(modFiles))
	fmt.Println()

	if len(newFiles) > 0 {
		fmt.Println("📄 New files to add:")
		for _, change := range newFiles {
			fmt.Printf("  + %s\n", change.Path)
		}
		fmt.Println()
	}

	if len(modFiles) > 0 {
		fmt.Println("📝 Modified files to update:")
		for _, change := range modFiles {
			fmt.Printf("  M %s\n", change.Path)
		}
		fmt.Println()
	}

	// Apply by default logic: apply unless --dry-run is specified
	shouldApply := !importLocalDryRun

	if importLocalDryRun {
		fmt.Printf("🔍 Dry-run mode: Would import %d changes\n", totalChanges)
		fmt.Println("💡 Run without --dry-run to apply these changes")
		return nil
	}

	if shouldApply {
		fmt.Printf("🚀 Applying %d changes...\n", totalChanges)

		// Copy new files
		for _, change := range newFiles {
			srcPath := filepath.Join(sourcePath, change.Path)
			dstPath := filepath.Join(destPath, change.Path)
			
			if err := fsutil.CopyFileWithDir(srcPath, dstPath); err != nil {
				return fmt.Errorf("failed to copy new file %s: %w", change.Path, err)
			}
			fmt.Printf("  ✓ Added %s\n", change.Path)
		}

		// Copy modified files
		for _, change := range modFiles {
			srcPath := filepath.Join(sourcePath, change.Path)
			dstPath := filepath.Join(destPath, change.Path)
			
			if err := fsutil.CopyFileWithDir(srcPath, dstPath); err != nil {
				return fmt.Errorf("failed to copy modified file %s: %w", change.Path, err)
			}
			fmt.Printf("  ✓ Updated %s\n", change.Path)
		}

		fmt.Println()
		fmt.Printf("✅ Successfully imported %d changes from .claude/ to internal/config/system/\n", totalChanges)
		
		// Regenerate manifest.json after successful import
		fmt.Println("🔄 Regenerating manifest.json...")
		if err := regenerateManifest(destPath); err != nil {
			fmt.Printf("⚠️  Warning: Failed to regenerate manifest: %v\n", err)
			// Don't fail the entire import for manifest issues
		} else {
			fmt.Println("   ✓ Manifest regenerated successfully")
		}
	}

	return nil
}

// filterChangesByGlobs filters changes based on glob patterns
func filterChangesByGlobs(changes []diff.Change, globs []string) []diff.Change {
	if len(globs) == 0 {
		return changes
	}

	var filtered []diff.Change
	for _, change := range changes {
		if matchesAnyGlob(change.Path, globs) {
			filtered = append(filtered, change)
		}
	}
	return filtered
}

// matchesAnyGlob checks if a path matches any of the provided glob patterns
func matchesAnyGlob(path string, globs []string) bool {
	for _, glob := range globs {
		// Convert glob to use forward slashes for consistency
		normalizedGlob := filepath.ToSlash(glob)
		normalizedPath := filepath.ToSlash(path)
		
		if matched, _ := filepath.Match(normalizedGlob, normalizedPath); matched {
			return true
		}
		
		// Handle ** patterns manually since filepath.Match doesn't support them
		if strings.Contains(normalizedGlob, "**") {
			if matchesDoubleStarGlob(normalizedPath, normalizedGlob) {
				return true
			}
		}
	}
	return false
}

// matchesDoubleStarGlob handles ** glob patterns
func matchesDoubleStarGlob(path, glob string) bool {
	// Simple implementation for ** patterns
	// Convert ** to a regex-like match
	
	if strings.HasPrefix(glob, "**/") {
		// **/ at start means any depth prefix
		suffix := glob[3:]
		if strings.HasSuffix(path, suffix) {
			return true
		}
		// Also check if any parent directory + suffix matches
		dirs := strings.Split(path, "/")
		for i := range dirs {
			if strings.HasSuffix(strings.Join(dirs[i:], "/"), suffix) {
				return true
			}
		}
	}
	
	if strings.HasSuffix(glob, "/**") {
		// /**  at end means any depth suffix
		prefix := glob[:len(glob)-3]
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	
	if glob == "**" {
		// ** alone matches everything
		return true
	}
	
	// Middle ** patterns: prefix/**/suffix
	parts := strings.Split(glob, "**")
	if len(parts) == 2 {
		prefix := strings.TrimSuffix(parts[0], "/")
		suffix := strings.TrimPrefix(parts[1], "/")
		
		if strings.HasPrefix(path, prefix) && strings.HasSuffix(path, suffix) {
			return true
		}
	}
	
	return false
}

// ManifestEntry represents a file entry in the manifest
type ManifestEntry struct {
	Path   string `json:"path"`
	Size   int64  `json:"size"`
	Sha256 string `json:"sha256"`
}

// regenerateManifest regenerates the manifest.json file for the given system directory
func regenerateManifest(systemPath string) error {
	var entries []ManifestEntry

	err := filepath.WalkDir(systemPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		
		// Skip directories
		if d.IsDir() {
			return nil
		}
		
		// Skip hooks directory (different Go module)
		if strings.Contains(path, "/hooks/") {
			return nil
		}
		
		// Skip manifest.json itself to avoid circular dependency
		if strings.HasSuffix(path, "manifest.json") {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", path, err)
		}

		hash := sha256.Sum256(content)
		relPath, err := filepath.Rel(systemPath, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path for %s: %w", path, err)
		}
		// Normalize path to use forward slashes for consistency
		relPath = filepath.ToSlash(relPath)
		// Prefix with "system/" to match the original manifest structure
		relPath = "system/" + relPath

		entries = append(entries, ManifestEntry{
			Path:   relPath,
			Size:   int64(len(content)),
			Sha256: hex.EncodeToString(hash[:]),
		})

		return nil
	})

	if err != nil {
		return fmt.Errorf("error walking directory: %w", err)
	}

	// Sort by path for stability
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Path < entries[j].Path
	})

	// Generate JSON with proper formatting
	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling JSON: %w", err)
	}

	// Write to manifest file
	manifestPath := filepath.Join(systemPath, "manifest.json")
	err = os.WriteFile(manifestPath, data, 0644)
	if err != nil {
		return fmt.Errorf("error writing manifest: %w", err)
	}

	return nil
}