package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"claude-wm-cli/internal/debug"
	"claude-wm-cli/internal/errors"
	"claude-wm-cli/internal/executor"
	"claude-wm-cli/internal/navigation"
	"claude-wm-cli/internal/workflow"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// InteractiveCmd represents the interactive command
var InteractiveCmd = &cobra.Command{
	Use:   "interactive",
	Short: "Interactive navigation through project workflow",
	Long: `Interactive provides an interactive menu system to guide you through
the Claude WM project workflow based on your current project state.

The navigation system automatically detects your current workflow position
and suggests appropriate next actions, making it easy to work with the 
project without memorizing commands.

FEATURES:
  • Context-aware menu options based on project state
  • Intelligent action suggestions with priority ranking
  • Visual project status display with progress indicators
  • Simple numbered menu interface with keyboard shortcuts
  • Graceful handling of missing or corrupted state files

SHORTCUTS:
  • 1, 2, 3... - Select numbered menu options
  • q, quit     - Quit navigation
  • b, back     - Go back to previous menu
  • h, help     - Show help information

EXAMPLES:
  claude-wm-cli interactive              # Start interactive navigation
  claude-wm-cli interactive --status     # Show status and exit
  claude-wm-cli interactive --suggest    # Show suggestions and exit`,
	Aliases: []string{"nav", "menu"},
	RunE:    runInteractive,
}

// Navigation command flags
var (
	showStatusOnly  bool
	showSuggestOnly bool
	showQuickStatus bool
	noInteractive   bool
	displayWidth    int
	maxSuggestions  int
)

func init() {
	rootCmd.AddCommand(InteractiveCmd)

	// Add flags for navigation command
	InteractiveCmd.Flags().BoolVar(&showStatusOnly, "status", false, "show project status and exit")
	InteractiveCmd.Flags().BoolVar(&showSuggestOnly, "suggest", false, "show suggestions and exit")
	InteractiveCmd.Flags().BoolVar(&showQuickStatus, "quick", false, "show quick one-line status")
	InteractiveCmd.Flags().BoolVar(&noInteractive, "no-interactive", false, "disable interactive mode")
	InteractiveCmd.Flags().IntVar(&displayWidth, "width", 80, "display width for formatting")
	InteractiveCmd.Flags().IntVar(&maxSuggestions, "max-suggestions", 5, "maximum number of suggestions to show")

	// Bind flags to viper
	viper.BindPFlag("interactive.status", InteractiveCmd.Flags().Lookup("status"))
	viper.BindPFlag("interactive.suggest", InteractiveCmd.Flags().Lookup("suggest"))
	viper.BindPFlag("interactive.quick", InteractiveCmd.Flags().Lookup("quick"))
	viper.BindPFlag("interactive.no-interactive", InteractiveCmd.Flags().Lookup("no-interactive"))
	viper.BindPFlag("interactive.width", InteractiveCmd.Flags().Lookup("width"))
	viper.BindPFlag("interactive.max-suggestions", InteractiveCmd.Flags().Lookup("max-suggestions"))
}

// runInteractive executes the interactive command
func runInteractive(cmd *cobra.Command, args []string) error {
	// Enable debug mode if flag is set
	debug.SetDebugMode(debugMode || viper.GetBool("debug"))
	
	debug.LogExecution("INTERACTIVE", "start navigation", "Initialize interactive menu system")
	
	// Get current working directory
	workDir, err := os.Getwd()
	if err != nil {
		return errors.NewCLIError("Failed to get current directory", 1).
			WithDetails(err.Error()).
			WithSuggestion("Ensure you have proper permissions to access the current directory")
	}

	// Initialize navigation components
	contextDetector := navigation.NewContextDetector(workDir)
	suggestionEngine := navigation.NewSuggestionEngine()
	menuDisplay := navigation.NewMenuDisplay()
	stateDisplay := navigation.NewProjectStateDisplay()

	// Set display width from flag
	stateDisplay.SetWidth(displayWidth)

	// Detect current project context
	projectContext, err := contextDetector.DetectContext()
	if err != nil {
		return errors.NewCLIError("Failed to detect project context", 1).
			WithDetails(err.Error()).
			WithSuggestion("Check that you're in a valid directory and have necessary permissions").
			WithContext("directory", workDir)
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "Detected project state: %s\n", projectContext.State.String())
		if len(projectContext.Issues) > 0 {
			fmt.Fprintf(os.Stderr, "Project issues detected: %d\n", len(projectContext.Issues))
		}
	}

	// Handle quick status flag
	if showQuickStatus {
		stateDisplay.DisplayQuickStatus(projectContext)
		return nil
	}

	// Handle status-only flag
	if showStatusOnly {
		stateDisplay.DisplayProjectOverview(projectContext)
		return nil
	}

	// Generate suggestions
	suggestions, err := suggestionEngine.GenerateSuggestions(projectContext)
	if err != nil {
		return errors.NewCLIError("Failed to generate suggestions", 1).
			WithDetails(err.Error()).
			WithSuggestion("Check project state and try again")
	}

	// Limit suggestions if requested
	if len(suggestions) > maxSuggestions {
		suggestions = suggestions[:maxSuggestions]
	}

	// Handle suggest-only flag
	if showSuggestOnly {
		displaySuggestions(suggestions, suggestionEngine)
		return nil
	}

	// Handle non-interactive mode
	if noInteractive {
		stateDisplay.DisplayWithSuggestions(projectContext, suggestions)
		return nil
	}

	// Start interactive navigation
	return runInteractiveNavigation(projectContext, suggestions, menuDisplay, stateDisplay, suggestionEngine)
}

// runInteractiveNavigation handles the interactive menu navigation
func runInteractiveNavigation(
	ctx *navigation.ProjectContext,
	suggestions []*navigation.Suggestion,
	menuDisplay *navigation.MenuDisplay,
	stateDisplay *navigation.ProjectStateDisplay,
	suggestionEngine *navigation.SuggestionEngine,
) error {
	for {
		// Display current state
		stateDisplay.DisplayProjectOverview(ctx)

		// Create main menu
		menu := createMainMenu(ctx, suggestions)

		// Show menu and get user choice
		result, err := menuDisplay.Show(menu)
		if err != nil {
			return errors.NewCLIError("Menu interaction failed", 1).
				WithDetails(err.Error()).
				WithSuggestion("Try restarting the navigation or check terminal compatibility")
		}

		// Handle menu result
		switch result.Action {
		case "quit":
			menuDisplay.ShowMessage("👋 Goodbye!")
			return nil

		case "help":
			displayNavigationHelp(menuDisplay)

		case "status":
			stateDisplay.DisplayProjectOverview(ctx)
			menuDisplay.WaitForKeyPress("")

		case "suggestions":
			displaySuggestions(suggestions, suggestionEngine)
			menuDisplay.WaitForKeyPress("")

		case "refresh":
			// Re-detect context and regenerate suggestions
			newCtx, err := navigation.NewContextDetector(ctx.ProjectPath).DetectContext()
			if err != nil {
				menuDisplay.ShowError(fmt.Sprintf("Failed to refresh context: %v", err))
				menuDisplay.WaitForKeyPress("")
				continue
			}
			ctx = newCtx

			newSuggestions, err := suggestionEngine.GenerateSuggestions(ctx)
			if err != nil {
				menuDisplay.ShowError(fmt.Sprintf("Failed to refresh suggestions: %v", err))
				menuDisplay.WaitForKeyPress("")
				continue
			}
			suggestions = newSuggestions
			menuDisplay.ShowSuccess("Context refreshed!")

		default:
			// Handle action execution
			err := executeAction(result.Action, ctx, menuDisplay)
			if err != nil {
				menuDisplay.ShowError(fmt.Sprintf("Failed to execute action: %v", err))
				menuDisplay.WaitForKeyPress("")
			}
		}
	}
}

// createMainMenu builds the main navigation menu based on context
func createMainMenu(ctx *navigation.ProjectContext, _ []*navigation.Suggestion) *navigation.Menu {
	menu := &navigation.Menu{
		Title:       "🧭 Claude WM CLI Navigation",
		Options:     []navigation.MenuOption{},
		ShowNumbers: true,
		ShowHelp:    true,
		AllowBack:   true,
		AllowQuit:   true,
	}


	// Helper function to add regular option
	addOption := func(id, label, description, action string) {
		menu.Options = append(menu.Options, navigation.MenuOption{
			ID:          id,
			Label:       label,
			Description: description,
			Action:      action,
			Enabled:     true,
		})
	}

	// === PROJECT UPDATE CYCLE === (always available)
	menu.Options = append(menu.Options, navigation.MenuOption{
		ID:          "section-header-project",
		Label:       "📋 PROJECT UPDATE CYCLE",
		Description: "",
		Action:      "",
		Enabled:     false,
	})
	addOption("project-import-feedback", "🔄 Import Feedback", "Import and process feedback from FEEDBACK.md", "project-import-feedback")
	addOption("project-challenge", "🤔 Challenge Docs", "Challenge existing documentation and assumptions", "project-challenge")
	addOption("project-enrich", "🌟 Enrich Context", "Enrich project context with additional information", "project-enrich")
	addOption("project-status-update", "📊 Status Update", "Update overall project status", "project-status-update")
	addOption("project-implementation-status", "🔍 Implementation Status", "Review and update implementation progress", "project-implementation-status")

	// === EPIC MANAGEMENT === (only if project is initialized)
	if ctx != nil && ctx.State != navigation.StateNotInitialized {
		menu.Options = append(menu.Options, navigation.MenuOption{
			ID:          "section-header-epic",
			Label:       "📚 EPIC MANAGEMENT",
			Description: "",
			Action:      "",
			Enabled:     false,
		})
		addOption("project-plan-epics", "📝 Plan Epics", "Plan and manage epic roadmap (create/update epics.json)", "project-plan-epics")
		addOption("epic-list", "📋 List Epics", "List all available epics", "epic-list")
	}

	// === CURRENT EPIC === (only if we have an active epic)
	if ctx != nil && ctx.CurrentEpic != nil {
		menu.Options = append(menu.Options, navigation.MenuOption{
			ID:          "section-header-current-epic",
			Label:       "🚧 CURRENT EPIC MANAGEMENT",
			Description: "",
			Action:      "",
			Enabled:     false, // Explicitly disabled
		})
		addOption("epic-select-stories", "🎯 Select Stories", "Select the most important story to work on", "/2-epic:1-start:1-Select-Stories")
		addOption("epic-plan-stories", "📋 Plan Stories", "Plan and organize stories for the epic", "/2-epic:1-start:2-Plan-stories")
		addOption("epic-complete", "✅ Complete Epic", "Mark epic as complete and archive", "/2-epic:2-manage:1-Complete-Epic")
		addOption("epic-status", "📊 Epic Status", "View current epic progress and status", "/2-epic:2-manage:2-Status-Epic")
	}

	// === TICKETS/INTERRUPTIONS === (always available)
	menu.Options = append(menu.Options, navigation.MenuOption{
		ID:          "section-header-tickets",
		Label:       "🎫 TICKETS & INTERRUPTIONS",
		Description: "",
		Action:      "",
		Enabled:     false, // Explicitly disabled
	})
	addOption("ticket-create", "🆘 Create Ticket", "Create interruption ticket or urgent task", "ticket-create")
	addOption("ticket-list", "📋 List Tickets", "List current tickets and interruptions", "ticket-list")
	if ctx != nil && ctx.CurrentTask != nil {
		addOption("ticket-current", "🎯 Current Ticket", "Work on current active ticket", "ticket-current")
	}

	// === SYSTEM === (always available)
	menu.Options = append(menu.Options, navigation.MenuOption{
		ID:          "section-header-system",
		Label:       "⚙️ SYSTEM",
		Description: "",
		Action:      "",
		Enabled:     false, // Explicitly disabled
	})
	addOption("status", "📊 Project Status", "Display detailed project state and progress", "status")
	addOption("suggestions", "💡 View Suggestions", "Show AI-generated action suggestions", "suggestions")
	addOption("refresh", "🔄 Refresh", "Re-scan project state and update context", "refresh")

	return menu
}

// executeAction handles the execution of selected actions
func executeAction(action string, ctx *navigation.ProjectContext, menuDisplay *navigation.MenuDisplay) error {
	switch action {
	// Project Update Cycle
	case "project-import-feedback":
		return executeProjectCommand([]string{"import-feedback"}, menuDisplay)
	case "project-challenge":
		return executeProjectCommand([]string{"challenge"}, menuDisplay)
	case "project-enrich":
		return executeProjectCommand([]string{"enrich"}, menuDisplay)
	case "project-status-update":
		return executeProjectCommand([]string{"status-update"}, menuDisplay)
	case "project-implementation-status":
		return executeProjectCommand([]string{"implementation-status"}, menuDisplay)
	case "project-plan-epics":
		return executeProjectCommand([]string{"plan-epics"}, menuDisplay)

	// Epic Management
	case "epic-list":
		return executeEpicCommand([]string{"list"}, menuDisplay)

	// Current Epic Management - Claude Commands
	case "/2-epic:1-start:1-Select-Stories":
		return executeClaudeCommandInteractive("/2-epic:1-start:1-Select-Stories", menuDisplay)
	case "/2-epic:1-start:2-Plan-stories":
		return executeClaudeCommandInteractive("/2-epic:1-start:2-Plan-stories", menuDisplay)
	case "/2-epic:2-manage:1-Complete-Epic":
		return executeClaudeCommandInteractive("/2-epic:2-manage:1-Complete-Epic", menuDisplay)
	case "/2-epic:2-manage:2-Status-Epic":
		return executeClaudeCommandInteractive("/2-epic:2-manage:2-Status-Epic", menuDisplay)

	// Task Management
	case "task-create":
		menuDisplay.ShowMessage("Create task functionality - use story commands to generate tasks")
		return nil

	// Ticket Management
	case "ticket-create":
		return executeTicketCommand([]string{"create"}, menuDisplay)
	case "ticket-list":
		return executeTicketCommand([]string{"list"}, menuDisplay)
	case "ticket-current":
		return executeTicketCommand([]string{"current"}, menuDisplay)

	// Legacy actions
	case "init-project":
		return executeInitProject(ctx, menuDisplay)

	default:
		menuDisplay.ShowWarning(fmt.Sprintf("Action '%s' not yet implemented", action))
		menuDisplay.ShowMessage("This action will be available in a future version.")
		return nil
	}
}

// executeInitProject handles project initialization
func executeInitProject(ctx *navigation.ProjectContext, menuDisplay *navigation.MenuDisplay) error {
	confirmed, err := menuDisplay.Confirm("Initialize project structure in current directory?")
	if err != nil {
		return err
	}

	if !confirmed {
		menuDisplay.ShowMessage("Project initialization cancelled")
		return nil
	}

	// Create basic project structure
	dirs := []string{
		"docs/1-project",
		"docs/2-current-epic",
		"docs/3-current-task",
	}

	for _, dir := range dirs {
		fullPath := filepath.Join(ctx.ProjectPath, dir)
		if err := os.MkdirAll(fullPath, 0755); err != nil {
			return errors.NewCLIError("Failed to create project directory", 1).
				WithDetails(err.Error()).
				WithContext("directory", fullPath)
		}
	}

	menuDisplay.ShowSuccess("✅ Project structure initialized!")
	menuDisplay.ShowMessage("You can now create your first epic.")
	return nil
}


// Helper functions to execute CLI commands

// executeProjectCommand executes a project subcommand
func executeProjectCommand(args []string, menuDisplay *navigation.MenuDisplay) error {
	// Use exec.Command to run the command in a subprocess to avoid stdin conflicts
	cmdArgs := append([]string{"project"}, args...)
	
	// Get the path to the current binary
	execPath, err := os.Executable()
	if err != nil {
		menuDisplay.ShowError(fmt.Sprintf("Failed to get executable path: %v", err))
		return err
	}
	
	// Use the build binary instead of the current executable if we're in development
	buildPath := filepath.Join(filepath.Dir(filepath.Dir(execPath)), "build", "claude-wm-cli")
	if _, err := os.Stat(buildPath); err == nil {
		execPath = buildPath
	}
	
	// Debug logging
	debug.LogCommandWithArgs("PROJECT", fmt.Sprintf("Execute project command: %s", args[0]), execPath, cmdArgs)
	
	// Add debug flag to subprocess if enabled
	if debugMode || viper.GetBool("debug") {
		cmdArgs = append(cmdArgs, "--debug")
	}
	
	cmd := exec.Command(execPath, cmdArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Run(); err != nil {
		debug.LogResult("PROJECT", fmt.Sprintf("project %s", args[0]), fmt.Sprintf("Command failed: %v", err), false)
		menuDisplay.ShowError(fmt.Sprintf("Failed to execute project %s: %v", args[0], err))
		return err
	}
	
	debug.LogResult("PROJECT", fmt.Sprintf("project %s", args[0]), "Command completed successfully", true)
	menuDisplay.ShowSuccess(fmt.Sprintf("✅ Project %s completed successfully", args[0]))
	return nil
}

// executeEpicCommand executes an epic subcommand
func executeEpicCommand(args []string, menuDisplay *navigation.MenuDisplay) error {
	cmdArgs := append([]string{"epic"}, args...)
	
	execPath, err := os.Executable()
	if err != nil {
		menuDisplay.ShowError(fmt.Sprintf("Failed to get executable path: %v", err))
		return err
	}
	
	buildPath := filepath.Join(filepath.Dir(filepath.Dir(execPath)), "build", "claude-wm-cli")
	if _, err := os.Stat(buildPath); err == nil {
		execPath = buildPath
	}
	
	cmd := exec.Command(execPath, cmdArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Run(); err != nil {
		menuDisplay.ShowError(fmt.Sprintf("Failed to execute epic %s: %v", args[0], err))
		return err
	}
	
	menuDisplay.ShowSuccess(fmt.Sprintf("✅ Epic %s completed successfully", args[0]))
	return nil
}


// executeTicketCommand executes a ticket subcommand  
func executeTicketCommand(args []string, menuDisplay *navigation.MenuDisplay) error {
	cmdArgs := append([]string{"ticket"}, args...)
	
	execPath, err := os.Executable()
	if err != nil {
		menuDisplay.ShowError(fmt.Sprintf("Failed to get executable path: %v", err))
		return err
	}
	
	buildPath := filepath.Join(filepath.Dir(filepath.Dir(execPath)), "build", "claude-wm-cli")
	if _, err := os.Stat(buildPath); err == nil {
		execPath = buildPath
	}
	
	cmd := exec.Command(execPath, cmdArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Run(); err != nil {
		menuDisplay.ShowError(fmt.Sprintf("Failed to execute ticket %s: %v", args[0], err))
		return err
	}
	
	menuDisplay.ShowSuccess(fmt.Sprintf("✅ Ticket %s completed successfully", args[0]))
	return nil
}

// executeClaudeCommandInteractive executes a Claude slash command from interactive menu
func executeClaudeCommandInteractive(command string, menuDisplay *navigation.MenuDisplay) error {
	menuDisplay.ShowMessage(fmt.Sprintf("🚀 Executing Claude command: %s", command))
	
	// Create Claude executor
	claudeExecutor := executor.NewClaudeExecutor()
	
	// Validate Claude is available
	if err := claudeExecutor.ValidateClaudeAvailable(); err != nil {
		menuDisplay.ShowError(fmt.Sprintf("Claude CLI not available: %v", err))
		menuDisplay.ShowMessage("💡 Please install Claude CLI to use this functionality")
		return err
	}
	
	// Execute the slash command
	description := fmt.Sprintf("Interactive menu command: %s", command)
	if err := claudeExecutor.ExecuteSlashCommand(command, description); err != nil {
		menuDisplay.ShowError(fmt.Sprintf("Failed to execute Claude command: %v", err))
		return err
	}
	
	menuDisplay.ShowSuccess(fmt.Sprintf("✅ Claude command %s completed successfully", command))
	return nil
}

// displaySuggestions shows all suggestions in a formatted way
func displaySuggestions(suggestions []*navigation.Suggestion, engine *navigation.SuggestionEngine) {
	if len(suggestions) == 0 {
		fmt.Println("No suggestions available")
		return
	}

	fmt.Printf("\n💡 Action Suggestions (%d):\n\n", len(suggestions))

	for i, suggestion := range suggestions {
		formatted := engine.FormatSuggestion(suggestion, true)
		fmt.Printf("  %d. %s\n", i+1, formatted)

		if len(suggestion.NextActions) > 0 {
			fmt.Printf("     → Next: %s\n", suggestion.NextActions[0])
		}
		fmt.Println()
	}
}

// displayNavigationHelp shows help information for navigation
func displayNavigationHelp(menuDisplay *navigation.MenuDisplay) {
	help := `
🧭 Navigation Help

KEYBOARD SHORTCUTS:
  • Numbers (1,2,3...)  - Select menu options
  • q, quit, exit       - Quit navigation
  • b, back            - Go back to previous menu  
  • h, help            - Show this help

MENU ACTIONS:
  • Project Status     - View detailed project state
  • Suggestions        - See all available actions
  • Refresh Context    - Re-scan project state

WORKFLOW STATES:
  • Not Initialized    - Project needs setup
  • Project Init       - Ready for epics
  • Has Epics         - Ready to start work
  • Epic In Progress   - Working on current epic
  • Story In Progress  - Working on current story
  • Task In Progress   - Working on current task

TIP: The navigation system automatically detects your current
     state and suggests the most appropriate next actions.
`
	menuDisplay.ShowMessage(help)
}

// Helper functions

// getPriorityIcon returns an icon for workflow priorities
func getPriorityIcon(priority workflow.Priority) string {
	switch priority {
	case workflow.PriorityP0:
		return "🔴 "
	case workflow.PriorityP1:
		return "🟡 "
	case workflow.PriorityP2:
		return "🟢 "
	default:
		return "⚪ "
	}
}

// truncateString truncates a string to the specified length
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
