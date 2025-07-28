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

	"claude-wm-cli/internal/debug"
	"claude-wm-cli/internal/executor"
	"claude-wm-cli/internal/ticket"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// ticketCmd represents the ticket command
var ticketCmd = &cobra.Command{
	Use:   "ticket",
	Short: "Manage interruption tickets and urgent tasks",
	Long: `Manage interruption tickets and urgent tasks that require immediate attention.

Tickets help track interruptions, bugs, urgent requests, and other tasks that
need to be handled outside of the normal epic/story workflow. Use tickets to
capture work that interrupts your current flow and needs tracking.

Available subcommands:
  create   Create a new ticket
  list     List tickets with filtering options
  show     Display detailed information about a ticket
  update   Update an existing ticket
  status   Change ticket status
  current  Set or show the current active ticket
  stats    Show ticket statistics and analytics

Examples:
  claude-wm-cli ticket create "Fix critical bug" --priority urgent --type bug
  claude-wm-cli ticket list --status open --priority high
  claude-wm-cli ticket current TICKET-001-FIX-BUG
  claude-wm-cli ticket status TICKET-001 --status resolved`,
}

// ticketCreateCmd represents the ticket create command
var ticketCreateCmd = &cobra.Command{
	Use:   "create <title>",
	Short: "Create a new ticket",
	Long: `Create a new ticket for tracking interruptions, bugs, or urgent tasks.

Tickets are used to track work that falls outside the normal epic/story workflow,
such as urgent bugs, interruptions, support requests, or ad-hoc tasks.

Examples:
  claude-wm-cli ticket create "Fix login bug"
  claude-wm-cli ticket create "Emergency deployment" --priority urgent --type interruption
  claude-wm-cli ticket create "Review PR #123" --description "Code review for authentication feature" --estimated-hours 2`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Enable debug mode if flag is set
		debug.SetDebugMode(debugMode || viper.GetBool("debug"))
		
		createTicket(args[0], cmd)
	},
}

// ticketListCmd represents the ticket list command
var ticketListCmd = &cobra.Command{
	Use:   "list",
	Short: "List tickets with filtering options",
	Long: `List tickets with optional filtering by status, priority, type, or assignment.

By default, shows open and in-progress tickets ordered by priority and creation date.
Use filters to focus on specific subsets of tickets.

Examples:
  claude-wm-cli ticket list                    # List all open tickets
  claude-wm-cli ticket list --status open     # List only open tickets
  claude-wm-cli ticket list --priority urgent # List urgent tickets
  claude-wm-cli ticket list --type bug        # List bug tickets
  claude-wm-cli ticket list --all             # Include closed tickets`,
	Run: func(cmd *cobra.Command, args []string) {
		// Enable debug mode if flag is set
		debug.SetDebugMode(debugMode || viper.GetBool("debug"))
		
		listTickets(cmd)
	},
}

// ticketShowCmd represents the ticket show command
var ticketShowCmd = &cobra.Command{
	Use:   "show <ticket-id>",
	Short: "Display detailed information about a ticket",
	Long: `Display comprehensive information about a specific ticket including
all properties, timeline, estimations, and related workflow context.

Examples:
  claude-wm-cli ticket show TICKET-001
  claude-wm-cli ticket show TICKET-001-FIX-BUG`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		showTicket(args[0])
	},
}

// ticketUpdateCmd represents the ticket update command
var ticketUpdateCmd = &cobra.Command{
	Use:   "update <ticket-id>",
	Short: "Update an existing ticket",
	Long: `Update properties of an existing ticket such as title, description,
priority, type, assignment, or estimations.

You can update multiple properties in a single command. The ticket's updated
timestamp will be automatically set.

Examples:
  claude-wm-cli ticket update TICKET-001 --title "New title"
  claude-wm-cli ticket update TICKET-001 --priority high --assigned-to john
  claude-wm-cli ticket update TICKET-001 --description "Updated description" --estimated-hours 4`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		updateTicket(args[0], cmd)
	},
}

// ticketStatusCmd represents the ticket status command
var ticketStatusCmd = &cobra.Command{
	Use:   "status <ticket-id> --status <new-status>",
	Short: "Change ticket status",
	Long: `Change the status of a ticket with proper transition validation.

Valid statuses: open, in_progress, resolved, closed
Transitions are validated to ensure proper workflow.

Examples:
  claude-wm-cli ticket status TICKET-001 --status in_progress
  claude-wm-cli ticket status TICKET-001 --status resolved
  claude-wm-cli ticket status TICKET-001 --status closed`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		changeTicketStatus(args[0], cmd)
	},
}

// ticketCurrentCmd represents the ticket current command  
var ticketCurrentCmd = &cobra.Command{
	Use:   "current [ticket-id]",
	Short: "Set or show the current active ticket",
	Long: `Set a ticket as the current active ticket or show the currently active ticket.

When a ticket is set as current, it will be automatically started if it's in open status.
Use without arguments to show the current ticket, or provide a ticket ID to set it.

Examples:
  claude-wm-cli ticket current                 # Show current ticket
  claude-wm-cli ticket current TICKET-001     # Set TICKET-001 as current
  claude-wm-cli ticket current --clear        # Clear current ticket`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Enable debug mode if flag is set
		debug.SetDebugMode(debugMode || viper.GetBool("debug"))
		
		manageCurrentTicket(args, cmd)
	},
}

// ticketStatsCmd represents the ticket stats command
var ticketStatsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show ticket statistics and analytics",
	Long: `Display analytics and statistics about tickets including counts by status,
priority, and type, as well as performance metrics like average resolution time.

Examples:
  claude-wm-cli ticket stats`,
	Run: func(cmd *cobra.Command, args []string) {
		showTicketStats()
	},
}

// Flag variables
var (
	ticketPriority       string
	ticketType           string
	ticketDescription    string
	ticketAssignedTo     string
	ticketEstimatedHours float64
	ticketStoryPoints    int
	ticketTags           []string
	ticketEpicID         string
	ticketStoryID        string
	ticketStatus         string
	ticketDueDate        string

	// List options
	listTicketStatus     string
	listTicketPriority   string
	listTicketType       string
	listTicketAssignedTo string
	listTicketAll        bool
	listTicketLimit      int

	// Current ticket options
	clearCurrent bool
)

func init() {
	rootCmd.AddCommand(ticketCmd)

	// Add subcommands
	ticketCmd.AddCommand(ticketCreateCmd)
	ticketCmd.AddCommand(ticketListCmd)
	ticketCmd.AddCommand(ticketShowCmd)
	ticketCmd.AddCommand(ticketUpdateCmd)
	ticketCmd.AddCommand(ticketStatusCmd)
	ticketCmd.AddCommand(ticketCurrentCmd)
	ticketCmd.AddCommand(ticketStatsCmd)

	// ticket create flags
	ticketCreateCmd.Flags().StringVarP(&ticketPriority, "priority", "p", "medium", "Ticket priority (low, medium, high, critical, urgent)")
	ticketCreateCmd.Flags().StringVarP(&ticketType, "type", "t", "task", "Ticket type (bug, feature, interruption, task, support)")
	ticketCreateCmd.Flags().StringVarP(&ticketDescription, "description", "d", "", "Ticket description")
	ticketCreateCmd.Flags().StringVarP(&ticketAssignedTo, "assigned-to", "a", "", "Assign ticket to someone")
	ticketCreateCmd.Flags().Float64Var(&ticketEstimatedHours, "estimated-hours", 0, "Estimated hours to complete")
	ticketCreateCmd.Flags().IntVar(&ticketStoryPoints, "story-points", 0, "Story points estimation")
	ticketCreateCmd.Flags().StringSliceVar(&ticketTags, "tags", []string{}, "Ticket tags (comma-separated)")
	ticketCreateCmd.Flags().StringVar(&ticketEpicID, "epic-id", "", "Related epic ID")
	ticketCreateCmd.Flags().StringVar(&ticketStoryID, "story-id", "", "Related story ID")
	ticketCreateCmd.Flags().StringVar(&ticketDueDate, "due-date", "", "Due date (YYYY-MM-DD format)")

	// ticket list flags
	ticketListCmd.Flags().StringVar(&listTicketStatus, "status", "", "Filter by status (open, in_progress, resolved, closed)")
	ticketListCmd.Flags().StringVar(&listTicketPriority, "priority", "", "Filter by priority (low, medium, high, critical, urgent)")
	ticketListCmd.Flags().StringVar(&listTicketType, "type", "", "Filter by type (bug, feature, interruption, task, support)")
	ticketListCmd.Flags().StringVar(&listTicketAssignedTo, "assigned-to", "", "Filter by assignee")
	ticketListCmd.Flags().BoolVar(&listTicketAll, "all", false, "Show all tickets including closed")
	ticketListCmd.Flags().IntVar(&listTicketLimit, "limit", 0, "Limit number of results")

	// ticket update flags
	ticketUpdateCmd.Flags().StringVar(&ticketPriority, "priority", "", "Update ticket priority")
	ticketUpdateCmd.Flags().StringVar(&ticketType, "type", "", "Update ticket type")
	ticketUpdateCmd.Flags().StringVar(&ticketDescription, "description", "", "Update ticket description")
	ticketUpdateCmd.Flags().StringVar(&ticketAssignedTo, "assigned-to", "", "Update ticket assignee")
	ticketUpdateCmd.Flags().Float64Var(&ticketEstimatedHours, "estimated-hours", -1, "Update estimated hours")
	ticketUpdateCmd.Flags().IntVar(&ticketStoryPoints, "story-points", -1, "Update story points")
	ticketUpdateCmd.Flags().StringSliceVar(&ticketTags, "tags", []string{}, "Update ticket tags")
	ticketUpdateCmd.Flags().StringVar(&ticketEpicID, "epic-id", "", "Update related epic ID")
	ticketUpdateCmd.Flags().StringVar(&ticketStoryID, "story-id", "", "Update related story ID")
	ticketUpdateCmd.Flags().StringVar(&ticketDueDate, "due-date", "", "Update due date (YYYY-MM-DD format)")
	ticketUpdateCmd.Flags().StringVar(&ticketTitle, "title", "", "Update ticket title")

	// ticket status flags
	ticketStatusCmd.Flags().StringVar(&ticketStatus, "status", "", "New status (open, in_progress, resolved, closed)")
	ticketStatusCmd.MarkFlagRequired("status")

	// ticket current flags
	ticketCurrentCmd.Flags().BoolVar(&clearCurrent, "clear", false, "Clear current ticket")
}

var ticketTitle string

func createTicket(title string, cmd *cobra.Command) {
	// Get current working directory
	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to get working directory: %v\n", err)
		os.Exit(1)
	}

	// Create Claude executor for enhanced ticket creation
	claudeExecutor := executor.NewClaudeExecutor()
	
	// Validate Claude is available
	if err := claudeExecutor.ValidateClaudeAvailable(); err != nil {
		debug.LogStub("TICKET", "createTicket", "Create ticket with Claude analysis but Claude CLI not available")
		fmt.Printf("⚠️  Claude CLI not found: %v\n", err)
		fmt.Println("📋 Falling back to basic ticket creation...")
	} else {
		// Execute Claude command for enhanced ticket creation
		prompt := "/3-current-task:1-tickets:CreateTicket"
		description := "Create ticket with AI-powered analysis and categorization"
		
		if err := claudeExecutor.ExecutePrompt(prompt, description); err != nil {
			debug.LogStub("TICKET", "createTicket", fmt.Sprintf("Enhanced ticket creation failed: %v", err))
			fmt.Printf("⚠️  Enhanced ticket creation failed: %v\n", err)
			fmt.Println("📋 Falling back to basic ticket creation...")
		} else {
			fmt.Println("✅ Enhanced ticket creation complete")
			return
		}
	}

	// Create ticket manager for fallback
	manager := ticket.NewManager(wd)

	// Parse priority
	var priority ticket.TicketPriority
	if ticketPriority != "" {
		priority = ticket.TicketPriority(ticketPriority)
		if !priority.IsValid() {
			fmt.Fprintf(os.Stderr, "Error: Invalid priority '%s'. Valid values: low, medium, high, critical, urgent\n", ticketPriority)
			os.Exit(1)
		}
	}

	// Parse type
	var ticketTypeVal ticket.TicketType
	if ticketType != "" {
		ticketTypeVal = ticket.TicketType(ticketType)
		if !ticketTypeVal.IsValid() {
			fmt.Fprintf(os.Stderr, "Error: Invalid type '%s'. Valid values: bug, feature, interruption, task, support\n", ticketType)
			os.Exit(1)
		}
	}

	// Parse due date
	var dueDate *time.Time
	if ticketDueDate != "" {
		parsed, err := time.Parse("2006-01-02", ticketDueDate)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Invalid due date format '%s'. Use YYYY-MM-DD format\n", ticketDueDate)
			os.Exit(1)
		}
		dueDate = &parsed
	}

	// Create ticket options
	options := ticket.TicketCreateOptions{
		Title:          title,
		Description:    ticketDescription,
		Type:           ticketTypeVal,
		Priority:       priority,
		RelatedEpicID:  ticketEpicID,
		RelatedStoryID: ticketStoryID,
		AssignedTo:     ticketAssignedTo,
		EstimatedHours: ticketEstimatedHours,
		StoryPoints:    ticketStoryPoints,
		Tags:           ticketTags,
		DueDate:        dueDate,
	}

	// Create the ticket
	newTicket, err := manager.CreateTicket(options)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to create ticket: %v\n", err)
		os.Exit(1)
	}

	// Display success message
	fmt.Printf("✅ Ticket created successfully!\n\n")
	fmt.Printf("🎫 Ticket Details:\n")
	fmt.Printf("   ID:          %s\n", newTicket.ID)
	fmt.Printf("   Title:       %s\n", newTicket.Title)
	fmt.Printf("   Type:        %s\n", newTicket.Type)
	fmt.Printf("   Priority:    %s\n", newTicket.Priority)
	fmt.Printf("   Status:      %s\n", newTicket.Status)
	if newTicket.Description != "" {
		fmt.Printf("   Description: %s\n", newTicket.Description)
	}
	if newTicket.AssignedTo != "" {
		fmt.Printf("   Assigned to: %s\n", newTicket.AssignedTo)
	}
	if newTicket.Estimations.EstimatedHours > 0 {
		fmt.Printf("   Estimated:   %.1f hours\n", newTicket.Estimations.EstimatedHours)
	}
	if len(newTicket.Tags) > 0 {
		fmt.Printf("   Tags:        %s\n", strings.Join(newTicket.Tags, ", "))
	}
	if newTicket.DueDate != nil {
		fmt.Printf("   Due date:    %s\n", newTicket.DueDate.Format("2006-01-02"))
	}
	fmt.Printf("   Created:     %s\n", newTicket.CreatedAt.Format("2006-01-02 15:04:05"))

	fmt.Printf("\n💡 Next steps:\n")
	fmt.Printf("   • Start this ticket: claude-wm-cli ticket current %s\n", newTicket.ID)
	fmt.Printf("   • List all tickets:  claude-wm-cli ticket list\n")
	fmt.Printf("   • Update ticket:     claude-wm-cli ticket update %s --status in_progress\n", newTicket.ID)
}

func listTickets(cmd *cobra.Command) {
	// Get current working directory
	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to get working directory: %v\n", err)
		os.Exit(1)
	}

	// Create Claude executor for enhanced ticket listing
	claudeExecutor := executor.NewClaudeExecutor()
	
	// Validate Claude is available
	if err := claudeExecutor.ValidateClaudeAvailable(); err != nil {
		debug.LogStub("TICKET", "listTickets", "List tickets with Claude analysis but Claude CLI not available")
		fmt.Printf("⚠️  Claude CLI not found: %v\n", err)
		fmt.Println("📋 Falling back to basic ticket listing...")
	} else {
		// Execute Claude command for enhanced ticket listing
		prompt := "/3-current-task:1-tickets:ListTickets"
		description := "List tickets with AI-powered analysis and prioritization"
		
		if err := claudeExecutor.ExecutePrompt(prompt, description); err != nil {
			debug.LogStub("TICKET", "listTickets", fmt.Sprintf("Enhanced ticket listing failed: %v", err))
			fmt.Printf("⚠️  Enhanced ticket listing failed: %v\n", err)
			fmt.Println("📋 Falling back to basic ticket listing...")
		} else {
			fmt.Println("✅ Enhanced ticket listing complete")
			return
		}
	}

	// Create ticket manager for fallback
	manager := ticket.NewManager(wd)

	// Parse filter options
	var statusFilter ticket.TicketStatus
	if listTicketStatus != "" {
		statusFilter = ticket.TicketStatus(listTicketStatus)
		if !statusFilter.IsValid() {
			fmt.Fprintf(os.Stderr, "Error: Invalid status '%s'. Valid values: open, in_progress, resolved, closed\n", listTicketStatus)
			os.Exit(1)
		}
	}

	var priorityFilter ticket.TicketPriority
	if listTicketPriority != "" {
		priorityFilter = ticket.TicketPriority(listTicketPriority)
		if !priorityFilter.IsValid() {
			fmt.Fprintf(os.Stderr, "Error: Invalid priority '%s'. Valid values: low, medium, high, critical, urgent\n", listTicketPriority)
			os.Exit(1)
		}
	}

	var typeFilter ticket.TicketType
	if listTicketType != "" {
		typeFilter = ticket.TicketType(listTicketType)
		if !typeFilter.IsValid() {
			fmt.Fprintf(os.Stderr, "Error: Invalid type '%s'. Valid values: bug, feature, interruption, task, support\n", listTicketType)
			os.Exit(1)
		}
	}

	// Create list options
	options := ticket.TicketListOptions{
		Status:     statusFilter,
		Priority:   priorityFilter,
		Type:       typeFilter,
		AssignedTo: listTicketAssignedTo,
		ShowClosed: listTicketAll,
		Limit:      listTicketLimit,
	}

	// Get tickets
	tickets, err := manager.ListTickets(options)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to list tickets: %v\n", err)
		os.Exit(1)
	}

	// Get current ticket
	currentTicket, _ := manager.GetCurrentTicket()
	var currentTicketID string
	if currentTicket != nil {
		currentTicketID = currentTicket.ID
	}

	// Display header
	fmt.Printf("🎫 Tickets\n")
	fmt.Printf("==========\n\n")

	if len(tickets) == 0 {
		fmt.Printf("No tickets found")
		if listTicketStatus != "" || listTicketPriority != "" || listTicketType != "" {
			fmt.Printf(" matching the specified filters")
		}
		fmt.Printf(".\n\n")
		fmt.Printf("💡 Create your first ticket: claude-wm-cli ticket create \"Ticket Title\"\n")
		return
	}

	// Create table writer
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	// Print header
	fmt.Fprintf(w, "ID\tTITLE\tTYPE\tSTATUS\tPRIORITY\tASSIGNED\tCREATED\tCURRENT\n")
	fmt.Fprintf(w, "──\t─────\t────\t──────\t────────\t────────\t───────\t───────\n")

	// Print each ticket
	for _, t := range tickets {
		isCurrent := ""
		if t.ID == currentTicketID {
			isCurrent = "→"
		}

		// Format icons
		typeIcon := getTicketTypeIcon(t.Type)
		statusIcon := getTicketStatusIcon(t.Status)
		priorityIcon := getTicketPriorityIcon(t.Priority)

		assignedTo := t.AssignedTo
		if assignedTo == "" {
			assignedTo = "-"
		}

		createdStr := t.CreatedAt.Format("Jan 02")

		fmt.Fprintf(w, "%s\t%s\t%s %s\t%s %s\t%s %s\t%s\t%s\t%s\n",
			t.ID,
			truncateTicketString(t.Title, 30),
			typeIcon, t.Type,
			statusIcon, t.Status,
			priorityIcon, t.Priority,
			assignedTo,
			createdStr,
			isCurrent)
	}

	w.Flush()

	// Show summary
	fmt.Printf("\n📊 Summary: %d ticket(s)", len(tickets))
	if currentTicketID != "" {
		fmt.Printf(" • Current: %s", currentTicketID)
	}
	fmt.Printf("\n\n")

	// Show next actions
	if len(tickets) > 0 && currentTicketID == "" {
		fmt.Printf("💡 Next steps:\n")
		fmt.Printf("   • Start working:     claude-wm-cli ticket current <ticket-id>\n")
		fmt.Printf("   • View details:      claude-wm-cli ticket show <ticket-id>\n")
	}
}

func showTicket(ticketID string) {
	// Get current working directory
	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to get working directory: %v\n", err)
		os.Exit(1)
	}

	// Create ticket manager
	manager := ticket.NewManager(wd)

	// Get the ticket
	t, err := manager.GetTicket(ticketID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to get ticket: %v\n", err)
		os.Exit(1)
	}

	// Check if it's the current ticket
	currentTicket, _ := manager.GetCurrentTicket()
	isCurrent := currentTicket != nil && currentTicket.ID == t.ID

	// Display ticket details
	fmt.Printf("🎫 Ticket Details\n")
	fmt.Printf("=================\n\n")

	fmt.Printf("🆔 ID:          %s", t.ID)
	if isCurrent {
		fmt.Printf(" (CURRENT)")
	}
	fmt.Printf("\n")

	fmt.Printf("📝 Title:       %s\n", t.Title)
	fmt.Printf("🏷️  Type:        %s %s\n", getTicketTypeIcon(t.Type), t.Type)
	fmt.Printf("📊 Status:      %s %s\n", getTicketStatusIcon(t.Status), t.Status)
	fmt.Printf("⚡ Priority:    %s %s\n", getTicketPriorityIcon(t.Priority), t.Priority)

	if t.Description != "" {
		fmt.Printf("📄 Description: %s\n", t.Description)
	}

	if t.AssignedTo != "" {
		fmt.Printf("👤 Assigned to: %s\n", t.AssignedTo)
	}

	// Estimations
	if t.Estimations.EstimatedHours > 0 || t.Estimations.StoryPoints > 0 {
		fmt.Printf("\n📈 Estimations:\n")
		if t.Estimations.EstimatedHours > 0 {
			fmt.Printf("   Estimated hours: %.1f\n", t.Estimations.EstimatedHours)
		}
		if t.Estimations.ActualHours > 0 {
			fmt.Printf("   Actual hours:    %.1f\n", t.Estimations.ActualHours)
		}
		if t.Estimations.StoryPoints > 0 {
			fmt.Printf("   Story points:    %d\n", t.Estimations.StoryPoints)
		}
	}

	// Related items
	if t.RelatedEpicID != "" || t.RelatedStoryID != "" {
		fmt.Printf("\n🔗 Related:\n")
		if t.RelatedEpicID != "" {
			fmt.Printf("   Epic:  %s\n", t.RelatedEpicID)
		}
		if t.RelatedStoryID != "" {
			fmt.Printf("   Story: %s\n", t.RelatedStoryID)
		}
	}

	if len(t.Tags) > 0 {
		fmt.Printf("\n🏷️  Tags:        %s\n", strings.Join(t.Tags, ", "))
	}

	if t.DueDate != nil {
		fmt.Printf("\n⏰ Due date:    %s", t.DueDate.Format("2006-01-02"))
		daysUntilDue := int(time.Until(*t.DueDate).Hours() / 24)
		if daysUntilDue < 0 {
			fmt.Printf(" (⚠️ %d days overdue)", -daysUntilDue)
		} else if daysUntilDue <= 3 {
			fmt.Printf(" (⚠️ due soon)")
		}
		fmt.Printf("\n")
	}

	// External reference
	if t.ExternalRef != nil {
		fmt.Printf("\n🔗 External:    %s %s", t.ExternalRef.System, t.ExternalRef.ID)
		if t.ExternalRef.URL != "" {
			fmt.Printf(" (%s)", t.ExternalRef.URL)
		}
		fmt.Printf("\n")
	}

	// Timestamps
	fmt.Printf("\n📅 Timeline:\n")
	fmt.Printf("   Created:    %s\n", t.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("   Updated:    %s\n", t.UpdatedAt.Format("2006-01-02 15:04:05"))
	if t.StartedAt != nil {
		fmt.Printf("   Started:    %s\n", t.StartedAt.Format("2006-01-02 15:04:05"))
	}
	if t.ResolvedAt != nil {
		fmt.Printf("   Resolved:   %s\n", t.ResolvedAt.Format("2006-01-02 15:04:05"))
	}
	if t.ClosedAt != nil {
		fmt.Printf("   Closed:     %s\n", t.ClosedAt.Format("2006-01-02 15:04:05"))
	}

	// Next actions
	fmt.Printf("\n💡 Available Actions:\n")
	if !isCurrent && (t.Status == ticket.TicketStatusOpen || t.Status == ticket.TicketStatusInProgress) {
		fmt.Printf("   • Start this ticket: claude-wm-cli ticket current %s\n", t.ID)
	}
	fmt.Printf("   • Update ticket:     claude-wm-cli ticket update %s --priority <priority>\n", t.ID)
	fmt.Printf("   • Change status:     claude-wm-cli ticket status %s --status <status>\n", t.ID)
	fmt.Printf("   • List all tickets:  claude-wm-cli ticket list\n")
}

func updateTicket(ticketID string, cmd *cobra.Command) {
	// Get current working directory
	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to get working directory: %v\n", err)
		os.Exit(1)
	}

	// Create ticket manager
	manager := ticket.NewManager(wd)

	// Build update options
	options := ticket.TicketUpdateOptions{}

	if ticketTitle != "" {
		options.Title = &ticketTitle
	}

	if ticketDescription != "" {
		options.Description = &ticketDescription
	}

	if ticketPriority != "" {
		priority := ticket.TicketPriority(ticketPriority)
		if !priority.IsValid() {
			fmt.Fprintf(os.Stderr, "Error: Invalid priority '%s'. Valid values: low, medium, high, critical, urgent\n", ticketPriority)
			os.Exit(1)
		}
		options.Priority = &priority
	}

	if ticketType != "" {
		ticketTypeVal := ticket.TicketType(ticketType)
		if !ticketTypeVal.IsValid() {
			fmt.Fprintf(os.Stderr, "Error: Invalid type '%s'. Valid values: bug, feature, interruption, task, support\n", ticketType)
			os.Exit(1)
		}
		options.Type = &ticketTypeVal
	}

	if ticketAssignedTo != "" {
		options.AssignedTo = &ticketAssignedTo
	}

	if ticketEstimatedHours >= 0 {
		options.EstimatedHours = &ticketEstimatedHours
	}

	if ticketStoryPoints >= 0 {
		options.StoryPoints = &ticketStoryPoints
	}

	if len(ticketTags) > 0 {
		options.Tags = &ticketTags
	}

	if ticketEpicID != "" {
		options.RelatedEpicID = &ticketEpicID
	}

	if ticketStoryID != "" {
		options.RelatedStoryID = &ticketStoryID
	}

	if ticketDueDate != "" {
		parsed, err := time.Parse("2006-01-02", ticketDueDate)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Invalid due date format '%s'. Use YYYY-MM-DD format\n", ticketDueDate)
			os.Exit(1)
		}
		options.DueDate = &parsed
	}

	// Check if any updates were specified
	if options.Title == nil && options.Description == nil && options.Priority == nil &&
		options.Type == nil && options.AssignedTo == nil && options.EstimatedHours == nil &&
		options.StoryPoints == nil && options.Tags == nil && options.RelatedEpicID == nil &&
		options.RelatedStoryID == nil && options.DueDate == nil {
		fmt.Fprintf(os.Stderr, "Error: No updates specified. Use flags like --title, --priority, --type, etc.\n")
		os.Exit(1)
	}

	// Update the ticket
	updatedTicket, err := manager.UpdateTicket(ticketID, options)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to update ticket: %v\n", err)
		os.Exit(1)
	}

	// Display success message
	fmt.Printf("✅ Ticket updated successfully!\n\n")
	fmt.Printf("🎫 Updated Ticket Details:\n")
	fmt.Printf("   ID:       %s\n", updatedTicket.ID)
	fmt.Printf("   Title:    %s\n", updatedTicket.Title)
	fmt.Printf("   Type:     %s\n", updatedTicket.Type)
	fmt.Printf("   Status:   %s\n", updatedTicket.Status)
	fmt.Printf("   Priority: %s\n", updatedTicket.Priority)
	fmt.Printf("   Updated:  %s\n", updatedTicket.UpdatedAt.Format("2006-01-02 15:04:05"))
}

func changeTicketStatus(ticketID string, cmd *cobra.Command) {
	// Get current working directory
	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to get working directory: %v\n", err)
		os.Exit(1)
	}

	// Create ticket manager
	manager := ticket.NewManager(wd)

	// Validate status
	newStatus := ticket.TicketStatus(ticketStatus)
	if !newStatus.IsValid() {
		fmt.Fprintf(os.Stderr, "Error: Invalid status '%s'. Valid values: open, in_progress, resolved, closed\n", ticketStatus)
		os.Exit(1)
	}

	// Update the ticket status
	options := ticket.TicketUpdateOptions{
		Status: &newStatus,
	}

	updatedTicket, err := manager.UpdateTicket(ticketID, options)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to update ticket status: %v\n", err)
		os.Exit(1)
	}

	// Display success message
	fmt.Printf("✅ Ticket status updated successfully!\n\n")
	fmt.Printf("🎫 %s\n", updatedTicket.ID)
	fmt.Printf("   Status: %s %s\n", getTicketStatusIcon(updatedTicket.Status), updatedTicket.Status)
	fmt.Printf("   Updated: %s\n", updatedTicket.UpdatedAt.Format("2006-01-02 15:04:05"))

	// Show status-specific information
	switch updatedTicket.Status {
	case ticket.TicketStatusInProgress:
		if updatedTicket.StartedAt != nil {
			fmt.Printf("   Started: %s\n", updatedTicket.StartedAt.Format("2006-01-02 15:04:05"))
		}
	case ticket.TicketStatusResolved:
		if updatedTicket.ResolvedAt != nil {
			fmt.Printf("   Resolved: %s\n", updatedTicket.ResolvedAt.Format("2006-01-02 15:04:05"))
			duration := updatedTicket.ResolvedAt.Sub(updatedTicket.CreatedAt)
			fmt.Printf("   Duration: %s\n", formatTicketDuration(duration))
		}
	case ticket.TicketStatusClosed:
		if updatedTicket.ClosedAt != nil {
			fmt.Printf("   Closed: %s\n", updatedTicket.ClosedAt.Format("2006-01-02 15:04:05"))
		}
	}
}

func manageCurrentTicket(args []string, cmd *cobra.Command) {
	// Get current working directory
	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to get working directory: %v\n", err)
		os.Exit(1)
	}

	// Create Claude executor for enhanced current ticket management
	claudeExecutor := executor.NewClaudeExecutor()
	
	// Validate Claude is available
	if err := claudeExecutor.ValidateClaudeAvailable(); err != nil {
		debug.LogStub("TICKET", "manageCurrentTicket", "Manage current ticket with Claude analysis but Claude CLI not available")
		fmt.Printf("⚠️  Claude CLI not found: %v\n", err)
		fmt.Println("📋 Falling back to basic current ticket management...")
	} else {
		// Execute Claude command for enhanced current ticket management
		prompt := "/3-current-task:1-tickets:CurrentTicket"
		description := "Manage current ticket with AI-powered context switching and focus"
		
		if err := claudeExecutor.ExecutePrompt(prompt, description); err != nil {
			debug.LogStub("TICKET", "manageCurrentTicket", fmt.Sprintf("Enhanced current ticket management failed: %v", err))
			fmt.Printf("⚠️  Enhanced current ticket management failed: %v\n", err)
			fmt.Println("📋 Falling back to basic current ticket management...")
		} else {
			fmt.Println("✅ Enhanced current ticket management complete")
			return
		}
	}

	// Create ticket manager for fallback
	manager := ticket.NewManager(wd)

	// Handle clear flag
	if clearCurrent {
		_, err := manager.SetCurrentTicket("")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to clear current ticket: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✅ Current ticket cleared.\n")
		return
	}

	// If no arguments, show current ticket
	if len(args) == 0 {
		currentTicket, err := manager.GetCurrentTicket()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to get current ticket: %v\n", err)
			os.Exit(1)
		}

		if currentTicket == nil {
			fmt.Printf("📋 No current ticket set.\n\n")
			fmt.Printf("💡 Set a current ticket: claude-wm-cli ticket current <ticket-id>\n")
			return
		}

		fmt.Printf("🎯 Current Ticket:\n")
		fmt.Printf("   ID:       %s\n", currentTicket.ID)
		fmt.Printf("   Title:    %s\n", currentTicket.Title)
		fmt.Printf("   Status:   %s %s\n", getTicketStatusIcon(currentTicket.Status), currentTicket.Status)
		fmt.Printf("   Priority: %s %s\n", getTicketPriorityIcon(currentTicket.Priority), currentTicket.Priority)
		return
	}

	// Set current ticket
	ticketID := args[0]
	selectedTicket, err := manager.SetCurrentTicket(ticketID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to set current ticket: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Current ticket set!\n\n")
	fmt.Printf("🎯 Active Ticket:\n")
	fmt.Printf("   ID:       %s\n", selectedTicket.ID)
	fmt.Printf("   Title:    %s\n", selectedTicket.Title)
	fmt.Printf("   Status:   %s %s\n", getTicketStatusIcon(selectedTicket.Status), selectedTicket.Status)
	fmt.Printf("   Priority: %s %s\n", getTicketPriorityIcon(selectedTicket.Priority), selectedTicket.Priority)

	if selectedTicket.Status == ticket.TicketStatusInProgress {
		fmt.Printf("\n💡 Ticket is now in progress!\n")
	}
}

func showTicketStats() {
	// Get current working directory
	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to get working directory: %v\n", err)
		os.Exit(1)
	}

	// Create ticket manager
	manager := ticket.NewManager(wd)

	// Get stats
	stats, err := manager.GetTicketStats()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to get ticket stats: %v\n", err)
		os.Exit(1)
	}

	// Display header
	fmt.Printf("📊 Ticket Statistics\n")
	fmt.Printf("====================\n\n")

	if stats.TotalTickets == 0 {
		fmt.Printf("No tickets found. Create your first ticket to get started!\n\n")
		fmt.Printf("💡 Create a ticket: claude-wm-cli ticket create \"Ticket Title\"\n")
		return
	}

	// Overall stats
	fmt.Printf("📈 Overall:\n")
	fmt.Printf("   Total tickets: %d\n", stats.TotalTickets)

	// By status
	fmt.Printf("\n📊 By Status:\n")
	for status, count := range stats.ByStatus {
		if count > 0 {
			fmt.Printf("   %s %-12s: %d\n", getTicketStatusIcon(status), status, count)
		}
	}

	// By priority
	fmt.Printf("\n⚡ By Priority:\n")
	priorityOrder := []ticket.TicketPriority{
		ticket.TicketPriorityUrgent,
		ticket.TicketPriorityCritical,
		ticket.TicketPriorityHigh,
		ticket.TicketPriorityMedium,
		ticket.TicketPriorityLow,
	}
	for _, priority := range priorityOrder {
		if count, exists := stats.ByPriority[priority]; exists && count > 0 {
			fmt.Printf("   %s %-12s: %d\n", getTicketPriorityIcon(priority), priority, count)
		}
	}

	// By type
	fmt.Printf("\n🏷️  By Type:\n")
	for ticketType, count := range stats.ByType {
		if count > 0 {
			fmt.Printf("   %s %-12s: %d\n", getTicketTypeIcon(ticketType), ticketType, count)
		}
	}

	// Performance metrics
	if stats.AverageResolutionTime > 0 {
		fmt.Printf("\n⏱️  Performance:\n")
		fmt.Printf("   Average resolution: %s\n", formatTicketDuration(stats.AverageResolutionTime))
	}

	if stats.OldestOpenTicket != nil {
		fmt.Printf("   Oldest open ticket: %s ago\n", formatTicketDuration(time.Since(*stats.OldestOpenTicket)))
	}
}

// Helper functions

func getTicketStatusIcon(status ticket.TicketStatus) string {
	switch status {
	case ticket.TicketStatusOpen:
		return "🔵"
	case ticket.TicketStatusInProgress:
		return "🟡"
	case ticket.TicketStatusResolved:
		return "🟢"
	case ticket.TicketStatusClosed:
		return "⚫"
	default:
		return "❓"
	}
}

func getTicketPriorityIcon(priority ticket.TicketPriority) string {
	switch priority {
	case ticket.TicketPriorityLow:
		return "🟢"
	case ticket.TicketPriorityMedium:
		return "🟡"
	case ticket.TicketPriorityHigh:
		return "🟠"
	case ticket.TicketPriorityCritical:
		return "🔴"
	case ticket.TicketPriorityUrgent:
		return "🚨"
	default:
		return "⚪"
	}
}

func getTicketTypeIcon(ticketType ticket.TicketType) string {
	switch ticketType {
	case ticket.TicketTypeBug:
		return "🐛"
	case ticket.TicketTypeFeature:
		return "✨"
	case ticket.TicketTypeInterruption:
		return "⚡"
	case ticket.TicketTypeTask:
		return "📋"
	case ticket.TicketTypeSupport:
		return "🆘"
	default:
		return "❓"
	}
}

func truncateTicketString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func formatTicketDuration(d time.Duration) string {
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
