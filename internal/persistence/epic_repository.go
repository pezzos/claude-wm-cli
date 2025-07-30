// Package persistence provides repository implementations for epic entities.
package persistence

import (
	"context"
	"fmt"
	"path/filepath"
	"reflect"

	"claude-wm-cli/internal/epic"
	"claude-wm-cli/internal/model"
)

// EpicRepository provides CRUD operations for Epic entities.
type EpicRepository struct {
	model.Repository[epic.Epic]
}

// NewEpicRepository creates a new repository for Epic entities.
//
// Parameters:
//   - dataDir: Directory where epic data will be stored
//
// Returns a repository that implements model.Repository[epic.Epic].
func NewEpicRepository(dataDir string) *EpicRepository {
	filePath := filepath.Join(dataDir, "epics.json")
	
	options := DefaultRepositoryOptions()
	options.EnableCache = true
	options.CacheTTL = DefaultRepositoryOptions().CacheTTL
	
	repo := NewJSONRepository[epic.Epic](
		filePath,
		validateEpic,
		options,
	)

	return &EpicRepository{
		Repository: repo,
	}
}

// validateEpic validates an Epic entity before persistence.
func validateEpic(e epic.Epic) error {
	var errors model.ValidationErrors

	// Validate required fields
	if e.ID == "" {
		errors.Add("id", e.ID, "ID is required")
	}
	if e.Title == "" {
		errors.Add("title", e.Title, "title is required")
	}

	// Validate Priority
	if !e.Priority.IsValid() {
		errors.Add("priority", string(e.Priority), "invalid priority value")
	}

	// Validate Status
	if !e.Status.IsValid() {
		errors.Add("status", string(e.Status), "invalid status value")
	}

	// Validate timestamps
	if e.CreatedAt.IsZero() {
		errors.Add("created_at", e.CreatedAt.String(), "created_at is required")
	}
	if e.UpdatedAt.IsZero() {
		errors.Add("updated_at", e.UpdatedAt.String(), "updated_at is required")
	}

	// Validate user stories
	for i, story := range e.UserStories {
		if story.ID == "" {
			errors.Add("user_stories", story.ID, 
				fmt.Sprintf("user story %d ID is required", i))
		}
		if story.Title == "" {
			errors.Add("user_stories", story.Title, 
				fmt.Sprintf("user story %d title is required", i))
		}
	}

	if errors.HasErrors() {
		return errors
	}

	return nil
}

// FindByStatus returns all epics with the specified status.
func (r *EpicRepository) FindByStatus(ctx context.Context, status epic.Status) ([]epic.Epic, error) {
	filter := &StatusFilter[epic.Epic]{
		Status: status,
	}
	return r.List(ctx, filter)
}

// FindByPriority returns all epics with the specified priority.
func (r *EpicRepository) FindByPriority(ctx context.Context, priority epic.Priority) ([]epic.Epic, error) {
	filter := &PriorityFilter[epic.Epic]{
		Priority: priority,
	}
	return r.List(ctx, filter)
}

// FindActive returns all epics that are currently active (in progress).
func (r *EpicRepository) FindActive(ctx context.Context) ([]epic.Epic, error) {
	return r.FindByStatus(ctx, epic.StatusInProgress)
}

// FindCompleted returns all epics that have been completed.
func (r *EpicRepository) FindCompleted(ctx context.Context) ([]epic.Epic, error) {
	return r.FindByStatus(ctx, epic.StatusCompleted)
}

// StatusFilter filters entities by status.
type StatusFilter[T any] struct {
	Status interface{} // Can be any status type (epic.Status, story.Status, etc.)
}

// Apply checks if an entity matches the status filter.
func (f *StatusFilter[T]) Apply(entity interface{}) bool {
	// Use reflection to get Status field
	v := reflect.ValueOf(entity)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	
	statusField := v.FieldByName("Status")
	if !statusField.IsValid() {
		return false
	}
	
	return statusField.Interface() == f.Status
}

// ToSQL converts the filter to SQL (not implemented for JSON repository).
func (f *StatusFilter[T]) ToSQL() (string, []interface{}) {
	return "status = ?", []interface{}{f.Status}
}

// Validate checks if the filter is valid.
func (f *StatusFilter[T]) Validate() error {
	if f.Status == nil {
		return model.NewValidationError("status cannot be nil")
	}
	return nil
}

// PriorityFilter filters entities by priority.
type PriorityFilter[T any] struct {
	Priority interface{} // Can be any priority type
}

// Apply checks if an entity matches the priority filter.
func (f *PriorityFilter[T]) Apply(entity interface{}) bool {
	v := reflect.ValueOf(entity)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	
	priorityField := v.FieldByName("Priority")
	if !priorityField.IsValid() {
		return false
	}
	
	return priorityField.Interface() == f.Priority
}

// ToSQL converts the filter to SQL (not implemented for JSON repository).
func (f *PriorityFilter[T]) ToSQL() (string, []interface{}) {
	return "priority = ?", []interface{}{f.Priority}
}

// Validate checks if the filter is valid.
func (f *PriorityFilter[T]) Validate() error {
	if f.Priority == nil {
		return model.NewValidationError("priority cannot be nil")
	}
	return nil
}