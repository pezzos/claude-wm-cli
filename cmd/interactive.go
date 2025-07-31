package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"claude-wm-cli/internal/config"
	"claude-wm-cli/internal/debug"
	"claude-wm-cli/internal/errors"
	"claude-wm-cli/internal/executor"
	"claude-wm-cli/internal/metrics"
	"claude-wm-cli/internal/navigation"
	"claude-wm-cli/internal/preprocessing"
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
	// Start performance monitoring for interactive command
	timer := metrics.InstrumentCommandInteractive("interactive")
	defer timer.Stop()

	// Enable debug mode if flag is set
	debug.SetDebugMode(debugMode || viper.GetBool("debug"))

	debug.LogExecution("INTERACTIVE", "start navigation", "Initialize interactive menu system")

	// Step 1: Working directory detection
	workDirStep := timer.ProfileStep("working_directory_detection")
	workDir, err := os.Getwd()
	if err != nil {
		workDirStep.StopWithError(err)
		timer.SetExitCode(1)
		return errors.NewCLIError("Failed to get current directory", 1).
			WithDetails(err.Error()).
			WithSuggestion("Ensure you have proper permissions to access the current directory")
	}
	workDirStep.SetMetadata("working_directory", workDir)
	workDirStep.Stop()

	// Step 2: Initialize navigation components
	initStep := timer.ProfileStep("navigation_initialization")
	contextDetector := navigation.NewContextDetector(workDir)
	suggestionEngine := navigation.NewSuggestionEngine()
	menuDisplay := navigation.NewMenuDisplay()
	stateDisplay := navigation.NewProjectStateDisplay()

	// Set display width from flag
	stateDisplay.SetWidth(displayWidth)
	initStep.SetMetadata("display_width", displayWidth)
	initStep.Stop()

	// Step 3: Detect current project context
	contextStep := timer.ProfileContextDetection(workDir)
	projectContext, err := contextDetector.DetectContext()
	if err != nil {
		contextStep.StopWithError(err)
		timer.SetExitCode(1)
		return errors.NewCLIError("Failed to detect project context", 1).
			WithDetails(err.Error()).
			WithSuggestion("Check that you're in a valid directory and have necessary permissions").
			WithContext("directory", workDir)
	}
	contextStep.SetMetadata("project_state", projectContext.State.String())
	contextStep.SetMetadata("issues_count", len(projectContext.Issues))
	contextStep.Stop()

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

// runInteractiveNavigation handles the interactive menu navigation with hierarchical support
func runInteractiveNavigation(
	ctx *navigation.ProjectContext,
	suggestions []*navigation.Suggestion,
	menuDisplay *navigation.MenuDisplay,
	stateDisplay *navigation.ProjectStateDisplay,
	suggestionEngine *navigation.SuggestionEngine,
) error {
	// Stack to track menu navigation
	var menuStack []string
	currentMenu := "main"

	for {
		// Display current state
		stateDisplay.DisplayProjectOverview(ctx)

		// Create appropriate menu based on current location
		var menu *navigation.Menu
		switch currentMenu {
		case "main":
			menu = createMainMenu(ctx, suggestions)
		case "project":
			menu = createProjectMenu(ctx)
		case "epics":
			menu = createEpicsMenu(ctx)
		case "current-epics":
			menu = createCurrentEpicMenu(ctx)
		case "current-story":
			menu = createCurrentStoryMenu(ctx)
		case "ticket":
			menu = createTicketMenu(ctx)
		case "claude":
			menu = createClaudeMenu(ctx)
		case "metrics":
			menu = createMetricsMenu(ctx)
		default:
			menu = createMainMenu(ctx, suggestions)
			currentMenu = "main"
		}

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

		case "back":
			// Navigate back to previous menu
			if len(menuStack) > 0 {
				currentMenu = menuStack[len(menuStack)-1]
				menuStack = menuStack[:len(menuStack)-1]
			} else {
				currentMenu = "main"
			}

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

		// Menu navigation actions
		case "project-menu":
			menuStack = append(menuStack, currentMenu)
			currentMenu = "project"

		case "epics-menu":
			menuStack = append(menuStack, currentMenu)
			currentMenu = "epics"

		case "current-epic-menu":
			menuStack = append(menuStack, currentMenu)
			currentMenu = "current-epics"

		case "current-story-menu":
			menuStack = append(menuStack, currentMenu)
			currentMenu = "current-story"

		case "ticket-menu":
			menuStack = append(menuStack, currentMenu)
			currentMenu = "ticket"

		case "claude-menu":
			menuStack = append(menuStack, currentMenu)
			currentMenu = "claude"

		case "metrics-menu":
			menuStack = append(menuStack, currentMenu)
			currentMenu = "metrics"

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

// createMainMenu builds the main navigation menu with hierarchical groups
func createMainMenu(_ *navigation.ProjectContext, _ []*navigation.Suggestion) *navigation.Menu {
	menu := &navigation.Menu{
		Title:       "🧭 Claude WM CLI Navigation",
		Options:     []navigation.MenuOption{},
		ShowNumbers: true,
		ShowHelp:    true,
		AllowBack:   false, // No back button for main menu
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

	// Main menu groups
	addOption("project-menu", "Project update cycle", "Init/Update", "project-menu")
	addOption("epics-menu", "Epics management", "Plan/Track", "epics-menu")
	addOption("current-epic-menu", "Current epic management", "Start epic/Plan stories/Complete epic", "current-epic-menu")
	addOption("current-story-menu", "Current story management", "Start story/Complete story", "current-story-menu")
	addOption("ticket-menu", "Ticket management", "Create/Plan/Execute/Complete", "ticket-menu")
	addOption("metrics-menu", "Performance metrics", "Analyze/Profile/Optimize", "metrics-menu")
	addOption("claude-menu", ".claude management", "Import/Install", "claude-menu")

	return menu
}

// createProjectMenu builds the project update cycle submenu
func createProjectMenu(_ *navigation.ProjectContext) *navigation.Menu {
	menu := &navigation.Menu{
		Title:       "📋 Project Update Cycle",
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

	// Init section header
	menu.Options = append(menu.Options, navigation.MenuOption{
		ID:          "init-header",
		Label:       "Init the Project",
		Description: "",
		Action:      "",
		Enabled:     false,
	})
	addOption("project-init", "🚀", "Initialize the current project", "init-project")

	// Update section header
	menu.Options = append(menu.Options, navigation.MenuOption{
		ID:          "update-header",
		Label:       "Update the Project",
		Description: "",
		Action:      "",
		Enabled:     false,
	})
	addOption("project-import-feedback", "🔄", "Import and process feedback from FEEDBACK.md", "/1-project:2-update:1-Import-feedback")
	addOption("project-challenge", "🤔", "Challenge existing documentation and assumptions", "/1-project:2-update:2-Challenge")
	addOption("project-enrich", "🌟", "Enrich project context with additional information", "/1-project:2-update:3-Enrich")
	addOption("project-status-update", "📊", "Analyze epics and provide project status with completion metrics and next actions. ", "/1-project:2-update:4-Status")
	addOption("project-implementation-status", "🔍", "Display working features, integration points, coverage", "/1-project:2-update:5-Implementation-Status")

	return menu
}

// createEpicsMenu builds the epics management submenu
func createEpicsMenu(_ *navigation.ProjectContext) *navigation.Menu {
	menu := &navigation.Menu{
		Title:       "📚 Epics Management",
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

	// Epics management options
	addOption("epics-plan", "📝 Plan Epics", "Create comprehensive EPICS.md with prioritized epic roadmap", "/1-project:3-epics:1-Plan-Epics")
	addOption("epics-update-implementation", "📊 Update Implementation", "Track and update epic implementation status across the project", "/1-project:3-epics:2-Update-Implementation")
	addOption("epic-list", "📋 List Epics", "List all available epics with status and progress", "epic-list")

	return menu
}

// createCurrentEpicMenu builds the epics management submenu
func createCurrentEpicMenu(_ *navigation.ProjectContext) *navigation.Menu {
	menu := &navigation.Menu{
		Title:       "📚 Current Epic Management",
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

	// Epics management options
	addOption("epic-select", "🎯 Select Epic", "Select the most important story to work on", "/2-epic:1-start:1-Select-Stories")
	addOption("epic-plan-stories", "📝 Plan Stories", "Plan and organize stories for the epic", "/2-epic:1-start:2-Plan-stories")
	addOption("story-list", "📋 List Stories", "List all stories in current epic with status and progress", "story-list")
	addOption("epic-complete", "✅ Complete Epic", "Mark epic as complete and archive", "/2-epic:2-manage:1-Complete-Epic")

	return menu
}

// createCurrentStoryMenu builds the current story management submenu
func createCurrentStoryMenu(_ *navigation.ProjectContext) *navigation.Menu {
	menu := &navigation.Menu{
		Title:       "📖 Current Story Management",
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

	// Current story management options
	addOption("story-start", "🚀 Start Story", "Identify highest priority unstarted story and start implementation", "/3-story:1-manage:1-Start-Story")
	addOption("task-list", "📋 List Tasks", "List all tasks in current story with status and priority", "task-list")
	addOption("story-complete", "✅ Complete Story", "Mark story complete and prepare for next story or epic completion", "/3-story:1-manage:2-Complete-Story")

	return menu
}

// createTicketMenu builds the ticket management submenu
func createTicketMenu(_ *navigation.ProjectContext) *navigation.Menu {
	menu := &navigation.Menu{
		Title:       "🎫 Ticket Management",
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

	// Create section header
	menu.Options = append(menu.Options, navigation.MenuOption{
		ID:          "create-header",
		Label:       "Create",
		Description: "",
		Action:      "",
		Enabled:     false,
	})
	addOption("ticket-from-story", "📋 From Story", "Generate implementation ticket from current story", "ticket-from-story")
	addOption("ticket-from-issue", "🐛 From Issue", "Create ticket from GitHub issue with analysis", "ticket-from-issue")
	addOption("ticket-from-input", "✏️  From Input", "Create custom ticket from direct user input", "ticket-from-input")

	// Execute section header
	menu.Options = append(menu.Options, navigation.MenuOption{
		ID:          "execute-header",
		Label:       "Execute",
		Description: "",
		Action:      "",
		Enabled:     false,
	})
	addOption("ticket-plan", "📝 Plan Ticket", "Create detailed implementation plan with research", "ticket-plan")
	addOption("ticket-test-design", "🧪 Test Design", "Design comprehensive test strategy", "ticket-test-design")
	addOption("ticket-implement", "⚡ Implement", "Execute intelligent implementation with MCP workflow", "/4-task:2-execute:3-Implement")
	addOption("ticket-validate", "✅ Validate", "Validate implementation against acceptance criteria", "ticket-validate")
	addOption("ticket-review", "👀 Review", "Final code review and quality assurance", "ticket-review")

	// Complete section header
	menu.Options = append(menu.Options, navigation.MenuOption{
		ID:          "complete-header",
		Label:       "Complete",
		Description: "",
		Action:      "",
		Enabled:     false,
	})
	addOption("ticket-execute-full", "⚡ Complete the current ticket", "Execute full workflow: Plan → Test → Implement → Validate → Review", "ticket-execute-full")
	addOption("ticket-execute-full-from-story", "⚡ Complete the current ticket from Story", "Execute full workflow: From Story → Plan → Test → Implement → Validate → Review", "ticket-execute-full-from-story")
	addOption("ticket-execute-full-from-issue", "⚡ Complete the current ticket from Issue", "Execute full workflow: From Issue → Plan → Test → Implement → Validate → Review", "ticket-execute-full-from-issue")
	addOption("ticket-execute-full-from-input", "⚡ Complete the current ticket from Input", "Execute full workflow: From Input → Plan → Test → Implement → Validate → Review", "ticket-execute-full-from-input")
	addOption("ticket-archive", "📦 Archive", "Archive completed ticket with summary", "ticket-archive")
	addOption("ticket-status", "📊 Status", "Update ticket status across documentation", "ticket-status")

	return menu
}

// createClaudeMenu builds the Claude management submenu
func createClaudeMenu(_ *navigation.ProjectContext) *navigation.Menu {
	menu := &navigation.Menu{
		Title:       "⚙️ .claude Management",
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

	// Configuration management options
	addOption("config-init", "🚀 Initialize configuration", "Set up package manager configuration structure", "config-init")
	addOption("config-sync", "🔄 Sync configuration", "Regenerate runtime config from templates and overrides", "config-sync")
	addOption("config-upgrade", "⬆️ Upgrade templates", "Update system templates while preserving customizations", "config-upgrade")

	return menu
}

// createMetricsMenu builds the Performance metrics submenu
func createMetricsMenu(_ *navigation.ProjectContext) *navigation.Menu {
	menu := &navigation.Menu{
		Title:       "📊 Performance Metrics",
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

	// Overview section header
	menu.Options = append(menu.Options, navigation.MenuOption{
		ID:          "overview-header",
		Label:       "Overview",
		Description: "",
		Action:      "",
		Enabled:     false,
	})
	addOption("metrics-status", "🔍 Status", "Show metrics collection status and database info", "metrics-status")
	addOption("metrics-commands", "📋 Commands", "List all commands with execution statistics", "metrics-commands")
	addOption("metrics-slow", "🐌 Slow Commands", "Show slowest commands that need optimization", "metrics-slow")

	// Analysis section header
	menu.Options = append(menu.Options, navigation.MenuOption{
		ID:          "analysis-header",
		Label:       "Detailed Analysis",
		Description: "",
		Action:      "",
		Enabled:     false,
	})
	addOption("metrics-command", "⚡ Analyze Command", "Detailed statistics for a specific command", "metrics-command")
	addOption("metrics-steps", "🔬 Analyze Steps", "Step-by-step performance breakdown", "metrics-steps")
	addOption("metrics-projects", "📈 Projects Comparison", "Compare performance across different projects", "metrics-projects")

	return menu
}

// executeAction handles the execution of selected actions
func executeAction(action string, ctx *navigation.ProjectContext, menuDisplay *navigation.MenuDisplay) error {
	switch action {
	// Claude slash commands - can start with '/'
	case "/1-project:1-start:1-Init-Project",
		"/1-project:2-update:1-Import-feedback",
		"/1-project:2-update:2-Challenge",
		"/1-project:2-update:3-Enrich",
		"/1-project:2-update:4-Status",
		"/1-project:2-update:5-Implementation-Status",
		"/1-project:3-epics:1-Plan-Epics",
		"/1-project:3-epics:2-Update-Implementation",
		"/2-epic:1-start:1-Select-Stories",
		"/2-epic:1-start:2-Plan-stories",
		"/2-epic:2-manage:1-Complete-Epic",
		"/2-epic:2-manage:2-Status-Epic",
		"/3-story:1-manage:1-Start-Story",
		"/3-story:1-manage:2-Complete-Story",
		"/4-task:1-start:1-From-story",
		"/4-task:1-start:2-From-issue",
		"/4-task:1-start:3-From-input",
		"/4-task:2-execute:1-Plan-Ticket",
		"/4-task:2-execute:2-Test-design",
		"/4-task:2-execute:3-Implement",
		"/4-task:2-execute:4-Validate-Ticket",
		"/4-task:2-execute:5-Review-Ticket",
		"/4-task:3-complete:1-Archive-Ticket",
		"/4-task:3-complete:2-Status-Ticket":
		return executeClaudeCommandInteractive(action, menuDisplay)

	// Legacy project actions (keeping for backward compatibility)
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

	// Story Management
	case "story-list":
		return executeStoryCommand([]string{"list"}, menuDisplay)

	// Task Management with Preprocessing
	case "ticket-from-story":
		return executeTaskFromStory(ctx, menuDisplay)
	case "ticket-from-issue":
		return executeTaskFromIssue(ctx, menuDisplay)
	case "ticket-from-input":
		return executeTaskFromInput(ctx, menuDisplay)
	case "ticket-plan":
		return executeTaskPlan(ctx, menuDisplay)
	case "ticket-test-design":
		return executeTaskTestDesign(ctx, menuDisplay)
	case "ticket-validate":
		return executeTaskValidate(ctx, menuDisplay)
	case "ticket-review":
		return executeTaskReview(ctx, menuDisplay)
	case "ticket-archive":
		return executeTaskArchive(ctx, menuDisplay)
	case "ticket-status":
		return executeTaskStatus(ctx, menuDisplay)

	// Legacy Ticket Management (keeping for compatibility)
	case "ticket-create":
		return executeTicketCommand([]string{"create"}, menuDisplay)
	case "task-list":
		return executeTaskListFromStory(ctx, menuDisplay)
	case "ticket-current":
		return executeTicketCommand([]string{"current"}, menuDisplay)
	case "ticket-execute-full":
		return executeTicketFullWorkflow(ctx, menuDisplay, "")
	case "ticket-execute-full-from-story":
		return executeTicketFullWorkflow(ctx, menuDisplay, "story")
	case "ticket-execute-full-from-issue":
		return executeTicketFullWorkflow(ctx, menuDisplay, "issue")
	case "ticket-execute-full-from-input":
		return executeTicketFullWorkflow(ctx, menuDisplay, "input")

	// Configuration Management
	case "config-init":
		return executeConfigInit(ctx, menuDisplay)
	case "config-sync":
		return executeConfigSync(ctx, menuDisplay)
	case "config-upgrade":
		return executeConfigUpgrade(ctx, menuDisplay)

	// Metrics Management
	case "metrics-status":
		return executeMetricsStatus(ctx, menuDisplay)
	case "metrics-commands":
		return executeMetricsCommands(ctx, menuDisplay)
	case "metrics-slow":
		return executeMetricsSlow(ctx, menuDisplay)
	case "metrics-projects":
		return executeMetricsProjects(ctx, menuDisplay)
	case "metrics-command":
		return executeMetricsCommand(ctx, menuDisplay)
	case "metrics-steps":
		return executeMetricsSteps(ctx, menuDisplay)

	// Legacy actions
	case "init-project":
		return executeInitProject(ctx, menuDisplay)

	default:
		menuDisplay.ShowWarning(fmt.Sprintf("Action '%s' not yet implemented", action))
		menuDisplay.ShowMessage("This action will be available in a future version.")
		return nil
	}
}

// executeInitProject handles comprehensive project initialization
func executeInitProject(ctx *navigation.ProjectContext, menuDisplay *navigation.MenuDisplay) error {
	confirmed, err := menuDisplay.Confirm("Initialize complete project structure in current directory?")
	if err != nil {
		return err
	}

	if !confirmed {
		menuDisplay.ShowMessage("Project initialization cancelled")
		return nil
	}

	menuDisplay.ShowMessage("🚀 Starting comprehensive project initialization...")

	// Step 1: Create all required directories
	if err := createProjectDirectories(ctx, menuDisplay); err != nil {
		return err
	}

	// Step 2: Import Claude commands and hooks
	menuDisplay.ShowMessage("📥 Importing Claude commands and hooks...")
	if err := executeClaudeImport(ctx, menuDisplay); err != nil {
		menuDisplay.ShowWarning(fmt.Sprintf("Failed to import Claude commands: %v", err))
		menuDisplay.ShowMessage("Continuing with manual setup...")
	}

	// Step 3: Copy template files if they don't exist
	if err := copyTemplateFiles(ctx, menuDisplay); err != nil {
		return err
	}

	// Step 4: Initialize Git with main/develop branches
	if err := initializeGitBranches(ctx, menuDisplay); err != nil {
		menuDisplay.ShowWarning(fmt.Sprintf("Git initialization failed: %v", err))
		menuDisplay.ShowMessage("You can initialize Git manually later")
	}

	menuDisplay.ShowSuccess("✅ Complete project structure initialized!")
	menuDisplay.ShowMessage("Your project is ready for development with:")
	menuDisplay.ShowMessage("  • Directory structure")
	menuDisplay.ShowMessage("  • Claude Code integration")
	menuDisplay.ShowMessage("  • Template files")
	menuDisplay.ShowMessage("  • Git repository with main/develop branches")
	return nil
}

// createProjectDirectories creates all required project directories
func createProjectDirectories(ctx *navigation.ProjectContext, menuDisplay *navigation.MenuDisplay) error {
	menuDisplay.ShowMessage("📁 Creating project directories...")

	dirs := []string{
		"docs/1-project",
		"docs/2-current-epic",
		"docs/3-current-task",
		"docs/archive",
		".claude-wm",
		".claude",
	}

	for _, dir := range dirs {
		fullPath := filepath.Join(ctx.ProjectPath, dir)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			if err := os.MkdirAll(fullPath, 0755); err != nil {
				return errors.NewCLIError("Failed to create project directory", 1).
					WithDetails(err.Error()).
					WithContext("directory", fullPath)
			}
			menuDisplay.ShowMessage(fmt.Sprintf("  ✓ Created %s", dir))
		} else {
			menuDisplay.ShowMessage(fmt.Sprintf("  ◦ %s already exists", dir))
		}
	}

	return nil
}

// copyTemplateFiles copies template files from runtime config to project root
func copyTemplateFiles(ctx *navigation.ProjectContext, menuDisplay *navigation.MenuDisplay) error {
	menuDisplay.ShowMessage("📄 Copying template files...")

	// Ensure config is initialized
	if err := config.EnsureConfigInitialized(ctx.ProjectPath); err != nil {
		return fmt.Errorf("failed to initialize config: %w", err)
	}

	templateDir, err := config.GetRuntimeConfigPath("commands/templates")
	if err != nil {
		return fmt.Errorf("failed to get template path: %w", err)
	}

	// Check if template directory exists
	if _, err := os.Stat(templateDir); os.IsNotExist(err) {
		menuDisplay.ShowWarning("Template directory not found. Skipping template file copy.")
		return nil
	}

	templateFiles := []string{"README.md", "METRICS.md", "CLAUDE.md"}

	for _, fileName := range templateFiles {
		sourcePath := filepath.Join(templateDir, fileName)
		destPath := filepath.Join(ctx.ProjectPath, fileName)

		// Check if source file exists
		if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
			menuDisplay.ShowMessage(fmt.Sprintf("  ◦ %s template not found, skipping", fileName))
			continue
		}

		// Check if destination file already exists
		if _, err := os.Stat(destPath); err == nil {
			menuDisplay.ShowMessage(fmt.Sprintf("  ◦ %s already exists, skipping", fileName))
			continue
		}

		// Copy the file
		if err := copyFile(sourcePath, destPath); err != nil {
			menuDisplay.ShowWarning(fmt.Sprintf("Failed to copy %s: %v", fileName, err))
			continue
		}

		menuDisplay.ShowMessage(fmt.Sprintf("  ✓ Copied %s", fileName))
	}

	return nil
}

// copyFile copies a file from source to destination
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

// initializeGitBranches initializes Git repository with main and develop branches
func initializeGitBranches(ctx *navigation.ProjectContext, menuDisplay *navigation.MenuDisplay) error {
	menuDisplay.ShowMessage("🌿 Initializing Git repository...")

	// Check if .git directory already exists
	gitDir := filepath.Join(ctx.ProjectPath, ".git")
	if _, err := os.Stat(gitDir); err == nil {
		menuDisplay.ShowMessage("  ◦ Git repository already exists")
		return ensureBranches(ctx, menuDisplay)
	}

	// Initialize git repository
	cmd := exec.Command("git", "init")
	cmd.Dir = ctx.ProjectPath
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to initialize git repository: %w", err)
	}
	menuDisplay.ShowMessage("  ✓ Initialized Git repository")

	// Create initial commit
	cmd = exec.Command("git", "add", ".")
	cmd.Dir = ctx.ProjectPath
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to add files to git: %w", err)
	}

	cmd = exec.Command("git", "commit", "-m", "Initial project setup with Claude WM CLI")
	cmd.Dir = ctx.ProjectPath
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create initial commit: %w", err)
	}
	menuDisplay.ShowMessage("  ✓ Created initial commit")

	return ensureBranches(ctx, menuDisplay)
}

// ensureBranches ensures main and develop branches exist
func ensureBranches(ctx *navigation.ProjectContext, menuDisplay *navigation.MenuDisplay) error {
	// Check current branch
	cmd := exec.Command("git", "branch", "--show-current")
	cmd.Dir = ctx.ProjectPath
	currentBranch, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get current branch: %w", err)
	}

	currentBranchName := strings.TrimSpace(string(currentBranch))

	// Rename current branch to main if it's not already main
	if currentBranchName != "main" && currentBranchName != "" {
		cmd = exec.Command("git", "branch", "-M", "main")
		cmd.Dir = ctx.ProjectPath
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to rename branch to main: %w", err)
		}
		menuDisplay.ShowMessage("  ✓ Renamed default branch to main")
	}

	// Create develop branch if it doesn't exist
	cmd = exec.Command("git", "branch")
	cmd.Dir = ctx.ProjectPath
	branches, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to list branches: %w", err)
	}

	if !strings.Contains(string(branches), "develop") {
		cmd = exec.Command("git", "checkout", "-b", "develop")
		cmd.Dir = ctx.ProjectPath
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to create develop branch: %w", err)
		}
		menuDisplay.ShowMessage("  ✓ Created develop branch")

		// Switch back to main
		cmd = exec.Command("git", "checkout", "main")
		cmd.Dir = ctx.ProjectPath
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to switch back to main: %w", err)
		}
		menuDisplay.ShowMessage("  ✓ Switched back to main branch")
	} else {
		menuDisplay.ShowMessage("  ◦ develop branch already exists")
	}

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

// executeStoryCommand executes a story subcommand
func executeStoryCommand(args []string, menuDisplay *navigation.MenuDisplay) error {
	cmdArgs := append([]string{"story"}, args...)

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
		menuDisplay.ShowError(fmt.Sprintf("Failed to execute story %s: %v", args[0], err))
		return err
	}

	menuDisplay.ShowSuccess(fmt.Sprintf("✅ Story %s completed successfully", args[0]))
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

// cleanCurrentTaskDirectory removes all files from docs/3-current-task/
func cleanCurrentTaskDirectory(projectPath string, menuDisplay *navigation.MenuDisplay) error {
	currentTaskDir := filepath.Join(projectPath, "docs/3-current-task")

	// Check if directory exists
	if _, err := os.Stat(currentTaskDir); os.IsNotExist(err) {
		// Create directory if it doesn't exist
		if err := os.MkdirAll(currentTaskDir, 0755); err != nil {
			return fmt.Errorf("failed to create current task directory: %w", err)
		}
		menuDisplay.ShowMessage("  ✓ Created docs/3-current-task/ directory")
		return nil
	}

	// Read directory contents
	files, err := os.ReadDir(currentTaskDir)
	if err != nil {
		return fmt.Errorf("failed to read current task directory: %w", err)
	}

	// Remove all files and subdirectories
	for _, file := range files {
		filePath := filepath.Join(currentTaskDir, file.Name())
		if err := os.RemoveAll(filePath); err != nil {
			return fmt.Errorf("failed to remove %s: %w", file.Name(), err)
		}
		menuDisplay.ShowMessage(fmt.Sprintf("  🗑️ Removed %s", file.Name()))
	}

	if len(files) == 0 {
		menuDisplay.ShowMessage("  ◦ docs/3-current-task/ was already empty")
	} else {
		menuDisplay.ShowMessage(fmt.Sprintf("  ✓ Cleaned %d items from docs/3-current-task/", len(files)))
	}

	return nil
}

// copyIterationsTemplate copies ITERATIONS.md from template to current task directory
func copyIterationsTemplate(projectPath string, menuDisplay *navigation.MenuDisplay) error {
	// Ensure config is initialized
	if err := config.EnsureConfigInitialized(projectPath); err != nil {
		return fmt.Errorf("failed to initialize config: %w", err)
	}

	manager := config.NewManager(projectPath)
	templatePath := manager.GetRuntimePath("commands/templates/ITERATIONS.md")
	destPath := filepath.Join(projectPath, "docs/3-current-task/ITERATIONS.md")

	// Check if template exists
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		menuDisplay.ShowWarning("⚠️  ITERATIONS.md template not found, skipping template copy")
		menuDisplay.ShowMessage("💡 You may need to run Claude import/install first")
		return nil
	}

	// Copy template file
	if err := copyFile(templatePath, destPath); err != nil {
		return fmt.Errorf("failed to copy ITERATIONS.md template: %w", err)
	}

	menuDisplay.ShowMessage("  ✓ Copied ITERATIONS.md template to docs/3-current-task/")
	return nil
}

// preprocessPlanTicketCommand handles preprocessing for Plan-Ticket command
func preprocessPlanTicketCommand(projectPath string, menuDisplay *navigation.MenuDisplay) error {
	menuDisplay.ShowMessage("📋 Preparing task workspace...")

	// Step 1: Clean current task directory
	if err := cleanCurrentTaskDirectory(projectPath, menuDisplay); err != nil {
		return fmt.Errorf("failed to clean current task directory: %w", err)
	}

	// Step 2: Copy ITERATIONS.md template
	if err := copyIterationsTemplate(projectPath, menuDisplay); err != nil {
		return fmt.Errorf("failed to copy ITERATIONS template: %w", err)
	}

	menuDisplay.ShowSuccess("✅ Task workspace prepared successfully")
	return nil
}

// executeClaudeCommandInteractive executes a Claude slash command from interactive menu
func executeClaudeCommandInteractive(command string, menuDisplay *navigation.MenuDisplay) error {
	// Start performance monitoring
	timer := metrics.InstrumentClaudeCommand(command)
	defer timer.Stop()

	menuDisplay.ShowMessage(fmt.Sprintf("🚀 Executing Claude command: %s", command))

	// Step 1: Preprocessing
	preprocessStep := timer.ProfileStep(metrics.StepPreprocessing)
	
	// Special preprocessing for Plan-Ticket command
	if command == "/4-task:2-execute:1-Plan-Ticket" {
		// Get current working directory for preprocessing
		workDir, err := os.Getwd()
		if err != nil {
			preprocessStep.StopWithError(err)
			menuDisplay.ShowError(fmt.Sprintf("Failed to get current directory: %v", err))
			timer.SetExitCode(1)
			return err
		}

		// Execute preprocessing
		if err := preprocessPlanTicketCommand(workDir, menuDisplay); err != nil {
			preprocessStep.StopWithError(err)
			menuDisplay.ShowError(fmt.Sprintf("Failed to prepare task workspace: %v", err))
			timer.SetExitCode(1)
			return err
		}
	}
	preprocessStep.Stop()

	// Step 2: Claude preparation and validation
	claudeValidationStep := timer.ProfileStep("claude_validation")
	claudeExecutor := executor.NewClaudeExecutor()

	// Validate Claude is available
	if err := claudeExecutor.ValidateClaudeAvailable(); err != nil {
		claudeValidationStep.StopWithError(err)
		menuDisplay.ShowError(fmt.Sprintf("Claude CLI not available: %v", err))
		menuDisplay.ShowMessage("💡 Please install Claude CLI to use this functionality")
		timer.SetExitCode(1)
		return err
	}
	claudeValidationStep.Stop()

	// Step 3: Claude execution (this is the potentially slow part)
	claudeExecutionStep := timer.ProfileClaudeExecution(command)
	description := fmt.Sprintf("Interactive menu command: %s", command)
	
	// Add additional context for Start Story command
	if command == "/3-story:1-manage:1-Start-Story" {
		claudeExecutionStep.SetMetadata("is_start_story", true)
		claudeExecutionStep.SetMetadata("workflow_phase", "story_start")
		claudeExecutionStep.SetMetadata("expected_operations", []string{
			"story_analysis", "priority_selection", "context_preparation", "claude_interaction",
		})
	}
	
	if err := claudeExecutor.ExecuteSlashCommand(command, description); err != nil {
		claudeExecutionStep.StopWithError(err)
		menuDisplay.ShowError(fmt.Sprintf("Failed to execute Claude command: %v", err))
		timer.SetExitCode(1)
		return err
	}
	claudeExecutionStep.Stop()

	// Step 4: Post-processing
	postProcessStep := timer.ProfileStep(metrics.StepPostprocessing)
	// Add any post-processing logic here if needed
	postProcessStep.Stop()

	timer.SetExitCode(0)
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

// executeClaudeImport handles importing commands and hooks from pezzos/.claude repository
// Legacy functions - these are now replaced by the config system

// executeClaudeImport - deprecated, use executeConfigInit instead
func executeClaudeImport(ctx *navigation.ProjectContext, menuDisplay *navigation.MenuDisplay) error {
	menuDisplay.ShowMessage("⚠️  This command is deprecated")
	menuDisplay.ShowMessage("🔄 Redirecting to new configuration system...")
	return executeConfigInit(ctx, menuDisplay)
}

// executeClaudeInit - deprecated, use executeConfigInit instead  
func executeClaudeInit(ctx *navigation.ProjectContext, menuDisplay *navigation.MenuDisplay) error {
	menuDisplay.ShowMessage("⚠️  This command is deprecated")
	menuDisplay.ShowMessage("🔄 Redirecting to new configuration system...")
	return executeConfigInit(ctx, menuDisplay)
}

// executeClaudeInstall - deprecated, use executeConfigSync instead
func executeClaudeInstall(ctx *navigation.ProjectContext, menuDisplay *navigation.MenuDisplay) error {
	menuDisplay.ShowMessage("⚠️  This command is deprecated")
	menuDisplay.ShowMessage("🔄 Redirecting to new configuration system...")
	return executeConfigSync(ctx, menuDisplay)
}

// Task execution functions with preprocessing integration

// executeTaskFromStory handles task creation from story with preprocessing
func executeTaskFromStory(ctx *navigation.ProjectContext, menuDisplay *navigation.MenuDisplay) error {
	// Step 1: Execute preprocessing
	if err := preprocessing.PreprocessFromStory(ctx.ProjectPath, menuDisplay); err != nil {
		menuDisplay.ShowError(fmt.Sprintf("Preprocessing failed: %v", err))
		return err
	}

	// Step 2: Execute Claude command for intelligent content generation
	return executeClaudeCommandInteractive("/4-task:1-start:1-From-story", menuDisplay)
}

// executeTaskFromIssue handles task creation from GitHub issue with preprocessing
func executeTaskFromIssue(ctx *navigation.ProjectContext, menuDisplay *navigation.MenuDisplay) error {
	// Step 1: Execute preprocessing
	if err := preprocessing.PreprocessFromIssue(ctx.ProjectPath, menuDisplay); err != nil {
		menuDisplay.ShowError(fmt.Sprintf("Preprocessing failed: %v", err))
		return err
	}

	// Step 2: Execute Claude command for intelligent content generation
	return executeClaudeCommandInteractive("/4-task:1-start:2-From-issue", menuDisplay)
}

// executeTaskFromInput handles task creation from user input with preprocessing
func executeTaskFromInput(ctx *navigation.ProjectContext, menuDisplay *navigation.MenuDisplay) error {
	// Get user input for task description
	fmt.Print("Enter task description: ")
	var description string
	fmt.Scanln(&description)

	if strings.TrimSpace(description) == "" {
		menuDisplay.ShowError("Task description cannot be empty")
		return fmt.Errorf("empty task description")
	}

	// Step 1: Execute preprocessing with user input
	if err := preprocessing.PreprocessFromInput(ctx.ProjectPath, description, menuDisplay); err != nil {
		menuDisplay.ShowError(fmt.Sprintf("Preprocessing failed: %v", err))
		return err
	}

	// Step 2: Execute Claude command for intelligent content generation
	return executeClaudeCommandInteractive("/4-task:1-start:3-From-input", menuDisplay)
}

// executeTaskPlan handles task planning with preprocessing
func executeTaskPlan(ctx *navigation.ProjectContext, menuDisplay *navigation.MenuDisplay) error {
	// Step 1: Execute preprocessing
	if err := preprocessing.PreprocessPlanTask(ctx.ProjectPath, menuDisplay); err != nil {
		menuDisplay.ShowError(fmt.Sprintf("Preprocessing failed: %v", err))
		return err
	}

	// Step 2: Execute Claude command for intelligent planning
	return executeClaudeCommandInteractive("/4-task:2-execute:1-Plan-Task", menuDisplay)
}

// executeTaskTestDesign handles test design with preprocessing
func executeTaskTestDesign(ctx *navigation.ProjectContext, menuDisplay *navigation.MenuDisplay) error {
	// Step 1: Execute preprocessing
	if err := preprocessing.PreprocessTestDesign(ctx.ProjectPath, menuDisplay); err != nil {
		menuDisplay.ShowError(fmt.Sprintf("Preprocessing failed: %v", err))
		return err
	}

	// Step 2: Execute Claude command for intelligent test design
	return executeClaudeCommandInteractive("/4-task:2-execute:2-Test-design", menuDisplay)
}

// executeTaskValidate handles task validation with preprocessing
func executeTaskValidate(ctx *navigation.ProjectContext, menuDisplay *navigation.MenuDisplay) error {
	// Step 1: Execute preprocessing
	if err := preprocessing.PreprocessValidateTask(ctx.ProjectPath, menuDisplay); err != nil {
		menuDisplay.ShowError(fmt.Sprintf("Preprocessing failed: %v", err))
		return err
	}

	// Step 2: Execute Claude command for intelligent validation
	return executeClaudeCommandInteractive("/4-task:2-execute:4-Validate-Task", menuDisplay)
}

// executeTaskReview handles task review with preprocessing
func executeTaskReview(ctx *navigation.ProjectContext, menuDisplay *navigation.MenuDisplay) error {
	// Step 1: Execute preprocessing
	if err := preprocessing.PreprocessReviewTask(ctx.ProjectPath, menuDisplay); err != nil {
		menuDisplay.ShowError(fmt.Sprintf("Preprocessing failed: %v", err))
		return err
	}

	// Step 2: Execute Claude command for intelligent review
	return executeClaudeCommandInteractive("/4-task:2-execute:5-Review-Task", menuDisplay)
}

// executeTaskArchive handles task archiving with preprocessing
func executeTaskArchive(ctx *navigation.ProjectContext, menuDisplay *navigation.MenuDisplay) error {
	// Step 1: Execute preprocessing
	if err := preprocessing.PreprocessArchiveTask(ctx.ProjectPath, menuDisplay); err != nil {
		menuDisplay.ShowError(fmt.Sprintf("Preprocessing failed: %v", err))
		return err
	}

	// Step 2: Execute Claude command for intelligent archiving
	return executeClaudeCommandInteractive("/4-task:3-complete:1-Archive-Task", menuDisplay)
}

// executeTaskStatus handles task status with preprocessing
func executeTaskStatus(ctx *navigation.ProjectContext, menuDisplay *navigation.MenuDisplay) error {
	// Step 1: Execute preprocessing and get status
	status, err := preprocessing.PreprocessStatusTask(ctx.ProjectPath, menuDisplay)
	if err != nil {
		menuDisplay.ShowError(fmt.Sprintf("Preprocessing failed: %v", err))
		return err
	}

	// Display status information
	menuDisplay.ShowMessage("📊 Task Status Report:")
	menuDisplay.ShowMessage(fmt.Sprintf("  %s", status.Message))
	menuDisplay.ShowMessage(fmt.Sprintf("  %s", status.Details))

	// Step 2: Execute Claude command for detailed status analysis
	return executeClaudeCommandInteractive("/4-task:3-complete:2-Status-Task", menuDisplay)
}

// executeTicketFullWorkflow executes the complete ticket workflow with iteration support
func executeTicketFullWorkflow(ctx *navigation.ProjectContext, menuDisplay *navigation.MenuDisplay, source string) error {
	menuDisplay.ShowMessage("🚀 Starting full ticket workflow with iteration support...")

	// Step 1: Initialize task based on source
	if err := initializeTaskFromSource(ctx, menuDisplay, source); err != nil {
		return err
	}

	// Main workflow loop with iteration support
	maxIterations := 3
	for iteration := 1; iteration <= maxIterations; iteration++ {
		menuDisplay.ShowMessage(fmt.Sprintf("🔄 Starting iteration %d/%d", iteration, maxIterations))

		// Step 2: Plan Task
		if err := executeTaskPlan(ctx, menuDisplay); err != nil {
			return fmt.Errorf("failed at planning step: %w", err)
		}

		// Step 3: Test Design
		if err := executeTaskTestDesign(ctx, menuDisplay); err != nil {
			return fmt.Errorf("failed at test design step: %w", err)
		}

		// Step 4: Implementation
		if err := executeClaudeCommandInteractive("/4-task:2-execute:3-Implement", menuDisplay); err != nil {
			return fmt.Errorf("failed at implementation step: %w", err)
		}

		// Step 5: Validation (with iteration check)
		validationResult, err := executeValidationWithIterationCheck(ctx, menuDisplay, iteration, maxIterations)
		if err != nil {
			return fmt.Errorf("failed at validation step: %w", err)
		}

		switch validationResult {
		case ValidationSuccess:
			menuDisplay.ShowSuccess("✅ Validation successful! Resetting iterations and proceeding to review...")
			
			// Reset iterations.json for review phase
			if err := resetIterationsAfterValidation(ctx.ProjectPath, menuDisplay); err != nil {
				menuDisplay.ShowWarning(fmt.Sprintf("Failed to reset iterations.json: %v", err))
			}
			
			// Enter review iteration loop (infinite until success or explicit failure)
			return executeReviewIterationLoop(ctx, menuDisplay)

		case ValidationFailedRetry:
			menuDisplay.ShowMessage(fmt.Sprintf("⚠️ Validation failed (iteration %d/%d). Retrying from planning step...", iteration, maxIterations))
			continue // Go to next iteration

		case ValidationFailedMaxReached:
			menuDisplay.ShowError(fmt.Sprintf("❌ Validation failed after %d iterations. Workflow stopped.", maxIterations))
			return fmt.Errorf("validation failed after maximum iterations (%d)", maxIterations)

		default:
			return fmt.Errorf("unknown validation result: %v", validationResult)
		}
	}

	return fmt.Errorf("workflow failed after %d iterations", maxIterations)
}

// ValidationResult represents the result of a validation step
type ValidationResult int

const (
	ValidationSuccess ValidationResult = iota
	ValidationFailedRetry
	ValidationFailedMaxReached
)

// initializeTaskFromSource initializes the task based on the source type
func initializeTaskFromSource(ctx *navigation.ProjectContext, menuDisplay *navigation.MenuDisplay, source string) error {
	switch source {
	case "story":
		return executeTaskFromStory(ctx, menuDisplay)
	case "issue":
		return executeTaskFromIssue(ctx, menuDisplay)
	case "input":
		return executeTaskFromInput(ctx, menuDisplay)
	case "":
		// No initialization needed - task already exists
		menuDisplay.ShowMessage("📋 Using existing task context...")
		return nil
	default:
		return fmt.Errorf("unknown task source: %s", source)
	}
}

// executeValidationWithIterationCheck executes validation and determines next action based on result
func executeValidationWithIterationCheck(ctx *navigation.ProjectContext, menuDisplay *navigation.MenuDisplay, currentIteration, maxIterations int) (ValidationResult, error) {
	// Execute preprocessing first
	if err := preprocessing.PreprocessValidateTask(ctx.ProjectPath, menuDisplay); err != nil {
		return ValidationFailedRetry, fmt.Errorf("preprocessing failed: %w", err)
	}

	// Check current iteration status from iterations.json
	iterationsPath := filepath.Join(ctx.ProjectPath, "docs/3-current-task/iterations.json")
	iterations, err := parseIterationsJSONFile(iterationsPath)
	if err != nil {
		menuDisplay.ShowWarning("⚠️ Could not read iterations.json, continuing with validation")
	}

	// Execute Claude validation command
	claudeExecutor := executor.NewClaudeExecutor()
	if err := claudeExecutor.ValidateClaudeAvailable(); err != nil {
		return ValidationFailedRetry, fmt.Errorf("Claude CLI not available: %w", err)
	}

	// Execute validation command and capture exit code
	description := fmt.Sprintf("Validation step (iteration %d/%d)", currentIteration, maxIterations)
	exitCode, err := claudeExecutor.ExecuteSlashCommandWithExitCode("/4-task:2-execute:4-Validate-Task", description)
	
	if err != nil {
		menuDisplay.ShowError(fmt.Sprintf("Failed to execute validation: %v", err))
		return ValidationFailedRetry, err
	}

	// Interpret Claude's exit code
	switch exitCode {
	case 0: // Success
		menuDisplay.ShowSuccess("✅ Validation passed!")
		return ValidationSuccess, nil

	case 1: // Needs iteration
		menuDisplay.ShowMessage("⚠️ Validation indicates iteration needed")
		
		// Check if we've reached max iterations
		if currentIteration >= maxIterations {
			// Update iterations.json to mark as blocked
			if iterations != nil {
				if err := updateIterationsAsBlocked(iterationsPath, iterations, "Maximum iterations reached"); err != nil {
					menuDisplay.ShowWarning(fmt.Sprintf("Failed to update iterations.json: %v", err))
				}
			}
			return ValidationFailedMaxReached, nil
		}
		
		// Update iterations.json for retry
		if iterations != nil {
			if err := updateIterationsForRetry(iterationsPath, iterations, currentIteration); err != nil {
				menuDisplay.ShowWarning(fmt.Sprintf("Failed to update iterations.json: %v", err))
			}
		}
		
		return ValidationFailedRetry, nil

	case 2: // Blocked
		menuDisplay.ShowError("❌ Validation indicates task is blocked")
		if iterations != nil {
			if err := updateIterationsAsBlocked(iterationsPath, iterations, "Validation blocked"); err != nil {
				menuDisplay.ShowWarning(fmt.Sprintf("Failed to update iterations.json: %v", err))
			}
		}
		return ValidationFailedMaxReached, fmt.Errorf("validation blocked")

	default:
		menuDisplay.ShowError(fmt.Sprintf("❌ Validation returned unexpected exit code: %d", exitCode))
		return ValidationFailedRetry, fmt.Errorf("unexpected exit code: %d", exitCode)
	}
}

// updateIterationsForRetry updates iterations.json for a retry scenario
func updateIterationsForRetry(iterationsPath string, iterations *preprocessing.IterationsData, currentIteration int) error {
	// Update current iteration
	iterations.TaskContext.CurrentIteration = currentIteration + 1
	
	// Add iteration record
	newIteration := preprocessing.Iteration{
		IterationNumber: currentIteration,
		Attempt: preprocessing.Attempt{
			StartedAt:      time.Now().Format(time.RFC3339),
			Approach:       "Validation-driven retry",
			Implementation: []string{"Validation failed", "Retrying from planning step"},
		},
		Result: preprocessing.Result{
			Success: false,
			Outcome: "❌ Failed",
			Details: "Validation did not pass, retrying from planning",
		},
		Learnings:   []string{"Validation failed", "Need to revisit planning and implementation"},
		CompletedAt: time.Now().Format(time.RFC3339),
	}
	
	iterations.Iterations = append(iterations.Iterations, newIteration)
	
	// Write back to file
	return writeJSONToFile(iterationsPath, iterations)
}

// updateIterationsAsBlocked updates iterations.json when max iterations reached or blocked
func updateIterationsAsBlocked(iterationsPath string, iterations *preprocessing.IterationsData, reason string) error {
	// Update final outcome
	iterations.FinalOutcome = preprocessing.FinalOutcome{
		Status:                "blocked",
		Solution:              "",
		TotalTimeHours:        0, // Would be calculated based on iterations
		Complexity:            "higher_than_estimated",
		OriginalEstimateHours: 0, // Would be from initial estimate
	}
	
	// Add recommendations
	iterations.Recommendations = append(iterations.Recommendations, 
		"Consider breaking down the task into smaller components",
		"Review approach and seek additional expertise",
		"Validate requirements and acceptance criteria",
	)
	
	// Write back to file
	return writeJSONToFile(iterationsPath, iterations)
}

// writeJSONToFile writes JSON data to a file
func writeJSONToFile(path string, data interface{}) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, jsonData, 0644)
}

// parseIterationsJSONFile parses iterations.json file locally
func parseIterationsJSONFile(path string) (*preprocessing.IterationsData, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var iterations preprocessing.IterationsData
	if err := json.Unmarshal(data, &iterations); err != nil {
		return nil, err
	}

	return &iterations, nil
}

// resetIterationsAfterValidation resets iterations.json by copying template after successful validation
func resetIterationsAfterValidation(projectPath string, menuDisplay *navigation.MenuDisplay) error {
	menuDisplay.ShowMessage("🔄 Resetting iterations.json for review phase...")
	
	// Ensure config is initialized
	if err := config.EnsureConfigInitialized(projectPath); err != nil {
		return fmt.Errorf("failed to initialize config: %w", err)
	}

	// Copy fresh template from runtime configuration
	manager := config.NewManager(projectPath)
	templatePath := manager.GetRuntimePath("commands/templates/iterations.json")
	destPath := filepath.Join(projectPath, "docs/3-current-task/iterations.json")
	
	if err := copyFile(templatePath, destPath); err != nil {
		return fmt.Errorf("failed to copy iterations.json template: %w", err)
	}
	
	// Initialize with review phase context
	if err := initializeIterationsForReviewPhase(destPath, projectPath); err != nil {
		return fmt.Errorf("failed to initialize iterations for review phase: %w", err)
	}
	
	menuDisplay.ShowMessage("  ✓ Iterations reset for review phase")
	return nil
}

// initializeIterationsForReviewPhase initializes iterations.json for review phase
func initializeIterationsForReviewPhase(iterationsPath, projectPath string) error {
	// Initialize iterations.json with review phase context
	iterationsData := preprocessing.IterationsData{
		TaskContext: preprocessing.TaskContext{
			TaskID:           "TASK-REVIEW",
			Title:            "Review Phase",
			CurrentIteration: 1,
			MaxIterations:    999, // No limit for review as requested
			Status:           "in_progress",
			Branch:           getCurrentGitBranch(projectPath),
			StartedAt:        time.Now().Format(time.RFC3339),
		},
		Iterations:      []preprocessing.Iteration{},
		FinalOutcome:    preprocessing.FinalOutcome{},
		Recommendations: []string{},
	}

	return writeJSONToFile(iterationsPath, iterationsData)
}

// executeReviewIterationLoop handles the review phase with iteration support
func executeReviewIterationLoop(ctx *navigation.ProjectContext, menuDisplay *navigation.MenuDisplay) error {
	menuDisplay.ShowMessage("👀 Starting review phase with iteration support...")
	
	reviewIteration := 1
	
	for {
		menuDisplay.ShowMessage(fmt.Sprintf("🔄 Review iteration %d", reviewIteration))
		
		// Execute review with iteration check
		reviewResult, err := executeReviewWithIterationCheck(ctx, menuDisplay, reviewIteration)
		if err != nil {
			return fmt.Errorf("failed at review step: %w", err)
		}
		
		switch reviewResult {
		case ReviewSuccess:
			menuDisplay.ShowSuccess("✅ Review successful! Proceeding to archive...")
			
			// Step 7: Archive
			if err := executeTaskArchive(ctx, menuDisplay); err != nil {
				return fmt.Errorf("failed at archive step: %w", err)
			}
			
			menuDisplay.ShowSuccess("🎉 Full ticket workflow completed successfully!")
			return nil
			
		case ReviewFailedRetry:
			menuDisplay.ShowMessage(fmt.Sprintf("⚠️ Review failed (iteration %d). Starting new implementation cycle...", reviewIteration))
			
			// Execute full implementation cycle: Plan → Test → Implement → Validate
			if err := executeImplementationCycleForReview(ctx, menuDisplay, reviewIteration); err != nil {
				return fmt.Errorf("failed during implementation cycle for review: %w", err)
			}
			
			reviewIteration++
			continue // Go back to review
			
		case ReviewBlocked:
			menuDisplay.ShowError("❌ Review indicates task is blocked")
			return fmt.Errorf("review blocked - task cannot be completed as specified")
			
		default:
			return fmt.Errorf("unknown review result: %v", reviewResult)
		}
	}
}

// ReviewResult represents the result of a review step
type ReviewResult int

const (
	ReviewSuccess ReviewResult = iota
	ReviewFailedRetry
	ReviewBlocked
)

// executeReviewWithIterationCheck executes review and determines next action based on result
func executeReviewWithIterationCheck(ctx *navigation.ProjectContext, menuDisplay *navigation.MenuDisplay, reviewIteration int) (ReviewResult, error) {
	// Execute preprocessing first
	if err := preprocessing.PreprocessReviewTask(ctx.ProjectPath, menuDisplay); err != nil {
		return ReviewFailedRetry, fmt.Errorf("preprocessing failed: %w", err)
	}
	
	// Execute Claude review command
	claudeExecutor := executor.NewClaudeExecutor()
	if err := claudeExecutor.ValidateClaudeAvailable(); err != nil {
		return ReviewFailedRetry, fmt.Errorf("Claude CLI not available: %w", err)
	}
	
	// Execute review command and capture exit code
	description := fmt.Sprintf("Review step (iteration %d)", reviewIteration)
	exitCode, err := claudeExecutor.ExecuteSlashCommandWithExitCode("/4-task:2-execute:5-Review-Task", description)
	
	if err != nil {
		menuDisplay.ShowError(fmt.Sprintf("Failed to execute review: %v", err))
		return ReviewFailedRetry, err
	}
	
	// Interpret Claude's exit code
	switch exitCode {
	case 0: // Success
		menuDisplay.ShowSuccess("✅ Review passed!")
		return ReviewSuccess, nil
		
	case 1: // Needs iteration
		menuDisplay.ShowMessage("⚠️ Review indicates iteration needed")
		
		// Update iterations.json for review retry with specific feedback
		iterationsPath := filepath.Join(ctx.ProjectPath, "docs/3-current-task/iterations.json")
		if err := updateIterationsForReviewRetry(iterationsPath, reviewIteration); err != nil {
			menuDisplay.ShowWarning(fmt.Sprintf("Failed to update iterations.json: %v", err))
		}
		
		return ReviewFailedRetry, nil
		
	case 2: // Blocked
		menuDisplay.ShowError("❌ Review indicates task is blocked")
		
		// Update iterations.json as blocked
		iterationsPath := filepath.Join(ctx.ProjectPath, "docs/3-current-task/iterations.json")
		iterations, err := parseIterationsJSONFile(iterationsPath)
		if err == nil {
			if err := updateIterationsAsBlocked(iterationsPath, iterations, "Review blocked"); err != nil {
				menuDisplay.ShowWarning(fmt.Sprintf("Failed to update iterations.json: %v", err))
			}
		}
		
		return ReviewBlocked, fmt.Errorf("review blocked")
		
	default:
		menuDisplay.ShowError(fmt.Sprintf("❌ Review returned unexpected exit code: %d", exitCode))
		return ReviewFailedRetry, fmt.Errorf("unexpected exit code: %d", exitCode)
	}
}

// executeImplementationCycleForReview executes the full implementation cycle when review fails
func executeImplementationCycleForReview(ctx *navigation.ProjectContext, menuDisplay *navigation.MenuDisplay, reviewIteration int) error {
	menuDisplay.ShowMessage(fmt.Sprintf("🔄 Starting implementation cycle for review iteration %d", reviewIteration))
	
	// Step 2: Plan Task (with review feedback from iterations.json)
	if err := executeTaskPlan(ctx, menuDisplay); err != nil {
		return fmt.Errorf("failed at planning step: %w", err)
	}
	
	// Step 3: Test Design
	if err := executeTaskTestDesign(ctx, menuDisplay); err != nil {
		return fmt.Errorf("failed at test design step: %w", err)
	}
	
	// Step 4: Implementation
	if err := executeClaudeCommandInteractive("/4-task:2-execute:3-Implement", menuDisplay); err != nil {
		return fmt.Errorf("failed at implementation step: %w", err)
	}
	
	// Step 5: Validation (simple execution without iteration - we assume it will pass)
	menuDisplay.ShowMessage("🔍 Quick validation before returning to review...")
	if err := preprocessing.PreprocessValidateTask(ctx.ProjectPath, menuDisplay); err != nil {
		menuDisplay.ShowWarning(fmt.Sprintf("Validation preprocessing failed: %v", err))
	}
	
	if err := executeClaudeCommandInteractive("/4-task:2-execute:4-Validate-Task", menuDisplay); err != nil {
		menuDisplay.ShowWarning(fmt.Sprintf("Validation failed: %v", err))
		// Continue anyway as review iteration will catch remaining issues
	}
	
	menuDisplay.ShowMessage("  ✓ Implementation cycle completed, returning to review")
	return nil
}

// updateIterationsForReviewRetry updates iterations.json for a review retry scenario
func updateIterationsForReviewRetry(iterationsPath string, reviewIteration int) error {
	iterations, err := parseIterationsJSONFile(iterationsPath)
	if err != nil {
		return err
	}
	
	// Update current iteration
	iterations.TaskContext.CurrentIteration = reviewIteration + 1
	
	// Add iteration record with review-specific context
	newIteration := preprocessing.Iteration{
		IterationNumber: reviewIteration,
		Attempt: preprocessing.Attempt{
			StartedAt:      time.Now().Format(time.RFC3339),
			Approach:       "Review-driven iteration",
			Implementation: []string{"Review identified issues", "Restarting from planning with review feedback"},
		},
		Result: preprocessing.Result{
			Success: false,
			Outcome: "❌ Failed",
			Details: "Review did not pass - implementation needs adjustments based on review feedback",
		},
		Learnings:   []string{"Review feedback requires implementation changes", "Need to revisit planning based on review insights"},
		CompletedAt: time.Now().Format(time.RFC3339),
	}
	
	iterations.Iterations = append(iterations.Iterations, newIteration)
	
	// Write back to file
	return writeJSONToFile(iterationsPath, iterations)
}

// getCurrentGitBranch gets the current git branch (helper function)
func getCurrentGitBranch(projectPath string) string {
	cmd := exec.Command("git", "branch", "--show-current")
	cmd.Dir = projectPath
	output, err := cmd.Output()
	if err != nil {
		return "main"
	}
	return strings.TrimSpace(string(output))
}

// executeConfigInit handles initializing the configuration workspace
func executeConfigInit(ctx *navigation.ProjectContext, menuDisplay *navigation.MenuDisplay) error {
	menuDisplay.ShowMessage("🚀 Initializing configuration workspace...")

	manager := config.NewManager(ctx.ProjectPath)

	// Initialize directory structure
	if err := manager.Initialize(); err != nil {
		menuDisplay.ShowError(fmt.Sprintf("Failed to initialize workspace: %v", err))
		return err
	}

	// Install default templates
	if err := manager.InstallSystemTemplates(); err != nil {
		menuDisplay.ShowError(fmt.Sprintf("Failed to install system templates: %v", err))
		return err
	}

	// Migrate from legacy structure if it exists
	legacyPath := filepath.Join(ctx.ProjectPath, ".claude-wm", ".claude")
	if err := manager.MigrateFromLegacy(legacyPath); err != nil {
		menuDisplay.ShowError(fmt.Sprintf("Migration failed: %v", err))
		return err
	}

	// Generate initial runtime configuration
	if err := manager.Sync(); err != nil {
		menuDisplay.ShowError(fmt.Sprintf("Failed to generate runtime configuration: %v", err))
		return err
	}

	menuDisplay.ShowSuccess("✅ Configuration workspace initialized successfully!")
	menuDisplay.ShowMessage("💡 Use 'claude-wm config show' to view your configuration")
	return nil
}

// executeConfigSync handles syncing the configuration
func executeConfigSync(ctx *navigation.ProjectContext, menuDisplay *navigation.MenuDisplay) error {
	menuDisplay.ShowMessage("🔄 Syncing configuration...")

	manager := config.NewManager(ctx.ProjectPath)
	if err := manager.Sync(); err != nil {
		menuDisplay.ShowError(fmt.Sprintf("Sync failed: %v", err))
		return err
	}

	menuDisplay.ShowSuccess("✅ Configuration synced successfully!")
	return nil
}

// executeConfigUpgrade handles upgrading system templates
func executeConfigUpgrade(ctx *navigation.ProjectContext, menuDisplay *navigation.MenuDisplay) error {
	menuDisplay.ShowMessage("⬆️  Upgrading system templates...")

	manager := config.NewManager(ctx.ProjectPath)

	// Reinstall system templates
	if err := manager.InstallSystemTemplates(); err != nil {
		menuDisplay.ShowError(fmt.Sprintf("Failed to upgrade system templates: %v", err))
		return err
	}

	// Regenerate runtime configuration
	if err := manager.Sync(); err != nil {
		menuDisplay.ShowError(fmt.Sprintf("Failed to sync after upgrade: %v", err))
		return err
	}

	menuDisplay.ShowSuccess("✅ System templates upgraded successfully!")
	menuDisplay.ShowMessage("💡 Your user customizations have been preserved")
	return nil
}

// Metrics execution functions

// executeMetricsStatus shows metrics collection status
func executeMetricsStatus(ctx *navigation.ProjectContext, menuDisplay *navigation.MenuDisplay) error {
	cmdArgs := []string{"metrics", "status"}
	return executeMetricsSubcommand(cmdArgs, menuDisplay)
}

// executeMetricsCommands shows all commands with statistics
func executeMetricsCommands(ctx *navigation.ProjectContext, menuDisplay *navigation.MenuDisplay) error {
	cmdArgs := []string{"metrics", "commands"}
	return executeMetricsSubcommand(cmdArgs, menuDisplay)
}

// executeMetricsSlow shows slowest commands
func executeMetricsSlow(ctx *navigation.ProjectContext, menuDisplay *navigation.MenuDisplay) error {
	cmdArgs := []string{"metrics", "slow"}
	return executeMetricsSubcommand(cmdArgs, menuDisplay)
}

// executeMetricsProjects shows project comparison
func executeMetricsProjects(ctx *navigation.ProjectContext, menuDisplay *navigation.MenuDisplay) error {
	cmdArgs := []string{"metrics", "projects"}
	return executeMetricsSubcommand(cmdArgs, menuDisplay)
}

// executeMetricsCommand shows detailed command analysis with user input
func executeMetricsCommand(ctx *navigation.ProjectContext, menuDisplay *navigation.MenuDisplay) error {
	// Get command name from user
	commandName, err := menuDisplay.PromptString("Enter command name to analyze (e.g., 'interactive', 'version')")
	if err != nil {
		return err
	}
	
	if strings.TrimSpace(commandName) == "" {
		menuDisplay.ShowError("Command name cannot be empty")
		return fmt.Errorf("empty command name")
	}
	
	cmdArgs := []string{"metrics", "command", strings.TrimSpace(commandName)}
	return executeMetricsSubcommand(cmdArgs, menuDisplay)
}

// executeMetricsSteps shows step-by-step analysis with user input
func executeMetricsSteps(ctx *navigation.ProjectContext, menuDisplay *navigation.MenuDisplay) error {
	// Get command name from user
	commandName, err := menuDisplay.PromptString("Enter command name to analyze steps (e.g., 'interactive')")
	if err != nil {
		return err
	}
	
	if strings.TrimSpace(commandName) == "" {
		menuDisplay.ShowError("Command name cannot be empty")
		return fmt.Errorf("empty command name")
	}
	
	cmdArgs := []string{"metrics", "steps", strings.TrimSpace(commandName)}
	return executeMetricsSubcommand(cmdArgs, menuDisplay)
}

// executeMetricsSubcommand executes a metrics subcommand
func executeMetricsSubcommand(args []string, menuDisplay *navigation.MenuDisplay) error {
	execPath, err := os.Executable()
	if err != nil {
		menuDisplay.ShowError(fmt.Sprintf("Failed to get executable path: %v", err))
		return err
	}

	buildPath := filepath.Join(filepath.Dir(filepath.Dir(execPath)), "build", "claude-wm-cli")
	if _, err := os.Stat(buildPath); err == nil {
		execPath = buildPath
	}

	cmd := exec.Command(execPath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		menuDisplay.ShowError(fmt.Sprintf("Failed to execute metrics %s: %v", args[1], err))
		return err
	}

	menuDisplay.ShowSuccess(fmt.Sprintf("✅ Metrics %s completed successfully", args[1]))
	return nil
}
