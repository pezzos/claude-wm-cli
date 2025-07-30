package epic

import (
	"time"

	"claude-wm-cli/internal/model"
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
// Now uses the centralized model.Priority type for consistency
type Priority = model.Priority

// Status represents the current status of an epic  
// Now uses the centralized model.Status type for consistency
type Status = model.Status

// Legacy priority constants for backward compatibility
// These map to the standardized P0-P3 system
const (
	PriorityLow      = model.PriorityP3 // Low priority maps to P3
	PriorityMedium   = model.PriorityP2 // Medium priority maps to P2
	PriorityHigh     = model.PriorityP1 // High priority maps to P1
	PriorityCritical = model.PriorityP0 // Critical priority maps to P0
)

// Legacy status constants for backward compatibility
const (
	StatusPlanned    = model.StatusPlanned
	StatusInProgress = model.StatusInProgress
	StatusOnHold     = model.StatusOnHold
	StatusCompleted  = model.StatusCompleted
	StatusCancelled  = model.StatusCancelled
)

// MigrateLegacyPriority converts legacy priority strings to standardized Priority
func MigrateLegacyPriority(legacy string) Priority {
	switch legacy {
	case "critical":
		return model.PriorityP0
	case "high":
		return model.PriorityP1
	case "medium":
		return model.PriorityP2
	case "low":
		return model.PriorityP3
	default:
		return model.PriorityP2 // Default to medium if unknown
	}
}

// MigrateLegacyStatus converts legacy status strings to standardized Status
func MigrateLegacyStatus(legacy string) Status {
	return model.StatusFromLegacy(legacy)
}

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

// Note: String() and IsValid() methods are now available through model.Priority and model.Status
// No need to redefine them here as they are inherited from the model types

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
