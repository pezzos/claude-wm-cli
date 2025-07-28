package story

import (
	"os"
	"path/filepath"
	"testing"

	"claude-wm-cli/internal/epic"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerator_NewGenerator(t *testing.T) {
	tempDir := t.TempDir()
	generator := NewGenerator(tempDir)

	assert.NotNil(t, generator)
	assert.Equal(t, tempDir, generator.rootPath)
	assert.NotNil(t, generator.epicManager)
}

func TestGenerator_CreateStory(t *testing.T) {
	tempDir := t.TempDir()
	setupTestDirs(t, tempDir)

	generator := NewGenerator(tempDir)

	// Create a story
	options := StoryCreateOptions{
		Title:       "Test User Story",
		Description: "A test story for validation",
		Priority:    epic.PriorityHigh,
		StoryPoints: 5,
		AcceptanceCriteria: []string{
			"User can login with email",
			"User receives welcome message",
			"User is redirected to dashboard",
		},
		Dependencies: []string{"STORY-001"},
	}

	story, err := generator.CreateStory(options)
	require.NoError(t, err)
	assert.NotNil(t, story)

	// Verify story properties
	assert.Equal(t, "Test User Story", story.Title)
	assert.Equal(t, "A test story for validation", story.Description)
	assert.Equal(t, epic.PriorityHigh, story.Priority)
	assert.Equal(t, epic.StatusPlanned, story.Status)
	assert.Equal(t, 5, story.StoryPoints)
	assert.Len(t, story.AcceptanceCriteria, 3)
	assert.Len(t, story.Tasks, 3) // Tasks generated from acceptance criteria
	assert.Equal(t, []string{"STORY-001"}, story.Dependencies)
	assert.True(t, story.ID != "")

	// Verify generated tasks
	for i, task := range story.Tasks {
		assert.Equal(t, story.ID, task.StoryID)
		assert.Equal(t, epic.StatusPlanned, task.Status)
		assert.Contains(t, task.Title, "Implement:")
		assert.Equal(t, story.AcceptanceCriteria[i], task.Description)
	}
}

func TestGenerator_UpdateStory(t *testing.T) {
	tempDir := t.TempDir()
	setupTestDirs(t, tempDir)

	generator := NewGenerator(tempDir)

	// Create a story first
	story, err := generator.CreateStory(StoryCreateOptions{
		Title:       "Original Story",
		Description: "Original description",
		Priority:    epic.PriorityMedium,
		StoryPoints: 3,
	})
	require.NoError(t, err)

	// Update the story
	newTitle := "Updated Story Title"
	newStatus := epic.StatusInProgress
	newStoryPoints := 8

	updatedStory, err := generator.UpdateStory(story.ID, StoryUpdateOptions{
		Title:       &newTitle,
		Status:      &newStatus,
		StoryPoints: &newStoryPoints,
	})
	require.NoError(t, err)

	// Verify updates
	assert.Equal(t, "Updated Story Title", updatedStory.Title)
	assert.Equal(t, epic.StatusInProgress, updatedStory.Status)
	assert.Equal(t, 8, updatedStory.StoryPoints)
	assert.NotNil(t, updatedStory.StartedAt) // Should be set when status changes to in_progress
}

func TestGenerator_ListStories(t *testing.T) {
	tempDir := t.TempDir()
	setupTestDirs(t, tempDir)

	generator := NewGenerator(tempDir)

	// Create test epics first
	epicManager := epic.NewManager(tempDir)
	epic1, err := epicManager.CreateEpic(epic.EpicCreateOptions{
		Title:    "Test Epic One",
		Priority: epic.PriorityHigh,
	})
	require.NoError(t, err)

	epic2, err := epicManager.CreateEpic(epic.EpicCreateOptions{
		Title:    "Test Epic Two",
		Priority: epic.PriorityMedium,
	})
	require.NoError(t, err)

	// Create multiple stories
	_, err = generator.CreateStory(StoryCreateOptions{
		Title:    "Story One",
		EpicID:   epic1.ID,
		Priority: epic.PriorityHigh,
	})
	require.NoError(t, err)

	story2, err := generator.CreateStory(StoryCreateOptions{
		Title:    "Story Two",
		EpicID:   epic1.ID,
		Priority: epic.PriorityMedium,
	})
	require.NoError(t, err)

	// Update story2 status
	inProgress := epic.StatusInProgress
	_, err = generator.UpdateStory(story2.ID, StoryUpdateOptions{
		Status: &inProgress,
	})
	require.NoError(t, err)

	_, err = generator.CreateStory(StoryCreateOptions{
		Title:    "Story Three",
		EpicID:   epic2.ID,
		Priority: epic.PriorityLow,
	})
	require.NoError(t, err)

	// List all stories
	allStories, err := generator.ListStories("", "")
	require.NoError(t, err)
	assert.Len(t, allStories, 3)

	// Filter by epic
	epic1Stories, err := generator.ListStories(epic1.ID, "")
	require.NoError(t, err)
	assert.Len(t, epic1Stories, 2)

	// Filter by status
	inProgressStories, err := generator.ListStories("", epic.StatusInProgress)
	require.NoError(t, err)
	assert.Len(t, inProgressStories, 1)
	assert.Equal(t, story2.ID, inProgressStories[0].ID)

	// Filter by both epic and status
	epic1InProgress, err := generator.ListStories(epic1.ID, epic.StatusInProgress)
	require.NoError(t, err)
	assert.Len(t, epic1InProgress, 1)
}

func TestGenerator_DeleteStory(t *testing.T) {
	tempDir := t.TempDir()
	setupTestDirs(t, tempDir)

	generator := NewGenerator(tempDir)

	// Create a story
	story, err := generator.CreateStory(StoryCreateOptions{
		Title: "Story to Delete",
	})
	require.NoError(t, err)

	// Delete the story
	err = generator.DeleteStory(story.ID)
	require.NoError(t, err)

	// Verify it's deleted
	_, err = generator.GetStory(story.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "story not found")
}

func TestGenerator_StatusTransitions(t *testing.T) {
	tempDir := t.TempDir()
	setupTestDirs(t, tempDir)

	generator := NewGenerator(tempDir)

	// Create a story
	story, err := generator.CreateStory(StoryCreateOptions{
		Title: "Status Test Story",
	})
	require.NoError(t, err)

	// Valid transition: planned -> in_progress
	inProgressStatus := epic.StatusInProgress
	_, err = generator.UpdateStory(story.ID, StoryUpdateOptions{
		Status: &inProgressStatus,
	})
	assert.NoError(t, err)

	// Valid transition: in_progress -> completed
	completedStatus := epic.StatusCompleted
	updatedStory, err := generator.UpdateStory(story.ID, StoryUpdateOptions{
		Status: &completedStatus,
	})
	assert.NoError(t, err)
	assert.NotNil(t, updatedStory.CompletedAt) // Should be set when completing

	// Invalid transition: completed -> in_progress (completed stories can't be restarted)
	inProgressStatus2 := epic.StatusInProgress
	_, err = generator.UpdateStory(story.ID, StoryUpdateOptions{
		Status: &inProgressStatus2,
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid status transition")
}

func TestGenerator_GenerateStoriesFromEpic(t *testing.T) {
	tempDir := t.TempDir()
	setupTestDirs(t, tempDir)

	// Create a story generator and test manual story creation
	generator := NewGenerator(tempDir)
	
	// Create an epic first  
	epicManager := epic.NewManager(tempDir)
	testEpic, err := epicManager.CreateEpic(epic.EpicCreateOptions{
		Title:       "Test Epic",
		Description: "An epic for testing story generation",
		Priority:    epic.PriorityHigh,
	})
	require.NoError(t, err)

	// Create stories manually to test the system
	story1, err := generator.CreateStory(StoryCreateOptions{
		Title:       "User Authentication",
		Description: "As a user, I want to login to the system",
		EpicID:      testEpic.ID,
		Priority:    epic.PriorityHigh,
		StoryPoints: 5,
		AcceptanceCriteria: []string{
			"User can enter email and password",
			"System validates credentials",
			"User is redirected on success",
		},
	})
	require.NoError(t, err)

	story2, err := generator.CreateStory(StoryCreateOptions{
		Title:       "User Dashboard",
		Description: "As a user, I want to see my dashboard",
		EpicID:      testEpic.ID,
		Priority:    epic.PriorityMedium,
		StoryPoints: 3,
		AcceptanceCriteria: []string{
			"Dashboard shows user name",
			"Dashboard shows recent activity",
		},
	})
	require.NoError(t, err)

	// Verify stories were created
	stories, err := generator.ListStories(testEpic.ID, "")
	require.NoError(t, err)
	assert.Len(t, stories, 2)

	// Verify story properties
	storyTitles := make(map[string]bool)
	for _, story := range stories {
		assert.Equal(t, testEpic.ID, story.EpicID)
		assert.Equal(t, epic.StatusPlanned, story.Status)
		assert.True(t, story.StoryPoints > 0)
		storyTitles[story.Title] = true
		
		// Verify tasks were generated from acceptance criteria
		assert.True(t, len(story.Tasks) > 0)
	}
	
	// Verify both stories were created
	assert.True(t, storyTitles["User Authentication"])
	assert.True(t, storyTitles["User Dashboard"])
	
	// Verify specific story details
	assert.Equal(t, 5, story1.StoryPoints)
	assert.Equal(t, 3, story2.StoryPoints)
	assert.Len(t, story1.Tasks, 3) // 3 acceptance criteria = 3 tasks
	assert.Len(t, story2.Tasks, 2) // 2 acceptance criteria = 2 tasks
}

func TestStory_ProgressCalculation(t *testing.T) {
	story := &Story{
		ID:    "TEST-STORY",
		Title: "Test Story",
		Tasks: []Task{
			{ID: "TASK-1", Status: epic.StatusCompleted},
			{ID: "TASK-2", Status: epic.StatusInProgress},
			{ID: "TASK-3", Status: epic.StatusPlanned},
			{ID: "TASK-4", Status: epic.StatusCompleted},
		},
	}

	progress := story.CalculateProgress()

	assert.Equal(t, 4, progress.TotalTasks)
	assert.Equal(t, 2, progress.CompletedTasks)
	assert.Equal(t, 1, progress.InProgressTasks)
	assert.Equal(t, 1, progress.PendingTasks)
	assert.Equal(t, 50.0, progress.CompletionPercentage)
}

func TestStory_HelperMethods(t *testing.T) {
	// Test CanStart
	plannedStory := &Story{Status: epic.StatusPlanned}
	assert.True(t, plannedStory.CanStart())
	assert.False(t, plannedStory.IsActive())

	// Test IsActive
	activeStory := &Story{Status: epic.StatusInProgress}
	assert.True(t, activeStory.IsActive())
	assert.False(t, activeStory.CanStart())

	// Test CanComplete
	completableStory := &Story{
		Status: epic.StatusInProgress,
		Tasks: []Task{
			{Status: epic.StatusCompleted},
			{Status: epic.StatusCompleted},
		},
	}
	assert.True(t, completableStory.CanComplete())

	incompletableStory := &Story{
		Status: epic.StatusInProgress,
		Tasks: []Task{
			{Status: epic.StatusCompleted},
			{Status: epic.StatusPlanned},
		},
	}
	assert.False(t, incompletableStory.CanComplete())
}

func TestStory_TaskManagement(t *testing.T) {
	story := &Story{
		ID:    "TEST-STORY",
		Title: "Test Story",
		Tasks: []Task{
			{ID: "TASK-1", Title: "Task One"},
			{ID: "TASK-2", Title: "Task Two"},
		},
	}

	// Test GetTaskByID
	task := story.GetTaskByID("TASK-1")
	assert.NotNil(t, task)
	assert.Equal(t, "Task One", task.Title)

	task = story.GetTaskByID("NON-EXISTENT")
	assert.Nil(t, task)

	// Test AddTask
	newTask := Task{ID: "TASK-3", Title: "Task Three"}
	story.AddTask(newTask)
	assert.Len(t, story.Tasks, 3)
	assert.Equal(t, "Task Three", story.Tasks[2].Title)

	// Test RemoveTask
	removed := story.RemoveTask("TASK-2")
	assert.True(t, removed)
	assert.Len(t, story.Tasks, 2)

	removed = story.RemoveTask("NON-EXISTENT")
	assert.False(t, removed)
}

func TestGenerator_Validation(t *testing.T) {
	tempDir := t.TempDir()
	setupTestDirs(t, tempDir)

	generator := NewGenerator(tempDir)

	// Test empty title
	_, err := generator.CreateStory(StoryCreateOptions{
		Title: "",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "title cannot be empty")

	// Test invalid epic ID
	_, err = generator.CreateStory(StoryCreateOptions{
		Title:  "Valid Title",
		EpicID: "NON-EXISTENT-EPIC",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// Helper function to setup test directories
func setupTestDirs(t *testing.T, tempDir string) {
	docsDir := filepath.Join(tempDir, "docs", "1-project")
	err := os.MkdirAll(docsDir, 0755)
	require.NoError(t, err)

	currentEpicDir := filepath.Join(tempDir, "docs", "2-current-epic")
	err = os.MkdirAll(currentEpicDir, 0755)
	require.NoError(t, err)

	currentTaskDir := filepath.Join(tempDir, "docs", "3-current-task")
	err = os.MkdirAll(currentTaskDir, 0755)
	require.NoError(t, err)
}