/*
Copyright © 2025 Claude WM CLI Team
*/
package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"claude-wm-cli/internal/epic"
	"claude-wm-cli/internal/story"

	"github.com/spf13/cobra"
)

// storyCmd represents the story command
var storyCmd = &cobra.Command{
	Use:   "story",
	Short: "Manage user stories within epics",
	Long: `Manage user stories including creation, updating, listing, and task management.

Stories are individual units of work that belong to epics. Use story
commands to break down epics into manageable development tasks.

Available subcommands:
  create     Create a new story
  list       List all stories with their status
  update     Update an existing story
  show       Display detailed information about a story
  generate   Generate stories from epic definitions

Examples:
  claude-wm-cli story create "User Login" --epic EPIC-001 --priority high
  claude-wm-cli story list --epic EPIC-001 --status in_progress
  claude-wm-cli story update STORY-001 --status completed
  claude-wm-cli story show STORY-001`,
}

// storyCreateCmd represents the story create command
var storyCreateCmd = &cobra.Command{
	Use:   "create <title>",
	Short: "Create a new story",
	Long: `Create a new user story with the specified title and optional parameters.

The story will be created and stored in the project's story collection.
You can specify epic, priority, description, story points, and acceptance criteria.

Examples:
  claude-wm-cli story create "User Registration"
  claude-wm-cli story create "API Integration" --epic EPIC-001 --priority high
  claude-wm-cli story create "UI Component" --story-points 5 --criteria "Component renders,Component is responsive"`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		createStory(args[0], cmd)
	},
}

// storyListCmd represents the story list command
var storyListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all stories with their status",
	Long: `List all stories in the project with their current status and progress.

You can filter the list by epic or status to focus on specific stories.
The list shows story ID, title, epic, status, priority, and progress.

Examples:
  claude-wm-cli story list                      # List all stories
  claude-wm-cli story list --epic EPIC-001     # List stories from specific epic
  claude-wm-cli story list --status planned    # List only planned stories`,
	Run: func(cmd *cobra.Command, args []string) {
		listStories(cmd)
	},
}

// storyUpdateCmd represents the story update command
var storyUpdateCmd = &cobra.Command{
	Use:   "update <story-id>",
	Short: "Update an existing story",
	Long: `Update the properties of an existing story such as title, description,
priority, status, story points, or acceptance criteria.

You can update multiple properties in a single command. The story's updated
timestamp will be automatically set.

Examples:
  claude-wm-cli story update STORY-001 --status in_progress
  claude-wm-cli story update STORY-001 --title "New Title" --priority critical
  claude-wm-cli story update STORY-001 --story-points 8 --criteria "New criteria"`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		updateStory(args[0], cmd)
	},
}

// storyShowCmd represents the story show command
var storyShowCmd = &cobra.Command{
	Use:   "show <story-id>",
	Short: "Display detailed information about a story",
	Long: `Display detailed information about a specific story including all
properties, acceptance criteria, tasks, progress metrics, and timestamps.

Examples:
  claude-wm-cli story show STORY-001
  claude-wm-cli story show STORY-001-USER-LOGIN`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		showStory(args[0])
	},
}

// storyGenerateCmd represents the story generate command
var storyGenerateCmd = &cobra.Command{
	Use:   "generate [epic-id]",
	Short: "Generate stories from epic definitions",
	Long: `Generate stories from epic user stories and acceptance criteria.

If no epic ID is provided, stories will be generated from all epics.
This command reads the epic definitions and creates corresponding stories
with tasks generated from acceptance criteria.

Examples:
  claude-wm-cli story generate                # Generate from all epics
  claude-wm-cli story generate EPIC-001      # Generate from specific epic`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		generateStories(args)
	},
}

// Flag variables
var (
	storyEpicID      string
	storyPriority    string
	storyDescription string
	storyPoints      int
	storyCriteria    []string
	storyStatus      string
	storyTitle       string
	listStoryEpic    string
	listStoryStatus  string
	dependencies     []string
)

func init() {
	rootCmd.AddCommand(storyCmd)

	// Add subcommands
	storyCmd.AddCommand(storyCreateCmd)
	storyCmd.AddCommand(storyListCmd)
	storyCmd.AddCommand(storyUpdateCmd)
	storyCmd.AddCommand(storyShowCmd)
	storyCmd.AddCommand(storyGenerateCmd)

	// story create flags
	storyCreateCmd.Flags().StringVar(&storyEpicID, "epic", "", "Epic ID to associate story with")
	storyCreateCmd.Flags().StringVarP(&storyPriority, "priority", "p", "medium", "Story priority (low, medium, high, critical)")
	storyCreateCmd.Flags().StringVarP(&storyDescription, "description", "d", "", "Story description")
	storyCreateCmd.Flags().IntVar(&storyPoints, "story-points", 0, "Story points for estimation")
	storyCreateCmd.Flags().StringSliceVar(&storyCriteria, "criteria", []string{}, "Acceptance criteria (comma-separated)")
	storyCreateCmd.Flags().StringSliceVar(&dependencies, "dependencies", []string{}, "Story dependencies (comma-separated)")

	// story list flags
	storyListCmd.Flags().StringVar(&listStoryEpic, "epic", "", "Filter by epic ID")
	storyListCmd.Flags().StringVar(&listStoryStatus, "status", "", "Filter by status (planned, in_progress, on_hold, completed, cancelled)")

	// story update flags
	storyUpdateCmd.Flags().StringVar(&storyTitle, "title", "", "Update story title")
	storyUpdateCmd.Flags().StringVar(&storyDescription, "description", "", "Update story description")
	storyUpdateCmd.Flags().StringVar(&storyPriority, "priority", "", "Update story priority")
	storyUpdateCmd.Flags().StringVar(&storyStatus, "status", "", "Update story status")
	storyUpdateCmd.Flags().IntVar(&storyPoints, "story-points", 0, "Update story points")
	storyUpdateCmd.Flags().StringSliceVar(&storyCriteria, "criteria", []string{}, "Update acceptance criteria")
	storyUpdateCmd.Flags().StringSliceVar(&dependencies, "dependencies", []string{}, "Update story dependencies")
}

func createStory(title string, cmd *cobra.Command) {
	// Get current working directory
	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to get working directory: %v\n", err)
		os.Exit(1)
	}

	// Create story generator
	generator := story.NewGenerator(wd)

	// Parse priority
	var priority epic.Priority
	if storyPriority != "" {
		priority = epic.Priority(storyPriority)
		if !priority.IsValid() {
			fmt.Fprintf(os.Stderr, "Error: Invalid priority '%s'. Valid values: low, medium, high, critical\n", storyPriority)
			os.Exit(1)
		}
	}

	// Create story options
	options := story.StoryCreateOptions{
		Title:              title,
		Description:        storyDescription,
		EpicID:             storyEpicID,
		Priority:           priority,
		StoryPoints:        storyPoints,
		AcceptanceCriteria: storyCriteria,
		Dependencies:       dependencies,
	}

	// Create the story
	newStory, err := generator.CreateStory(options)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to create story: %v\n", err)
		os.Exit(1)
	}

	// Display success message
	fmt.Printf("✅ Story created successfully!\n\n")
	fmt.Printf("📝 Story Details:\n")
	fmt.Printf("   ID:          %s\n", newStory.ID)
	fmt.Printf("   Title:       %s\n", newStory.Title)
	fmt.Printf("   Epic ID:     %s\n", newStory.EpicID)
	fmt.Printf("   Priority:    %s\n", newStory.Priority)
	fmt.Printf("   Status:      %s\n", newStory.Status)
	fmt.Printf("   Points:      %d\n", newStory.StoryPoints)
	if newStory.Description != "" {
		fmt.Printf("   Description: %s\n", newStory.Description)
	}
	if len(newStory.AcceptanceCriteria) > 0 {
		fmt.Printf("   Criteria:    %s\n", strings.Join(newStory.AcceptanceCriteria, ", "))
	}
	if len(newStory.Tasks) > 0 {
		fmt.Printf("   Tasks:       %d generated from criteria\n", len(newStory.Tasks))
	}
	fmt.Printf("   Created:     %s\n", newStory.CreatedAt.Format("2006-01-02 15:04:05"))

	fmt.Printf("\n💡 Next steps:\n")
	fmt.Printf("   • View story details: claude-wm-cli story show %s\n", newStory.ID)
	fmt.Printf("   • Update story:       claude-wm-cli story update %s --status in_progress\n", newStory.ID)
	fmt.Printf("   • List all stories:   claude-wm-cli story list\n")
}

func listStories(cmd *cobra.Command) {
	// Get current working directory
	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to get working directory: %v\n", err)
		os.Exit(1)
	}

	// Create story generator
	generator := story.NewGenerator(wd)

	// Parse filter options
	var statusFilter epic.Status
	if listStoryStatus != "" {
		statusFilter = epic.Status(listStoryStatus)
		if !statusFilter.IsValid() {
			fmt.Fprintf(os.Stderr, "Error: Invalid status '%s'. Valid values: planned, in_progress, on_hold, completed, cancelled\n", listStoryStatus)
			os.Exit(1)
		}
	}

	// Get stories
	stories, err := generator.ListStories(listStoryEpic, statusFilter)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to list stories: %v\n", err)
		os.Exit(1)
	}

	// Display header
	fmt.Printf("📋 Project Stories\n")
	fmt.Printf("==================\n\n")

	if len(stories) == 0 {
		fmt.Printf("No stories found")
		if listStoryEpic != "" || listStoryStatus != "" {
			fmt.Printf(" matching the specified filters")
		}
		fmt.Printf(".\n\n")
		fmt.Printf("💡 Create your first story: claude-wm-cli story create \"Story Title\"\n")
		return
	}

	// Create table writer
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	// Print header
	fmt.Fprintf(w, "ID\tTITLE\tEPIC\tSTATUS\tPRIORITY\tPOINTS\tPROGRESS\tCREATED\n")
	fmt.Fprintf(w, "──\t─────\t────\t──────\t────────\t──────\t────────\t───────\n")

	// Print each story
	for _, st := range stories {
		// Format status with emoji
		statusIcon := getStoryStatusIcon(st.Status)
		priorityIcon := getStoryPriorityIcon(st.Priority)

		// Calculate progress
		progress := st.CalculateProgress()
		progressStr := fmt.Sprintf("%.0f%%", progress.CompletionPercentage)
		if progress.TotalTasks > 0 {
			progressStr += fmt.Sprintf(" (%d/%d)", progress.CompletedTasks, progress.TotalTasks)
		}

		createdStr := st.CreatedAt.Format("Jan 02")
		epicStr := st.EpicID
		if epicStr == "" {
			epicStr = "-"
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s %s\t%s %s\t%d\t%s\t%s\n",
			st.ID,
			truncateStoryString(st.Title, 25),
			epicStr,
			statusIcon, st.Status,
			priorityIcon, st.Priority,
			st.StoryPoints,
			progressStr,
			createdStr)
	}

	w.Flush()

	// Show summary
	fmt.Printf("\n📊 Summary: %d story(ies)\n\n", len(stories))
}

func updateStory(storyID string, cmd *cobra.Command) {
	// Get current working directory
	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to get working directory: %v\n", err)
		os.Exit(1)
	}

	// Create story generator
	generator := story.NewGenerator(wd)

	// Build update options
	options := story.StoryUpdateOptions{}

	if storyTitle != "" {
		options.Title = &storyTitle
	}

	if storyDescription != "" {
		options.Description = &storyDescription
	}

	if storyPriority != "" {
		priority := epic.Priority(storyPriority)
		if !priority.IsValid() {
			fmt.Fprintf(os.Stderr, "Error: Invalid priority '%s'. Valid values: low, medium, high, critical\n", storyPriority)
			os.Exit(1)
		}
		options.Priority = &priority
	}

	if storyStatus != "" {
		status := epic.Status(storyStatus)
		if !status.IsValid() {
			fmt.Fprintf(os.Stderr, "Error: Invalid status '%s'. Valid values: planned, in_progress, on_hold, completed, cancelled\n", storyStatus)
			os.Exit(1)
		}
		options.Status = &status
	}

	if storyPoints > 0 {
		options.StoryPoints = &storyPoints
	}

	if len(storyCriteria) > 0 {
		options.AcceptanceCriteria = &storyCriteria
	}

	if len(dependencies) > 0 {
		options.Dependencies = &dependencies
	}

	// Check if any updates were specified
	if options.Title == nil && options.Description == nil && options.Priority == nil &&
		options.Status == nil && options.StoryPoints == nil && options.AcceptanceCriteria == nil &&
		options.Dependencies == nil {
		fmt.Fprintf(os.Stderr, "Error: No updates specified. Use flags like --title, --status, --priority, etc.\n")
		os.Exit(1)
	}

	// Update the story
	updatedStory, err := generator.UpdateStory(storyID, options)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to update story: %v\n", err)
		os.Exit(1)
	}

	// Display success message
	fmt.Printf("✅ Story updated successfully!\n\n")
	fmt.Printf("📝 Updated Story Details:\n")
	fmt.Printf("   ID:          %s\n", updatedStory.ID)
	fmt.Printf("   Title:       %s\n", updatedStory.Title)
	fmt.Printf("   Epic ID:     %s\n", updatedStory.EpicID)
	fmt.Printf("   Priority:    %s\n", updatedStory.Priority)
	fmt.Printf("   Status:      %s\n", updatedStory.Status)
	fmt.Printf("   Points:      %d\n", updatedStory.StoryPoints)
	if updatedStory.Description != "" {
		fmt.Printf("   Description: %s\n", updatedStory.Description)
	}
	if len(updatedStory.AcceptanceCriteria) > 0 {
		fmt.Printf("   Criteria:    %s\n", strings.Join(updatedStory.AcceptanceCriteria, ", "))
	}
	fmt.Printf("   Updated:     %s\n", updatedStory.UpdatedAt.Format("2006-01-02 15:04:05"))
}

func showStory(storyID string) {
	// Get current working directory
	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to get working directory: %v\n", err)
		os.Exit(1)
	}

	// Create story generator
	generator := story.NewGenerator(wd)

	// Get the story
	st, err := generator.GetStory(storyID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to get story: %v\n", err)
		os.Exit(1)
	}

	// Display story details
	fmt.Printf("📋 Story Details\n")
	fmt.Printf("================\n\n")

	fmt.Printf("🆔 ID:          %s\n", st.ID)
	fmt.Printf("📝 Title:       %s\n", st.Title)
	fmt.Printf("📊 Status:      %s %s\n", getStoryStatusIcon(st.Status), st.Status)
	fmt.Printf("⚡ Priority:    %s %s\n", getStoryPriorityIcon(st.Priority), st.Priority)
	fmt.Printf("🎯 Points:      %d\n", st.StoryPoints)

	if st.EpicID != "" {
		fmt.Printf("📚 Epic:        %s\n", st.EpicID)
	}

	if st.Description != "" {
		fmt.Printf("📄 Description: %s\n", st.Description)
	}

	if len(st.AcceptanceCriteria) > 0 {
		fmt.Printf("✅ Criteria:\n")
		for i, criteria := range st.AcceptanceCriteria {
			fmt.Printf("   %d. %s\n", i+1, criteria)
		}
	}

	if len(st.Dependencies) > 0 {
		fmt.Printf("🔗 Dependencies: %s\n", strings.Join(st.Dependencies, ", "))
	}

	// Progress section
	progress := st.CalculateProgress()
	fmt.Printf("\n📈 Progress:\n")
	fmt.Printf("   Overall:     %.0f%%\n", progress.CompletionPercentage)
	fmt.Printf("   Tasks:       %d/%d completed\n", progress.CompletedTasks, progress.TotalTasks)
	if progress.InProgressTasks > 0 {
		fmt.Printf("   In Progress: %d tasks\n", progress.InProgressTasks)
	}
	if progress.PendingTasks > 0 {
		fmt.Printf("   Pending:     %d tasks\n", progress.PendingTasks)
	}

	// Tasks section
	if len(st.Tasks) > 0 {
		fmt.Printf("\n📋 Tasks (%d):\n", len(st.Tasks))
		for i, task := range st.Tasks {
			fmt.Printf("   %d. %s %s\n", i+1, getStoryStatusIcon(task.Status), task.Title)
		}
	} else {
		fmt.Printf("\n📋 Tasks: None defined yet\n")
	}

	// Timestamps
	fmt.Printf("\n📅 Timestamps:\n")
	fmt.Printf("   Created:    %s\n", st.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("   Updated:    %s\n", st.UpdatedAt.Format("2006-01-02 15:04:05"))
	if st.StartedAt != nil {
		fmt.Printf("   Started:    %s\n", st.StartedAt.Format("2006-01-02 15:04:05"))
	}
	if st.CompletedAt != nil {
		fmt.Printf("   Completed:  %s\n", st.CompletedAt.Format("2006-01-02 15:04:05"))
	}

	// Next actions
	fmt.Printf("\n💡 Available Actions:\n")
	if st.CanStart() {
		fmt.Printf("   • Start story:       claude-wm-cli story update %s --status in_progress\n", st.ID)
	}
	if st.IsActive() {
		fmt.Printf("   • Complete story:    claude-wm-cli story update %s --status completed\n", st.ID)
		fmt.Printf("   • Put on hold:       claude-wm-cli story update %s --status on_hold\n", st.ID)
	}
	fmt.Printf("   • Update story:      claude-wm-cli story update %s --title \"New Title\"\n", st.ID)
	fmt.Printf("   • List all stories:  claude-wm-cli story list\n")
}

func generateStories(args []string) {
	// Get current working directory
	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to get working directory: %v\n", err)
		os.Exit(1)
	}

	// Create story generator
	generator := story.NewGenerator(wd)

	var err2 error
	if len(args) > 0 {
		// Generate from specific epic
		epicID := args[0]
		fmt.Printf("🔄 Generating stories from epic: %s\n\n", epicID)
		err2 = generator.GenerateStoriesFromEpic(epicID)
	} else {
		// Generate from all epics
		fmt.Printf("🔄 Generating stories from all epics\n\n")
		err2 = generator.GenerateStoriesFromAllEpics()
	}

	if err2 != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to generate stories: %v\n", err2)
		os.Exit(1)
	}

	fmt.Printf("✅ Stories generated successfully!\n\n")
	fmt.Printf("💡 Next steps:\n")
	fmt.Printf("   • List generated stories: claude-wm-cli story list\n")
	fmt.Printf("   • View story details:     claude-wm-cli story show <story-id>\n")
}

// Helper functions

func getStoryStatusIcon(status epic.Status) string {
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

func getStoryPriorityIcon(priority epic.Priority) string {
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

func truncateStoryString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
