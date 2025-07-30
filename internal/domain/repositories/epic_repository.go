// Package repositories contains repository interfaces that define data access contracts.
// These interfaces are part of the domain layer and represent the application's data needs.
// Infrastructure layer implementations will provide concrete implementations.
package repositories

import (
	"context"

	"claude-wm-cli/internal/domain/entities"
	"claude-wm-cli/internal/domain/valueobjects"
)

// EpicRepository defines the contract for epic data persistence.
// This interface follows the Repository pattern and is implemented by infrastructure layer.
type EpicRepository interface {
	// Create saves a new epic to the repository.
	Create(ctx context.Context, epic *entities.Epic) error

	// GetByID retrieves an epic by its unique identifier.
	GetByID(ctx context.Context, id string) (*entities.Epic, error)

	// Update saves changes to an existing epic.
	Update(ctx context.Context, epic *entities.Epic) error

	// Delete removes an epic from the repository.
	Delete(ctx context.Context, id string) error

	// List retrieves epics based on the provided filter criteria.
	List(ctx context.Context, filter EpicFilter) ([]*entities.Epic, error)

	// Exists checks if an epic with the given ID exists.
	Exists(ctx context.Context, id string) (bool, error)

	// Count returns the total number of epics matching the filter.
	Count(ctx context.Context, filter EpicFilter) (int, error)

	// GetByStatus retrieves all epics with the specified status.
	GetByStatus(ctx context.Context, status valueobjects.Status) ([]*entities.Epic, error)

	// GetByPriority retrieves all epics with the specified priority.
	GetByPriority(ctx context.Context, priority valueobjects.Priority) ([]*entities.Epic, error)

	// GetByTag retrieves all epics that have the specified tag.
	GetByTag(ctx context.Context, tag string) ([]*entities.Epic, error)

	// Search performs a text search across epic titles, descriptions, and tags.
	Search(ctx context.Context, query string) ([]*entities.Epic, error)

	// GetDependents returns epics that depend on the specified epic.
	GetDependents(ctx context.Context, epicID string) ([]*entities.Epic, error)

	// GetBlocked returns epics that are blocked by unresolved dependencies.
	GetBlocked(ctx context.Context) ([]*entities.Epic, error)

	// GetActive returns all epics that are currently in progress.
	GetActive(ctx context.Context) ([]*entities.Epic, error)

	// GetOverdue returns epics that are past their end date and not completed.
	GetOverdue(ctx context.Context) ([]*entities.Epic, error)
}

// EpicFilter defines criteria for filtering epics in queries.
type EpicFilter struct {
	// Status filters epics by their current status
	Status *valueobjects.Status

	// Priority filters epics by their priority level
	Priority *valueobjects.Priority

	// Tags filters epics that contain all specified tags
	Tags []string

	// HasDependencies filters epics based on whether they have dependencies
	HasDependencies *bool

	// IsBlocked filters epics that are blocked by unresolved dependencies
	IsBlocked *bool

	// CreatedAfter filters epics created after the specified time
	CreatedAfter *int64

	// CreatedBefore filters epics created before the specified time
	CreatedBefore *int64

	// UpdatedAfter filters epics updated after the specified time
	UpdatedAfter *int64

	// UpdatedBefore filters epics updated before the specified time
	UpdatedBefore *int64

	// Limit limits the number of results returned (0 = no limit)
	Limit int

	// Offset specifies the number of results to skip for pagination
	Offset int

	// SortBy specifies the field to sort by ("created_at", "updated_at", "title", "priority")
	SortBy string

	// SortOrder specifies the sort direction ("desc", "asc")
	SortOrder string
}

// EpicAggregate provides aggregate operations on epic collections.
type EpicAggregate interface {
	// GetStatistics returns statistical information about epics.
	GetStatistics(ctx context.Context, filter EpicFilter) (*EpicStatistics, error)

	// GetProgressSummary returns a summary of progress across all epics.
	GetProgressSummary(ctx context.Context, filter EpicFilter) (*ProgressSummary, error)

	// GetDependencyGraph returns the dependency relationships between epics.
	GetDependencyGraph(ctx context.Context) (*DependencyGraph, error)

	// GetWorkloadDistribution returns workload distribution by priority and status.
	GetWorkloadDistribution(ctx context.Context) (*WorkloadDistribution, error)
}

// EpicStatistics contains statistical information about epic collections.
type EpicStatistics struct {
	TotalEpics           int                          `json:"total_epics"`
	EpicsByStatus        map[valueobjects.Status]int  `json:"epics_by_status"`
	EpicsByPriority      map[valueobjects.Priority]int `json:"epics_by_priority"`
	AverageStoryPoints   float64                      `json:"average_story_points"`
	TotalStoryPoints     int                          `json:"total_story_points"`
	CompletedStoryPoints int                          `json:"completed_story_points"`
	OverallProgress      float64                      `json:"overall_progress"`
	BlockedEpics         int                          `json:"blocked_epics"`
	OverdueEpics         int                          `json:"overdue_epics"`
}

// ProgressSummary contains progress information across epic collections.
type ProgressSummary struct {
	TotalStoryPoints     int     `json:"total_story_points"`
	CompletedStoryPoints int     `json:"completed_story_points"`
	TotalStories         int     `json:"total_stories"`
	CompletedStories     int     `json:"completed_stories"`
	OverallProgress      float64 `json:"overall_progress"`
	EstimatedHours       int     `json:"estimated_hours"`
	ActualHours          int     `json:"actual_hours"`
	Velocity             float64 `json:"velocity"`
}

// DependencyGraph represents the dependency relationships between epics.
type DependencyGraph struct {
	Nodes []DependencyNode `json:"nodes"`
	Edges []DependencyEdge `json:"edges"`
}

// DependencyNode represents an epic in the dependency graph.
type DependencyNode struct {
	EpicID   string                `json:"epic_id"`
	Title    string                `json:"title"`
	Status   valueobjects.Status   `json:"status"`
	Priority valueobjects.Priority `json:"priority"`
}

// DependencyEdge represents a dependency relationship between epics.
type DependencyEdge struct {
	FromEpicID string `json:"from_epic_id"`
	ToEpicID   string `json:"to_epic_id"`
	Type       string `json:"type"` // "blocks", "relates_to", etc.
}

// WorkloadDistribution contains workload distribution information.
type WorkloadDistribution struct {
	ByPriority map[valueobjects.Priority]WorkloadMetrics `json:"by_priority"`
	ByStatus   map[valueobjects.Status]WorkloadMetrics   `json:"by_status"`
	Total      WorkloadMetrics                           `json:"total"`
}

// WorkloadMetrics contains metrics about workload.
type WorkloadMetrics struct {
	EpicCount            int     `json:"epic_count"`
	TotalStoryPoints     int     `json:"total_story_points"`
	CompletedStoryPoints int     `json:"completed_story_points"`
	EstimatedHours       int     `json:"estimated_hours"`
	ActualHours          int     `json:"actual_hours"`
	ProgressPercentage   float64 `json:"progress_percentage"`
}