package story

import (
	"time"

	"claude-wm-cli/internal/model"
)

// Status represents the status of a story or task
type Status = model.Status

// Priority represents the priority of a story
type Priority = model.Priority

// Story represents an individual user story
type Story struct {
	ID                 string     `json:"id"`
	Title              string     `json:"title"`
	Description        string     `json:"description"`
	EpicID             string     `json:"epic_id"`
	Status             Status     `json:"status"`
	Priority           Priority   `json:"priority"`
	StoryPoints        int        `json:"story_points"`
	AcceptanceCriteria []string   `json:"acceptance_criteria"`
	Tasks              []Task     `json:"tasks"`
	Dependencies       []string   `json:"dependencies,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
	StartedAt          *time.Time `json:"started_at,omitempty"`
	CompletedAt        *time.Time `json:"completed_at,omitempty"`
}

// Task represents a task within a story (generated from acceptance criteria)
type Task struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      Status    `json:"status"`
	StoryID     string    `json:"story_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// StoryCollection represents the collection of all stories
type StoryCollection struct {
	Stories      map[string]*Story  `json:"stories"`
	CurrentStory string             `json:"current_story,omitempty"`
	Metadata     CollectionMetadata `json:"metadata"`
}

// CollectionMetadata contains metadata about the story collection
type CollectionMetadata struct {
	Version      string    `json:"version"`
	LastUpdated  time.Time `json:"last_updated"`
	TotalStories int       `json:"total_stories"`
	TotalTasks   int       `json:"total_tasks"`
}

// StoryCreateOptions contains options for creating a new story
type StoryCreateOptions struct {
	Title              string
	Description        string
	EpicID             string
	Priority           Priority
	StoryPoints        int
	AcceptanceCriteria []string
	Dependencies       []string
}

// StoryUpdateOptions contains options for updating an existing story
type StoryUpdateOptions struct {
	Title              *string
	Description        *string
	Status             *Status
	Priority           *Priority
	StoryPoints        *int
	AcceptanceCriteria *[]string
	Dependencies       *[]string
}

// TaskCreateOptions contains options for creating a new task
type TaskCreateOptions struct {
	Title       string
	Description string
	StoryID     string
}

// TaskUpdateOptions contains options for updating an existing task
type TaskUpdateOptions struct {
	Title       *string
	Description *string
	Status      *Status
}

// ProgressMetrics contains progress information for a story
type ProgressMetrics struct {
	CompletionPercentage float64 `json:"completion_percentage"`
	TotalTasks           int     `json:"total_tasks"`
	CompletedTasks       int     `json:"completed_tasks"`
	InProgressTasks      int     `json:"in_progress_tasks"`
	PendingTasks         int     `json:"pending_tasks"`
}

// CalculateProgress calculates and updates story progress metrics
func (s *Story) CalculateProgress() ProgressMetrics {
	totalTasks := len(s.Tasks)
	if totalTasks == 0 {
		return ProgressMetrics{
			CompletionPercentage: 0.0,
			TotalTasks:           0,
			CompletedTasks:       0,
			InProgressTasks:      0,
			PendingTasks:         0,
		}
	}

	var completedTasks, inProgressTasks, pendingTasks int
	for _, task := range s.Tasks {
		switch task.Status {
		case model.StatusCompleted:
			completedTasks++
		case model.StatusInProgress:
			inProgressTasks++
		case model.StatusPlanned:
			pendingTasks++
		}
	}

	completionPercentage := float64(completedTasks) / float64(totalTasks) * 100.0

	return ProgressMetrics{
		CompletionPercentage: completionPercentage,
		TotalTasks:           totalTasks,
		CompletedTasks:       completedTasks,
		InProgressTasks:      inProgressTasks,
		PendingTasks:         pendingTasks,
	}
}

// IsActive returns true if the story is currently being worked on
func (s *Story) IsActive() bool {
	return s.Status == model.StatusInProgress
}

// CanStart returns true if the story can be started
func (s *Story) CanStart() bool {
	return s.Status == model.StatusPlanned
}

// CanComplete returns true if the story can be completed
func (s *Story) CanComplete() bool {
	if s.Status != model.StatusInProgress {
		return false
	}

	// Check if all tasks are completed
	for _, task := range s.Tasks {
		if task.Status != model.StatusCompleted {
			return false
		}
	}

	return true
}

// GetTaskByID returns a task by its ID
func (s *Story) GetTaskByID(taskID string) *Task {
	for i := range s.Tasks {
		if s.Tasks[i].ID == taskID {
			return &s.Tasks[i]
		}
	}
	return nil
}

// AddTask adds a new task to the story
func (s *Story) AddTask(task Task) {
	s.Tasks = append(s.Tasks, task)
	s.UpdatedAt = time.Now()
}

// RemoveTask removes a task from the story
func (s *Story) RemoveTask(taskID string) bool {
	for i, task := range s.Tasks {
		if task.ID == taskID {
			s.Tasks = append(s.Tasks[:i], s.Tasks[i+1:]...)
			s.UpdatedAt = time.Now()
			return true
		}
	}
	return false
}
