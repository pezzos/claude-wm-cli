// Package valueobjects contains domain value objects that encapsulate business rules and behaviors.
// Value objects are immutable and are compared by their values rather than identity.
package valueobjects

import (
	"fmt"
)

// Priority represents the importance level of work items.
// This is a domain value object that encapsulates business rules about priority levels.
type Priority string

const (
	// P0 represents critical priority - requires immediate attention
	P0 Priority = "P0"
	// P1 represents high priority - important, to be done soon
	P1 Priority = "P1"
	// P2 represents medium priority - normal priority work
	P2 Priority = "P2"
	// P3 represents low priority - can be deferred if needed
	P3 Priority = "P3"
)

// NewPriority creates a new Priority value object with validation.
func NewPriority(value string) (Priority, error) {
	p := Priority(value)
	if !p.IsValid() {
		return "", fmt.Errorf("invalid priority: %s", value)
	}
	return p, nil
}

// NewPriorityFromLegacy converts legacy priority strings to standardized Priority.
// This maintains backward compatibility with existing data formats.
func NewPriorityFromLegacy(legacy string) Priority {
	switch legacy {
	case "critical":
		return P0
	case "high":
		return P1
	case "medium":
		return P2
	case "low":
		return P3
	default:
		return P2 // Default to medium if unknown
	}
}

// String returns the string representation of Priority.
func (p Priority) String() string {
	return string(p)
}

// IsValid validates that the Priority value is one of the defined constants.
func (p Priority) IsValid() bool {
	switch p {
	case P0, P1, P2, P3:
		return true
	default:
		return false
	}
}

// Weight returns a numeric weight for priority comparison and sorting.
// Higher values indicate higher priority.
func (p Priority) Weight() int {
	switch p {
	case P0:
		return 4
	case P1:
		return 3
	case P2:
		return 2
	case P3:
		return 1
	default:
		return 0
	}
}

// IsCritical returns true if this is a critical priority (P0).
func (p Priority) IsCritical() bool {
	return p == P0
}

// IsHigh returns true if this is a high priority (P1).
func (p Priority) IsHigh() bool {
	return p == P1
}

// IsHigherThan returns true if this priority has higher weight than the other.
func (p Priority) IsHigherThan(other Priority) bool {
	return p.Weight() > other.Weight()
}

// IsLowerThan returns true if this priority has lower weight than the other.
func (p Priority) IsLowerThan(other Priority) bool {
	return p.Weight() < other.Weight()
}

// Equal checks if two priorities are equal.
func (p Priority) Equal(other Priority) bool {
	return p == other
}

// ToLegacyString returns the legacy string representation for backward compatibility.
func (p Priority) ToLegacyString() string {
	switch p {
	case P0:
		return "critical"
	case P1:
		return "high"
	case P2:
		return "medium"
	case P3:
		return "low"
	default:
		return "medium"
	}
}

// AllPriorities returns all valid priority values.
func AllPriorities() []Priority {
	return []Priority{P0, P1, P2, P3}
}