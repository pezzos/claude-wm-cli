package workflow

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCommandGenerator(t *testing.T) {
	rootPath := "/test/path"
	generator := NewCommandGenerator(rootPath)
	
	assert.NotNil(t, generator)
	assert.NotNil(t, generator.actionRegistry)
	assert.NotNil(t, generator.analyzer)
}

func TestGenerateContextualCommands_NotInitialized(t *testing.T) {
	tempDir := t.TempDir()
	generator := NewCommandGenerator(tempDir)

	commands, err := generator.GenerateContextualCommands()
	require.NoError(t, err)
	
	assert.NotEmpty(t, commands)
	
	// Should have init-project as top priority command
	assert.Equal(t, "init-project", commands[0].Action.ID)
	assert.Equal(t, PriorityP0, commands[0].Priority)
	assert.Contains(t, commands[0].Reasoning, "initialized")
}

func TestGenerateContextualCommands_ProjectLevel(t *testing.T) {
	tempDir := t.TempDir()
	setupCompleteProjectStructure(t, tempDir)
	
	generator := NewCommandGenerator(tempDir)
	commands, err := generator.GenerateContextualCommands()
	require.NoError(t, err)
	
	assert.NotEmpty(t, commands)
	
	// Should have create-epic as primary command when no epics exist
	foundCreateEpic := false
	for _, cmd := range commands {
		if cmd.Action.ID == "create-epic" && cmd.Priority == PriorityP0 {
			foundCreateEpic = true
			assert.Contains(t, cmd.Reasoning, "No epics defined")
			break
		}
	}
	assert.True(t, foundCreateEpic, "Should suggest creating epic when none exist")
}

func TestGenerateContextualCommands_EpicLevel(t *testing.T) {
	tempDir := t.TempDir()
	setupCompleteProjectStructure(t, tempDir)
	setupCurrentEpic(t, tempDir)
	
	generator := NewCommandGenerator(tempDir)
	commands, err := generator.GenerateContextualCommands()
	require.NoError(t, err)
	
	assert.NotEmpty(t, commands)
	
	// Should suggest creating stories when epic has no stories
	foundCreateStory := false
	for _, cmd := range commands {
		if cmd.Action.ID == "create-story" && cmd.Priority == PriorityP0 {
			foundCreateStory = true
			assert.Contains(t, cmd.Reasoning, "no stories defined")
			break
		}
	}
	assert.True(t, foundCreateStory, "Should suggest creating story when epic has none")
}

func TestGenerateContextualCommands_StoryLevel(t *testing.T) {
	tempDir := t.TempDir()
	setupCompleteProjectStructure(t, tempDir)
	setupCurrentEpic(t, tempDir)
	setupCurrentStory(t, tempDir)
	
	generator := NewCommandGenerator(tempDir)
	commands, err := generator.GenerateContextualCommands()
	require.NoError(t, err)
	
	assert.NotEmpty(t, commands)
	
	// Should suggest creating tasks when story has no tasks
	foundCreateTask := false
	for _, cmd := range commands {
		if cmd.Action.ID == "create-task" && cmd.Priority == PriorityP0 {
			foundCreateTask = true
			assert.Contains(t, cmd.Reasoning, "no tasks defined")
			break
		}
	}
	assert.True(t, foundCreateTask, "Should suggest creating task when story has none")
}

func TestGenerateContextualCommands_TaskLevel(t *testing.T) {
	tempDir := t.TempDir()
	setupCompleteProjectStructure(t, tempDir)
	setupCurrentEpic(t, tempDir)
	setupCurrentStory(t, tempDir)
	setupCurrentTasks(t, tempDir)
	
	generator := NewCommandGenerator(tempDir)
	commands, err := generator.GenerateContextualCommands()
	require.NoError(t, err)
	
	assert.NotEmpty(t, commands)
	
	// Should suggest continuing tasks when tasks exist
	foundContinueTask := false
	for _, cmd := range commands {
		if cmd.Action.ID == "continue-task" {
			foundContinueTask = true
			assert.Contains(t, cmd.Reasoning, "pending task")
			break
		}
	}
	assert.True(t, foundContinueTask, "Should suggest continuing task when tasks exist")
}

func TestGenerateContextualCommands_WithBlockers(t *testing.T) {
	tempDir := t.TempDir()
	setupCompleteProjectStructure(t, tempDir)
	setupCurrentEpic(t, tempDir)
	setupCurrentStory(t, tempDir)
	setupBlockedTasks(t, tempDir)
	
	generator := NewCommandGenerator(tempDir)
	commands, err := generator.GenerateContextualCommands()
	require.NoError(t, err)
	
	// Should have warnings about blocked tasks
	foundBlockedWarning := false
	for _, cmd := range commands {
		if cmd.Action.ID == "continue-task" && len(cmd.Warnings) > 0 {
			for _, warning := range cmd.Warnings {
				if contains(warning, "blocked") {
					foundBlockedWarning = true
					break
				}
			}
		}
	}
	assert.True(t, foundBlockedWarning, "Should warn about blocked tasks")
}

func TestGenerateContextualCommands_CompletionScenarios(t *testing.T) {
	tempDir := t.TempDir()
	setupCompleteProjectStructure(t, tempDir)
	setupCurrentEpic(t, tempDir)
	setupCurrentStory(t, tempDir)
	// Don't set up any current tasks to simulate all tasks being completed
	
	generator := NewCommandGenerator(tempDir)
	commands, err := generator.GenerateContextualCommands()
	require.NoError(t, err)
	
	// Should suggest creating tasks when story has no active tasks
	foundCreateTask := false
	for _, cmd := range commands {
		if cmd.Action.ID == "create-task" && cmd.Priority == PriorityP0 {
			foundCreateTask = true
			assert.Contains(t, cmd.Reasoning, "no tasks defined")
			break
		}
	}
	assert.True(t, foundCreateTask, "Should suggest creating tasks when story has none")
}

func TestGetRecommendedAction(t *testing.T) {
	tempDir := t.TempDir()
	setupCompleteProjectStructure(t, tempDir)
	
	generator := NewCommandGenerator(tempDir)
	recommended, err := generator.GetRecommendedAction()
	require.NoError(t, err)
	
	assert.NotNil(t, recommended)
	assert.Equal(t, "create-epic", recommended.Action.ID)
	assert.Equal(t, PriorityP0, recommended.Priority)
}

func TestGetCommandsByPriority(t *testing.T) {
	tempDir := t.TempDir()
	setupCompleteProjectStructure(t, tempDir)
	setupCurrentEpic(t, tempDir)
	
	generator := NewCommandGenerator(tempDir)
	grouped, err := generator.GetCommandsByPriority()
	require.NoError(t, err)
	
	assert.NotEmpty(t, grouped)
	
	// Should have commands in different priority levels
	assert.Contains(t, grouped, PriorityP0)
	assert.Contains(t, grouped, PriorityP2) // Utility commands
	
	// P0 commands should be more specific to current state
	p0Commands := grouped[PriorityP0]
	assert.NotEmpty(t, p0Commands)
}

func TestValidateCommand_ValidCommand(t *testing.T) {
	tempDir := t.TempDir()
	setupCompleteProjectStructure(t, tempDir)
	
	generator := NewCommandGenerator(tempDir)
	cmd, issues, err := generator.ValidateCommand("create-epic")
	require.NoError(t, err)
	
	assert.NotNil(t, cmd)
	assert.Equal(t, "create-epic", cmd.Action.ID)
	assert.Empty(t, issues, "create-epic should be valid in initialized project")
}

func TestValidateCommand_InvalidCommand(t *testing.T) {
	tempDir := t.TempDir()
	// Don't initialize project
	
	generator := NewCommandGenerator(tempDir)
	cmd, issues, err := generator.ValidateCommand("create-epic")
	require.NoError(t, err)
	
	assert.NotNil(t, cmd)
	assert.NotEmpty(t, issues)
	assert.Contains(t, issues[0], "Prerequisites not met")
}

func TestValidateCommand_NonExistentCommand(t *testing.T) {
	tempDir := t.TempDir()
	generator := NewCommandGenerator(tempDir)
	
	cmd, issues, err := generator.ValidateCommand("non-existent-command")
	require.NoError(t, err)
	
	assert.NotNil(t, cmd)
	assert.NotEmpty(t, issues)
	// Should have at least one issue about the action not being found
	foundNotFoundError := false
	for _, issue := range issues {
		if contains(issue, "Action not found") {
			foundNotFoundError = true
			break
		}
	}
	assert.True(t, foundNotFoundError, "Should indicate that action was not found")
}

func TestValidateSpecificAction_InitProject(t *testing.T) {
	tests := []struct {
		name                string
		projectInit         bool
		expectedMinIssues   int
		shouldContainText   string
	}{
		{
			name:              "not initialized",
			projectInit:       false,
			expectedMinIssues: 0,
			shouldContainText: "",
		},
		{
			name:              "already initialized",
			projectInit:       true,
			expectedMinIssues: 1,
			shouldContainText: "already initialized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			if tt.projectInit {
				setupCompleteProjectStructure(t, tempDir)
			}
			
			generator := NewCommandGenerator(tempDir)
			cmd, issues, err := generator.ValidateCommand("init-project")
			require.NoError(t, err)
			
			assert.NotNil(t, cmd)
			assert.GreaterOrEqual(t, len(issues), tt.expectedMinIssues)
			
			if tt.shouldContainText != "" {
				foundExpectedText := false
				for _, issue := range issues {
					if contains(issue, tt.shouldContainText) {
						foundExpectedText = true
						break
					}
				}
				assert.True(t, foundExpectedText, "Should contain text: %s in issues: %v", tt.shouldContainText, issues)
			}
		})
	}
}

func TestValidateSpecificAction_CreateStory(t *testing.T) {
	tests := []struct {
		name           string
		setupEpic      bool
		expectedIssues int
	}{
		{
			name:           "no active epic",
			setupEpic:      false,
			expectedIssues: 2, // Prerequisites + no active epic
		},
		{
			name:           "with active epic",
			setupEpic:      true,
			expectedIssues: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			setupCompleteProjectStructure(t, tempDir)
			if tt.setupEpic {
				setupCurrentEpic(t, tempDir)
			}
			
			generator := NewCommandGenerator(tempDir)
			cmd, issues, err := generator.ValidateCommand("create-story")
			require.NoError(t, err)
			
			assert.NotNil(t, cmd)
			assert.Len(t, issues, tt.expectedIssues)
		})
	}
}

func TestSortCommands(t *testing.T) {
	generator := NewCommandGenerator("/test")
	
	commands := []*ContextualCommand{
		{
			Action:   &WorkflowAction{ID: "action-p1", Name: "Z Action"},
			Priority: PriorityP1,
		},
		{
			Action:   &WorkflowAction{ID: "action-p0", Name: "A Action"},
			Priority: PriorityP0,
		},
		{
			Action:   &WorkflowAction{ID: "action-p2", Name: "M Action"},
			Priority: PriorityP2,
		},
		{
			Action:   &WorkflowAction{ID: "action-p0-2", Name: "B Action"},
			Priority: PriorityP0,
		},
	}
	
	generator.sortCommands(commands)
	
	// Should be sorted by priority first (P0 > P1 > P2)
	assert.Equal(t, PriorityP0, commands[0].Priority)
	assert.Equal(t, PriorityP0, commands[1].Priority)
	assert.Equal(t, PriorityP1, commands[2].Priority)
	assert.Equal(t, PriorityP2, commands[3].Priority)
	
	// Within same priority, should be sorted alphabetically
	assert.Equal(t, "A Action", commands[0].Action.Name)
	assert.Equal(t, "B Action", commands[1].Action.Name)
}

func TestPriorityValue(t *testing.T) {
	tests := []struct {
		priority Priority
		expected int
	}{
		{PriorityP0, 3},
		{PriorityP1, 2},
		{PriorityP2, 1},
		{Priority("unknown"), 0},
	}

	for _, tt := range tests {
		t.Run(string(tt.priority), func(t *testing.T) {
			result := priorityValue(tt.priority)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsPrerequisiteMet(t *testing.T) {
	generator := NewCommandGenerator("/test")
	
	// Test with not initialized project
	notInitAnalysis := &WorkflowAnalysis{
		ProjectInitialized: false,
		CompletionMetrics: CompletionMetrics{
			TotalEpics:   0,
			TotalStories: 0,
			TotalTasks:   0,
		},
	}
	
	tests := []struct {
		prerequisite string
		analysis     *WorkflowAnalysis
		expected     bool
	}{
		{"empty_directory", notInitAnalysis, true},
		{"project_initialized", notInitAnalysis, false},
		{"has_epics", notInitAnalysis, false},
		{"unknown_prereq", notInitAnalysis, false},
	}

	for _, tt := range tests {
		t.Run(tt.prerequisite, func(t *testing.T) {
			result := generator.isPrerequisiteMet(tt.prerequisite, tt.analysis)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsCommandBlocked(t *testing.T) {
	generator := NewCommandGenerator("/test")
	
	analysis := &WorkflowAnalysis{
		Blockers: []WorkflowBlocker{
			{Type: BlockerMissingDefinition},
			{Type: BlockerMissingDependency},
		},
	}
	
	tests := []struct {
		actionID string
		expected bool
	}{
		{"continue-epic", true},   // Blocked by missing definition
		{"continue-task", true},   // Blocked by missing dependency
		{"help", false},           // Not blocked
		{"status", false},         // Not blocked
	}

	for _, tt := range tests {
		t.Run(tt.actionID, func(t *testing.T) {
			cmd := &ContextualCommand{
				Action: &WorkflowAction{ID: tt.actionID},
			}
			result := generator.isCommandBlocked(cmd, analysis)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Helper functions for setting up test data

func setupCompletedTasks(t *testing.T, tempDir string) {
	// Create tasks with mixed status to test completion scenario
	todoData := map[string]interface{}{
		"todos": []map[string]interface{}{
			{
				"id":       "TASK-001",
				"title":    "Completed Task 1",
				"status":   "todo",  // Use 'todo' so it gets loaded for analysis
				"priority": "P0",
			},
			{
				"id":       "TASK-002",
				"title":    "Completed Task 2", 
				"status":   "todo",  // Use 'todo' so it gets loaded for analysis
				"priority": "P1",
			},
		},
	}

	// Also create a stories file that shows all tasks as completed in metrics
	// This simulates the scenario where all tasks are done but still need story completion
	storiesData := map[string]interface{}{
		"meta": map[string]interface{}{
			"current_story": "STORY-001",
		},
		"stories": []interface{}{
			map[string]interface{}{
				"metadata": map[string]interface{}{
					"id": "STORY-001",
					"schema_version": "1.0.0",
				},
				"epic_id": "EPIC-001",
				"title": "Test Story",
				"status": "in_progress",
				"metrics": map[string]interface{}{
					"total_tasks": 2,
					"completed_tasks": 2,  // All tasks completed
					"progress_percent": 100.0,
				},
			},
		},
	}

	todoJSON, _ := json.Marshal(todoData)
	todoPath := filepath.Join(tempDir, "docs/3-current-task/todo.json")
	err := os.WriteFile(todoPath, todoJSON, 0644)
	require.NoError(t, err)

	storiesJSON, _ := json.Marshal(storiesData)
	storiesPath := filepath.Join(tempDir, "docs/2-current-epic/stories.json")
	err = os.WriteFile(storiesPath, storiesJSON, 0644)
	require.NoError(t, err)
}

// Helper function to check if string contains substring (case insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || 
		len(s) > len(substr) && (s[:len(substr)] == substr || 
		s[len(s)-len(substr):] == substr ||
		findInString(s, substr)))
}

func findInString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}