package navigation

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"claude-wm-cli/internal/workflow"
)

func TestSuggestionEngine_GenerateSuggestions_NotInitialized(t *testing.T) {
	engine := NewSuggestionEngine()
	ctx := &ProjectContext{
		State: StateNotInitialized,
	}
	
	suggestions, err := engine.GenerateSuggestions(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, suggestions)
	
	// Should suggest init-project as top priority
	topSuggestion := suggestions[0]
	assert.Equal(t, "init-project", topSuggestion.Action.ID)
	assert.Equal(t, workflow.PriorityP0, topSuggestion.Priority)
	assert.Contains(t, topSuggestion.Reasoning, "not initialized")
}

func TestSuggestionEngine_GenerateSuggestions_ProjectInitialized(t *testing.T) {
	engine := NewSuggestionEngine()
	ctx := &ProjectContext{
		State: StateProjectInitialized,
	}
	
	suggestions, err := engine.GenerateSuggestions(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, suggestions)
	
	// Should suggest create-epic as top priority
	topSuggestion := suggestions[0]
	assert.Equal(t, "create-epic", topSuggestion.Action.ID)
	assert.Equal(t, workflow.PriorityP0, topSuggestion.Priority)
}

func TestSuggestionEngine_GenerateSuggestions_HasEpics(t *testing.T) {
	engine := NewSuggestionEngine()
	ctx := &ProjectContext{
		State: StateHasEpics,
	}
	
	suggestions, err := engine.GenerateSuggestions(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, suggestions)
	
	// Should suggest start-epic as top priority
	topSuggestion := suggestions[0]
	assert.Equal(t, "start-epic", topSuggestion.Action.ID)
	assert.Equal(t, workflow.PriorityP0, topSuggestion.Priority)
}

func TestSuggestionEngine_GenerateSuggestions_EpicInProgress(t *testing.T) {
	engine := NewSuggestionEngine()
	ctx := &ProjectContext{
		State: StateEpicInProgress,
		CurrentEpic: &EpicContext{
			ID:       "EPIC-001",
			Title:    "Test Epic",
			Progress: 0.5,
		},
	}
	
	suggestions, err := engine.GenerateSuggestions(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, suggestions)
	
	// Should suggest continue-epic since no story is active
	topSuggestion := suggestions[0]
	assert.Equal(t, "continue-epic", topSuggestion.Action.ID)
	assert.Equal(t, workflow.PriorityP1, topSuggestion.Priority)
	assert.Contains(t, topSuggestion.Reasoning, "Test Epic")
}

func TestSuggestionEngine_GenerateSuggestions_EpicInProgressWithStory(t *testing.T) {
	engine := NewSuggestionEngine()
	ctx := &ProjectContext{
		State: StateEpicInProgress,
		CurrentEpic: &EpicContext{
			ID:       "EPIC-001",
			Title:    "Test Epic",
			Progress: 0.5,
		},
		CurrentStory: &StoryContext{
			ID:       "STORY-001",
			Title:    "Test Story",
			Progress: 0.3,
		},
	}
	
	suggestions, err := engine.GenerateSuggestions(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, suggestions)
	
	// Should suggest continue-story since story is active
	topSuggestion := suggestions[0]
	assert.Equal(t, "continue-story", topSuggestion.Action.ID)
	assert.Contains(t, topSuggestion.Reasoning, "Test Story")
}

func TestSuggestionEngine_GenerateSuggestions_StoryInProgress(t *testing.T) {
	engine := NewSuggestionEngine()
	ctx := &ProjectContext{
		State: StateStoryInProgress,
		CurrentStory: &StoryContext{
			ID:       "STORY-001",
			Title:    "Test Story",
			Progress: 0.5,
		},
	}
	
	suggestions, err := engine.GenerateSuggestions(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, suggestions)
	
	// Should suggest continue-story since no task is active
	topSuggestion := suggestions[0]
	assert.Equal(t, "continue-story", topSuggestion.Action.ID)
}

func TestSuggestionEngine_GenerateSuggestions_TaskInProgress(t *testing.T) {
	engine := NewSuggestionEngine()
	ctx := &ProjectContext{
		State: StateTaskInProgress,
		CurrentTask: &TaskContext{
			ID:    "TASK-001",
			Title: "Test Task",
		},
	}
	
	suggestions, err := engine.GenerateSuggestions(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, suggestions)
	
	// Should suggest continue-task
	topSuggestion := suggestions[0]
	assert.Equal(t, "continue-task", topSuggestion.Action.ID)
	assert.Contains(t, topSuggestion.Reasoning, "Test Task")
}

func TestSuggestionEngine_GenerateSuggestions_NearCompletion(t *testing.T) {
	engine := NewSuggestionEngine()
	ctx := &ProjectContext{
		State: StateEpicInProgress,
		CurrentEpic: &EpicContext{
			ID:       "EPIC-001",
			Title:    "Nearly Done Epic",
			Progress: 0.9, // 90% complete
		},
	}
	
	suggestions, err := engine.GenerateSuggestions(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, suggestions)
	
	// Should include complete-epic suggestion
	var hasCompleteEpic bool
	for _, suggestion := range suggestions {
		if suggestion.Action.ID == "complete-epic" {
			hasCompleteEpic = true
			assert.Contains(t, suggestion.Reasoning, "90%")
			break
		}
	}
	assert.True(t, hasCompleteEpic, "Should suggest completing nearly done epic")
}

func TestSuggestionEngine_GenerateSuggestions_WithIssues(t *testing.T) {
	engine := NewSuggestionEngine()
	ctx := &ProjectContext{
		State:  StateProjectInitialized,
		Issues: []string{"Missing configuration", "Corrupted state file"},
	}
	
	suggestions, err := engine.GenerateSuggestions(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, suggestions)
	
	// Should include fix-issues suggestion
	var hasFixIssues bool
	for _, suggestion := range suggestions {
		if suggestion.Action.ID == "fix-issues" {
			hasFixIssues = true
			assert.Contains(t, suggestion.Reasoning, "2 project issues")
			break
		}
	}
	assert.True(t, hasFixIssues, "Should suggest fixing issues when issues exist")
}

func TestSuggestionEngine_SortSuggestions(t *testing.T) {
	engine := NewSuggestionEngine()
	
	suggestions := []*Suggestion{
		{
			Action:   &workflow.WorkflowAction{ID: "low-priority"},
			Priority: workflow.PriorityP2,
			Urgency:  1,
		},
		{
			Action:   &workflow.WorkflowAction{ID: "high-priority-low-urgency"},
			Priority: workflow.PriorityP0,
			Urgency:  1,
		},
		{
			Action:   &workflow.WorkflowAction{ID: "high-priority-high-urgency"},
			Priority: workflow.PriorityP0,
			Urgency:  10,
		},
		{
			Action:   &workflow.WorkflowAction{ID: "medium-priority"},
			Priority: workflow.PriorityP1,
			Urgency:  5,
		},
	}
	
	engine.sortSuggestions(suggestions)
	
	// Should be sorted by priority first, then urgency
	assert.Equal(t, "high-priority-high-urgency", suggestions[0].Action.ID)
	assert.Equal(t, "high-priority-low-urgency", suggestions[1].Action.ID)
	assert.Equal(t, "medium-priority", suggestions[2].Action.ID)
	assert.Equal(t, "low-priority", suggestions[3].Action.ID)
}

func TestSuggestionEngine_GetTopSuggestion(t *testing.T) {
	engine := NewSuggestionEngine()
	ctx := &ProjectContext{
		State: StateNotInitialized,
	}
	
	suggestion, err := engine.GetTopSuggestion(ctx)
	require.NoError(t, err)
	require.NotNil(t, suggestion)
	
	assert.Equal(t, "init-project", suggestion.Action.ID)
	assert.Equal(t, workflow.PriorityP0, suggestion.Priority)
}

func TestSuggestionEngine_GetSuggestionsByPriority(t *testing.T) {
	engine := NewSuggestionEngine()
	ctx := &ProjectContext{
		State: StateProjectInitialized,
	}
	
	grouped, err := engine.GetSuggestionsByPriority(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, grouped)
	
	// Should have P0 suggestions
	p0Suggestions := grouped[workflow.PriorityP0]
	assert.NotEmpty(t, p0Suggestions)
	assert.Equal(t, "create-epic", p0Suggestions[0].Action.ID)
	
	// Should have P2 suggestions (help, status, etc.)
	p2Suggestions := grouped[workflow.PriorityP2]
	assert.NotEmpty(t, p2Suggestions)
}

func TestSuggestionEngine_FilterDuplicates(t *testing.T) {
	engine := NewSuggestionEngine()
	
	suggestions := []*Suggestion{
		{
			Action: &workflow.WorkflowAction{ID: "help"},
		},
		{
			Action: &workflow.WorkflowAction{ID: "status"},
		},
		{
			Action: &workflow.WorkflowAction{ID: "help"}, // Duplicate
		},
	}
	
	ctx := &ProjectContext{State: StateProjectInitialized}
	filtered := engine.filterSuggestions(suggestions, ctx)
	
	// Should remove duplicate
	assert.Len(t, filtered, 2)
	
	ids := make(map[string]bool)
	for _, suggestion := range filtered {
		ids[suggestion.Action.ID] = true
	}
	
	assert.True(t, ids["help"])
	assert.True(t, ids["status"])
}

func TestSuggestionEngine_FormatSuggestion(t *testing.T) {
	engine := NewSuggestionEngine()
	
	suggestion := &Suggestion{
		Action: &workflow.WorkflowAction{
			ID:   "test-action",
			Name: "Test Action",
		},
		Priority:  workflow.PriorityP1,
		Reasoning: "This is a test suggestion",
	}
	
	// Format without reasoning
	formatted := engine.FormatSuggestion(suggestion, false)
	assert.Equal(t, "[P1] Test Action", formatted)
	
	// Format with reasoning
	formatted = engine.FormatSuggestion(suggestion, true)
	assert.Equal(t, "[P1] Test Action - This is a test suggestion", formatted)
}

func TestSuggestionEngine_FormatSuggestion_Nil(t *testing.T) {
	engine := NewSuggestionEngine()
	
	// Test nil suggestion
	formatted := engine.FormatSuggestion(nil, true)
	assert.Equal(t, "No suggestion available", formatted)
	
	// Test suggestion with nil action
	suggestion := &Suggestion{Action: nil}
	formatted = engine.FormatSuggestion(suggestion, true)
	assert.Equal(t, "No suggestion available", formatted)
}

func TestSuggestionEngine_GenerateSuggestions_NilContext(t *testing.T) {
	engine := NewSuggestionEngine()
	
	suggestions, err := engine.GenerateSuggestions(nil)
	assert.Error(t, err)
	assert.Nil(t, suggestions)
	assert.Contains(t, err.Error(), "context is nil")
}

func TestSuggestionEngine_EmptyEpicSuggestsCreateStory(t *testing.T) {
	engine := NewSuggestionEngine()
	ctx := &ProjectContext{
		State: StateEpicInProgress,
		CurrentEpic: &EpicContext{
			ID:           "EPIC-001",
			Title:        "Empty Epic",
			TotalStories: 0, // No stories yet
		},
	}
	
	suggestions, err := engine.GenerateSuggestions(ctx)
	require.NoError(t, err)
	
	// Should suggest creating a story
	var hasCreateStory bool
	for _, suggestion := range suggestions {
		if suggestion.Action.ID == "create-story" {
			hasCreateStory = true
			assert.Contains(t, suggestion.Reasoning, "no stories")
			break
		}
	}
	assert.True(t, hasCreateStory, "Should suggest creating story for empty epic")
}

func TestSuggestionEngine_EmptyStorySuggestsCreateTask(t *testing.T) {
	engine := NewSuggestionEngine()
	ctx := &ProjectContext{
		State: StateStoryInProgress,
		CurrentStory: &StoryContext{
			ID:         "STORY-001",
			Title:      "Empty Story",
			TotalTasks: 0, // No tasks yet
		},
	}
	
	suggestions, err := engine.GenerateSuggestions(ctx)
	require.NoError(t, err)
	
	// Should suggest creating a task
	var hasCreateTask bool
	for _, suggestion := range suggestions {
		if suggestion.Action.ID == "create-task" {
			hasCreateTask = true
			assert.Contains(t, suggestion.Reasoning, "no tasks")
			break
		}
	}
	assert.True(t, hasCreateTask, "Should suggest creating task for empty story")
}