// Package entities contains domain entities that represent core business objects.
// Entities have identity and lifecycle, and encapsulate business rules and behaviors.
package entities

import (
	"fmt"
	"strings"
	"time"

	"claude-wm-cli/internal/domain/valueobjects"
)

// Epic represents a large unit of work composed of multiple user stories.
// This is a domain entity that encapsulates all business logic related to epics.
type Epic struct {
	id           string
	title        string
	description  string
	priority     valueobjects.Priority
	status       valueobjects.Status
	startDate    *time.Time
	endDate      *time.Time
	duration     string
	tags         []string
	dependencies []string
	userStories  []UserStory
	progress     ProgressMetrics
	createdAt    time.Time
	updatedAt    time.Time
}

// UserStory represents a user story within an epic.
type UserStory struct {
	ID          string                `json:"id"`
	Title       string                `json:"title"`
	Description string                `json:"description"`
	Priority    valueobjects.Priority `json:"priority"`
	Status      valueobjects.Status   `json:"status"`
	StoryPoints int                   `json:"story_points,omitempty"`
	Tags        []string              `json:"tags,omitempty"`
}

// ProgressMetrics tracks the progress of an epic.
type ProgressMetrics struct {
	TotalStoryPoints     int        `json:"total_story_points"`
	CompletedStoryPoints int        `json:"completed_story_points"`
	TotalStories         int        `json:"total_stories"`
	CompletedStories     int        `json:"completed_stories"`
	CompletionPercentage float64    `json:"completion_percentage"`
	EstimatedHours       int        `json:"estimated_hours"`
	ActualHours          int        `json:"actual_hours"`
	EstimatedEndDate     *time.Time `json:"estimated_end_date,omitempty"`
}

// NewEpic creates a new Epic with the provided details.
// This is the factory method that ensures proper initialization.
func NewEpic(id, title, description string, priority valueobjects.Priority) (*Epic, error) {
	if err := validateEpicID(id); err != nil {
		return nil, err
	}
	if err := validateEpicTitle(title); err != nil {
		return nil, err
	}
	if err := validateEpicDescription(description); err != nil {
		return nil, err
	}
	if !priority.IsValid() {
		return nil, fmt.Errorf("invalid priority: %s", priority)
	}

	now := time.Now()
	return &Epic{
		id:          id,
		title:       title,
		description: description,
		priority:    priority,
		status:      valueobjects.Planned,
		tags:        []string{},
		dependencies: []string{},
		userStories: []UserStory{},
		progress:    ProgressMetrics{},
		createdAt:   now,
		updatedAt:   now,
	}, nil
}

// ID returns the epic's unique identifier.
func (e *Epic) ID() string {
	return e.id
}

// Title returns the epic's title.
func (e *Epic) Title() string {
	return e.title
}

// Description returns the epic's description.
func (e *Epic) Description() string {
	return e.description
}

// Priority returns the epic's priority.
func (e *Epic) Priority() valueobjects.Priority {
	return e.priority
}

// Status returns the epic's current status.
func (e *Epic) Status() valueobjects.Status {
	return e.status
}

// StartDate returns the epic's start date.
func (e *Epic) StartDate() *time.Time {
	return e.startDate
}

// EndDate returns the epic's end date.
func (e *Epic) EndDate() *time.Time {
	return e.endDate
}

// Duration returns the epic's duration string.
func (e *Epic) Duration() string {
	return e.duration
}

// Tags returns the epic's tags.
func (e *Epic) Tags() []string {
	// Return a copy to maintain immutability
	tags := make([]string, len(e.tags))
	copy(tags, e.tags)
	return tags
}

// Dependencies returns the epic's dependencies.
func (e *Epic) Dependencies() []string {
	// Return a copy to maintain immutability
	deps := make([]string, len(e.dependencies))
	copy(deps, e.dependencies)
	return deps
}

// UserStories returns the epic's user stories.
func (e *Epic) UserStories() []UserStory {
	// Return a copy to maintain immutability
	stories := make([]UserStory, len(e.userStories))
	copy(stories, e.userStories)
	return stories
}

// Progress returns the epic's progress metrics.
func (e *Epic) Progress() ProgressMetrics {
	return e.progress
}

// CreatedAt returns when the epic was created.
func (e *Epic) CreatedAt() time.Time {
	return e.createdAt
}

// UpdatedAt returns when the epic was last updated.
func (e *Epic) UpdatedAt() time.Time {
	return e.updatedAt
}

// UpdateTitle updates the epic's title with validation.
func (e *Epic) UpdateTitle(title string) error {
	if err := validateEpicTitle(title); err != nil {
		return err
	}
	e.title = title
	e.updatedAt = time.Now()
	return nil
}

// UpdateDescription updates the epic's description with validation.
func (e *Epic) UpdateDescription(description string) error {
	if err := validateEpicDescription(description); err != nil {
		return err
	}
	e.description = description
	e.updatedAt = time.Now()
	return nil
}

// UpdatePriority updates the epic's priority with validation.
func (e *Epic) UpdatePriority(priority valueobjects.Priority) error {
	if !priority.IsValid() {
		return fmt.Errorf("invalid priority: %s", priority)
	}
	e.priority = priority
	e.updatedAt = time.Now()
	return nil
}

// TransitionTo changes the epic's status if the transition is valid.
func (e *Epic) TransitionTo(newStatus valueobjects.Status) error {
	if !e.status.CanTransitionTo(newStatus) {
		return fmt.Errorf("invalid status transition from %s to %s", e.status, newStatus)
	}
	e.status = newStatus
	e.updatedAt = time.Now()
	return nil
}

// SetDuration updates the epic's duration.
func (e *Epic) SetDuration(duration string) {
	e.duration = duration
	e.updatedAt = time.Now()
}

// SetStartDate updates the epic's start date.
func (e *Epic) SetStartDate(startDate *time.Time) {
	e.startDate = startDate
	e.updatedAt = time.Now()
}

// SetEndDate updates the epic's end date.
func (e *Epic) SetEndDate(endDate *time.Time) {
	e.endDate = endDate
	e.updatedAt = time.Now()
}

// AddTag adds a tag to the epic if it doesn't already exist.
func (e *Epic) AddTag(tag string) {
	if tag == "" || e.HasTag(tag) {
		return
	}
	e.tags = append(e.tags, tag)
	e.updatedAt = time.Now()
}

// RemoveTag removes a tag from the epic.
func (e *Epic) RemoveTag(tag string) {
	for i, t := range e.tags {
		if t == tag {
			e.tags = append(e.tags[:i], e.tags[i+1:]...)
			e.updatedAt = time.Now()
			break
		}
	}
}

// HasTag checks if the epic has a specific tag.
func (e *Epic) HasTag(tag string) bool {
	for _, t := range e.tags {
		if t == tag {
			return true
		}
	}
	return false
}

// AddDependency adds a dependency to the epic if it doesn't already exist.
func (e *Epic) AddDependency(dependency string) {
	if dependency == "" || e.HasDependency(dependency) || dependency == e.id {
		return
	}
	e.dependencies = append(e.dependencies, dependency)
	e.updatedAt = time.Now()
}

// RemoveDependency removes a dependency from the epic.
func (e *Epic) RemoveDependency(dependency string) {
	for i, d := range e.dependencies {
		if d == dependency {
			e.dependencies = append(e.dependencies[:i], e.dependencies[i+1:]...)
			e.updatedAt = time.Now()
			break
		}
	}
}

// HasDependency checks if the epic has a specific dependency.
func (e *Epic) HasDependency(dependency string) bool {
	for _, d := range e.dependencies {
		if d == dependency {
			return true
		}
	}
	return false
}

// AddUserStory adds a user story to the epic.
func (e *Epic) AddUserStory(story UserStory) error {
	if story.ID == "" {
		return fmt.Errorf("user story ID cannot be empty")
	}
	
	// Check for duplicate IDs
	for _, existingStory := range e.userStories {
		if existingStory.ID == story.ID {
			return fmt.Errorf("user story with ID %s already exists", story.ID)
		}
	}
	
	e.userStories = append(e.userStories, story)
	e.CalculateProgress()
	e.updatedAt = time.Now()
	return nil
}

// RemoveUserStory removes a user story from the epic.
func (e *Epic) RemoveUserStory(storyID string) {
	for i, story := range e.userStories {
		if story.ID == storyID {
			e.userStories = append(e.userStories[:i], e.userStories[i+1:]...)
			e.CalculateProgress()
			e.updatedAt = time.Now()
			break
		}
	}
}

// UpdateUserStory updates an existing user story.
func (e *Epic) UpdateUserStory(storyID string, updatedStory UserStory) error {
	for i, story := range e.userStories {
		if story.ID == storyID {
			updatedStory.ID = storyID // Preserve the ID
			e.userStories[i] = updatedStory
			e.CalculateProgress()
			e.updatedAt = time.Now()
			return nil
		}
	}
	return fmt.Errorf("user story with ID %s not found", storyID)
}

// CalculateProgress updates the progress metrics for the epic.
func (e *Epic) CalculateProgress() {
	if len(e.userStories) == 0 {
		e.progress = ProgressMetrics{}
		return
	}

	totalStoryPoints := 0
	completedStoryPoints := 0
	completedStories := 0

	for _, story := range e.userStories {
		totalStoryPoints += story.StoryPoints
		if story.Status == valueobjects.Completed {
			completedStoryPoints += story.StoryPoints
			completedStories++
		}
	}

	completionPercentage := float64(0)
	if totalStoryPoints > 0 {
		completionPercentage = float64(completedStoryPoints) / float64(totalStoryPoints) * 100
	} else if len(e.userStories) > 0 {
		completionPercentage = float64(completedStories) / float64(len(e.userStories)) * 100
	}

	e.progress = ProgressMetrics{
		TotalStoryPoints:     totalStoryPoints,
		CompletedStoryPoints: completedStoryPoints,
		TotalStories:         len(e.userStories),
		CompletedStories:     completedStories,
		CompletionPercentage: completionPercentage,
	}
}

// IsActive returns true if the epic is currently active (in progress).
func (e *Epic) IsActive() bool {
	return e.status.IsActive()
}

// CanStart returns true if the epic can be started.
func (e *Epic) CanStart() bool {
	return e.status == valueobjects.Planned
}

// CanComplete returns true if the epic can be completed.
func (e *Epic) CanComplete() bool {
	return e.status == valueobjects.InProgress && e.progress.CompletionPercentage >= 100
}

// IsCompleted returns true if the epic is completed.
func (e *Epic) IsCompleted() bool {
	return e.status == valueobjects.Completed
}

// GetSearchableText returns text content for search indexing.
func (e *Epic) GetSearchableText() string {
	parts := []string{e.title, e.description}
	
	if len(e.tags) > 0 {
		parts = append(parts, strings.Join(e.tags, " "))
	}
	
	for _, story := range e.userStories {
		if story.Title != "" {
			parts = append(parts, story.Title)
		}
	}
	
	return strings.Join(parts, " ")
}

// Validate validates the epic's current state.
func (e *Epic) Validate() error {
	if err := validateEpicID(e.id); err != nil {
		return err
	}
	if err := validateEpicTitle(e.title); err != nil {
		return err
	}
	if err := validateEpicDescription(e.description); err != nil {
		return err
	}
	if !e.priority.IsValid() {
		return fmt.Errorf("invalid priority: %s", e.priority)
	}
	if !e.status.IsValid() {
		return fmt.Errorf("invalid status: %s", e.status)
	}
	return nil
}

// Validation functions

func validateEpicID(id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("epic ID cannot be empty")
	}
	if len(id) > 100 {
		return fmt.Errorf("epic ID too long (max 100 characters): got %d", len(id))
	}
	return nil
}

func validateEpicTitle(title string) error {
	if strings.TrimSpace(title) == "" {
		return fmt.Errorf("epic title cannot be empty")
	}
	if len(title) > 200 {
		return fmt.Errorf("epic title too long (max 200 characters): got %d", len(title))
	}
	return nil
}

func validateEpicDescription(description string) error {
	if strings.TrimSpace(description) == "" {
		return fmt.Errorf("epic description cannot be empty")
	}
	if len(description) > 2000 {
		return fmt.Errorf("epic description too long (max 2000 characters): got %d", len(description))
	}
	return nil
}