/*
Copyright © 2025 Claude WM CLI Team
*/
package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"claude-wm-cli/internal/epic"

	"github.com/spf13/cobra"
)

// epicCmd represents the epic command
var epicCmd = &cobra.Command{
	Use:   "epic",
	Short: "Manage project epics",
	Long: `Manage project epics including creation, updating, listing, and selection.

Epics are large units of work that organize multiple user stories. Use epic
commands to structure your project workflow and track progress across major
features or initiatives.

Available subcommands:
  create   Create a new epic
  list     List all epics with their status
  update   Update an existing epic
  select   Set an epic as the current active epic
  show     Display detailed information about an epic

Examples:
  claude-wm-cli epic create "User Authentication" --priority high
  claude-wm-cli epic list --status in_progress
  claude-wm-cli epic select EPIC-001-USER-AUTH
  claude-wm-cli epic update EPIC-001 --status completed`,
}

// epicCreateCmd represents the epic create command
var epicCreateCmd = &cobra.Command{
	Use:   "create <title>",
	Short: "Create a new epic",
	Long: `Create a new epic with the specified title and optional parameters.

The epic will be created with a unique ID and stored in the project's epic
collection. You can specify priority, description, duration, and tags.

Examples:
  claude-wm-cli epic create "User Authentication System"
  claude-wm-cli epic create "API Integration" --priority high --description "Integrate with external APIs"
  claude-wm-cli epic create "UI Redesign" --priority medium --duration "2 weeks" --tags ui,design`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		createEpic(args[0], cmd)
	},
}

// epicListCmd represents the epic list command
var epicListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all epics with their status",
	Long: `List all epics in the project with their current status and progress.

You can filter the list by status or priority to focus on specific epics.
The list shows epic ID, title, status, priority, and completion percentage.

Examples:
  claude-wm-cli epic list                    # List all epics
  claude-wm-cli epic list --status planned  # List only planned epics
  claude-wm-cli epic list --priority high   # List only high priority epics
  claude-wm-cli epic list --all             # Show all epics including completed`,
	Run: func(cmd *cobra.Command, args []string) {
		listEpics(cmd)
	},
}

// epicUpdateCmd represents the epic update command
var epicUpdateCmd = &cobra.Command{
	Use:   "update <epic-id>",
	Short: "Update an existing epic",
	Long: `Update the properties of an existing epic such as title, description,
priority, status, duration, or tags.

You can update multiple properties in a single command. The epic's updated
timestamp will be automatically set.

Examples:
  claude-wm-cli epic update EPIC-001 --status in_progress
  claude-wm-cli epic update EPIC-001 --title "New Title" --priority critical
  claude-wm-cli epic update EPIC-001 --description "Updated description" --duration "3 weeks"`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		updateEpic(args[0], cmd)
	},
}

// epicSelectCmd represents the epic select command
var epicSelectCmd = &cobra.Command{
	Use:   "select <epic-id>",
	Short: "Set an epic as the current active epic",
	Long: `Set the specified epic as the current active epic for the project.

This will make the epic the focus of your workflow and automatically start
it if it's in planned status. Only one epic can be active at a time.

Examples:
  claude-wm-cli epic select EPIC-001-USER-AUTH
  claude-wm-cli epic select EPIC-002`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		selectEpic(args[0])
	},
}

// epicShowCmd represents the epic show command
var epicShowCmd = &cobra.Command{
	Use:   "show <epic-id>",
	Short: "Display detailed information about an epic",
	Long: `Display detailed information about a specific epic including all
properties, user stories, progress metrics, and timestamps.

Examples:
  claude-wm-cli epic show EPIC-001
  claude-wm-cli epic show EPIC-001-USER-AUTH`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		showEpic(args[0])
	},
}

// epicHistoryCmd represents the epic history command
var epicHistoryCmd = &cobra.Command{
	Use:   "history <epic-id>",
	Short: "Show state transition history of an epic",
	Long: `Display the complete state transition history of an epic including
timestamps, reasons for transitions, and metadata.

Examples:
  claude-wm-cli epic history EPIC-001
  claude-wm-cli epic history EPIC-001-USER-AUTH`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		showEpicHistory(args[0])
	},
}

// epicMetricsCmd represents the epic metrics command
var epicMetricsCmd = &cobra.Command{
	Use:   "metrics <epic-id>",
	Short: "Show advanced metrics for an epic",
	Long: `Display advanced metrics for an epic including duration analytics,
velocity, estimated completion, and state transition analysis.

Examples:
  claude-wm-cli epic metrics EPIC-001
  claude-wm-cli epic metrics EPIC-001-USER-AUTH`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		showEpicMetrics(args[0])
	},
}

// epicDashboardCmd represents the epic dashboard command
var epicDashboardCmd = &cobra.Command{
	Use:   "dashboard",
	Short: "Display epic progress dashboard",
	Long: `Display a comprehensive dashboard showing progress, risk analysis, and 
velocity metrics for all epics in the project.

The dashboard provides:
- Overall project progress summary
- Individual epic progress with visual progress bars
- Risk assessment and alerts for high-risk epics
- Velocity tracking and timeline analysis
- Recommendations for improving epic delivery

Examples:
  claude-wm-cli epic dashboard`,
	Run: func(cmd *cobra.Command, args []string) {
		showEpicDashboard()
	},
}

// Flag variables
var (
	epicPriority    string
	epicDescription string
	epicDuration    string
	epicTags        []string
	epicStatus      string
	listStatus      string
	listPriority    string
	listAll         bool
)

func init() {
	rootCmd.AddCommand(epicCmd)

	// Add subcommands
	epicCmd.AddCommand(epicCreateCmd)
	epicCmd.AddCommand(epicListCmd)
	epicCmd.AddCommand(epicUpdateCmd)
	epicCmd.AddCommand(epicSelectCmd)
	epicCmd.AddCommand(epicShowCmd)
	epicCmd.AddCommand(epicHistoryCmd)
	epicCmd.AddCommand(epicMetricsCmd)
	epicCmd.AddCommand(epicDashboardCmd)

	// epic create flags
	epicCreateCmd.Flags().StringVarP(&epicPriority, "priority", "p", "medium", "Epic priority (low, medium, high, critical)")
	epicCreateCmd.Flags().StringVarP(&epicDescription, "description", "d", "", "Epic description")
	epicCreateCmd.Flags().StringVar(&epicDuration, "duration", "", "Estimated duration (e.g., '2 weeks', '1 month')")
	epicCreateCmd.Flags().StringSliceVarP(&epicTags, "tags", "t", []string{}, "Epic tags (comma-separated)")

	// epic list flags
	epicListCmd.Flags().StringVar(&listStatus, "status", "", "Filter by status (planned, in_progress, on_hold, completed, cancelled)")
	epicListCmd.Flags().StringVar(&listPriority, "priority", "", "Filter by priority (low, medium, high, critical)")
	epicListCmd.Flags().BoolVar(&listAll, "all", false, "Show all epics including completed and cancelled")

	// epic update flags
	epicUpdateCmd.Flags().StringVar(&epicPriority, "priority", "", "Update epic priority")
	epicUpdateCmd.Flags().StringVar(&epicDescription, "description", "", "Update epic description")
	epicUpdateCmd.Flags().StringVar(&epicDuration, "duration", "", "Update estimated duration")
	epicUpdateCmd.Flags().StringSliceVar(&epicTags, "tags", []string{}, "Update epic tags")
	epicUpdateCmd.Flags().StringVar(&epicStatus, "status", "", "Update epic status")
	epicUpdateCmd.Flags().StringVar(&epicTitle, "title", "", "Update epic title")
}

var epicTitle string

func createEpic(title string, cmd *cobra.Command) {
	// Get current working directory
	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to get working directory: %v\n", err)
		os.Exit(1)
	}

	// Create epic manager
	manager := epic.NewManager(wd)

	// Parse priority
	var priority epic.Priority
	if epicPriority != "" {
		priority = epic.Priority(epicPriority)
		if !priority.IsValid() {
			fmt.Fprintf(os.Stderr, "Error: Invalid priority '%s'. Valid values: low, medium, high, critical\n", epicPriority)
			os.Exit(1)
		}
	}

	// Create epic options
	options := epic.EpicCreateOptions{
		Title:        title,
		Description:  epicDescription,
		Priority:     priority,
		Duration:     epicDuration,
		Tags:         epicTags,
		Dependencies: []string{}, // TODO: Add dependencies support in future
	}

	// Create the epic
	newEpic, err := manager.CreateEpic(options)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to create epic: %v\n", err)
		os.Exit(1)
	}

	// Display success message
	fmt.Printf("✅ Epic created successfully!\n\n")
	fmt.Printf("📝 Epic Details:\n")
	fmt.Printf("   ID:          %s\n", newEpic.ID)
	fmt.Printf("   Title:       %s\n", newEpic.Title)
	fmt.Printf("   Priority:    %s\n", newEpic.Priority)
	fmt.Printf("   Status:      %s\n", newEpic.Status)
	if newEpic.Description != "" {
		fmt.Printf("   Description: %s\n", newEpic.Description)
	}
	if newEpic.Duration != "" {
		fmt.Printf("   Duration:    %s\n", newEpic.Duration)
	}
	if len(newEpic.Tags) > 0 {
		fmt.Printf("   Tags:        %s\n", strings.Join(newEpic.Tags, ", "))
	}
	fmt.Printf("   Created:     %s\n", newEpic.CreatedAt.Format("2006-01-02 15:04:05"))

	fmt.Printf("\n💡 Next steps:\n")
	fmt.Printf("   • Select this epic:  claude-wm-cli epic select %s\n", newEpic.ID)
	fmt.Printf("   • List all epics:    claude-wm-cli epic list\n")
	fmt.Printf("   • Update this epic:  claude-wm-cli epic update %s --status in_progress\n", newEpic.ID)
}

func listEpics(cmd *cobra.Command) {
	// Get current working directory
	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to get working directory: %v\n", err)
		os.Exit(1)
	}

	// Create epic manager
	manager := epic.NewManager(wd)

	// Parse filter options
	var statusFilter epic.Status
	if listStatus != "" {
		statusFilter = epic.Status(listStatus)
		if !statusFilter.IsValid() {
			fmt.Fprintf(os.Stderr, "Error: Invalid status '%s'. Valid values: planned, in_progress, on_hold, completed, cancelled\n", listStatus)
			os.Exit(1)
		}
	}

	var priorityFilter epic.Priority
	if listPriority != "" {
		priorityFilter = epic.Priority(listPriority)
		if !priorityFilter.IsValid() {
			fmt.Fprintf(os.Stderr, "Error: Invalid priority '%s'. Valid values: low, medium, high, critical\n", listPriority)
			os.Exit(1)
		}
	}

	// Create list options
	options := epic.EpicListOptions{
		Status:   statusFilter,
		Priority: priorityFilter,
		ShowAll:  listAll,
	}

	// Get epics
	epics, err := manager.ListEpics(options)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to list epics: %v\n", err)
		os.Exit(1)
	}

	// Get current epic
	currentEpic, _ := manager.GetCurrentEpic()
	var currentEpicID string
	if currentEpic != nil {
		currentEpicID = currentEpic.ID
	}

	// Display header
	fmt.Printf("📋 Project Epics\n")
	fmt.Printf("================\n\n")

	if len(epics) == 0 {
		fmt.Printf("No epics found")
		if listStatus != "" || listPriority != "" {
			fmt.Printf(" matching the specified filters")
		}
		fmt.Printf(".\n\n")
		fmt.Printf("💡 Create your first epic: claude-wm-cli epic create \"Epic Title\"\n")
		return
	}

	// Create table writer
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	// Print header
	fmt.Fprintf(w, "ID\tTITLE\tSTATUS\tPRIORITY\tPROGRESS\tCREATED\tCURRENT\n")
	fmt.Fprintf(w, "──\t─────\t──────\t────────\t────────\t───────\t───────\n")

	// Print each epic
	for _, ep := range epics {
		isCurrent := ""
		if ep.ID == currentEpicID {
			isCurrent = "→"
		}

		// Format status with emoji
		statusIcon := getEpicStatusIcon(ep.Status)
		priorityIcon := getEpicPriorityIcon(ep.Priority)

		progressStr := fmt.Sprintf("%.0f%%", ep.Progress.CompletionPercentage)
		if ep.Progress.TotalStories > 0 {
			progressStr += fmt.Sprintf(" (%d/%d)", ep.Progress.CompletedStories, ep.Progress.TotalStories)
		}

		createdStr := ep.CreatedAt.Format("Jan 02")

		fmt.Fprintf(w, "%s\t%s\t%s %s\t%s %s\t%s\t%s\t%s\n",
			ep.ID,
			truncateEpicString(ep.Title, 30),
			statusIcon, ep.Status,
			priorityIcon, ep.Priority,
			progressStr,
			createdStr,
			isCurrent)
	}

	w.Flush()

	// Show summary
	fmt.Printf("\n📊 Summary: %d epic(s)", len(epics))
	if currentEpicID != "" {
		fmt.Printf(" • Current: %s", currentEpicID)
	}
	fmt.Printf("\n\n")

	// Show next actions
	if len(epics) > 0 && currentEpicID == "" {
		fmt.Printf("💡 Next steps:\n")
		fmt.Printf("   • Select an epic:  claude-wm-cli epic select <epic-id>\n")
		fmt.Printf("   • View details:    claude-wm-cli epic show <epic-id>\n")
	}
}

func updateEpic(epicID string, cmd *cobra.Command) {
	// Get current working directory
	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to get working directory: %v\n", err)
		os.Exit(1)
	}

	// Create epic manager
	manager := epic.NewManager(wd)

	// Build update options
	options := epic.EpicUpdateOptions{}

	if epicTitle != "" {
		options.Title = &epicTitle
	}

	if epicDescription != "" {
		options.Description = &epicDescription
	}

	if epicPriority != "" {
		priority := epic.Priority(epicPriority)
		if !priority.IsValid() {
			fmt.Fprintf(os.Stderr, "Error: Invalid priority '%s'. Valid values: low, medium, high, critical\n", epicPriority)
			os.Exit(1)
		}
		options.Priority = &priority
	}

	if epicStatus != "" {
		status := epic.Status(epicStatus)
		if !status.IsValid() {
			fmt.Fprintf(os.Stderr, "Error: Invalid status '%s'. Valid values: planned, in_progress, on_hold, completed, cancelled\n", epicStatus)
			os.Exit(1)
		}
		options.Status = &status
	}

	if epicDuration != "" {
		options.Duration = &epicDuration
	}

	if len(epicTags) > 0 {
		options.Tags = &epicTags
	}

	// Check if any updates were specified
	if options.Title == nil && options.Description == nil && options.Priority == nil &&
		options.Status == nil && options.Duration == nil && options.Tags == nil {
		fmt.Fprintf(os.Stderr, "Error: No updates specified. Use flags like --title, --status, --priority, etc.\n")
		os.Exit(1)
	}

	// Update the epic
	updatedEpic, err := manager.UpdateEpic(epicID, options)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to update epic: %v\n", err)
		os.Exit(1)
	}

	// Display success message
	fmt.Printf("✅ Epic updated successfully!\n\n")
	fmt.Printf("📝 Updated Epic Details:\n")
	fmt.Printf("   ID:          %s\n", updatedEpic.ID)
	fmt.Printf("   Title:       %s\n", updatedEpic.Title)
	fmt.Printf("   Priority:    %s\n", updatedEpic.Priority)
	fmt.Printf("   Status:      %s\n", updatedEpic.Status)
	if updatedEpic.Description != "" {
		fmt.Printf("   Description: %s\n", updatedEpic.Description)
	}
	if updatedEpic.Duration != "" {
		fmt.Printf("   Duration:    %s\n", updatedEpic.Duration)
	}
	if len(updatedEpic.Tags) > 0 {
		fmt.Printf("   Tags:        %s\n", strings.Join(updatedEpic.Tags, ", "))
	}
	fmt.Printf("   Updated:     %s\n", updatedEpic.UpdatedAt.Format("2006-01-02 15:04:05"))
}

func selectEpic(epicID string) {
	// Get current working directory
	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to get working directory: %v\n", err)
		os.Exit(1)
	}

	// Create epic manager
	manager := epic.NewManager(wd)

	// Select the epic
	selectedEpic, err := manager.SelectEpic(epicID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to select epic: %v\n", err)
		os.Exit(1)
	}

	// Display success message
	fmt.Printf("✅ Epic selected successfully!\n\n")
	fmt.Printf("🎯 Active Epic:\n")
	fmt.Printf("   ID:       %s\n", selectedEpic.ID)
	fmt.Printf("   Title:    %s\n", selectedEpic.Title)
	fmt.Printf("   Status:   %s\n", selectedEpic.Status)
	fmt.Printf("   Priority: %s\n", selectedEpic.Priority)
	fmt.Printf("   Progress: %.0f%%", selectedEpic.Progress.CompletionPercentage)
	if selectedEpic.Progress.TotalStories > 0 {
		fmt.Printf(" (%d/%d stories)", selectedEpic.Progress.CompletedStories, selectedEpic.Progress.TotalStories)
	}
	fmt.Printf("\n")

	fmt.Printf("\n💡 Next steps:\n")
	fmt.Printf("   • View epic details: claude-wm-cli epic show %s\n", selectedEpic.ID)
	fmt.Printf("   • Check status:      claude-wm-cli status\n")
	if selectedEpic.Status == epic.StatusInProgress {
		fmt.Printf("   • Create stories:    claude-wm-cli story create\n")
	}
}

func showEpic(epicID string) {
	// Get current working directory
	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to get working directory: %v\n", err)
		os.Exit(1)
	}

	// Create epic manager
	manager := epic.NewManager(wd)

	// Get the epic
	ep, err := manager.GetEpic(epicID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to get epic: %v\n", err)
		os.Exit(1)
	}

	// Check if it's the current epic
	currentEpic, _ := manager.GetCurrentEpic()
	isCurrent := currentEpic != nil && currentEpic.ID == ep.ID

	// Display epic details
	fmt.Printf("📋 Epic Details\n")
	fmt.Printf("===============\n\n")

	fmt.Printf("🆔 ID:          %s", ep.ID)
	if isCurrent {
		fmt.Printf(" (CURRENT)")
	}
	fmt.Printf("\n")

	fmt.Printf("📝 Title:       %s\n", ep.Title)
	fmt.Printf("📊 Status:      %s %s\n", getEpicStatusIcon(ep.Status), ep.Status)
	fmt.Printf("⚡ Priority:    %s %s\n", getEpicPriorityIcon(ep.Priority), ep.Priority)

	if ep.Description != "" {
		fmt.Printf("📄 Description: %s\n", ep.Description)
	}

	if ep.Duration != "" {
		fmt.Printf("⏱️  Duration:    %s\n", ep.Duration)
	}

	if len(ep.Tags) > 0 {
		fmt.Printf("🏷️  Tags:        %s\n", strings.Join(ep.Tags, ", "))
	}

	if len(ep.Dependencies) > 0 {
		fmt.Printf("🔗 Dependencies: %s\n", strings.Join(ep.Dependencies, ", "))
	}

	// Progress section
	fmt.Printf("\n📈 Progress:\n")
	fmt.Printf("   Overall:         %.0f%%\n", ep.Progress.CompletionPercentage)
	fmt.Printf("   Stories:         %d/%d completed\n", ep.Progress.CompletedStories, ep.Progress.TotalStories)
	if ep.Progress.TotalStoryPoints > 0 {
		fmt.Printf("   Story Points:    %d/%d completed\n", ep.Progress.CompletedStoryPoints, ep.Progress.TotalStoryPoints)
	}

	// Timestamps
	fmt.Printf("\n📅 Timestamps:\n")
	fmt.Printf("   Created:    %s\n", ep.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("   Updated:    %s\n", ep.UpdatedAt.Format("2006-01-02 15:04:05"))
	if ep.StartDate != nil {
		fmt.Printf("   Started:    %s\n", ep.StartDate.Format("2006-01-02 15:04:05"))
	}
	if ep.EndDate != nil {
		fmt.Printf("   Completed:  %s\n", ep.EndDate.Format("2006-01-02 15:04:05"))
	}

	// User stories section
	if len(ep.UserStories) > 0 {
		fmt.Printf("\n👥 User Stories (%d):\n", len(ep.UserStories))
		for i, story := range ep.UserStories {
			fmt.Printf("   %d. %s %s (%s, %s", i+1, getEpicStatusIcon(story.Status), story.Title, story.Status, story.Priority)
			if story.StoryPoints > 0 {
				fmt.Printf(", %d pts", story.StoryPoints)
			}
			fmt.Printf(")\n")
		}
	} else {
		fmt.Printf("\n👥 User Stories: None defined yet\n")
	}

	// Next actions
	fmt.Printf("\n💡 Available Actions:\n")
	if !isCurrent && (ep.Status == epic.StatusPlanned || ep.Status == epic.StatusInProgress) {
		fmt.Printf("   • Select this epic:  claude-wm-cli epic select %s\n", ep.ID)
	}
	fmt.Printf("   • Update this epic:  claude-wm-cli epic update %s --status <status>\n", ep.ID)
	fmt.Printf("   • List all epics:    claude-wm-cli epic list\n")
	if isCurrent && ep.Status == epic.StatusInProgress {
		fmt.Printf("   • Create stories:    claude-wm-cli story create\n")
	}
}

// Helper functions

func getEpicStatusIcon(status epic.Status) string {
	switch status {
	case epic.StatusPlanned:
		return "📋"
	case epic.StatusInProgress:
		return "🚧"
	case epic.StatusOnHold:
		return "⏸️"
	case epic.StatusCompleted:
		return "✅"
	case epic.StatusCancelled:
		return "❌"
	default:
		return "❓"
	}
}

func getEpicPriorityIcon(priority epic.Priority) string {
	switch priority {
	case epic.PriorityLow:
		return "🟢"
	case epic.PriorityMedium:
		return "🟡"
	case epic.PriorityHigh:
		return "🟠"
	case epic.PriorityCritical:
		return "🔴"
	default:
		return "⚪"
	}
}

func truncateEpicString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func showEpicHistory(epicID string) {
	// Get current working directory
	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to get working directory: %v\n", err)
		os.Exit(1)
	}

	// Create epic manager
	manager := epic.NewManager(wd)

	// Check if epic exists
	ep, err := manager.GetEpic(epicID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to get epic: %v\n", err)
		os.Exit(1)
	}

	// Get state history
	history := manager.GetEpicStateHistory(epicID)

	// Display header
	fmt.Printf("📊 Epic State History: %s\n", ep.Title)
	fmt.Printf("===========================================\n\n")

	if len(history) == 0 {
		fmt.Printf("No state transitions recorded for this epic.\n")
		return
	}

	// Display each transition
	for i, transition := range history {
		fmt.Printf("%d. %s → %s\n", i+1,
			getEpicStatusIcon(transition.FromStatus),
			getEpicStatusIcon(transition.ToStatus))
		fmt.Printf("   Status: %s → %s\n", transition.FromStatus, transition.ToStatus)
		fmt.Printf("   Time:   %s\n", transition.Timestamp.Format("2006-01-02 15:04:05"))
		fmt.Printf("   Reason: %s\n", transition.Reason)
		if transition.TriggeredBy != "" {
			fmt.Printf("   By:     %s\n", transition.TriggeredBy)
		}

		// Show metadata if any
		if len(transition.Metadata) > 0 {
			fmt.Printf("   Data:   ")
			for key, value := range transition.Metadata {
				fmt.Printf("%s=%v ", key, value)
			}
			fmt.Printf("\n")
		}

		if i < len(history)-1 {
			fmt.Printf("\n")
		}
	}

	fmt.Printf("\n📈 Summary: %d state transitions\n", len(history))
	if len(history) > 0 {
		fmt.Printf("   Latest: %s (%s)\n",
			history[len(history)-1].ToStatus,
			history[len(history)-1].Timestamp.Format("Jan 02 15:04"))
	}
}

func showEpicMetrics(epicID string) {
	// Get current working directory
	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to get working directory: %v\n", err)
		os.Exit(1)
	}

	// Create epic manager
	manager := epic.NewManager(wd)

	// Check if epic exists
	ep, err := manager.GetEpic(epicID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to get epic: %v\n", err)
		os.Exit(1)
	}

	// Get advanced metrics
	metrics, err := manager.GetEpicAdvancedMetrics(epicID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to get epic metrics: %v\n", err)
		os.Exit(1)
	}

	// Display header
	fmt.Printf("📊 Epic Advanced Metrics: %s\n", ep.Title)
	fmt.Printf("=======================================\n\n")

	// Basic metrics
	fmt.Printf("📈 Basic Progress:\n")
	fmt.Printf("   Overall:         %.1f%%\n", metrics.BasicMetrics.CompletionPercentage)
	fmt.Printf("   Stories:         %d/%d completed\n",
		metrics.BasicMetrics.CompletedStories, metrics.BasicMetrics.TotalStories)
	if metrics.BasicMetrics.TotalStoryPoints > 0 {
		fmt.Printf("   Story Points:    %d/%d completed\n",
			metrics.BasicMetrics.CompletedStoryPoints, metrics.BasicMetrics.TotalStoryPoints)
	}

	// Duration metrics
	fmt.Printf("\n⏱️  Duration Analysis:\n")
	if metrics.TotalDuration > 0 {
		fmt.Printf("   Total Duration:  %s\n", formatDuration(metrics.TotalDuration))
		fmt.Printf("   Duration (Days): %d days\n", metrics.DurationDays)
	} else {
		fmt.Printf("   Duration:        Not started yet\n")
	}

	// State transition metrics
	fmt.Printf("\n🔄 State Transitions:\n")
	fmt.Printf("   Total Transitions: %d\n", metrics.StateTransitions)
	if metrics.LastTransition != nil {
		fmt.Printf("   Last Transition:   %s → %s (%s)\n",
			metrics.LastTransition.FromStatus,
			metrics.LastTransition.ToStatus,
			metrics.LastTransition.Timestamp.Format("Jan 02 15:04"))
	}
	if metrics.AvgTransitionTime > 0 {
		fmt.Printf("   Avg Between:       %s\n", formatDuration(metrics.AvgTransitionTime))
	}

	// Velocity and predictions
	fmt.Printf("\n🎯 Velocity & Predictions:\n")
	if metrics.EstimatedCompletion != nil {
		fmt.Printf("   Est. Completion:   %s\n", metrics.EstimatedCompletion.Format("2006-01-02 15:04"))

		// Calculate time remaining
		if metrics.EstimatedCompletion.After(time.Now()) {
			remaining := metrics.EstimatedCompletion.Sub(time.Now())
			fmt.Printf("   Time Remaining:    %s\n", formatDuration(remaining))
		} else {
			overdue := time.Since(*metrics.EstimatedCompletion)
			fmt.Printf("   Overdue By:        %s\n", formatDuration(overdue))
		}
	} else {
		fmt.Printf("   Est. Completion:   Unable to calculate\n")
	}

	// Summary
	fmt.Printf("\n📋 Calculated: %s\n", metrics.CalculatedAt.Format("2006-01-02 15:04:05"))
}

func formatDuration(d time.Duration) string {
	days := int(d.Hours() / 24)
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm", days, hours, minutes)
	} else if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	} else {
		return fmt.Sprintf("%dm", minutes)
	}
}

func showEpicDashboard() {
	// Get current working directory
	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to get working directory: %v\n", err)
		os.Exit(1)
	}

	// Create epic manager and dashboard
	manager := epic.NewManager(wd)
	dashboard := epic.NewDashboard(manager)

	// Display the dashboard
	if err := dashboard.DisplayEpicDashboard(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to display dashboard: %v\n", err)
		os.Exit(1)
	}
}
