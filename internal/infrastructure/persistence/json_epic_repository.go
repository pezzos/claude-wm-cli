// Package persistence contains infrastructure implementations for data persistence.
// This package implements repository interfaces defined in the domain layer.
package persistence

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"claude-wm-cli/internal/domain/entities"
	"claude-wm-cli/internal/domain/repositories"
	"claude-wm-cli/internal/domain/valueobjects"
)

// JSONEpicRepository implements the EpicRepository interface using JSON file storage.
// This is an infrastructure concern that implements the domain repository interface.
type JSONEpicRepository struct {
	filePath string
	data     *EpicCollection
}

// EpicCollection represents the JSON structure for storing epics.
type EpicCollection struct {
	ProjectID   string                  `json:"project_id"`
	Epics       map[string]*EpicData    `json:"epics"`
	CurrentEpic string                  `json:"current_epic,omitempty"`
	Metadata    CollectionMetadata      `json:"metadata"`
}

// EpicData represents the JSON structure for a single epic.
type EpicData struct {
	ID           string                `json:"id"`
	Title        string                `json:"title"`
	Description  string                `json:"description"`
	Priority     string                `json:"priority"`
	Status       string                `json:"status"`
	StartDate    *time.Time            `json:"start_date,omitempty"`
	EndDate      *time.Time            `json:"end_date,omitempty"`
	Duration     string                `json:"duration,omitempty"`
	Tags         []string              `json:"tags,omitempty"`
	Dependencies []string              `json:"dependencies,omitempty"`
	UserStories  []UserStoryData       `json:"user_stories,omitempty"`
	Progress     ProgressMetricsData   `json:"progress"`
	CreatedAt    time.Time             `json:"created_at"`
	UpdatedAt    time.Time             `json:"updated_at"`
}

// UserStoryData represents the JSON structure for a user story.
type UserStoryData struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Priority    string   `json:"priority"`
	Status      string   `json:"status"`
	StoryPoints int      `json:"story_points,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

// ProgressMetricsData represents the JSON structure for progress metrics.
type ProgressMetricsData struct {
	TotalStoryPoints     int        `json:"total_story_points"`
	CompletedStoryPoints int        `json:"completed_story_points"`
	TotalStories         int        `json:"total_stories"`
	CompletedStories     int        `json:"completed_stories"`
	CompletionPercentage float64    `json:"completion_percentage"`
	EstimatedHours       int        `json:"estimated_hours"`
	ActualHours          int        `json:"actual_hours"`
	EstimatedEndDate     *time.Time `json:"estimated_end_date,omitempty"`
}

// CollectionMetadata contains metadata about the epic collection.
type CollectionMetadata struct {
	Version     string    `json:"version"`
	LastUpdated time.Time `json:"last_updated"`
	TotalEpics  int       `json:"total_epics"`
}

// NewJSONEpicRepository creates a new JSON-based epic repository.
func NewJSONEpicRepository(filePath string) (*JSONEpicRepository, error) {
	repo := &JSONEpicRepository{
		filePath: filePath,
	}

	if err := repo.load(); err != nil {
		return nil, fmt.Errorf("failed to load epic repository: %w", err)
	}

	return repo, nil
}

// Create saves a new epic to the repository.
func (r *JSONEpicRepository) Create(ctx context.Context, epic *entities.Epic) error {
	if _, exists := r.data.Epics[epic.ID()]; exists {
		return fmt.Errorf("epic with ID %s already exists", epic.ID())
	}

	epicData := r.domainToData(epic)
	r.data.Epics[epic.ID()] = epicData
	r.data.Metadata.TotalEpics = len(r.data.Epics)
	r.data.Metadata.LastUpdated = time.Now()

	return r.save()
}

// GetByID retrieves an epic by its unique identifier.
func (r *JSONEpicRepository) GetByID(ctx context.Context, id string) (*entities.Epic, error) {
	epicData, exists := r.data.Epics[id]
	if !exists {
		return nil, fmt.Errorf("epic with ID %s not found", id)
	}

	return r.dataToDomain(epicData)
}

// Update saves changes to an existing epic.
func (r *JSONEpicRepository) Update(ctx context.Context, epic *entities.Epic) error {
	if _, exists := r.data.Epics[epic.ID()]; !exists {
		return fmt.Errorf("epic with ID %s not found", epic.ID())
	}

	epicData := r.domainToData(epic)
	r.data.Epics[epic.ID()] = epicData
	r.data.Metadata.LastUpdated = time.Now()

	return r.save()
}

// Delete removes an epic from the repository.
func (r *JSONEpicRepository) Delete(ctx context.Context, id string) error {
	if _, exists := r.data.Epics[id]; !exists {
		return fmt.Errorf("epic with ID %s not found", id)
	}

	delete(r.data.Epics, id)
	r.data.Metadata.TotalEpics = len(r.data.Epics)
	r.data.Metadata.LastUpdated = time.Now()

	return r.save()
}

// List retrieves epics based on the provided filter criteria.
func (r *JSONEpicRepository) List(ctx context.Context, filter repositories.EpicFilter) ([]*entities.Epic, error) {
	var epics []*entities.Epic

	for _, epicData := range r.data.Epics {
		if r.matchesFilter(epicData, filter) {
			epic, err := r.dataToDomain(epicData)
			if err != nil {
				continue // Skip invalid epics
			}
			epics = append(epics, epic)
		}
	}

	// Apply sorting and pagination
	epics = r.applySorting(epics, filter.SortBy, filter.SortOrder)
	epics = r.applyPagination(epics, filter.Offset, filter.Limit)

	return epics, nil
}

// Exists checks if an epic with the given ID exists.
func (r *JSONEpicRepository) Exists(ctx context.Context, id string) (bool, error) {
	_, exists := r.data.Epics[id]
	return exists, nil
}

// Count returns the total number of epics matching the filter.
func (r *JSONEpicRepository) Count(ctx context.Context, filter repositories.EpicFilter) (int, error) {
	count := 0
	for _, epicData := range r.data.Epics {
		if r.matchesFilter(epicData, filter) {
			count++
		}
	}
	return count, nil
}

// GetByStatus retrieves all epics with the specified status.
func (r *JSONEpicRepository) GetByStatus(ctx context.Context, status valueobjects.Status) ([]*entities.Epic, error) {
	filter := repositories.EpicFilter{
		Status: &status,
	}
	return r.List(ctx, filter)
}

// GetByPriority retrieves all epics with the specified priority.
func (r *JSONEpicRepository) GetByPriority(ctx context.Context, priority valueobjects.Priority) ([]*entities.Epic, error) {
	filter := repositories.EpicFilter{
		Priority: &priority,
	}
	return r.List(ctx, filter)
}

// GetByTag retrieves all epics that have the specified tag.
func (r *JSONEpicRepository) GetByTag(ctx context.Context, tag string) ([]*entities.Epic, error) {
	filter := repositories.EpicFilter{
		Tags: []string{tag},
	}
	return r.List(ctx, filter)
}

// Search performs a text search across epic titles, descriptions, and tags.
func (r *JSONEpicRepository) Search(ctx context.Context, query string) ([]*entities.Epic, error) {
	var epics []*entities.Epic
	queryLower := strings.ToLower(query)

	for _, epicData := range r.data.Epics {
		// Search in title, description, and tags
		searchText := strings.ToLower(fmt.Sprintf("%s %s %s", 
			epicData.Title, epicData.Description, strings.Join(epicData.Tags, " ")))
		
		if strings.Contains(searchText, queryLower) {
			epic, err := r.dataToDomain(epicData)
			if err != nil {
				continue // Skip invalid epics
			}
			epics = append(epics, epic)
		}
	}

	return epics, nil
}

// GetDependents returns epics that depend on the specified epic.
func (r *JSONEpicRepository) GetDependents(ctx context.Context, epicID string) ([]*entities.Epic, error) {
	var dependents []*entities.Epic

	for _, epicData := range r.data.Epics {
		for _, depID := range epicData.Dependencies {
			if depID == epicID {
				epic, err := r.dataToDomain(epicData)
				if err != nil {
					continue // Skip invalid epics
				}
				dependents = append(dependents, epic)
				break
			}
		}
	}

	return dependents, nil
}

// GetBlocked returns epics that are blocked by unresolved dependencies.
func (r *JSONEpicRepository) GetBlocked(ctx context.Context) ([]*entities.Epic, error) {
	var blocked []*entities.Epic

	for _, epicData := range r.data.Epics {
		if len(epicData.Dependencies) == 0 {
			continue
		}

		isBlocked := false
		for _, depID := range epicData.Dependencies {
			if depData, exists := r.data.Epics[depID]; exists {
				if depData.Status != string(valueobjects.Completed) {
					isBlocked = true
					break
				}
			} else {
				isBlocked = true // Dependency doesn't exist
				break
			}
		}

		if isBlocked {
			epic, err := r.dataToDomain(epicData)
			if err != nil {
				continue // Skip invalid epics
			}
			blocked = append(blocked, epic)
		}
	}

	return blocked, nil
}

// GetActive returns all epics that are currently in progress.
func (r *JSONEpicRepository) GetActive(ctx context.Context) ([]*entities.Epic, error) {
	status := valueobjects.InProgress
	return r.GetByStatus(ctx, status)
}

// GetOverdue returns epics that are past their end date and not completed.
func (r *JSONEpicRepository) GetOverdue(ctx context.Context) ([]*entities.Epic, error) {
	var overdue []*entities.Epic
	now := time.Now()

	for _, epicData := range r.data.Epics {
		if epicData.Status == string(valueobjects.Completed) || epicData.Status == string(valueobjects.Cancelled) {
			continue
		}

		if epicData.EndDate != nil && epicData.EndDate.Before(now) {
			epic, err := r.dataToDomain(epicData)
			if err != nil {
				continue // Skip invalid epics
			}
			overdue = append(overdue, epic)
		}
	}

	return overdue, nil
}

// Helper methods

func (r *JSONEpicRepository) load() error {
	if err := r.ensureFileExists(); err != nil {
		return err
	}

	data, err := os.ReadFile(r.filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var collection EpicCollection
	if err := json.Unmarshal(data, &collection); err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	r.data = &collection
	return nil
}

func (r *JSONEpicRepository) save() error {
	data, err := json.MarshalIndent(r.data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	if err := os.WriteFile(r.filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

func (r *JSONEpicRepository) ensureFileExists() error {
	if _, err := os.Stat(r.filePath); os.IsNotExist(err) {
		// Create directory if it doesn't exist
		dir := filepath.Dir(r.filePath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}

		// Create empty collection
		collection := EpicCollection{
			ProjectID: "default",
			Epics:     make(map[string]*EpicData),
			Metadata: CollectionMetadata{
				Version:     "1.0",
				LastUpdated: time.Now(),
				TotalEpics:  0,
			},
		}

		data, err := json.MarshalIndent(collection, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal initial JSON: %w", err)
		}

		if err := os.WriteFile(r.filePath, data, 0644); err != nil {
			return fmt.Errorf("failed to write initial file: %w", err)
		}
	}

	return nil
}

func (r *JSONEpicRepository) domainToData(epic *entities.Epic) *EpicData {
	userStories := make([]UserStoryData, len(epic.UserStories()))
	for i, story := range epic.UserStories() {
		userStories[i] = UserStoryData{
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
	return &EpicData{
		ID:           epic.ID(),
		Title:        epic.Title(),
		Description:  epic.Description(),
		Priority:     epic.Priority().String(),
		Status:       epic.Status().String(),
		StartDate:    epic.StartDate(),
		EndDate:      epic.EndDate(),
		Duration:     epic.Duration(),
		Tags:         epic.Tags(),
		Dependencies: epic.Dependencies(),
		UserStories:  userStories,
		Progress: ProgressMetricsData{
			TotalStoryPoints:     progress.TotalStoryPoints,
			CompletedStoryPoints: progress.CompletedStoryPoints,
			TotalStories:         progress.TotalStories,
			CompletedStories:     progress.CompletedStories,
			CompletionPercentage: progress.CompletionPercentage,
			EstimatedHours:       progress.EstimatedHours,
			ActualHours:          progress.ActualHours,
			EstimatedEndDate:     progress.EstimatedEndDate,
		},
		CreatedAt: epic.CreatedAt(),
		UpdatedAt: epic.UpdatedAt(),
	}
}

func (r *JSONEpicRepository) dataToDomain(data *EpicData) (*entities.Epic, error) {
	priority, err := valueobjects.NewPriority(data.Priority)
	if err != nil {
		priority = valueobjects.P2 // Default fallback
	}

	epic, err := entities.NewEpic(data.ID, data.Title, data.Description, priority)
	if err != nil {
		return nil, fmt.Errorf("failed to create epic: %w", err)
	}

	// Set status
	if status, err := valueobjects.NewStatus(data.Status); err == nil {
		epic.TransitionTo(status)
	}

	// Set optional fields
	epic.SetStartDate(data.StartDate)
	epic.SetEndDate(data.EndDate)
	epic.SetDuration(data.Duration)

	// Set tags
	for _, tag := range data.Tags {
		epic.AddTag(tag)
	}

	// Set dependencies
	for _, dep := range data.Dependencies {
		epic.AddDependency(dep)
	}

	// Set user stories
	for _, storyData := range data.UserStories {
		storyPriority, _ := valueobjects.NewPriority(storyData.Priority)
		storyStatus, _ := valueobjects.NewStatus(storyData.Status)
		
		story := entities.UserStory{
			ID:          storyData.ID,
			Title:       storyData.Title,
			Description: storyData.Description,
			Priority:    storyPriority,
			Status:      storyStatus,
			StoryPoints: storyData.StoryPoints,
			Tags:        storyData.Tags,
		}
		
		epic.AddUserStory(story)
	}

	return epic, nil
}

func (r *JSONEpicRepository) matchesFilter(data *EpicData, filter repositories.EpicFilter) bool {
	// Status filter
	if filter.Status != nil && data.Status != filter.Status.String() {
		return false
	}

	// Priority filter
	if filter.Priority != nil && data.Priority != filter.Priority.String() {
		return false
	}

	// Tags filter (epic must have all specified tags)
	if len(filter.Tags) > 0 {
		for _, filterTag := range filter.Tags {
			found := false
			for _, epicTag := range data.Tags {
				if epicTag == filterTag {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
	}

	// Dependencies filter
	if filter.HasDependencies != nil {
		hasDeps := len(data.Dependencies) > 0
		if *filter.HasDependencies != hasDeps {
			return false
		}
	}

	// Time filters
	if filter.CreatedAfter != nil && data.CreatedAt.Unix() <= *filter.CreatedAfter {
		return false
	}
	if filter.CreatedBefore != nil && data.CreatedAt.Unix() >= *filter.CreatedBefore {
		return false
	}
	if filter.UpdatedAfter != nil && data.UpdatedAt.Unix() <= *filter.UpdatedAfter {
		return false
	}
	if filter.UpdatedBefore != nil && data.UpdatedAt.Unix() >= *filter.UpdatedBefore {
		return false
	}

	return true
}

func (r *JSONEpicRepository) applySorting(epics []*entities.Epic, sortBy, sortOrder string) []*entities.Epic {
	// Implement sorting logic based on sortBy and sortOrder
	// For now, return as-is (could implement proper sorting later)
	return epics
}

func (r *JSONEpicRepository) applyPagination(epics []*entities.Epic, offset, limit int) []*entities.Epic {
	if offset >= len(epics) {
		return []*entities.Epic{}
	}

	end := len(epics)
	if limit > 0 && offset+limit < end {
		end = offset + limit
	}

	return epics[offset:end]
}