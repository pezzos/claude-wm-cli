package workflow

import (
	"fmt"
	"strings"
)

// ValidationResult represents the result of dependency validation
type ValidationResult struct {
	IsValid      bool                  `json:"is_valid"`
	Violations   []DependencyViolation `json:"violations,omitempty"`
	Suggestions  []string              `json:"suggestions,omitempty"`
	Warnings     []string              `json:"warnings,omitempty"`
	CanOverride  bool                  `json:"can_override"`
	OverrideRisk string                `json:"override_risk,omitempty"`
}

// DependencyViolation represents a specific violation of workflow dependencies
type DependencyViolation struct {
	Type          ViolationType `json:"type"`
	Severity      Severity      `json:"severity"`
	Description   string        `json:"description"`
	Prerequisite  string        `json:"prerequisite,omitempty"`
	CurrentState  string        `json:"current_state,omitempty"`
	RequiredState string        `json:"required_state,omitempty"`
	Blocker       string        `json:"blocker,omitempty"`
}

// ViolationType categorizes different types of dependency violations
type ViolationType string

const (
	ViolationMissingPrerequisite ViolationType = "missing_prerequisite"
	ViolationInvalidState        ViolationType = "invalid_state"
	ViolationCircularDependency  ViolationType = "circular_dependency"
	ViolationBlockingCondition   ViolationType = "blocking_condition"
	ViolationIncompatibleAction  ViolationType = "incompatible_action"
)

// Severity levels for dependency violations
type Severity string

const (
	SeverityCritical Severity = "critical"
	SeverityHigh     Severity = "high"
	SeverityMedium   Severity = "medium"
	SeverityLow      Severity = "low"
	SeverityWarning  Severity = "warning"
)

// DependencyEnforcer handles validation and enforcement of workflow dependencies
type DependencyEnforcer struct {
	analyzer         *WorkflowAnalyzer
	commandGenerator *CommandGenerator
}

// NewDependencyEnforcer creates a new dependency enforcer
func NewDependencyEnforcer(rootPath string) *DependencyEnforcer {
	return &DependencyEnforcer{
		analyzer:         NewWorkflowAnalyzer(rootPath),
		commandGenerator: NewCommandGenerator(rootPath),
	}
}

// ValidateActionExecution validates if an action can be executed in the current workflow state
func (de *DependencyEnforcer) ValidateActionExecution(actionID string, allowOverride bool) (*ValidationResult, error) {
	// Get current workflow analysis
	analysis, err := de.analyzer.AnalyzeWorkflowPosition()
	if err != nil {
		return nil, fmt.Errorf("failed to analyze workflow: %w", err)
	}

	// Get action from registry
	action, exists := de.commandGenerator.actionRegistry.GetAction(actionID)
	if !exists {
		return &ValidationResult{
			IsValid: false,
			Violations: []DependencyViolation{
				{
					Type:        ViolationInvalidState,
					Severity:    SeverityHigh,
					Description: fmt.Sprintf("Action '%s' is not recognized", actionID),
				},
			},
			Suggestions: []string{"Use 'help' command to see available actions"},
			CanOverride: false,
		}, nil
	}

	result := &ValidationResult{
		IsValid:     true,
		Violations:  []DependencyViolation{},
		Suggestions: []string{},
		Warnings:    []string{},
		CanOverride: true,
	}

	// Validate prerequisites
	de.validatePrerequisites(action, analysis, result)

	// Validate workflow state compatibility
	de.validateWorkflowState(action, analysis, result)

	// Validate against current blockers
	de.validateAgainstBlockers(action, analysis, result)

	// Check for circular dependencies
	de.validateCircularDependencies(action, analysis, result)

	// Validate action-specific constraints
	de.validateActionSpecificConstraints(actionID, analysis, result)

	// Generate suggestions for resolution
	de.generateResolutionSuggestions(action, analysis, result)

	// Determine if override is allowed
	de.evaluateOverrideCapability(action, analysis, result, allowOverride)

	// Set overall validity
	result.IsValid = len(result.Violations) == 0 ||
		(allowOverride && result.CanOverride && de.onlyNonCriticalViolations(result.Violations))

	return result, nil
}

// validatePrerequisites checks if all action prerequisites are met
func (de *DependencyEnforcer) validatePrerequisites(action *WorkflowAction, analysis *WorkflowAnalysis, result *ValidationResult) {
	for _, prereq := range action.Prerequisites {
		if !de.isPrerequisiteMet(prereq, analysis) {
			violation := DependencyViolation{
				Type:          ViolationMissingPrerequisite,
				Severity:      de.getPrerequisiteSeverity(prereq),
				Description:   fmt.Sprintf("Prerequisite '%s' is not met", prereq),
				Prerequisite:  prereq,
				CurrentState:  de.getCurrentStateForPrerequisite(prereq, analysis),
				RequiredState: de.getRequiredStateForPrerequisite(prereq),
			}
			result.Violations = append(result.Violations, violation)
		}
	}
}

// validateWorkflowState checks if the action is compatible with current workflow state
func (de *DependencyEnforcer) validateWorkflowState(action *WorkflowAction, analysis *WorkflowAnalysis, result *ValidationResult) {
	incompatibilities := de.getStateIncompatibilities(action.ID, analysis)
	for _, incompatibility := range incompatibilities {
		violation := DependencyViolation{
			Type:        ViolationIncompatibleAction,
			Severity:    SeverityMedium,
			Description: incompatibility,
		}
		result.Violations = append(result.Violations, violation)
	}
}

// validateAgainstBlockers checks if current workflow blockers prevent action execution
func (de *DependencyEnforcer) validateAgainstBlockers(action *WorkflowAction, analysis *WorkflowAnalysis, result *ValidationResult) {
	for _, blocker := range analysis.Blockers {
		if de.isActionBlockedBy(action.ID, blocker) {
			violation := DependencyViolation{
				Type:        ViolationBlockingCondition,
				Severity:    de.mapBlockerSeverity(blocker.Severity),
				Description: fmt.Sprintf("Action blocked by: %s", blocker.Description),
				Blocker:     blocker.Entity,
			}
			result.Violations = append(result.Violations, violation)
		}
	}
}

// validateCircularDependencies checks for potential circular dependencies
func (de *DependencyEnforcer) validateCircularDependencies(action *WorkflowAction, analysis *WorkflowAnalysis, result *ValidationResult) {
	// For now, implement basic circular dependency detection
	// This could be enhanced with a more sophisticated dependency graph analysis

	if action.ID == "complete-epic" && analysis.CurrentEpic != nil {
		// Check if epic has incomplete dependencies
		if len(action.Blocks) > 0 {
			for _, blockedAction := range action.Blocks {
				if de.isActionExecutable(blockedAction, analysis) {
					violation := DependencyViolation{
						Type:        ViolationCircularDependency,
						Severity:    SeverityHigh,
						Description: fmt.Sprintf("Completing epic would block '%s' which is still needed", blockedAction),
					}
					result.Violations = append(result.Violations, violation)
				}
			}
		}
	}
}

// validateActionSpecificConstraints checks constraints specific to certain actions
func (de *DependencyEnforcer) validateActionSpecificConstraints(actionID string, analysis *WorkflowAnalysis, result *ValidationResult) {
	switch actionID {
	case "init-project":
		if analysis.ProjectInitialized {
			violation := DependencyViolation{
				Type:          ViolationInvalidState,
				Severity:      SeverityHigh,
				Description:   "Project is already initialized",
				CurrentState:  "initialized",
				RequiredState: "not_initialized",
			}
			result.Violations = append(result.Violations, violation)
		}

	case "create-epic":
		if !analysis.ProjectInitialized {
			violation := DependencyViolation{
				Type:          ViolationMissingPrerequisite,
				Severity:      SeverityCritical,
				Description:   "Project must be initialized before creating epics",
				Prerequisite:  "project_initialized",
				CurrentState:  "not_initialized",
				RequiredState: "initialized",
			}
			result.Violations = append(result.Violations, violation)
		}

	case "start-epic":
		if analysis.CompletionMetrics.TotalEpics == 0 {
			violation := DependencyViolation{
				Type:          ViolationMissingPrerequisite,
				Severity:      SeverityHigh,
				Description:   "No epics available to start",
				Prerequisite:  "has_epics",
				CurrentState:  "no_epics",
				RequiredState: "epics_available",
			}
			result.Violations = append(result.Violations, violation)
		}
		if analysis.CurrentEpic != nil {
			violation := DependencyViolation{
				Type:          ViolationInvalidState,
				Severity:      SeverityMedium,
				Description:   "Another epic is already active",
				CurrentState:  fmt.Sprintf("epic_%s_active", analysis.CurrentEpic.Metadata.ID),
				RequiredState: "no_active_epic",
			}
			result.Violations = append(result.Violations, violation)
		}

	case "complete-epic":
		if analysis.CurrentEpic == nil {
			violation := DependencyViolation{
				Type:          ViolationMissingPrerequisite,
				Severity:      SeverityHigh,
				Description:   "No epic is currently active",
				CurrentState:  "no_active_epic",
				RequiredState: "epic_active",
			}
			result.Violations = append(result.Violations, violation)
		} else if analysis.CompletionMetrics.EpicProgress < 100 {
			violation := DependencyViolation{
				Type:          ViolationInvalidState,
				Severity:      SeverityMedium,
				Description:   fmt.Sprintf("Epic is only %.1f%% complete", analysis.CompletionMetrics.EpicProgress),
				CurrentState:  fmt.Sprintf("%.1f%%_complete", analysis.CompletionMetrics.EpicProgress),
				RequiredState: "100%_complete",
			}
			result.Violations = append(result.Violations, violation)
		}

	case "create-story":
		if analysis.CurrentEpic == nil {
			violation := DependencyViolation{
				Type:          ViolationMissingPrerequisite,
				Severity:      SeverityHigh,
				Description:   "No epic is currently active",
				CurrentState:  "no_active_epic",
				RequiredState: "epic_active",
			}
			result.Violations = append(result.Violations, violation)
		}

	case "create-task":
		if analysis.CurrentStory == nil {
			violation := DependencyViolation{
				Type:          ViolationMissingPrerequisite,
				Severity:      SeverityHigh,
				Description:   "No story is currently active",
				CurrentState:  "no_active_story",
				RequiredState: "story_active",
			}
			result.Violations = append(result.Violations, violation)
		}
	}
}

// generateResolutionSuggestions creates actionable suggestions to resolve violations
func (de *DependencyEnforcer) generateResolutionSuggestions(action *WorkflowAction, analysis *WorkflowAnalysis, result *ValidationResult) {
	if len(result.Violations) == 0 {
		return
	}

	suggestionSet := make(map[string]bool) // Use map to avoid duplicates

	for _, violation := range result.Violations {
		switch violation.Type {
		case ViolationMissingPrerequisite:
			suggestions := de.getSuggestionsForPrerequisite(violation.Prerequisite, analysis)
			for _, suggestion := range suggestions {
				suggestionSet[suggestion] = true
			}

		case ViolationInvalidState:
			suggestions := de.getSuggestionsForStateTransition(violation.CurrentState, violation.RequiredState, analysis)
			for _, suggestion := range suggestions {
				suggestionSet[suggestion] = true
			}

		case ViolationBlockingCondition:
			suggestionSet["Resolve workflow blockers before proceeding"] = true
			if violation.Blocker != "" {
				suggestionSet[fmt.Sprintf("Address issue with '%s'", violation.Blocker)] = true
			}
		}
	}

	// Convert set to slice
	for suggestion := range suggestionSet {
		result.Suggestions = append(result.Suggestions, suggestion)
	}
}

// evaluateOverrideCapability determines if the action can be overridden and at what risk
func (de *DependencyEnforcer) evaluateOverrideCapability(action *WorkflowAction, analysis *WorkflowAnalysis, result *ValidationResult, allowOverride bool) {
	if !allowOverride {
		result.CanOverride = false
		return
	}

	// Critical violations cannot be overridden
	hasCriticalViolations := false
	for _, violation := range result.Violations {
		if violation.Severity == SeverityCritical {
			hasCriticalViolations = true
			break
		}
	}

	if hasCriticalViolations {
		result.CanOverride = false
		result.OverrideRisk = "Critical dependencies prevent override"
		return
	}

	// Evaluate override risk based on action and violations
	riskLevel := de.calculateOverrideRisk(action, result.Violations)
	result.CanOverride = true
	result.OverrideRisk = riskLevel

	// Add warnings for override
	if len(result.Violations) > 0 {
		result.Warnings = append(result.Warnings, "Overriding dependencies may cause workflow inconsistencies")
		if riskLevel == "high" {
			result.Warnings = append(result.Warnings, "High risk: proceeding may require manual cleanup")
		}
	}
}

// Helper methods

func (de *DependencyEnforcer) isPrerequisiteMet(prerequisite string, analysis *WorkflowAnalysis) bool {
	// Reuse the implementation from command generator
	generator := de.commandGenerator
	return generator.isPrerequisiteMet(prerequisite, analysis)
}

func (de *DependencyEnforcer) getPrerequisiteSeverity(prerequisite string) Severity {
	// Map prerequisites to severity levels
	criticalPrereqs := []string{"project_initialized"}
	for _, critical := range criticalPrereqs {
		if prerequisite == critical {
			return SeverityCritical
		}
	}
	return SeverityHigh
}

func (de *DependencyEnforcer) getCurrentStateForPrerequisite(prerequisite string, analysis *WorkflowAnalysis) string {
	switch prerequisite {
	case "project_initialized":
		if analysis.ProjectInitialized {
			return "initialized"
		}
		return "not_initialized"
	case "has_epics":
		return fmt.Sprintf("%d_epics", analysis.CompletionMetrics.TotalEpics)
	case "epic_in_progress":
		if analysis.CurrentEpic != nil {
			return fmt.Sprintf("epic_%s_active", analysis.CurrentEpic.Metadata.ID)
		}
		return "no_active_epic"
	case "story_in_progress":
		if analysis.CurrentStory != nil {
			return fmt.Sprintf("story_%s_active", analysis.CurrentStory.Metadata.ID)
		}
		return "no_active_story"
	default:
		return "unknown"
	}
}

func (de *DependencyEnforcer) getRequiredStateForPrerequisite(prerequisite string) string {
	switch prerequisite {
	case "project_initialized":
		return "initialized"
	case "has_epics":
		return "epics_available"
	case "epic_in_progress":
		return "epic_active"
	case "story_in_progress":
		return "story_active"
	default:
		return "required_state"
	}
}

func (de *DependencyEnforcer) getStateIncompatibilities(actionID string, analysis *WorkflowAnalysis) []string {
	var incompatibilities []string

	// Define incompatible state combinations
	switch actionID {
	case "init-project":
		if analysis.ProjectInitialized {
			incompatibilities = append(incompatibilities, "Cannot initialize an already initialized project")
		}
	case "create-epic":
		if analysis.CurrentEpic != nil && analysis.CompletionMetrics.EpicProgress < 100 {
			incompatibilities = append(incompatibilities, "Consider completing current epic before creating a new one")
		}
	}

	return incompatibilities
}

func (de *DependencyEnforcer) isActionBlockedBy(actionID string, blocker WorkflowBlocker) bool {
	// Map blocker types to affected actions
	switch blocker.Type {
	case BlockerMissingDefinition:
		blockedActions := []string{"continue-epic", "continue-story", "complete-epic", "complete-story"}
		return de.containsString(blockedActions, actionID)
	case BlockerMissingDependency:
		blockedActions := []string{"continue-task", "complete-task"}
		return de.containsString(blockedActions, actionID)
	case BlockerInconsistentState:
		blockedActions := []string{"complete-epic", "complete-story", "complete-task"}
		return de.containsString(blockedActions, actionID)
	default:
		return false
	}
}

func (de *DependencyEnforcer) mapBlockerSeverity(severity string) Severity {
	switch severity {
	case "critical":
		return SeverityCritical
	case "high":
		return SeverityHigh
	case "medium":
		return SeverityMedium
	case "low":
		return SeverityLow
	default:
		return SeverityMedium
	}
}

func (de *DependencyEnforcer) isActionExecutable(actionID string, analysis *WorkflowAnalysis) bool {
	// Simple check if action can be executed (this could be more sophisticated)
	switch actionID {
	case "create-story":
		return analysis.CurrentEpic != nil
	case "create-task":
		return analysis.CurrentStory != nil
	case "continue-task":
		return len(analysis.CurrentTasks) > 0
	default:
		return true
	}
}

func (de *DependencyEnforcer) getSuggestionsForPrerequisite(prerequisite string, analysis *WorkflowAnalysis) []string {
	switch prerequisite {
	case "project_initialized":
		return []string{"Run 'init-project' to initialize the project structure"}
	case "has_epics":
		return []string{"Create an epic using 'create-epic' command"}
	case "epic_in_progress":
		if analysis.CompletionMetrics.TotalEpics > 0 {
			return []string{"Start an existing epic using 'start-epic' command"}
		}
		return []string{"Create and start an epic using 'create-epic' command"}
	case "story_in_progress":
		return []string{"Create a story using 'create-story' command"}
	case "no_active_epic":
		return []string{"Complete or switch the current epic before starting a new one"}
	default:
		return []string{fmt.Sprintf("Ensure '%s' prerequisite is met", prerequisite)}
	}
}

func (de *DependencyEnforcer) getSuggestionsForStateTransition(currentState, requiredState string, analysis *WorkflowAnalysis) []string {
	if strings.Contains(currentState, "not_initialized") && requiredState == "initialized" {
		return []string{"Initialize the project using 'init-project' command"}
	}
	if strings.Contains(currentState, "no_epics") && strings.Contains(requiredState, "epics") {
		return []string{"Create epics using 'create-epic' command"}
	}
	if strings.Contains(currentState, "no_active_epic") && strings.Contains(requiredState, "epic") {
		return []string{"Start an epic using 'start-epic' command"}
	}
	return []string{fmt.Sprintf("Transition from '%s' to '%s'", currentState, requiredState)}
}

func (de *DependencyEnforcer) calculateOverrideRisk(action *WorkflowAction, violations []DependencyViolation) string {
	highRiskActions := []string{"init-project", "complete-epic", "complete-story"}
	if de.containsString(highRiskActions, action.ID) {
		return "high"
	}

	highSeverityCount := 0
	for _, violation := range violations {
		if violation.Severity == SeverityHigh {
			highSeverityCount++
		}
	}

	if highSeverityCount > 2 {
		return "high"
	} else if highSeverityCount > 0 {
		return "medium"
	}
	return "low"
}

func (de *DependencyEnforcer) onlyNonCriticalViolations(violations []DependencyViolation) bool {
	for _, violation := range violations {
		if violation.Severity == SeverityCritical {
			return false
		}
	}
	return true
}

func (de *DependencyEnforcer) containsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// ValidateWorkflowTransition validates if a transition between workflow states is valid
func (de *DependencyEnforcer) ValidateWorkflowTransition(fromState, toState WorkflowPosition, actionID string) (*ValidationResult, error) {
	validTransitions := map[WorkflowPosition][]WorkflowPosition{
		PositionNotInitialized: {PositionProjectLevel},
		PositionProjectLevel:   {PositionEpicLevel},
		PositionEpicLevel:      {PositionStoryLevel, PositionProjectLevel},
		PositionStoryLevel:     {PositionTaskLevel, PositionEpicLevel},
		PositionTaskLevel:      {PositionStoryLevel},
	}

	result := &ValidationResult{
		IsValid:     true,
		Violations:  []DependencyViolation{},
		Suggestions: []string{},
		CanOverride: true,
	}

	// Check if transition is valid
	allowedTransitions, exists := validTransitions[fromState]
	if !exists {
		result.IsValid = false
		result.Violations = append(result.Violations, DependencyViolation{
			Type:        ViolationInvalidState,
			Severity:    SeverityHigh,
			Description: fmt.Sprintf("Unknown workflow state: %s", fromState),
		})
		return result, nil
	}

	validTransition := false
	for _, allowed := range allowedTransitions {
		if allowed == toState {
			validTransition = true
			break
		}
	}

	if !validTransition {
		result.IsValid = false
		result.Violations = append(result.Violations, DependencyViolation{
			Type:          ViolationInvalidState,
			Severity:      SeverityMedium,
			Description:   fmt.Sprintf("Invalid transition from %s to %s", fromState, toState),
			CurrentState:  string(fromState),
			RequiredState: fmt.Sprintf("one of: %v", allowedTransitions),
		})
		result.Suggestions = append(result.Suggestions,
			fmt.Sprintf("Valid transitions from %s are: %v", fromState, allowedTransitions))
	}

	return result, nil
}

// GetAllowedActions returns actions that can be executed without violations
func (de *DependencyEnforcer) GetAllowedActions() ([]*WorkflowAction, error) {
	allActions := de.commandGenerator.actionRegistry.GetAllActions()
	var allowedActions []*WorkflowAction

	for _, action := range allActions {
		result, err := de.ValidateActionExecution(action.ID, false)
		if err != nil {
			continue // Skip actions that can't be validated
		}

		if result.IsValid {
			allowedActions = append(allowedActions, action)
		}
	}

	return allowedActions, nil
}

// GetBlockedActions returns actions that are currently blocked with their violations
func (de *DependencyEnforcer) GetBlockedActions() (map[string]*ValidationResult, error) {
	allActions := de.commandGenerator.actionRegistry.GetAllActions()
	blockedActions := make(map[string]*ValidationResult)

	for _, action := range allActions {
		result, err := de.ValidateActionExecution(action.ID, false)
		if err != nil {
			continue // Skip actions that can't be validated
		}

		if !result.IsValid {
			blockedActions[action.ID] = result
		}
	}

	return blockedActions, nil
}
