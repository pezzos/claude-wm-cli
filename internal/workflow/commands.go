package workflow

import (
	"fmt"
	"sort"
)

// ContextualCommand represents a command that's appropriate for the current workflow state
type ContextualCommand struct {
	Action        *WorkflowAction `json:"action"`
	Priority      Priority        `json:"priority"`
	Reasoning     string          `json:"reasoning"`
	Prerequisites []string        `json:"prerequisites,omitempty"`
	NextActions   []string        `json:"next_actions,omitempty"`
	Warnings      []string        `json:"warnings,omitempty"`
}

// CommandGenerator generates contextual commands based on workflow analysis
type CommandGenerator struct {
	actionRegistry *ActionRegistry
	analyzer       *WorkflowAnalyzer
}

// NewCommandGenerator creates a new command generator
func NewCommandGenerator(rootPath string) *CommandGenerator {
	return &CommandGenerator{
		actionRegistry: NewActionRegistry(),
		analyzer:       NewWorkflowAnalyzer(rootPath),
	}
}

// GenerateContextualCommands generates appropriate commands for the current workflow state
func (cg *CommandGenerator) GenerateContextualCommands() ([]*ContextualCommand, error) {
	// First, analyze the current workflow state
	analysis, err := cg.analyzer.AnalyzeWorkflowPosition()
	if err != nil {
		return nil, fmt.Errorf("failed to analyze workflow: %w", err)
	}

	var commands []*ContextualCommand

	// Generate primary commands based on current position
	primaryCommands := cg.generatePrimaryCommands(analysis)
	commands = append(commands, primaryCommands...)

	// Generate secondary commands based on capabilities
	secondaryCommands := cg.generateSecondaryCommands(analysis)
	commands = append(commands, secondaryCommands...)

	// Generate utility commands (always available)
	utilityCommands := cg.generateUtilityCommands(analysis)
	commands = append(commands, utilityCommands...)

	// Filter commands based on prerequisites and blockers
	filteredCommands := cg.filterCommands(commands, analysis)

	// Sort commands by priority and relevance
	cg.sortCommands(filteredCommands)

	return filteredCommands, nil
}

// generatePrimaryCommands creates the most relevant commands for the current workflow position
func (cg *CommandGenerator) generatePrimaryCommands(analysis *WorkflowAnalysis) []*ContextualCommand {
	var commands []*ContextualCommand

	switch analysis.Position {
	case PositionNotInitialized:
		commands = append(commands, cg.createCommand("init-project", PriorityP0,
			"Project structure needs to be initialized",
			[]string{}, []string{"create-epic"}))

	case PositionProjectLevel:
		if analysis.CompletionMetrics.TotalEpics == 0 {
			commands = append(commands, cg.createCommand("create-epic", PriorityP0,
				"No epics defined - create your first epic to organize work",
				[]string{}, []string{"start-epic"}))
		} else {
			commands = append(commands, cg.createCommand("start-epic", PriorityP0,
				"Select an existing epic to start working",
				[]string{}, []string{"create-story"}))
		}

	case PositionEpicLevel:
		if analysis.CompletionMetrics.TotalStories == 0 {
			commands = append(commands, cg.createCommand("create-story", PriorityP0,
				"Epic selected but no stories defined - break down into user stories",
				[]string{}, []string{"create-task"}))
		} else {
			commands = append(commands, cg.createCommand("continue-epic", PriorityP1,
				"Continue working on the current epic",
				[]string{}, []string{"create-story", "continue-story"}))
		}

	case PositionStoryLevel:
		if analysis.CompletionMetrics.TotalTasks == 0 {
			commands = append(commands, cg.createCommand("create-task", PriorityP0,
				"Story selected but no tasks defined - create implementation tasks",
				[]string{}, []string{"continue-task"}))
		} else {
			commands = append(commands, cg.createCommand("continue-story", PriorityP1,
				"Continue working on the current story",
				[]string{}, []string{"create-task", "continue-task"}))
		}

	case PositionTaskLevel:
		// Determine the most appropriate task action
		inProgressTasks := 0
		todoTasks := 0
		blockedTasks := 0

		for _, task := range analysis.CurrentTasks {
			switch task.Status {
			case "in_progress":
				inProgressTasks++
			case "todo":
				todoTasks++
			case "blocked":
				blockedTasks++
			}
		}

		if inProgressTasks > 0 {
			commands = append(commands, cg.createCommand("continue-task", PriorityP0,
				fmt.Sprintf("Continue working on %d task(s) in progress", inProgressTasks),
				[]string{}, []string{"complete-task"}))
		} else if todoTasks > 0 {
			commands = append(commands, cg.createCommand("continue-task", PriorityP1,
				fmt.Sprintf("Start working on %d pending task(s)", todoTasks),
				[]string{}, []string{"complete-task"}))
		}

		if blockedTasks > 0 {
			warnings := []string{fmt.Sprintf("%d task(s) are blocked", blockedTasks)}
			commands = append(commands, cg.createCommandWithWarnings("continue-task", PriorityP2,
				"Resolve blocked tasks to continue progress",
				[]string{}, []string{}, warnings))
		}

		// Check if all tasks are completed
		if analysis.CompletionMetrics.CompletedTasks == analysis.CompletionMetrics.TotalTasks &&
			analysis.CompletionMetrics.TotalTasks > 0 {
			commands = append(commands, cg.createCommand("complete-story", PriorityP0,
				"All tasks completed - mark story as complete",
				[]string{}, []string{"create-story", "complete-epic"}))
		}
	}

	return commands
}

// generateSecondaryCommands creates additional relevant commands
func (cg *CommandGenerator) generateSecondaryCommands(analysis *WorkflowAnalysis) []*ContextualCommand {
	var commands []*ContextualCommand

	// Add creation commands based on current level
	switch analysis.Position {
	case PositionEpicLevel, PositionStoryLevel, PositionTaskLevel:
		if analysis.Position >= PositionEpicLevel {
			commands = append(commands, cg.createCommand("create-epic", PriorityP2,
				"Create a new epic for future work",
				[]string{}, []string{"start-epic"}))
		}
		if analysis.Position >= PositionStoryLevel {
			commands = append(commands, cg.createCommand("create-story", PriorityP2,
				"Add another story to the current epic",
				[]string{}, []string{"create-task"}))
		}
		if analysis.Position >= PositionTaskLevel {
			commands = append(commands, cg.createCommand("create-task", PriorityP2,
				"Add more tasks to the current story",
				[]string{}, []string{"continue-task"}))
		}
	}

	// Add completion commands if appropriate
	if analysis.CurrentEpic != nil && analysis.CompletionMetrics.EpicProgress >= 100 {
		commands = append(commands, cg.createCommand("complete-epic", PriorityP1,
			"Epic appears to be complete - mark as finished",
			[]string{}, []string{"create-epic", "start-epic"}))
	}

	// Add list commands for navigation
	if analysis.CompletionMetrics.TotalEpics > 1 {
		commands = append(commands, cg.createCommand("list-epics", PriorityP2,
			"View all epics and their status",
			[]string{}, []string{"start-epic"}))
	}

	if analysis.CompletionMetrics.TotalStories > 1 {
		commands = append(commands, cg.createCommand("list-stories", PriorityP2,
			"View all stories in the current epic",
			[]string{}, []string{"continue-story"}))
	}

	if analysis.CompletionMetrics.TotalTasks > 1 {
		commands = append(commands, cg.createCommand("list-tasks", PriorityP2,
			"View all tasks in the current story",
			[]string{}, []string{"continue-task"}))
	}

	return commands
}

// generateUtilityCommands creates utility commands that are always relevant
func (cg *CommandGenerator) generateUtilityCommands(analysis *WorkflowAnalysis) []*ContextualCommand {
	var commands []*ContextualCommand

	// Status command - always useful
	reasoning := "View current project status and progress"
	if len(analysis.Blockers) > 0 {
		reasoning = fmt.Sprintf("View status and resolve %d blocker(s)", len(analysis.Blockers))
	}

	commands = append(commands, cg.createCommand("status", PriorityP2, reasoning, []string{}, []string{}))

	// Help command - always available
	commands = append(commands, cg.createCommand("help", PriorityP2,
		"Get help with available commands", []string{}, []string{}))

	return commands
}

// createCommand creates a contextual command with the given parameters
func (cg *CommandGenerator) createCommand(actionID string, priority Priority, reasoning string,
	prerequisites []string, nextActions []string) *ContextualCommand {
	return cg.createCommandWithWarnings(actionID, priority, reasoning, prerequisites, nextActions, []string{})
}

// createCommandWithWarnings creates a contextual command with warnings
func (cg *CommandGenerator) createCommandWithWarnings(actionID string, priority Priority, reasoning string,
	prerequisites []string, nextActions []string, warnings []string) *ContextualCommand {

	action, exists := cg.actionRegistry.GetAction(actionID)
	if !exists {
		// Create a basic action if not found in registry
		action = &WorkflowAction{
			ID:          actionID,
			Name:        actionID,
			Description: "Action not found in registry",
			Priority:    priority,
		}
	}

	return &ContextualCommand{
		Action:        action,
		Priority:      priority,
		Reasoning:     reasoning,
		Prerequisites: prerequisites,
		NextActions:   nextActions,
		Warnings:      warnings,
	}
}

// filterCommands removes commands that can't be executed due to unmet prerequisites or blockers
func (cg *CommandGenerator) filterCommands(commands []*ContextualCommand, analysis *WorkflowAnalysis) []*ContextualCommand {
	var filtered []*ContextualCommand

	for _, cmd := range commands {
		// Check if command is blocked by workflow blockers
		if cg.isCommandBlocked(cmd, analysis) {
			// Add warning but keep the command
			cmd.Warnings = append(cmd.Warnings, "Command may be blocked by current workflow issues")
		}

		// Check prerequisites
		if cg.checkPrerequisites(cmd, analysis) {
			filtered = append(filtered, cmd)
		} else {
			// Command prerequisites not met - skip or add with warning
			cmd.Priority = PriorityP2 // Lower priority for commands with unmet prerequisites
			cmd.Warnings = append(cmd.Warnings, "Some prerequisites may not be met")
			filtered = append(filtered, cmd)
		}
	}

	return filtered
}

// isCommandBlocked checks if a command is blocked by current workflow issues
func (cg *CommandGenerator) isCommandBlocked(cmd *ContextualCommand, analysis *WorkflowAnalysis) bool {
	for _, blocker := range analysis.Blockers {
		// Check if this specific action would be affected by the blocker
		switch blocker.Type {
		case BlockerMissingDefinition:
			if cmd.Action.ID == "continue-epic" || cmd.Action.ID == "continue-story" {
				return true
			}
		case BlockerInconsistentState:
			if cmd.Action.ID == "complete-epic" || cmd.Action.ID == "complete-story" {
				return true
			}
		case BlockerMissingDependency:
			if cmd.Action.ID == "continue-task" || cmd.Action.ID == "complete-task" {
				return true
			}
		}
	}
	return false
}

// checkPrerequisites verifies that command prerequisites are met
func (cg *CommandGenerator) checkPrerequisites(cmd *ContextualCommand, analysis *WorkflowAnalysis) bool {
	if len(cmd.Prerequisites) == 0 && len(cmd.Action.Prerequisites) == 0 {
		return true // No prerequisites to check
	}

	// Combine prerequisites from command and action
	allPrereqs := append(cmd.Prerequisites, cmd.Action.Prerequisites...)

	for _, prereq := range allPrereqs {
		if !cg.isPrerequisiteMet(prereq, analysis) {
			return false
		}
	}

	return true
}

// isPrerequisiteMet checks if a specific prerequisite is satisfied
func (cg *CommandGenerator) isPrerequisiteMet(prerequisite string, analysis *WorkflowAnalysis) bool {
	switch prerequisite {
	case "empty_directory":
		return !analysis.ProjectInitialized
	case "project_initialized":
		return analysis.ProjectInitialized
	case "has_epics":
		return analysis.CompletionMetrics.TotalEpics > 0
	case "no_active_epic":
		return analysis.CurrentEpic == nil
	case "epic_in_progress":
		return analysis.CurrentEpic != nil
	case "story_in_progress":
		return analysis.CurrentStory != nil
	case "task_in_progress":
		return len(analysis.CurrentTasks) > 0
	case "all_tasks_complete":
		return analysis.CompletionMetrics.TotalTasks > 0 &&
			analysis.CompletionMetrics.CompletedTasks == analysis.CompletionMetrics.TotalTasks
	case "all_stories_complete":
		return analysis.CompletionMetrics.TotalStories > 0 &&
			analysis.CompletionMetrics.CompletedStories == analysis.CompletionMetrics.TotalStories
	default:
		// Unknown prerequisite - assume it's not met
		return false
	}
}

// sortCommands sorts commands by priority and relevance
func (cg *CommandGenerator) sortCommands(commands []*ContextualCommand) {
	sort.Slice(commands, func(i, j int) bool {
		// First sort by priority (P0 > P1 > P2)
		if commands[i].Priority != commands[j].Priority {
			return priorityValue(commands[i].Priority) > priorityValue(commands[j].Priority)
		}

		// Then sort by warning count (fewer warnings first)
		if len(commands[i].Warnings) != len(commands[j].Warnings) {
			return len(commands[i].Warnings) < len(commands[j].Warnings)
		}

		// Finally sort alphabetically by action name
		return commands[i].Action.Name < commands[j].Action.Name
	})
}

// priorityValue converts priority to numeric value for sorting
func priorityValue(p Priority) int {
	switch p {
	case PriorityP0:
		return 3
	case PriorityP1:
		return 2
	case PriorityP2:
		return 1
	default:
		return 0
	}
}

// GetRecommendedAction returns the single most recommended action for the current state
func (cg *CommandGenerator) GetRecommendedAction() (*ContextualCommand, error) {
	commands, err := cg.GenerateContextualCommands()
	if err != nil {
		return nil, err
	}

	if len(commands) == 0 {
		return nil, fmt.Errorf("no commands available")
	}

	// Return the highest priority command (already sorted)
	return commands[0], nil
}

// GetCommandsByPriority returns commands grouped by priority level
func (cg *CommandGenerator) GetCommandsByPriority() (map[Priority][]*ContextualCommand, error) {
	commands, err := cg.GenerateContextualCommands()
	if err != nil {
		return nil, err
	}

	grouped := make(map[Priority][]*ContextualCommand)
	for _, cmd := range commands {
		grouped[cmd.Priority] = append(grouped[cmd.Priority], cmd)
	}

	return grouped, nil
}

// ValidateCommand checks if a specific command can be executed in the current context
func (cg *CommandGenerator) ValidateCommand(actionID string) (*ContextualCommand, []string, error) {
	analysis, err := cg.analyzer.AnalyzeWorkflowPosition()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to analyze workflow: %w", err)
	}

	action, exists := cg.actionRegistry.GetAction(actionID)
	if !exists {
		// Create a basic action for unknown commands
		action = &WorkflowAction{
			ID:          actionID,
			Name:        actionID,
			Description: "Unknown action",
			Priority:    PriorityP2,
		}
	}

	cmd := &ContextualCommand{
		Action:   action,
		Priority: action.Priority,
	}

	var issues []string

	// Add issue if action was not found in registry
	if !exists {
		issues = append(issues, "Action not found in registry")
	}

	// Check prerequisites
	if !cg.checkPrerequisites(cmd, analysis) {
		issues = append(issues, "Prerequisites not met")
	}

	// Check for blockers
	if cg.isCommandBlocked(cmd, analysis) {
		issues = append(issues, "Command is blocked by workflow issues")
	}

	// Add specific validation based on action
	specificIssues := cg.validateSpecificAction(actionID, analysis)
	issues = append(issues, specificIssues...)

	return cmd, issues, nil
}

// validateSpecificAction performs action-specific validation
func (cg *CommandGenerator) validateSpecificAction(actionID string, analysis *WorkflowAnalysis) []string {
	var issues []string

	switch actionID {
	case "init-project":
		if analysis.ProjectInitialized {
			issues = append(issues, "Project is already initialized")
		}

	case "create-epic":
		if !analysis.ProjectInitialized {
			issues = append(issues, "Project must be initialized first")
		}

	case "start-epic":
		if analysis.CompletionMetrics.TotalEpics == 0 {
			issues = append(issues, "No epics available to start")
		}
		if analysis.CurrentEpic != nil {
			issues = append(issues, "An epic is already active")
		}

	case "create-story":
		if analysis.CurrentEpic == nil {
			issues = append(issues, "No epic is currently active")
		}

	case "create-task":
		if analysis.CurrentStory == nil {
			issues = append(issues, "No story is currently active")
		}

	case "continue-task":
		if len(analysis.CurrentTasks) == 0 {
			issues = append(issues, "No tasks are currently defined")
		}

	case "complete-epic":
		if analysis.CurrentEpic == nil {
			issues = append(issues, "No epic is currently active")
		}
		if analysis.CompletionMetrics.EpicProgress < 100 {
			issues = append(issues, "Epic is not fully complete")
		}

	case "complete-story":
		if analysis.CurrentStory == nil {
			issues = append(issues, "No story is currently active")
		}
		if analysis.CompletionMetrics.StoryProgress < 100 {
			issues = append(issues, "Story is not fully complete")
		}
	}

	return issues
}
