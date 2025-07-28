package navigation

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"claude-wm-cli/internal/workflow"
)

// TestNavigationSystemIntegration tests the complete navigation workflow
func TestNavigationSystemIntegration(t *testing.T) {
	tempDir := t.TempDir()
	
	// Test progression through complete workflow states
	testCases := []struct {
		name           string
		setupFunc      func(string) error
		expectedState  WorkflowState
		expectedAction string
	}{
		{
			name:           "not_initialized",
			setupFunc:      setupEmptyProject,
			expectedState:  StateNotInitialized,
			expectedAction: "init-project",
		},
		{
			name:           "project_initialized",
			setupFunc:      setupProjectStructure,
			expectedState:  StateProjectInitialized,
			expectedAction: "create-epic",
		},
		{
			name:           "has_epics",
			setupFunc:      setupWithEpics,
			expectedState:  StateHasEpics,
			expectedAction: "start-epic",
		},
		{
			name:           "epic_in_progress",
			setupFunc:      setupEpicInProgress,
			expectedState:  StateEpicInProgress,
			expectedAction: "continue-epic",
		},
		{
			name:           "story_in_progress",
			setupFunc:      setupStoryInProgress,
			expectedState:  StateStoryInProgress,
			expectedAction: "continue-story",
		},
		{
			name:           "task_in_progress",
			setupFunc:      setupTaskInProgress,
			expectedState:  StateTaskInProgress,
			expectedAction: "continue-task",
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup project state
			projectDir := filepath.Join(tempDir, tc.name)
			err := os.MkdirAll(projectDir, 0755)
			require.NoError(t, err)
			
			err = tc.setupFunc(projectDir)
			require.NoError(t, err)
			
			// Test context detection
			detector := NewContextDetector(projectDir)
			ctx, err := detector.DetectContext()
			require.NoError(t, err)
			assert.Equal(t, tc.expectedState, ctx.State)
			
			// Test suggestion generation
			engine := NewSuggestionEngine()
			suggestions, err := engine.GenerateSuggestions(ctx)
			require.NoError(t, err)
			require.NotEmpty(t, suggestions)
			
			// Check top suggestion
			topSuggestion := suggestions[0]
			assert.Equal(t, tc.expectedAction, topSuggestion.Action.ID)
			
			// Test display system
			display := NewProjectStateDisplay()
			assert.NotPanics(t, func() {
				display.DisplayProjectOverview(ctx)
			})
		})
	}
}

// TestNavigationErrorHandling tests error scenarios
func TestNavigationErrorHandling(t *testing.T) {
	tempDir := t.TempDir()
	
	testCases := []struct {
		name      string
		setupFunc func(string) error
		expectErr bool
		hasIssues bool
	}{
		{
			name:      "corrupted_epic_file",
			setupFunc: setupCorruptedEpic,
			expectErr: false, // Should handle gracefully
			hasIssues: true,
		},
		{
			name:      "missing_current_epic",
			setupFunc: setupMissingCurrentEpic,
			expectErr: false,
			hasIssues: false,
		},
		{
			name:      "permission_denied",
			setupFunc: setupPermissionDenied,
			expectErr: false, // Context detection should handle this
			hasIssues: false,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			projectDir := filepath.Join(tempDir, tc.name)
			err := os.MkdirAll(projectDir, 0755)
			require.NoError(t, err)
			
			err = tc.setupFunc(projectDir)
			if tc.expectErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			
			detector := NewContextDetector(projectDir)
			ctx, err := detector.DetectContext()
			
			if tc.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tc.hasIssues {
					assert.NotEmpty(t, ctx.Issues)
				}
			}
		})
	}
}

// TestNavigationPerformance tests performance with large files
func TestNavigationPerformance(t *testing.T) {
	tempDir := t.TempDir()
	
	// Create large epic file
	err := setupLargeEpicFile(tempDir)
	require.NoError(t, err)
	
	detector := NewContextDetector(tempDir)
	ctx, err := detector.DetectContext()
	require.NoError(t, err)
	
	// Should handle large files without issues
	assert.Equal(t, StateHasEpics, ctx.State)
	
	// Test suggestion generation with large context
	engine := NewSuggestionEngine()
	suggestions, err := engine.GenerateSuggestions(ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, suggestions)
}

// TestMenuInteractionSimulation simulates user interactions
func TestMenuInteractionSimulation(t *testing.T) {
	tempDir := t.TempDir()
	setupProjectStructure(tempDir)
	
	// Test context detection works
	detector := NewContextDetector(tempDir)
	_, err := detector.DetectContext()
	require.NoError(t, err)
	
	// Test menu creation
	menu := &Menu{
		Title:       "Test Menu",
		ShowNumbers: true,
		AllowBack:   true,
		AllowQuit:   true,
		Options: []MenuOption{
			{
				ID:      "test1",
				Label:   "Test Option 1",
				Action:  "action1",
				Enabled: true,
			},
			{
				ID:      "test2",
				Label:   "Test Option 2",
				Action:  "action2",
				Enabled: true,
			},
		},
	}
	
	display := &MenuDisplay{}
	
	// Test different input scenarios
	testInputs := []struct {
		input          string
		expectedAction string
	}{
		{"1", "action1"},
		{"2", "action2"},
		{"test1", "action1"},
		{"test2", "action2"},
		{"q", "quit"},
		{"quit", "quit"},
		{"b", "back"},
		{"back", "back"},
	}
	
	for _, testInput := range testInputs {
		t.Run("input_"+testInput.input, func(t *testing.T) {
			result := display.processInput(menu, testInput.input)
			if testInput.expectedAction == "quit" || testInput.expectedAction == "back" {
				require.NotNil(t, result)
				assert.Equal(t, testInput.expectedAction, result.Action)
			} else {
				require.NotNil(t, result)
				assert.Equal(t, testInput.expectedAction, result.Action)
			}
		})
	}
}

// TestStateTransitions tests workflow state transitions
func TestStateTransitions(t *testing.T) {
	tempDir := t.TempDir()
	
	// Test complete workflow progression
	transitions := []struct {
		name       string
		setup      func(string) error
		fromState  WorkflowState
		toState    WorkflowState
		action     string
	}{
		{
			name:      "init_to_project",
			setup:     setupProjectStructure,
			fromState: StateNotInitialized,
			toState:   StateProjectInitialized,
			action:    "init-project",
		},
		{
			name:      "project_to_epics",
			setup:     setupWithEpics,
			fromState: StateProjectInitialized,
			toState:   StateHasEpics,
			action:    "create-epic",
		},
		{
			name:      "epics_to_progress",
			setup:     setupEpicInProgress,
			fromState: StateHasEpics,
			toState:   StateEpicInProgress,
			action:    "start-epic",
		},
	}
	
	for _, transition := range transitions {
		t.Run(transition.name, func(t *testing.T) {
			projectDir := filepath.Join(tempDir, transition.name)
			err := os.MkdirAll(projectDir, 0755)
			require.NoError(t, err)
			
			// Test initial state (empty)
			detector := NewContextDetector(projectDir)
			ctx, err := detector.DetectContext()
			require.NoError(t, err)
			assert.Equal(t, StateNotInitialized, ctx.State)
			
			// Apply transition setup
			err = transition.setup(projectDir)
			require.NoError(t, err)
			
			// Test final state
			ctx, err = detector.DetectContext()
			require.NoError(t, err)
			assert.Equal(t, transition.toState, ctx.State)
			
			// Test that suggestions include the next appropriate action
			engine := NewSuggestionEngine()
			suggestions, err := engine.GenerateSuggestions(ctx)
			require.NoError(t, err)
			require.NotEmpty(t, suggestions)
		})
	}
}

// Setup functions for different project states

func setupEmptyProject(dir string) error {
	// Empty directory
	return nil
}

func setupProjectStructure(dir string) error {
	dirs := []string{
		"docs/1-project",
		"docs/2-current-epic",
		"docs/3-current-task",
	}
	
	for _, d := range dirs {
		if err := os.MkdirAll(filepath.Join(dir, d), 0755); err != nil {
			return err
		}
	}
	return nil
}

func setupWithEpics(dir string) error {
	if err := setupProjectStructure(dir); err != nil {
		return err
	}
	
	epicsData := map[string]interface{}{
		"epics": []map[string]interface{}{
			{
				"id":     "EPIC-001",
				"title":  "Test Epic",
				"status": "todo",
			},
		},
	}
	
	data, err := json.Marshal(epicsData)
	if err != nil {
		return err
	}
	
	return os.WriteFile(filepath.Join(dir, "docs/1-project/epics.json"), data, 0644)
}

func setupEpicInProgress(dir string) error {
	if err := setupWithEpics(dir); err != nil {
		return err
	}
	
	currentEpicData := map[string]interface{}{
		"epic": map[string]interface{}{
			"id":       "EPIC-001",
			"title":    "Test Epic",
			"status":   "🚧 In Progress",
			"priority": "high",
			"userStories": []map[string]interface{}{
				{
					"id":     "US-001",
					"title":  "Test Story",
					"status": "todo",
				},
			},
		},
	}
	
	data, err := json.Marshal(currentEpicData)
	if err != nil {
		return err
	}
	
	return os.WriteFile(filepath.Join(dir, "docs/2-current-epic/current-epic.json"), data, 0644)
}

func setupStoryInProgress(dir string) error {
	if err := setupEpicInProgress(dir); err != nil {
		return err
	}
	
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
	if err != nil {
		return err
	}
	
	return os.WriteFile(filepath.Join(dir, "docs/2-current-epic/stories.json"), data, 0644)
}

func setupTaskInProgress(dir string) error {
	if err := setupStoryInProgress(dir); err != nil {
		return err
	}
	
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
	if err != nil {
		return err
	}
	
	return os.WriteFile(filepath.Join(dir, "docs/3-current-task/todo.json"), data, 0644)
}

func setupCorruptedEpic(dir string) error {
	if err := setupProjectStructure(dir); err != nil {
		return err
	}
	
	// Write invalid JSON
	return os.WriteFile(filepath.Join(dir, "docs/1-project/epics.json"), []byte("invalid json"), 0644)
}

func setupMissingCurrentEpic(dir string) error {
	if err := setupWithEpics(dir); err != nil {
		return err
	}
	
	// current-epic.json is missing (epics exist but none is current)
	return nil
}

func setupPermissionDenied(dir string) error {
	if err := setupProjectStructure(dir); err != nil {
		return err
	}
	
	// Create a file with restricted permissions
	epicsPath := filepath.Join(dir, "docs/1-project/epics.json")
	if err := os.WriteFile(epicsPath, []byte("{}"), 0000); err != nil {
		return err
	}
	
	return nil
}

func setupLargeEpicFile(dir string) error {
	if err := setupProjectStructure(dir); err != nil {
		return err
	}
	
	// Create large epics file with many epics
	epics := make([]map[string]interface{}, 100)
	for i := 0; i < 100; i++ {
		epics[i] = map[string]interface{}{
			"id":     "EPIC-" + string(rune(i+1)),
			"title":  "Large Epic " + string(rune(i+1)),
			"status": "todo",
			"userStories": make([]map[string]interface{}, 10),
		}
	}
	
	epicsData := map[string]interface{}{
		"epics": epics,
	}
	
	data, err := json.Marshal(epicsData)
	if err != nil {
		return err
	}
	
	return os.WriteFile(filepath.Join(dir, "docs/1-project/epics.json"), data, 0644)
}

// TestFullWorkflowCoverage ensures all workflow states are covered
func TestFullWorkflowCoverage(t *testing.T) {
	allStates := []WorkflowState{
		StateNotInitialized,
		StateProjectInitialized,
		StateHasEpics,
		StateEpicInProgress,
		StateStoryInProgress,
		StateTaskInProgress,
	}
	
	for _, state := range allStates {
		t.Run(state.String(), func(t *testing.T) {
			// Test that each state has a string representation
			assert.NotEmpty(t, state.String())
			
			// Test that each state has appropriate suggestions
			engine := NewSuggestionEngine()
			ctx := &ProjectContext{State: state}
			
			suggestions, err := engine.GenerateSuggestions(ctx)
			assert.NoError(t, err)
			assert.NotEmpty(t, suggestions, "State %s should have at least one suggestion", state.String())
			
			// Test that each state has appropriate display
			display := NewProjectStateDisplay()
			assert.NotPanics(t, func() {
				display.DisplayProjectOverview(ctx)
			})
		})
	}
}

// TestSuggestionEngineRobustness tests suggestion engine with various contexts
func TestSuggestionEngineRobustness(t *testing.T) {
	engine := NewSuggestionEngine()
	
	testCases := []struct {
		name string
		ctx  *ProjectContext
	}{
		{
			name: "nil_epic",
			ctx: &ProjectContext{
				State:       StateEpicInProgress,
				CurrentEpic: nil,
			},
		},
		{
			name: "nil_story",
			ctx: &ProjectContext{
				State:        StateStoryInProgress,
				CurrentStory: nil,
			},
		},
		{
			name: "nil_task",
			ctx: &ProjectContext{
				State:       StateTaskInProgress,
				CurrentTask: nil,
			},
		},
		{
			name: "with_issues",
			ctx: &ProjectContext{
				State:  StateProjectInitialized,
				Issues: []string{"Test issue 1", "Test issue 2"},
			},
		},
		{
			name: "near_completion",
			ctx: &ProjectContext{
				State: StateEpicInProgress,
				CurrentEpic: &EpicContext{
					Progress: 0.95, // 95% complete
					Title:    "Nearly Done Epic",
				},
			},
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			suggestions, err := engine.GenerateSuggestions(tc.ctx)
			assert.NoError(t, err)
			assert.NotEmpty(t, suggestions)
			
			// Test that suggestions are properly prioritized
			for i := 1; i < len(suggestions); i++ {
				prev := suggestions[i-1]
				curr := suggestions[i]
				
				// Lower priority items should not come before higher priority items
				if prev.Priority == workflow.PriorityP1 {
					assert.NotEqual(t, workflow.PriorityP0, curr.Priority, "P1 should not come before P0")
				}
				if prev.Priority == workflow.PriorityP2 {
					assert.NotEqual(t, workflow.PriorityP0, curr.Priority, "P2 should not come before P0")
					assert.NotEqual(t, workflow.PriorityP1, curr.Priority, "P2 should not come before P1")
				}
			}
		})
	}
}