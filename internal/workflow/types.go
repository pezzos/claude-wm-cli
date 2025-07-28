package workflow

// Priority represents the priority level of a workflow action
type Priority string

const (
	PriorityP0 Priority = "P0" // Critical/Blocking
	PriorityP1 Priority = "P1" // Important
	PriorityP2 Priority = "P2" // Nice to have
)

// String returns the string representation of the priority
func (p Priority) String() string {
	return string(p)
}

// WorkflowAction represents an action that can be performed in the workflow
type WorkflowAction struct {
	ID            string
	Name          string
	Description   string
	Priority      Priority
	Prerequisites []string // List of required conditions/states
	Blocks        []string // List of actions this blocks
}

// ActionRegistry contains all available workflow actions
type ActionRegistry struct {
	actions map[string]*WorkflowAction
}

// NewActionRegistry creates a new action registry with default actions
func NewActionRegistry() *ActionRegistry {
	registry := &ActionRegistry{
		actions: make(map[string]*WorkflowAction),
	}

	// Register default workflow actions
	registry.registerDefaultActions()
	return registry
}

// registerDefaultActions registers the standard workflow actions
func (ar *ActionRegistry) registerDefaultActions() {
	defaultActions := []*WorkflowAction{
		{
			ID:            "init-project",
			Name:          "Initialize Project",
			Description:   "Set up the project structure and initialize the workflow system",
			Priority:      PriorityP0,
			Prerequisites: []string{"empty_directory"},
		},
		{
			ID:            "create-epic",
			Name:          "Create Epic",
			Description:   "Create a new epic to organize work",
			Priority:      PriorityP0,
			Prerequisites: []string{"project_initialized"},
		},
		{
			ID:            "start-epic",
			Name:          "Start Epic",
			Description:   "Begin work on an existing epic",
			Priority:      PriorityP0,
			Prerequisites: []string{"has_epics", "no_active_epic"},
		},
		{
			ID:            "continue-epic",
			Name:          "Continue Epic",
			Description:   "Continue working on the current epic",
			Priority:      PriorityP1,
			Prerequisites: []string{"epic_in_progress"},
		},
		{
			ID:            "create-story",
			Name:          "Create Story",
			Description:   "Create a new user story within the current epic",
			Priority:      PriorityP1,
			Prerequisites: []string{"epic_in_progress"},
		},
		{
			ID:            "continue-story",
			Name:          "Continue Story",
			Description:   "Continue working on the current story",
			Priority:      PriorityP1,
			Prerequisites: []string{"story_in_progress"},
		},
		{
			ID:            "create-task",
			Name:          "Create Task",
			Description:   "Create a new task within the current story",
			Priority:      PriorityP1,
			Prerequisites: []string{"story_in_progress"},
		},
		{
			ID:            "continue-task",
			Name:          "Continue Task",
			Description:   "Continue working on the current task",
			Priority:      PriorityP1,
			Prerequisites: []string{"task_in_progress"},
		},
		{
			ID:            "complete-task",
			Name:          "Complete Task",
			Description:   "Mark the current task as completed",
			Priority:      PriorityP1,
			Prerequisites: []string{"task_in_progress"},
		},
		{
			ID:            "complete-story",
			Name:          "Complete Story",
			Description:   "Mark the current story as completed",
			Priority:      PriorityP1,
			Prerequisites: []string{"story_in_progress", "all_tasks_complete"},
		},
		{
			ID:            "complete-epic",
			Name:          "Complete Epic",
			Description:   "Mark the current epic as completed",
			Priority:      PriorityP1,
			Prerequisites: []string{"epic_in_progress", "all_stories_complete"},
		},
		{
			ID:            "list-epics",
			Name:          "List Epics",
			Description:   "Show all available epics and their status",
			Priority:      PriorityP2,
			Prerequisites: []string{"has_epics"},
		},
		{
			ID:            "list-stories",
			Name:          "List Stories",
			Description:   "Show all stories in the current epic",
			Priority:      PriorityP2,
			Prerequisites: []string{"epic_in_progress"},
		},
		{
			ID:            "list-tasks",
			Name:          "List Tasks",
			Description:   "Show all tasks in the current story",
			Priority:      PriorityP2,
			Prerequisites: []string{"story_in_progress"},
		},
		{
			ID:            "status",
			Name:          "Show Status",
			Description:   "Display current project status and progress",
			Priority:      PriorityP2,
			Prerequisites: []string{"project_initialized"},
		},
		{
			ID:            "help",
			Name:          "Show Help",
			Description:   "Display available commands and usage information",
			Priority:      PriorityP2,
			Prerequisites: []string{},
		},
	}

	for _, action := range defaultActions {
		ar.actions[action.ID] = action
	}
}

// GetAction returns an action by ID
func (ar *ActionRegistry) GetAction(id string) (*WorkflowAction, bool) {
	action, exists := ar.actions[id]
	return action, exists
}

// GetAllActions returns all registered actions
func (ar *ActionRegistry) GetAllActions() []*WorkflowAction {
	actions := make([]*WorkflowAction, 0, len(ar.actions))
	for _, action := range ar.actions {
		actions = append(actions, action)
	}
	return actions
}

// RegisterAction adds a new action to the registry
func (ar *ActionRegistry) RegisterAction(action *WorkflowAction) {
	ar.actions[action.ID] = action
}
