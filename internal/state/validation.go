package state

import (
	"fmt"
	"regexp"
	"time"
)

// ValidationError represents a validation failure
type ValidationError struct {
	Field    string      `json:"field"`
	Value    interface{} `json:"value"`
	Rule     string      `json:"rule"`
	Message  string      `json:"message"`
	Severity string      `json:"severity"`
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("validation error in field '%s': %s", e.Field, e.Message)
}

// ValidationResult contains the outcome of validation
type ValidationResult struct {
	Valid    bool              `json:"valid"`
	Errors   []ValidationError `json:"errors,omitempty"`
	Warnings []ValidationError `json:"warnings,omitempty"`
}

// Validator provides schema validation functionality
type Validator struct {
	schema *SchemaDefinition
	rules  map[string][]ValidationRule
}

// NewValidator creates a new validator with default schema
func NewValidator() *Validator {
	return &Validator{
		schema: getDefaultSchema(),
		rules:  getDefaultValidationRules(),
	}
}

// ValidateProject validates a project state
func (v *Validator) ValidateProject(project *ProjectState) *ValidationResult {
	result := &ValidationResult{Valid: true}

	// Validate required fields
	if project.Name == "" {
		result.addError("name", project.Name, "required", "Project name is required")
	}

	if project.Metadata.ID == "" {
		result.addError("metadata.id", project.Metadata.ID, "required", "Project ID is required")
	}

	// Validate name format
	if project.Name != "" && !isValidName(project.Name) {
		result.addError("name", project.Name, "format", "Project name must contain only letters, numbers, hyphens, and underscores")
	}

	// Validate status
	if !isValidStatus(project.Status) {
		result.addError("status", project.Status, "enum", "Invalid project status")
	}

	// Validate settings
	if project.Settings.DefaultTimeout < 1 || project.Settings.DefaultTimeout > 3600 {
		result.addError("settings.default_timeout", project.Settings.DefaultTimeout, "range", "Default timeout must be between 1 and 3600 seconds")
	}

	if project.Settings.DefaultRetries < 0 || project.Settings.DefaultRetries > 10 {
		result.addError("settings.default_retries", project.Settings.DefaultRetries, "range", "Default retries must be between 0 and 10")
	}

	// Validate metrics
	if project.Metrics.ProgressPercent < 0 || project.Metrics.ProgressPercent > 100 {
		result.addError("metrics.progress_percent", project.Metrics.ProgressPercent, "range", "Progress percent must be between 0 and 100")
	}

	return result
}

// ValidateEpic validates an epic state
func (v *Validator) ValidateEpic(epic *EpicState) *ValidationResult {
	result := &ValidationResult{Valid: true}

	// Validate required fields
	if epic.Title == "" {
		result.addError("title", epic.Title, "required", "Epic title is required")
	}

	if epic.ProjectID == "" {
		result.addError("project_id", epic.ProjectID, "required", "Epic must belong to a project")
	}

	if epic.Metadata.ID == "" {
		result.addError("metadata.id", epic.Metadata.ID, "required", "Epic ID is required")
	}

	// Validate priority
	if !isValidPriority(epic.Priority) {
		result.addError("priority", epic.Priority, "enum", "Invalid epic priority")
	}

	// Validate status
	if !isValidStatus(epic.Status) {
		result.addError("status", epic.Status, "enum", "Invalid epic status")
	}

	// Validate dates
	if epic.StartDate != nil && epic.EndDate != nil && epic.StartDate.After(*epic.EndDate) {
		result.addError("end_date", epic.EndDate, "logic", "End date must be after start date")
	}

	// Validate hours
	if epic.EstimatedHours < 0 {
		result.addError("estimated_hours", epic.EstimatedHours, "range", "Estimated hours cannot be negative")
	}

	if epic.ActualHours < 0 {
		result.addError("actual_hours", epic.ActualHours, "range", "Actual hours cannot be negative")
	}

	// Check for circular dependencies
	if containsString(epic.Dependencies, epic.Metadata.ID) {
		result.addError("dependencies", epic.Dependencies, "logic", "Epic cannot depend on itself")
	}

	return result
}

// ValidateStory validates a story state
func (v *Validator) ValidateStory(story *StoryState) *ValidationResult {
	result := &ValidationResult{Valid: true}

	// Validate required fields
	if story.Title == "" {
		result.addError("title", story.Title, "required", "Story title is required")
	}

	if story.EpicID == "" {
		result.addError("epic_id", story.EpicID, "required", "Story must belong to an epic")
	}

	if story.Metadata.ID == "" {
		result.addError("metadata.id", story.Metadata.ID, "required", "Story ID is required")
	}

	// Validate priority
	if !isValidPriority(story.Priority) {
		result.addError("priority", story.Priority, "enum", "Invalid story priority")
	}

	// Validate status
	if !isValidStatus(story.Status) {
		result.addError("status", story.Status, "enum", "Invalid story status")
	}

	// Validate story points
	if story.StoryPoints < 0 || story.StoryPoints > 100 {
		result.addWarning("story_points", story.StoryPoints, "range", "Story points should typically be between 1 and 21")
	}

	// Validate hours
	if story.EstimatedHours < 0 {
		result.addError("estimated_hours", story.EstimatedHours, "range", "Estimated hours cannot be negative")
	}

	if story.ActualHours < 0 {
		result.addError("actual_hours", story.ActualHours, "range", "Actual hours cannot be negative")
	}

	// Validate acceptance criteria
	for i, criteria := range story.AcceptanceCriteria {
		if criteria.Description == "" {
			result.addError(fmt.Sprintf("acceptance_criteria[%d].description", i), criteria.Description, "required", "Acceptance criteria description is required")
		}
	}

	// Check for circular dependencies
	if containsString(story.Dependencies, story.Metadata.ID) {
		result.addError("dependencies", story.Dependencies, "logic", "Story cannot depend on itself")
	}

	return result
}

// ValidateTask validates a task state
func (v *Validator) ValidateTask(task *TaskState) *ValidationResult {
	result := &ValidationResult{Valid: true}

	// Validate required fields
	if task.Title == "" {
		result.addError("title", task.Title, "required", "Task title is required")
	}

	if task.StoryID == "" {
		result.addError("story_id", task.StoryID, "required", "Task must belong to a story")
	}

	if task.Metadata.ID == "" {
		result.addError("metadata.id", task.Metadata.ID, "required", "Task ID is required")
	}

	// Validate priority
	if !isValidPriority(task.Priority) {
		result.addError("priority", task.Priority, "enum", "Invalid task priority")
	}

	// Validate status
	if !isValidStatus(task.Status) {
		result.addError("status", task.Status, "enum", "Invalid task status")
	}

	// Validate type
	if !isValidTaskType(task.Type) {
		result.addError("type", task.Type, "enum", "Invalid task type")
	}

	// Validate hours
	if task.EstimatedHours < 0 {
		result.addError("estimated_hours", task.EstimatedHours, "range", "Estimated hours cannot be negative")
	}

	if task.ActualHours < 0 {
		result.addError("actual_hours", task.ActualHours, "range", "Actual hours cannot be negative")
	}

	// Validate commands
	for i, cmd := range task.Commands {
		if cmd.Command == "" {
			result.addError(fmt.Sprintf("commands[%d].command", i), cmd.Command, "required", "Command text is required")
		}

		if cmd.Timeout < 0 || cmd.Timeout > 3600 {
			result.addError(fmt.Sprintf("commands[%d].timeout", i), cmd.Timeout, "range", "Command timeout must be between 0 and 3600 seconds")
		}

		if cmd.Retries < 0 || cmd.Retries > 10 {
			result.addError(fmt.Sprintf("commands[%d].retries", i), cmd.Retries, "range", "Command retries must be between 0 and 10")
		}
	}

	// Check for circular dependencies
	if containsString(task.Dependencies, task.Metadata.ID) {
		result.addError("dependencies", task.Dependencies, "logic", "Task cannot depend on itself")
	}

	return result
}

// ValidateStateCollection validates the entire state collection
func (v *Validator) ValidateStateCollection(state *StateCollection) *ValidationResult {
	result := &ValidationResult{Valid: true}

	// Validate projects
	for id, project := range state.Projects {
		if project.Metadata.ID != id {
			result.addError(fmt.Sprintf("projects[%s].metadata.id", id), project.Metadata.ID, "consistency", "Project ID mismatch with map key")
		}

		projectResult := v.ValidateProject(&project)
		result.merge(projectResult, fmt.Sprintf("projects[%s]", id))
	}

	// Validate epics
	for id, epic := range state.Epics {
		if epic.Metadata.ID != id {
			result.addError(fmt.Sprintf("epics[%s].metadata.id", id), epic.Metadata.ID, "consistency", "Epic ID mismatch with map key")
		}

		// Check project reference
		if _, exists := state.Projects[epic.ProjectID]; !exists {
			result.addError(fmt.Sprintf("epics[%s].project_id", id), epic.ProjectID, "reference", "Referenced project does not exist")
		}

		epicResult := v.ValidateEpic(&epic)
		result.merge(epicResult, fmt.Sprintf("epics[%s]", id))
	}

	// Validate stories
	for id, story := range state.Stories {
		if story.Metadata.ID != id {
			result.addError(fmt.Sprintf("stories[%s].metadata.id", id), story.Metadata.ID, "consistency", "Story ID mismatch with map key")
		}

		// Check epic reference
		if _, exists := state.Epics[story.EpicID]; !exists {
			result.addError(fmt.Sprintf("stories[%s].epic_id", id), story.EpicID, "reference", "Referenced epic does not exist")
		}

		storyResult := v.ValidateStory(&story)
		result.merge(storyResult, fmt.Sprintf("stories[%s]", id))
	}

	// Validate tasks
	for id, task := range state.Tasks {
		if task.Metadata.ID != id {
			result.addError(fmt.Sprintf("tasks[%s].metadata.id", id), task.Metadata.ID, "consistency", "Task ID mismatch with map key")
		}

		// Check story reference
		if _, exists := state.Stories[task.StoryID]; !exists {
			result.addError(fmt.Sprintf("tasks[%s].story_id", id), task.StoryID, "reference", "Referenced story does not exist")
		}

		taskResult := v.ValidateTask(&task)
		result.merge(taskResult, fmt.Sprintf("tasks[%s]", id))
	}

	return result
}

// Helper methods
func (r *ValidationResult) addError(field string, value interface{}, rule, message string) {
	r.Valid = false
	r.Errors = append(r.Errors, ValidationError{
		Field:    field,
		Value:    value,
		Rule:     rule,
		Message:  message,
		Severity: "error",
	})
}

func (r *ValidationResult) addWarning(field string, value interface{}, rule, message string) {
	r.Warnings = append(r.Warnings, ValidationError{
		Field:    field,
		Value:    value,
		Rule:     rule,
		Message:  message,
		Severity: "warning",
	})
}

func (r *ValidationResult) merge(other *ValidationResult, prefix string) {
	if !other.Valid {
		r.Valid = false
	}

	for _, err := range other.Errors {
		err.Field = prefix + "." + err.Field
		r.Errors = append(r.Errors, err)
	}

	for _, warn := range other.Warnings {
		warn.Field = prefix + "." + warn.Field
		r.Warnings = append(r.Warnings, warn)
	}
}

// Validation helper functions
func isValidName(name string) bool {
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]+$`, name)
	return matched && len(name) > 0 && len(name) <= 100
}

func isValidStatus(status Status) bool {
	validStatuses := []Status{StatusTodo, StatusInProgress, StatusDone, StatusBlocked, StatusCancelled}
	for _, valid := range validStatuses {
		if status == valid {
			return true
		}
	}
	return false
}

func isValidPriority(priority Priority) bool {
	validPriorities := []Priority{PriorityP0, PriorityP1, PriorityP2, PriorityP3}
	for _, valid := range validPriorities {
		if priority == valid {
			return true
		}
	}
	return false
}

func isValidTaskType(taskType TaskType) bool {
	validTypes := []TaskType{
		TaskTypeDevelopment, TaskTypeBug, TaskTypeTesting,
		TaskTypeResearch, TaskTypeDocumentation, TaskTypeDeployment, TaskTypeRefactoring,
	}
	for _, valid := range validTypes {
		if taskType == valid {
			return true
		}
	}
	return false
}

func containsString(slice []string, str string) bool {
	for _, item := range slice {
		if item == str {
			return true
		}
	}
	return false
}

// getDefaultSchema returns the default schema definition
func getDefaultSchema() *SchemaDefinition {
	return &SchemaDefinition{
		Version:   SchemaVersion,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Entities: map[string]interface{}{
			"project": ProjectState{},
			"epic":    EpicState{},
			"story":   StoryState{},
			"task":    TaskState{},
		},
		Indexes: []string{
			"metadata.id",
			"status",
			"priority",
			"assignee",
			"project_id",
			"epic_id",
			"story_id",
		},
	}
}

// getDefaultValidationRules returns default validation rules
func getDefaultValidationRules() map[string][]ValidationRule {
	return map[string][]ValidationRule{
		"project": {
			{Field: "name", Type: "required", Message: "Project name is required", Severity: "error"},
			{Field: "name", Type: "format", Value: "^[a-zA-Z0-9_-]+$", Message: "Invalid name format", Severity: "error"},
		},
		"epic": {
			{Field: "title", Type: "required", Message: "Epic title is required", Severity: "error"},
			{Field: "project_id", Type: "required", Message: "Epic must belong to a project", Severity: "error"},
		},
		"story": {
			{Field: "title", Type: "required", Message: "Story title is required", Severity: "error"},
			{Field: "epic_id", Type: "required", Message: "Story must belong to an epic", Severity: "error"},
		},
		"task": {
			{Field: "title", Type: "required", Message: "Task title is required", Severity: "error"},
			{Field: "story_id", Type: "required", Message: "Task must belong to a story", Severity: "error"},
		},
	}
}
