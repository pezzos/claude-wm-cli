package workflow

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
	
	"claude-wm-cli/internal/epic"
	"claude-wm-cli/internal/ticket"
)

const (
	InterruptionStackFileName = "interruption-stack.json"
	StackVersion             = "1.0.0"
)

// InterruptionStack manages the workflow state during interruptions
type InterruptionStack struct {
	rootPath      string
	epicManager   *epic.Manager
	ticketManager *ticket.Manager
}

// NewInterruptionStack creates a new interruption stack manager
func NewInterruptionStack(rootPath string) *InterruptionStack {
	return &InterruptionStack{
		rootPath:      rootPath,
		epicManager:   epic.NewManager(rootPath),
		ticketManager: ticket.NewManager(rootPath),
	}
}

// WorkflowContext represents the complete state of a workflow at a point in time
type WorkflowContext struct {
	ID               string                 `json:"id"`
	Name             string                 `json:"name"`
	Description      string                 `json:"description,omitempty"`
	Type             WorkflowContextType    `json:"type"`
	
	// Epic context
	CurrentEpicID    string                 `json:"current_epic_id,omitempty"`
	EpicState        *epic.Epic            `json:"epic_state,omitempty"`
	
	// Story context  
	CurrentStoryID   string                 `json:"current_story_id,omitempty"`
	StoryState       interface{}            `json:"story_state,omitempty"`
	
	// Ticket context
	CurrentTicketID  string                 `json:"current_ticket_id,omitempty"`
	TicketState      *ticket.Ticket        `json:"ticket_state,omitempty"`
	
	// File and git context
	WorkingDirectory string                 `json:"working_directory"`
	GitBranch        string                 `json:"git_branch,omitempty"`
	GitCommit        string                 `json:"git_commit,omitempty"`
	ModifiedFiles    []string               `json:"modified_files,omitempty"`
	
	// User notes and metadata
	UserNotes        string                 `json:"user_notes,omitempty"`
	Tags             []string               `json:"tags,omitempty"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
	
	// Timestamps
	CreatedAt        time.Time              `json:"created_at"`
	SavedAt          time.Time              `json:"saved_at"`
	LastAccessedAt   time.Time              `json:"last_accessed_at"`
	
	// Context relationships
	ParentContextID  string                 `json:"parent_context_id,omitempty"`
	ChildContextIDs  []string               `json:"child_context_ids,omitempty"`
}

// WorkflowContextType defines the type of workflow context
type WorkflowContextType string

const (
	WorkflowContextTypeNormal       WorkflowContextType = "normal"
	WorkflowContextTypeInterruption WorkflowContextType = "interruption"
	WorkflowContextTypeEmergency    WorkflowContextType = "emergency"
	WorkflowContextTypeHotfix       WorkflowContextType = "hotfix"
	WorkflowContextTypeExperiment   WorkflowContextType = "experiment"
)

// InterruptionStackData contains the complete interruption stack state
type InterruptionStackData struct {
	Version          string                        `json:"version"`
	CurrentContext   *WorkflowContext              `json:"current_context,omitempty"`
	ContextStack     []*WorkflowContext            `json:"context_stack"`
	ContextHistory   []*WorkflowContext            `json:"context_history"`
	ActiveContexts   map[string]*WorkflowContext   `json:"active_contexts"`
	Metadata         StackMetadata                 `json:"metadata"`
}

// StackMetadata contains metadata about the interruption stack
type StackMetadata struct {
	LastUpdated      time.Time                     `json:"last_updated"`
	TotalInterruptions int                         `json:"total_interruptions"`
	ActiveInterruptions int                        `json:"active_interruptions"`
	MaxStackDepth    int                           `json:"max_stack_depth"`
	CurrentStackDepth int                          `json:"current_stack_depth"`
}

// ContextSaveOptions configures how context is saved
type ContextSaveOptions struct {
	Name             string                 `json:"name"`
	Description      string                 `json:"description,omitempty"`
	Type             WorkflowContextType    `json:"type"`
	UserNotes        string                 `json:"user_notes,omitempty"`
	Tags             []string               `json:"tags,omitempty"`
	IncludeFileState bool                   `json:"include_file_state"`
	IncludeGitState  bool                   `json:"include_git_state"`
	AutoSave         bool                   `json:"auto_save"`
}

// ContextRestoreOptions configures how context is restored
type ContextRestoreOptions struct {
	RestoreFiles     bool                   `json:"restore_files"`
	RestoreGitState  bool                   `json:"restore_git_state"`
	RestoreTickets   bool                   `json:"restore_tickets"`
	RestoreEpics     bool                   `json:"restore_epics"`
	Force            bool                   `json:"force"`
	BackupCurrent    bool                   `json:"backup_current"`
}

// SaveCurrentContext saves the current workflow state to the interruption stack
func (is *InterruptionStack) SaveCurrentContext(options ContextSaveOptions) (*WorkflowContext, error) {
	// Generate unique context ID
	contextID := fmt.Sprintf("ctx-%d", time.Now().UnixNano())
	
	// Capture current state
	context := &WorkflowContext{
		ID:               contextID,
		Name:             options.Name,
		Description:      options.Description,
		Type:             options.Type,
		WorkingDirectory: is.rootPath,
		UserNotes:        options.UserNotes,
		Tags:             options.Tags,
		Metadata:         make(map[string]interface{}),
		CreatedAt:        time.Now(),
		SavedAt:          time.Now(),
		LastAccessedAt:   time.Now(),
	}
	
	// Capture epic context
	if err := is.captureEpicContext(context); err != nil {
		return nil, fmt.Errorf("failed to capture epic context: %w", err)
	}
	
	// Capture ticket context
	if err := is.captureTicketContext(context); err != nil {
		return nil, fmt.Errorf("failed to capture ticket context: %w", err)
	}
	
	// Capture file and git state if requested
	if options.IncludeFileState {
		if err := is.captureFileState(context); err != nil {
			return nil, fmt.Errorf("failed to capture file state: %w", err)
		}
	}
	
	if options.IncludeGitState {
		if err := is.captureGitState(context); err != nil {
			return nil, fmt.Errorf("failed to capture git state: %w", err)
		}
	}
	
	// Load current stack
	stackData, err := is.loadStack()
	if err != nil {
		return nil, fmt.Errorf("failed to load interruption stack: %w", err)
	}
	
	// Add to stack if this is an interruption
	if options.Type == WorkflowContextTypeInterruption || 
	   options.Type == WorkflowContextTypeEmergency ||
	   options.Type == WorkflowContextTypeHotfix {
		
		// Set parent relationship
		if stackData.CurrentContext != nil {
			context.ParentContextID = stackData.CurrentContext.ID
			stackData.CurrentContext.ChildContextIDs = append(
				stackData.CurrentContext.ChildContextIDs, contextID)
			// Update parent context in active contexts map
			stackData.ActiveContexts[stackData.CurrentContext.ID] = stackData.CurrentContext
		}
		
		// Push current context to stack
		if stackData.CurrentContext != nil {
			stackData.ContextStack = append(stackData.ContextStack, stackData.CurrentContext)
		}
		
		// Update metadata
		stackData.Metadata.TotalInterruptions++
		stackData.Metadata.ActiveInterruptions++
		stackData.Metadata.CurrentStackDepth = len(stackData.ContextStack)
		if stackData.Metadata.CurrentStackDepth > stackData.Metadata.MaxStackDepth {
			stackData.Metadata.MaxStackDepth = stackData.Metadata.CurrentStackDepth
		}
	}
	
	// Set as current context
	stackData.CurrentContext = context
	stackData.ActiveContexts[contextID] = context
	
	// Add to history
	stackData.ContextHistory = append(stackData.ContextHistory, context)
	
	// Keep history limited to last 50 contexts
	if len(stackData.ContextHistory) > 50 {
		stackData.ContextHistory = stackData.ContextHistory[1:]
	}
	
	// Update metadata
	stackData.Metadata.LastUpdated = time.Now()
	
	// Save stack
	if err := is.saveStack(stackData); err != nil {
		return nil, fmt.Errorf("failed to save interruption stack: %w", err)
	}
	
	return context, nil
}

// RestoreContext restores a workflow context from the stack
func (is *InterruptionStack) RestoreContext(contextID string, options ContextRestoreOptions) error {
	// Load current stack
	stackData, err := is.loadStack()
	if err != nil {
		return fmt.Errorf("failed to load interruption stack: %w", err)
	}
	
	// Find the context to restore
	var contextToRestore *WorkflowContext
	var isFromStack bool
	
	// Check if it's the current context
	if stackData.CurrentContext != nil && stackData.CurrentContext.ID == contextID {
		return fmt.Errorf("context %s is already current", contextID)
	}
	
	// Look in the stack
	for i, ctx := range stackData.ContextStack {
		if ctx.ID == contextID {
			contextToRestore = ctx
			isFromStack = true
			// Remove from stack
			stackData.ContextStack = append(stackData.ContextStack[:i], stackData.ContextStack[i+1:]...)
			break
		}
	}
	
	// Look in active contexts if not found in stack
	if contextToRestore == nil {
		if ctx, exists := stackData.ActiveContexts[contextID]; exists {
			contextToRestore = ctx
		}
	}
	
	// Look in history if still not found
	if contextToRestore == nil {
		for _, ctx := range stackData.ContextHistory {
			if ctx.ID == contextID {
				contextToRestore = ctx
				break
			}
		}
	}
	
	if contextToRestore == nil {
		return fmt.Errorf("context %s not found", contextID)
	}
	
	// Backup current context if requested
	if options.BackupCurrent && stackData.CurrentContext != nil {
		backupOptions := ContextSaveOptions{
			Name:             fmt.Sprintf("Backup before restoring %s", contextID),
			Type:             WorkflowContextTypeNormal,
			IncludeFileState: true,
			IncludeGitState:  true,
			AutoSave:         true,
		}
		_, err := is.SaveCurrentContext(backupOptions)
		if err != nil {
			return fmt.Errorf("failed to backup current context: %w", err)
		}
	}
	
	// Restore the context
	if err := is.restoreWorkflowState(contextToRestore, options); err != nil {
		return fmt.Errorf("failed to restore workflow state: %w", err)
	}
	
	// Update stack state
	if isFromStack {
		stackData.Metadata.ActiveInterruptions--
		stackData.Metadata.CurrentStackDepth = len(stackData.ContextStack)
	}
	
	// Set as current context
	stackData.CurrentContext = contextToRestore
	stackData.CurrentContext.LastAccessedAt = time.Now()
	
	// Update metadata
	stackData.Metadata.LastUpdated = time.Now()
	
	// Save stack
	if err := is.saveStack(stackData); err != nil {
		return fmt.Errorf("failed to save updated stack: %w", err)
	}
	
	return nil
}

// PopContext restores the most recent context from the stack
func (is *InterruptionStack) PopContext(options ContextRestoreOptions) (*WorkflowContext, error) {
	// Load current stack
	stackData, err := is.loadStack()
	if err != nil {
		return nil, fmt.Errorf("failed to load interruption stack: %w", err)
	}
	
	// Check if stack is empty
	if len(stackData.ContextStack) == 0 {
		return nil, fmt.Errorf("interruption stack is empty")
	}
	
	// Get the most recent context
	contextToRestore := stackData.ContextStack[len(stackData.ContextStack)-1]
	
	// Restore it
	if err := is.RestoreContext(contextToRestore.ID, options); err != nil {
		return nil, fmt.Errorf("failed to restore context: %w", err)
	}
	
	return contextToRestore, nil
}

// GetCurrentContext returns the current workflow context
func (is *InterruptionStack) GetCurrentContext() (*WorkflowContext, error) {
	stackData, err := is.loadStack()
	if err != nil {
		return nil, fmt.Errorf("failed to load interruption stack: %w", err)
	}
	
	return stackData.CurrentContext, nil
}

// ListContexts returns all contexts (stack + history)
func (is *InterruptionStack) ListContexts() (*InterruptionStackData, error) {
	return is.loadStack()
}

// GetStackDepth returns the current interruption stack depth
func (is *InterruptionStack) GetStackDepth() (int, error) {
	stackData, err := is.loadStack()
	if err != nil {
		return 0, fmt.Errorf("failed to load interruption stack: %w", err)
	}
	
	return len(stackData.ContextStack), nil
}

// ClearStack clears all contexts from the stack
func (is *InterruptionStack) ClearStack() error {
	stackData, err := is.loadStack()
	if err != nil {
		return fmt.Errorf("failed to load interruption stack: %w", err)
	}
	
	// Clear stack and active contexts
	stackData.ContextStack = []*WorkflowContext{}
	stackData.ActiveContexts = make(map[string]*WorkflowContext)
	stackData.CurrentContext = nil
	
	// Reset metadata
	stackData.Metadata.ActiveInterruptions = 0
	stackData.Metadata.CurrentStackDepth = 0
	stackData.Metadata.LastUpdated = time.Now()
	
	return is.saveStack(stackData)
}

// Private helper methods

func (is *InterruptionStack) captureEpicContext(context *WorkflowContext) error {
	// Get current epic
	currentEpic, err := is.epicManager.GetCurrentEpic()
	if err != nil {
		// Not an error if no current epic
		return nil
	}
	
	if currentEpic != nil {
		context.CurrentEpicID = currentEpic.ID
		context.EpicState = currentEpic
	}
	
	return nil
}

func (is *InterruptionStack) captureTicketContext(context *WorkflowContext) error {
	// Get current ticket
	currentTicket, err := is.ticketManager.GetCurrentTicket()
	if err != nil {
		// Not an error if no current ticket
		return nil
	}
	
	if currentTicket != nil {
		context.CurrentTicketID = currentTicket.ID
		context.TicketState = currentTicket
	}
	
	return nil
}

func (is *InterruptionStack) captureFileState(context *WorkflowContext) error {
	// TODO: Implement file state capture
	// This could include:
	// - List of modified files
	// - File checksums for change detection
	// - Temporary file backups
	
	context.Metadata["file_state_captured"] = true
	return nil
}

func (is *InterruptionStack) captureGitState(context *WorkflowContext) error {
	// TODO: Implement git state capture
	// This could include:
	// - Current branch
	// - Current commit
	// - Staged changes
	// - Stash state
	
	context.Metadata["git_state_captured"] = true
	return nil
}

func (is *InterruptionStack) restoreWorkflowState(context *WorkflowContext, options ContextRestoreOptions) error {
	// Restore epic context
	if options.RestoreEpics && context.CurrentEpicID != "" {
		_, err := is.epicManager.SelectEpic(context.CurrentEpicID)
		if err != nil {
			if !options.Force {
				return fmt.Errorf("failed to restore epic %s: %w", context.CurrentEpicID, err)
			}
			// Log warning but continue
			fmt.Printf("Warning: failed to restore epic %s: %v\n", context.CurrentEpicID, err)
		}
	}
	
	// Restore ticket context
	if options.RestoreTickets && context.CurrentTicketID != "" {
		_, err := is.ticketManager.SetCurrentTicket(context.CurrentTicketID)
		if err != nil {
			if !options.Force {
				return fmt.Errorf("failed to restore ticket %s: %w", context.CurrentTicketID, err)
			}
			// Log warning but continue
			fmt.Printf("Warning: failed to restore ticket %s: %v\n", context.CurrentTicketID, err)
		}
	}
	
	// Restore file state
	if options.RestoreFiles {
		if err := is.restoreFileState(context); err != nil {
			if !options.Force {
				return fmt.Errorf("failed to restore file state: %w", err)
			}
			fmt.Printf("Warning: failed to restore file state: %v\n", err)
		}
	}
	
	// Restore git state
	if options.RestoreGitState {
		if err := is.restoreGitState(context); err != nil {
			if !options.Force {
				return fmt.Errorf("failed to restore git state: %w", err)
			}
			fmt.Printf("Warning: failed to restore git state: %v\n", err)
		}
	}
	
	return nil
}

func (is *InterruptionStack) restoreFileState(context *WorkflowContext) error {
	// TODO: Implement file state restoration
	return nil
}

func (is *InterruptionStack) restoreGitState(context *WorkflowContext) error {
	// TODO: Implement git state restoration
	return nil
}

func (is *InterruptionStack) loadStack() (*InterruptionStackData, error) {
	stackPath := filepath.Join(is.rootPath, "docs", "2-current-epic", InterruptionStackFileName)
	
	// Check if file exists
	if _, err := os.Stat(stackPath); os.IsNotExist(err) {
		// Return empty stack
		return &InterruptionStackData{
			Version:        StackVersion,
			ContextStack:   []*WorkflowContext{},
			ContextHistory: []*WorkflowContext{},
			ActiveContexts: make(map[string]*WorkflowContext),
			Metadata: StackMetadata{
				LastUpdated: time.Now(),
			},
		}, nil
	}
	
	// Read file
	data, err := os.ReadFile(stackPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read stack file: %w", err)
	}
	
	var stackData InterruptionStackData
	if err := json.Unmarshal(data, &stackData); err != nil {
		return nil, fmt.Errorf("failed to parse stack file: %w", err)
	}
	
	// Validate and migrate if needed
	if err := is.validateAndMigrateStack(&stackData); err != nil {
		return nil, fmt.Errorf("failed to validate stack: %w", err)
	}
	
	return &stackData, nil
}

func (is *InterruptionStack) saveStack(stackData *InterruptionStackData) error {
	stackPath := filepath.Join(is.rootPath, "docs", "2-current-epic", InterruptionStackFileName)
	
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(stackPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	
	// Update metadata
	stackData.Metadata.LastUpdated = time.Now()
	stackData.Version = StackVersion
	
	// Marshal to JSON
	data, err := json.MarshalIndent(stackData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal stack data: %w", err)
	}
	
	// Write file atomically
	tempPath := stackPath + ".tmp"
	if err := os.WriteFile(tempPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write temp stack file: %w", err)
	}
	
	if err := os.Rename(tempPath, stackPath); err != nil {
		os.Remove(tempPath)
		return fmt.Errorf("failed to replace stack file: %w", err)
	}
	
	return nil
}

func (is *InterruptionStack) validateAndMigrateStack(stackData *InterruptionStackData) error {
	// Initialize maps if nil
	if stackData.ActiveContexts == nil {
		stackData.ActiveContexts = make(map[string]*WorkflowContext)
	}
	
	if stackData.ContextStack == nil {
		stackData.ContextStack = []*WorkflowContext{}
	}
	
	if stackData.ContextHistory == nil {
		stackData.ContextHistory = []*WorkflowContext{}
	}
	
	// Set version if empty
	if stackData.Version == "" {
		stackData.Version = StackVersion
	}
	
	// Validate contexts and populate active contexts map
	for _, ctx := range stackData.ContextStack {
		if ctx != nil && ctx.ID != "" {
			stackData.ActiveContexts[ctx.ID] = ctx
		}
	}
	
	// Add current context to active contexts
	if stackData.CurrentContext != nil && stackData.CurrentContext.ID != "" {
		stackData.ActiveContexts[stackData.CurrentContext.ID] = stackData.CurrentContext
	}
	
	// Update metadata counters
	stackData.Metadata.CurrentStackDepth = len(stackData.ContextStack)
	stackData.Metadata.ActiveInterruptions = len(stackData.ContextStack)
	
	return nil
}