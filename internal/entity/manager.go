// Package entity provides generic entity management functionality.
// This package eliminates CRUD duplication across epic, story, and ticket packages.
package entity

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"claude-wm-cli/internal/model"
)

// BaseEntity defines the common interface for all entities.
type BaseEntity interface {
	GetID() string
	SetID(string)
	GetStatus() model.Status
	SetStatus(model.Status)
	GetPriority() model.Priority
	SetPriority(model.Priority)
	GetCreatedAt() time.Time
	SetCreatedAt(time.Time)
	GetUpdatedAt() time.Time  
	SetUpdatedAt(time.Time)
	Validate() error
}

// EntityManager provides generic CRUD operations for any entity type.
// This eliminates the duplication found in epic/story/ticket managers.
type EntityManager[T BaseEntity] struct {
	repo        model.Repository[T]
	entityType  string
	validator   ValidatorFunc[T]
	beforeSave  []HookFunc[T]
	afterSave   []HookFunc[T]
	beforeDelete []HookFunc[T]
	afterDelete  []HookFunc[T]
}

// ValidatorFunc defines the signature for entity validation functions.
type ValidatorFunc[T BaseEntity] func(T) error

// HookFunc defines the signature for entity lifecycle hooks.
type HookFunc[T BaseEntity] func(context.Context, T) error

// ManagerOptions configures entity manager behavior.
type ManagerOptions[T BaseEntity] struct {
	EntityType   string
	Validator    ValidatorFunc[T]
	BeforeSave   []HookFunc[T]
	AfterSave    []HookFunc[T]
	BeforeDelete []HookFunc[T]
	AfterDelete  []HookFunc[T]
}

// NewEntityManager creates a new generic entity manager.
//
// Parameters:
//   - repo: Repository implementation for the entity type
//   - options: Configuration options for the manager
//
// Returns a manager that provides standard CRUD operations with hooks and validation.
func NewEntityManager[T BaseEntity](repo model.Repository[T], options ManagerOptions[T]) *EntityManager[T] {
	if options.EntityType == "" {
		// Extract entity type name from the generic type
		var zero T
		t := reflect.TypeOf(zero)
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
		options.EntityType = t.Name()
	}

	return &EntityManager[T]{
		repo:         repo,
		entityType:   options.EntityType,
		validator:    options.Validator,
		beforeSave:   options.BeforeSave,
		afterSave:    options.AfterSave,
		beforeDelete: options.BeforeDelete,
		afterDelete:  options.AfterDelete,
	}
}

// CreateEntityRequest represents a request to create a new entity.
type CreateEntityRequest[T BaseEntity] struct {
	Entity   T
	SkipHooks bool
}

// CreateEntityResponse represents the response from creating an entity.
type CreateEntityResponse[T BaseEntity] struct {
	Entity    T
	Created   bool
	Warnings  []string
	Metadata  map[string]interface{}
}

// Create creates a new entity with full lifecycle management.
func (m *EntityManager[T]) Create(ctx context.Context, req CreateEntityRequest[T]) (*CreateEntityResponse[T], error) {
	// Set timestamps
	now := time.Now()
	req.Entity.SetCreatedAt(now)
	req.Entity.SetUpdatedAt(now)

	// Generate ID if not set
	if req.Entity.GetID() == "" {
		id, err := m.generateID(req.Entity)
		if err != nil {
			return nil, model.NewInternalError("failed to generate entity ID").
				WithCause(err).
				WithContext(m.entityType)
		}
		req.Entity.SetID(id)
	}

	// Validate entity
	if m.validator != nil {
		if err := m.validator(req.Entity); err != nil {
			return nil, model.NewValidationError("entity validation failed").
				WithCause(err).
				WithContext(fmt.Sprintf("%s ID: %s", m.entityType, req.Entity.GetID()))
		}
	}

	// Entity self-validation
	if err := req.Entity.Validate(); err != nil {
		return nil, model.NewValidationError("entity self-validation failed").
			WithCause(err).
			WithContext(fmt.Sprintf("%s ID: %s", m.entityType, req.Entity.GetID()))
	}

	// Execute beforeSave hooks
	if !req.SkipHooks {
		for _, hook := range m.beforeSave {
			if err := hook(ctx, req.Entity); err != nil {
				return nil, model.NewInternalError("beforeSave hook failed").
					WithCause(err).
					WithContext(fmt.Sprintf("%s ID: %s", m.entityType, req.Entity.GetID()))
			}
		}
	}

	// Save to repository
	if err := m.repo.Create(ctx, req.Entity); err != nil {
		return nil, model.NewInternalError(fmt.Sprintf("failed to create %s", m.entityType)).
			WithCause(err).
			WithContext(req.Entity.GetID()).
			WithSuggestions([]string{
				"Check if entity with same ID already exists",
				"Verify repository configuration",
				"Ensure data directory is writable",
			})
	}

	// Execute afterSave hooks
	var warnings []string
	if !req.SkipHooks {
		for _, hook := range m.afterSave {
			if err := hook(ctx, req.Entity); err != nil {
				// AfterSave hooks failures are non-critical - collect as warnings
				warnings = append(warnings, fmt.Sprintf("afterSave hook warning: %v", err))
			}
		}
	}

	return &CreateEntityResponse[T]{
		Entity:   req.Entity,
		Created:  true,
		Warnings: warnings,
		Metadata: map[string]interface{}{
			"entity_type": m.entityType,
			"created_at":  req.Entity.GetCreatedAt(),
		},
	}, nil
}

// UpdateEntityRequest represents a request to update an entity.
type UpdateEntityRequest[T BaseEntity] struct {
	ID        string
	Updates   func(T) T  // Function to apply updates
	SkipHooks bool
}

// UpdateEntityResponse represents the response from updating an entity.
type UpdateEntityResponse[T BaseEntity] struct {
	Entity    T
	Updated   bool
	Warnings  []string
	Changes   []string
	Metadata  map[string]interface{}
}

// Update updates an existing entity with change tracking.
func (m *EntityManager[T]) Update(ctx context.Context, req UpdateEntityRequest[T]) (*UpdateEntityResponse[T], error) {
	// Get existing entity
	existing, err := m.repo.Read(ctx, req.ID)
	if err != nil {
		return nil, model.NewNotFoundError(m.entityType).
			WithContext(req.ID).
			WithSuggestion(fmt.Sprintf("Use List() to see available %ss", m.entityType))
	}

	// Apply updates
	updated := req.Updates(existing)
	updated.SetUpdatedAt(time.Now())

	// Track changes (simplified - in real implementation, use reflection for detailed tracking)
	var changes []string
	if existing.GetStatus() != updated.GetStatus() {
		changes = append(changes, fmt.Sprintf("status: %s → %s", existing.GetStatus(), updated.GetStatus()))
	}
	if existing.GetPriority() != updated.GetPriority() {
		changes = append(changes, fmt.Sprintf("priority: %s → %s", existing.GetPriority(), updated.GetPriority()))
	}

	// Validate updated entity
	if m.validator != nil {
		if err := m.validator(updated); err != nil {
			return nil, model.NewValidationError("updated entity validation failed").
				WithCause(err).
				WithContext(fmt.Sprintf("%s ID: %s", m.entityType, req.ID))
		}
	}

	if err := updated.Validate(); err != nil {
		return nil, model.NewValidationError("updated entity self-validation failed").
			WithCause(err).
			WithContext(fmt.Sprintf("%s ID: %s", m.entityType, req.ID))
	}

	// Execute beforeSave hooks
	if !req.SkipHooks {
		for _, hook := range m.beforeSave {
			if err := hook(ctx, updated); err != nil {
				return nil, model.NewInternalError("beforeSave hook failed during update").
					WithCause(err).
					WithContext(fmt.Sprintf("%s ID: %s", m.entityType, req.ID))
			}
		}
	}

	// Save to repository
	if err := m.repo.Update(ctx, req.ID, updated); err != nil {
		return nil, model.NewInternalError(fmt.Sprintf("failed to update %s", m.entityType)).
			WithCause(err).
			WithContext(req.ID).
			WithSuggestions([]string{
				"Check if entity still exists",
				"Verify repository configuration",
				"Ensure data directory is writable",
			})
	}

	// Execute afterSave hooks
	var warnings []string
	if !req.SkipHooks {
		for _, hook := range m.afterSave {
			if err := hook(ctx, updated); err != nil {
				warnings = append(warnings, fmt.Sprintf("afterSave hook warning: %v", err))
			}
		}
	}

	return &UpdateEntityResponse[T]{
		Entity:   updated,
		Updated:  len(changes) > 0,
		Warnings: warnings,
		Changes:  changes,
		Metadata: map[string]interface{}{
			"entity_type": m.entityType,
			"updated_at":  updated.GetUpdatedAt(),
			"changes":     len(changes),
		},
	}, nil
}

// DeleteEntityRequest represents a request to delete an entity.
type DeleteEntityRequest struct {
	ID        string
	SkipHooks bool
	Force     bool  // Skip dependency checks
}

// DeleteEntityResponse represents the response from deleting an entity.
type DeleteEntityResponse[T BaseEntity] struct {
	Deleted   bool
	Entity    T  // The deleted entity (for audit/undo)
	Warnings  []string
	Metadata  map[string]interface{}
}

// Delete removes an entity with lifecycle management.
func (m *EntityManager[T]) Delete(ctx context.Context, req DeleteEntityRequest) (*DeleteEntityResponse[T], error) {
	// Get entity before deletion (for hooks and audit)
	entity, err := m.repo.Read(ctx, req.ID)
	if err != nil {
		return nil, model.NewNotFoundError(m.entityType).
			WithContext(req.ID).
			WithSuggestion(fmt.Sprintf("Use List() to see available %ss", m.entityType))
	}

	// Execute beforeDelete hooks
	if !req.SkipHooks {
		for _, hook := range m.beforeDelete {
			if err := hook(ctx, entity); err != nil {
				if !req.Force {
					return nil, model.NewValidationError("beforeDelete hook failed").
						WithCause(err).
						WithContext(fmt.Sprintf("%s ID: %s", m.entityType, req.ID)).
						WithSuggestion("Use Force=true to override, or resolve the dependency")
				}
			}
		}
	}

	// Delete from repository
	if err := m.repo.Delete(ctx, req.ID); err != nil {
		return nil, model.NewInternalError(fmt.Sprintf("failed to delete %s", m.entityType)).
			WithCause(err).
			WithContext(req.ID).
			WithSuggestions([]string{
				"Check if entity still exists",
				"Verify repository configuration",
				"Ensure data directory is writable",
			})
	}

	// Execute afterDelete hooks
	var warnings []string
	if !req.SkipHooks {
		for _, hook := range m.afterDelete {
			if err := hook(ctx, entity); err != nil {
				warnings = append(warnings, fmt.Sprintf("afterDelete hook warning: %v", err))
			}
		}
	}

	return &DeleteEntityResponse[T]{
		Deleted:  true,
		Entity:   entity,
		Warnings: warnings,
		Metadata: map[string]interface{}{
			"entity_type": m.entityType,
			"deleted_at":  time.Now(),
		},
	}, nil
}

// ListEntitiesRequest represents a request to list entities.
type ListEntitiesRequest struct {
	Filter    model.Filter
	SortBy    string
	SortOrder string  // "asc" or "desc"
	Limit     int
	Offset    int
}

// ListEntitiesResponse represents the response from listing entities.
type ListEntitiesResponse[T BaseEntity] struct {
	Entities []T
	Total    int
	Filtered int
	Metadata map[string]interface{}
}

// List retrieves entities with filtering, sorting, and pagination.
func (m *EntityManager[T]) List(ctx context.Context, req ListEntitiesRequest) (*ListEntitiesResponse[T], error) {
	// Get entities from repository
	entities, err := m.repo.List(ctx, req.Filter)
	if err != nil {
		return nil, model.NewInternalError(fmt.Sprintf("failed to list %ss", m.entityType)).
			WithCause(err).
			WithSuggestions([]string{
				"Check repository configuration",
				"Verify data directory exists and is readable",
			})
	}

	total := len(entities)

	// Apply sorting (simplified - in real implementation, use reflection)
	// TODO: Implement generic sorting based on req.SortBy and req.SortOrder

	// Apply pagination
	filtered := len(entities)
	if req.Limit > 0 {
		start := req.Offset
		end := start + req.Limit
		if start >= len(entities) {
			entities = []T{}
		} else {
			if end > len(entities) {
				end = len(entities)
			}
			entities = entities[start:end]
		}
	}

	return &ListEntitiesResponse[T]{
		Entities: entities,
		Total:    total,
		Filtered: filtered,
		Metadata: map[string]interface{}{
			"entity_type": m.entityType,
			"query_time":  time.Now(),
			"has_filter":  req.Filter != nil,
			"paginated":   req.Limit > 0,
		},
	}, nil
}

// Get retrieves a single entity by ID.
func (m *EntityManager[T]) Get(ctx context.Context, id string) (T, error) {
	entity, err := m.repo.Read(ctx, id)
	if err != nil {
		var zero T
		return zero, model.NewNotFoundError(m.entityType).
			WithContext(id).
			WithSuggestion(fmt.Sprintf("Use List() to see available %ss", m.entityType))
	}
	return entity, nil
}

// Exists checks if an entity exists by ID.
func (m *EntityManager[T]) Exists(ctx context.Context, id string) (bool, error) {
	exists, err := m.repo.Exists(ctx, id)
	if err != nil {
		return false, model.NewInternalError("failed to check entity existence").
			WithCause(err).
			WithContext(fmt.Sprintf("%s ID: %s", m.entityType, id))
	}
	return exists, nil
}

// Count returns the total number of entities matching the filter.
func (m *EntityManager[T]) Count(ctx context.Context, filter model.Filter) (int, error) {
	count, err := m.repo.Count(ctx, filter)
	if err != nil {
		return 0, model.NewInternalError(fmt.Sprintf("failed to count %ss", m.entityType)).
			WithCause(err).
			WithSuggestion("Check repository configuration")
	}
	return count, nil
}

// generateID generates a unique ID for the entity.
func (m *EntityManager[T]) generateID(entity T) (string, error) {
	// Simple implementation - in production, use UUID or more sophisticated ID generation
	timestamp := time.Now().Unix()
	entityType := m.entityType
	return fmt.Sprintf("%s-%d", entityType, timestamp), nil
}

// GetEntityType returns the entity type managed by this manager.
func (m *EntityManager[T]) GetEntityType() string {
	return m.entityType
}

// GetRepository returns the underlying repository (for advanced operations).
func (m *EntityManager[T]) GetRepository() model.Repository[T] {
	return m.repo
}