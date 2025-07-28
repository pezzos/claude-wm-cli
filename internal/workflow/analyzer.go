package workflow

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"claude-wm-cli/internal/project"
	"claude-wm-cli/internal/state"
)

// WorkflowPosition represents the current position in the workflow hierarchy
type WorkflowPosition string

const (
	PositionUnknown       WorkflowPosition = "unknown"
	PositionNotInitialized WorkflowPosition = "not_initialized"
	PositionProjectLevel   WorkflowPosition = "project"
	PositionEpicLevel      WorkflowPosition = "epic"
	PositionStoryLevel     WorkflowPosition = "story"
	PositionTaskLevel      WorkflowPosition = "task"
)

// String returns the string representation of the workflow position
func (wp WorkflowPosition) String() string {
	return string(wp)
}

// WorkflowAnalysis contains the complete analysis of the current workflow state
type WorkflowAnalysis struct {
	Position            WorkflowPosition    `json:"position"`
	ProjectInitialized  bool               `json:"project_initialized"`
	CurrentEpic         *state.EpicState   `json:"current_epic,omitempty"`
	CurrentStory        *state.StoryState  `json:"current_story,omitempty"`
	CurrentTasks        []state.TaskState  `json:"current_tasks,omitempty"`
	CompletionMetrics   CompletionMetrics  `json:"completion_metrics"`
	Blockers           []WorkflowBlocker   `json:"blockers,omitempty"`
	Recommendations    []string           `json:"recommendations,omitempty"`
	RootPath           string             `json:"root_path"`
	AnalyzedAt         time.Time          `json:"analyzed_at"`
	Issues             []string           `json:"issues,omitempty"`
}

// CompletionMetrics tracks progress at all levels of the workflow
type CompletionMetrics struct {
	ProjectProgress  float64 `json:"project_progress"`  // Overall project completion %
	EpicProgress     float64 `json:"epic_progress"`     // Current epic completion %
	StoryProgress    float64 `json:"story_progress"`    // Current story completion %
	TotalEpics       int     `json:"total_epics"`
	CompletedEpics   int     `json:"completed_epics"`
	TotalStories     int     `json:"total_stories"`
	CompletedStories int     `json:"completed_stories"`
	TotalTasks       int     `json:"total_tasks"`
	CompletedTasks   int     `json:"completed_tasks"`
	EstimatedHours   float64 `json:"estimated_hours"`
	ActualHours      float64 `json:"actual_hours"`
}

// WorkflowBlocker represents an issue that prevents workflow progression
type WorkflowBlocker struct {
	Type        BlockerType `json:"type"`
	Severity    string      `json:"severity"` // critical, high, medium, low
	Description string      `json:"description"`
	Entity      string      `json:"entity,omitempty"`    // ID of blocked entity
	Suggestion  string      `json:"suggestion,omitempty"`
}

// BlockerType categorizes different types of workflow blockers
type BlockerType string

const (
	BlockerMissingDependency  BlockerType = "missing_dependency"
	BlockerInconsistentState  BlockerType = "inconsistent_state"
	BlockerMissingDefinition  BlockerType = "missing_definition"
	BlockerConfigurationError BlockerType = "configuration_error"
	BlockerStatusMismatch     BlockerType = "status_mismatch"
)

// WorkflowAnalyzer analyzes the current state of the workflow
type WorkflowAnalyzer struct {
	rootPath string
}

// NewWorkflowAnalyzer creates a new workflow analyzer
func NewWorkflowAnalyzer(rootPath string) *WorkflowAnalyzer {
	return &WorkflowAnalyzer{
		rootPath: rootPath,
	}
}

// AnalyzeWorkflowPosition performs a comprehensive analysis of the current workflow state
func (wa *WorkflowAnalyzer) AnalyzeWorkflowPosition() (*WorkflowAnalysis, error) {
	analysis := &WorkflowAnalysis{
		RootPath:   wa.rootPath,
		AnalyzedAt: time.Now(),
		Issues:     []string{},
		Blockers:   []WorkflowBlocker{},
		Recommendations: []string{},
	}

	// First, check if project is initialized using the project detection system
	projectResult, err := project.DetectProjectInitialization(wa.rootPath)
	if err != nil {
		return nil, fmt.Errorf("failed to detect project initialization: %w", err)
	}

	analysis.ProjectInitialized = projectResult.Status == project.Complete
	analysis.Issues = append(analysis.Issues, projectResult.Issues...)

	if !analysis.ProjectInitialized {
		analysis.Position = PositionNotInitialized
		analysis.Recommendations = append(analysis.Recommendations, "Initialize project structure")
		return analysis, nil
	}

	// Analyze current workflow position
	if err := wa.analyzeCurrentPosition(analysis); err != nil {
		return nil, fmt.Errorf("failed to analyze current position: %w", err)
	}

	// Calculate completion metrics
	if err := wa.calculateCompletionMetrics(analysis); err != nil {
		analysis.Issues = append(analysis.Issues, fmt.Sprintf("Failed to calculate metrics: %v", err))
	}

	// Detect blockers and generate recommendations
	wa.detectBlockers(analysis)
	wa.generateRecommendations(analysis)

	return analysis, nil
}

// analyzeCurrentPosition determines the current position in the workflow hierarchy
func (wa *WorkflowAnalyzer) analyzeCurrentPosition(analysis *WorkflowAnalysis) error {
	// Check for current epic
	currentEpic, err := wa.loadCurrentEpic()
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to load current epic: %w", err)
		}
		// No current epic - we're at project level
		analysis.Position = PositionProjectLevel
		return nil
	}

	analysis.CurrentEpic = currentEpic
	analysis.Position = PositionEpicLevel

	// Check for current story
	currentStory, err := wa.loadCurrentStory()
	if err != nil {
		if !os.IsNotExist(err) {
			analysis.Issues = append(analysis.Issues, fmt.Sprintf("Failed to load current story: %v", err))
		}
		// Have epic but no current story
		return nil
	}

	analysis.CurrentStory = currentStory
	analysis.Position = PositionStoryLevel

	// Check for current tasks
	currentTasks, err := wa.loadCurrentTasks()
	if err != nil {
		analysis.Issues = append(analysis.Issues, fmt.Sprintf("Failed to load current tasks: %v", err))
		return nil
	}

	if len(currentTasks) > 0 {
		analysis.CurrentTasks = currentTasks
		analysis.Position = PositionTaskLevel
	}

	return nil
}

// loadCurrentEpic loads the currently active epic
func (wa *WorkflowAnalyzer) loadCurrentEpic() (*state.EpicState, error) {
	epicPath := filepath.Join(wa.rootPath, "docs/2-current-epic/current-epic.json")
	
	data, err := os.ReadFile(epicPath)
	if err != nil {
		return nil, err
	}

	var epic state.EpicState
	if err := json.Unmarshal(data, &epic); err != nil {
		return nil, fmt.Errorf("invalid epic JSON: %w", err)
	}

	return &epic, nil
}

// loadCurrentStory loads the currently active story
func (wa *WorkflowAnalyzer) loadCurrentStory() (*state.StoryState, error) {
	// Try to load from stories.json and find the current one
	storiesPath := filepath.Join(wa.rootPath, "docs/2-current-epic/stories.json")
	
	data, err := os.ReadFile(storiesPath)
	if err != nil {
		return nil, err
	}

	var storiesFile struct {
		Stories []state.StoryState `json:"stories"`
		Meta    struct {
			CurrentStory string `json:"current_story,omitempty"`
		} `json:"meta,omitempty"`
	}

	if err := json.Unmarshal(data, &storiesFile); err != nil {
		return nil, fmt.Errorf("invalid stories JSON: %w", err)
	}

	// Find current story (either by meta.current_story or by status in_progress)
	for _, story := range storiesFile.Stories {
		if story.Metadata.ID == storiesFile.Meta.CurrentStory || 
		   story.Status == state.StatusInProgress {
			return &story, nil
		}
	}

	return nil, os.ErrNotExist // No current story found
}

// loadCurrentTasks loads currently active tasks
func (wa *WorkflowAnalyzer) loadCurrentTasks() ([]state.TaskState, error) {
	tasksDir := filepath.Join(wa.rootPath, "docs/3-current-task")
	
	// Check if tasks directory exists
	if _, err := os.Stat(tasksDir); err != nil {
		if os.IsNotExist(err) {
			return []state.TaskState{}, nil
		}
		return nil, err
	}

	// Read all JSON files in the tasks directory
	files, err := filepath.Glob(filepath.Join(tasksDir, "*.json"))
	if err != nil {
		return nil, err
	}

	var tasks []state.TaskState
	for _, file := range files {
		// Skip certain files that aren't task definitions
		if filepath.Base(file) == "current-task.json" {
			continue
		}

		data, err := os.ReadFile(file)
		if err != nil {
			continue // Skip files we can't read
		}

		// Try to parse as todo file first (common format)
		var todoFile struct {
			Todos []struct {
				ID       string `json:"id"`
				Title    string `json:"title"`
				Status   string `json:"status"`
				Priority string `json:"priority"`
			} `json:"todos"`
		}

		if err := json.Unmarshal(data, &todoFile); err == nil {
			// Convert todo items to task states
			for _, todo := range todoFile.Todos {
				if todo.Status == "todo" || todo.Status == "in_progress" || todo.Status == "blocked" {
					task := state.TaskState{
						Metadata: state.Metadata{
							ID: todo.ID,
						},
						Title:       todo.Title,
						Description: todo.Title,
						Status:      state.Status(todo.Status),
						Priority:    state.Priority(todo.Priority),
					}
					tasks = append(tasks, task)
				}
			}
			continue
		}

		// Try to parse as task state
		var task state.TaskState
		if err := json.Unmarshal(data, &task); err == nil {
			if task.Status == state.StatusTodo || task.Status == state.StatusInProgress || task.Status == state.StatusBlocked {
				tasks = append(tasks, task)
			}
		}
	}

	return tasks, nil
}

// calculateCompletionMetrics computes progress metrics at all levels
func (wa *WorkflowAnalyzer) calculateCompletionMetrics(analysis *WorkflowAnalysis) error {
	metrics := &analysis.CompletionMetrics

	// Load all epics to calculate project-level metrics
	epics, err := wa.loadAllEpics()
	if err != nil {
		return fmt.Errorf("failed to load epics: %w", err)
	}

	metrics.TotalEpics = len(epics)
	for _, epic := range epics {
		if epic.Status == state.StatusDone {
			metrics.CompletedEpics++
		}
		metrics.EstimatedHours += epic.EstimatedHours
		metrics.ActualHours += epic.ActualHours
	}

	if metrics.TotalEpics > 0 {
		metrics.ProjectProgress = float64(metrics.CompletedEpics) / float64(metrics.TotalEpics) * 100
	}

	// Calculate epic-level metrics if we have a current epic
	if analysis.CurrentEpic != nil {
		if err := wa.calculateEpicMetrics(analysis, metrics); err != nil {
			return fmt.Errorf("failed to calculate epic metrics: %w", err)
		}
	}

	// Calculate story-level metrics if we have a current story
	if analysis.CurrentStory != nil {
		wa.calculateStoryMetrics(analysis, metrics)
	}

	return nil
}

// loadAllEpics loads all epics from the epics.json file
func (wa *WorkflowAnalyzer) loadAllEpics() ([]state.EpicState, error) {
	epicsPath := filepath.Join(wa.rootPath, "docs/1-project/epics.json")
	
	data, err := os.ReadFile(epicsPath)
	if err != nil {
		return nil, err
	}

	var epicsFile struct {
		Epics []state.EpicState `json:"epics"`
	}

	if err := json.Unmarshal(data, &epicsFile); err != nil {
		return nil, fmt.Errorf("invalid epics JSON: %w", err)
	}

	return epicsFile.Epics, nil
}

// calculateEpicMetrics calculates metrics for the current epic
func (wa *WorkflowAnalyzer) calculateEpicMetrics(analysis *WorkflowAnalysis, metrics *CompletionMetrics) error {
	// Load stories for the current epic
	storiesPath := filepath.Join(wa.rootPath, "docs/2-current-epic/stories.json")
	
	data, err := os.ReadFile(storiesPath)
	if err != nil {
		return err
	}

	var storiesFile struct {
		Stories []state.StoryState `json:"stories"`
	}

	if err := json.Unmarshal(data, &storiesFile); err != nil {
		return fmt.Errorf("invalid stories JSON: %w", err)
	}

	metrics.TotalStories = len(storiesFile.Stories)
	for _, story := range storiesFile.Stories {
		if story.Status == state.StatusDone {
			metrics.CompletedStories++
		}
	}

	if metrics.TotalStories > 0 {
		metrics.EpicProgress = float64(metrics.CompletedStories) / float64(metrics.TotalStories) * 100
	}

	return nil
}

// calculateStoryMetrics calculates metrics for the current story
func (wa *WorkflowAnalyzer) calculateStoryMetrics(analysis *WorkflowAnalysis, metrics *CompletionMetrics) {
	// Count tasks from current tasks
	metrics.TotalTasks = len(analysis.CurrentTasks)
	for _, task := range analysis.CurrentTasks {
		if task.Status == state.StatusDone {
			metrics.CompletedTasks++
		}
	}

	if metrics.TotalTasks > 0 {
		metrics.StoryProgress = float64(metrics.CompletedTasks) / float64(metrics.TotalTasks) * 100
	}
}

// detectBlockers identifies issues that prevent workflow progression
func (wa *WorkflowAnalyzer) detectBlockers(analysis *WorkflowAnalysis) {
	// Check for epic without stories
	if analysis.CurrentEpic != nil && analysis.CurrentStory == nil {
		analysis.Blockers = append(analysis.Blockers, WorkflowBlocker{
			Type:        BlockerMissingDefinition,
			Severity:    "high",
			Description: "Epic is selected but no stories are defined",
			Entity:      analysis.CurrentEpic.Metadata.ID,
			Suggestion:  "Create stories for the current epic",
		})
	}

	// Check for story without tasks
	if analysis.CurrentStory != nil && len(analysis.CurrentTasks) == 0 {
		analysis.Blockers = append(analysis.Blockers, WorkflowBlocker{
			Type:        BlockerMissingDefinition,
			Severity:    "medium",
			Description: "Story is selected but no tasks are defined",
			Entity:      analysis.CurrentStory.Metadata.ID,
			Suggestion:  "Create tasks for the current story",
		})
	}

	// Check for inconsistent states
	if analysis.CurrentEpic != nil && analysis.CurrentEpic.Status == state.StatusDone {
		if analysis.Position >= PositionStoryLevel {
			analysis.Blockers = append(analysis.Blockers, WorkflowBlocker{
				Type:        BlockerInconsistentState,
				Severity:    "critical",
				Description: "Epic is marked as done but work is still in progress",
				Entity:      analysis.CurrentEpic.Metadata.ID,
				Suggestion:  "Review epic status or complete remaining work",
			})
		}
	}

	// Check for blocked tasks
	for _, task := range analysis.CurrentTasks {
		if task.Status == state.StatusBlocked {
			analysis.Blockers = append(analysis.Blockers, WorkflowBlocker{
				Type:        BlockerMissingDependency,
				Severity:    "high",
				Description: fmt.Sprintf("Task '%s' is blocked", task.Title),
				Entity:      task.Metadata.ID,
				Suggestion:  "Resolve blocking dependencies",
			})
		}
	}
}

// generateRecommendations creates actionable recommendations based on the current state
func (wa *WorkflowAnalyzer) generateRecommendations(analysis *WorkflowAnalysis) {
	switch analysis.Position {
	case PositionNotInitialized:
		analysis.Recommendations = append(analysis.Recommendations, "Initialize project structure")

	case PositionProjectLevel:
		analysis.Recommendations = append(analysis.Recommendations, "Create your first epic to start organizing work")

	case PositionEpicLevel:
		analysis.Recommendations = append(analysis.Recommendations, "Break down the epic into user stories")

	case PositionStoryLevel:
		analysis.Recommendations = append(analysis.Recommendations, "Create tasks to implement the current story")

	case PositionTaskLevel:
		// Find tasks that are todo or in_progress
		inProgressTasks := 0
		todoTasks := 0
		for _, task := range analysis.CurrentTasks {
			if task.Status == state.StatusInProgress {
				inProgressTasks++
			} else if task.Status == state.StatusTodo {
				todoTasks++
			}
		}

		if inProgressTasks > 0 {
			analysis.Recommendations = append(analysis.Recommendations, "Continue working on the current task")
		} else if todoTasks > 0 {
			analysis.Recommendations = append(analysis.Recommendations, "Start working on the next task")
		} else {
			analysis.Recommendations = append(analysis.Recommendations, "Mark the current story as complete")
		}
	}

	// Add recommendations based on completion metrics
	if analysis.CompletionMetrics.ProjectProgress > 80 {
		analysis.Recommendations = append(analysis.Recommendations, "Project is near completion - review final deliverables")
	}

	// Add recommendations based on blockers
	if len(analysis.Blockers) > 0 {
		analysis.Recommendations = append(analysis.Recommendations, "Resolve workflow blockers to continue progress")
	}
}

// GetWorkflowCapabilities returns the current capabilities based on workflow position
func (wa *WorkflowAnalyzer) GetWorkflowCapabilities(analysis *WorkflowAnalysis) []string {
	capabilities := []string{}

	switch analysis.Position {
	case PositionNotInitialized:
		capabilities = append(capabilities, "init-project")

	case PositionProjectLevel:
		capabilities = append(capabilities, "create-epic", "list-epics", "status")

	case PositionEpicLevel:
		capabilities = append(capabilities, "create-story", "start-epic", "list-stories", "create-epic", "status")

	case PositionStoryLevel:
		capabilities = append(capabilities, "create-task", "continue-story", "list-tasks", "create-story", "status")

	case PositionTaskLevel:
		capabilities = append(capabilities, "continue-task", "complete-task", "create-task", "status")
	}

	// Always add help
	capabilities = append(capabilities, "help")

	return capabilities
}