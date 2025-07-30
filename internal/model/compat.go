// Package model provides backward compatibility for the Clean Architecture migration.
// This file contains compatibility functions and aliases to maintain existing functionality.
package model

import (
	"claude-wm-cli/internal/domain/valueobjects"
)

// Backward compatibility functions for Status

// GetValidTransitions provides backward compatibility for status transitions.
func (s Status) GetValidTransitions() []Status {
	domainStatus := valueobjects.Status(s)
	domainTransitions := domainStatus.GetValidTransitions()
	
	transitions := make([]Status, len(domainTransitions))
	for i, t := range domainTransitions {
		transitions[i] = Status(t)
	}
	return transitions
}

// Backward compatibility functions for Epic validation

// ValidateEpic provides backward compatibility for epic validation.
func ValidateEpic(epic interface{}) error {
	// This is a placeholder for backward compatibility
	// In practice, the new domain entities have built-in validation
	return nil
}

// Additional compatibility helpers can be added here as needed