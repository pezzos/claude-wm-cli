// Package entity provides common interfaces and utilities for entity management.
package entity

import (
	"context"
	"time"

	"claude-wm-cli/internal/model"
)

// EntityService defines the interface for entity services that provide business logic.
// This interface abstracts the common operations that all entity services should support.
type EntityService[T BaseEntity] interface {
	// Core CRUD operations
	Create(ctx context.Context, req CreateEntityRequest[T]) (*CreateEntityResponse[T], error)
	Update(ctx context.Context, req UpdateEntityRequest[T]) (*UpdateEntityResponse[T], error)
	Delete(ctx context.Context, req DeleteEntityRequest) (*DeleteEntityResponse[T], error)
	Get(ctx context.Context, id string) (T, error)
	List(ctx context.Context, req ListEntitiesRequest) (*ListEntitiesResponse[T], error)
	
	// Utility operations
	Exists(ctx context.Context, id string) (bool, error)
	Count(ctx context.Context, filter model.Filter) (int, error)
	
	// Entity type information
	GetEntityType() string
}

// WorkflowEntity defines entities that participate in workflow state transitions.
type WorkflowEntity interface {
	BaseEntity
	CanTransitionTo(newStatus model.Status) bool
	GetValidTransitions() []model.Status
	GetWorkflowContext() map[string]interface{}
}

// ProgressTrackingEntity defines entities that track completion progress.
type ProgressTrackingEntity interface {
	BaseEntity
	GetCompletionPercentage() float64
	GetProgressMetrics() map[string]interface{}
	UpdateProgress() error
}

// HierarchicalEntity defines entities that have parent-child relationships.
type HierarchicalEntity interface {
	BaseEntity
	GetParentID() string
	SetParentID(string)
	GetChildIDs() []string
	AddChildID(string)
	RemoveChildID(string)
}

// TaggableEntity defines entities that support tagging.
type TaggableEntity interface {
	BaseEntity
	GetTags() []string
	SetTags([]string)
	AddTag(string)
	RemoveTag(string)
	HasTag(string) bool
}

// TimestampedEntity provides detailed timestamp tracking.
type TimestampedEntity interface {
	BaseEntity
	GetCompletedAt() *time.Time
	SetCompletedAt(*time.Time)
	GetStartedAt() *time.Time
	SetStartedAt(*time.Time)
	GetDueDate() *time.Time
	SetDueDate(*time.Time)
}

// VersionedEntity defines entities that support versioning.
type VersionedEntity interface {
	BaseEntity
	GetVersion() int
	SetVersion(int)
	IncrementVersion()
}

// AuditableEntity defines entities that track changes for audit purposes.
type AuditableEntity interface {
	BaseEntity
	GetCreatedBy() string
	SetCreatedBy(string)
	GetUpdatedBy() string
	SetUpdatedBy(string)
	GetChangeHistory() []ChangeRecord
	AddChangeRecord(ChangeRecord)
}

// ChangeRecord represents a single change to an entity.
type ChangeRecord struct {
	Timestamp time.Time              `json:"timestamp"`
	User      string                 `json:"user"`
	Action    string                 `json:"action"` // create, update, delete, etc.
	Changes   map[string]interface{} `json:"changes"`
	Reason    string                 `json:"reason,omitempty"`
}

// DependentEntity defines entities that have dependencies on other entities.
type DependentEntity interface {
	BaseEntity
	GetDependencies() []string
	SetDependencies([]string)
	AddDependency(string)
	RemoveDependency(string)
	HasDependency(string) bool
	GetBlockedBy() []string
	GetBlocking() []string
}

// SearchableEntity defines entities that support full-text search.
type SearchableEntity interface {
	BaseEntity
	GetSearchableText() string
	GetSearchKeywords() []string
	GetSearchMetadata() map[string]interface{}
}

// EntityFactory defines a factory for creating entities of a specific type.
type EntityFactory[T BaseEntity] interface {
	CreateEmpty() T
	CreateFromTemplate(templateID string) (T, error)
	CreateFromData(data map[string]interface{}) (T, error)
	GetEntityType() string
	GetDefaultValues() map[string]interface{}
}

// EntityValidator provides validation services for entities.
type EntityValidator[T BaseEntity] interface {
	Validate(entity T) error
	ValidateField(entity T, fieldName string) error
	ValidateTransition(entity T, newStatus model.Status) error
	GetValidationRules() map[string]interface{}
}

// EntityTransformer provides transformation services between entity types.
type EntityTransformer[From BaseEntity, To BaseEntity] interface {
	Transform(from From) (To, error)
	CanTransform(from From) bool
	GetTransformationRules() map[string]interface{}
}

// EntityRepository is an alias for the model.Repository interface
// with additional documentation for entity-specific usage.
type EntityRepository[T BaseEntity] interface {
	model.Repository[T]
}

// EntityNotificationService handles notifications for entity lifecycle events.
type EntityNotificationService interface {
	OnEntityCreated(ctx context.Context, entityType string, entity interface{}) error
	OnEntityUpdated(ctx context.Context, entityType string, entity interface{}, changes []string) error
	OnEntityDeleted(ctx context.Context, entityType string, entityID string) error
	OnEntityStatusChanged(ctx context.Context, entityType string, entityID string, oldStatus, newStatus model.Status) error
}

// EntityCacheService provides caching services for entities.
type EntityCacheService[T BaseEntity] interface {
	Get(ctx context.Context, id string) (T, bool)
	Set(ctx context.Context, id string, entity T, ttl time.Duration) error
	Delete(ctx context.Context, id string) error
	Clear(ctx context.Context) error
	GetStats() map[string]interface{}
}

// EntityMetricsService provides metrics and analytics for entities.
type EntityMetricsService interface {
	RecordEntityCreated(entityType string)
	RecordEntityUpdated(entityType string)
	RecordEntityDeleted(entityType string)
	RecordEntityStatusTransition(entityType string, fromStatus, toStatus model.Status)
	GetEntityMetrics(entityType string) map[string]interface{}
	GetOverallMetrics() map[string]interface{}
}

// EntityIndexService provides indexing and search capabilities.
type EntityIndexService[T BaseEntity] interface {
	Index(ctx context.Context, entity T) error
	Remove(ctx context.Context, id string) error
	Search(ctx context.Context, query string) ([]T, error)
	Reindex(ctx context.Context) error
	GetIndexStats() map[string]interface{}
}

// EntityExportService provides export capabilities for entities.
type EntityExportService[T BaseEntity] interface {
	ExportToJSON(ctx context.Context, entities []T) ([]byte, error)
	ExportToCSV(ctx context.Context, entities []T) ([]byte, error)
	ExportToMarkdown(ctx context.Context, entities []T) ([]byte, error)
	ImportFromJSON(ctx context.Context, data []byte) ([]T, error)
	GetSupportedFormats() []string
}

// Common filter implementations for entities
type StatusFilter struct {
	Status model.Status
}

func (f *StatusFilter) Apply(entity interface{}) bool {
	if be, ok := entity.(BaseEntity); ok {
		return be.GetStatus() == f.Status
	}
	return false
}

type PriorityFilter struct {
	Priority model.Priority
}

func (f *PriorityFilter) Apply(entity interface{}) bool {
	if be, ok := entity.(BaseEntity); ok {
		return be.GetPriority() == f.Priority
	}
	return false
}

type DateRangeFilter struct {
	After  *time.Time
	Before *time.Time
}

func (f *DateRangeFilter) Apply(entity interface{}) bool {
	if be, ok := entity.(BaseEntity); ok {
		createdAt := be.GetCreatedAt()
		if f.After != nil && createdAt.Before(*f.After) {
			return false
		}
		if f.Before != nil && createdAt.After(*f.Before) {
			return false
		}
		return true
	}
	return false
}

// MultiFilter allows combining multiple filters with AND logic.
type MultiFilter struct {
	Filters []model.Filter
}

func (f *MultiFilter) Apply(entity interface{}) bool {
	for _, filter := range f.Filters {
		if !filter.Apply(entity) {
			return false
		}
	}
	return true
}

// OrFilter allows combining multiple filters with OR logic.
type OrFilter struct {
	Filters []model.Filter
}

func (f *OrFilter) Apply(entity interface{}) bool {
	for _, filter := range f.Filters {
		if filter.Apply(entity) {
			return true
		}
	}
	return false
}