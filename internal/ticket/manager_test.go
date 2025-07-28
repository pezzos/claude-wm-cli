package ticket

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestManager_NewManager(t *testing.T) {
	tempDir := t.TempDir()
	setupTestDirs(t, tempDir)

	manager := NewManager(tempDir)
	assert.NotNil(t, manager)
	assert.Equal(t, tempDir, manager.rootPath)
	assert.NotNil(t, manager.epicManager)
}

func TestManager_CreateTicket(t *testing.T) {
	tempDir := t.TempDir()
	setupTestDirs(t, tempDir)

	manager := NewManager(tempDir)

	// Test basic ticket creation
	options := TicketCreateOptions{
		Title:       "Test Bug Fix",
		Description: "Fix the test bug",
		Type:        TicketTypeBug,
		Priority:    TicketPriorityHigh,
	}

	ticket, err := manager.CreateTicket(options)
	require.NoError(t, err)
	require.NotNil(t, ticket)

	assert.Equal(t, "Test Bug Fix", ticket.Title)
	assert.Equal(t, "Fix the test bug", ticket.Description)
	assert.Equal(t, TicketTypeBug, ticket.Type)
	assert.Equal(t, TicketPriorityHigh, ticket.Priority)
	assert.Equal(t, TicketStatusOpen, ticket.Status)
	assert.NotEmpty(t, ticket.ID)
	assert.False(t, ticket.CreatedAt.IsZero())
	assert.False(t, ticket.UpdatedAt.IsZero())
}

func TestManager_CreateTicketValidation(t *testing.T) {
	tempDir := t.TempDir()
	setupTestDirs(t, tempDir)

	manager := NewManager(tempDir)

	// Test empty title
	options := TicketCreateOptions{
		Title: "",
	}

	_, err := manager.CreateTicket(options)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "title cannot be empty")

	// Test invalid type
	options = TicketCreateOptions{
		Title: "Valid Title",
		Type:  TicketType("invalid"),
	}

	_, err = manager.CreateTicket(options)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid ticket type")

	// Test invalid priority
	options = TicketCreateOptions{
		Title:    "Valid Title",
		Priority: TicketPriority("invalid"),
	}

	_, err = manager.CreateTicket(options)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid ticket priority")
}

func TestManager_UpdateTicket(t *testing.T) {
	tempDir := t.TempDir()
	setupTestDirs(t, tempDir)

	manager := NewManager(tempDir)

	// Create a ticket first
	options := TicketCreateOptions{
		Title:    "Original Title",
		Type:     TicketTypeBug,
		Priority: TicketPriorityMedium,
	}

	ticket, err := manager.CreateTicket(options)
	require.NoError(t, err)

	// Update the ticket
	newTitle := "Updated Title"
	newPriority := TicketPriorityHigh
	updateOptions := TicketUpdateOptions{
		Title:    &newTitle,
		Priority: &newPriority,
	}

	updatedTicket, err := manager.UpdateTicket(ticket.ID, updateOptions)
	require.NoError(t, err)

	assert.Equal(t, "Updated Title", updatedTicket.Title)
	assert.Equal(t, TicketPriorityHigh, updatedTicket.Priority)
	assert.True(t, updatedTicket.UpdatedAt.After(ticket.UpdatedAt))
}

func TestManager_StatusTransitions(t *testing.T) {
	tempDir := t.TempDir()
	setupTestDirs(t, tempDir)

	manager := NewManager(tempDir)

	// Create a ticket
	options := TicketCreateOptions{
		Title: "Status Test Ticket",
		Type:  TicketTypeTask,
	}

	ticket, err := manager.CreateTicket(options)
	require.NoError(t, err)
	assert.Equal(t, TicketStatusOpen, ticket.Status)

	// Test valid transition: open -> in_progress
	newStatus := TicketStatusInProgress
	updateOptions := TicketUpdateOptions{
		Status: &newStatus,
	}

	updatedTicket, err := manager.UpdateTicket(ticket.ID, updateOptions)
	require.NoError(t, err)
	assert.Equal(t, TicketStatusInProgress, updatedTicket.Status)
	assert.NotNil(t, updatedTicket.StartedAt)

	// Test valid transition: in_progress -> resolved
	newStatus = TicketStatusResolved
	updateOptions = TicketUpdateOptions{
		Status: &newStatus,
	}

	updatedTicket, err = manager.UpdateTicket(ticket.ID, updateOptions)
	require.NoError(t, err)
	assert.Equal(t, TicketStatusResolved, updatedTicket.Status)
	assert.NotNil(t, updatedTicket.ResolvedAt)

	// Test valid transition: resolved -> closed
	newStatus = TicketStatusClosed
	updateOptions = TicketUpdateOptions{
		Status: &newStatus,
	}

	updatedTicket, err = manager.UpdateTicket(ticket.ID, updateOptions)
	require.NoError(t, err)
	assert.Equal(t, TicketStatusClosed, updatedTicket.Status)
	assert.NotNil(t, updatedTicket.ClosedAt)
}

func TestManager_InvalidStatusTransition(t *testing.T) {
	tempDir := t.TempDir()
	setupTestDirs(t, tempDir)

	manager := NewManager(tempDir)

	// Create a ticket and move it to resolved
	options := TicketCreateOptions{
		Title: "Status Test Ticket",
	}

	ticket, err := manager.CreateTicket(options)
	require.NoError(t, err)

	// Move to resolved first
	newStatus := TicketStatusInProgress
	updateOptions := TicketUpdateOptions{
		Status: &newStatus,
	}
	updatedTicket, err := manager.UpdateTicket(ticket.ID, updateOptions)
	require.NoError(t, err)

	newStatus = TicketStatusResolved
	updateOptions = TicketUpdateOptions{
		Status: &newStatus,
	}
	updatedTicket, err = manager.UpdateTicket(ticket.ID, updateOptions)
	require.NoError(t, err)

	// Try invalid transition: resolved -> open (should fail)
	newStatus = TicketStatusOpen
	updateOptions = TicketUpdateOptions{
		Status: &newStatus,
	}

	_, err = manager.UpdateTicket(updatedTicket.ID, updateOptions)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid status transition")
}

func TestManager_ListTickets(t *testing.T) {
	tempDir := t.TempDir()
	setupTestDirs(t, tempDir)

	manager := NewManager(tempDir)

	// Create multiple tickets
	tickets := []TicketCreateOptions{
		{
			Title:    "High Priority Bug",
			Type:     TicketTypeBug,
			Priority: TicketPriorityHigh,
		},
		{
			Title:    "Medium Priority Task",
			Type:     TicketTypeTask,
			Priority: TicketPriorityMedium,
		},
		{
			Title:    "Urgent Feature",
			Type:     TicketTypeFeature,
			Priority: TicketPriorityUrgent,
		},
	}

	var createdTickets []*Ticket
	for _, options := range tickets {
		ticket, err := manager.CreateTicket(options)
		require.NoError(t, err)
		createdTickets = append(createdTickets, ticket)
	}

	// Test listing all tickets
	allTickets, err := manager.ListTickets(TicketListOptions{})
	require.NoError(t, err)
	assert.Len(t, allTickets, 3)

	// Check sorting (should be by priority: urgent > high > medium)
	assert.Equal(t, TicketPriorityUrgent, allTickets[0].Priority)
	assert.Equal(t, TicketPriorityHigh, allTickets[1].Priority)
	assert.Equal(t, TicketPriorityMedium, allTickets[2].Priority)

	// Test filtering by priority
	highPriorityTickets, err := manager.ListTickets(TicketListOptions{
		Priority: TicketPriorityHigh,
	})
	require.NoError(t, err)
	assert.Len(t, highPriorityTickets, 1)
	assert.Equal(t, "High Priority Bug", highPriorityTickets[0].Title)

	// Test filtering by type
	bugTickets, err := manager.ListTickets(TicketListOptions{
		Type: TicketTypeBug,
	})
	require.NoError(t, err)
	assert.Len(t, bugTickets, 1)
	assert.Equal(t, "High Priority Bug", bugTickets[0].Title)
}

func TestManager_CurrentTicket(t *testing.T) {
	tempDir := t.TempDir()
	setupTestDirs(t, tempDir)

	manager := NewManager(tempDir)

	// Initially no current ticket
	currentTicket, err := manager.GetCurrentTicket()
	require.NoError(t, err)
	assert.Nil(t, currentTicket)

	// Create a ticket
	options := TicketCreateOptions{
		Title: "Current Test Ticket",
	}

	ticket, err := manager.CreateTicket(options)
	require.NoError(t, err)

	// Set as current
	selectedTicket, err := manager.SetCurrentTicket(ticket.ID)
	require.NoError(t, err)
	assert.Equal(t, ticket.ID, selectedTicket.ID)
	assert.Equal(t, TicketStatusInProgress, selectedTicket.Status) // Should auto-start

	// Get current ticket
	currentTicket, err = manager.GetCurrentTicket()
	require.NoError(t, err)
	assert.NotNil(t, currentTicket)
	assert.Equal(t, ticket.ID, currentTicket.ID)

	// Clear current ticket
	_, err = manager.SetCurrentTicket("")
	require.NoError(t, err)

	currentTicket, err = manager.GetCurrentTicket()
	require.NoError(t, err)
	assert.Nil(t, currentTicket)
}

func TestManager_GetTicketStats(t *testing.T) {
	tempDir := t.TempDir()
	setupTestDirs(t, tempDir)

	manager := NewManager(tempDir)

	// Create tickets with different properties
	tickets := []TicketCreateOptions{
		{Title: "Bug 1", Type: TicketTypeBug, Priority: TicketPriorityHigh},
		{Title: "Bug 2", Type: TicketTypeBug, Priority: TicketPriorityMedium},
		{Title: "Feature 1", Type: TicketTypeFeature, Priority: TicketPriorityHigh},
		{Title: "Task 1", Type: TicketTypeTask, Priority: TicketPriorityLow},
	}

	var createdTickets []*Ticket
	for _, options := range tickets {
		ticket, err := manager.CreateTicket(options)
		require.NoError(t, err)
		createdTickets = append(createdTickets, ticket)
	}

	// Resolve one ticket to test resolution metrics
	// First move to in_progress, then to resolved
	inProgressStatus := TicketStatusInProgress
	_, err := manager.UpdateTicket(createdTickets[0].ID, TicketUpdateOptions{
		Status: &inProgressStatus,
	})
	require.NoError(t, err)

	resolvedStatus := TicketStatusResolved
	_, err = manager.UpdateTicket(createdTickets[0].ID, TicketUpdateOptions{
		Status: &resolvedStatus,
	})
	require.NoError(t, err)

	// Get stats
	stats, err := manager.GetTicketStats()
	require.NoError(t, err)

	assert.Equal(t, 4, stats.TotalTickets)
	assert.Equal(t, 3, stats.ByStatus[TicketStatusOpen])
	assert.Equal(t, 1, stats.ByStatus[TicketStatusResolved])
	assert.Equal(t, 2, stats.ByPriority[TicketPriorityHigh])
	assert.Equal(t, 1, stats.ByPriority[TicketPriorityMedium])
	assert.Equal(t, 1, stats.ByPriority[TicketPriorityLow])
	assert.Equal(t, 2, stats.ByType[TicketTypeBug])
	assert.Equal(t, 1, stats.ByType[TicketTypeFeature])
	assert.Equal(t, 1, stats.ByType[TicketTypeTask])
}

func TestManager_DeleteTicket(t *testing.T) {
	tempDir := t.TempDir()
	setupTestDirs(t, tempDir)

	manager := NewManager(tempDir)

	// Create a ticket
	options := TicketCreateOptions{
		Title: "Ticket to Delete",
	}

	ticket, err := manager.CreateTicket(options)
	require.NoError(t, err)

	// Set as current
	_, err = manager.SetCurrentTicket(ticket.ID)
	require.NoError(t, err)

	// Delete the ticket
	err = manager.DeleteTicket(ticket.ID)
	require.NoError(t, err)

	// Verify it's gone
	_, err = manager.GetTicket(ticket.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	// Verify current ticket is cleared
	currentTicket, err := manager.GetCurrentTicket()
	require.NoError(t, err)
	assert.Nil(t, currentTicket)
}

func TestManager_GenerateTicketID(t *testing.T) {
	tempDir := t.TempDir()
	setupTestDirs(t, tempDir)

	manager := NewManager(tempDir)
	collection := &TicketCollection{
		Tickets: make(map[string]*Ticket),
	}

	// Test basic ID generation
	id1 := manager.generateTicketID("Fix Login Bug", collection)
	assert.Equal(t, "TICKET-001-FIX-LOGIN-BUG", id1)

	// Test uniqueness
	collection.Tickets[id1] = &Ticket{ID: id1}
	id2 := manager.generateTicketID("Fix Login Bug", collection)
	assert.Equal(t, "TICKET-002-FIX-LOGIN-BUG", id2)

	// Test special character removal
	id3 := manager.generateTicketID("Fix @#$% Bug!!!", collection)
	assert.Equal(t, "TICKET-001-FIX--BUG", id3)

	// Test empty title fallback
	id4 := manager.generateTicketID("", collection)
	assert.Equal(t, "TICKET-001-TICKET", id4)
}

func TestTicketTypes_IsValid(t *testing.T) {
	// Test valid statuses
	assert.True(t, TicketStatusOpen.IsValid())
	assert.True(t, TicketStatusInProgress.IsValid())
	assert.True(t, TicketStatusResolved.IsValid())
	assert.True(t, TicketStatusClosed.IsValid())

	// Test invalid status
	assert.False(t, TicketStatus("invalid").IsValid())

	// Test valid priorities
	assert.True(t, TicketPriorityLow.IsValid())
	assert.True(t, TicketPriorityMedium.IsValid())
	assert.True(t, TicketPriorityHigh.IsValid())
	assert.True(t, TicketPriorityCritical.IsValid())
	assert.True(t, TicketPriorityUrgent.IsValid())

	// Test invalid priority
	assert.False(t, TicketPriority("invalid").IsValid())

	// Test valid types
	assert.True(t, TicketTypeBug.IsValid())
	assert.True(t, TicketTypeFeature.IsValid())
	assert.True(t, TicketTypeInterruption.IsValid())
	assert.True(t, TicketTypeTask.IsValid())
	assert.True(t, TicketTypeSupport.IsValid())

	// Test invalid type
	assert.False(t, TicketType("invalid").IsValid())
}

func TestManager_PersistenceAndRecovery(t *testing.T) {
	tempDir := t.TempDir()
	setupTestDirs(t, tempDir)

	manager := NewManager(tempDir)

	// Create a ticket
	options := TicketCreateOptions{
		Title:       "Persistence Test",
		Description: "Test persistence",
		Type:        TicketTypeBug,
		Priority:    TicketPriorityHigh,
	}

	ticket, err := manager.CreateTicket(options)
	require.NoError(t, err)

	// Create new manager (simulate restart)
	manager2 := NewManager(tempDir)

	// Retrieve the ticket
	retrievedTicket, err := manager2.GetTicket(ticket.ID)
	require.NoError(t, err)

	assert.Equal(t, ticket.ID, retrievedTicket.ID)
	assert.Equal(t, ticket.Title, retrievedTicket.Title)
	assert.Equal(t, ticket.Description, retrievedTicket.Description)
	assert.Equal(t, ticket.Type, retrievedTicket.Type)
	assert.Equal(t, ticket.Priority, retrievedTicket.Priority)
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
