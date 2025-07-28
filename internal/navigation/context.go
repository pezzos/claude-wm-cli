package navigation

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// WorkflowState represents the current state of the project workflow
type WorkflowState int

const (
	StateNotInitialized WorkflowState = iota
	StateProjectInitialized
	StateHasEpics
	StateEpicInProgress
	StateStoryInProgress
	StateTaskInProgress
)

// String returns a human-readable representation of the WorkflowState
func (ws WorkflowState) String() string {
	switch ws {
	case StateNotInitialized:
		return "Not Initialized"
	case StateProjectInitialized:
		return "Project Initialized"
	case StateHasEpics:
		return "Has Epics"
	case StateEpicInProgress:
		return "Epic In Progress"
	case StateStoryInProgress:
		return "Story In Progress"
	case StateTaskInProgress:
		return "Task In Progress"
	default:
		return "Unknown State"
	}
}

// ProjectContext contains information about the current project state
type ProjectContext struct {
	State            WorkflowState
	ProjectPath      string
	CurrentEpic      *EpicContext
	CurrentStory     *StoryContext
	CurrentTask      *TaskContext
	AvailableActions []string
	Issues           []string // List of issues or warnings about project state
}

// EpicContext contains information about the current epic
type EpicContext struct {
	ID               string
	Title            string
	Status           string
	Priority         string
	Progress         float64 // 0.0 to 1.0
	TotalStories     int
	CompletedStories int
}

// StoryContext contains information about the current story
type StoryContext struct {
	ID             string
	Title          string
	Status         string
	Priority       string
	Progress       float64
	TotalTasks     int
	CompletedTasks int
}

// TaskContext contains information about the current task
type TaskContext struct {
	ID             string
	Title          string
	Status         string
	Priority       string
	EstimatedHours int
}

// ContextDetector is responsible for analyzing project state
type ContextDetector struct {
	projectPath string
}

// NewContextDetector creates a new context detector for the given project path
func NewContextDetector(projectPath string) *ContextDetector {
	return &ContextDetector{
		projectPath: projectPath,
	}
}

// DetectContext analyzes the current project state and returns context information
func (cd *ContextDetector) DetectContext() (*ProjectContext, error) {
	ctx := &ProjectContext{
		ProjectPath:      cd.projectPath,
		AvailableActions: []string{},
		Issues:           []string{},
	}

	// Check if docs directory exists
	docsPath := filepath.Join(cd.projectPath, "docs")
	if !cd.pathExists(docsPath) {
		ctx.State = StateNotInitialized
		ctx.AvailableActions = append(ctx.AvailableActions, "init-project")
		return ctx, nil
	}

	// Check project structure
	if err := cd.validateProjectStructure(ctx); err != nil {
		ctx.Issues = append(ctx.Issues, fmt.Sprintf("Project structure issue: %v", err))
	}

	// Detect current state based on existing files
	if err := cd.detectCurrentState(ctx); err != nil {
		return nil, fmt.Errorf("failed to detect current state: %w", err)
	}

	// Determine available actions based on current state
	cd.determineAvailableActions(ctx)

	return ctx, nil
}

// pathExists checks if a path exists
func (cd *ContextDetector) pathExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// validateProjectStructure validates the expected project directory structure
func (cd *ContextDetector) validateProjectStructure(ctx *ProjectContext) error {
	requiredDirs := []string{
		"docs/1-project",
		"docs/2-current-epic",
		"docs/3-current-task",
	}

	for _, dir := range requiredDirs {
		dirPath := filepath.Join(cd.projectPath, dir)
		if !cd.pathExists(dirPath) {
			return fmt.Errorf("missing required directory: %s", dir)
		}
	}

	ctx.State = StateProjectInitialized
	return nil
}

// detectCurrentState analyzes existing files to determine the current workflow state
func (cd *ContextDetector) detectCurrentState(ctx *ProjectContext) error {
	// Check for epics.json
	epicsPath := filepath.Join(cd.projectPath, "docs/1-project/epics.json")
	if cd.pathExists(epicsPath) {
		ctx.State = StateHasEpics

		// Validate epics.json file
		if err := cd.validateEpicsFile(epicsPath); err != nil {
			ctx.Issues = append(ctx.Issues, fmt.Sprintf("Invalid epics.json: %v", err))
		}

		// Try to load epic context
		if epicCtx, err := cd.loadEpicContext(); err != nil {
			ctx.Issues = append(ctx.Issues, fmt.Sprintf("Failed to load epic context: %v", err))
		} else if epicCtx != nil {
			ctx.CurrentEpic = epicCtx
			ctx.State = StateEpicInProgress
		}
	}

	// Check for current story
	if ctx.CurrentEpic != nil {
		if storyCtx, err := cd.loadStoryContext(); err != nil {
			ctx.Issues = append(ctx.Issues, fmt.Sprintf("Failed to load story context: %v", err))
		} else if storyCtx != nil {
			ctx.CurrentStory = storyCtx
			ctx.State = StateStoryInProgress
		}
	}

	// Check for current task
	if ctx.CurrentStory != nil {
		if taskCtx, err := cd.loadTaskContext(); err != nil {
			ctx.Issues = append(ctx.Issues, fmt.Sprintf("Failed to load task context: %v", err))
		} else if taskCtx != nil {
			ctx.CurrentTask = taskCtx
			ctx.State = StateTaskInProgress
		}
	}

	return nil
}

// validateEpicsFile validates that epics.json contains valid JSON
func (cd *ContextDetector) validateEpicsFile(epicsPath string) error {
	data, err := os.ReadFile(epicsPath)
	if err != nil {
		return fmt.Errorf("failed to read epics.json: %w", err)
	}

	var epicsData map[string]interface{}
	if err := json.Unmarshal(data, &epicsData); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	return nil
}

// loadEpicContext loads the current epic context from state files
func (cd *ContextDetector) loadEpicContext() (*EpicContext, error) {
	currentEpicPath := filepath.Join(cd.projectPath, "docs/2-current-epic/current-epic.json")
	if !cd.pathExists(currentEpicPath) {
		return nil, nil
	}

	data, err := os.ReadFile(currentEpicPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read current-epic.json: %w", err)
	}

	var epicData struct {
		Epic struct {
			ID          string `json:"id"`
			Title       string `json:"title"`
			Status      string `json:"status"`
			Priority    string `json:"priority"`
			UserStories []struct {
				Status string `json:"status"`
			} `json:"userStories"`
		} `json:"epic"`
	}

	if err := json.Unmarshal(data, &epicData); err != nil {
		return nil, fmt.Errorf("failed to parse current-epic.json: %w", err)
	}

	// Calculate progress
	totalStories := len(epicData.Epic.UserStories)
	completedStories := 0
	for _, story := range epicData.Epic.UserStories {
		if story.Status == "completed" {
			completedStories++
		}
	}

	progress := 0.0
	if totalStories > 0 {
		progress = float64(completedStories) / float64(totalStories)
	}

	return &EpicContext{
		ID:               epicData.Epic.ID,
		Title:            epicData.Epic.Title,
		Status:           epicData.Epic.Status,
		Priority:         epicData.Epic.Priority,
		Progress:         progress,
		TotalStories:     totalStories,
		CompletedStories: completedStories,
	}, nil
}

// loadStoryContext loads the current story context from state files
func (cd *ContextDetector) loadStoryContext() (*StoryContext, error) {
	storiesPath := filepath.Join(cd.projectPath, "docs/2-current-epic/stories.json")
	if !cd.pathExists(storiesPath) {
		return nil, nil
	}

	data, err := os.ReadFile(storiesPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read stories.json: %w", err)
	}

	var storiesData struct {
		Stories map[string]struct {
			ID       string `json:"id"`
			Title    string `json:"title"`
			Status   string `json:"status"`
			Priority string `json:"priority"`
		} `json:"stories"`
	}

	if err := json.Unmarshal(data, &storiesData); err != nil {
		return nil, fmt.Errorf("failed to parse stories.json: %w", err)
	}

	// Find current story (first in_progress story)
	for _, story := range storiesData.Stories {
		if story.Status == "in_progress" || story.Status == "todo" {
			return &StoryContext{
				ID:       story.ID,
				Title:    story.Title,
				Status:   story.Status,
				Priority: story.Priority,
				Progress: 0.0, // TODO: Calculate from tasks
			}, nil
		}
	}

	return nil, nil
}

// loadTaskContext loads the current task context from todo files
func (cd *ContextDetector) loadTaskContext() (*TaskContext, error) {
	taskDir := filepath.Join(cd.projectPath, "docs/3-current-task")

	// Look for todo files in current task directory
	entries, err := os.ReadDir(taskDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read task directory: %w", err)
	}

	// Find the most recent todo file
	var todoFile string
	for _, entry := range entries {
		if filepath.Ext(entry.Name()) == ".json" &&
			(entry.Name() == "todo.json" || entry.Name() == "todo-epic-001.json") {
			todoFile = filepath.Join(taskDir, entry.Name())
			break
		}
	}

	if todoFile == "" {
		return nil, nil
	}

	data, err := os.ReadFile(todoFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read todo file: %w", err)
	}

	var todoData struct {
		Todos []struct {
			ID             string `json:"id"`
			Title          string `json:"title"`
			Status         string `json:"status"`
			Priority       string `json:"priority"`
			EstimatedHours int    `json:"estimatedHours"`
		} `json:"todos"`
	}

	if err := json.Unmarshal(data, &todoData); err != nil {
		return nil, fmt.Errorf("failed to parse todo file: %w", err)
	}

	// Find current task (first in_progress or todo task)
	for _, task := range todoData.Todos {
		if task.Status == "in_progress" || task.Status == "todo" {
			return &TaskContext{
				ID:             task.ID,
				Title:          task.Title,
				Status:         task.Status,
				Priority:       task.Priority,
				EstimatedHours: task.EstimatedHours,
			}, nil
		}
	}

	return nil, nil
}

// determineAvailableActions determines what actions are available based on current state
func (cd *ContextDetector) determineAvailableActions(ctx *ProjectContext) {
	switch ctx.State {
	case StateNotInitialized:
		ctx.AvailableActions = []string{
			"init-project",
			"help",
		}
	case StateProjectInitialized:
		ctx.AvailableActions = []string{
			"create-epic",
			"list-epics",
			"help",
		}
	case StateHasEpics:
		ctx.AvailableActions = []string{
			"start-epic",
			"create-epic",
			"list-epics",
			"help",
		}
	case StateEpicInProgress:
		ctx.AvailableActions = []string{
			"continue-epic",
			"list-stories",
			"create-story",
			"switch-epic",
			"help",
		}
	case StateStoryInProgress:
		ctx.AvailableActions = []string{
			"continue-story",
			"list-tasks",
			"create-task",
			"complete-story",
			"help",
		}
	case StateTaskInProgress:
		ctx.AvailableActions = []string{
			"continue-task",
			"complete-task",
			"create-task",
			"switch-task",
			"help",
		}
	}

	// Always add common actions
	commonActions := []string{"status", "interactive", "exit"}
	ctx.AvailableActions = append(ctx.AvailableActions, commonActions...)
}

// GetRecommendedAction returns the most recommended action based on current state
func (cd *ContextDetector) GetRecommendedAction(ctx *ProjectContext) string {
	switch ctx.State {
	case StateNotInitialized:
		return "init-project"
	case StateProjectInitialized:
		return "create-epic"
	case StateHasEpics:
		return "start-epic"
	case StateEpicInProgress:
		if ctx.CurrentStory == nil {
			return "continue-epic"
		}
		return "continue-story"
	case StateStoryInProgress:
		if ctx.CurrentTask == nil {
			return "continue-story"
		}
		return "continue-task"
	case StateTaskInProgress:
		return "continue-task"
	default:
		return "help"
	}
}
