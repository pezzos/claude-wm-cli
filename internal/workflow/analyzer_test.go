package workflow

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"claude-wm-cli/internal/state"
)

func TestWorkflowPosition_String(t *testing.T) {
	tests := []struct {
		position WorkflowPosition
		expected string
	}{
		{PositionUnknown, "unknown"},
		{PositionNotInitialized, "not_initialized"},
		{PositionProjectLevel, "project"},
		{PositionEpicLevel, "epic"},
		{PositionStoryLevel, "story"},
		{PositionTaskLevel, "task"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.position.String())
		})
	}
}

func TestNewWorkflowAnalyzer(t *testing.T) {
	rootPath := "/test/path"
	analyzer := NewWorkflowAnalyzer(rootPath)

	assert.NotNil(t, analyzer)
	assert.Equal(t, rootPath, analyzer.rootPath)
}

func TestAnalyzeWorkflowPosition_NotInitialized(t *testing.T) {
	tempDir := t.TempDir()
	analyzer := NewWorkflowAnalyzer(tempDir)

	analysis, err := analyzer.AnalyzeWorkflowPosition()
	require.NoError(t, err)

	assert.Equal(t, PositionNotInitialized, analysis.Position)
	assert.False(t, analysis.ProjectInitialized)
	assert.Contains(t, analysis.Recommendations, "Initialize project structure")
	assert.Equal(t, tempDir, analysis.RootPath)
	assert.WithinDuration(t, time.Now(), analysis.AnalyzedAt, time.Second)
}

func TestAnalyzeWorkflowPosition_ProjectLevel(t *testing.T) {
	tempDir := t.TempDir()
	setupCompleteProjectStructure(t, tempDir)

	analyzer := NewWorkflowAnalyzer(tempDir)
	analysis, err := analyzer.AnalyzeWorkflowPosition()
	require.NoError(t, err)

	assert.Equal(t, PositionProjectLevel, analysis.Position)
	assert.True(t, analysis.ProjectInitialized)
	assert.Nil(t, analysis.CurrentEpic)
	assert.Nil(t, analysis.CurrentStory)
	assert.Empty(t, analysis.CurrentTasks)
	assert.Contains(t, analysis.Recommendations, "Create your first epic to start organizing work")
}

func TestAnalyzeWorkflowPosition_EpicLevel(t *testing.T) {
	tempDir := t.TempDir()
	setupCompleteProjectStructure(t, tempDir)
	setupCurrentEpic(t, tempDir)

	analyzer := NewWorkflowAnalyzer(tempDir)
	analysis, err := analyzer.AnalyzeWorkflowPosition()
	require.NoError(t, err)

	assert.Equal(t, PositionEpicLevel, analysis.Position)
	assert.True(t, analysis.ProjectInitialized)
	assert.NotNil(t, analysis.CurrentEpic)
	assert.Equal(t, "EPIC-001", analysis.CurrentEpic.Metadata.ID)
	assert.Nil(t, analysis.CurrentStory)
	assert.Contains(t, analysis.Recommendations, "Break down the epic into user stories")
}

func TestAnalyzeWorkflowPosition_StoryLevel(t *testing.T) {
	tempDir := t.TempDir()
	setupCompleteProjectStructure(t, tempDir)
	setupCurrentEpic(t, tempDir)
	setupCurrentStory(t, tempDir)

	analyzer := NewWorkflowAnalyzer(tempDir)
	analysis, err := analyzer.AnalyzeWorkflowPosition()
	require.NoError(t, err)

	assert.Equal(t, PositionStoryLevel, analysis.Position)
	assert.True(t, analysis.ProjectInitialized)
	assert.NotNil(t, analysis.CurrentEpic)
	assert.NotNil(t, analysis.CurrentStory)
	assert.Equal(t, "STORY-001", analysis.CurrentStory.Metadata.ID)
	assert.Empty(t, analysis.CurrentTasks)
	assert.Contains(t, analysis.Recommendations, "Create tasks to implement the current story")
}

func TestAnalyzeWorkflowPosition_TaskLevel(t *testing.T) {
	tempDir := t.TempDir()
	setupCompleteProjectStructure(t, tempDir)
	setupCurrentEpic(t, tempDir)
	setupCurrentStory(t, tempDir)
	setupCurrentTasks(t, tempDir)

	analyzer := NewWorkflowAnalyzer(tempDir)
	analysis, err := analyzer.AnalyzeWorkflowPosition()
	require.NoError(t, err)

	assert.Equal(t, PositionTaskLevel, analysis.Position)
	assert.True(t, analysis.ProjectInitialized)
	assert.NotNil(t, analysis.CurrentEpic)
	assert.NotNil(t, analysis.CurrentStory)
	assert.NotEmpty(t, analysis.CurrentTasks)
	assert.Contains(t, analysis.Recommendations, "Start working on the next task")
}

func TestAnalyzeWorkflowPosition_WithBlockers(t *testing.T) {
	tempDir := t.TempDir()
	setupCompleteProjectStructure(t, tempDir)
	setupCurrentEpic(t, tempDir)
	// Don't setup current story - this should create a blocker

	analyzer := NewWorkflowAnalyzer(tempDir)
	analysis, err := analyzer.AnalyzeWorkflowPosition()
	require.NoError(t, err)

	assert.Len(t, analysis.Blockers, 1)
	assert.Equal(t, BlockerMissingDefinition, analysis.Blockers[0].Type)
	assert.Contains(t, analysis.Blockers[0].Description, "Epic is selected but no stories are defined")
	assert.Equal(t, "EPIC-001", analysis.Blockers[0].Entity)
}

func TestCalculateCompletionMetrics(t *testing.T) {
	tempDir := t.TempDir()
	setupCompleteProjectStructure(t, tempDir)
	setupMultipleEpics(t, tempDir)
	setupCurrentEpic(t, tempDir)
	setupCurrentStory(t, tempDir)
	setupCurrentTasks(t, tempDir)

	analyzer := NewWorkflowAnalyzer(tempDir)
	analysis, err := analyzer.AnalyzeWorkflowPosition()
	require.NoError(t, err)

	metrics := analysis.CompletionMetrics
	assert.Equal(t, 3, metrics.TotalEpics)     // From setupMultipleEpics
	assert.Equal(t, 1, metrics.CompletedEpics) // One epic marked as done
	assert.InDelta(t, 33.33, metrics.ProjectProgress, 0.1)

	assert.Equal(t, 2, metrics.TotalStories)     // From setupCurrentStory
	assert.Equal(t, 1, metrics.CompletedStories) // One story marked as done
	assert.Equal(t, 50.0, metrics.EpicProgress)

	assert.Equal(t, 2, metrics.TotalTasks)     // From setupCurrentTasks
	assert.Equal(t, 0, metrics.CompletedTasks) // No tasks marked as done
	assert.Equal(t, 0.0, metrics.StoryProgress)
}

func TestDetectBlockers_InconsistentState(t *testing.T) {
	tempDir := t.TempDir()
	setupCompleteProjectStructure(t, tempDir)
	setupCompletedEpicWithActiveStory(t, tempDir)

	analyzer := NewWorkflowAnalyzer(tempDir)
	analysis, err := analyzer.AnalyzeWorkflowPosition()
	require.NoError(t, err)

	// Should detect inconsistent state: epic is done but story is in progress
	hasInconsistentStateBlocker := false
	for _, blocker := range analysis.Blockers {
		if blocker.Type == BlockerInconsistentState {
			hasInconsistentStateBlocker = true
			assert.Equal(t, "critical", blocker.Severity)
			assert.Contains(t, blocker.Description, "Epic is marked as done but work is still in progress")
		}
	}
	assert.True(t, hasInconsistentStateBlocker, "Should detect inconsistent state blocker")
}

func TestDetectBlockers_BlockedTasks(t *testing.T) {
	tempDir := t.TempDir()
	setupCompleteProjectStructure(t, tempDir)
	setupCurrentEpic(t, tempDir)
	setupCurrentStory(t, tempDir)
	setupBlockedTasks(t, tempDir)

	analyzer := NewWorkflowAnalyzer(tempDir)
	analysis, err := analyzer.AnalyzeWorkflowPosition()
	require.NoError(t, err)

	// Should detect blocked task
	hasBlockedTaskBlocker := false
	for _, blocker := range analysis.Blockers {
		if blocker.Type == BlockerMissingDependency {
			hasBlockedTaskBlocker = true
			assert.Equal(t, "high", blocker.Severity)
			assert.Contains(t, blocker.Description, "is blocked")
		}
	}
	assert.True(t, hasBlockedTaskBlocker, "Should detect blocked task blocker")
}

func TestGetWorkflowCapabilities(t *testing.T) {
	tests := []struct {
		name                 string
		position             WorkflowPosition
		expectedCapabilities []string
		mustContain          []string
	}{
		{
			name:        "not initialized",
			position:    PositionNotInitialized,
			mustContain: []string{"init-project", "help"},
		},
		{
			name:        "project level",
			position:    PositionProjectLevel,
			mustContain: []string{"create-epic", "list-epics", "status", "help"},
		},
		{
			name:        "epic level",
			position:    PositionEpicLevel,
			mustContain: []string{"create-story", "start-epic", "list-stories", "help"},
		},
		{
			name:        "story level",
			position:    PositionStoryLevel,
			mustContain: []string{"create-task", "continue-story", "list-tasks", "help"},
		},
		{
			name:        "task level",
			position:    PositionTaskLevel,
			mustContain: []string{"continue-task", "complete-task", "create-task", "help"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer := NewWorkflowAnalyzer("/test")
			analysis := &WorkflowAnalysis{Position: tt.position}

			capabilities := analyzer.GetWorkflowCapabilities(analysis)

			for _, expectedCap := range tt.mustContain {
				assert.Contains(t, capabilities, expectedCap,
					"Capabilities should contain %s for position %s", expectedCap, tt.position)
			}
		})
	}
}

func TestLoadCurrentTasks_TodoFormat(t *testing.T) {
	tempDir := t.TempDir()
	setupCompleteProjectStructure(t, tempDir)

	// Create a todo file in the expected format
	todoData := map[string]interface{}{
		"todos": []map[string]interface{}{
			{
				"id":       "TASK-001",
				"title":    "Test Task 1",
				"status":   "todo",
				"priority": "P0",
			},
			{
				"id":       "TASK-002",
				"title":    "Test Task 2",
				"status":   "in_progress",
				"priority": "P1",
			},
			{
				"id":       "TASK-003",
				"title":    "Test Task 3",
				"status":   "completed",
				"priority": "P2",
			},
		},
	}

	todoJSON, _ := json.Marshal(todoData)
	todoPath := filepath.Join(tempDir, "docs/3-current-task/todo.json")
	os.WriteFile(todoPath, todoJSON, 0644)

	analyzer := NewWorkflowAnalyzer(tempDir)
	tasks, err := analyzer.loadCurrentTasks()
	require.NoError(t, err)

	// Should only load todo and in_progress tasks, not completed ones
	assert.Len(t, tasks, 2)
	assert.Equal(t, "TASK-001", tasks[0].Metadata.ID)
	assert.Equal(t, "Test Task 1", tasks[0].Title)
	assert.Equal(t, state.StatusTodo, tasks[0].Status)
	assert.Equal(t, "TASK-002", tasks[1].Metadata.ID)
	assert.Equal(t, state.StatusInProgress, tasks[1].Status)
}

// Helper functions for setting up test data

func setupCompleteProjectStructure(t *testing.T, tempDir string) {
	// Create directory structure
	dirs := []string{
		"docs/1-project",
		"docs/2-current-epic",
		"docs/3-current-task",
	}
	for _, dir := range dirs {
		err := os.MkdirAll(filepath.Join(tempDir, dir), 0755)
		require.NoError(t, err)
	}

	// Create basic epics.json
	epicsData := map[string]interface{}{
		"epics": []interface{}{},
	}
	epicsJSON, _ := json.Marshal(epicsData)
	epicsPath := filepath.Join(tempDir, "docs/1-project/epics.json")
	err := os.WriteFile(epicsPath, epicsJSON, 0644)
	require.NoError(t, err)
}

func setupCurrentEpic(t *testing.T, tempDir string) {
	epicData := state.EpicState{
		Metadata: state.Metadata{
			ID:            "EPIC-001",
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
			SchemaVersion: "1.0.0",
		},
		Title:       "Test Epic",
		Description: "A test epic for workflow analysis",
		Priority:    state.PriorityP0,
		Status:      state.StatusInProgress,
	}

	epicJSON, _ := json.Marshal(epicData)
	epicPath := filepath.Join(tempDir, "docs/2-current-epic/current-epic.json")
	err := os.WriteFile(epicPath, epicJSON, 0644)
	require.NoError(t, err)
}

func setupCurrentStory(t *testing.T, tempDir string) {
	storiesData := map[string]interface{}{
		"meta": map[string]interface{}{
			"current_story": "STORY-001",
		},
		"stories": []state.StoryState{
			{
				Metadata: state.Metadata{
					ID:            "STORY-001",
					CreatedAt:     time.Now(),
					UpdatedAt:     time.Now(),
					SchemaVersion: "1.0.0",
				},
				EpicID:      "EPIC-001",
				Title:       "Test Story 1",
				Description: "A test story for workflow analysis",
				Priority:    state.PriorityP0,
				Status:      state.StatusInProgress,
			},
			{
				Metadata: state.Metadata{
					ID:            "STORY-002",
					CreatedAt:     time.Now(),
					UpdatedAt:     time.Now(),
					SchemaVersion: "1.0.0",
				},
				EpicID:      "EPIC-001",
				Title:       "Test Story 2",
				Description: "Another test story",
				Priority:    state.PriorityP1,
				Status:      state.StatusDone,
			},
		},
	}

	storiesJSON, _ := json.Marshal(storiesData)
	storiesPath := filepath.Join(tempDir, "docs/2-current-epic/stories.json")
	err := os.WriteFile(storiesPath, storiesJSON, 0644)
	require.NoError(t, err)
}

func setupCurrentTasks(t *testing.T, tempDir string) {
	todoData := map[string]interface{}{
		"todos": []map[string]interface{}{
			{
				"id":       "TASK-001",
				"title":    "Test Task 1",
				"status":   "todo",
				"priority": "P0",
			},
			{
				"id":       "TASK-002",
				"title":    "Test Task 2",
				"status":   "todo",
				"priority": "P1",
			},
		},
	}

	todoJSON, _ := json.Marshal(todoData)
	todoPath := filepath.Join(tempDir, "docs/3-current-task/todo.json")
	err := os.WriteFile(todoPath, todoJSON, 0644)
	require.NoError(t, err)
}

func setupMultipleEpics(t *testing.T, tempDir string) {
	epicsData := map[string]interface{}{
		"epics": []state.EpicState{
			{
				Metadata: state.Metadata{
					ID:            "EPIC-001",
					CreatedAt:     time.Now(),
					UpdatedAt:     time.Now(),
					SchemaVersion: "1.0.0",
				},
				Title:          "Test Epic 1",
				Status:         state.StatusInProgress,
				EstimatedHours: 40.0,
				ActualHours:    20.0,
			},
			{
				Metadata: state.Metadata{
					ID:            "EPIC-002",
					CreatedAt:     time.Now(),
					UpdatedAt:     time.Now(),
					SchemaVersion: "1.0.0",
				},
				Title:          "Test Epic 2",
				Status:         state.StatusDone,
				EstimatedHours: 30.0,
				ActualHours:    35.0,
			},
			{
				Metadata: state.Metadata{
					ID:            "EPIC-003",
					CreatedAt:     time.Now(),
					UpdatedAt:     time.Now(),
					SchemaVersion: "1.0.0",
				},
				Title:          "Test Epic 3",
				Status:         state.StatusTodo,
				EstimatedHours: 50.0,
				ActualHours:    0.0,
			},
		},
	}

	epicsJSON, _ := json.Marshal(epicsData)
	epicsPath := filepath.Join(tempDir, "docs/1-project/epics.json")
	err := os.WriteFile(epicsPath, epicsJSON, 0644)
	require.NoError(t, err)
}

func setupCompletedEpicWithActiveStory(t *testing.T, tempDir string) {
	// Set up a completed epic
	epicData := state.EpicState{
		Metadata: state.Metadata{
			ID:            "EPIC-001",
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
			SchemaVersion: "1.0.0",
		},
		Title:       "Completed Epic",
		Description: "An epic that's marked as done",
		Priority:    state.PriorityP0,
		Status:      state.StatusDone, // This is the inconsistency
	}

	epicJSON, _ := json.Marshal(epicData)
	epicPath := filepath.Join(tempDir, "docs/2-current-epic/current-epic.json")
	err := os.WriteFile(epicPath, epicJSON, 0644)
	require.NoError(t, err)

	// Set up an active story (this creates the inconsistency)
	setupCurrentStory(t, tempDir)
}

func setupBlockedTasks(t *testing.T, tempDir string) {
	todoData := map[string]interface{}{
		"todos": []map[string]interface{}{
			{
				"id":       "TASK-001",
				"title":    "Blocked Task",
				"status":   "blocked",
				"priority": "P0",
			},
			{
				"id":       "TASK-002",
				"title":    "Normal Task",
				"status":   "todo",
				"priority": "P1",
			},
		},
	}

	todoJSON, _ := json.Marshal(todoData)
	todoPath := filepath.Join(tempDir, "docs/3-current-task/todo.json")
	err := os.WriteFile(todoPath, todoJSON, 0644)
	require.NoError(t, err)
}
