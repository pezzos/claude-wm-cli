package navigation

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"claude-wm-cli/internal/workflow"
)

// captureOutput captures stdout during test execution
func captureOutput(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

func TestProjectStateDisplay_CreateProgressBar(t *testing.T) {
	display := NewProjectStateDisplay()
	
	tests := []struct {
		progress float64
		width    int
		expected string
	}{
		{0.0, 10, "[░░░░░░░░░░]"},
		{0.5, 10, "[█████░░░░░]"},
		{1.0, 10, "[██████████]"},
		{0.25, 8, "[██░░░░░░]"},
		{0.75, 4, "[███░]"},
		{-0.1, 5, "[░░░░░]"}, // Negative should be clamped to 0
		{1.5, 5, "[█████]"},  // Over 1 should be clamped to 1
	}
	
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%.2f_%d", tt.progress, tt.width), func(t *testing.T) {
			result := display.createProgressBar(tt.progress, tt.width)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestProjectStateDisplay_GetStateIcon(t *testing.T) {
	display := NewProjectStateDisplay()
	
	tests := []struct {
		state    WorkflowState
		expected string
	}{
		{StateNotInitialized, "🆕 "},
		{StateProjectInitialized, "📁 "},
		{StateHasEpics, "📚 "},
		{StateEpicInProgress, "🚧 "},
		{StateStoryInProgress, "📖 "},
		{StateTaskInProgress, "⚡ "},
	}
	
	for _, tt := range tests {
		t.Run(tt.state.String(), func(t *testing.T) {
			result := display.getStateIcon(tt.state)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestProjectStateDisplay_GetStatusIcon(t *testing.T) {
	display := NewProjectStateDisplay()
	
	tests := []struct {
		status   string
		expected string
	}{
		{"completed", "✅"},
		{"done", "✅"},
		{"in_progress", "🚧"},
		{"progress", "🚧"},
		{"todo", "⏳"},
		{"pending", "⏳"},
		{"blocked", "🚫"},
		{"review", "👀"},
		{"unknown_status", "📋"},
	}
	
	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			result := display.getStatusIcon(tt.status)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestProjectStateDisplay_GetPriorityIcon(t *testing.T) {
	display := NewProjectStateDisplay()
	
	tests := []struct {
		priority string
		expected string
	}{
		{"high", "🔴 "},
		{"P0", "🔴 "},
		{"medium", "🟡 "},
		{"P1", "🟡 "},
		{"low", "🟢 "},
		{"P2", "🟢 "},
		{"unknown", "⚪ "},
	}
	
	for _, tt := range tests {
		t.Run(tt.priority, func(t *testing.T) {
			result := display.getPriorityIcon(tt.priority)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestProjectStateDisplay_GetProjectName(t *testing.T) {
	display := NewProjectStateDisplay()
	
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "extract from path",
			path:     "/home/user/projects/my-project",
			expected: "my-project",
		},
		{
			name:     "extract from nested path",
			path:     "/very/deep/path/to/awesome-project",
			expected: "awesome-project",
		},
		{
			name:     "empty path",
			path:     "",
			expected: "Claude WM Project",
		},
		{
			name:     "root path",
			path:     "/",
			expected: "Claude WM Project",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &ProjectContext{ProjectPath: tt.path}
			result := display.getProjectName(ctx)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestProjectStateDisplay_GetNextMilestone(t *testing.T) {
	display := NewProjectStateDisplay()
	
	tests := []struct {
		state    WorkflowState
		expected string
	}{
		{StateNotInitialized, "Initialize project structure"},
		{StateProjectInitialized, "Create first epic"},
		{StateHasEpics, "Start working on an epic"},
		{StateEpicInProgress, "Complete current epic"},
		{StateTaskInProgress, "Continue current work"},
	}
	
	for _, tt := range tests {
		t.Run(tt.state.String(), func(t *testing.T) {
			ctx := &ProjectContext{State: tt.state}
			result := display.getNextMilestone(ctx)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestProjectStateDisplay_DisplayQuickStatus(t *testing.T) {
	display := NewProjectStateDisplay()
	
	tests := []struct {
		name     string
		ctx      *ProjectContext
		contains []string
	}{
		{
			name: "not initialized",
			ctx: &ProjectContext{
				State: StateNotInitialized,
			},
			contains: []string{"🆕", "Not Initialized"},
		},
		{
			name: "with epic",
			ctx: &ProjectContext{
				State: StateEpicInProgress,
				CurrentEpic: &EpicContext{
					Title:    "Test Epic",
					Progress: 0.75,
				},
			},
			contains: []string{"🚧", "Epic In Progress", "Test Epic", "75%"},
		},
		{
			name: "with story and task",
			ctx: &ProjectContext{
				State: StateTaskInProgress,
				CurrentEpic: &EpicContext{
					Title:    "Test Epic",
					Progress: 0.5,
				},
				CurrentStory: &StoryContext{
					Title: "Test Story",
				},
				CurrentTask: &TaskContext{
					Title: "Test Task",
				},
			},
			contains: []string{"Task In Progress", "Test Epic", "Test Story", "Test Task"},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := captureOutput(func() {
				display.DisplayQuickStatus(tt.ctx)
			})
			
			for _, expected := range tt.contains {
				assert.Contains(t, output, expected)
			}
		})
	}
}

func TestProjectStateDisplay_DisplayProgressSummary(t *testing.T) {
	display := NewProjectStateDisplay()
	
	ctx := &ProjectContext{
		State: StateStoryInProgress,
		CurrentEpic: &EpicContext{
			Title:    "Test Epic",
			Progress: 0.6,
		},
		CurrentStory: &StoryContext{
			Title:          "Test Story",
			CompletedTasks: 3,
			TotalTasks:     5,
		},
	}
	
	output := captureOutput(func() {
		display.DisplayProgressSummary(ctx)
	})
	
	assert.Contains(t, output, "Progress Summary")
	assert.Contains(t, output, "Epic:")
	assert.Contains(t, output, "60.0%")
	assert.Contains(t, output, "Story:")
	assert.Contains(t, output, "60.0%") // 3/5 = 60%
}

func TestProjectStateDisplay_DisplayActionSummary(t *testing.T) {
	display := NewProjectStateDisplay()
	
	tests := []struct {
		name    string
		actions []string
		want    []string
	}{
		{
			name:    "no actions",
			actions: []string{},
			want:    []string{"No actions available"},
		},
		{
			name:    "few actions",
			actions: []string{"action1", "action2", "action3"},
			want:    []string{"Available Actions (3)", "action1", "action2", "action3"},
		},
		{
			name:    "many actions",
			actions: []string{"a1", "a2", "a3", "a4", "a5", "a6", "a7", "a8", "a9", "a10"},
			want:    []string{"Available Actions (10)", "a1", "and 2 more"},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &ProjectContext{
				AvailableActions: tt.actions,
			}
			
			output := captureOutput(func() {
				display.DisplayActionSummary(ctx)
			})
			
			for _, expected := range tt.want {
				assert.Contains(t, output, expected)
			}
		})
	}
}

func TestProjectStateDisplay_DisplayWithSuggestions(t *testing.T) {
	display := NewProjectStateDisplay()
	
	ctx := &ProjectContext{
		State: StateNotInitialized,
		AvailableActions: []string{"init-project"},
	}
	
	suggestions := []*Suggestion{
		{
			Action: &workflow.WorkflowAction{
				ID:   "init-project",
				Name: "Initialize Project",
			},
			Priority:  workflow.PriorityP0,
			Reasoning: "Project needs to be initialized",
		},
		{
			Action: &workflow.WorkflowAction{
				ID:   "help",
				Name: "Show Help",
			},
			Priority:  workflow.PriorityP2,
			Reasoning: "Get help with commands",
		},
	}
	
	output := captureOutput(func() {
		display.DisplayWithSuggestions(ctx, suggestions)
	})
	
	assert.Contains(t, output, "Not Initialized")
	assert.Contains(t, output, "Recommended Actions")
	assert.Contains(t, output, "Initialize Project")
	assert.Contains(t, output, "Project needs to be initialized")
}

func TestProjectStateDisplay_DisplayProjectOverview_NotInitialized(t *testing.T) {
	display := NewProjectStateDisplay()
	
	ctx := &ProjectContext{
		State:       StateNotInitialized,
		ProjectPath: "/test/project",
	}
	
	output := captureOutput(func() {
		display.DisplayProjectOverview(ctx)
	})
	
	assert.Contains(t, output, "project")
	assert.Contains(t, output, "Not Initialized")
	assert.Contains(t, output, "/test/project")
}

func TestProjectStateDisplay_DisplayProjectOverview_WithEpic(t *testing.T) {
	display := NewProjectStateDisplay()
	
	ctx := &ProjectContext{
		State: StateEpicInProgress,
		CurrentEpic: &EpicContext{
			ID:               "EPIC-001",
			Title:            "Test Epic",
			Status:           "in_progress",
			Priority:         "high",
			Progress:         0.4,
			CompletedStories: 2,
			TotalStories:     5,
		},
		ProjectPath: "/test/project",
	}
	
	output := captureOutput(func() {
		display.DisplayProjectOverview(ctx)
	})
	
	assert.Contains(t, output, "Test Epic")
	assert.Contains(t, output, "EPIC-001")
	assert.Contains(t, output, "40.0%")
	assert.Contains(t, output, "2/5 stories")
	assert.Contains(t, output, "🔴") // High priority icon
}

func TestProjectStateDisplay_DisplayProjectOverview_WithIssues(t *testing.T) {
	display := NewProjectStateDisplay()
	
	ctx := &ProjectContext{
		State: StateProjectInitialized,
		Issues: []string{
			"Missing configuration file",
			"Corrupted state",
		},
	}
	
	output := captureOutput(func() {
		display.DisplayProjectOverview(ctx)
	})
	
	assert.Contains(t, output, "Issues (2)")
	assert.Contains(t, output, "Missing configuration file")
	assert.Contains(t, output, "Corrupted state")
}

func TestProjectStateDisplay_DisplayProjectOverview_ManyIssues(t *testing.T) {
	display := NewProjectStateDisplay()
	
	ctx := &ProjectContext{
		State: StateProjectInitialized,
		Issues: []string{"issue1", "issue2", "issue3", "issue4", "issue5", "issue6", "issue7"},
	}
	
	output := captureOutput(func() {
		display.DisplayProjectOverview(ctx)
	})
	
	assert.Contains(t, output, "Issues (7)")
	assert.Contains(t, output, "and 2 more") // Should limit to first 5
}

func TestProjectStateDisplay_SetWidth(t *testing.T) {
	display := NewProjectStateDisplay()
	
	// Test default width
	assert.Equal(t, 80, display.width)
	
	// Test setting custom width
	display.SetWidth(120)
	assert.Equal(t, 120, display.width)
}

func TestProjectStateDisplay_PrintCentered(t *testing.T) {
	display := NewProjectStateDisplay()
	display.SetWidth(20)
	
	tests := []struct {
		text     string
		expected string
	}{
		{
			text:     "Hello",
			expected: "       Hello        ",
		},
		{
			text:     "This is a very long text that exceeds width",
			expected: "This is a very long text that exceeds width",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			output := captureOutput(func() {
				display.printCentered(tt.text)
			})
			
			// Remove newline for comparison
			output = strings.TrimRight(output, "\n")
			assert.Equal(t, tt.expected, output)
		})
	}
}

func TestProjectStateDisplay_PrintSeparator(t *testing.T) {
	display := NewProjectStateDisplay()
	display.SetWidth(10)
	
	output := captureOutput(func() {
		display.printSeparator("=")
	})
	
	expected := "==========\n"
	assert.Equal(t, expected, output)
}

func TestProjectStateDisplay_CompleteWorkflow(t *testing.T) {
	display := NewProjectStateDisplay()
	
	// Test complete workflow with all components
	ctx := &ProjectContext{
		State:       StateTaskInProgress,
		ProjectPath: "/home/user/awesome-project",
		CurrentEpic: &EpicContext{
			ID:               "EPIC-001",
			Title:            "Build Amazing Feature",
			Status:           "in_progress",
			Priority:         "high",
			Progress:         0.67,
			CompletedStories: 2,
			TotalStories:     3,
		},
		CurrentStory: &StoryContext{
			ID:             "STORY-003",
			Title:          "Implement Core Logic",
			Status:         "in_progress",
			Priority:       "medium",
			CompletedTasks: 4,
			TotalTasks:     6,
		},
		CurrentTask: &TaskContext{
			ID:             "TASK-015",
			Title:          "Write Unit Tests",
			Status:         "in_progress",
			Priority:       "P1",
			EstimatedHours: 3,
		},
		AvailableActions: []string{
			"continue-task",
			"complete-task",
			"help",
		},
		Issues: []string{
			"Test coverage below 80%",
		},
	}
	
	output := captureOutput(func() {
		display.DisplayProjectOverview(ctx)
	})
	
	// Verify all major components are displayed
	assert.Contains(t, output, "awesome-project")
	assert.Contains(t, output, "Task In Progress")
	assert.Contains(t, output, "Build Amazing Feature")
	assert.Contains(t, output, "67.0%")
	assert.Contains(t, output, "2/3 stories")
	assert.Contains(t, output, "Implement Core Logic")
	assert.Contains(t, output, "4/6 tasks")
	assert.Contains(t, output, "Write Unit Tests")
	assert.Contains(t, output, "3 hours")
	assert.Contains(t, output, "Issues (1)")
	assert.Contains(t, output, "Test coverage")
	assert.Contains(t, output, "3 actions available")
}

func TestProjectStateDisplay_NilValues(t *testing.T) {
	display := NewProjectStateDisplay()
	
	// Test handling of nil contexts gracefully
	ctx := &ProjectContext{
		State:        StateEpicInProgress,
		CurrentEpic:  nil, // Should handle nil epic
		CurrentStory: nil, // Should handle nil story
		CurrentTask:  nil, // Should handle nil task
	}
	
	// Should not panic
	require.NotPanics(t, func() {
		output := captureOutput(func() {
			display.DisplayProjectOverview(ctx)
		})
		
		assert.Contains(t, output, "No active epic")
	})
}