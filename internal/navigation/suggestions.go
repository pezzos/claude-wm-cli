package navigation

import (
	"fmt"
	"sort"
	"strings"

	"claude-wm-cli/internal/workflow"
)

// Suggestion represents a suggested action for the user
type Suggestion struct {
	Action      *workflow.WorkflowAction
	Priority    workflow.Priority
	Reasoning   string   // Why this action is suggested
	Urgency     int      // 1-10 scale for ordering within priority
	Conditions  []string // Current conditions that make this suggestion valid
	NextActions []string // What actions become available after this one
}

// SuggestionEngine generates contextual suggestions based on project state
type SuggestionEngine struct {
	actionRegistry *workflow.ActionRegistry
}

// NewSuggestionEngine creates a new suggestion engine
func NewSuggestionEngine() *SuggestionEngine {
	return &SuggestionEngine{
		actionRegistry: workflow.NewActionRegistry(),
	}
}

// GenerateSuggestions analyzes the project context and returns prioritized suggestions
func (se *SuggestionEngine) GenerateSuggestions(ctx *ProjectContext) ([]*Suggestion, error) {
	if ctx == nil {
		return nil, fmt.Errorf("project context is nil")
	}

	var suggestions []*Suggestion

	// Generate state-specific suggestions
	stateSuggestions := se.generateStateSuggestions(ctx)
	suggestions = append(suggestions, stateSuggestions...)

	// Generate context-specific suggestions
	contextSuggestions := se.generateContextSuggestions(ctx)
	suggestions = append(suggestions, contextSuggestions...)

	// Generate general suggestions that are always available
	generalSuggestions := se.generateGeneralSuggestions(ctx)
	suggestions = append(suggestions, generalSuggestions...)

	// Sort suggestions by priority and urgency
	se.sortSuggestions(suggestions)

	// Remove duplicates and filter by context
	suggestions = se.filterSuggestions(suggestions, ctx)

	return suggestions, nil
}

// generateStateSuggestions generates suggestions based on the current workflow state
func (se *SuggestionEngine) generateStateSuggestions(ctx *ProjectContext) []*Suggestion {
	var suggestions []*Suggestion

	switch ctx.State {
	case StateNotInitialized:
		suggestions = append(suggestions, &Suggestion{
			Action:      se.getAction("init-project"),
			Priority:    workflow.PriorityP0,
			Reasoning:   "Project is not initialized. You need to set up the project structure before you can start working.",
			Urgency:     10,
			Conditions:  []string{"no_docs_directory"},
			NextActions: []string{"create-epic"},
		})

	case StateProjectInitialized:
		suggestions = append(suggestions, &Suggestion{
			Action:      se.getAction("create-epic"),
			Priority:    workflow.PriorityP0,
			Reasoning:   "Project is initialized but has no epics. Create your first epic to start organizing work.",
			Urgency:     10,
			Conditions:  []string{"docs_structure_exists", "no_epics"},
			NextActions: []string{"start-epic"},
		})

	case StateHasEpics:
		suggestions = append(suggestions, &Suggestion{
			Action:      se.getAction("start-epic"),
			Priority:    workflow.PriorityP0,
			Reasoning:   "Epics are available but none are currently active. Start an epic to begin working.",
			Urgency:     10,
			Conditions:  []string{"epics_exist", "no_active_epic"},
			NextActions: []string{"continue-epic", "create-story"},
		})

	case StateEpicInProgress:
		if ctx.CurrentStory == nil {
			epicTitle := "current epic"
			if ctx.CurrentEpic != nil {
				epicTitle = fmt.Sprintf("'%s'", ctx.CurrentEpic.Title)
			}
			suggestions = append(suggestions, &Suggestion{
				Action:      se.getAction("continue-epic"),
				Priority:    workflow.PriorityP1,
				Reasoning:   fmt.Sprintf("Epic %s is in progress but no story is active. Continue with the next story in the epic.", epicTitle),
				Urgency:     9,
				Conditions:  []string{"epic_active", "no_active_story"},
				NextActions: []string{"continue-story", "create-task"},
			})
		} else {
			storyTitle := "current story"
			if ctx.CurrentStory != nil {
				storyTitle = fmt.Sprintf("'%s'", ctx.CurrentStory.Title)
			}
			epicTitle := "current epic"
			if ctx.CurrentEpic != nil {
				epicTitle = fmt.Sprintf("'%s'", ctx.CurrentEpic.Title)
			}
			suggestions = append(suggestions, &Suggestion{
				Action:      se.getAction("continue-story"),
				Priority:    workflow.PriorityP1,
				Reasoning:   fmt.Sprintf("Continue working on story %s in epic %s.", storyTitle, epicTitle),
				Urgency:     9,
				Conditions:  []string{"story_active"},
				NextActions: []string{"continue-task", "complete-story"},
			})
		}

	case StateStoryInProgress:
		if ctx.CurrentTask == nil {
			storyTitle := "current story"
			if ctx.CurrentStory != nil {
				storyTitle = fmt.Sprintf("'%s'", ctx.CurrentStory.Title)
			}
			suggestions = append(suggestions, &Suggestion{
				Action:      se.getAction("continue-story"),
				Priority:    workflow.PriorityP1,
				Reasoning:   fmt.Sprintf("Story %s is in progress. Continue with the next task or create a new one.", storyTitle),
				Urgency:     9,
				Conditions:  []string{"story_active", "no_active_task"},
				NextActions: []string{"continue-task", "create-task"},
			})
		} else {
			taskTitle := "current task"
			if ctx.CurrentTask != nil {
				taskTitle = fmt.Sprintf("'%s'", ctx.CurrentTask.Title)
			}
			suggestions = append(suggestions, &Suggestion{
				Action:      se.getAction("continue-task"),
				Priority:    workflow.PriorityP1,
				Reasoning:   fmt.Sprintf("Continue working on task %s.", taskTitle),
				Urgency:     10,
				Conditions:  []string{"task_active"},
				NextActions: []string{"complete-task"},
			})
		}

	case StateTaskInProgress:
		taskTitle := "current task"
		if ctx.CurrentTask != nil {
			taskTitle = fmt.Sprintf("'%s'", ctx.CurrentTask.Title)
		}
		suggestions = append(suggestions, &Suggestion{
			Action:      se.getAction("continue-task"),
			Priority:    workflow.PriorityP1,
			Reasoning:   fmt.Sprintf("Task %s is in progress. Continue working on it.", taskTitle),
			Urgency:     10,
			Conditions:  []string{"task_active"},
			NextActions: []string{"complete-task", "create-task"},
		})
	}

	return suggestions
}

// generateContextSuggestions generates suggestions based on specific context information
func (se *SuggestionEngine) generateContextSuggestions(ctx *ProjectContext) []*Suggestion {
	var suggestions []*Suggestion

	// Suggest completion actions if progress indicates near completion
	if ctx.CurrentEpic != nil && ctx.CurrentEpic.Progress > 0.8 {
		suggestions = append(suggestions, &Suggestion{
			Action:      se.getAction("complete-epic"),
			Priority:    workflow.PriorityP1,
			Reasoning:   fmt.Sprintf("Epic '%s' is %.0f%% complete. Consider completing it.", ctx.CurrentEpic.Title, ctx.CurrentEpic.Progress*100),
			Urgency:     8,
			Conditions:  []string{"epic_near_completion"},
			NextActions: []string{"start-epic", "create-epic"},
		})
	}

	if ctx.CurrentStory != nil && ctx.CurrentStory.Progress > 0.8 {
		suggestions = append(suggestions, &Suggestion{
			Action:      se.getAction("complete-story"),
			Priority:    workflow.PriorityP1,
			Reasoning:   fmt.Sprintf("Story '%s' is %.0f%% complete. Consider completing it.", ctx.CurrentStory.Title, ctx.CurrentStory.Progress*100),
			Urgency:     8,
			Conditions:  []string{"story_near_completion"},
			NextActions: []string{"continue-epic", "create-story"},
		})
	}

	// Suggest creating new work items if current level is empty
	if ctx.State >= StateEpicInProgress {
		if ctx.CurrentEpic != nil && ctx.CurrentEpic.TotalStories == 0 {
			suggestions = append(suggestions, &Suggestion{
				Action:      se.getAction("create-story"),
				Priority:    workflow.PriorityP1,
				Reasoning:   "Current epic has no stories. Create the first story to start work.",
				Urgency:     9,
				Conditions:  []string{"epic_active", "no_stories"},
				NextActions: []string{"continue-story"},
			})
		}

		if ctx.CurrentStory != nil && ctx.CurrentStory.TotalTasks == 0 {
			suggestions = append(suggestions, &Suggestion{
				Action:      se.getAction("create-task"),
				Priority:    workflow.PriorityP1,
				Reasoning:   "Current story has no tasks. Create the first task to begin implementation.",
				Urgency:     9,
				Conditions:  []string{"story_active", "no_tasks"},
				NextActions: []string{"continue-task"},
			})
		}
	}

	// Suggest addressing issues if any exist
	if len(ctx.Issues) > 0 {
		suggestions = append(suggestions, &Suggestion{
			Action: &workflow.WorkflowAction{
				ID:          "fix-issues",
				Name:        "Fix Issues",
				Description: "Address project issues and warnings",
				Priority:    workflow.PriorityP1,
			},
			Priority:    workflow.PriorityP1,
			Reasoning:   fmt.Sprintf("There are %d project issues that need attention: %s", len(ctx.Issues), strings.Join(ctx.Issues[:1], ", ")),
			Urgency:     7,
			Conditions:  []string{"has_issues"},
			NextActions: []string{"status"},
		})
	}

	return suggestions
}

// generateGeneralSuggestions generates general suggestions that are usually available
func (se *SuggestionEngine) generateGeneralSuggestions(ctx *ProjectContext) []*Suggestion {
	var suggestions []*Suggestion

	// Always suggest status if project is initialized
	if ctx.State > StateNotInitialized {
		suggestions = append(suggestions, &Suggestion{
			Action:      se.getAction("status"),
			Priority:    workflow.PriorityP2,
			Reasoning:   "Check current project status and progress.",
			Urgency:     5,
			Conditions:  []string{"project_initialized"},
			NextActions: []string{},
		})
	}

	// Always suggest help
	suggestions = append(suggestions, &Suggestion{
		Action:      se.getAction("help"),
		Priority:    workflow.PriorityP2,
		Reasoning:   "Get help with available commands and workflows.",
		Urgency:     3,
		Conditions:  []string{},
		NextActions: []string{},
	})

	// Suggest list commands based on state
	if ctx.State >= StateHasEpics {
		suggestions = append(suggestions, &Suggestion{
			Action:      se.getAction("list-epics"),
			Priority:    workflow.PriorityP2,
			Reasoning:   "View all available epics and their status.",
			Urgency:     4,
			Conditions:  []string{"has_epics"},
			NextActions: []string{"start-epic"},
		})
	}

	if ctx.State >= StateEpicInProgress {
		suggestions = append(suggestions, &Suggestion{
			Action:      se.getAction("list-stories"),
			Priority:    workflow.PriorityP2,
			Reasoning:   "View all stories in the current epic.",
			Urgency:     4,
			Conditions:  []string{"epic_active"},
			NextActions: []string{"continue-story"},
		})
	}

	if ctx.State >= StateStoryInProgress {
		suggestions = append(suggestions, &Suggestion{
			Action:      se.getAction("list-tasks"),
			Priority:    workflow.PriorityP2,
			Reasoning:   "View all tasks in the current story.",
			Urgency:     4,
			Conditions:  []string{"story_active"},
			NextActions: []string{"continue-task"},
		})
	}

	return suggestions
}

// getAction is a helper to safely get an action from the registry
func (se *SuggestionEngine) getAction(id string) *workflow.WorkflowAction {
	if action, exists := se.actionRegistry.GetAction(id); exists {
		return action
	}

	// Return a fallback action if not found
	return &workflow.WorkflowAction{
		ID:          id,
		Name:        strings.Title(strings.ReplaceAll(id, "-", " ")),
		Description: fmt.Sprintf("Execute %s action", id),
		Priority:    workflow.PriorityP2,
	}
}

// sortSuggestions sorts suggestions by priority (P0 > P1 > P2) and then by urgency (higher first)
func (se *SuggestionEngine) sortSuggestions(suggestions []*Suggestion) {
	sort.Slice(suggestions, func(i, j int) bool {
		// First sort by priority
		priorityOrder := map[workflow.Priority]int{
			workflow.PriorityP0: 3,
			workflow.PriorityP1: 2,
			workflow.PriorityP2: 1,
		}

		iPriority := priorityOrder[suggestions[i].Priority]
		jPriority := priorityOrder[suggestions[j].Priority]

		if iPriority != jPriority {
			return iPriority > jPriority
		}

		// If same priority, sort by urgency
		return suggestions[i].Urgency > suggestions[j].Urgency
	})
}

// filterSuggestions removes duplicates and filters suggestions based on context
func (se *SuggestionEngine) filterSuggestions(suggestions []*Suggestion, ctx *ProjectContext) []*Suggestion {
	seen := make(map[string]bool)
	var filtered []*Suggestion

	for _, suggestion := range suggestions {
		// Skip if we've already seen this action
		if seen[suggestion.Action.ID] {
			continue
		}
		seen[suggestion.Action.ID] = true

		// Filter based on context conditions
		if se.shouldIncludeSuggestion(suggestion, ctx) {
			filtered = append(filtered, suggestion)
		}
	}

	return filtered
}

// shouldIncludeSuggestion determines if a suggestion should be included based on context
func (se *SuggestionEngine) shouldIncludeSuggestion(suggestion *Suggestion, ctx *ProjectContext) bool {
	// Always include P0 suggestions
	if suggestion.Priority == workflow.PriorityP0 {
		return true
	}

	// Limit total number of suggestions to avoid overwhelming the user
	// This could be made configurable
	maxSuggestions := 8

	// For now, we'll include all suggestions but this could be enhanced
	// with more sophisticated filtering logic
	_ = maxSuggestions

	return true
}

// GetTopSuggestion returns the highest priority suggestion
func (se *SuggestionEngine) GetTopSuggestion(ctx *ProjectContext) (*Suggestion, error) {
	suggestions, err := se.GenerateSuggestions(ctx)
	if err != nil {
		return nil, err
	}

	if len(suggestions) == 0 {
		return nil, fmt.Errorf("no suggestions available")
	}

	return suggestions[0], nil
}

// GetSuggestionsByPriority returns suggestions grouped by priority
func (se *SuggestionEngine) GetSuggestionsByPriority(ctx *ProjectContext) (map[workflow.Priority][]*Suggestion, error) {
	suggestions, err := se.GenerateSuggestions(ctx)
	if err != nil {
		return nil, err
	}

	grouped := make(map[workflow.Priority][]*Suggestion)
	for _, suggestion := range suggestions {
		grouped[suggestion.Priority] = append(grouped[suggestion.Priority], suggestion)
	}

	return grouped, nil
}

// FormatSuggestion returns a formatted string representation of a suggestion
func (se *SuggestionEngine) FormatSuggestion(suggestion *Suggestion, includeReasoning bool) string {
	if suggestion == nil || suggestion.Action == nil {
		return "No suggestion available"
	}

	formatted := fmt.Sprintf("[%s] %s", suggestion.Priority, suggestion.Action.Name)

	if includeReasoning && suggestion.Reasoning != "" {
		formatted += fmt.Sprintf(" - %s", suggestion.Reasoning)
	}

	return formatted
}
