/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"claude-wm-cli/internal/git"
	"claude-wm-cli/internal/state"

	"github.com/spf13/cobra"
)

var (
	gitConfigFile   string
	gitFormat       string
	gitLimit        int
	gitStrategy     string
	gitMaxDepth     int
	gitVerify       bool
	gitCreateBackup bool
)

// gitCmd represents the git command
var gitCmd = &cobra.Command{
	Use:   "git",
	Short: "Git integration for state versioning",
	Long: `Git integration provides versioning and recovery capabilities for state files.

FEATURES:
  • Automatic state versioning with semantic commits
  • Point-in-time recovery using Git history
  • Corruption detection and automatic recovery
  • Branch management for different contexts
  • Integration with atomic file operations

COMMANDS:
  • status      - Show Git repository status
  • log         - Show commit history  
  • recover     - Recover from state corruption
  • backup      - Create recovery points
  • config      - Manage Git configuration

Examples:
  claude-wm-cli git status                          # Show repo status
  claude-wm-cli git log --limit 10                 # Show last 10 commits
  claude-wm-cli git recover --strategy automatic   # Auto-recover from corruption
  claude-wm-cli git backup "milestone checkpoint"  # Create recovery point`,
}

// gitStatusCmd shows Git repository status
var gitStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show Git repository status",
	Long: `Display the current status of the Git repository including:
- Current branch and remote tracking
- Staged and unstaged changes  
- Untracked files
- Commit information`,
	Run: func(cmd *cobra.Command, args []string) {
		showGitStatus()
	},
}

// gitLogCmd shows commit history
var gitLogCmd = &cobra.Command{
	Use:   "log",
	Short: "Show commit history",
	Long: `Display commit history with state versioning information.
Supports multiple output formats and filtering options.`,
	Run: func(cmd *cobra.Command, args []string) {
		showGitLog()
	},
}

// gitRecoverCmd recovers from state corruption
var gitRecoverCmd = &cobra.Command{
	Use:   "recover [files...]",
	Short: "Recover from state corruption",
	Long: `Recover corrupted state files using Git history.
Supports multiple recovery strategies and automatic corruption detection.

STRATEGIES:
  • automatic    - Fully automated recovery (default)
  • conservative - Safe recovery with user confirmation
  • aggressive   - Fast recovery that may lose recent changes
  • interactive  - User-guided recovery process

Examples:
  claude-wm-cli git recover                           # Auto-detect and recover
  claude-wm-cli git recover --strategy conservative  # Safe recovery
  claude-wm-cli git recover stories.json              # Recover specific file`,
	Run: func(cmd *cobra.Command, args []string) {
		recoverFromCorruption(args)
	},
}

// gitBackupCmd creates recovery points
var gitBackupCmd = &cobra.Command{
	Use:   "backup [description]",
	Short: "Create recovery point",
	Long: `Create a tagged recovery point for the current state.
This allows easy restoration to known good states.`,
	Run: func(cmd *cobra.Command, args []string) {
		createBackup(args)
	},
}

// gitConfigCmd manages Git configuration
var gitConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage Git configuration",
	Long: `Display or update Git integration configuration.
Shows current settings and allows modifications.`,
	Run: func(cmd *cobra.Command, args []string) {
		manageGitConfig()
	},
}

func showGitStatus() {
	workingDir, _ := os.Getwd()
	config := git.DefaultGitConfig()

	repo := git.NewRepository(workingDir, config)

	if !repo.IsRepository() {
		fmt.Println("❌ Not a Git repository")
		fmt.Println("💡 Run 'claude-wm-cli git init' to initialize Git integration")
		return
	}

	status, err := repo.GetStatus()
	if err != nil {
		fmt.Printf("❌ Failed to get Git status: %v\n", err)
		return
	}

	fmt.Println("📊 Git Repository Status")
	fmt.Println("========================")
	fmt.Printf("Branch: %s\n", status.Branch)

	if status.Remote != "" {
		fmt.Printf("Remote: %s", status.Remote)
		if status.Ahead > 0 || status.Behind > 0 {
			fmt.Printf(" (ahead %d, behind %d)", status.Ahead, status.Behind)
		}
		fmt.Println()
	}

	if status.Clean {
		fmt.Println("✅ Working tree clean")
	} else {
		fmt.Printf("📝 Changes: %d staged, %d modified, %d untracked",
			status.Staged, status.Modified, status.Untracked)
		if status.Conflicted > 0 {
			fmt.Printf(", %d conflicted", status.Conflicted)
		}
		fmt.Println()

		if gitFormat == "json" {
			jsonData, _ := json.MarshalIndent(status, "", "  ")
			fmt.Println(string(jsonData))
		} else {
			// Show file details
			for _, file := range status.Files {
				icon := getFileStatusIcon(file.Status)
				fmt.Printf("  %s %s\n", icon, file.Path)
			}
		}
	}

	// Show last commit info
	if status.LastCommit != "" {
		fmt.Printf("\nLast commit: %s\n", status.LastCommit)
		if status.LastCommitMsg != "" {
			fmt.Printf("Message: %s\n", status.LastCommitMsg)
		}
	}
}

func showGitLog() {
	workingDir, _ := os.Getwd()
	config := git.DefaultGitConfig()

	repo := git.NewRepository(workingDir, config)

	if !repo.IsRepository() {
		fmt.Println("❌ Not a Git repository")
		return
	}

	commits, err := repo.GetLog(gitLimit)
	if err != nil {
		fmt.Printf("❌ Failed to get commit history: %v\n", err)
		return
	}

	if len(commits) == 0 {
		fmt.Println("📝 No commits found")
		return
	}

	fmt.Printf("📜 Commit History (last %d commits)\n", len(commits))
	fmt.Println("====================================")

	if gitFormat == "json" {
		jsonData, _ := json.MarshalIndent(commits, "", "  ")
		fmt.Println(string(jsonData))
		return
	}

	for i, commit := range commits {
		fmt.Printf("%d. %s %s\n", i+1, commit.ShortHash, commit.Message)
		fmt.Printf("   Author: %s <%s>\n", commit.Author, commit.Email)
		fmt.Printf("   Date: %s\n", commit.Date.Format("2006-01-02 15:04:05"))

		if len(commit.Files) > 0 {
			fmt.Printf("   Files: %s", strings.Join(commit.Files, ", "))
			if len(commit.Files) > 3 {
				fmt.Printf(" (+%d more)", len(commit.Files)-3)
			}
			fmt.Println()
		}

		if i < len(commits)-1 {
			fmt.Println()
		}
	}
}

func recoverFromCorruption(targetFiles []string) {
	workingDir, _ := os.Getwd()
	config := git.DefaultGitConfig()

	// Initialize Git repository and version manager
	repo := git.NewRepository(workingDir, config)
	if !repo.IsRepository() {
		fmt.Println("❌ Not a Git repository - cannot perform recovery")
		return
	}

	atomicWriter := state.NewAtomicWriter("")
	versionManager, err := git.NewStateVersionManager(workingDir, config, atomicWriter)
	if err != nil {
		fmt.Printf("❌ Failed to initialize version manager: %v\n", err)
		return
	}

	corruptionDetector := state.NewCorruptionDetector(atomicWriter)
	recoveryEngine := git.NewGitRecoveryEngine(repo, versionManager, corruptionDetector, config)

	fmt.Println("🔍 Analyzing state corruption...")

	// Auto-detect corrupted files if none specified
	if len(targetFiles) == 0 {
		targetFiles = autoDetectCorruptedFiles(corruptionDetector, workingDir)
		if len(targetFiles) == 0 {
			fmt.Println("✅ No corruption detected")
			return
		}
	}

	fmt.Printf("🚨 Found %d corrupted files: %s\n", len(targetFiles), strings.Join(targetFiles, ", "))

	// Create recovery options
	strategy := git.RecoveryStrategy(gitStrategy)
	options := &git.RecoveryOptions{
		Strategy:        strategy,
		MaxSearchDepth:  gitMaxDepth,
		VerifyIntegrity: gitVerify,
		CreateBackup:    gitCreateBackup,
		TargetFiles:     targetFiles,
	}

	fmt.Printf("🔧 Starting %s recovery...\n", strategy)

	// Perform recovery
	result, err := recoveryEngine.AutoRecover(targetFiles, options)
	if err != nil {
		fmt.Printf("❌ Recovery failed: %v\n", err)
		return
	}

	// Display results
	displayRecoveryResult(result)
}

func createBackup(args []string) {
	workingDir, _ := os.Getwd()
	config := git.DefaultGitConfig()

	description := "manual backup"
	if len(args) > 0 {
		description = strings.Join(args, " ")
	}

	atomicWriter := state.NewAtomicWriter("")
	versionManager, err := git.NewStateVersionManager(workingDir, config, atomicWriter)
	if err != nil {
		fmt.Printf("❌ Failed to initialize version manager: %v\n", err)
		return
	}

	// Find state files to backup
	stateFiles := findStateFiles(workingDir)
	if len(stateFiles) == 0 {
		fmt.Println("📝 No state files found to backup")
		return
	}

	fmt.Printf("💾 Creating backup of %d state files...\n", len(stateFiles))

	recoveryPoint, err := versionManager.CreateRecoveryPoint("manual", description, stateFiles...)
	if err != nil {
		fmt.Printf("❌ Failed to create backup: %v\n", err)
		return
	}

	fmt.Printf("✅ Backup created successfully\n")
	fmt.Printf("📋 Recovery point: %s\n", recoveryPoint.Commit.ShortHash)
	fmt.Printf("📝 Description: %s\n", recoveryPoint.Description)
	fmt.Printf("📁 Files: %d\n", len(recoveryPoint.StateFiles))
}

func manageGitConfig() {
	config := git.DefaultGitConfig()

	fmt.Println("⚙️  Git Integration Configuration")
	fmt.Println("================================")

	if gitFormat == "json" {
		jsonData, _ := json.MarshalIndent(config, "", "  ")
		fmt.Println(string(jsonData))
		return
	}

	fmt.Printf("Enabled: %v\n", config.Enabled)
	fmt.Printf("Repository: %s\n", config.RepositoryPath)
	fmt.Printf("Auto-commit: %v\n", config.AutoCommit)
	fmt.Printf("Branch: %s\n", config.Branch)
	fmt.Printf("Max commits: %d\n", config.MaxCommits)
	fmt.Printf("Auto-push: %v\n", config.AutoPush)
	fmt.Printf("Conflict strategy: %s\n", config.ConflictStrategy)

	if config.Username != "" {
		fmt.Printf("User: %s <%s>\n", config.Username, config.Email)
	}

	if config.RemoteURL != "" {
		fmt.Printf("Remote: %s\n", config.RemoteURL)
	}
}

// Helper functions

func getFileStatusIcon(status string) string {
	switch status {
	case "M":
		return "📝" // Modified
	case "A":
		return "➕" // Added
	case "D":
		return "🗑️" // Deleted
	case "R":
		return "📋" // Renamed
	case "C":
		return "📄" // Copied
	case "U":
		return "⚠️" // Unmerged/Conflicted
	case "?":
		return "❓" // Untracked
	default:
		return "📄"
	}
}

func autoDetectCorruptedFiles(detector *state.CorruptionDetector, workingDir string) []string {
	var corruptedFiles []string

	// Scan for JSON files in common state directories
	stateDirs := []string{
		filepath.Join(workingDir, "docs", "current-epic"),
		filepath.Join(workingDir, "docs", "project"),
		filepath.Join(workingDir, "state"),
		workingDir,
	}

	for _, dir := range stateDirs {
		if reports, err := detector.ScanDirectory(dir); err == nil {
			for _, report := range reports {
				if report.IsCorrupted {
					corruptedFiles = append(corruptedFiles, report.FilePath)
				}
			}
		}
	}

	return corruptedFiles
}

func findStateFiles(workingDir string) []string {
	var stateFiles []string

	// Common state file patterns
	patterns := []string{
		"*.json",
		"state/*.json",
		"docs/**/*.json",
	}

	for _, pattern := range patterns {
		if matches, err := filepath.Glob(filepath.Join(workingDir, pattern)); err == nil {
			for _, match := range matches {
				// Check if file exists and is readable
				if info, err := os.Stat(match); err == nil && !info.IsDir() {
					stateFiles = append(stateFiles, match)
				}
			}
		}
	}

	return stateFiles
}

func displayRecoveryResult(result *git.RecoveryResult) {
	fmt.Printf("\n🎯 Recovery Result: %s\n", getResultIcon(result.Success))
	fmt.Println("========================")
	fmt.Printf("Strategy: %s\n", result.Strategy)
	fmt.Printf("Duration: %v\n", result.Duration)
	fmt.Printf("Steps executed: %d\n", len(result.StepsExecuted))

	if len(result.StepsFailed) > 0 {
		fmt.Printf("Steps failed: %d\n", len(result.StepsFailed))
	}

	if len(result.FilesRecovered) > 0 {
		fmt.Printf("Files recovered: %s\n", strings.Join(result.FilesRecovered, ", "))
	}

	if result.BackupCreated != "" {
		fmt.Printf("Backup created: %s\n", result.BackupCreated)
	}

	if result.IntegrityCheck {
		fmt.Println("✅ Integrity check passed")
	} else {
		fmt.Println("⚠️ Integrity check failed")
	}

	if len(result.Warnings) > 0 {
		fmt.Println("\n⚠️ Warnings:")
		for _, warning := range result.Warnings {
			fmt.Printf("  - %s\n", warning)
		}
	}

	if len(result.NextSteps) > 0 {
		fmt.Println("\n📋 Next Steps:")
		for _, step := range result.NextSteps {
			fmt.Printf("  - %s\n", step)
		}
	}
}

func getResultIcon(success bool) string {
	if success {
		return "✅ SUCCESS"
	}
	return "❌ FAILED"
}

func init() {
	rootCmd.AddCommand(gitCmd)

	// Add subcommands
	gitCmd.AddCommand(gitStatusCmd)
	gitCmd.AddCommand(gitLogCmd)
	gitCmd.AddCommand(gitRecoverCmd)
	gitCmd.AddCommand(gitBackupCmd)
	gitCmd.AddCommand(gitConfigCmd)

	// Global flags
	gitCmd.PersistentFlags().StringVarP(&gitConfigFile, "config", "c", "", "Git configuration file")
	gitCmd.PersistentFlags().StringVarP(&gitFormat, "format", "f", "text", "Output format (text, json)")

	// Log command flags
	gitLogCmd.Flags().IntVarP(&gitLimit, "limit", "l", 10, "Maximum number of commits to show")

	// Recover command flags
	gitRecoverCmd.Flags().StringVarP(&gitStrategy, "strategy", "s", "automatic",
		"Recovery strategy (automatic, conservative, aggressive, interactive)")
	gitRecoverCmd.Flags().IntVar(&gitMaxDepth, "max-depth", 50, "Maximum search depth in commit history")
	gitRecoverCmd.Flags().BoolVar(&gitVerify, "verify", true, "Verify integrity after recovery")
	gitRecoverCmd.Flags().BoolVar(&gitCreateBackup, "backup", true, "Create backup before recovery")
}
