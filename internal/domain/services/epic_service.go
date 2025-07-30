// Package services contains domain services that implement business logic.
// Domain services encapsulate business rules that don't naturally belong to entities.
package services

import (
	"context"
	"fmt"
	"time"

	"claude-wm-cli/internal/domain/entities"
	"claude-wm-cli/internal/domain/repositories"
	"claude-wm-cli/internal/domain/valueobjects"
)

// EpicDomainService implements business logic for epic management.
// This is a domain service that encapsulates business rules that involve multiple entities
// or don't naturally belong to a single entity.
type EpicDomainService struct {
	epicRepo repositories.EpicRepository
}

// NewEpicDomainService creates a new instance of EpicDomainService.
func NewEpicDomainService(epicRepo repositories.EpicRepository) *EpicDomainService {
	return &EpicDomainService{
		epicRepo: epicRepo,
	}
}

// ValidateEpicCreation validates that an epic can be created with the given parameters.
func (s *EpicDomainService) ValidateEpicCreation(ctx context.Context, id, title, description string, priority valueobjects.Priority) error {
	// Check if epic with same ID already exists
	exists, err := s.epicRepo.Exists(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check epic existence: %w", err)
	}
	if exists {
		return fmt.Errorf("epic with ID %s already exists", id)
	}

	// Validate business rules for epic creation
	if title == "" {
		return fmt.Errorf("epic title is required")
	}
	if description == "" {
		return fmt.Errorf("epic description is required")
	}
	if !priority.IsValid() {
		return fmt.Errorf("invalid priority: %s", priority)
	}

	return nil
}

// CanTransitionEpicStatus checks if an epic can transition to a new status.
// This implements business rules around status transitions.
func (s *EpicDomainService) CanTransitionEpicStatus(ctx context.Context, epicID string, newStatus valueobjects.Status) error {
	epic, err := s.epicRepo.GetByID(ctx, epicID)
	if err != nil {
		return fmt.Errorf("failed to get epic: %w", err)
	}

	// Check basic status transition rules
	if !epic.Status().CanTransitionTo(newStatus) {
		return fmt.Errorf("cannot transition from %s to %s", epic.Status(), newStatus)
	}

	// Additional business rules for specific transitions
	switch newStatus {
	case valueobjects.InProgress:
		return s.validateStartEpic(ctx, epic)
	case valueobjects.Completed:
		return s.validateCompleteEpic(ctx, epic)
	case valueobjects.Blocked:
		return s.validateBlockEpic(ctx, epic)
	}

	return nil
}

// validateStartEpic validates that an epic can be started.
func (s *EpicDomainService) validateStartEpic(ctx context.Context, epic *entities.Epic) error {
	// Check if dependencies are resolved
	for _, depID := range epic.Dependencies() {
		depEpic, err := s.epicRepo.GetByID(ctx, depID)
		if err != nil {
			return fmt.Errorf("failed to check dependency %s: %w", depID, err)
		}
		if !depEpic.IsCompleted() {
			return fmt.Errorf("epic cannot be started: dependency %s is not completed", depID)
		}
	}

	// Check if epic has at least one user story
	if len(epic.UserStories()) == 0 {
		return fmt.Errorf("epic cannot be started without user stories")
	}

	return nil
}

// validateCompleteEpic validates that an epic can be completed.
func (s *EpicDomainService) validateCompleteEpic(ctx context.Context, epic *entities.Epic) error {
	// Check if all user stories are completed
	for _, story := range epic.UserStories() {
		if story.Status != valueobjects.Completed {
			return fmt.Errorf("epic cannot be completed: user story %s is not completed", story.ID)
		}
	}

	// Check if progress is 100%
	if epic.Progress().CompletionPercentage < 100.0 {
		return fmt.Errorf("epic cannot be completed: progress is only %.1f%%", epic.Progress().CompletionPercentage)
	}

	return nil
}

// validateBlockEpic validates that an epic can be blocked.
func (s *EpicDomainService) validateBlockEpic(ctx context.Context, epic *entities.Epic) error {
	// Epic can be blocked if it's currently in progress
	if epic.Status() != valueobjects.InProgress {
		return fmt.Errorf("only in-progress epics can be blocked")
	}
	return nil
}

// CalculateDependencyImpact calculates the impact of completing or cancelling an epic.
func (s *EpicDomainService) CalculateDependencyImpact(ctx context.Context, epicID string) (*DependencyImpact, error) {
	// Get epics that depend on this epic
	dependents, err := s.epicRepo.GetDependents(ctx, epicID)
	if err != nil {
		return nil, fmt.Errorf("failed to get dependents: %w", err)
	}

	impact := &DependencyImpact{
		EpicID:           epicID,
		DirectDependents: make([]string, len(dependents)),
		TotalImpact:      len(dependents),
	}

	for i, dep := range dependents {
		impact.DirectDependents[i] = dep.ID()
	}

	// Calculate cascade impact (epics that depend on the dependents)
	cascadeImpact := make(map[string]bool)
	for _, dep := range dependents {
		cascade, err := s.epicRepo.GetDependents(ctx, dep.ID())
		if err != nil {
			continue // Skip errors in cascade calculation
		}
		for _, cascadeDep := range cascade {
			cascadeImpact[cascadeDep.ID()] = true
		}
	}

	impact.CascadeDependents = make([]string, 0, len(cascadeImpact))
	for id := range cascadeImpact {
		impact.CascadeDependents = append(impact.CascadeDependents, id)
	}
	impact.TotalImpact += len(cascadeImpact)

	return impact, nil
}

// ValidateEpicDeletion validates that an epic can be safely deleted.
func (s *EpicDomainService) ValidateEpicDeletion(ctx context.Context, epicID string) error {
	// Check if other epics depend on this epic
	dependents, err := s.epicRepo.GetDependents(ctx, epicID)
	if err != nil {
		return fmt.Errorf("failed to check dependents: %w", err)
	}

	if len(dependents) > 0 {
		dependentIDs := make([]string, len(dependents))
		for i, dep := range dependents {
			dependentIDs[i] = dep.ID()
		}
		return fmt.Errorf("epic cannot be deleted: %d epics depend on it: %v", len(dependents), dependentIDs)
	}

	return nil
}

// SuggestEpicPriority suggests an appropriate priority for an epic based on business rules.
func (s *EpicDomainService) SuggestEpicPriority(ctx context.Context, tags []string, dependencies []string) valueobjects.Priority {
	// Business rules for priority suggestion
	for _, tag := range tags {
		switch tag {
		case "critical", "security", "bug", "hotfix":
			return valueobjects.P0
		case "important", "feature", "customer-request":
			return valueobjects.P1
		}
	}

	// If epic has many dependencies, it might be foundational work
	if len(dependencies) == 0 {
		return valueobjects.P1 // High priority for independent work
	}

	// Default to medium priority
	return valueobjects.P2
}

// CalculateEstimatedEndDate calculates when an epic is likely to be completed.
func (s *EpicDomainService) CalculateEstimatedEndDate(epic *entities.Epic, velocity float64) *time.Time {
	progress := epic.Progress()
	if progress.CompletionPercentage >= 100.0 {
		return nil // Already completed
	}

	if velocity <= 0 {
		return nil // Cannot calculate without velocity
	}

	remainingPoints := progress.TotalStoryPoints - progress.CompletedStoryPoints
	if remainingPoints <= 0 {
		return nil // No remaining work
	}

	daysRemaining := float64(remainingPoints) / velocity
	endDate := time.Now().Add(time.Duration(daysRemaining) * 24 * time.Hour)
	return &endDate
}

// ValidateEpicDependencies validates that epic dependencies are valid and don't create cycles.
func (s *EpicDomainService) ValidateEpicDependencies(ctx context.Context, epicID string, dependencies []string) error {
	// Check if dependencies exist
	for _, depID := range dependencies {
		exists, err := s.epicRepo.Exists(ctx, depID)
		if err != nil {
			return fmt.Errorf("failed to check dependency %s: %w", depID, err)
		}
		if !exists {
			return fmt.Errorf("dependency epic %s does not exist", depID)
		}
	}

	// Check for self-dependency
	for _, depID := range dependencies {
		if depID == epicID {
			return fmt.Errorf("epic cannot depend on itself")
		}
	}

	// Check for circular dependencies
	return s.validateNoCycles(ctx, epicID, dependencies)
}

// validateNoCycles checks for circular dependencies using depth-first search.
func (s *EpicDomainService) validateNoCycles(ctx context.Context, epicID string, newDependencies []string) error {
	visited := make(map[string]bool)
	recursionStack := make(map[string]bool)

	// Check each new dependency for cycles
	for _, depID := range newDependencies {
		if s.hasCycle(ctx, depID, epicID, visited, recursionStack) {
			return fmt.Errorf("adding dependency %s would create a circular dependency", depID)
		}
	}

	return nil
}

// hasCycle performs DFS to detect cycles in the dependency graph.
func (s *EpicDomainService) hasCycle(ctx context.Context, currentID, targetID string, visited, recursionStack map[string]bool) bool {
	if currentID == targetID {
		return true
	}

	if visited[currentID] {
		return recursionStack[currentID]
	}

	visited[currentID] = true
	recursionStack[currentID] = true

	// Get dependencies of current epic
	epic, err := s.epicRepo.GetByID(ctx, currentID)
	if err != nil {
		return false // Skip on error
	}

	for _, depID := range epic.Dependencies() {
		if s.hasCycle(ctx, depID, targetID, visited, recursionStack) {
			return true
		}
	}

	recursionStack[currentID] = false
	return false
}

// DependencyImpact represents the impact of changing an epic on its dependents.
type DependencyImpact struct {
	EpicID             string   `json:"epic_id"`
	DirectDependents   []string `json:"direct_dependents"`
	CascadeDependents  []string `json:"cascade_dependents"`
	TotalImpact        int      `json:"total_impact"`
}