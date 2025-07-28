package ticket

import (
	"time"
)

// TicketStatus represents the current state of a ticket
type TicketStatus string

const (
	TicketStatusOpen       TicketStatus = "open"
	TicketStatusInProgress TicketStatus = "in_progress"
	TicketStatusResolved   TicketStatus = "resolved"
	TicketStatusClosed     TicketStatus = "closed"
)

// IsValid checks if the ticket status is valid
func (ts TicketStatus) IsValid() bool {
	switch ts {
	case TicketStatusOpen, TicketStatusInProgress, TicketStatusResolved, TicketStatusClosed:
		return true
	default:
		return false
	}
}

// TicketPriority represents the urgency level of a ticket
type TicketPriority string

const (
	TicketPriorityLow      TicketPriority = "low"
	TicketPriorityMedium   TicketPriority = "medium"
	TicketPriorityHigh     TicketPriority = "high"
	TicketPriorityCritical TicketPriority = "critical"
	TicketPriorityUrgent   TicketPriority = "urgent"
)

// IsValid checks if the ticket priority is valid
func (tp TicketPriority) IsValid() bool {
	switch tp {
	case TicketPriorityLow, TicketPriorityMedium, TicketPriorityHigh, TicketPriorityCritical, TicketPriorityUrgent:
		return true
	default:
		return false
	}
}

// TicketType categorizes the nature of the ticket
type TicketType string

const (
	TicketTypeBug          TicketType = "bug"
	TicketTypeFeature      TicketType = "feature"
	TicketTypeInterruption TicketType = "interruption"
	TicketTypeTask         TicketType = "task"
	TicketTypeSupport      TicketType = "support"
)

// IsValid checks if the ticket type is valid
func (tt TicketType) IsValid() bool {
	switch tt {
	case TicketTypeBug, TicketTypeFeature, TicketTypeInterruption, TicketTypeTask, TicketTypeSupport:
		return true
	default:
		return false
	}
}

// Ticket represents an interruption or urgent task
type Ticket struct {
	ID          string         `json:"id"`
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Type        TicketType     `json:"type"`
	Status      TicketStatus   `json:"status"`
	Priority    TicketPriority `json:"priority"`

	// Workflow integration
	RelatedEpicID  string `json:"related_epic_id,omitempty"`
	RelatedStoryID string `json:"related_story_id,omitempty"`

	// Interruption context
	InterruptedTask string                 `json:"interrupted_task,omitempty"`
	WorkflowContext map[string]interface{} `json:"workflow_context,omitempty"`

	// Assignment and tracking
	AssignedTo  string           `json:"assigned_to,omitempty"`
	Estimations TicketEstimation `json:"estimations"`
	Tags        []string         `json:"tags,omitempty"`

	// External references
	ExternalRef *ExternalReference `json:"external_ref,omitempty"`

	// Timestamps
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	StartedAt  *time.Time `json:"started_at,omitempty"`
	ResolvedAt *time.Time `json:"resolved_at,omitempty"`
	ClosedAt   *time.Time `json:"closed_at,omitempty"`
	DueDate    *time.Time `json:"due_date,omitempty"`
}

// TicketEstimation contains time and effort estimates
type TicketEstimation struct {
	EstimatedHours float64 `json:"estimated_hours,omitempty"`
	ActualHours    float64 `json:"actual_hours,omitempty"`
	StoryPoints    int     `json:"story_points,omitempty"`
	Complexity     string  `json:"complexity,omitempty"` // simple, medium, complex
}

// ExternalReference links tickets to external systems
type ExternalReference struct {
	System   string                 `json:"system"` // "github", "jira", "linear", etc.
	ID       string                 `json:"id"`     // External ID
	URL      string                 `json:"url,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// TicketCollection manages all tickets in the project
type TicketCollection struct {
	Tickets       map[string]*Ticket `json:"tickets"`
	CurrentTicket string             `json:"current_ticket,omitempty"`
	WorkflowState *WorkflowState     `json:"workflow_state,omitempty"`
	Metadata      TicketMetadata     `json:"metadata"`
}

// WorkflowState preserves context during interruptions
type WorkflowState struct {
	PreInterruption  *WorkflowContext `json:"pre_interruption,omitempty"`
	CurrentContext   *WorkflowContext `json:"current_context,omitempty"`
	InterruptionTime time.Time        `json:"interruption_time"`
	CanResume        bool             `json:"can_resume"`
}

// WorkflowContext captures the current state of work
type WorkflowContext struct {
	CurrentEpic      string    `json:"current_epic,omitempty"`
	CurrentStory     string    `json:"current_story,omitempty"`
	CurrentTask      string    `json:"current_task,omitempty"`
	WorkingDirectory string    `json:"working_directory"`
	GitBranch        string    `json:"git_branch,omitempty"`
	GitCommit        string    `json:"git_commit,omitempty"`
	ActiveFiles      []string  `json:"active_files,omitempty"`
	Notes            string    `json:"notes,omitempty"`
	Timestamp        time.Time `json:"timestamp"`
}

// TicketMetadata contains collection-level information
type TicketMetadata struct {
	Version         string    `json:"version"`
	LastUpdated     time.Time `json:"last_updated"`
	TotalTickets    int       `json:"total_tickets"`
	OpenTickets     int       `json:"open_tickets"`
	ResolvedTickets int       `json:"resolved_tickets"`
}

// TicketCreateOptions contains parameters for creating a new ticket
type TicketCreateOptions struct {
	Title          string
	Description    string
	Type           TicketType
	Priority       TicketPriority
	RelatedEpicID  string
	RelatedStoryID string
	AssignedTo     string
	EstimatedHours float64
	StoryPoints    int
	Tags           []string
	DueDate        *time.Time
	ExternalRef    *ExternalReference
}

// TicketUpdateOptions contains parameters for updating an existing ticket
type TicketUpdateOptions struct {
	Title          *string
	Description    *string
	Type           *TicketType
	Status         *TicketStatus
	Priority       *TicketPriority
	RelatedEpicID  *string
	RelatedStoryID *string
	AssignedTo     *string
	EstimatedHours *float64
	ActualHours    *float64
	StoryPoints    *int
	Tags           *[]string
	DueDate        *time.Time
	ExternalRef    *ExternalReference
}

// TicketListOptions contains parameters for filtering tickets
type TicketListOptions struct {
	Status         TicketStatus
	Priority       TicketPriority
	Type           TicketType
	AssignedTo     string
	RelatedEpicID  string
	RelatedStoryID string
	ShowClosed     bool
	Limit          int
}

// TicketStats provides analytics on ticket collection
type TicketStats struct {
	TotalTickets          int                    `json:"total_tickets"`
	ByStatus              map[TicketStatus]int   `json:"by_status"`
	ByPriority            map[TicketPriority]int `json:"by_priority"`
	ByType                map[TicketType]int     `json:"by_type"`
	AverageResolutionTime time.Duration          `json:"avg_resolution_time"`
	OldestOpenTicket      *time.Time             `json:"oldest_open_ticket,omitempty"`
	RecentActivity        []TicketActivity       `json:"recent_activity"`
}

// TicketActivity represents a change in ticket state
type TicketActivity struct {
	TicketID  string      `json:"ticket_id"`
	Action    string      `json:"action"` // created, updated, status_changed, etc.
	OldValue  interface{} `json:"old_value,omitempty"`
	NewValue  interface{} `json:"new_value,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
	User      string      `json:"user,omitempty"`
}
