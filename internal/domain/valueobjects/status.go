// Package valueobjects contains domain value objects that encapsulate business rules and behaviors.
package valueobjects

import (
	"fmt"
)

// Status represents the current state of work items in the workflow.
// This is a domain value object that encapsulates the state machine logic.
type Status string

const (
	// Planned represents initial state - work is planned but not started
	Planned Status = "planned"
	// InProgress represents work that is actively being done
	InProgress Status = "in_progress"
	// Blocked represents work that is blocked by external dependencies
	Blocked Status = "blocked"
	// OnHold represents work that is paused but can be resumed
	OnHold Status = "on_hold"
	// Completed represents work that has been finished successfully
	Completed Status = "completed"
	// Cancelled represents work that has been cancelled and won't be completed
	Cancelled Status = "cancelled"
)

// NewStatus creates a new Status value object with validation.
func NewStatus(value string) (Status, error) {
	s := Status(value)
	if !s.IsValid() {
		return "", fmt.Errorf("invalid status: %s", value)
	}
	return s, nil
}

// NewStatusFromLegacy converts legacy status strings to standardized Status.
// This maintains backward compatibility with existing data formats.
func NewStatusFromLegacy(legacy string) Status {
	switch legacy {
	case "todo":
		return Planned
	case "done":
		return Completed
	default:
		// Try exact match first
		if Status(legacy).IsValid() {
			return Status(legacy)
		}
		return Planned // Default to planned if unknown
	}
}

// String returns the string representation of Status.
func (s Status) String() string {
	return string(s)
}

// IsValid validates that the Status value is one of the defined constants.
func (s Status) IsValid() bool {
	switch s {
	case Planned, InProgress, Blocked, OnHold, Completed, Cancelled:
		return true
	default:
		return false
	}
}

// IsActive returns true if the status represents active work.
func (s Status) IsActive() bool {
	return s == InProgress
}

// IsTerminal returns true if the status represents completed work (success or failure).
func (s Status) IsTerminal() bool {
	return s == Completed || s == Cancelled
}

// IsBlocked returns true if the status represents blocked work.
func (s Status) IsBlocked() bool {
	return s == Blocked || s == OnHold
}

// IsSuccessful returns true if the work was completed successfully.
func (s Status) IsSuccessful() bool {
	return s == Completed
}

// IsCancelled returns true if the work was cancelled.
func (s Status) IsCancelled() bool {
	return s == Cancelled
}

// CanTransitionTo checks if a transition from current status to target status is valid.
// This encapsulates the state machine logic for workflow transitions.
func (s Status) CanTransitionTo(target Status) bool {
	// Define valid state transitions based on business rules
	transitions := map[Status][]Status{
		Planned:    {InProgress, OnHold, Cancelled},
		InProgress: {Blocked, OnHold, Completed, Cancelled},
		Blocked:    {InProgress, OnHold, Cancelled},
		OnHold:     {Planned, InProgress, Cancelled},
		Completed:  {}, // Terminal state - no transitions allowed
		Cancelled:  {Planned}, // Can be reopened
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

// GetValidTransitions returns all valid target statuses from the current status.
func (s Status) GetValidTransitions() []Status {
	transitions := map[Status][]Status{
		Planned:    {InProgress, OnHold, Cancelled},
		InProgress: {Blocked, OnHold, Completed, Cancelled},
		Blocked:    {InProgress, OnHold, Cancelled},
		OnHold:     {Planned, InProgress, Cancelled},
		Completed:  {},
		Cancelled:  {Planned},
	}

	if validTargets, exists := transitions[s]; exists {
		return validTargets
	}
	return []Status{}
}

// Equal checks if two statuses are equal.
func (s Status) Equal(other Status) bool {
	return s == other
}

// ToLegacyString returns the legacy string representation for backward compatibility.
func (s Status) ToLegacyString() string {
	switch s {
	case Planned:
		return "todo"
	case Completed:
		return "done"
	default:
		return string(s)
	}
}

// AllStatuses returns all valid status values.
func AllStatuses() []Status {
	return []Status{Planned, InProgress, Blocked, OnHold, Completed, Cancelled}
}

// WorkflowPhases returns statuses grouped by workflow phase.
func WorkflowPhases() map[string][]Status {
	return map[string][]Status{
		"planning":   {Planned},
		"active":     {InProgress},
		"blocked":    {Blocked, OnHold},
		"completed":  {Completed, Cancelled},
	}
}