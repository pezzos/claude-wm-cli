/*
Copyright © 2025 Claude WM CLI Team
*/
package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"claude-wm-cli/internal/workflow"

	"github.com/spf13/cobra"
)

// interruptCmd represents the interrupt command
var interruptCmd = &cobra.Command{
	Use:   "interrupt",
	Short: "Manage workflow interruptions and context switching",
	Long: `Manage workflow interruptions while preserving current work context.

The interrupt system allows you to:
- Start interruption workflows for urgent tasks
- Preserve your current workflow state
- Resume previous workflows after interruption completion
- View and manage the interruption stack

Available subcommands:
  start     Start a new interruption workflow
  resume    Resume the previous workflow from the stack
  status    Show current interruption stack and context
  clear     Clear the interruption stack (emergency use)

Examples:
  claude-wm-cli interrupt start --name "Urgent Bug Fix" --type emergency
  claude-wm-cli interrupt resume
  claude-wm-cli interrupt status
  claude-wm-cli interrupt clear --confirm`,
}

// interruptStartCmd represents the interrupt start command
var interruptStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a new interruption workflow",
	Long: `Start a new interruption workflow, saving the current context.

This command saves your current workflow state and creates a new context for handling
the interruption. Your current epic, story, and ticket states are preserved.

The interruption workflow allows nested interruptions (e.g., emergency within an
existing interruption), maintaining a proper context stack.

Examples:
  claude-wm-cli interrupt start --name "Critical Bug Fix"
  claude-wm-cli interrupt start --name "Security Patch" --type emergency
  claude-wm-cli interrupt start --name "Hotfix Deploy" --type hotfix --notes "Production issue"`,
	Run: func(cmd *cobra.Command, args []string) {
		startInterruption(cmd)
	},
}

// interruptResumeCmd represents the interrupt resume command
var interruptResumeCmd = &cobra.Command{
	Use:   "resume [context-id]",
	Short: "Resume the previous workflow from the interruption stack",
	Long: `Resume the previous workflow by popping from the interruption stack.

If no context ID is provided, resumes the most recent context from the stack.
If a specific context ID is provided, resumes that specific context and removes
it from the stack.

This restores your previous epic, story, and ticket states, allowing you to
continue where you left off before the interruption.

Examples:
  claude-wm-cli interrupt resume                    # Resume most recent
  claude-wm-cli interrupt resume ctx-12345         # Resume specific context
  claude-wm-cli interrupt resume --force           # Force resume with conflicts`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var contextID string
		if len(args) > 0 {
			contextID = args[0]
		}
		resumeInterruption(cmd, contextID)
	},
}

// interruptStatusCmd represents the interrupt status command
var interruptStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current interruption stack and context information",
	Long: `Display detailed information about the current interruption stack.

This includes:
- Current workflow context
- Interruption stack depth and contents
- Context history and relationships
- Epic, story, and ticket states for each context
- Timing information and metadata

Examples:
  claude-wm-cli interrupt status
  claude-wm-cli interrupt status --verbose
  claude-wm-cli interrupt status --format json`,
	Run: func(cmd *cobra.Command, args []string) {
		showInterruptionStatus(cmd)
	},
}

// interruptClearCmd represents the interrupt clear command
var interruptClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear the entire interruption stack (emergency use)",
	Long: `Clear the entire interruption stack and reset to a clean state.

WARNING: This is an emergency command that will permanently remove all saved
contexts from the interruption stack. Use with extreme caution as this cannot
be undone.

This should only be used when the interruption stack is corrupted or when you
need to forcibly reset the workflow state.

Examples:
  claude-wm-cli interrupt clear --confirm
  claude-wm-cli interrupt clear --confirm --backup`,
	Run: func(cmd *cobra.Command, args []string) {
		clearInterruptionStack(cmd)
	},
}

// Flag variables
var (
	// Start flags
	interruptName        string
	interruptDescription string
	interruptType        string
	interruptNotes       string
	interruptTags        []string
	includeFileState     bool
	includeGitState      bool

	// Resume flags
	resumeForce     bool
	restoreFiles    bool
	restoreGitState bool
	restoreTickets  bool
	restoreEpics    bool
	backupCurrent   bool

	// Status flags
	statusVerbose bool
	statusFormat  string

	// Clear flags
	clearConfirm bool
	clearBackup  bool
)

func init() {
	rootCmd.AddCommand(interruptCmd)

	// Add subcommands
	interruptCmd.AddCommand(interruptStartCmd)
	interruptCmd.AddCommand(interruptResumeCmd)
	interruptCmd.AddCommand(interruptStatusCmd)
	interruptCmd.AddCommand(interruptClearCmd)

	// interrupt start flags
	interruptStartCmd.Flags().StringVar(&interruptName, "name", "", "Name for the interruption context (required)")
	interruptStartCmd.Flags().StringVar(&interruptDescription, "description", "", "Description of the interruption")
	interruptStartCmd.Flags().StringVar(&interruptType, "type", "interruption", "Type of interruption (interruption, emergency, hotfix, experiment)")
	interruptStartCmd.Flags().StringVar(&interruptNotes, "notes", "", "User notes for the interruption")
	interruptStartCmd.Flags().StringSliceVar(&interruptTags, "tags", []string{}, "Tags for the interruption context")
	interruptStartCmd.Flags().BoolVar(&includeFileState, "include-files", true, "Include file state in context")
	interruptStartCmd.Flags().BoolVar(&includeGitState, "include-git", true, "Include git state in context")
	interruptStartCmd.MarkFlagRequired("name")

	// interrupt resume flags
	interruptResumeCmd.Flags().BoolVar(&resumeForce, "force", false, "Force resume even with conflicts")
	interruptResumeCmd.Flags().BoolVar(&restoreFiles, "restore-files", true, "Restore file state")
	interruptResumeCmd.Flags().BoolVar(&restoreGitState, "restore-git", true, "Restore git state")
	interruptResumeCmd.Flags().BoolVar(&restoreTickets, "restore-tickets", true, "Restore ticket state")
	interruptResumeCmd.Flags().BoolVar(&restoreEpics, "restore-epics", true, "Restore epic state")
	interruptResumeCmd.Flags().BoolVar(&backupCurrent, "backup-current", true, "Backup current context before resuming")

	// interrupt status flags
	interruptStatusCmd.Flags().BoolVar(&statusVerbose, "verbose", false, "Show verbose context information")
	interruptStatusCmd.Flags().StringVar(&statusFormat, "format", "table", "Output format (table, json)")

	// interrupt clear flags
	interruptClearCmd.Flags().BoolVar(&clearConfirm, "confirm", false, "Confirm clearing the interruption stack")
	interruptClearCmd.Flags().BoolVar(&clearBackup, "backup", false, "Create backup before clearing")
	interruptClearCmd.MarkFlagRequired("confirm")
}

func startInterruption(cmd *cobra.Command) {
	// Get current working directory
	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to get working directory: %v\n", err)
		os.Exit(1)
	}

	// Create interruption stack
	stack := workflow.NewInterruptionStack(wd)

	// Validate interruption type
	contextType, err := parseContextType(interruptType)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Invalid interruption type '%s': %v\n", interruptType, err)
		os.Exit(1)
	}

	// Build save options
	saveOptions := workflow.ContextSaveOptions{
		Name:             interruptName,
		Description:      interruptDescription,
		Type:             contextType,
		UserNotes:        interruptNotes,
		Tags:             interruptTags,
		IncludeFileState: includeFileState,
		IncludeGitState:  includeGitState,
	}

	// Check current stack depth for warnings
	currentDepth, err := stack.GetStackDepth()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to check stack depth: %v\n", err)
		os.Exit(1)
	}

	if currentDepth >= 3 {
		fmt.Printf("⚠️  Warning: Deep nesting detected (%d levels). Consider completing some interruptions first.\n\n", currentDepth)
	}

	fmt.Printf("💾 Starting interruption workflow...\n")
	fmt.Printf("   Name:        %s\n", interruptName)
	fmt.Printf("   Type:        %s\n", interruptType)
	if interruptDescription != "" {
		fmt.Printf("   Description: %s\n", interruptDescription)
	}
	fmt.Printf("   Stack depth: %d → %d\n", currentDepth, currentDepth+1)
	fmt.Printf("\n")

	// Save current context
	context, err := stack.SaveCurrentContext(saveOptions)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to start interruption: %v\n", err)
		os.Exit(1)
	}

	// Display success
	fmt.Printf("✅ Interruption started successfully!\n\n")
	fmt.Printf("📝 Context Details:\n")
	fmt.Printf("   Context ID:  %s\n", context.ID)
	fmt.Printf("   Created:     %s\n", context.CreatedAt.Format("2006-01-02 15:04:05"))
	if context.ParentContextID != "" {
		fmt.Printf("   Parent:      %s\n", context.ParentContextID)
	}

	// Show preserved state
	fmt.Printf("\n🔒 Preserved State:\n")
	if context.CurrentEpicID != "" {
		fmt.Printf("   Epic:        %s\n", context.CurrentEpicID)
	}
	if context.CurrentTicketID != "" {
		fmt.Printf("   Ticket:      %s\n", context.CurrentTicketID)
	}
	if includeFileState {
		fmt.Printf("   File state:  Captured\n")
	}
	if includeGitState {
		fmt.Printf("   Git state:   Captured\n")
	}

	fmt.Printf("\n💡 Next steps:\n")
	fmt.Printf("   • Work on your interruption task\n")
	fmt.Printf("   • When finished: claude-wm-cli interrupt resume\n")
	fmt.Printf("   • Check status:  claude-wm-cli interrupt status\n")
}

func resumeInterruption(cmd *cobra.Command, contextID string) {
	// Get current working directory
	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to get working directory: %v\n", err)
		os.Exit(1)
	}

	// Create interruption stack
	stack := workflow.NewInterruptionStack(wd)

	// Check if stack is empty
	stackDepth, err := stack.GetStackDepth()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to check stack depth: %v\n", err)
		os.Exit(1)
	}

	if stackDepth == 0 {
		fmt.Printf("ℹ️  No interruptions to resume. The stack is empty.\n")
		fmt.Printf("\n💡 Available actions:\n")
		fmt.Printf("   • Start work:       claude-wm-cli interrupt start --name \"Task Name\"\n")
		fmt.Printf("   • Check status:     claude-wm-cli interrupt status\n")
		return
	}

	// Build restore options
	restoreOptions := workflow.ContextRestoreOptions{
		RestoreFiles:    restoreFiles,
		RestoreGitState: restoreGitState,
		RestoreTickets:  restoreTickets,
		RestoreEpics:    restoreEpics,
		Force:           resumeForce,
		BackupCurrent:   backupCurrent,
	}

	fmt.Printf("🔄 Resuming workflow...\n")

	var restoredContext *workflow.WorkflowContext

	if contextID != "" {
		// Resume specific context
		fmt.Printf("   Target:      Context %s\n", contextID)
		err = stack.RestoreContext(contextID, restoreOptions)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to resume context %s: %v\n", contextID, err)
			if !resumeForce {
				fmt.Printf("\n💡 Try again with --force to override conflicts\n")
			}
			os.Exit(1)
		}

		// Get the restored context for display
		currentCtx, _ := stack.GetCurrentContext()
		restoredContext = currentCtx
	} else {
		// Pop most recent context
		fmt.Printf("   Target:      Most recent interruption\n")
		var err error
		restoredContext, err = stack.PopContext(restoreOptions)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to resume workflow: %v\n", err)
			if !resumeForce {
				fmt.Printf("\n💡 Try again with --force to override conflicts\n")
			}
			os.Exit(1)
		}
	}

	// Get updated stack depth
	newDepth, _ := stack.GetStackDepth()
	fmt.Printf("   Stack depth: %d → %d\n", stackDepth, newDepth)
	fmt.Printf("\n")

	// Display success
	fmt.Printf("✅ Workflow resumed successfully!\n\n")
	fmt.Printf("📝 Resumed Context:\n")
	fmt.Printf("   Context ID:  %s\n", restoredContext.ID)
	fmt.Printf("   Name:        %s\n", restoredContext.Name)
	fmt.Printf("   Type:        %s\n", restoredContext.Type)
	fmt.Printf("   Saved:       %s\n", restoredContext.SavedAt.Format("2006-01-02 15:04:05"))

	// Show restored state
	fmt.Printf("\n🔓 Restored State:\n")
	if restoredContext.CurrentEpicID != "" {
		fmt.Printf("   Epic:        %s\n", restoredContext.CurrentEpicID)
	}
	if restoredContext.CurrentTicketID != "" {
		fmt.Printf("   Ticket:      %s\n", restoredContext.CurrentTicketID)
	}
	if restoreFiles && restoredContext.Metadata["file_state_captured"] == true {
		fmt.Printf("   File state:  Restored\n")
	}
	if restoreGitState && restoredContext.Metadata["git_state_captured"] == true {
		fmt.Printf("   Git state:   Restored\n")
	}

	if newDepth > 0 {
		fmt.Printf("\n💡 Next steps:\n")
		fmt.Printf("   • Continue work or resume again: claude-wm-cli interrupt resume\n")
		fmt.Printf("   • Check status:                  claude-wm-cli interrupt status\n")
	} else {
		fmt.Printf("\n💡 You're back to your original workflow!\n")
	}
}

func showInterruptionStatus(cmd *cobra.Command) {
	// Get current working directory
	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to get working directory: %v\n", err)
		os.Exit(1)
	}

	// Create interruption stack
	stack := workflow.NewInterruptionStack(wd)

	// Get stack data
	stackData, err := stack.ListContexts()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to get interruption status: %v\n", err)
		os.Exit(1)
	}

	if statusFormat == "json" {
		displayStatusJSON(stackData)
		return
	}

	// Display header
	fmt.Printf("📊 Interruption Stack Status\n")
	fmt.Printf("============================\n\n")

	// Current context
	fmt.Printf("🎯 Current Context:\n")
	if stackData.CurrentContext != nil {
		ctx := stackData.CurrentContext
		fmt.Printf("   Context ID:   %s\n", ctx.ID)
		fmt.Printf("   Name:         %s\n", ctx.Name)
		fmt.Printf("   Type:         %s\n", ctx.Type)
		fmt.Printf("   Created:      %s\n", ctx.CreatedAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("   Last Access:  %s\n", ctx.LastAccessedAt.Format("2006-01-02 15:04:05"))

		if statusVerbose {
			if ctx.Description != "" {
				fmt.Printf("   Description:  %s\n", ctx.Description)
			}
			if ctx.UserNotes != "" {
				fmt.Printf("   Notes:        %s\n", ctx.UserNotes)
			}
			if len(ctx.Tags) > 0 {
				fmt.Printf("   Tags:         %s\n", strings.Join(ctx.Tags, ", "))
			}
			if ctx.CurrentEpicID != "" {
				fmt.Printf("   Epic:         %s\n", ctx.CurrentEpicID)
			}
			if ctx.CurrentTicketID != "" {
				fmt.Printf("   Ticket:       %s\n", ctx.CurrentTicketID)
			}
		}
	} else {
		fmt.Printf("   No current context\n")
	}

	// Stack information
	fmt.Printf("\n📚 Stack Information:\n")
	fmt.Printf("   Stack depth:        %d\n", stackData.Metadata.CurrentStackDepth)
	fmt.Printf("   Active interruptions: %d\n", stackData.Metadata.ActiveInterruptions)
	fmt.Printf("   Total interruptions:  %d\n", stackData.Metadata.TotalInterruptions)
	fmt.Printf("   Max depth reached:    %d\n", stackData.Metadata.MaxStackDepth)
	fmt.Printf("   Last updated:         %s\n", stackData.Metadata.LastUpdated.Format("2006-01-02 15:04:05"))

	// Interruption stack
	if len(stackData.ContextStack) > 0 {
		fmt.Printf("\n⬆️  Interruption Stack (most recent first):\n")
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintf(w, "POSITION\tCONTEXT ID\tNAME\tTYPE\tCREATED\n")
		fmt.Fprintf(w, "────────\t──────────\t────\t────\t───────\n")

		for i := len(stackData.ContextStack) - 1; i >= 0; i-- {
			ctx := stackData.ContextStack[i]
			position := len(stackData.ContextStack) - i
			fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\n",
				position,
				truncateInterruptString(ctx.ID, 12),
				truncateInterruptString(ctx.Name, 20),
				ctx.Type,
				ctx.CreatedAt.Format("Jan 02 15:04"))
		}
		w.Flush()
	} else {
		fmt.Printf("\n⬆️  Interruption Stack: Empty\n")
	}

	// Context history (if verbose)
	if statusVerbose && len(stackData.ContextHistory) > 0 {
		fmt.Printf("\n📜 Context History (last 10):\n")
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintf(w, "CONTEXT ID\tNAME\tTYPE\tCREATED\tPARENT\n")
		fmt.Fprintf(w, "──────────\t────\t────\t───────\t──────\n")

		start := len(stackData.ContextHistory) - 10
		if start < 0 {
			start = 0
		}

		for i := start; i < len(stackData.ContextHistory); i++ {
			ctx := stackData.ContextHistory[i]
			parentID := truncateInterruptString(ctx.ParentContextID, 8)
			if parentID == "" {
				parentID = "-"
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
				truncateInterruptString(ctx.ID, 12),
				truncateInterruptString(ctx.Name, 20),
				ctx.Type,
				ctx.CreatedAt.Format("Jan 02 15:04"),
				parentID)
		}
		w.Flush()
	}

	// Available actions
	fmt.Printf("\n💡 Available Actions:\n")
	if stackData.Metadata.CurrentStackDepth > 0 {
		fmt.Printf("   • Resume workflow:  claude-wm-cli interrupt resume\n")
	}
	fmt.Printf("   • Start interruption: claude-wm-cli interrupt start --name \"Task Name\"\n")
	if stackData.Metadata.CurrentStackDepth > 0 || len(stackData.ContextHistory) > 0 {
		fmt.Printf("   • Clear stack:        claude-wm-cli interrupt clear --confirm\n")
	}
}

func clearInterruptionStack(cmd *cobra.Command) {
	// Get current working directory
	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to get working directory: %v\n", err)
		os.Exit(1)
	}

	// Create interruption stack
	stack := workflow.NewInterruptionStack(wd)

	// Get current stack status
	stackDepth, err := stack.GetStackDepth()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to check stack depth: %v\n", err)
		os.Exit(1)
	}

	if stackDepth == 0 {
		fmt.Printf("ℹ️  Interruption stack is already empty.\n")
		return
	}

	// Create backup if requested
	if clearBackup {
		fmt.Printf("💾 Creating backup before clearing...\n")
		// TODO: Implement backup functionality
		fmt.Printf("   Backup functionality not yet implemented\n")
	}

	fmt.Printf("⚠️  WARNING: This will permanently clear %d interruptions from the stack.\n", stackDepth)
	fmt.Printf("   This action cannot be undone!\n\n")

	if !clearConfirm {
		fmt.Printf("❌ Operation cancelled. Use --confirm to proceed.\n")
		return
	}

	// Clear the stack
	err = stack.ClearStack()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to clear interruption stack: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Interruption stack cleared successfully!\n\n")
	fmt.Printf("📝 Summary:\n")
	fmt.Printf("   Cleared contexts: %d\n", stackDepth)
	fmt.Printf("   Current state:    Clean\n")

	fmt.Printf("\n💡 Next steps:\n")
	fmt.Printf("   • Start new work:   claude-wm-cli interrupt start --name \"Task Name\"\n")
	fmt.Printf("   • Check status:     claude-wm-cli interrupt status\n")
}

// Helper functions

func parseContextType(typeStr string) (workflow.WorkflowContextType, error) {
	switch strings.ToLower(typeStr) {
	case "normal":
		return workflow.WorkflowContextTypeNormal, nil
	case "interruption", "interrupt":
		return workflow.WorkflowContextTypeInterruption, nil
	case "emergency":
		return workflow.WorkflowContextTypeEmergency, nil
	case "hotfix":
		return workflow.WorkflowContextTypeHotfix, nil
	case "experiment":
		return workflow.WorkflowContextTypeExperiment, nil
	default:
		return "", fmt.Errorf("invalid context type '%s'. Valid types: normal, interruption, emergency, hotfix, experiment", typeStr)
	}
}

func displayStatusJSON(stackData *workflow.InterruptionStackData) {
	// TODO: Implement JSON marshaling and output
	// Convert to a more user-friendly JSON structure when implemented
	_ = stackData // Suppress unused parameter warning

	fmt.Printf("JSON output not yet implemented\n")
	fmt.Printf("Use 'claude-wm-cli interrupt status' for table format\n")
}

func truncateInterruptString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}
