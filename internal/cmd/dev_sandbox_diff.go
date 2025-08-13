package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"claude-wm-cli/internal/diff"
	"claude-wm-cli/internal/fsutil"
	"github.com/spf13/cobra"
)

// Flags for the diff command
var (
	diffApply       bool
	diffDryRun      bool
	diffAllowDelete bool
	diffOnlyPattern []string
)

// DiffResult represents the result of a diff operation
type DiffResult struct {
	Plan    []DiffAction `json:"plan"`
	Summary DiffSummary  `json:"summary"`
}

// DiffAction represents a single action to be taken
type DiffAction struct {
	Path   string `json:"path"`
	Action string `json:"action"` // "new", "mod", "del"
	Status string `json:"status"` // "planned", "applied", "skipped"
}

// DiffSummary provides a summary of the diff operation
type DiffSummary struct {
	New       int `json:"new"`
	Modified  int `json:"modified"`
	Deleted   int `json:"deleted"`
	Total     int `json:"total"`
	Applied   int `json:"applied,omitempty"`
	Skipped   int `json:"skipped,omitempty"`
}

// DevSandboxDiffCmd compares sandbox with source and allows selective upstreaming
var DevSandboxDiffCmd = &cobra.Command{
	Use:   "diff",
	Short: "Compare sandbox with source and selectively upstream changes",
	Long: `Compare the sandbox directory (.wm/sandbox/claude/) with the source
directory (internal/config/system/) and apply changes by default.

This command helps you selectively upstream changes from your experimental
sandbox environment back to the main source tree.

By default, applies all changes automatically. Use --dry-run to see what would
be changed without applying anything.

Examples:
  # Apply all changes (default behavior)
  claude-wm dev sandbox diff

  # Show what would be applied without doing it
  claude-wm dev sandbox diff --dry-run

  # Apply only agent-related changes
  claude-wm dev sandbox diff --only "agents/**"

  # Apply changes including deletions
  claude-wm dev sandbox diff --allow-delete

  # Dry run with specific patterns
  claude-wm dev sandbox diff --dry-run --only "agents/**"`,
	RunE: runDevSandboxDiff,
}

func init() {
	DevSandboxDiffCmd.Flags().BoolVar(&diffApply, "apply", false, "Explicitly apply changes (default behavior, this flag is optional)")
	DevSandboxDiffCmd.Flags().BoolVar(&diffDryRun, "dry-run", false, "Show what would be done without actually doing it")
	DevSandboxDiffCmd.Flags().BoolVar(&diffAllowDelete, "allow-delete", false, "Include file deletions in the changes")
	DevSandboxDiffCmd.Flags().StringArrayVar(&diffOnlyPattern, "only", []string{}, "Include only files matching these glob patterns (can be specified multiple times)")
}

// runDevSandboxDiff implements the dev sandbox diff logic
func runDevSandboxDiff(cmd *cobra.Command, args []string) error {
	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Define paths
	sandboxPath := filepath.Join(cwd, ".wm", "sandbox", "claude")
	systemPath := filepath.Join(cwd, "internal", "config", "system")

	// Check if sandbox exists
	if _, err := os.Stat(sandboxPath); os.IsNotExist(err) {
		return fmt.Errorf("sandbox directory not found: %s\nRun 'claude-wm dev sandbox' first to create it", sandboxPath)
	}

	// Check if source system directory exists
	if _, err := os.Stat(systemPath); os.IsNotExist(err) {
		return fmt.Errorf("source system directory not found: %s", systemPath)
	}

	// Perform diff analysis
	sandboxFS := os.DirFS(sandboxPath)
	systemFS := os.DirFS(systemPath)

	changes, err := diff.DiffTrees(sandboxFS, ".", systemFS, ".")
	if err != nil {
		return fmt.Errorf("failed to compare directories: %w", err)
	}

	// Filter changes based on --only patterns
	if len(diffOnlyPattern) > 0 {
		changes = filterChangesByPattern(changes, diffOnlyPattern)
	}

	// Create diff result
	result := createDiffResult(changes, diffAllowDelete)

	// Apply by default behavior:
	// - If --dry-run is specified: show what would be applied (don't actually apply)
	// - Otherwise: apply the changes (default behavior)
	
	if diffDryRun {
		// Dry-run mode: show what would be applied without actually doing it
		return showDryRunPlan(result)
	}
	
	// Default behavior: apply the changes
	return applyDiffChanges(result, sandboxPath, systemPath)
}

// filterChangesByPattern filters changes based on glob patterns
func filterChangesByPattern(changes []diff.Change, patterns []string) []diff.Change {
	if len(patterns) == 0 {
		return changes
	}

	var filtered []diff.Change
	for _, change := range changes {
		for _, pattern := range patterns {
			matched, err := filepath.Match(pattern, change.Path)
			if err != nil {
				// If pattern is invalid, skip it
				continue
			}
			if matched {
				filtered = append(filtered, change)
				break // Found match, no need to check other patterns
			}
			// Try Unix-style glob matching for patterns like "agents/**"
			if strings.Contains(pattern, "**") {
				if matchesGlobPattern(change.Path, pattern) {
					filtered = append(filtered, change)
					break
				}
			}
		}
	}

	return filtered
}

// matchesGlobPattern provides enhanced ** glob matching
func matchesGlobPattern(path, pattern string) bool {
	// Handle ** patterns more comprehensively
	if strings.Contains(pattern, "**") {
		// Pattern like "prefix/**" - matches anything under prefix/
		if strings.HasSuffix(pattern, "/**") {
			prefix := strings.TrimSuffix(pattern, "/**")
			return strings.HasPrefix(path, prefix+"/") || path == prefix
		}
		
		// Pattern like "**/suffix" - matches suffix anywhere in tree
		if strings.HasPrefix(pattern, "**/") {
			suffix := strings.TrimPrefix(pattern, "**/")
			return strings.HasSuffix(path, "/"+suffix) || path == suffix
		}
		
		// Pattern like "prefix/**/suffix" - matches prefix...suffix with anything in between
		if strings.Contains(pattern, "/**/") {
			parts := strings.SplitN(pattern, "/**/", 2)
			if len(parts) == 2 {
				prefix := parts[0]
				suffix := parts[1]
				return (strings.HasPrefix(path, prefix+"/") || path == prefix) && 
					   (strings.HasSuffix(path, "/"+suffix) || strings.HasSuffix(path, suffix))
			}
		}
		
		// Pattern with just ** (matches everything)
		if pattern == "**" {
			return true
		}
	}
	return false
}

// createDiffResult creates a structured diff result from changes
func createDiffResult(changes []diff.Change, allowDelete bool) DiffResult {
	var actions []DiffAction
	summary := DiffSummary{}

	for _, change := range changes {
		action := DiffAction{
			Path:   change.Path,
			Action: string(change.Type),
			Status: "planned",
		}

		// Skip delete actions if not allowed
		if change.Type == diff.ChangeDel && !allowDelete {
			action.Status = "skipped"
			summary.Skipped++
		} else {
			switch change.Type {
			case diff.ChangeNew:
				summary.New++
			case diff.ChangeMod:
				summary.Modified++
			case diff.ChangeDel:
				summary.Deleted++
			}
		}

		actions = append(actions, action)
		summary.Total++
	}

	return DiffResult{
		Plan:    actions,
		Summary: summary,
	}
}

// showDiffPlan displays the diff plan in a readable format
func showDiffPlan(result DiffResult) error {
	fmt.Printf("📊 Sandbox → Source Diff Analysis\n")
	fmt.Printf("═══════════════════════════════════\n\n")

	if result.Summary.Total == 0 {
		fmt.Printf("✅ No differences found between sandbox and source.\n")
		return nil
	}

	// Print summary
	fmt.Printf("📋 Summary:\n")
	if result.Summary.New > 0 {
		fmt.Printf("   📄 New files:      %d\n", result.Summary.New)
	}
	if result.Summary.Modified > 0 {
		fmt.Printf("   📝 Modified files: %d\n", result.Summary.Modified)
	}
	if result.Summary.Deleted > 0 {
		fmt.Printf("   🗑️  Deleted files:  %d\n", result.Summary.Deleted)
	}
	if result.Summary.Skipped > 0 {
		fmt.Printf("   ⏭️  Skipped files:  %d (use --allow-delete to include deletions)\n", result.Summary.Skipped)
	}
	fmt.Printf("   📊 Total changes:  %d\n\n", result.Summary.Total)

	// Print detailed plan
	fmt.Printf("📋 Detailed Plan:\n")
	for _, action := range result.Plan {
		var icon string
		switch action.Action {
		case "new":
			icon = "📄"
		case "mod":
			icon = "📝"
		case "del":
			icon = "🗑️"
		default:
			icon = "❓"
		}

		status := ""
		if action.Status == "skipped" {
			status = " (skipped)"
		}

		fmt.Printf("   %s %-3s %s%s\n", icon, action.Action, action.Path, status)
	}

	fmt.Printf("\n💡 Next Steps:\n")
	fmt.Printf("   • Review the changes above\n")
	fmt.Printf("   • Use --apply to apply all planned changes\n")
	fmt.Printf("   • Use --only <pattern> to apply specific files only\n")
	fmt.Printf("   • Use --dry-run with --apply to see exactly what would be done\n")

	return nil
}

// showDryRunPlan displays what would be applied in dry-run mode
func showDryRunPlan(result DiffResult) error {
	fmt.Printf("🔍 Dry Run: What Would Be Applied\n")
	fmt.Printf("═══════════════════════════════════\n\n")

	plannedActions := 0
	for _, action := range result.Plan {
		if action.Status == "planned" {
			plannedActions++
			var icon string
			switch action.Action {
			case "new":
				icon = "📄"
				fmt.Printf("   %s CREATE %s\n", icon, action.Path)
			case "mod":
				icon = "📝"
				fmt.Printf("   %s UPDATE %s\n", icon, action.Path)
			case "del":
				icon = "🗑️"
				fmt.Printf("   %s DELETE %s\n", icon, action.Path)
			}
		}
	}

	if plannedActions == 0 {
		fmt.Printf("✅ No actions to apply.\n")
	} else {
		fmt.Printf("\n📊 Summary: %d actions would be applied\n", plannedActions)
		fmt.Printf("\n💡 Remove --dry-run to actually apply these changes.\n")
	}

	return nil
}

// applyDiffChanges applies the planned changes to the file system
func applyDiffChanges(result DiffResult, sandboxPath, systemPath string) error {
	fmt.Printf("🚀 Applying Changes: Sandbox → Source\n")
	fmt.Printf("═══════════════════════════════════════\n\n")

	applied := 0
	skipped := 0

	for i, action := range result.Plan {
		if action.Status != "planned" {
			skipped++
			continue
		}

		sandboxFile := filepath.Join(sandboxPath, action.Path)
		systemFile := filepath.Join(systemPath, action.Path)

		var err error
		switch action.Action {
		case "new", "mod":
			fmt.Printf("📄 Copying %s\n", action.Path)
			err = fsutil.CopyFileWithDir(sandboxFile, systemFile)
			if err == nil {
				result.Plan[i].Status = "applied"
				applied++
			}
		case "del":
			fmt.Printf("🗑️  Deleting %s\n", action.Path)
			err = os.Remove(systemFile)
			if err == nil {
				result.Plan[i].Status = "applied"
				applied++
			}
		}

		if err != nil {
			fmt.Printf("   ❌ Error: %v\n", err)
			result.Plan[i].Status = "failed"
		}
	}

	fmt.Printf("\n📊 Results Summary:\n")
	fmt.Printf("   ✅ Applied:  %d\n", applied)
	if skipped > 0 {
		fmt.Printf("   ⏭️  Skipped:  %d\n", skipped)
	}

	failedCount := 0
	for _, action := range result.Plan {
		if action.Status == "failed" {
			failedCount++
		}
	}
	if failedCount > 0 {
		fmt.Printf("   ❌ Failed:   %d\n", failedCount)
	}

	if applied > 0 {
		fmt.Printf("\n✅ Successfully updated source directory with %d changes.\n", applied)
	}

	return nil
}

