package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"claude-wm-cli/internal/errors"
	"claude-wm-cli/internal/navigation"
	"claude-wm-cli/internal/workflow"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// navigateCmd represents the navigate command
var navigateCmd = &cobra.Command{
	Use:   "navigate",
	Short: "Interactive navigation through project workflow",
	Long: `Navigate provides an interactive menu system to guide you through
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
  claude-wm-cli navigate              # Start interactive navigation
  claude-wm-cli navigate --status     # Show status and exit
  claude-wm-cli navigate --suggest    # Show suggestions and exit`,
	Aliases: []string{"nav", "menu"},
	RunE:    runNavigate,
}

// Navigation command flags
var (
	showStatusOnly    bool
	showSuggestOnly   bool
	showQuickStatus   bool
	noInteractive     bool
	displayWidth      int
	maxSuggestions    int
)

func init() {
	rootCmd.AddCommand(navigateCmd)

	// Add flags for navigation command
	navigateCmd.Flags().BoolVar(&showStatusOnly, "status", false, "show project status and exit")
	navigateCmd.Flags().BoolVar(&showSuggestOnly, "suggest", false, "show suggestions and exit")
	navigateCmd.Flags().BoolVar(&showQuickStatus, "quick", false, "show quick one-line status")
	navigateCmd.Flags().BoolVar(&noInteractive, "no-interactive", false, "disable interactive mode")
	navigateCmd.Flags().IntVar(&displayWidth, "width", 80, "display width for formatting")
	navigateCmd.Flags().IntVar(&maxSuggestions, "max-suggestions", 5, "maximum number of suggestions to show")

	// Bind flags to viper
	viper.BindPFlag("navigate.status", navigateCmd.Flags().Lookup("status"))
	viper.BindPFlag("navigate.suggest", navigateCmd.Flags().Lookup("suggest"))
	viper.BindPFlag("navigate.quick", navigateCmd.Flags().Lookup("quick"))
	viper.BindPFlag("navigate.no-interactive", navigateCmd.Flags().Lookup("no-interactive"))
	viper.BindPFlag("navigate.width", navigateCmd.Flags().Lookup("width"))
	viper.BindPFlag("navigate.max-suggestions", navigateCmd.Flags().Lookup("max-suggestions"))
}

// runNavigate executes the navigate command
func runNavigate(cmd *cobra.Command, args []string) error {
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

// createMainMenu builds the main navigation menu based on context and suggestions
func createMainMenu(ctx *navigation.ProjectContext, suggestions []*navigation.Suggestion) *navigation.Menu {
	builder := navigation.NewMenuBuilder("🧭 Project Navigation")

	// Add top suggestions as menu options
	addedSuggestions := 0
	for _, suggestion := range suggestions {
		if addedSuggestions >= 3 { // Limit to top 3 suggestions
			break
		}

		priorityIcon := getPriorityIcon(suggestion.Priority)
		label := fmt.Sprintf("%s%s", priorityIcon, suggestion.Action.Name)
		description := truncateString(suggestion.Reasoning, 60)

		builder.AddOption(
			suggestion.Action.ID,
			label,
			description,
			suggestion.Action.ID,
		)
		addedSuggestions++
	}

	// Add separator if we have suggestions
	if addedSuggestions > 0 {
		builder.AddSeparator()
	}

	// Add standard navigation options
	builder.AddOption("status", "📊 Show Project Status", "Display detailed project state and progress", "status")
	builder.AddOption("suggestions", "💡 View All Suggestions", "Show all available action suggestions", "suggestions")
	builder.AddOption("refresh", "🔄 Refresh Context", "Re-scan project state and update suggestions", "refresh")

	return builder.Build()
}

// executeAction handles the execution of selected actions
func executeAction(action string, ctx *navigation.ProjectContext, menuDisplay *navigation.MenuDisplay) error {
	switch action {
	case "init-project":
		return executeInitProject(ctx, menuDisplay)
	case "create-epic":
		return executeCreateEpic(ctx, menuDisplay)
	case "start-epic":
		return executeStartEpic(ctx, menuDisplay)
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

// executeCreateEpic handles epic creation (placeholder)
func executeCreateEpic(ctx *navigation.ProjectContext, menuDisplay *navigation.MenuDisplay) error {
	epicTitle, err := menuDisplay.PromptString("Enter epic title")
	if err != nil {
		return err
	}

	if epicTitle == "" {
		menuDisplay.ShowWarning("Epic title cannot be empty")
		return nil
	}

	menuDisplay.ShowMessage(fmt.Sprintf("Would create epic: %s", epicTitle))
	menuDisplay.ShowWarning("Epic creation not yet fully implemented")
	return nil
}

// executeStartEpic handles epic selection and start (placeholder)
func executeStartEpic(ctx *navigation.ProjectContext, menuDisplay *navigation.MenuDisplay) error {
	menuDisplay.ShowMessage("Would show available epics for selection")
	menuDisplay.ShowWarning("Epic selection not yet fully implemented")
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