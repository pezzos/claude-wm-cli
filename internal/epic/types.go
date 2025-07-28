package epic

import (
	"time"
)

// Epic represents a large unit of work composed of multiple user stories
type Epic struct {
	ID           string          `json:"id"`
	Title        string          `json:"title"`
	Description  string          `json:"description"`
	Priority     Priority        `json:"priority"`
	Status       Status          `json:"status"`
	StartDate    *time.Time      `json:"start_date,omitempty"`
	EndDate      *time.Time      `json:"end_date,omitempty"`
	Duration     string          `json:"duration,omitempty"`
	Tags         []string        `json:"tags,omitempty"`
	Dependencies []string        `json:"dependencies,omitempty"`
	UserStories  []UserStory     `json:"user_stories,omitempty"`
	Progress     ProgressMetrics `json:"progress"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

// Priority represents the priority level of an epic
type Priority string

const (
	PriorityLow      Priority = "low"
	PriorityMedium   Priority = "medium"
	PriorityHigh     Priority = "high"
	PriorityCritical Priority = "critical"
)

// Status represents the current status of an epic
type Status string

const (
	StatusPlanned    Status = "planned"
	StatusInProgress Status = "in_progress"
	StatusOnHold     Status = "on_hold"
	StatusCompleted  Status = "completed"
	StatusCancelled  Status = "cancelled"
)

// UserStory represents a user story within an epic
type UserStory struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Priority    Priority `json:"priority"`
	Status      Status   `json:"status"`
	StoryPoints int      `json:"story_points,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

// ProgressMetrics tracks the progress of an epic
type ProgressMetrics struct {
	TotalStoryPoints     int        `json:"total_story_points"`
	CompletedStoryPoints int        `json:"completed_story_points"`
	TotalStories         int        `json:"total_stories"`
	CompletedStories     int        `json:"completed_stories"`
	CompletionPercentage float64    `json:"completion_percentage"`
	EstimatedEndDate     *time.Time `json:"estimated_end_date,omitempty"`
}

// EpicCollection represents a collection of epics with metadata
type EpicCollection struct {
	ProjectID   string             `json:"project_id"`
	Epics       map[string]*Epic   `json:"epics"`
	CurrentEpic string             `json:"current_epic,omitempty"`
	Metadata    CollectionMetadata `json:"metadata"`
}

// CollectionMetadata contains metadata about the epic collection
type CollectionMetadata struct {
	Version     string    `json:"version"`
	LastUpdated time.Time `json:"last_updated"`
	TotalEpics  int       `json:"total_epics"`
}

// EpicCreateOptions contains options for creating a new epic
type EpicCreateOptions struct {
	Title        string
	Description  string
	Priority     Priority
	Duration     string
	Tags         []string
	Dependencies []string
}

// EpicUpdateOptions contains options for updating an epic
type EpicUpdateOptions struct {
	Title        *string
	Description  *string
	Priority     *Priority
	Status       *Status
	Duration     *string
	Tags         *[]string
	Dependencies *[]string
}

// EpicListOptions contains options for listing epics
type EpicListOptions struct {
	Status   Status
	Priority Priority
	ShowAll  bool
}

// String returns the string representation of priority
func (p Priority) String() string {
	return string(p)
}

// String returns the string representation of status
func (s Status) String() string {
	return string(s)
}

// IsValid checks if the priority is valid
func (p Priority) IsValid() bool {
	switch p {
	case PriorityLow, PriorityMedium, PriorityHigh, PriorityCritical:
		return true
	default:
		return false
	}
}

// IsValid checks if the status is valid
func (s Status) IsValid() bool {
	switch s {
	case StatusPlanned, StatusInProgress, StatusOnHold, StatusCompleted, StatusCancelled:
		return true
	default:
		return false
	}
}

// CalculateProgress updates the progress metrics for an epic
func (e *Epic) CalculateProgress() {
	if len(e.UserStories) == 0 {
		e.Progress = ProgressMetrics{}
		return
	}

	totalStoryPoints := 0
	completedStoryPoints := 0
	completedStories := 0

	for _, story := range e.UserStories {
		totalStoryPoints += story.StoryPoints
		if story.Status == StatusCompleted {
			completedStoryPoints += story.StoryPoints
			completedStories++
		}
	}

	completionPercentage := float64(0)
	if totalStoryPoints > 0 {
		completionPercentage = float64(completedStoryPoints) / float64(totalStoryPoints) * 100
	} else if len(e.UserStories) > 0 {
		completionPercentage = float64(completedStories) / float64(len(e.UserStories)) * 100
	}

	e.Progress = ProgressMetrics{
		TotalStoryPoints:     totalStoryPoints,
		CompletedStoryPoints: completedStoryPoints,
		TotalStories:         len(e.UserStories),
		CompletedStories:     completedStories,
		CompletionPercentage: completionPercentage,
	}
}

// IsActive returns true if the epic is currently active (in progress)
func (e *Epic) IsActive() bool {
	return e.Status == StatusInProgress
}

// CanStart returns true if the epic can be started
func (e *Epic) CanStart() bool {
	return e.Status == StatusPlanned
}

// CanComplete returns true if the epic can be completed
func (e *Epic) CanComplete() bool {
	return e.Status == StatusInProgress && e.Progress.CompletionPercentage >= 100
}
