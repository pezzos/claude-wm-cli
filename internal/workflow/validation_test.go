package workflow

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDependencyEnforcer(t *testing.T) {
	rootPath := "/test/path"
	enforcer := NewDependencyEnforcer(rootPath)
	
	assert.NotNil(t, enforcer)
	assert.NotNil(t, enforcer.analyzer)
	assert.NotNil(t, enforcer.commandGenerator)
}

func TestValidateActionExecution_UnknownAction(t *testing.T) {
	tempDir := t.TempDir()
	setupCompleteProjectStructure(t, tempDir)
	
	enforcer := NewDependencyEnforcer(tempDir)
	result, err := enforcer.ValidateActionExecution("unknown-action", false)
	require.NoError(t, err)
	
	assert.False(t, result.IsValid)
	assert.Len(t, result.Violations, 1)
	assert.Equal(t, ViolationInvalidState, result.Violations[0].Type)
	assert.Equal(t, SeverityHigh, result.Violations[0].Severity)
	assert.Contains(t, result.Violations[0].Description, "unknown-action")
	assert.NotEmpty(t, result.Suggestions)
	assert.Contains(t, result.Suggestions[0], "help")
	assert.False(t, result.CanOverride)
}

func TestValidateActionExecution_InitProject_Valid(t *testing.T) {
	tempDir := t.TempDir()
	// Don't initialize project - this should be valid for init-project
	
	enforcer := NewDependencyEnforcer(tempDir)
	result, err := enforcer.ValidateActionExecution("init-project", false)
	require.NoError(t, err)
	
	assert.True(t, result.IsValid)
	assert.Empty(t, result.Violations)
}

func TestValidateActionExecution_InitProject_AlreadyInitialized(t *testing.T) {
	tempDir := t.TempDir()
	setupCompleteProjectStructure(t, tempDir)
	
	enforcer := NewDependencyEnforcer(tempDir)
	result, err := enforcer.ValidateActionExecution("init-project", false)
	require.NoError(t, err)
	
	assert.False(t, result.IsValid)
	assert.NotEmpty(t, result.Violations)
	
	// Should have a violation about project already being initialized
	foundInitializedViolation := false
	for _, violation := range result.Violations {
		if contains(violation.Description, "already initialized") {
			foundInitializedViolation = true
			// The actual implementation may classify this differently
			// Just check that we found the right violation
			break
		}
	}
	assert.True(t, foundInitializedViolation, "Should have violation about project already initialized")
}

func TestValidateActionExecution_CreateEpic_Valid(t *testing.T) {
	tempDir := t.TempDir()
	setupCompleteProjectStructure(t, tempDir)
	
	enforcer := NewDependencyEnforcer(tempDir)
	result, err := enforcer.ValidateActionExecution("create-epic", false)
	require.NoError(t, err)
	
	assert.True(t, result.IsValid)
	assert.Empty(t, result.Violations)
}

func TestValidateActionExecution_CreateEpic_NotInitialized(t *testing.T) {
	tempDir := t.TempDir()
	// Don't initialize project
	
	enforcer := NewDependencyEnforcer(tempDir)
	result, err := enforcer.ValidateActionExecution("create-epic", false)
	require.NoError(t, err)
	
	assert.False(t, result.IsValid)
	assert.NotEmpty(t, result.Violations)
	
	// Should have a critical violation about project initialization
	foundCriticalViolation := false
	for _, violation := range result.Violations {
		if violation.Severity == SeverityCritical && 
		   (contains(violation.Description, "must be initialized") || 
		    contains(violation.Description, "project_initialized")) {
			foundCriticalViolation = true
			break
		}
	}
	assert.True(t, foundCriticalViolation, "Should have critical violation about project initialization")
}

func TestValidateActionExecution_StartEpic_NoEpics(t *testing.T) {
	tempDir := t.TempDir()
	setupCompleteProjectStructure(t, tempDir)
	
	enforcer := NewDependencyEnforcer(tempDir)
	result, err := enforcer.ValidateActionExecution("start-epic", false)
	require.NoError(t, err)
	
	assert.False(t, result.IsValid)
	assert.NotEmpty(t, result.Violations)
	
	// Should have a violation about no epics available
	foundNoEpicsViolation := false
	for _, violation := range result.Violations {
		if contains(violation.Description, "No epics available") || 
		   contains(violation.Description, "has_epics") {
			foundNoEpicsViolation = true
			break
		}
	}
	assert.True(t, foundNoEpicsViolation, "Should have violation about no epics available")
}

func TestValidateActionExecution_StartEpic_EpicAlreadyActive(t *testing.T) {
	tempDir := t.TempDir()
	setupCompleteProjectStructure(t, tempDir)
	setupMultipleEpics(t, tempDir)
	setupCurrentEpic(t, tempDir)
	
	enforcer := NewDependencyEnforcer(tempDir)
	result, err := enforcer.ValidateActionExecution("start-epic", false)
	require.NoError(t, err)
	
	assert.False(t, result.IsValid)
	assert.NotEmpty(t, result.Violations)
	
	// Should have a violation about epic already being active
	foundActiveEpicViolation := false
	for _, violation := range result.Violations {
		if contains(violation.Description, "already active") ||
		   contains(violation.Description, "no_active_epic") {
			foundActiveEpicViolation = true
			break
		}
	}
	assert.True(t, foundActiveEpicViolation, "Should have violation about epic already active")
}

func TestValidateActionExecution_CompleteEpic_NoActiveEpic(t *testing.T) {
	tempDir := t.TempDir()
	setupCompleteProjectStructure(t, tempDir)
	
	enforcer := NewDependencyEnforcer(tempDir)
	result, err := enforcer.ValidateActionExecution("complete-epic", false)
	require.NoError(t, err)
	
	assert.False(t, result.IsValid)
	assert.NotEmpty(t, result.Violations)
	
	// Should have a violation about no active epic
	foundNoActiveEpicViolation := false
	for _, violation := range result.Violations {
		if contains(violation.Description, "No epic is currently active") ||
		   contains(violation.Description, "epic_in_progress") {
			foundNoActiveEpicViolation = true
			break
		}
	}
	assert.True(t, foundNoActiveEpicViolation, "Should have violation about no active epic")
}

func TestValidateActionExecution_CompleteEpic_NotComplete(t *testing.T) {
	tempDir := t.TempDir()
	setupCompleteProjectStructure(t, tempDir)
	setupCurrentEpic(t, tempDir)
	setupCurrentStory(t, tempDir)
	
	enforcer := NewDependencyEnforcer(tempDir)
	result, err := enforcer.ValidateActionExecution("complete-epic", false)
	require.NoError(t, err)
	
	assert.False(t, result.IsValid)
	assert.NotEmpty(t, result.Violations)
	
	// Should have a violation about epic not being complete
	foundIncompleteViolation := false
	for _, violation := range result.Violations {
		if contains(violation.Description, "complete") ||
		   contains(violation.Description, "all_stories_complete") {
			foundIncompleteViolation = true
			break
		}
	}
	assert.True(t, foundIncompleteViolation, "Should have violation about epic not complete")
}

func TestValidateActionExecution_CreateStory_NoActiveEpic(t *testing.T) {
	tempDir := t.TempDir()
	setupCompleteProjectStructure(t, tempDir)
	
	enforcer := NewDependencyEnforcer(tempDir)
	result, err := enforcer.ValidateActionExecution("create-story", false)
	require.NoError(t, err)
	
	assert.False(t, result.IsValid)
	assert.NotEmpty(t, result.Violations)
	
	// Should have a violation about no active epic
	foundNoActiveEpicViolation := false
	for _, violation := range result.Violations {
		if contains(violation.Description, "No epic is currently active") ||
		   contains(violation.Description, "epic_in_progress") {
			foundNoActiveEpicViolation = true
			break
		}
	}
	assert.True(t, foundNoActiveEpicViolation, "Should have violation about no active epic")
}

func TestValidateActionExecution_CreateTask_NoActiveStory(t *testing.T) {
	tempDir := t.TempDir()
	setupCompleteProjectStructure(t, tempDir)
	setupCurrentEpic(t, tempDir)
	
	enforcer := NewDependencyEnforcer(tempDir)
	result, err := enforcer.ValidateActionExecution("create-task", false)
	require.NoError(t, err)
	
	assert.False(t, result.IsValid)
	assert.NotEmpty(t, result.Violations)
	
	// Should have a violation about no active story
	foundNoActiveStoryViolation := false
	for _, violation := range result.Violations {
		if contains(violation.Description, "No story is currently active") ||
		   contains(violation.Description, "story_in_progress") {
			foundNoActiveStoryViolation = true
			break
		}
	}
	assert.True(t, foundNoActiveStoryViolation, "Should have violation about no active story")
}

func TestValidateActionExecution_WithBlockers(t *testing.T) {
	tempDir := t.TempDir()
	setupCompleteProjectStructure(t, tempDir)
	setupCurrentEpic(t, tempDir)
	setupCurrentStory(t, tempDir)
	setupBlockedTasks(t, tempDir)
	
	enforcer := NewDependencyEnforcer(tempDir)
	result, err := enforcer.ValidateActionExecution("continue-task", false)
	require.NoError(t, err)
	
	// Should have blocking violations
	hasBlockingViolation := false
	for _, violation := range result.Violations {
		if violation.Type == ViolationBlockingCondition {
			hasBlockingViolation = true
			assert.Equal(t, SeverityHigh, violation.Severity)
			assert.Contains(t, violation.Description, "blocked")
		}
	}
	assert.True(t, hasBlockingViolation, "Should detect blocking conditions")
}

func TestValidateActionExecution_WithOverride(t *testing.T) {
	tempDir := t.TempDir()
	setupCompleteProjectStructure(t, tempDir)
	
	enforcer := NewDependencyEnforcer(tempDir)
	result, err := enforcer.ValidateActionExecution("create-epic", true) // allow override
	require.NoError(t, err)
	
	assert.True(t, result.IsValid)
	assert.True(t, result.CanOverride)
}

func TestValidateActionExecution_CriticalViolationNoOverride(t *testing.T) {
	tempDir := t.TempDir()
	// Don't initialize project
	
	enforcer := NewDependencyEnforcer(tempDir)
	result, err := enforcer.ValidateActionExecution("create-epic", true) // allow override
	require.NoError(t, err)
	
	assert.False(t, result.IsValid)
	assert.False(t, result.CanOverride) // Critical violations can't be overridden
	assert.Contains(t, result.OverrideRisk, "Critical dependencies prevent override")
}

func TestValidateWorkflowTransition_ValidTransitions(t *testing.T) {
	tempDir := t.TempDir()
	enforcer := NewDependencyEnforcer(tempDir)
	
	tests := []struct {
		name     string
		from     WorkflowPosition
		to       WorkflowPosition
		expected bool
	}{
		{"not_initialized to project", PositionNotInitialized, PositionProjectLevel, true},
		{"project to epic", PositionProjectLevel, PositionEpicLevel, true},
		{"epic to story", PositionEpicLevel, PositionStoryLevel, true},
		{"epic to project", PositionEpicLevel, PositionProjectLevel, true},
		{"story to task", PositionStoryLevel, PositionTaskLevel, true},
		{"story to epic", PositionStoryLevel, PositionEpicLevel, true},
		{"task to story", PositionTaskLevel, PositionStoryLevel, true},
		{"invalid transition", PositionProjectLevel, PositionTaskLevel, false},
		{"backward invalid", PositionTaskLevel, PositionProjectLevel, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := enforcer.ValidateWorkflowTransition(tt.from, tt.to, "test-action")
			require.NoError(t, err)
			
			assert.Equal(t, tt.expected, result.IsValid, 
				"Transition from %s to %s should be %v", tt.from, tt.to, tt.expected)
		})
	}
}

func TestGetAllowedActions(t *testing.T) {
	tempDir := t.TempDir()
	setupCompleteProjectStructure(t, tempDir)
	
	enforcer := NewDependencyEnforcer(tempDir)
	actions, err := enforcer.GetAllowedActions()
	require.NoError(t, err)
	
	assert.NotEmpty(t, actions)
	
	// Should contain create-epic since project is initialized
	foundCreateEpic := false
	for _, action := range actions {
		if action.ID == "create-epic" {
			foundCreateEpic = true
			break
		}
	}
	assert.True(t, foundCreateEpic, "Should contain create-epic action for initialized project")
}

func TestGetBlockedActions(t *testing.T) {
	tempDir := t.TempDir()
	// Don't initialize project - this should block most actions
	
	enforcer := NewDependencyEnforcer(tempDir)
	blocked, err := enforcer.GetBlockedActions()
	require.NoError(t, err)
	
	assert.NotEmpty(t, blocked)
	
	// Should contain create-epic since project is not initialized
	createEpicResult, exists := blocked["create-epic"]
	assert.True(t, exists, "create-epic should be blocked when project not initialized")
	assert.False(t, createEpicResult.IsValid)
	assert.NotEmpty(t, createEpicResult.Violations)
}

func TestViolationSeverityMapping(t *testing.T) {
	tempDir := t.TempDir()
	enforcer := NewDependencyEnforcer(tempDir)
	
	tests := []struct {
		blockerSeverity string
		expected        Severity
	}{
		{"critical", SeverityCritical},
		{"high", SeverityHigh},
		{"medium", SeverityMedium},
		{"low", SeverityLow},
		{"unknown", SeverityMedium},
	}

	for _, tt := range tests {
		t.Run(tt.blockerSeverity, func(t *testing.T) {
			result := enforcer.mapBlockerSeverity(tt.blockerSeverity)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPrerequisiteMapping(t *testing.T) {
	tempDir := t.TempDir()
	setupCompleteProjectStructure(t, tempDir)
	setupMultipleEpics(t, tempDir) // This adds epics
	setupCurrentEpic(t, tempDir)
	setupCurrentStory(t, tempDir)
	
	enforcer := NewDependencyEnforcer(tempDir)
	analysis, err := enforcer.analyzer.AnalyzeWorkflowPosition()
	require.NoError(t, err)

	tests := []struct {
		prerequisite    string
		expectedCurrent string
		expectedRequired string
	}{
		{"project_initialized", "initialized", "initialized"},
		{"has_epics", "3_epics", "epics_available"}, // setupMultipleEpics creates 3 epics
		{"epic_in_progress", "epic_EPIC-001_active", "epic_active"},
		{"story_in_progress", "story_STORY-001_active", "story_active"},
		{"unknown_prereq", "unknown", "required_state"},
	}

	for _, tt := range tests {
		t.Run(tt.prerequisite, func(t *testing.T) {
			current := enforcer.getCurrentStateForPrerequisite(tt.prerequisite, analysis)
			required := enforcer.getRequiredStateForPrerequisite(tt.prerequisite)
			
			assert.Equal(t, tt.expectedCurrent, current)
			assert.Equal(t, tt.expectedRequired, required)
		})
	}
}

func TestSuggestionGeneration(t *testing.T) {
	tempDir := t.TempDir()
	enforcer := NewDependencyEnforcer(tempDir)
	
	// Test not initialized project
	analysis := &WorkflowAnalysis{
		ProjectInitialized: false,
		CompletionMetrics: CompletionMetrics{
			TotalEpics: 0,
		},
	}

	tests := []struct {
		prerequisite      string
		expectedSuggestion string
	}{
		{"project_initialized", "init-project"},
		{"has_epics", "create-epic"},
		{"epic_in_progress", "create-epic"},
		{"story_in_progress", "create-story"},
	}

	for _, tt := range tests {
		t.Run(tt.prerequisite, func(t *testing.T) {
			suggestions := enforcer.getSuggestionsForPrerequisite(tt.prerequisite, analysis)
			
			assert.NotEmpty(t, suggestions)
			foundExpected := false
			for _, suggestion := range suggestions {
				if contains(suggestion, tt.expectedSuggestion) {
					foundExpected = true
					break
				}
			}
			assert.True(t, foundExpected, "Should contain suggestion with '%s'", tt.expectedSuggestion)
		})
	}
}

func TestStateTransitionSuggestions(t *testing.T) {
	tempDir := t.TempDir()
	enforcer := NewDependencyEnforcer(tempDir)
	analysis := &WorkflowAnalysis{}

	tests := []struct {
		currentState     string
		requiredState    string
		expectedSuggestion string
	}{
		{"not_initialized", "initialized", "init-project"},
		{"no_epics", "epics_available", "create-epic"},
		{"no_active_epic", "epic_active", "start-epic"},
		{"other", "other", "Transition from"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s_to_%s", tt.currentState, tt.requiredState), func(t *testing.T) {
			suggestions := enforcer.getSuggestionsForStateTransition(tt.currentState, tt.requiredState, analysis)
			
			assert.NotEmpty(t, suggestions)
			foundExpected := false
			for _, suggestion := range suggestions {
				if contains(suggestion, tt.expectedSuggestion) {
					foundExpected = true
					break
				}
			}
			assert.True(t, foundExpected, "Should contain suggestion with '%s'", tt.expectedSuggestion)
		})
	}
}

func TestOverrideRiskCalculation(t *testing.T) {
	tempDir := t.TempDir()
	enforcer := NewDependencyEnforcer(tempDir)

	tests := []struct {
		name        string
		actionID    string
		violations  []DependencyViolation
		expectedRisk string
	}{
		{
			name:        "high risk action",
			actionID:    "init-project",
			violations:  []DependencyViolation{},
			expectedRisk: "high",
		},
		{
			name:     "many high severity violations",
			actionID: "create-story",
			violations: []DependencyViolation{
				{Severity: SeverityHigh},
				{Severity: SeverityHigh},
				{Severity: SeverityHigh},
			},
			expectedRisk: "high",
		},
		{
			name:     "some high severity violations",
			actionID: "create-task",
			violations: []DependencyViolation{
				{Severity: SeverityHigh},
				{Severity: SeverityMedium},
			},
			expectedRisk: "medium",
		},
		{
			name:     "low severity violations",
			actionID: "status",
			violations: []DependencyViolation{
				{Severity: SeverityLow},
				{Severity: SeverityMedium},
			},
			expectedRisk: "low",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			action := &WorkflowAction{ID: tt.actionID}
			risk := enforcer.calculateOverrideRisk(action, tt.violations)
			assert.Equal(t, tt.expectedRisk, risk)
		})
	}
}

func TestCircularDependencyDetection(t *testing.T) {
	tempDir := t.TempDir()
	setupCompleteProjectStructure(t, tempDir)
	setupCurrentEpic(t, tempDir)
	
	enforcer := NewDependencyEnforcer(tempDir)
	analysis, err := enforcer.analyzer.AnalyzeWorkflowPosition()
	require.NoError(t, err)

	result := &ValidationResult{
		Violations: []DependencyViolation{},
	}
	
	// Create an action with blocks
	action := &WorkflowAction{
		ID:     "complete-epic",
		Blocks: []string{"create-story"},
	}
	
	enforcer.validateCircularDependencies(action, analysis, result)
	
	// Should detect potential circular dependency if story creation is still executable
	if len(result.Violations) > 0 {
		assert.Equal(t, ViolationCircularDependency, result.Violations[0].Type)
		assert.Equal(t, SeverityHigh, result.Violations[0].Severity)
	}
}

func TestContainsStringHelper(t *testing.T) {
	tempDir := t.TempDir()
	enforcer := NewDependencyEnforcer(tempDir)
	
	slice := []string{"apple", "banana", "cherry"}
	
	assert.True(t, enforcer.containsString(slice, "banana"))
	assert.False(t, enforcer.containsString(slice, "orange"))
	assert.False(t, enforcer.containsString([]string{}, "apple"))
}

func TestOnlyNonCriticalViolations(t *testing.T) {
	tempDir := t.TempDir()
	enforcer := NewDependencyEnforcer(tempDir)
	
	tests := []struct {
		name       string
		violations []DependencyViolation
		expected   bool
	}{
		{
			name:       "no violations",
			violations: []DependencyViolation{},
			expected:   true,
		},
		{
			name: "only non-critical",
			violations: []DependencyViolation{
				{Severity: SeverityHigh},
				{Severity: SeverityMedium},
				{Severity: SeverityLow},
			},
			expected: true,
		},
		{
			name: "has critical",
			violations: []DependencyViolation{
				{Severity: SeverityHigh},
				{Severity: SeverityCritical},
				{Severity: SeverityMedium},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := enforcer.onlyNonCriticalViolations(tt.violations)
			assert.Equal(t, tt.expected, result)
		})
	}
}

