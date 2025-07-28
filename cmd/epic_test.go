package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"claude-wm-cli/internal/epic"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEpicManager_CreateEpic(t *testing.T) {
	tempDir := t.TempDir()

	// Create the directory structure
	docsDir := filepath.Join(tempDir, "docs", "1-project")
	err := os.MkdirAll(docsDir, 0755)
	require.NoError(t, err)

	manager := epic.NewManager(tempDir)

	options := epic.EpicCreateOptions{
		Title:       "Test Epic",
		Description: "Test description",
		Priority:    epic.PriorityHigh,
		Duration:    "2 weeks",
		Tags:        []string{"test", "validation"},
	}

	createdEpic, err := manager.CreateEpic(options)
	require.NoError(t, err)
	assert.NotNil(t, createdEpic)

	assert.Equal(t, "Test Epic", createdEpic.Title)
	assert.Equal(t, "Test description", createdEpic.Description)
	assert.Equal(t, epic.PriorityHigh, createdEpic.Priority)
	assert.Equal(t, epic.StatusPlanned, createdEpic.Status)
	assert.Equal(t, "2 weeks", createdEpic.Duration)
	assert.Equal(t, []string{"test", "validation"}, createdEpic.Tags)
	assert.True(t, createdEpic.ID != "")
}

func TestEpicManager_ListEpics(t *testing.T) {
	tempDir := t.TempDir()

	// Create the directory structure
	docsDir := filepath.Join(tempDir, "docs", "1-project")
	err := os.MkdirAll(docsDir, 0755)
	require.NoError(t, err)

	manager := epic.NewManager(tempDir)

	// Create multiple epics
	epic1, err := manager.CreateEpic(epic.EpicCreateOptions{
		Title:    "Epic One",
		Priority: epic.PriorityHigh,
	})
	require.NoError(t, err)

	epic2, err := manager.CreateEpic(epic.EpicCreateOptions{
		Title:    "Epic Two",
		Priority: epic.PriorityMedium,
	})
	require.NoError(t, err)

	// List all epics
	epics, err := manager.ListEpics(epic.EpicListOptions{})
	require.NoError(t, err)
	assert.Len(t, epics, 2)

	// Check that epics are sorted by creation date (newest first)
	assert.Equal(t, epic2.ID, epics[0].ID) // Latest created should be first
	assert.Equal(t, epic1.ID, epics[1].ID)
}

func TestEpicManager_UpdateEpic(t *testing.T) {
	tempDir := t.TempDir()

	// Create the directory structure
	docsDir := filepath.Join(tempDir, "docs", "1-project")
	err := os.MkdirAll(docsDir, 0755)
	require.NoError(t, err)

	manager := epic.NewManager(tempDir)

	// Create an epic
	createdEpic, err := manager.CreateEpic(epic.EpicCreateOptions{
		Title:    "Original Title",
		Priority: epic.PriorityMedium,
	})
	require.NoError(t, err)

	// Update the epic
	newTitle := "Updated Title"
	newPriority := epic.PriorityHigh
	newStatus := epic.StatusInProgress

	updatedEpic, err := manager.UpdateEpic(createdEpic.ID, epic.EpicUpdateOptions{
		Title:    &newTitle,
		Priority: &newPriority,
		Status:   &newStatus,
	})
	require.NoError(t, err)

	assert.Equal(t, "Updated Title", updatedEpic.Title)
	assert.Equal(t, epic.PriorityHigh, updatedEpic.Priority)
	assert.Equal(t, epic.StatusInProgress, updatedEpic.Status)
	assert.NotNil(t, updatedEpic.StartDate) // Should be set when status changes to in_progress
}

func TestEpicManager_SelectEpic(t *testing.T) {
	tempDir := t.TempDir()

	// Create the directory structure
	docsDir := filepath.Join(tempDir, "docs", "1-project")
	err := os.MkdirAll(docsDir, 0755)
	require.NoError(t, err)

	manager := epic.NewManager(tempDir)

	// Create an epic
	createdEpic, err := manager.CreateEpic(epic.EpicCreateOptions{
		Title:    "Selectable Epic",
		Priority: epic.PriorityHigh,
	})
	require.NoError(t, err)

	// Select the epic
	selectedEpic, err := manager.SelectEpic(createdEpic.ID)
	require.NoError(t, err)

	assert.Equal(t, createdEpic.ID, selectedEpic.ID)
	assert.Equal(t, epic.StatusInProgress, selectedEpic.Status) // Should auto-start

	// Verify it's set as current epic
	currentEpic, err := manager.GetCurrentEpic()
	require.NoError(t, err)
	assert.Equal(t, createdEpic.ID, currentEpic.ID)
}

func TestEpicManager_GetEpic(t *testing.T) {
	tempDir := t.TempDir()

	// Create the directory structure
	docsDir := filepath.Join(tempDir, "docs", "1-project")
	err := os.MkdirAll(docsDir, 0755)
	require.NoError(t, err)

	manager := epic.NewManager(tempDir)

	// Create an epic
	createdEpic, err := manager.CreateEpic(epic.EpicCreateOptions{
		Title:       "Get Epic Test",
		Description: "Epic to retrieve",
		Priority:    epic.PriorityMedium,
	})
	require.NoError(t, err)

	// Get the epic
	retrievedEpic, err := manager.GetEpic(createdEpic.ID)
	require.NoError(t, err)

	assert.Equal(t, createdEpic.ID, retrievedEpic.ID)
	assert.Equal(t, "Get Epic Test", retrievedEpic.Title)
	assert.Equal(t, "Epic to retrieve", retrievedEpic.Description)
	assert.Equal(t, epic.PriorityMedium, retrievedEpic.Priority)

	// Test getting non-existent epic
	_, err = manager.GetEpic("NON-EXISTENT")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "epic not found")
}

func TestEpicManager_DeleteEpic(t *testing.T) {
	tempDir := t.TempDir()

	// Create the directory structure
	docsDir := filepath.Join(tempDir, "docs", "1-project")
	err := os.MkdirAll(docsDir, 0755)
	require.NoError(t, err)

	manager := epic.NewManager(tempDir)

	// Create an epic
	createdEpic, err := manager.CreateEpic(epic.EpicCreateOptions{
		Title: "Epic to Delete",
	})
	require.NoError(t, err)

	// Select it as current
	_, err = manager.SelectEpic(createdEpic.ID)
	require.NoError(t, err)

	// Delete the epic
	err = manager.DeleteEpic(createdEpic.ID)
	require.NoError(t, err)

	// Verify it's deleted
	_, err = manager.GetEpic(createdEpic.ID)
	assert.Error(t, err)

	// Verify current epic is cleared
	_, err = manager.GetCurrentEpic()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no epic is currently active")
}

func TestEpicValidation(t *testing.T) {
	tempDir := t.TempDir()

	// Create the directory structure
	docsDir := filepath.Join(tempDir, "docs", "1-project")
	err := os.MkdirAll(docsDir, 0755)
	require.NoError(t, err)

	manager := epic.NewManager(tempDir)

	// Test empty title
	_, err = manager.CreateEpic(epic.EpicCreateOptions{
		Title: "",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "title cannot be empty")

	// Test invalid priority
	_, err = manager.CreateEpic(epic.EpicCreateOptions{
		Title:    "Valid Title",
		Priority: "invalid",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid priority")
}

func TestEpicStatusTransitions(t *testing.T) {
	tempDir := t.TempDir()

	// Create the directory structure
	docsDir := filepath.Join(tempDir, "docs", "1-project")
	err := os.MkdirAll(docsDir, 0755)
	require.NoError(t, err)

	manager := epic.NewManager(tempDir)

	// Create an epic
	createdEpic, err := manager.CreateEpic(epic.EpicCreateOptions{
		Title: "Status Test Epic",
	})
	require.NoError(t, err)

	// Valid transition: planned -> in_progress
	inProgressStatus := epic.StatusInProgress
	_, err = manager.UpdateEpic(createdEpic.ID, epic.EpicUpdateOptions{
		Status: &inProgressStatus,
	})
	assert.NoError(t, err)

	// Valid transition: in_progress -> completed
	completedStatus := epic.StatusCompleted
	updatedEpic, err := manager.UpdateEpic(createdEpic.ID, epic.EpicUpdateOptions{
		Status: &completedStatus,
	})
	assert.NoError(t, err)
	assert.NotNil(t, updatedEpic.EndDate) // Should be set when completing

	// Invalid transition: completed -> in_progress (completed epics can't be restarted)
	inProgressStatus2 := epic.StatusInProgress
	_, err = manager.UpdateEpic(createdEpic.ID, epic.EpicUpdateOptions{
		Status: &inProgressStatus2,
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid status transition")
}

func TestEpicHelperFunctions(t *testing.T) {
	// Test priority validation
	assert.True(t, epic.PriorityHigh.IsValid())
	assert.True(t, epic.PriorityMedium.IsValid())
	assert.True(t, epic.PriorityLow.IsValid())
	assert.True(t, epic.PriorityCritical.IsValid())
	assert.False(t, epic.Priority("invalid").IsValid())

	// Test status validation
	assert.True(t, epic.StatusPlanned.IsValid())
	assert.True(t, epic.StatusInProgress.IsValid())
	assert.True(t, epic.StatusCompleted.IsValid())
	assert.True(t, epic.StatusCancelled.IsValid())
	assert.True(t, epic.StatusOnHold.IsValid())
	assert.False(t, epic.Status("invalid").IsValid())

	// Test epic state checks
	plannedEpic := &epic.Epic{Status: epic.StatusPlanned}
	assert.True(t, plannedEpic.CanStart())
	assert.False(t, plannedEpic.IsActive())
	assert.False(t, plannedEpic.CanComplete())

	inProgressEpic := &epic.Epic{Status: epic.StatusInProgress}
	assert.False(t, inProgressEpic.CanStart())
	assert.True(t, inProgressEpic.IsActive())
	// Set progress to 100% for completion test
	inProgressEpic.Progress.CompletionPercentage = 100
	assert.True(t, inProgressEpic.CanComplete())

	completedEpic := &epic.Epic{Status: epic.StatusCompleted}
	assert.False(t, completedEpic.CanStart())
	assert.False(t, completedEpic.IsActive())
	assert.False(t, completedEpic.CanComplete())
}
