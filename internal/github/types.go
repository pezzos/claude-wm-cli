package github

import (
	"time"

	"claude-wm-cli/internal/ticket"
)

// GitHubConfig contains GitHub API configuration
type GitHubConfig struct {
	Token     string `json:"token,omitempty"`
	Owner     string `json:"owner"`                // Repository owner/organization
	Repo      string `json:"repo"`                 // Repository name
	BaseURL   string `json:"base_url,omitempty"`   // For GitHub Enterprise
	UploadURL string `json:"upload_url,omitempty"` // For GitHub Enterprise
}

// IssueSyncOptions configures how GitHub issues are synchronized
type IssueSyncOptions struct {
	// Filter options
	State     string     `json:"state,omitempty"`     // open, closed, all
	Labels    []string   `json:"labels,omitempty"`    // Filter by labels
	Assignee  string     `json:"assignee,omitempty"`  // Filter by assignee
	Creator   string     `json:"creator,omitempty"`   // Filter by creator
	Mentioned string     `json:"mentioned,omitempty"` // Filter by mentioned user
	Milestone string     `json:"milestone,omitempty"` // Filter by milestone
	Since     *time.Time `json:"since,omitempty"`     // Only issues updated after this time

	// Sync behavior
	CreateNew      bool `json:"create_new"`           // Create tickets for new issues
	UpdateExisting bool `json:"update_existing"`      // Update existing tickets
	CloseResolved  bool `json:"close_resolved"`       // Close tickets for closed issues
	MaxIssues      int  `json:"max_issues,omitempty"` // Limit number of issues to fetch

	// Mapping options
	LabelToPriority map[string]ticket.TicketPriority `json:"label_to_priority,omitempty"`
	LabelToType     map[string]ticket.TicketType     `json:"label_to_type,omitempty"`
	DefaultPriority ticket.TicketPriority            `json:"default_priority,omitempty"`
	DefaultType     ticket.TicketType                `json:"default_type,omitempty"`
}

// IssueSyncResult contains the results of a sync operation
type IssueSyncResult struct {
	TotalIssues     int              `json:"total_issues"`
	CreatedTickets  int              `json:"created_tickets"`
	UpdatedTickets  int              `json:"updated_tickets"`
	SkippedIssues   int              `json:"skipped_issues"`
	ErrorCount      int              `json:"error_count"`
	ProcessedIssues []ProcessedIssue `json:"processed_issues"`
	Errors          []string         `json:"errors,omitempty"`
	SyncTime        time.Time        `json:"sync_time"`
	RateLimitInfo   *RateLimitInfo   `json:"rate_limit_info,omitempty"`
}

// ProcessedIssue represents an issue that was processed during sync
type ProcessedIssue struct {
	IssueNumber int    `json:"issue_number"`
	IssueURL    string `json:"issue_url"`
	TicketID    string `json:"ticket_id,omitempty"`
	Action      string `json:"action"` // created, updated, skipped, error
	Reason      string `json:"reason,omitempty"`
	Error       string `json:"error,omitempty"`
}

// RateLimitInfo contains GitHub API rate limit information
type RateLimitInfo struct {
	Limit     int       `json:"limit"`
	Remaining int       `json:"remaining"`
	ResetTime time.Time `json:"reset_time"`
}

// IssueMapping defines how GitHub issues map to tickets
type IssueMapping struct {
	// GitHub issue data
	Number    int              `json:"number"`
	Title     string           `json:"title"`
	Body      string           `json:"body"`
	State     string           `json:"state"` // open, closed
	Labels    []GitHubLabel    `json:"labels"`
	Assignees []GitHubUser     `json:"assignees"`
	Author    GitHubUser       `json:"author"`
	Milestone *GitHubMilestone `json:"milestone,omitempty"`
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`
	ClosedAt  *time.Time       `json:"closed_at,omitempty"`
	URL       string           `json:"url"`
	HTMLURL   string           `json:"html_url"`

	// Derived ticket properties
	TicketType     ticket.TicketType     `json:"ticket_type"`
	TicketPriority ticket.TicketPriority `json:"ticket_priority"`
	TicketStatus   ticket.TicketStatus   `json:"ticket_status"`
	AssignedTo     string                `json:"assigned_to,omitempty"`
	Tags           []string              `json:"tags"`

	// Metadata
	SyncedAt     time.Time  `json:"synced_at"`
	LastSyncedAt *time.Time `json:"last_synced_at,omitempty"`
}

// GitHubLabel represents a GitHub issue label
type GitHubLabel struct {
	Name        string `json:"name"`
	Color       string `json:"color"`
	Description string `json:"description,omitempty"`
}

// GitHubUser represents a GitHub user
type GitHubUser struct {
	Login     string `json:"login"`
	ID        int64  `json:"id"`
	AvatarURL string `json:"avatar_url,omitempty"`
	HTMLURL   string `json:"html_url,omitempty"`
}

// GitHubMilestone represents a GitHub milestone
type GitHubMilestone struct {
	Title       string     `json:"title"`
	Number      int        `json:"number"`
	State       string     `json:"state"`
	Description string     `json:"description,omitempty"`
	DueOn       *time.Time `json:"due_on,omitempty"`
}

// AuthConfig contains GitHub authentication configuration
type AuthConfig struct {
	Token          string `json:"token,omitempty"`
	TokenFile      string `json:"token_file,omitempty"`
	AppID          int64  `json:"app_id,omitempty"`
	InstallationID int64  `json:"installation_id,omitempty"`
	PrivateKeyPath string `json:"private_key_path,omitempty"`
}

// WebhookConfig contains GitHub webhook configuration for real-time sync
type WebhookConfig struct {
	Enabled  bool   `json:"enabled"`
	Secret   string `json:"secret,omitempty"`
	Port     int    `json:"port,omitempty"`
	Path     string `json:"path,omitempty"`
	AutoSync bool   `json:"auto_sync"` // Automatically sync when webhook received
}

// SyncHistory tracks GitHub sync operations
type SyncHistory struct {
	LastSync           time.Time         `json:"last_sync"`
	LastSuccessfulSync time.Time         `json:"last_successful_sync"`
	TotalSyncs         int               `json:"total_syncs"`
	SuccessfulSyncs    int               `json:"successful_syncs"`
	FailedSyncs        int               `json:"failed_syncs"`
	LastErrors         []string          `json:"last_errors,omitempty"`
	RecentResults      []IssueSyncResult `json:"recent_results,omitempty"`
}

// GitHubIntegrationConfig contains the complete configuration
type GitHubIntegrationConfig struct {
	GitHub  GitHubConfig     `json:"github"`
	Auth    AuthConfig       `json:"auth"`
	Sync    IssueSyncOptions `json:"sync"`
	Webhook WebhookConfig    `json:"webhook"`
	History SyncHistory      `json:"history"`
	Enabled bool             `json:"enabled"`
}

// DefaultConfig returns a default GitHub integration configuration
func DefaultConfig() GitHubIntegrationConfig {
	return GitHubIntegrationConfig{
		Sync: IssueSyncOptions{
			State:           "open",
			CreateNew:       true,
			UpdateExisting:  true,
			CloseResolved:   false,
			MaxIssues:       100,
			DefaultPriority: ticket.TicketPriorityMedium,
			DefaultType:     ticket.TicketTypeBug,
			LabelToPriority: map[string]ticket.TicketPriority{
				"urgent":            ticket.TicketPriorityUrgent,
				"critical":          ticket.TicketPriorityCritical,
				"high":              ticket.TicketPriorityHigh,
				"medium":            ticket.TicketPriorityMedium,
				"low":               ticket.TicketPriorityLow,
				"priority/urgent":   ticket.TicketPriorityUrgent,
				"priority/critical": ticket.TicketPriorityCritical,
				"priority/high":     ticket.TicketPriorityHigh,
				"priority/medium":   ticket.TicketPriorityMedium,
				"priority/low":      ticket.TicketPriorityLow,
			},
			LabelToType: map[string]ticket.TicketType{
				"bug":              ticket.TicketTypeBug,
				"feature":          ticket.TicketTypeFeature,
				"enhancement":      ticket.TicketTypeFeature,
				"task":             ticket.TicketTypeTask,
				"support":          ticket.TicketTypeSupport,
				"help":             ticket.TicketTypeSupport,
				"type/bug":         ticket.TicketTypeBug,
				"type/feature":     ticket.TicketTypeFeature,
				"type/enhancement": ticket.TicketTypeFeature,
				"type/task":        ticket.TicketTypeTask,
				"type/support":     ticket.TicketTypeSupport,
			},
		},
		Webhook: WebhookConfig{
			Enabled:  false,
			Port:     8080,
			Path:     "/webhook/github",
			AutoSync: true,
		},
		Enabled: true,
	}
}
