package navigation

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkflowState_String(t *testing.T) {
	tests := []struct {
		state    WorkflowState
		expected string
	}{
		{StateNotInitialized, "Not Initialized"},
		{StateProjectInitialized, "Project Initialized"},
		{StateHasEpics, "Has Epics"},
		{StateEpicInProgress, "Epic In Progress"},
		{StateStoryInProgress, "Story In Progress"},
		{StateTaskInProgress, "Task In Progress"},
		{WorkflowState(999), "Unknown State"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.state.String())
		})
	}
}

func TestContextDetector_DetectContext_NotInitialized(t *testing.T) {
	// Create temporary directory without docs structure
	tempDir := t.TempDir()

	detector := NewContextDetector(tempDir)
	ctx, err := detector.DetectContext()

	require.NoError(t, err)
	assert.Equal(t, StateNotInitialized, ctx.State)
	assert.Contains(t, ctx.AvailableActions, "init-project")
	assert.Equal(t, tempDir, ctx.ProjectPath)
}

func TestContextDetector_DetectContext_ProjectInitialized(t *testing.T) {
	// Create temporary directory with docs structure but no epics
	tempDir := t.TempDir()
	createProjectStructure(t, tempDir)

	detector := NewContextDetector(tempDir)
	ctx, err := detector.DetectContext()

	require.NoError(t, err)
	assert.Equal(t, StateProjectInitialized, ctx.State)
	assert.Contains(t, ctx.AvailableActions, "create-epic")
}

func TestContextDetector_DetectContext_HasEpics(t *testing.T) {
	// Create temporary directory with epics.json but no current epic
	tempDir := t.TempDir()
	createProjectStructure(t, tempDir)
	createEpicsFile(t, tempDir, false)

	detector := NewContextDetector(tempDir)
	ctx, err := detector.DetectContext()

	require.NoError(t, err)
	assert.Equal(t, StateHasEpics, ctx.State)
	assert.Contains(t, ctx.AvailableActions, "start-epic")
}

func TestContextDetector_DetectContext_EpicInProgress(t *testing.T) {
	// Create temporary directory with current epic
	tempDir := t.TempDir()
	createProjectStructure(t, tempDir)
	createEpicsFile(t, tempDir, true)
	createCurrentEpicFile(t, tempDir)

	detector := NewContextDetector(tempDir)
	ctx, err := detector.DetectContext()

	require.NoError(t, err)
	assert.Equal(t, StateEpicInProgress, ctx.State)
	assert.NotNil(t, ctx.CurrentEpic)
	assert.Equal(t, "EPIC-001", ctx.CurrentEpic.ID)
	assert.Equal(t, "Test Epic", ctx.CurrentEpic.Title)
	assert.Contains(t, ctx.AvailableActions, "continue-epic")
}

func TestContextDetector_DetectContext_StoryInProgress(t *testing.T) {
	// Create temporary directory with current story
	tempDir := t.TempDir()
	createProjectStructure(t, tempDir)
	createEpicsFile(t, tempDir, true)
	createCurrentEpicFile(t, tempDir)
	createStoriesFile(t, tempDir)

	detector := NewContextDetector(tempDir)
	ctx, err := detector.DetectContext()

	require.NoError(t, err)
	assert.Equal(t, StateStoryInProgress, ctx.State)
	assert.NotNil(t, ctx.CurrentStory)
	assert.Equal(t, "STORY-001", ctx.CurrentStory.ID)
	assert.Contains(t, ctx.AvailableActions, "continue-story")
}

func TestContextDetector_DetectContext_TaskInProgress(t *testing.T) {
	// Create temporary directory with current task
	tempDir := t.TempDir()
	createProjectStructure(t, tempDir)
	createEpicsFile(t, tempDir, true)
	createCurrentEpicFile(t, tempDir)
	createStoriesFile(t, tempDir)
	createTodoFile(t, tempDir)

	detector := NewContextDetector(tempDir)
	ctx, err := detector.DetectContext()

	require.NoError(t, err)
	assert.Equal(t, StateTaskInProgress, ctx.State)
	assert.NotNil(t, ctx.CurrentTask)
	assert.Equal(t, "TASK-001", ctx.CurrentTask.ID)
	assert.Contains(t, ctx.AvailableActions, "continue-task")
}

func TestContextDetector_GetRecommendedAction(t *testing.T) {
	detector := NewContextDetector("")

	tests := []struct {
		state    WorkflowState
		expected string
	}{
		{StateNotInitialized, "init-project"},
		{StateProjectInitialized, "create-epic"},
		{StateHasEpics, "start-epic"},
		{StateEpicInProgress, "continue-epic"},
		{StateStoryInProgress, "continue-story"},
		{StateTaskInProgress, "continue-task"},
	}

	for _, tt := range tests {
		t.Run(tt.state.String(), func(t *testing.T) {
			ctx := &ProjectContext{State: tt.state}
			action := detector.GetRecommendedAction(ctx)
			assert.Equal(t, tt.expected, action)
		})
	}
}

func TestContextDetector_HandleCorruptedFiles(t *testing.T) {
	// Create temporary directory with corrupted JSON files
	tempDir := t.TempDir()
	createProjectStructure(t, tempDir)

	// Create corrupted epics.json
	epicsPath := filepath.Join(tempDir, "docs/1-project/epics.json")
	err := os.WriteFile(epicsPath, []byte("invalid json"), 0644)
	require.NoError(t, err)

	detector := NewContextDetector(tempDir)
	ctx, err := detector.DetectContext()

	require.NoError(t, err)
	assert.Equal(t, StateHasEpics, ctx.State) // Should still detect epics.json exists
	assert.NotEmpty(t, ctx.Issues)            // Should report issues
}

// Helper functions for tests

func createProjectStructure(t *testing.T, tempDir string) {
	dirs := []string{
		"docs/1-project",
		"docs/2-current-epic",
		"docs/3-current-task",
	}

	for _, dir := range dirs {
		err := os.MkdirAll(filepath.Join(tempDir, dir), 0755)
		require.NoError(t, err)
	}
}

func createEpicsFile(t *testing.T, tempDir string, withCurrentEpic bool) {
	epicsData := map[string]interface{}{
		"epics": []map[string]interface{}{
			{
				"id":     "EPIC-001",
				"title":  "Test Epic",
				"status": "todo",
			},
		},
	}

	if withCurrentEpic {
		epicsData["epics"].([]map[string]interface{})[0]["status"] = "in_progress"
	}

	data, err := json.Marshal(epicsData)
	require.NoError(t, err)

	epicsPath := filepath.Join(tempDir, "docs/1-project/epics.json")
	err = os.WriteFile(epicsPath, data, 0644)
	require.NoError(t, err)
}

func createCurrentEpicFile(t *testing.T, tempDir string) {
	epicData := map[string]interface{}{
		"epic": map[string]interface{}{
			"id":       "EPIC-001",
			"title":    "Test Epic",
			"status":   "🚧 In Progress",
			"priority": "high",
			"userStories": []map[string]interface{}{
				{
					"id":     "US-001",
					"title":  "Test Story",
					"status": "completed",
				},
				{
					"id":     "US-002",
					"title":  "Another Story",
					"status": "todo",
				},
			},
		},
	}

	data, err := json.Marshal(epicData)
	require.NoError(t, err)

	epicPath := filepath.Join(tempDir, "docs/2-current-epic/current-epic.json")
	err = os.WriteFile(epicPath, data, 0644)
	require.NoError(t, err)
}

func createStoriesFile(t *testing.T, tempDir string) {
	storiesData := map[string]interface{}{
		"stories": []map[string]interface{}{
			{
				"id":       "STORY-001",
				"title":    "Test Story",
				"status":   "in_progress",
				"priority": "high",
			},
		},
	}

	data, err := json.Marshal(storiesData)
	require.NoError(t, err)

	storiesPath := filepath.Join(tempDir, "docs/2-current-epic/stories.json")
	err = os.WriteFile(storiesPath, data, 0644)
	require.NoError(t, err)
}

func createTodoFile(t *testing.T, tempDir string) {
	todoData := map[string]interface{}{
		"todos": []map[string]interface{}{
			{
				"id":             "TASK-001",
				"title":          "Test Task",
				"status":         "in_progress",
				"priority":       "P0",
				"estimatedHours": 3,
			},
		},
	}

	data, err := json.Marshal(todoData)
	require.NoError(t, err)

	todoPath := filepath.Join(tempDir, "docs/3-current-task/todo-epic-001.json")
	err = os.WriteFile(todoPath, data, 0644)
	require.NoError(t, err)
}
