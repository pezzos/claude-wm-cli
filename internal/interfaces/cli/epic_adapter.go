// Package cli contains interface adapters that connect the external CLI interface to the application layer.
// These adapters convert between CLI concerns and domain/application concerns.
package cli

import (
	"context"
	"fmt"
	"strings"
	"time"

	"claude-wm-cli/internal/domain/entities"
	"claude-wm-cli/internal/domain/repositories"
	"claude-wm-cli/internal/domain/services"
	"claude-wm-cli/internal/domain/valueobjects"
)

// EpicCLIAdapter adapts CLI commands to domain operations.
// This is an interface adapter that converts CLI-specific concerns to clean domain operations.
type EpicCLIAdapter struct {
	epicRepo    repositories.EpicRepository
	epicService *services.EpicDomainService
}

// NewEpicCLIAdapter creates a new epic CLI adapter.
func NewEpicCLIAdapter(epicRepo repositories.EpicRepository, epicService *services.EpicDomainService) *EpicCLIAdapter {
	return &EpicCLIAdapter{
		epicRepo:    epicRepo,
		epicService: epicService,
	}
}

// CreateEpicRequest represents the CLI request to create an epic.
type CreateEpicRequest struct {
	ID          string
	Title       string
	Description string
	Priority    string
	Tags        []string
	Duration    string
}

// UpdateEpicRequest represents the CLI request to update an epic.
type UpdateEpicRequest struct {
	ID          string
	Title       *string
	Description *string
	Priority    *string
	Status      *string
	Tags        *[]string
	Duration    *string
}

// EpicListOptions represents CLI options for listing epics.
type EpicListOptions struct {
	Status      string
	Priority    string
	Tags        []string
	ShowAll     bool
	SearchQuery string
	Limit       int
	Offset      int
}

// EpicResponse represents the CLI response for epic operations.
type EpicResponse struct {
	ID                   string                    `json:"id"`
	Title                string                    `json:"title"`
	Description          string                    `json:"description"`
	Priority             string                    `json:"priority"`
	Status               string                    `json:"status"`
	Tags                 []string                  `json:"tags"`
	Dependencies         []string                  `json:"dependencies"`
	UserStories          []UserStoryResponse       `json:"user_stories"`
	Progress             ProgressResponse          `json:"progress"`
	CreatedAt            time.Time                 `json:"created_at"`
	UpdatedAt            time.Time                 `json:"updated_at"`
	StartDate            *time.Time                `json:"start_date,omitempty"`
	EndDate              *time.Time                `json:"end_date,omitempty"`
	Duration             string                    `json:"duration,omitempty"`
}

// UserStoryResponse represents the CLI response for user stories.
type UserStoryResponse struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Priority    string   `json:"priority"`
	Status      string   `json:"status"`
	StoryPoints int      `json:"story_points"`
	Tags        []string `json:"tags"`
}

// ProgressResponse represents the CLI response for progress metrics.
type ProgressResponse struct {
	TotalStoryPoints     int     `json:"total_story_points"`
	CompletedStoryPoints int     `json:"completed_story_points"`
	TotalStories         int     `json:"total_stories"`
	CompletedStories     int     `json:"completed_stories"`
	CompletionPercentage float64 `json:"completion_percentage"`
	EstimatedHours       int     `json:"estimated_hours"`
	ActualHours          int     `json:"actual_hours"`
}

// CreateEpic creates a new epic from CLI request.
func (a *EpicCLIAdapter) CreateEpic(ctx context.Context, req CreateEpicRequest) (*EpicResponse, error) {
	// Convert CLI priority to domain priority
	priority, err := a.parsePriority(req.Priority)
	if err != nil {
		return nil, fmt.Errorf("invalid priority: %w", err)
	}

	// Validate creation using domain service
	if err := a.epicService.ValidateEpicCreation(ctx, req.ID, req.Title, req.Description, priority); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Create domain entity
	epic, err := entities.NewEpic(req.ID, req.Title, req.Description, priority)
	if err != nil {
		return nil, fmt.Errorf("failed to create epic: %w", err)
	}

	// Set optional fields
	if req.Duration != "" {
		epic.SetDuration(req.Duration)
	}

	// Add tags
	for _, tag := range req.Tags {
		epic.AddTag(tag)
	}

	// Persist the epic
	if err := a.epicRepo.Create(ctx, epic); err != nil {
		return nil, fmt.Errorf("failed to save epic: %w", err)
	}

	return a.entityToResponse(epic), nil
}

// GetEpic retrieves an epic by ID.
func (a *EpicCLIAdapter) GetEpic(ctx context.Context, id string) (*EpicResponse, error) {
	epic, err := a.epicRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get epic: %w", err)
	}

	return a.entityToResponse(epic), nil
}

// UpdateEpic updates an existing epic.
func (a *EpicCLIAdapter) UpdateEpic(ctx context.Context, req UpdateEpicRequest) (*EpicResponse, error) {
	// Get existing epic
	epic, err := a.epicRepo.GetByID(ctx, req.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get epic: %w", err)
	}

	// Apply updates
	if req.Title != nil {
		if err := epic.UpdateTitle(*req.Title); err != nil {
			return nil, fmt.Errorf("failed to update title: %w", err)
		}
	}

	if req.Description != nil {
		if err := epic.UpdateDescription(*req.Description); err != nil {
			return nil, fmt.Errorf("failed to update description: %w", err)
		}
	}

	if req.Priority != nil {
		priority, err := a.parsePriority(*req.Priority)
		if err != nil {
			return nil, fmt.Errorf("invalid priority: %w", err)
		}
		if err := epic.UpdatePriority(priority); err != nil {
			return nil, fmt.Errorf("failed to update priority: %w", err)
		}
	}

	if req.Status != nil {
		status, err := a.parseStatus(*req.Status)
		if err != nil {
			return nil, fmt.Errorf("invalid status: %w", err)
		}

		// Validate status transition using domain service
		if err := a.epicService.CanTransitionEpicStatus(ctx, req.ID, status); err != nil {
			return nil, fmt.Errorf("invalid status transition: %w", err)
		}

		if err := epic.TransitionTo(status); err != nil {
			return nil, fmt.Errorf("failed to update status: %w", err)
		}
	}

	if req.Duration != nil {
		epic.SetDuration(*req.Duration)
	}

	if req.Tags != nil {
		// Clear existing tags and add new ones
		for _, tag := range epic.Tags() {
			epic.RemoveTag(tag)
		}
		for _, tag := range *req.Tags {
			epic.AddTag(tag)
		}
	}

	// Persist changes
	if err := a.epicRepo.Update(ctx, epic); err != nil {
		return nil, fmt.Errorf("failed to save epic: %w", err)
	}

	return a.entityToResponse(epic), nil
}

// DeleteEpic deletes an epic.
func (a *EpicCLIAdapter) DeleteEpic(ctx context.Context, id string) error {
	// Validate deletion using domain service
	if err := a.epicService.ValidateEpicDeletion(ctx, id); err != nil {
		return fmt.Errorf("cannot delete epic: %w", err)
	}

	if err := a.epicRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete epic: %w", err)
	}

	return nil
}

// ListEpics lists epics based on CLI options.
func (a *EpicCLIAdapter) ListEpics(ctx context.Context, opts EpicListOptions) ([]*EpicResponse, error) {
	filter := repositories.EpicFilter{
		Limit:  opts.Limit,
		Offset: opts.Offset,
	}

	// Parse CLI options to domain filter
	if opts.Status != "" {
		status, err := a.parseStatus(opts.Status)
		if err != nil {
			return nil, fmt.Errorf("invalid status filter: %w", err)
		}
		filter.Status = &status
	}

	if opts.Priority != "" {
		priority, err := a.parsePriority(opts.Priority)
		if err != nil {
			return nil, fmt.Errorf("invalid priority filter: %w", err)
		}
		filter.Priority = &priority
	}

	if len(opts.Tags) > 0 {
		filter.Tags = opts.Tags
	}

	var epics []*entities.Epic
	var err error

	// Handle search query
	if opts.SearchQuery != "" {
		epics, err = a.epicRepo.Search(ctx, opts.SearchQuery)
	} else {
		epics, err = a.epicRepo.List(ctx, filter)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to list epics: %w", err)
	}

	// Convert to CLI responses
	responses := make([]*EpicResponse, len(epics))
	for i, epic := range epics {
		responses[i] = a.entityToResponse(epic)
	}

	return responses, nil
}

// GetEpicStatistics returns epic statistics for CLI display.
func (a *EpicCLIAdapter) GetEpicStatistics(ctx context.Context) (*StatisticsResponse, error) {
	// Get all epics
	epics, err := a.epicRepo.List(ctx, repositories.EpicFilter{})
	if err != nil {
		return nil, fmt.Errorf("failed to get epics: %w", err)
	}

	stats := &StatisticsResponse{
		TotalEpics:      len(epics),
		EpicsByStatus:   make(map[string]int),
		EpicsByPriority: make(map[string]int),
	}

	totalStoryPoints := 0
	completedStoryPoints := 0
	blockedCount := 0
	overdueCount := 0
	now := time.Now()

	for _, epic := range epics {
		// Count by status
		stats.EpicsByStatus[epic.Status().String()]++

		// Count by priority
		stats.EpicsByPriority[epic.Priority().String()]++

		// Aggregate progress
		progress := epic.Progress()
		totalStoryPoints += progress.TotalStoryPoints
		completedStoryPoints += progress.CompletedStoryPoints

		// Check if blocked (has unresolved dependencies)
		if len(epic.Dependencies()) > 0 {
			blockedCount++ // Simplified check
		}

		// Check if overdue
		if epic.EndDate() != nil && epic.EndDate().Before(now) && !epic.IsCompleted() {
			overdueCount++
		}
	}

	stats.TotalStoryPoints = totalStoryPoints
	stats.CompletedStoryPoints = completedStoryPoints
	if totalStoryPoints > 0 {
		stats.OverallProgress = float64(completedStoryPoints) / float64(totalStoryPoints) * 100
	}
	stats.BlockedEpics = blockedCount
	stats.OverdueEpics = overdueCount

	return stats, nil
}

// StatisticsResponse represents CLI response for epic statistics.
type StatisticsResponse struct {
	TotalEpics           int            `json:"total_epics"`
	EpicsByStatus        map[string]int `json:"epics_by_status"`
	EpicsByPriority      map[string]int `json:"epics_by_priority"`
	TotalStoryPoints     int            `json:"total_story_points"`
	CompletedStoryPoints int            `json:"completed_story_points"`
	OverallProgress      float64        `json:"overall_progress"`
	BlockedEpics         int            `json:"blocked_epics"`
	OverdueEpics         int            `json:"overdue_epics"`
}

// Helper methods for parsing CLI inputs to domain values

func (a *EpicCLIAdapter) parsePriority(priorityStr string) (valueobjects.Priority, error) {
	// Handle legacy format
	if priorityStr == "critical" || priorityStr == "high" || priorityStr == "medium" || priorityStr == "low" {
		return valueobjects.NewPriorityFromLegacy(priorityStr), nil
	}

	// Handle standard format
	return valueobjects.NewPriority(strings.ToUpper(priorityStr))
}

func (a *EpicCLIAdapter) parseStatus(statusStr string) (valueobjects.Status, error) {
	// Handle legacy format
	if statusStr == "todo" || statusStr == "done" {
		return valueobjects.NewStatusFromLegacy(statusStr), nil
	}

	// Handle standard format
	return valueobjects.NewStatus(strings.ToLower(statusStr))
}

// Helper method to convert domain entity to CLI response

func (a *EpicCLIAdapter) entityToResponse(epic *entities.Epic) *EpicResponse {
	userStories := make([]UserStoryResponse, len(epic.UserStories()))
	for i, story := range epic.UserStories() {
		userStories[i] = UserStoryResponse{
			ID:          story.ID,
			Title:       story.Title,
			Description: story.Description,
			Priority:    story.Priority.String(),
			Status:      story.Status.String(),
			StoryPoints: story.StoryPoints,
			Tags:        story.Tags,
		}
	}

	progress := epic.Progress()
	return &EpicResponse{
		ID:           epic.ID(),
		Title:        epic.Title(),
		Description:  epic.Description(),
		Priority:     epic.Priority().String(),
		Status:       epic.Status().String(),
		Tags:         epic.Tags(),
		Dependencies: epic.Dependencies(),
		UserStories:  userStories,
		Progress: ProgressResponse{
			TotalStoryPoints:     progress.TotalStoryPoints,
			CompletedStoryPoints: progress.CompletedStoryPoints,
			TotalStories:         progress.TotalStories,
			CompletedStories:     progress.CompletedStories,
			CompletionPercentage: progress.CompletionPercentage,
			EstimatedHours:       progress.EstimatedHours,
			ActualHours:          progress.ActualHours,
		},
		CreatedAt: epic.CreatedAt(),
		UpdatedAt: epic.UpdatedAt(),
		StartDate: epic.StartDate(),
		EndDate:   epic.EndDate(),
		Duration:  epic.Duration(),
	}
}