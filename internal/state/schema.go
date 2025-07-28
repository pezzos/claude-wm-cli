package state

import (
	"encoding/json"
	"time"
)

// SchemaVersion represents the current schema version for migration compatibility
const SchemaVersion = "1.0.0"

// Priority levels for tasks and stories
type Priority string

const (
	PriorityP0 Priority = "P0" // Critical
	PriorityP1 Priority = "P1" // High
	PriorityP2 Priority = "P2" // Medium
	PriorityP3 Priority = "P3" // Low
)

// Status represents the current state of an entity
type Status string

const (
	StatusTodo       Status = "todo"
	StatusInProgress Status = "in_progress"
	StatusDone       Status = "done"
	StatusBlocked    Status = "blocked"
	StatusCancelled  Status = "cancelled"
)

// Metadata contains common fields for all state entities
type Metadata struct {
	ID            string    `json:"id"`
	Version       string    `json:"version"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	CreatedBy     string    `json:"created_by,omitempty"`
	UpdatedBy     string    `json:"updated_by,omitempty"`
	SchemaVersion string    `json:"schema_version"`
}

// ProjectState represents the overall project configuration and status
type ProjectState struct {
	Metadata     Metadata               `json:"metadata"`
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	Repository   string                 `json:"repository,omitempty"`
	Status       Status                 `json:"status"`
	Settings     ProjectSettings        `json:"settings"`
	Metrics      ProjectMetrics         `json:"metrics"`
	Tags         []string               `json:"tags,omitempty"`
	CustomFields map[string]interface{} `json:"custom_fields,omitempty"`
}

// ProjectSettings contains project-wide configuration
type ProjectSettings struct {
	DefaultTimeout    int      `json:"default_timeout"`
	DefaultRetries    int      `json:"default_retries"`
	WorkingDirectory  string   `json:"working_directory,omitempty"`
	Environment       []string `json:"environment,omitempty"`
	NotificationEmail string   `json:"notification_email,omitempty"`
	AutoBackup        bool     `json:"auto_backup"`
	GitIntegration    bool     `json:"git_integration"`
}

// ProjectMetrics tracks project progress and performance
type ProjectMetrics struct {
	TotalEpics       int       `json:"total_epics"`
	CompletedEpics   int       `json:"completed_epics"`
	TotalStories     int       `json:"total_stories"`
	CompletedStories int       `json:"completed_stories"`
	TotalTasks       int       `json:"total_tasks"`
	CompletedTasks   int       `json:"completed_tasks"`
	ProgressPercent  float64   `json:"progress_percent"`
	LastActivity     time.Time `json:"last_activity"`
	EstimatedHours   float64   `json:"estimated_hours"`
	ActualHours      float64   `json:"actual_hours"`
}

// EpicState represents a major feature or initiative
type EpicState struct {
	Metadata       Metadata               `json:"metadata"`
	ProjectID      string                 `json:"project_id"`
	Title          string                 `json:"title"`
	Description    string                 `json:"description"`
	Priority       Priority               `json:"priority"`
	Status         Status                 `json:"status"`
	StartDate      *time.Time             `json:"start_date,omitempty"`
	EndDate        *time.Time             `json:"end_date,omitempty"`
	EstimatedHours float64                `json:"estimated_hours"`
	ActualHours    float64                `json:"actual_hours"`
	Assignee       string                 `json:"assignee,omitempty"`
	Reporter       string                 `json:"reporter,omitempty"`
	Labels         []string               `json:"labels,omitempty"`
	Dependencies   []string               `json:"dependencies,omitempty"` // Epic IDs
	Blocks         []string               `json:"blocks,omitempty"`       // Epic IDs
	Stories        []string               `json:"stories,omitempty"`      // Story IDs
	Metrics        EpicMetrics            `json:"metrics"`
	CustomFields   map[string]interface{} `json:"custom_fields,omitempty"`
}

// EpicMetrics tracks epic-specific progress
type EpicMetrics struct {
	TotalStories     int     `json:"total_stories"`
	CompletedStories int     `json:"completed_stories"`
	TotalTasks       int     `json:"total_tasks"`
	CompletedTasks   int     `json:"completed_tasks"`
	ProgressPercent  float64 `json:"progress_percent"`
	Velocity         float64 `json:"velocity"` // Stories per day
	BurndownRate     float64 `json:"burndown_rate"`
}

// StoryState represents a user story or feature requirement
type StoryState struct {
	Metadata           Metadata               `json:"metadata"`
	EpicID             string                 `json:"epic_id"`
	Title              string                 `json:"title"`
	Description        string                 `json:"description"`
	UserStory          string                 `json:"user_story,omitempty"` // As a... I want... So that...
	Priority           Priority               `json:"priority"`
	Status             Status                 `json:"status"`
	StoryPoints        int                    `json:"story_points,omitempty"`
	EstimatedHours     float64                `json:"estimated_hours"`
	ActualHours        float64                `json:"actual_hours"`
	Assignee           string                 `json:"assignee,omitempty"`
	Reporter           string                 `json:"reporter,omitempty"`
	Labels             []string               `json:"labels,omitempty"`
	Dependencies       []string               `json:"dependencies,omitempty"` // Story IDs
	Blocks             []string               `json:"blocks,omitempty"`       // Story IDs
	Tasks              []string               `json:"tasks,omitempty"`        // Task IDs
	AcceptanceCriteria []AcceptanceCriterion  `json:"acceptance_criteria,omitempty"`
	TestCases          []TestCase             `json:"test_cases,omitempty"`
	Metrics            StoryMetrics           `json:"metrics"`
	CustomFields       map[string]interface{} `json:"custom_fields,omitempty"`
}

// AcceptanceCriterion defines when a story is considered complete
type AcceptanceCriterion struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	Completed   bool   `json:"completed"`
	TestCommand string `json:"test_command,omitempty"`
}

// TestCase represents a test scenario for a story
type TestCase struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Steps       []string `json:"steps,omitempty"`
	Expected    string   `json:"expected"`
	Actual      string   `json:"actual,omitempty"`
	Status      Status   `json:"status"`
	Automated   bool     `json:"automated"`
	Command     string   `json:"command,omitempty"`
}

// StoryMetrics tracks story-specific progress
type StoryMetrics struct {
	TotalTasks        int     `json:"total_tasks"`
	CompletedTasks    int     `json:"completed_tasks"`
	ProgressPercent   float64 `json:"progress_percent"`
	CompletedCriteria int     `json:"completed_criteria"`
	TotalCriteria     int     `json:"total_criteria"`
	PassedTests       int     `json:"passed_tests"`
	TotalTests        int     `json:"total_tests"`
	CycleTime         float64 `json:"cycle_time"` // Hours from start to done
}

// TaskState represents an individual work item
type TaskState struct {
	Metadata       Metadata               `json:"metadata"`
	StoryID        string                 `json:"story_id"`
	Title          string                 `json:"title"`
	Description    string                 `json:"description"`
	Priority       Priority               `json:"priority"`
	Status         Status                 `json:"status"`
	Type           TaskType               `json:"type"`
	EstimatedHours float64                `json:"estimated_hours"`
	ActualHours    float64                `json:"actual_hours"`
	Assignee       string                 `json:"assignee,omitempty"`
	Reporter       string                 `json:"reporter,omitempty"`
	Labels         []string               `json:"labels,omitempty"`
	Dependencies   []string               `json:"dependencies,omitempty"` // Task IDs
	Blocks         []string               `json:"blocks,omitempty"`       // Task IDs
	Commands       []TaskCommand          `json:"commands,omitempty"`
	Notes          []TaskNote             `json:"notes,omitempty"`
	Attachments    []TaskAttachment       `json:"attachments,omitempty"`
	Metrics        TaskMetrics            `json:"metrics"`
	CustomFields   map[string]interface{} `json:"custom_fields,omitempty"`
}

// TaskType categorizes the type of work
type TaskType string

const (
	TaskTypeDevelopment   TaskType = "development"
	TaskTypeBug           TaskType = "bug"
	TaskTypeTesting       TaskType = "testing"
	TaskTypeResearch      TaskType = "research"
	TaskTypeDocumentation TaskType = "documentation"
	TaskTypeDeployment    TaskType = "deployment"
	TaskTypeRefactoring   TaskType = "refactoring"
)

// TaskCommand represents a command to execute for this task
type TaskCommand struct {
	ID          string            `json:"id"`
	Command     string            `json:"command"`
	Description string            `json:"description,omitempty"`
	WorkingDir  string            `json:"working_dir,omitempty"`
	Environment map[string]string `json:"environment,omitempty"`
	Timeout     int               `json:"timeout,omitempty"`
	Retries     int               `json:"retries,omitempty"`
	Status      Status            `json:"status"`
	Output      string            `json:"output,omitempty"`
	Error       string            `json:"error,omitempty"`
	ExitCode    int               `json:"exit_code,omitempty"`
	ExecutedAt  *time.Time        `json:"executed_at,omitempty"`
	Duration    float64           `json:"duration,omitempty"` // seconds
}

// TaskNote represents a note or comment on a task
type TaskNote struct {
	ID        string    `json:"id"`
	Author    string    `json:"author"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	Type      NoteType  `json:"type"`
}

// NoteType categorizes different types of notes
type NoteType string

const (
	NoteTypeComment  NoteType = "comment"
	NoteTypeProgress NoteType = "progress"
	NoteTypeBlocking NoteType = "blocking"
	NoteTypeDecision NoteType = "decision"
)

// TaskAttachment represents a file or link attached to a task
type TaskAttachment struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Path        string    `json:"path,omitempty"`
	URL         string    `json:"url,omitempty"`
	Type        string    `json:"type"`
	Size        int64     `json:"size,omitempty"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	CreatedBy   string    `json:"created_by"`
}

// TaskMetrics tracks task-specific metrics
type TaskMetrics struct {
	CommandsExecuted     int        `json:"commands_executed"`
	SuccessfulCommands   int        `json:"successful_commands"`
	FailedCommands       int        `json:"failed_commands"`
	TotalExecutionTime   float64    `json:"total_execution_time"`   // seconds
	AverageExecutionTime float64    `json:"average_execution_time"` // seconds
	LastExecutedAt       *time.Time `json:"last_executed_at,omitempty"`
}

// StateCollection represents the complete state of the system
type StateCollection struct {
	Metadata Metadata                `json:"metadata"`
	Projects map[string]ProjectState `json:"projects"`
	Epics    map[string]EpicState    `json:"epics"`
	Stories  map[string]StoryState   `json:"stories"`
	Tasks    map[string]TaskState    `json:"tasks"`
	Index    StateIndex              `json:"index"`
}

// StateIndex provides quick lookups and relationships
type StateIndex struct {
	ProjectEpics   map[string][]string `json:"project_epics"`   // project_id -> epic_ids
	EpicStories    map[string][]string `json:"epic_stories"`    // epic_id -> story_ids
	StoryTasks     map[string][]string `json:"story_tasks"`     // story_id -> task_ids
	Dependencies   map[string][]string `json:"dependencies"`    // entity_id -> dependency_ids
	Assignees      map[string][]string `json:"assignees"`       // assignee -> entity_ids
	Labels         map[string][]string `json:"labels"`          // label -> entity_ids
	StatusCounts   map[Status]int      `json:"status_counts"`   // status -> count
	PriorityCounts map[Priority]int    `json:"priority_counts"` // priority -> count
}

// ValidationRule represents a validation constraint
type ValidationRule struct {
	Field    string      `json:"field"`
	Type     string      `json:"type"` // required, format, range, enum
	Value    interface{} `json:"value,omitempty"`
	Message  string      `json:"message"`
	Severity string      `json:"severity"` // error, warning, info
}

// SchemaDefinition contains the complete schema with validation rules
type SchemaDefinition struct {
	Version    string                      `json:"version"`
	Entities   map[string]interface{}      `json:"entities"`
	Rules      map[string][]ValidationRule `json:"rules"`
	Indexes    []string                    `json:"indexes"`
	Migrations []SchemaMigration           `json:"migrations"`
	CreatedAt  time.Time                   `json:"created_at"`
	UpdatedAt  time.Time                   `json:"updated_at"`
}

// SchemaMigration represents a schema change
type SchemaMigration struct {
	FromVersion string    `json:"from_version"`
	ToVersion   string    `json:"to_version"`
	Description string    `json:"description"`
	Script      string    `json:"script"` // Migration script or instructions
	CreatedAt   time.Time `json:"created_at"`
	Reversible  bool      `json:"reversible"`
}

// Helper methods for JSON marshaling with validation
func (p *ProjectState) MarshalJSON() ([]byte, error) {
	// Update metadata before marshaling
	p.Metadata.UpdatedAt = time.Now()
	p.Metadata.SchemaVersion = SchemaVersion

	type Alias ProjectState
	return json.Marshal((*Alias)(p))
}

func (e *EpicState) MarshalJSON() ([]byte, error) {
	e.Metadata.UpdatedAt = time.Now()
	e.Metadata.SchemaVersion = SchemaVersion

	type Alias EpicState
	return json.Marshal((*Alias)(e))
}

func (s *StoryState) MarshalJSON() ([]byte, error) {
	s.Metadata.UpdatedAt = time.Now()
	s.Metadata.SchemaVersion = SchemaVersion

	type Alias StoryState
	return json.Marshal((*Alias)(s))
}

func (t *TaskState) MarshalJSON() ([]byte, error) {
	t.Metadata.UpdatedAt = time.Now()
	t.Metadata.SchemaVersion = SchemaVersion

	type Alias TaskState
	return json.Marshal((*Alias)(t))
}
