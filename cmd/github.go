/*
Copyright © 2025 Claude WM CLI Team
*/
package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	"claude-wm-cli/internal/github"
	"claude-wm-cli/internal/ticket"

	"github.com/spf13/cobra"
)

// githubCmd represents the github command
var githubCmd = &cobra.Command{
	Use:   "github",
	Short: "GitHub issue integration for ticket management",
	Long: `Integrate with GitHub issues to automatically create and sync tickets.

The GitHub integration allows you to:
- Sync GitHub issues to tickets automatically
- Create tickets from specific GitHub issues
- Configure issue-to-ticket mapping (labels, priorities, types)
- Maintain bidirectional synchronization

Available subcommands:
  config   Configure GitHub integration settings
  sync     Sync GitHub issues to tickets
  issue    Import a specific GitHub issue as a ticket
  status   Show GitHub integration status and history
  test     Test GitHub API connection

Examples:
  claude-wm-cli github config --owner myorg --repo myproject --token $GITHUB_TOKEN
  claude-wm-cli github sync --create-new --update-existing
  claude-wm-cli github issue 123
  claude-wm-cli github status`,
}

// githubConfigCmd represents the github config command
var githubConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Configure GitHub integration settings",
	Long: `Configure GitHub repository, authentication, and sync settings.

This command sets up the GitHub integration by configuring:
- Repository owner and name
- Authentication (token or token file)
- Sync options (which issues to sync, how to map them)
- Label mappings for priority and type detection

Examples:
  claude-wm-cli github config --owner myorg --repo myproject --token $GITHUB_TOKEN
  claude-wm-cli github config --token-file ~/.github/token
  claude-wm-cli github config --show
  claude-wm-cli github config --disable`,
	Run: func(cmd *cobra.Command, args []string) {
		configureGitHub(cmd)
	},
}

// githubSyncCmd represents the github sync command
var githubSyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync GitHub issues to tickets",
	Long: `Sync GitHub issues to the local ticket system.

This command fetches GitHub issues from the configured repository and
creates or updates corresponding tickets based on the sync configuration.

You can control what gets synced using various filters and options.

Examples:
  claude-wm-cli github sync                           # Sync with default settings
  claude-wm-cli github sync --state open             # Only sync open issues
  claude-wm-cli github sync --labels bug,critical    # Only sync issues with specific labels
  claude-wm-cli github sync --create-new --update-existing
  claude-wm-cli github sync --max-issues 50          # Limit to 50 issues`,
	Run: func(cmd *cobra.Command, args []string) {
		syncGitHubIssues(cmd)
	},
}

// githubIssueCmd represents the github issue command
var githubIssueCmd = &cobra.Command{
	Use:   "issue <issue-number>",
	Short: "Import a specific GitHub issue as a ticket",
	Long: `Import a specific GitHub issue by its number and create a corresponding ticket.

This is useful for importing individual issues without running a full sync,
or for testing the integration with a specific issue.

Examples:
  claude-wm-cli github issue 123
  claude-wm-cli github issue 456`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		importGitHubIssue(args[0])
	},
}

// githubStatusCmd represents the github status command
var githubStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show GitHub integration status and history",
	Long: `Display the current GitHub integration status, configuration summary,
and synchronization history.

This includes:
- Configuration status (enabled/disabled, repository info)
- Recent sync results and statistics
- Rate limit information
- Authentication status

Examples:
  claude-wm-cli github status`,
	Run: func(cmd *cobra.Command, args []string) {
		showGitHubStatus()
	},
}

// githubTestCmd represents the github test command
var githubTestCmd = &cobra.Command{
	Use:   "test",
	Short: "Test GitHub API connection",
	Long: `Test the GitHub API connection using the current configuration.

This command verifies:
- Authentication credentials
- Repository access permissions
- API rate limits
- Network connectivity

Examples:
  claude-wm-cli github test`,
	Run: func(cmd *cobra.Command, args []string) {
		testGitHubConnection()
	},
}

// Flag variables
var (
	// Config flags
	githubOwner     string
	githubRepo      string
	githubToken     string
	githubTokenFile string
	githubBaseURL   string
	githubUploadURL string
	githubShow      bool
	githubDisable   bool
	githubEnable    bool

	// Sync flags
	syncState          string
	syncLabels         []string
	syncAssignee       string
	syncCreateNew      bool
	syncUpdateExisting bool
	syncCloseResolved  bool
	syncMaxIssues      int

	// Priority and type mappings
	priorityMappings []string
	typeMappings     []string
)

func init() {
	rootCmd.AddCommand(githubCmd)

	// Add subcommands
	githubCmd.AddCommand(githubConfigCmd)
	githubCmd.AddCommand(githubSyncCmd)
	githubCmd.AddCommand(githubIssueCmd)
	githubCmd.AddCommand(githubStatusCmd)
	githubCmd.AddCommand(githubTestCmd)

	// github config flags
	githubConfigCmd.Flags().StringVar(&githubOwner, "owner", "", "GitHub repository owner/organization")
	githubConfigCmd.Flags().StringVar(&githubRepo, "repo", "", "GitHub repository name")
	githubConfigCmd.Flags().StringVar(&githubToken, "token", "", "GitHub personal access token")
	githubConfigCmd.Flags().StringVar(&githubTokenFile, "token-file", "", "Path to file containing GitHub token")
	githubConfigCmd.Flags().StringVar(&githubBaseURL, "base-url", "", "GitHub Enterprise base URL")
	githubConfigCmd.Flags().StringVar(&githubUploadURL, "upload-url", "", "GitHub Enterprise upload URL")
	githubConfigCmd.Flags().BoolVar(&githubShow, "show", false, "Show current configuration")
	githubConfigCmd.Flags().BoolVar(&githubDisable, "disable", false, "Disable GitHub integration")
	githubConfigCmd.Flags().BoolVar(&githubEnable, "enable", false, "Enable GitHub integration")
	githubConfigCmd.Flags().StringSliceVar(&priorityMappings, "priority-map", []string{}, "Label to priority mappings (format: label=priority)")
	githubConfigCmd.Flags().StringSliceVar(&typeMappings, "type-map", []string{}, "Label to type mappings (format: label=type)")

	// github sync flags
	githubSyncCmd.Flags().StringVar(&syncState, "state", "", "Issue state filter (open, closed, all)")
	githubSyncCmd.Flags().StringSliceVar(&syncLabels, "labels", []string{}, "Filter by issue labels (comma-separated)")
	githubSyncCmd.Flags().StringVar(&syncAssignee, "assignee", "", "Filter by assignee")
	githubSyncCmd.Flags().BoolVar(&syncCreateNew, "create-new", false, "Create tickets for new issues")
	githubSyncCmd.Flags().BoolVar(&syncUpdateExisting, "update-existing", false, "Update existing tickets")
	githubSyncCmd.Flags().BoolVar(&syncCloseResolved, "close-resolved", false, "Close tickets for closed issues")
	githubSyncCmd.Flags().IntVar(&syncMaxIssues, "max-issues", 0, "Maximum number of issues to process")
}

func configureGitHub(cmd *cobra.Command) {
	// Get current working directory
	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to get working directory: %v\n", err)
		os.Exit(1)
	}

	// Create GitHub integration
	integration := github.NewIntegration(wd)

	// Load existing configuration
	if err := integration.LoadConfig(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Handle show flag
	if githubShow {
		showGitHubConfig(integration)
		return
	}

	// Handle enable/disable flags
	if githubDisable {
		config := github.DefaultConfig()
		config.Enabled = false
		if err := integration.UpdateConfig(config); err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to disable GitHub integration: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✅ GitHub integration disabled.\n")
		return
	}

	if githubEnable {
		config := github.DefaultConfig()
		config.Enabled = true
		if err := integration.UpdateConfig(config); err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to enable GitHub integration: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✅ GitHub integration enabled.\n")
		return
	}

	// Build configuration from flags
	config := github.DefaultConfig()

	if githubOwner != "" {
		config.GitHub.Owner = githubOwner
	}
	if githubRepo != "" {
		config.GitHub.Repo = githubRepo
	}
	if githubToken != "" {
		config.Auth.Token = githubToken
	}
	if githubTokenFile != "" {
		config.Auth.TokenFile = githubTokenFile
	}
	if githubBaseURL != "" {
		config.GitHub.BaseURL = githubBaseURL
	}
	if githubUploadURL != "" {
		config.GitHub.UploadURL = githubUploadURL
	}

	// Parse priority mappings
	if len(priorityMappings) > 0 {
		for _, mapping := range priorityMappings {
			parts := strings.SplitN(mapping, "=", 2)
			if len(parts) == 2 {
				label := strings.TrimSpace(parts[0])
				priority := ticket.TicketPriority(strings.TrimSpace(parts[1]))
				if priority.IsValid() {
					config.Sync.LabelToPriority[label] = priority
				} else {
					fmt.Fprintf(os.Stderr, "Warning: Invalid priority '%s' for label '%s'\n", parts[1], label)
				}
			}
		}
	}

	// Parse type mappings
	if len(typeMappings) > 0 {
		for _, mapping := range typeMappings {
			parts := strings.SplitN(mapping, "=", 2)
			if len(parts) == 2 {
				label := strings.TrimSpace(parts[0])
				ticketType := ticket.TicketType(strings.TrimSpace(parts[1]))
				if ticketType.IsValid() {
					config.Sync.LabelToType[label] = ticketType
				} else {
					fmt.Fprintf(os.Stderr, "Warning: Invalid type '%s' for label '%s'\n", parts[1], label)
				}
			}
		}
	}

	// Update configuration
	if err := integration.UpdateConfig(config); err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to update configuration: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ GitHub integration configured successfully!\n\n")
	fmt.Printf("📝 Configuration:\n")
	fmt.Printf("   Repository: %s/%s\n", config.GitHub.Owner, config.GitHub.Repo)
	fmt.Printf("   Enabled:    %t\n", config.Enabled)
	if config.GitHub.BaseURL != "" {
		fmt.Printf("   Base URL:   %s\n", config.GitHub.BaseURL)
	}

	fmt.Printf("\n💡 Next steps:\n")
	fmt.Printf("   • Test connection:  claude-wm-cli github test\n")
	fmt.Printf("   • Sync issues:      claude-wm-cli github sync --create-new\n")
	fmt.Printf("   • View status:      claude-wm-cli github status\n")
}

func syncGitHubIssues(cmd *cobra.Command) {
	// Get current working directory
	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to get working directory: %v\n", err)
		os.Exit(1)
	}

	// Create GitHub integration
	integration := github.NewIntegration(wd)

	// Load configuration
	if err := integration.LoadConfig(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize integration
	config := github.DefaultConfig()
	if err := integration.Initialize(config); err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to initialize GitHub integration: %v\n", err)
		os.Exit(1)
	}

	// Build sync options from flags
	syncOptions := &github.IssueSyncOptions{
		State:           syncState,
		Labels:          syncLabels,
		Assignee:        syncAssignee,
		CreateNew:       syncCreateNew,
		UpdateExisting:  syncUpdateExisting,
		CloseResolved:   syncCloseResolved,
		MaxIssues:       syncMaxIssues,
		DefaultPriority: ticket.TicketPriorityMedium,
		DefaultType:     ticket.TicketTypeBug,
		LabelToPriority: config.Sync.LabelToPriority,
		LabelToType:     config.Sync.LabelToType,
	}

	// If no flags provided, use defaults
	if !syncCreateNew && !syncUpdateExisting {
		syncOptions.CreateNew = true
		syncOptions.UpdateExisting = true
	}

	if syncOptions.State == "" {
		syncOptions.State = "open"
	}

	fmt.Printf("🔄 Syncing GitHub issues...\n")
	fmt.Printf("   Repository: %s/%s\n", config.GitHub.Owner, config.GitHub.Repo)
	fmt.Printf("   State:      %s\n", syncOptions.State)
	if len(syncOptions.Labels) > 0 {
		fmt.Printf("   Labels:     %s\n", strings.Join(syncOptions.Labels, ", "))
	}
	if syncOptions.Assignee != "" {
		fmt.Printf("   Assignee:   %s\n", syncOptions.Assignee)
	}
	fmt.Printf("   Create new: %t\n", syncOptions.CreateNew)
	fmt.Printf("   Update:     %t\n", syncOptions.UpdateExisting)
	fmt.Printf("\n")

	// Perform sync
	result, err := integration.SyncIssues(syncOptions)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Sync failed: %v\n", err)
		os.Exit(1)
	}

	// Display results
	fmt.Printf("✅ Sync completed!\n\n")
	fmt.Printf("📊 Results:\n")
	fmt.Printf("   Total issues:    %d\n", result.TotalIssues)
	fmt.Printf("   Created tickets: %d\n", result.CreatedTickets)
	fmt.Printf("   Updated tickets: %d\n", result.UpdatedTickets)
	fmt.Printf("   Skipped issues:  %d\n", result.SkippedIssues)
	fmt.Printf("   Errors:          %d\n", result.ErrorCount)

	if result.RateLimitInfo != nil {
		fmt.Printf("\n⏱️  Rate Limit:\n")
		fmt.Printf("   Remaining: %d/%d\n", result.RateLimitInfo.Remaining, result.RateLimitInfo.Limit)
		fmt.Printf("   Reset:     %s\n", result.RateLimitInfo.ResetTime.Format("15:04:05"))
	}

	// Show errors if any
	if len(result.Errors) > 0 {
		fmt.Printf("\n⚠️  Errors:\n")
		for _, errMsg := range result.Errors {
			fmt.Printf("   • %s\n", errMsg)
		}
	}

	// Show detailed results for recent issues
	if len(result.ProcessedIssues) > 0 {
		fmt.Printf("\n📋 Processed Issues:\n")
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintf(w, "ISSUE\tACTION\tTICKET ID\tREASON\n")
		fmt.Fprintf(w, "─────\t──────\t─────────\t──────\n")

		for _, processed := range result.ProcessedIssues {
			reason := processed.Reason
			if processed.Error != "" {
				reason = processed.Error
			}
			fmt.Fprintf(w, "#%d\t%s\t%s\t%s\n",
				processed.IssueNumber,
				processed.Action,
				processed.TicketID,
				truncateGitHubString(reason, 40))
		}
		w.Flush()
	}

	fmt.Printf("\n💡 Next steps:\n")
	fmt.Printf("   • List tickets:    claude-wm-cli ticket list\n")
	fmt.Printf("   • View status:     claude-wm-cli github status\n")
}

func importGitHubIssue(issueNumberStr string) {
	// Parse issue number
	issueNumber, err := strconv.Atoi(issueNumberStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Invalid issue number '%s'\n", issueNumberStr)
		os.Exit(1)
	}

	// Get current working directory
	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to get working directory: %v\n", err)
		os.Exit(1)
	}

	// Create GitHub integration
	integration := github.NewIntegration(wd)

	// Load configuration
	if err := integration.LoadConfig(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize integration
	config := github.DefaultConfig()
	if err := integration.Initialize(config); err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to initialize GitHub integration: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("📥 Importing GitHub issue #%d...\n", issueNumber)

	// Import the issue
	processed, err := integration.GetIssueByNumber(issueNumber)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to import issue: %v\n", err)
		os.Exit(1)
	}

	// Display result
	fmt.Printf("✅ Issue imported successfully!\n\n")
	fmt.Printf("📝 Result:\n")
	fmt.Printf("   Issue:     #%d\n", processed.IssueNumber)
	fmt.Printf("   Action:    %s\n", processed.Action)
	fmt.Printf("   Ticket ID: %s\n", processed.TicketID)
	if processed.Reason != "" {
		fmt.Printf("   Reason:    %s\n", processed.Reason)
	}
	if processed.Error != "" {
		fmt.Printf("   Error:     %s\n", processed.Error)
	}
	fmt.Printf("   URL:       %s\n", processed.IssueURL)

	if processed.TicketID != "" {
		fmt.Printf("\n💡 Next steps:\n")
		fmt.Printf("   • View ticket:  claude-wm-cli ticket show %s\n", processed.TicketID)
		fmt.Printf("   • Start work:   claude-wm-cli ticket current %s\n", processed.TicketID)
	}
}

func showGitHubStatus() {
	// Get current working directory
	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to get working directory: %v\n", err)
		os.Exit(1)
	}

	// Create GitHub integration
	integration := github.NewIntegration(wd)

	// Load configuration
	if err := integration.LoadConfig(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	config := github.DefaultConfig()
	history := integration.GetSyncHistory()

	// Display header
	fmt.Printf("📊 GitHub Integration Status\n")
	fmt.Printf("=============================\n\n")

	// Configuration status
	fmt.Printf("⚙️  Configuration:\n")
	fmt.Printf("   Enabled:    %t\n", config.Enabled)
	if config.GitHub.Owner != "" && config.GitHub.Repo != "" {
		fmt.Printf("   Repository: %s/%s\n", config.GitHub.Owner, config.GitHub.Repo)
	} else {
		fmt.Printf("   Repository: Not configured\n")
	}

	if config.GitHub.BaseURL != "" {
		fmt.Printf("   Base URL:   %s\n", config.GitHub.BaseURL)
	}

	// Authentication status
	fmt.Printf("\n🔐 Authentication:\n")
	if config.Auth.Token != "" {
		fmt.Printf("   Method:     Personal Access Token\n")
		fmt.Printf("   Status:     Configured\n")
	} else if config.Auth.TokenFile != "" {
		fmt.Printf("   Method:     Token File (%s)\n", config.Auth.TokenFile)
		fmt.Printf("   Status:     Configured\n")
	} else {
		fmt.Printf("   Method:     None\n")
		fmt.Printf("   Status:     ⚠️ Not configured\n")
	}

	// Sync history
	fmt.Printf("\n📈 Sync History:\n")
	if history.TotalSyncs == 0 {
		fmt.Printf("   No syncs performed yet\n")
	} else {
		fmt.Printf("   Total syncs:      %d\n", history.TotalSyncs)
		fmt.Printf("   Successful:       %d\n", history.SuccessfulSyncs)
		fmt.Printf("   Failed:           %d\n", history.FailedSyncs)
		fmt.Printf("   Last sync:        %s\n", history.LastSync.Format("2006-01-02 15:04:05"))
		fmt.Printf("   Last successful:  %s\n", history.LastSuccessfulSync.Format("2006-01-02 15:04:05"))
	}

	// Recent results
	if len(history.RecentResults) > 0 {
		fmt.Printf("\n📋 Recent Sync Results:\n")
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintf(w, "DATE\tISSUES\tCREATED\tUPDATED\tERRORS\n")
		fmt.Fprintf(w, "────\t──────\t───────\t───────\t──────\n")

		// Show last 5 results
		start := len(history.RecentResults) - 5
		if start < 0 {
			start = 0
		}

		for i := start; i < len(history.RecentResults); i++ {
			result := history.RecentResults[i]
			fmt.Fprintf(w, "%s\t%d\t%d\t%d\t%d\n",
				result.SyncTime.Format("Jan 02 15:04"),
				result.TotalIssues,
				result.CreatedTickets,
				result.UpdatedTickets,
				result.ErrorCount)
		}
		w.Flush()
	}

	// Show errors if any
	if len(history.LastErrors) > 0 {
		fmt.Printf("\n⚠️  Recent Errors:\n")
		for _, errMsg := range history.LastErrors {
			fmt.Printf("   • %s\n", errMsg)
		}
	}

	// Next steps
	fmt.Printf("\n💡 Available Actions:\n")
	if !config.Enabled {
		fmt.Printf("   • Enable:       claude-wm-cli github config --enable\n")
	}
	if config.GitHub.Owner == "" || config.GitHub.Repo == "" {
		fmt.Printf("   • Configure:    claude-wm-cli github config --owner <owner> --repo <repo>\n")
	}
	if config.Auth.Token == "" && config.Auth.TokenFile == "" {
		fmt.Printf("   • Set token:    claude-wm-cli github config --token <token>\n")
	}
	fmt.Printf("   • Test:         claude-wm-cli github test\n")
	fmt.Printf("   • Sync:         claude-wm-cli github sync --create-new\n")
}

func testGitHubConnection() {
	// Get current working directory
	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to get working directory: %v\n", err)
		os.Exit(1)
	}

	// Create GitHub integration
	integration := github.NewIntegration(wd)

	// Load configuration
	if err := integration.LoadConfig(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("🔧 Testing GitHub connection...\n")

	// Initialize integration (this will test the connection)
	config := github.DefaultConfig()
	if err := integration.Initialize(config); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Connection test failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ GitHub connection test successful!\n\n")
	fmt.Printf("📝 Connection Details:\n")
	fmt.Printf("   Repository: %s/%s\n", config.GitHub.Owner, config.GitHub.Repo)
	if config.GitHub.BaseURL != "" {
		fmt.Printf("   Base URL:   %s\n", config.GitHub.BaseURL)
	}

	fmt.Printf("\n💡 Next steps:\n")
	fmt.Printf("   • Sync issues:  claude-wm-cli github sync --create-new\n")
	fmt.Printf("   • View status:  claude-wm-cli github status\n")
}

func showGitHubConfig(integration *github.Integration) {
	config := github.DefaultConfig()

	fmt.Printf("📋 GitHub Integration Configuration\n")
	fmt.Printf("===================================\n\n")

	fmt.Printf("🔧 General:\n")
	fmt.Printf("   Enabled:    %t\n", config.Enabled)
	fmt.Printf("   Repository: %s/%s\n", config.GitHub.Owner, config.GitHub.Repo)
	if config.GitHub.BaseURL != "" {
		fmt.Printf("   Base URL:   %s\n", config.GitHub.BaseURL)
	}

	fmt.Printf("\n🔐 Authentication:\n")
	if config.Auth.Token != "" {
		fmt.Printf("   Token:      [CONFIGURED]\n")
	} else if config.Auth.TokenFile != "" {
		fmt.Printf("   Token File: %s\n", config.Auth.TokenFile)
	} else {
		fmt.Printf("   Token:      [NOT CONFIGURED]\n")
	}

	fmt.Printf("\n🔄 Sync Options:\n")
	fmt.Printf("   State:         %s\n", config.Sync.State)
	fmt.Printf("   Create new:    %t\n", config.Sync.CreateNew)
	fmt.Printf("   Update:        %t\n", config.Sync.UpdateExisting)
	fmt.Printf("   Close resolved: %t\n", config.Sync.CloseResolved)
	fmt.Printf("   Max issues:    %d\n", config.Sync.MaxIssues)

	if len(config.Sync.LabelToPriority) > 0 {
		fmt.Printf("\n🏷️  Priority Mappings:\n")
		for label, priority := range config.Sync.LabelToPriority {
			fmt.Printf("   %s → %s\n", label, priority)
		}
	}

	if len(config.Sync.LabelToType) > 0 {
		fmt.Printf("\n🏷️  Type Mappings:\n")
		for label, ticketType := range config.Sync.LabelToType {
			fmt.Printf("   %s → %s\n", label, ticketType)
		}
	}
}

// Helper functions

func truncateGitHubString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
