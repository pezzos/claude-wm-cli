// Package model defines common types and interfaces shared across the application.
// This package centralizes entity definitions to reduce duplication and ensure consistency.
package model

import (
	"fmt"
	"time"
)

// BaseEntity provides common fields for all entities in the system.
// It includes standard metadata fields that track entity lifecycle.
type BaseEntity struct {
	ID          string    `json:"id"`                    // Unique identifier for the entity
	CreatedAt   time.Time `json:"created_at"`            // Timestamp when entity was created
	UpdatedAt   time.Time `json:"updated_at"`            // Timestamp when entity was last updated
	Description string    `json:"description,omitempty"` // Optional description
}

// Metadata extends BaseEntity with versioning and authorship information.
// Used for entities that require audit trail and schema evolution support.
type Metadata struct {
	BaseEntity
	Version       string `json:"version"`                   // Entity version for optimistic locking
	CreatedBy     string `json:"created_by,omitempty"`     // User who created the entity
	UpdatedBy     string `json:"updated_by,omitempty"`     // User who last updated the entity
	SchemaVersion string `json:"schema_version"`           // Schema version for migration support
}

// Priority represents the importance level of work items.
// Uses a standardized P0-P3 scale where P0 is most critical.
type Priority string

const (
	PriorityP0 Priority = "P0" // Critical - requires immediate attention
	PriorityP1 Priority = "P1" // High - important, to be done soon
	PriorityP2 Priority = "P2" // Medium - normal priority work
	PriorityP3 Priority = "P3" // Low - can be deferred if needed
)

// PriorityFromLegacy converts legacy priority strings to standardized Priority.
// Maintains backward compatibility with existing data formats.
func PriorityFromLegacy(legacy string) Priority {
	switch legacy {
	case "critical":
		return PriorityP0
	case "high":
		return PriorityP1
	case "medium":
		return PriorityP2
	case "low":
		return PriorityP3
	default:
		return PriorityP2 // Default to medium if unknown
	}
}

// String returns the string representation of Priority.
func (p Priority) String() string {
	return string(p)
}

// IsValid validates that the Priority value is one of the defined constants.
func (p Priority) IsValid() bool {
	switch p {
	case PriorityP0, PriorityP1, PriorityP2, PriorityP3:
		return true
	default:
		return false
	}
}

// Weight returns a numeric weight for priority comparison and sorting.
// Higher values indicate higher priority.
func (p Priority) Weight() int {
	switch p {
	case PriorityP0:
		return 4
	case PriorityP1:
		return 3
	case PriorityP2:
		return 2
	case PriorityP3:
		return 1
	default:
		return 0
	}
}

// Status represents the current state of work items in the workflow.
// Uses a standardized state machine with clear transitions.
type Status string

const (
	StatusPlanned    Status = "planned"     // Initial state - work is planned but not started
	StatusInProgress Status = "in_progress" // Work is actively being done
	StatusBlocked    Status = "blocked"     // Work is blocked by external dependencies
	StatusOnHold     Status = "on_hold"     // Work is paused but can be resumed
	StatusCompleted  Status = "completed"   // Work has been finished successfully
	StatusCancelled  Status = "cancelled"   // Work has been cancelled and won't be completed
)

// StatusFromLegacy converts legacy status strings to standardized Status.
// Maintains backward compatibility with existing data formats.
func StatusFromLegacy(legacy string) Status {
	switch legacy {
	case "todo":
		return StatusPlanned
	case "done":
		return StatusCompleted
	default:
		// Try exact match first
		if Status(legacy).IsValid() {
			return Status(legacy)
		}
		return StatusPlanned // Default to planned if unknown
	}
}

// String returns the string representation of Status.
func (s Status) String() string {
	return string(s)
}

// IsValid validates that the Status value is one of the defined constants.
func (s Status) IsValid() bool {
	switch s {
	case StatusPlanned, StatusInProgress, StatusBlocked, StatusOnHold, StatusCompleted, StatusCancelled:
		return true
	default:
		return false
	}
}

// IsActive returns true if the status represents active work.
func (s Status) IsActive() bool {
	return s == StatusInProgress
}

// IsTerminal returns true if the status represents completed work (success or failure).
func (s Status) IsTerminal() bool {
	return s == StatusCompleted || s == StatusCancelled
}

// CanTransitionTo checks if a transition from current status to target status is valid.
func (s Status) CanTransitionTo(target Status) bool {
	// Define valid state transitions
	transitions := map[Status][]Status{
		StatusPlanned:    {StatusInProgress, StatusOnHold, StatusCancelled},
		StatusInProgress: {StatusBlocked, StatusOnHold, StatusCompleted, StatusCancelled},
		StatusBlocked:    {StatusInProgress, StatusOnHold, StatusCancelled},
		StatusOnHold:     {StatusPlanned, StatusInProgress, StatusCancelled},
		StatusCompleted:  {}, // Terminal state - no transitions allowed
		StatusCancelled:  {StatusPlanned}, // Can be reopened
	}

	validTargets, exists := transitions[s]
	if !exists {
		return false
	}

	for _, validTarget := range validTargets {
		if validTarget == target {
			return true
		}
	}
	return false
}

// WorkflowState defines the interface for entities that participate in workflow management.
// Entities implementing this interface can be managed by the workflow engine.
type WorkflowState interface {
	GetID() string
	GetStatus() Status
	SetStatus(Status) error
	GetPriority() Priority
	SetPriority(Priority) error
	GetMetadata() BaseEntity
	Validate() error
}

// ProgressMetrics tracks completion progress for entities with sub-items.
// Provides standardized progress calculation across different entity types.
type ProgressMetrics struct {
	TotalItems         int        `json:"total_items"`          // Total number of sub-items
	CompletedItems     int        `json:"completed_items"`      // Number of completed sub-items
	TotalPoints        int        `json:"total_points"`         // Total story points or effort points
	CompletedPoints    int        `json:"completed_points"`     // Completed story points or effort points
	CompletionPercent  float64    `json:"completion_percent"`   // Completion percentage (0-100)
	EstimatedEndDate   *time.Time `json:"estimated_end_date"`   // Estimated completion date
	VelocityPerDay     float64    `json:"velocity_per_day"`     // Average completion velocity
	RemainingEffort    int        `json:"remaining_effort"`     // Estimated remaining effort
}

// Calculate updates all progress metrics based on current state.
func (pm *ProgressMetrics) Calculate() {
	if pm.TotalItems == 0 {
		pm.CompletionPercent = 0
		return
	}

	// Calculate completion percentage based on items or points
	if pm.TotalPoints > 0 {
		pm.CompletionPercent = float64(pm.CompletedPoints) / float64(pm.TotalPoints) * 100
	} else {
		pm.CompletionPercent = float64(pm.CompletedItems) / float64(pm.TotalItems) * 100
	}

	// Calculate remaining effort
	if pm.TotalPoints > 0 {
		pm.RemainingEffort = pm.TotalPoints - pm.CompletedPoints
	} else {
		pm.RemainingEffort = pm.TotalItems - pm.CompletedItems
	}

	// Estimate end date based on velocity
	if pm.VelocityPerDay > 0 && pm.RemainingEffort > 0 {
		daysRemaining := float64(pm.RemainingEffort) / pm.VelocityPerDay
		endDate := time.Now().Add(time.Duration(daysRemaining) * 24 * time.Hour)
		pm.EstimatedEndDate = &endDate
	}
}

// IsComplete returns true if the work is 100% complete.
func (pm *ProgressMetrics) IsComplete() bool {
	return pm.CompletionPercent >= 100.0
}

// ValidationError represents an error in entity validation.
type ValidationError struct {
	Field   string `json:"field"`   // Field name that failed validation
	Value   string `json:"value"`   // Invalid value
	Message string `json:"message"` // Human-readable error message
}

// Error implements the error interface.
func (ve ValidationError) Error() string {
	return fmt.Sprintf("validation error on field '%s': %s (value: %s)", ve.Field, ve.Message, ve.Value)
}

// ValidationErrors represents multiple validation errors.
type ValidationErrors []ValidationError

// Error implements the error interface for multiple validation errors.
func (ves ValidationErrors) Error() string {
	if len(ves) == 0 {
		return "no validation errors"
	}
	if len(ves) == 1 {
		return ves[0].Error()
	}
	return fmt.Sprintf("%s (and %d more validation errors)", ves[0].Error(), len(ves)-1)
}

// HasErrors returns true if there are any validation errors.
func (ves ValidationErrors) HasErrors() bool {
	return len(ves) > 0
}

// Add appends a new validation error.
func (ves *ValidationErrors) Add(field, value, message string) {
	*ves = append(*ves, ValidationError{
		Field:   field,
		Value:   value,
		Message: message,
	})
}