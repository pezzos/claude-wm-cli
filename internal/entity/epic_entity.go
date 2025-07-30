// Package entity provides Epic entity implementation using the new generic system.
package entity

import (
	"fmt"
	"strings"
	"time"

	"claude-wm-cli/internal/epic"
	"claude-wm-cli/internal/model"
)

// EpicEntity wraps the existing epic.Epic to implement BaseEntity interface.
// This allows using the generic EntityManager while preserving existing Epic functionality.
type EpicEntity struct {
	*epic.Epic
}

// NewEpicEntity creates a new EpicEntity wrapper.
func NewEpicEntity(e *epic.Epic) *EpicEntity {
	if e == nil {
		e = &epic.Epic{}
	}
	return &EpicEntity{Epic: e}
}

// BaseEntity interface implementation

func (e *EpicEntity) GetID() string {
	return e.Epic.ID
}

func (e *EpicEntity) SetID(id string) {
	e.Epic.ID = id
}

func (e *EpicEntity) GetStatus() model.Status {
	return model.Status(e.Epic.Status)
}

func (e *EpicEntity) SetStatus(status model.Status) {
	e.Epic.Status = epic.Status(status)
}

func (e *EpicEntity) GetPriority() model.Priority {
	return model.Priority(e.Epic.Priority)
}

func (e *EpicEntity) SetPriority(priority model.Priority) {
	e.Epic.Priority = epic.Priority(priority)
}

func (e *EpicEntity) GetCreatedAt() time.Time {
	return e.Epic.CreatedAt
}

func (e *EpicEntity) SetCreatedAt(t time.Time) {
	e.Epic.CreatedAt = t
}

func (e *EpicEntity) GetUpdatedAt() time.Time {
	return e.Epic.UpdatedAt
}

func (e *EpicEntity) SetUpdatedAt(t time.Time) {
	e.Epic.UpdatedAt = t
}

func (e *EpicEntity) Validate() error {
	return e.Epic.Validate()
}

// WorkflowEntity interface implementation

func (e *EpicEntity) CanTransitionTo(newStatus model.Status) bool {
	return e.Epic.Status.CanTransitionTo(epic.Status(newStatus))
}

func (e *EpicEntity) GetValidTransitions() []model.Status {
	epicTransitions := e.Epic.Status.GetValidTransitions()
	transitions := make([]model.Status, len(epicTransitions))
	for i, t := range epicTransitions {
		transitions[i] = model.Status(t)
	}
	return transitions
}

func (e *EpicEntity) GetWorkflowContext() map[string]interface{} {
	return map[string]interface{}{
		"current_status":     e.Epic.Status,
		"valid_transitions":  e.GetValidTransitions(),
		"has_dependencies":   len(e.Epic.Dependencies) > 0,
		"completion_percent": e.GetCompletionPercentage(),
		"blocked_by":         e.GetBlockedBy(),
	}
}

// ProgressTrackingEntity interface implementation

func (e *EpicEntity) GetCompletionPercentage() float64 {
	e.Epic.CalculateProgress()
	return e.Epic.Progress.CompletionPercentage
}

func (e *EpicEntity) GetProgressMetrics() map[string]interface{} {
	e.Epic.CalculateProgress()
	return map[string]interface{}{
		"completion_percentage": e.Epic.Progress.CompletionPercentage,
		"total_stories":        e.Epic.Progress.TotalStories,
		"completed_stories":    e.Epic.Progress.CompletedStories,
		"total_story_points":   e.Epic.Progress.TotalStoryPoints,
		"completed_points":     e.Epic.Progress.CompletedStoryPoints,
		"estimated_hours":      e.Epic.Progress.EstimatedHours,
		"actual_hours":         e.Epic.Progress.ActualHours,
	}
}

func (e *EpicEntity) UpdateProgress() error {
	e.Epic.CalculateProgress()
	return nil
}

// HierarchicalEntity interface implementation (Epics can have Stories as children)

func (e *EpicEntity) GetParentID() string {
	// Epics typically don't have parents, but could be part of larger initiatives
	return ""
}

func (e *EpicEntity) SetParentID(parentID string) {
	// Implementation depends on whether you want to support epic hierarchies
}

func (e *EpicEntity) GetChildIDs() []string {
	childIDs := make([]string, len(e.Epic.UserStories))
	for i, story := range e.Epic.UserStories {
		childIDs[i] = story.ID
	}
	return childIDs
}

func (e *EpicEntity) AddChildID(childID string) {
	// This would typically be handled by the story creation process
	// For now, we'll add a placeholder story
	e.Epic.UserStories = append(e.Epic.UserStories, epic.UserStory{
		ID: childID,
	})
}

func (e *EpicEntity) RemoveChildID(childID string) {
	for i, story := range e.Epic.UserStories {
		if story.ID == childID {
			e.Epic.UserStories = append(e.Epic.UserStories[:i], e.Epic.UserStories[i+1:]...)
			break
		}
	}
}

// TaggableEntity interface implementation

func (e *EpicEntity) GetTags() []string {
	return e.Epic.Tags
}

func (e *EpicEntity) SetTags(tags []string) {
	e.Epic.Tags = tags
}

func (e *EpicEntity) AddTag(tag string) {
	if !e.HasTag(tag) {
		e.Epic.Tags = append(e.Epic.Tags, tag)
	}
}

func (e *EpicEntity) RemoveTag(tag string) {
	for i, t := range e.Epic.Tags {
		if t == tag {
			e.Epic.Tags = append(e.Epic.Tags[:i], e.Epic.Tags[i+1:]...)
			break
		}
	}
}

func (e *EpicEntity) HasTag(tag string) bool {
	for _, t := range e.Epic.Tags {
		if t == tag {
			return true
		}
	}
	return false
}

// TimestampedEntity interface implementation

func (e *EpicEntity) GetCompletedAt() *time.Time {
	if e.Epic.Status == epic.StatusCompleted {
		// In a real implementation, you'd store this separately
		return &e.Epic.UpdatedAt
	}
	return nil
}

func (e *EpicEntity) SetCompletedAt(t *time.Time) {
	// In a real implementation, add CompletedAt field to Epic struct
}

func (e *EpicEntity) GetStartedAt() *time.Time {
	if e.Epic.Status != epic.StatusPlanned {
		// In a real implementation, you'd store this separately
		return &e.Epic.CreatedAt
	}
	return nil
}

func (e *EpicEntity) SetStartedAt(t *time.Time) {
	// In a real implementation, add StartedAt field to Epic struct
}

func (e *EpicEntity) GetDueDate() *time.Time {
	// Epics don't currently have due dates, but could be added
	return nil
}

func (e *EpicEntity) SetDueDate(t *time.Time) {
	// In a real implementation, add DueDate field to Epic struct
}

// DependentEntity interface implementation

func (e *EpicEntity) GetDependencies() []string {
	return e.Epic.Dependencies
}

func (e *EpicEntity) SetDependencies(deps []string) {
	e.Epic.Dependencies = deps
}

func (e *EpicEntity) AddDependency(dep string) {
	if !e.HasDependency(dep) {
		e.Epic.Dependencies = append(e.Epic.Dependencies, dep)
	}
}

func (e *EpicEntity) RemoveDependency(dep string) {
	for i, d := range e.Epic.Dependencies {
		if d == dep {
			e.Epic.Dependencies = append(e.Epic.Dependencies[:i], e.Epic.Dependencies[i+1:]...)
			break
		}
	}
}

func (e *EpicEntity) HasDependency(dep string) bool {
	for _, d := range e.Epic.Dependencies {
		if d == dep {
			return true
		}
	}
	return false
}

func (e *EpicEntity) GetBlockedBy() []string {
	// This would require a more sophisticated dependency analysis
	// For now, return dependencies that are not completed
	return e.Epic.Dependencies // Simplified
}

func (e *EpicEntity) GetBlocking() []string {
	// This would require analyzing other epics that depend on this one
	return []string{} // Would need to be implemented with cross-entity analysis
}

// SearchableEntity interface implementation

func (e *EpicEntity) GetSearchableText() string {
	parts := []string{
		e.Epic.Title,
		e.Epic.Description,
	}
	
	// Add tags
	if len(e.Epic.Tags) > 0 {
		parts = append(parts, strings.Join(e.Epic.Tags, " "))
	}
	
	// Add user story titles
	for _, story := range e.Epic.UserStories {
		if story.Title != "" {
			parts = append(parts, story.Title)
		}
	}
	
	return strings.Join(parts, " ")
}

func (e *EpicEntity) GetSearchKeywords() []string {
	keywords := []string{
		e.Epic.ID,
		string(e.Epic.Status),
		string(e.Epic.Priority),
	}
	
	// Add tags
	keywords = append(keywords, e.Epic.Tags...)
	
	// Add extracted keywords from title and description
	titleWords := strings.Fields(strings.ToLower(e.Epic.Title))
	keywords = append(keywords, titleWords...)
	
	return keywords
}

func (e *EpicEntity) GetSearchMetadata() map[string]interface{} {
	return map[string]interface{}{
		"entity_type":       "epic",
		"status":           string(e.Epic.Status),
		"priority":         string(e.Epic.Priority),
		"tags":             e.Epic.Tags,
		"has_stories":      len(e.Epic.UserStories) > 0,
		"story_count":      len(e.Epic.UserStories),
		"completion":       e.GetCompletionPercentage(),
		"dependencies":     len(e.Epic.Dependencies),
		"created_at":       e.Epic.CreatedAt,
		"updated_at":       e.Epic.UpdatedAt,
	}
}

// EpicEntityFactory implements EntityFactory for Epic entities.
type EpicEntityFactory struct{}

func (f *EpicEntityFactory) CreateEmpty() *EpicEntity {
	return NewEpicEntity(&epic.Epic{
		Status:    epic.StatusPlanned,
		Priority:  epic.PriorityMedium,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})
}

func (f *EpicEntityFactory) CreateFromTemplate(templateID string) (*EpicEntity, error) {
	// In a real implementation, load template from storage
	switch templateID {
	case "feature":
		return NewEpicEntity(&epic.Epic{
			Title:       "New Feature Epic",
			Description: "Epic for implementing a new feature",
			Priority:    epic.PriorityHigh,
			Status:      epic.StatusPlanned,
			Tags:        []string{"feature", "new"},
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}), nil
	case "bugfix":
		return NewEpicEntity(&epic.Epic{
			Title:       "Bug Fix Epic",
			Description: "Epic for fixing critical bugs",
			Priority:    epic.PriorityCritical,
			Status:      epic.StatusPlanned,
			Tags:        []string{"bugfix", "critical"},
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}), nil
	default:
		return nil, model.NewValidationError("unknown template").
			WithContext(templateID).
			WithSuggestions([]string{"feature", "bugfix"})
	}
}

func (f *EpicEntityFactory) CreateFromData(data map[string]interface{}) (*EpicEntity, error) {
	e := f.CreateEmpty()
	
	if title, ok := data["title"].(string); ok {
		e.Epic.Title = title
	}
	
	if desc, ok := data["description"].(string); ok {
		e.Epic.Description = desc
	}
	
	if priority, ok := data["priority"].(string); ok {
		e.Epic.Priority = epic.Priority(priority)
	}
	
	if tags, ok := data["tags"].([]interface{}); ok {
		stringTags := make([]string, len(tags))
		for i, tag := range tags {
			if tagStr, ok := tag.(string); ok {
				stringTags[i] = tagStr
			}
		}
		e.Epic.Tags = stringTags
	}
	
	return e, nil
}

func (f *EpicEntityFactory) GetEntityType() string {
	return "epic"
}

func (f *EpicEntityFactory) GetDefaultValues() map[string]interface{} {
	return map[string]interface{}{
		"status":   string(epic.StatusPlanned),
		"priority": string(epic.PriorityMedium),
		"tags":     []string{},
	}
}

// EpicEntityValidator implements EntityValidator for Epic entities.
type EpicEntityValidator struct{}

func (v *EpicEntityValidator) Validate(entity *EpicEntity) error {
	if entity.Epic.Title == "" {
		return model.NewValidationError("epic title is required").
			WithSuggestion("Provide a descriptive title for the epic")
	}
	
	if len(entity.Epic.Title) > 200 {
		return model.NewValidationError("epic title too long").
			WithContext(fmt.Sprintf("current: %d chars, max: 200", len(entity.Epic.Title))).
			WithSuggestion("Shorten the epic title to under 200 characters")
	}
	
	if entity.Epic.Description == "" {
		return model.NewValidationError("epic description is required").
			WithSuggestion("Provide a detailed description of the epic")
	}
	
	return entity.Epic.Validate()
}

func (v *EpicEntityValidator) ValidateField(entity *EpicEntity, fieldName string) error {
	switch fieldName {
	case "title":
		if entity.Epic.Title == "" {
			return model.NewValidationError("title is required")
		}
		if len(entity.Epic.Title) > 200 {
			return model.NewValidationError("title too long")
		}
	case "description":
		if entity.Epic.Description == "" {
			return model.NewValidationError("description is required")
		}
	case "priority":
		if !entity.Epic.Priority.IsValid() {
			return model.NewValidationError("invalid priority")
		}
	case "status":
		if !entity.Epic.Status.IsValid() {
			return model.NewValidationError("invalid status")
		}
	default:
		return model.NewValidationError("unknown field").WithContext(fieldName)
	}
	return nil
}

func (v *EpicEntityValidator) ValidateTransition(entity *EpicEntity, newStatus model.Status) error {
	if !entity.CanTransitionTo(newStatus) {
		return model.NewWorkflowViolationError(
			model.Status(entity.Epic.Status),
			newStatus,
		)
	}
	return nil
}

func (v *EpicEntityValidator) GetValidationRules() map[string]interface{} {
	return map[string]interface{}{
		"title": map[string]interface{}{
			"required":  true,
			"max_length": 200,
		},
		"description": map[string]interface{}{
			"required": true,
		},
		"priority": map[string]interface{}{
			"required": true,
			"enum":     []string{"critical", "high", "medium", "low"},
		},
		"status": map[string]interface{}{
			"required": true,
			"enum":     []string{"planned", "in-progress", "completed", "cancelled"},
		},
	}
}