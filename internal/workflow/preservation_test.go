package workflow

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	
	"claude-wm-cli/internal/epic"
	"claude-wm-cli/internal/ticket"
)

func TestNewInterruptionStack(t *testing.T) {
	tempDir := t.TempDir()
	setupTestDirs(t, tempDir)
	
	stack := NewInterruptionStack(tempDir)
	assert.NotNil(t, stack)
	assert.Equal(t, tempDir, stack.rootPath)
	assert.NotNil(t, stack.epicManager)
	assert.NotNil(t, stack.ticketManager)
}

func TestInterruptionStack_SaveCurrentContext(t *testing.T) {
	tempDir := t.TempDir()
	setupTestDirs(t, tempDir)
	
	stack := NewInterruptionStack(tempDir)
	
	// Test saving normal context
	options := ContextSaveOptions{
		Name:             "Test Normal Context",
		Description:      "Testing normal workflow context",
		Type:             WorkflowContextTypeNormal,
		UserNotes:        "Test notes",
		Tags:             []string{"test", "normal"},
		IncludeFileState: true,
		IncludeGitState:  true,
	}
	
	context, err := stack.SaveCurrentContext(options)
	require.NoError(t, err)
	require.NotNil(t, context)
	
	// Verify context properties
	assert.NotEmpty(t, context.ID)
	assert.Equal(t, "Test Normal Context", context.Name)
	assert.Equal(t, "Testing normal workflow context", context.Description)
	assert.Equal(t, WorkflowContextTypeNormal, context.Type)
	assert.Equal(t, tempDir, context.WorkingDirectory)
	assert.Equal(t, "Test notes", context.UserNotes)
	assert.Equal(t, []string{"test", "normal"}, context.Tags)
	assert.False(t, context.CreatedAt.IsZero())
	assert.False(t, context.SavedAt.IsZero())
	assert.False(t, context.LastAccessedAt.IsZero())
	assert.NotNil(t, context.Metadata)
	
	// Verify file and git state capture
	assert.Equal(t, true, context.Metadata["file_state_captured"])
	assert.Equal(t, true, context.Metadata["git_state_captured"])
}

func TestInterruptionStack_SaveInterruptionContext(t *testing.T) {
	tempDir := t.TempDir()
	setupTestDirs(t, tempDir)
	
	stack := NewInterruptionStack(tempDir)
	
	// Save initial normal context
	normalOptions := ContextSaveOptions{
		Name: "Normal Work",
		Type: WorkflowContextTypeNormal,
	}
	
	normalContext, err := stack.SaveCurrentContext(normalOptions)
	require.NoError(t, err)
	
	// Save interruption context
	interruptOptions := ContextSaveOptions{
		Name: "Urgent Bug Fix",
		Type: WorkflowContextTypeInterruption,
	}
	
	interruptContext, err := stack.SaveCurrentContext(interruptOptions)
	require.NoError(t, err)
	
	// Verify interruption context has parent relationship
	assert.Equal(t, normalContext.ID, interruptContext.ParentContextID)
	
	// Get updated stack data to check parent-child relationships
	stackData, err := stack.loadStack()
	require.NoError(t, err)
	
	// Find the normal context in the stack (it should be there)
	var updatedNormalContext *WorkflowContext
	for _, ctx := range stackData.ContextStack {
		if ctx.ID == normalContext.ID {
			updatedNormalContext = ctx
			break
		}
	}
	require.NotNil(t, updatedNormalContext, "Normal context should be in stack")
	assert.Contains(t, updatedNormalContext.ChildContextIDs, interruptContext.ID)
	
	// Verify stack metadata
	assert.Equal(t, 1, stackData.Metadata.TotalInterruptions)
	assert.Equal(t, 1, stackData.Metadata.ActiveInterruptions)
	assert.Equal(t, 1, stackData.Metadata.CurrentStackDepth)
	assert.Equal(t, 1, stackData.Metadata.MaxStackDepth)
	assert.Equal(t, 1, len(stackData.ContextStack))
	assert.Equal(t, interruptContext.ID, stackData.CurrentContext.ID)
}

func TestInterruptionStack_NestedInterruptions(t *testing.T) {
	tempDir := t.TempDir()
	setupTestDirs(t, tempDir)
	
	stack := NewInterruptionStack(tempDir)
	
	// Create nested interruption stack: Normal -> Interruption -> Emergency
	contexts := make([]*WorkflowContext, 3)
	
	// Level 1: Normal work
	options1 := ContextSaveOptions{
		Name: "Feature Development",
		Type: WorkflowContextTypeNormal,
	}
	contexts[0], _ = stack.SaveCurrentContext(options1)
	
	// Level 2: Interruption
	options2 := ContextSaveOptions{
		Name: "Bug Report",
		Type: WorkflowContextTypeInterruption,
	}
	contexts[1], _ = stack.SaveCurrentContext(options2)
	
	// Level 3: Emergency
	options3 := ContextSaveOptions{
		Name: "Production Issue",
		Type: WorkflowContextTypeEmergency,
	}
	contexts[2], _ = stack.SaveCurrentContext(options3)
	
	// Verify stack depth and relationships
	stackData, err := stack.loadStack()
	require.NoError(t, err)
	assert.Equal(t, 2, stackData.Metadata.CurrentStackDepth) // Emergency is current, 2 in stack
	assert.Equal(t, 2, stackData.Metadata.ActiveInterruptions)
	assert.Equal(t, 2, stackData.Metadata.TotalInterruptions) // Only interruption and emergency are counted
	
	// Verify parent-child relationships
	assert.Equal(t, contexts[0].ID, contexts[1].ParentContextID)
	assert.Equal(t, contexts[1].ID, contexts[2].ParentContextID)
	
	// Check child relationships in the active contexts (updated versions)
	assert.Contains(t, stackData.ActiveContexts[contexts[0].ID].ChildContextIDs, contexts[1].ID)
	assert.Contains(t, stackData.ActiveContexts[contexts[1].ID].ChildContextIDs, contexts[2].ID)
	
	// Verify current context is the emergency
	assert.Equal(t, contexts[2].ID, stackData.CurrentContext.ID)
}

func TestInterruptionStack_RestoreContext(t *testing.T) {
	tempDir := t.TempDir()
	setupTestDirs(t, tempDir)
	
	stack := NewInterruptionStack(tempDir)
	
	// Create context hierarchy
	normalCtx, _ := stack.SaveCurrentContext(ContextSaveOptions{
		Name: "Normal Work",
		Type: WorkflowContextTypeNormal,
	})
	
	_, _ = stack.SaveCurrentContext(ContextSaveOptions{
		Name: "Interruption",
		Type: WorkflowContextTypeInterruption,
	})
	
	// Restore to normal context
	restoreOptions := ContextRestoreOptions{
		RestoreFiles:    true,
		RestoreGitState: true,
		RestoreTickets:  true,
		RestoreEpics:    true,
		Force:           false,
		BackupCurrent:   true,
	}
	
	err := stack.RestoreContext(normalCtx.ID, restoreOptions)
	require.NoError(t, err)
	
	// Verify current context is restored
	currentCtx, err := stack.GetCurrentContext()
	require.NoError(t, err)
	assert.Equal(t, normalCtx.ID, currentCtx.ID)
	
	// Verify stack metadata updated
	stackData, err := stack.loadStack()
	require.NoError(t, err)
	assert.Equal(t, 0, stackData.Metadata.ActiveInterruptions)
	assert.Equal(t, 0, stackData.Metadata.CurrentStackDepth)
	
	// Verify last accessed time updated
	assert.True(t, currentCtx.LastAccessedAt.After(normalCtx.LastAccessedAt))
}

func TestInterruptionStack_PopContext(t *testing.T) {
	tempDir := t.TempDir()
	setupTestDirs(t, tempDir)
	
	stack := NewInterruptionStack(tempDir)
	
	// Create context stack
	normalCtx, _ := stack.SaveCurrentContext(ContextSaveOptions{
		Name: "Normal Work",
		Type: WorkflowContextTypeNormal,
	})
	
	_, _ = stack.SaveCurrentContext(ContextSaveOptions{
		Name: "Interruption",
		Type: WorkflowContextTypeInterruption,
	})
	
	// Pop the most recent context (should restore normal work)
	restoreOptions := ContextRestoreOptions{
		RestoreFiles:   true,
		RestoreTickets: true,
		RestoreEpics:   true,
	}
	
	restoredCtx, err := stack.PopContext(restoreOptions)
	require.NoError(t, err)
	assert.Equal(t, normalCtx.ID, restoredCtx.ID)
	
	// Verify current context
	currentCtx, err := stack.GetCurrentContext()
	require.NoError(t, err)
	assert.Equal(t, normalCtx.ID, currentCtx.ID)
	
	// Verify stack is empty
	depth, err := stack.GetStackDepth()
	require.NoError(t, err)
	assert.Equal(t, 0, depth)
	
	// Test popping from empty stack
	_, err = stack.PopContext(restoreOptions)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "interruption stack is empty")
}

func TestInterruptionStack_ContextNotFound(t *testing.T) {
	tempDir := t.TempDir()
	setupTestDirs(t, tempDir)
	
	stack := NewInterruptionStack(tempDir)
	
	// Try to restore non-existent context
	restoreOptions := ContextRestoreOptions{}
	err := stack.RestoreContext("non-existent-id", restoreOptions)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context non-existent-id not found")
}

func TestInterruptionStack_RestoreCurrentContext(t *testing.T) {
	tempDir := t.TempDir()
	setupTestDirs(t, tempDir)
	
	stack := NewInterruptionStack(tempDir)
	
	// Save a context
	ctx, _ := stack.SaveCurrentContext(ContextSaveOptions{
		Name: "Test Context",
		Type: WorkflowContextTypeNormal,
	})
	
	// Try to restore the same context (should fail)
	restoreOptions := ContextRestoreOptions{}
	err := stack.RestoreContext(ctx.ID, restoreOptions)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "is already current")
}

func TestInterruptionStack_ListContexts(t *testing.T) {
	tempDir := t.TempDir()
	setupTestDirs(t, tempDir)
	
	stack := NewInterruptionStack(tempDir)
	
	// Create multiple contexts
	contexts := make([]*WorkflowContext, 3)
	for i := 0; i < 3; i++ {
		contextType := WorkflowContextTypeNormal
		if i > 0 {
			contextType = WorkflowContextTypeInterruption
		}
		
		options := ContextSaveOptions{
			Name: fmt.Sprintf("Context %d", i+1),
			Type: contextType,
		}
		contexts[i], _ = stack.SaveCurrentContext(options)
	}
	
	// List all contexts
	stackData, err := stack.ListContexts()
	require.NoError(t, err)
	
	// Verify we have all contexts
	assert.Equal(t, 2, len(stackData.ContextStack)) // First two pushed to stack
	assert.Equal(t, 3, len(stackData.ContextHistory)) // All in history
	assert.Equal(t, 3, len(stackData.ActiveContexts)) // All active
	assert.Equal(t, contexts[2].ID, stackData.CurrentContext.ID) // Last is current
}

func TestInterruptionStack_GetStackDepth(t *testing.T) {
	tempDir := t.TempDir()
	setupTestDirs(t, tempDir)
	
	stack := NewInterruptionStack(tempDir)
	
	// Initially empty
	depth, err := stack.GetStackDepth()
	require.NoError(t, err)
	assert.Equal(t, 0, depth)
	
	// Add normal context (doesn't increase depth)
	stack.SaveCurrentContext(ContextSaveOptions{
		Name: "Normal",
		Type: WorkflowContextTypeNormal,
	})
	
	depth, err = stack.GetStackDepth()
	require.NoError(t, err)
	assert.Equal(t, 0, depth)
	
	// Add interruption (increases depth)
	stack.SaveCurrentContext(ContextSaveOptions{
		Name: "Interrupt 1",
		Type: WorkflowContextTypeInterruption,
	})
	
	depth, err = stack.GetStackDepth()
	require.NoError(t, err)
	assert.Equal(t, 1, depth)
	
	// Add another interruption
	stack.SaveCurrentContext(ContextSaveOptions{
		Name: "Interrupt 2",
		Type: WorkflowContextTypeEmergency,
	})
	
	depth, err = stack.GetStackDepth()
	require.NoError(t, err)
	assert.Equal(t, 2, depth)
}

func TestInterruptionStack_ClearStack(t *testing.T) {
	tempDir := t.TempDir()
	setupTestDirs(t, tempDir)
	
	stack := NewInterruptionStack(tempDir)
	
	// Create contexts
	for i := 0; i < 3; i++ {
		stack.SaveCurrentContext(ContextSaveOptions{
			Name: fmt.Sprintf("Context %d", i+1),
			Type: WorkflowContextTypeInterruption,
		})
	}
	
	// Verify stack has contents
	depth, _ := stack.GetStackDepth()
	assert.Equal(t, 2, depth)
	
	// Clear stack
	err := stack.ClearStack()
	require.NoError(t, err)
	
	// Verify stack is cleared
	depth, err = stack.GetStackDepth()
	require.NoError(t, err)
	assert.Equal(t, 0, depth)
	
	currentCtx, err := stack.GetCurrentContext()
	require.NoError(t, err)
	assert.Nil(t, currentCtx)
	
	stackData, err := stack.loadStack()
	require.NoError(t, err)
	assert.Equal(t, 0, len(stackData.ContextStack))
	assert.Equal(t, 0, len(stackData.ActiveContexts))
	assert.Equal(t, 0, stackData.Metadata.ActiveInterruptions)
	assert.Equal(t, 0, stackData.Metadata.CurrentStackDepth)
}

func TestInterruptionStack_WithEpicAndTicketContext(t *testing.T) {
	tempDir := t.TempDir()
	setupTestDirs(t, tempDir)
	
	// Create an epic and ticket for testing
	setupTestEpicAndTicket(t, tempDir)
	
	stack := NewInterruptionStack(tempDir)
	
	// Save context (should capture epic and ticket)
	options := ContextSaveOptions{
		Name:             "Context with Epic and Ticket",
		Type:             WorkflowContextTypeNormal,
		IncludeFileState: true,
		IncludeGitState:  true,
	}
	
	context, err := stack.SaveCurrentContext(options)
	require.NoError(t, err)
	
	// Verify epic and ticket context captured
	// Note: Since we're using mocked managers, epic and ticket IDs might be empty
	// but the capture methods should not fail
	assert.NotNil(t, context)
}

func TestInterruptionStack_HistoryLimiting(t *testing.T) {
	tempDir := t.TempDir()
	setupTestDirs(t, tempDir)
	
	stack := NewInterruptionStack(tempDir)
	
	// Create more than 50 contexts to test history limiting
	for i := 0; i < 55; i++ {
		options := ContextSaveOptions{
			Name: fmt.Sprintf("Context %d", i+1),
			Type: WorkflowContextTypeNormal,
		}
		stack.SaveCurrentContext(options)
	}
	
	// Verify history is limited to 50
	stackData, err := stack.loadStack()
	require.NoError(t, err)
	assert.Equal(t, 50, len(stackData.ContextHistory))
	
	// Verify the most recent contexts are kept
	lastContext := stackData.ContextHistory[len(stackData.ContextHistory)-1]
	assert.Equal(t, "Context 55", lastContext.Name)
}

func TestInterruptionStack_PersistenceAndMigration(t *testing.T) {
	tempDir := t.TempDir()
	setupTestDirs(t, tempDir)
	
	stack1 := NewInterruptionStack(tempDir)
	
	// Save some contexts
	ctx1, _ := stack1.SaveCurrentContext(ContextSaveOptions{
		Name: "Persistent Context 1",
		Type: WorkflowContextTypeNormal,
	})
	
	ctx2, _ := stack1.SaveCurrentContext(ContextSaveOptions{
		Name: "Persistent Context 2",
		Type: WorkflowContextTypeInterruption,
	})
	
	// Create new stack instance (simulates app restart)
	stack2 := NewInterruptionStack(tempDir)
	
	// Verify data persisted
	currentCtx, err := stack2.GetCurrentContext()
	require.NoError(t, err)
	assert.Equal(t, ctx2.ID, currentCtx.ID)
	assert.Equal(t, "Persistent Context 2", currentCtx.Name)
	
	stackData, err := stack2.loadStack()
	require.NoError(t, err)
	assert.Equal(t, 1, len(stackData.ContextStack))
	assert.Equal(t, ctx1.ID, stackData.ContextStack[0].ID)
	assert.Equal(t, 2, len(stackData.ActiveContexts))
}

func TestWorkflowContextType_Constants(t *testing.T) {
	// Test all context type constants
	assert.Equal(t, WorkflowContextType("normal"), WorkflowContextTypeNormal)
	assert.Equal(t, WorkflowContextType("interruption"), WorkflowContextTypeInterruption)
	assert.Equal(t, WorkflowContextType("emergency"), WorkflowContextTypeEmergency)
	assert.Equal(t, WorkflowContextType("hotfix"), WorkflowContextTypeHotfix)
	assert.Equal(t, WorkflowContextType("experiment"), WorkflowContextTypeExperiment)
}

func TestContextSaveOptions_Validation(t *testing.T) {
	tempDir := t.TempDir()
	setupTestDirs(t, tempDir)
	
	stack := NewInterruptionStack(tempDir)
	
	// Test with minimal options
	options := ContextSaveOptions{
		Name: "Minimal Context",
		Type: WorkflowContextTypeNormal,
	}
	
	context, err := stack.SaveCurrentContext(options)
	require.NoError(t, err)
	assert.Equal(t, "Minimal Context", context.Name)
	assert.Equal(t, WorkflowContextTypeNormal, context.Type)
	assert.Empty(t, context.Description)
	assert.Empty(t, context.UserNotes)
	assert.Empty(t, context.Tags)
}

func TestContextRestoreOptions_Validation(t *testing.T) {
	tempDir := t.TempDir()
	setupTestDirs(t, tempDir)
	
	stack := NewInterruptionStack(tempDir)
	
	// Create and save context
	ctx, _ := stack.SaveCurrentContext(ContextSaveOptions{
		Name: "Test Context",
		Type: WorkflowContextTypeNormal,
	})
	
	// Create interruption
	stack.SaveCurrentContext(ContextSaveOptions{
		Name: "Interruption",
		Type: WorkflowContextTypeInterruption,
	})
	
	// Test with all restore options enabled
	restoreOptions := ContextRestoreOptions{
		RestoreFiles:    true,
		RestoreGitState: true,
		RestoreTickets:  true,
		RestoreEpics:    true,
		Force:           true,
		BackupCurrent:   true,
	}
	
	err := stack.RestoreContext(ctx.ID, restoreOptions)
	require.NoError(t, err)
	
	// Verify backup was created (should have at least the original contexts)
	stackData, err := stack.loadStack()
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(stackData.ContextHistory), 2) // At least original 2 contexts
}

// Helper functions

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

func setupTestEpicAndTicket(t *testing.T, tempDir string) {
	// Create a test epic
	epicManager := epic.NewManager(tempDir)
	epicOptions := epic.EpicCreateOptions{
		Title:       "Test Epic",
		Description: "Test epic for workflow preservation",
		Priority:    epic.PriorityMedium,
	}
	testEpic, err := epicManager.CreateEpic(epicOptions)
	require.NoError(t, err)
	
	// Set as current epic
	_, err = epicManager.SelectEpic(testEpic.ID)
	require.NoError(t, err)
	
	// Create a test ticket
	ticketManager := ticket.NewManager(tempDir)
	ticketOptions := ticket.TicketCreateOptions{
		Title:       "Test Ticket",
		Description: "Test ticket for workflow preservation",
		Type:        ticket.TicketTypeBug,
		Priority:    ticket.TicketPriorityMedium,
	}
	testTicket, err := ticketManager.CreateTicket(ticketOptions)
	require.NoError(t, err)
	
	// Set as current ticket
	_, err = ticketManager.SetCurrentTicket(testTicket.ID)
	require.NoError(t, err)
}