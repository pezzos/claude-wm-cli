// Package application provides application services that encapsulate business logic.
// These services act as orchestrators between the CLI interface and domain entities.
package application

import (
	"context"
	"fmt"
	"strings"
	"time"

	"claude-wm-cli/internal/entity"
	"claude-wm-cli/internal/epic"
	"claude-wm-cli/internal/model"
	"claude-wm-cli/internal/persistence"
)

// EpicApplicationService provides high-level operations for Epic management.
// This service encapsulates all business logic, leaving CLI commands simple.
type EpicApplicationService struct {
	entityManager    *entity.EntityManager[*entity.EpicEntity]
	notificationSvc  entity.EntityNotificationService
	metricsSvc       entity.EntityMetricsService
	cacheSvc         entity.EntityCacheService[*entity.EpicEntity]
	factory          *entity.EpicEntityFactory
	validator        *entity.EpicEntityValidator
}

// EpicApplicationServiceConfig configures the Epic application service.
type EpicApplicationServiceConfig struct {
	DataDirectory   string
	EnableCache     bool
	EnableMetrics   bool
	EnableNotifications bool
}

// NewEpicApplicationService creates a new Epic application service.
func NewEpicApplicationService(config EpicApplicationServiceConfig) (*EpicApplicationService, error) {
	// Create repository
	factory := &entity.EpicEntityFactory{}
	validator := &entity.EpicEntityValidator{}
	
	// Create a custom validator function that works with EpicEntity
	validatorFunc := func(e *entity.EpicEntity) error {
		return validator.Validate(e)
	}
	
	repo := persistence.NewJSONRepository[*entity.EpicEntity](
		fmt.Sprintf("%s/epics.json", config.DataDirectory),
		validatorFunc,
		persistence.DefaultRepositoryOptions(),
	)
	
	// Create entity manager with hooks
	managerOptions := entity.ManagerOptions[*entity.EpicEntity]{
		EntityType: "epic",
		Validator:  validatorFunc,
		BeforeSave: []entity.HookFunc[*entity.EpicEntity]{
			validateEpicBusinessRules,
			calculateEpicProgress,
			updateEpicTimestamps,
		},
		AfterSave: []entity.HookFunc[*entity.EpicEntity]{
			notifyEpicChanges,
			updateEpicMetrics,
		},
		BeforeDelete: []entity.HookFunc[*entity.EpicEntity]{
			checkEpicDependencies,
		},
		AfterDelete: []entity.HookFunc[*entity.EpicEntity]{
			cleanupEpicReferences,
		},
	}
	
	entityManager := entity.NewEntityManager(repo, managerOptions)
	
	return &EpicApplicationService{
		entityManager: entityManager,
		factory:       factory,
		validator:     validator,
		// notificationSvc, metricsSvc, cacheSvc would be injected in real implementation
	}, nil
}

// CreateEpicWorkflowRequest represents a request to create an epic with full workflow.
type CreateEpicWorkflowRequest struct {
	Title        string            `json:"title"`
	Description  string            `json:"description"`
	Priority     string            `json:"priority"`
	Tags         []string          `json:"tags"`
	Dependencies []string          `json:"dependencies"`
	Duration     time.Duration     `json:"duration"`
	Template     string            `json:"template,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	SkipValidation bool             `json:"skip_validation,omitempty"`
}

// CreateEpicWorkflowResponse represents the response from creating an epic.
type CreateEpicWorkflowResponse struct {
	Epic        *entity.EpicEntity        `json:"epic"`
	Warnings    []string                  `json:"warnings,omitempty"`
	Suggestions []string                  `json:"suggestions,omitempty"`
	NextActions []string                  `json:"next_actions,omitempty"`
	Metadata    map[string]interface{}    `json:"metadata,omitempty"`
}

// CreateEpicWorkflow creates a new epic with full business logic and workflow integration.
func (s *EpicApplicationService) CreateEpicWorkflow(ctx context.Context, req CreateEpicWorkflowRequest) (*CreateEpicWorkflowResponse, error) {
	// Create entity from template or empty
	var epicEntity *entity.EpicEntity
	var err error
	
	if req.Template != "" {
		epicEntity, err = s.factory.CreateFromTemplate(req.Template)
		if err != nil {
			return nil, model.NewValidationError("failed to create epic from template").
				WithCause(err).
				WithContext(req.Template).
				WithSuggestions([]string{"feature", "bugfix"})
		}
	} else {
		epicEntity = s.factory.CreateEmpty()
	}
	
	// Apply request data
	epicEntity.Epic.Title = req.Title
	epicEntity.Epic.Description = req.Description
	epicEntity.Epic.Tags = req.Tags
	epicEntity.Epic.Dependencies = req.Dependencies
	epicEntity.Epic.Duration = req.Duration
	
	// Parse and validate priority
	if req.Priority != "" {
		priority := epic.Priority(req.Priority)
		if !priority.IsValid() {
			return nil, model.NewValidationError("invalid priority").
				WithContext(req.Priority).
				WithSuggestions([]string{"critical", "high", "medium", "low"})
		}
		epicEntity.Epic.Priority = priority
	}
	
	// Business logic validations
	warnings := []string{}
	suggestions := []string{}
	
	// Check for similar epics (duplicate detection)
	if duplicates, err := s.findSimilarEpics(ctx, epicEntity); err == nil && len(duplicates) > 0 {
		warnings = append(warnings, fmt.Sprintf("Found %d similar epics", len(duplicates)))
		suggestions = append(suggestions, "Review existing epics to avoid duplication")
	}
	
	// Validate dependencies
	if dependencyWarnings, err := s.validateDependencies(ctx, epicEntity.Epic.Dependencies); err == nil {
		warnings = append(warnings, dependencyWarnings...)
	}
	
	// Create the epic using entity manager
	createReq := entity.CreateEntityRequest[*entity.EpicEntity]{
		Entity:    epicEntity,
		SkipHooks: req.SkipValidation,
	}
	
	createResp, err := s.entityManager.Create(ctx, createReq)
	if err != nil {
		return nil, err // EntityManager already provides rich error context
	}
	
	// Generate next actions based on epic state
	nextActions := []string{
		fmt.Sprintf("epic show %s", createResp.Entity.GetID()),
		fmt.Sprintf("story create --epic %s", createResp.Entity.GetID()),
	}
	
	if len(epicEntity.Epic.Dependencies) > 0 {
		nextActions = append(nextActions, "Review and validate epic dependencies")
	}
	
	// Combine warnings
	allWarnings := append(warnings, createResp.Warnings...)
	
	return &CreateEpicWorkflowResponse{
		Epic:        createResp.Entity,
		Warnings:    allWarnings,
		Suggestions: suggestions,
		NextActions: nextActions,
		Metadata: map[string]interface{}{
			"creation_time":     createResp.Entity.GetCreatedAt(),
			"template_used":     req.Template != "",
			"has_dependencies":  len(epicEntity.Epic.Dependencies) > 0,
			"estimated_duration": req.Duration.String(),
		},
	}, nil
}

// UpdateEpicWorkflowRequest represents a request to update an epic.
type UpdateEpicWorkflowRequest struct {
	ID           string                 `json:"id"`
	Title        *string               `json:"title,omitempty"`
	Description  *string               `json:"description,omitempty"`
	Priority     *string               `json:"priority,omitempty"`
	Status       *string               `json:"status,omitempty"`
	Tags         *[]string             `json:"tags,omitempty"`
	Dependencies *[]string             `json:"dependencies,omitempty"`
	Duration     *time.Duration        `json:"duration,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	Force        bool                  `json:"force,omitempty"`
}

// UpdateEpicWorkflowResponse represents the response from updating an epic.
type UpdateEpicWorkflowResponse struct {
	Epic         *entity.EpicEntity        `json:"epic"`
	Changes      []string                  `json:"changes"`
	Warnings     []string                  `json:"warnings,omitempty"`
	Suggestions  []string                  `json:"suggestions,omitempty"`
	NextActions  []string                  `json:"next_actions,omitempty"`
	Metadata     map[string]interface{}    `json:"metadata,omitempty"`
}

// UpdateEpicWorkflow updates an epic with full business logic validation.
func (s *EpicApplicationService) UpdateEpicWorkflow(ctx context.Context, req UpdateEpicWorkflowRequest) (*UpdateEpicWorkflowResponse, error) {
	// Define update function
	updateFunc := func(e *entity.EpicEntity) *entity.EpicEntity {
		if req.Title != nil {
			e.Epic.Title = *req.Title
		}
		if req.Description != nil {
			e.Epic.Description = *req.Description
		}
		if req.Priority != nil {
			e.Epic.Priority = epic.Priority(*req.Priority)
		}
		if req.Status != nil {
			e.Epic.Status = epic.Status(*req.Status)
		}
		if req.Tags != nil {
			e.Epic.Tags = *req.Tags
		}
		if req.Dependencies != nil {
			e.Epic.Dependencies = *req.Dependencies
		}
		if req.Duration != nil {
			e.Epic.Duration = *req.Duration
		}
		return e
	}
	
	// Validate status transition if requested
	if req.Status != nil {
		existing, err := s.entityManager.Get(ctx, req.ID)
		if err != nil {
			return nil, err
		}
		
		newStatus := model.Status(*req.Status)
		if !existing.CanTransitionTo(newStatus) && !req.Force {
			return nil, model.NewWorkflowViolationError(
				existing.GetStatus(),
				newStatus,
			).WithSuggestions([]string{
				"Use --force to override workflow validation",
				"Check valid transitions with 'epic workflow'",
			})
		}
	}
	
	// Validate dependencies if changed
	warnings := []string{}
	if req.Dependencies != nil {
		if dependencyWarnings, err := s.validateDependencies(ctx, *req.Dependencies); err == nil {
			warnings = append(warnings, dependencyWarnings...)
		}
	}
	
	// Update using entity manager
	updateReq := entity.UpdateEntityRequest[*entity.EpicEntity]{
		ID:        req.ID,
		Updates:   updateFunc,
		SkipHooks: req.Force,
	}
	
	updateResp, err := s.entityManager.Update(ctx, updateReq)
	if err != nil {
		return nil, err
	}
	
	// Generate next actions based on new state
	nextActions := s.generateNextActions(updateResp.Entity)
	
	return &UpdateEpicWorkflowResponse{
		Epic:        updateResp.Entity,
		Changes:     updateResp.Changes,
		Warnings:    append(warnings, updateResp.Warnings...),
		NextActions: nextActions,
		Metadata: map[string]interface{}{
			"update_time":      updateResp.Entity.GetUpdatedAt(),
			"changes_count":    len(updateResp.Changes),
			"workflow_status":  updateResp.Entity.GetStatus(),
		},
	}, nil
}

// ListEpicsWorkflowRequest represents a request to list epics with filtering.
type ListEpicsWorkflowRequest struct {
	Status       string    `json:"status,omitempty"`
	Priority     string    `json:"priority,omitempty"`
	Tags         []string  `json:"tags,omitempty"`
	Search       string    `json:"search,omitempty"`
	CreatedAfter *time.Time `json:"created_after,omitempty"`
	CreatedBefore *time.Time `json:"created_before,omitempty"`
	SortBy       string    `json:"sort_by,omitempty"`
	SortOrder    string    `json:"sort_order,omitempty"`
	Limit        int       `json:"limit,omitempty"`
	Offset       int       `json:"offset,omitempty"`
	IncludeStats bool      `json:"include_stats,omitempty"`
}

// ListEpicsWorkflowResponse represents the response from listing epics.
type ListEpicsWorkflowResponse struct {
	Epics       []*entity.EpicEntity      `json:"epics"`
	Total       int                       `json:"total"`
	Filtered    int                       `json:"filtered"`
	Stats       map[string]interface{}    `json:"stats,omitempty"`
	Suggestions []string                  `json:"suggestions,omitempty"`
	Metadata    map[string]interface{}    `json:"metadata"`
}

// ListEpicsWorkflow lists epics with advanced filtering and business logic.
func (s *EpicApplicationService) ListEpicsWorkflow(ctx context.Context, req ListEpicsWorkflowRequest) (*ListEpicsWorkflowResponse, error) {
	// Build filters
	var filters []model.Filter
	
	if req.Status != "" {
		filters = append(filters, &entity.StatusFilter{Status: model.Status(req.Status)})
	}
	
	if req.Priority != "" {
		filters = append(filters, &entity.PriorityFilter{Priority: model.Priority(req.Priority)})
	}
	
	if req.CreatedAfter != nil || req.CreatedBefore != nil {
		filters = append(filters, &entity.DateRangeFilter{
			After:  req.CreatedAfter,
			Before: req.CreatedBefore,
		})
	}
	
	// Combine filters
	var filter model.Filter
	if len(filters) > 1 {
		filter = &entity.MultiFilter{Filters: filters}
	} else if len(filters) == 1 {
		filter = filters[0]
	}
	
	// List using entity manager
	listReq := entity.ListEntitiesRequest{
		Filter:    filter,
		SortBy:    req.SortBy,
		SortOrder: req.SortOrder,
		Limit:     req.Limit,
		Offset:    req.Offset,
	}
	
	listResp, err := s.entityManager.List(ctx, listReq)
	if err != nil {
		return nil, err
	}
	
	// Apply additional filtering (search, tags)
	epics := listResp.Entities
	if req.Search != "" {
		epics = s.filterBySearch(epics, req.Search)
	}
	
	if len(req.Tags) > 0 {
		epics = s.filterByTags(epics, req.Tags)
	}
	
	// Generate stats if requested
	var stats map[string]interface{}
	if req.IncludeStats {
		stats = s.generateEpicStats(listResp.Entities)
	}
	
	// Generate suggestions based on results
	suggestions := s.generateListSuggestions(epics, req)
	
	return &ListEpicsWorkflowResponse{
		Epics:       epics,
		Total:       listResp.Total,
		Filtered:    len(epics),
		Stats:       stats,
		Suggestions: suggestions,
		Metadata: map[string]interface{}{
			"query_time":       time.Now(),
			"filters_applied":  len(filters),
			"search_applied":   req.Search != "",
			"tags_filtered":    len(req.Tags),
		},
	}, nil
}

// DeleteEpicWorkflowRequest represents a request to delete an epic.
type DeleteEpicWorkflowRequest struct {
	ID            string `json:"id"`
	Force         bool   `json:"force,omitempty"`
	SkipValidation bool   `json:"skip_validation,omitempty"`
}

// DeleteEpicWorkflowResponse represents the response from deleting an epic.
type DeleteEpicWorkflowResponse struct {
	Deleted     bool                      `json:"deleted"`
	Epic        *entity.EpicEntity        `json:"epic"` // For audit/undo
	Warnings    []string                  `json:"warnings,omitempty"`
	NextActions []string                  `json:"next_actions,omitempty"`
	Metadata    map[string]interface{}    `json:"metadata"`
}

// DeleteEpicWorkflow deletes an epic with full business logic validation.
func (s *EpicApplicationService) DeleteEpicWorkflow(ctx context.Context, req DeleteEpicWorkflowRequest) (*DeleteEpicWorkflowResponse, error) {
	// Get epic before deletion (for validation and audit)
	epic, err := s.entityManager.Get(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	
	// Business validation
	warnings := []string{}
	if epic.Epic.Status == epic.StatusInProgress && !req.Force {
		return nil, model.NewValidationError("cannot delete epic in progress").
			WithContext(req.ID).
			WithSuggestions([]string{
				"Complete or cancel the epic first",
				"Use --force to override",
			})
	}
	
	if len(epic.Epic.UserStories) > 0 && !req.Force {
		warnings = append(warnings, fmt.Sprintf("Epic has %d user stories that will be orphaned", len(epic.Epic.UserStories)))
	}
	
	// Delete using entity manager
	deleteReq := entity.DeleteEntityRequest{
		ID:        req.ID,
		Force:     req.Force,
		SkipHooks: req.SkipValidation,
	}
	
	deleteResp, err := s.entityManager.Delete(ctx, deleteReq)
	if err != nil {
		return nil, err
	}
	
	nextActions := []string{
		"epic list", // Show remaining epics
	}
	
	if len(epic.Epic.UserStories) > 0 {
		nextActions = append(nextActions, "story list --orphaned")
	}
	
	return &DeleteEpicWorkflowResponse{
		Deleted:     deleteResp.Deleted,
		Epic:        epic,
		Warnings:    append(warnings, deleteResp.Warnings...),
		NextActions: nextActions,
		Metadata: map[string]interface{}{
			"deleted_at":       time.Now(),
			"had_stories":      len(epic.Epic.UserStories) > 0,
			"story_count":      len(epic.Epic.UserStories),
		},
	}, nil
}

// Helper methods for business logic

func (s *EpicApplicationService) findSimilarEpics(ctx context.Context, epic *entity.EpicEntity) ([]*entity.EpicEntity, error) {
	// Simplified similarity detection - in real implementation, use fuzzy matching
	allEpics, err := s.entityManager.List(ctx, entity.ListEntitiesRequest{})
	if err != nil {
		return nil, err
	}
	
	var similar []*entity.EpicEntity
	for _, existing := range allEpics.Entities {
		if strings.Contains(strings.ToLower(existing.Epic.Title), strings.ToLower(epic.Epic.Title)) {
			similar = append(similar, existing)
		}
	}
	
	return similar, nil
}

func (s *EpicApplicationService) validateDependencies(ctx context.Context, dependencies []string) ([]string, error) {
	var warnings []string
	
	for _, depID := range dependencies {
		exists, err := s.entityManager.Exists(ctx, depID)
		if err != nil {
			return warnings, err
		}
		if !exists {
			warnings = append(warnings, fmt.Sprintf("Dependency '%s' not found", depID))
		}
	}
	
	return warnings, nil
}

func (s *EpicApplicationService) generateNextActions(epic *entity.EpicEntity) []string {
	actions := []string{
		fmt.Sprintf("epic show %s", epic.GetID()),
	}
	
	switch epic.Epic.Status {
	case epic.StatusPlanned:
		actions = append(actions, fmt.Sprintf("epic start %s", epic.GetID()))
		actions = append(actions, fmt.Sprintf("story create --epic %s", epic.GetID()))
	case epic.StatusInProgress:
		actions = append(actions, fmt.Sprintf("story list --epic %s", epic.GetID()))
		actions = append(actions, fmt.Sprintf("epic progress %s", epic.GetID()))
	case epic.StatusCompleted:
		actions = append(actions, fmt.Sprintf("epic archive %s", epic.GetID()))
	}
	
	return actions
}

func (s *EpicApplicationService) filterBySearch(epics []*entity.EpicEntity, search string) []*entity.EpicEntity {
	if search == "" {
		return epics
	}
	
	searchLower := strings.ToLower(search)
	var filtered []*entity.EpicEntity
	
	for _, epic := range epics {
		if strings.Contains(strings.ToLower(epic.GetSearchableText()), searchLower) {
			filtered = append(filtered, epic)
		}
	}
	
	return filtered
}

func (s *EpicApplicationService) filterByTags(epics []*entity.EpicEntity, tags []string) []*entity.EpicEntity {
	if len(tags) == 0 {
		return epics
	}
	
	var filtered []*entity.EpicEntity
	
	for _, epic := range epics {
		hasAllTags := true
		for _, tag := range tags {
			if !epic.HasTag(tag) {
				hasAllTags = false
				break
			}
		}
		if hasAllTags {
			filtered = append(filtered, epic)
		}
	}
	
	return filtered
}

func (s *EpicApplicationService) generateEpicStats(epics []*entity.EpicEntity) map[string]interface{} {
	stats := map[string]interface{}{
		"total": len(epics),
		"by_status": make(map[string]int),
		"by_priority": make(map[string]int),
		"completion_avg": 0.0,
	}
	
	var totalCompletion float64
	statusCount := make(map[string]int)
	priorityCount := make(map[string]int)
	
	for _, epic := range epics {
		statusCount[string(epic.Epic.Status)]++
		priorityCount[string(epic.Epic.Priority)]++
		totalCompletion += epic.GetCompletionPercentage()
	}
	
	stats["by_status"] = statusCount
	stats["by_priority"] = priorityCount
	
	if len(epics) > 0 {
		stats["completion_avg"] = totalCompletion / float64(len(epics))
	}
	
	return stats
}

func (s *EpicApplicationService) generateListSuggestions(epics []*entity.EpicEntity, req ListEpicsWorkflowRequest) []string {
	var suggestions []string
	
	if len(epics) == 0 {
		suggestions = append(suggestions, "No epics found. Create your first epic with 'epic create'")
		return suggestions
	}
	
	// Count by status
	statusCount := make(map[epic.Status]int)
	for _, epic := range epics {
		statusCount[epic.Epic.Status]++
	}
	
	if statusCount[epic.StatusPlanned] > statusCount[epic.StatusInProgress] {
		suggestions = append(suggestions, "Consider starting some planned epics")
	}
	
	if statusCount[epic.StatusInProgress] > 5 {
		suggestions = append(suggestions, "You have many epics in progress. Consider focusing on fewer at once")
	}
	
	return suggestions
}

// Hook functions for entity lifecycle

func validateEpicBusinessRules(ctx context.Context, epic *entity.EpicEntity) error {
	// Business rule: Critical epics should have descriptions > 100 chars
	if epic.Epic.Priority == epic.PriorityCritical && len(epic.Epic.Description) < 100 {
		return model.NewValidationError("critical epics require detailed descriptions").
			WithContext(fmt.Sprintf("current: %d chars, required: 100+", len(epic.Epic.Description))).
			WithSuggestion("Provide more details for critical epics")
	}
	
	return nil
}

func calculateEpicProgress(ctx context.Context, epic *entity.EpicEntity) error {
	epic.Epic.CalculateProgress()
	return nil
}

func updateEpicTimestamps(ctx context.Context, epic *entity.EpicEntity) error {
	epic.Epic.UpdatedAt = time.Now()
	return nil
}

func notifyEpicChanges(ctx context.Context, epic *entity.EpicEntity) error {
	// In real implementation, send notifications
	return nil
}

func updateEpicMetrics(ctx context.Context, epic *entity.EpicEntity) error {
	// In real implementation, update metrics
	return nil
}

func checkEpicDependencies(ctx context.Context, epic *entity.EpicEntity) error {
	// In real implementation, check if other epics depend on this one
	return nil
}

func cleanupEpicReferences(ctx context.Context, epic *entity.EpicEntity) error {
	// In real implementation, clean up references to this epic
	return nil
}