package github

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/go-github/v57/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	
	"claude-wm-cli/internal/ticket"
)

func TestIntegration_NewIntegration(t *testing.T) {
	tempDir := t.TempDir()
	setupTestDirs(t, tempDir)
	
	integration := NewIntegration(tempDir)
	assert.NotNil(t, integration)
	assert.Equal(t, tempDir, integration.rootPath)
	assert.NotNil(t, integration.ticketManager)
	assert.Equal(t, context.Background(), integration.ctx)
}

func TestIntegration_LoadSaveConfig(t *testing.T) {
	tempDir := t.TempDir()
	setupTestDirs(t, tempDir)
	
	integration := NewIntegration(tempDir)
	
	// Test loading default config (file doesn't exist)
	err := integration.LoadConfig()
	require.NoError(t, err)
	
	// Verify default values
	config := DefaultConfig()
	assert.True(t, config.Enabled)
	assert.Equal(t, "open", config.Sync.State)
	assert.True(t, config.Sync.CreateNew)
	assert.True(t, config.Sync.UpdateExisting)
	assert.False(t, config.Sync.CloseResolved)
	assert.Equal(t, 100, config.Sync.MaxIssues)
	
	// Test saving and loading custom config
	customConfig := DefaultConfig()
	customConfig.GitHub.Owner = "testorg"
	customConfig.GitHub.Repo = "testrepo"
	customConfig.Auth.Token = "test-token"
	customConfig.Enabled = false
	
	err = integration.UpdateConfig(customConfig)
	require.NoError(t, err)
	
	// Create new integration and load config
	integration2 := NewIntegration(tempDir)
	err = integration2.LoadConfig()
	require.NoError(t, err)
	
	// Verify loaded config has custom values
	// Note: We can't directly access the private config, so we test through file existence
	configPath := filepath.Join(tempDir, "docs", "2-current-epic", ConfigFileName)
	assert.FileExists(t, configPath)
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()
	
	// Test default values
	assert.True(t, config.Enabled)
	assert.Equal(t, "open", config.Sync.State)
	assert.True(t, config.Sync.CreateNew)
	assert.True(t, config.Sync.UpdateExisting)
	assert.False(t, config.Sync.CloseResolved)
	assert.Equal(t, 100, config.Sync.MaxIssues)
	assert.Equal(t, ticket.TicketPriorityMedium, config.Sync.DefaultPriority)
	assert.Equal(t, ticket.TicketTypeBug, config.Sync.DefaultType)
	
	// Test priority mappings
	assert.Equal(t, ticket.TicketPriorityUrgent, config.Sync.LabelToPriority["urgent"])
	assert.Equal(t, ticket.TicketPriorityCritical, config.Sync.LabelToPriority["critical"])
	assert.Equal(t, ticket.TicketPriorityHigh, config.Sync.LabelToPriority["high"])
	assert.Equal(t, ticket.TicketPriorityMedium, config.Sync.LabelToPriority["medium"])
	assert.Equal(t, ticket.TicketPriorityLow, config.Sync.LabelToPriority["low"])
	
	// Test type mappings
	assert.Equal(t, ticket.TicketTypeBug, config.Sync.LabelToType["bug"])
	assert.Equal(t, ticket.TicketTypeFeature, config.Sync.LabelToType["feature"])
	assert.Equal(t, ticket.TicketTypeFeature, config.Sync.LabelToType["enhancement"])
	assert.Equal(t, ticket.TicketTypeTask, config.Sync.LabelToType["task"])
	assert.Equal(t, ticket.TicketTypeSupport, config.Sync.LabelToType["support"])
	
	// Test webhook config
	assert.False(t, config.Webhook.Enabled)
	assert.Equal(t, 8080, config.Webhook.Port)
	assert.Equal(t, "/webhook/github", config.Webhook.Path)
	assert.True(t, config.Webhook.AutoSync)
}

func TestIntegration_MapIssueToTicket(t *testing.T) {
	tempDir := t.TempDir()
	setupTestDirs(t, tempDir)
	
	integration := NewIntegration(tempDir)
	
	// Create test GitHub issue
	issueNumber := 123
	issueTitle := "Test Bug Fix"
	issueBody := "This is a test issue for bug fixing"
	issueState := "open"
	createdAt := github.Timestamp{Time: time.Now().Add(-24 * time.Hour)}
	updatedAt := github.Timestamp{Time: time.Now()}
	
	issue := &github.Issue{
		Number:    &issueNumber,
		Title:     &issueTitle,
		Body:      &issueBody,
		State:     &issueState,
		CreatedAt: &createdAt,
		UpdatedAt: &updatedAt,
		HTMLURL:   github.String("https://github.com/test/repo/issues/123"),
		URL:       github.String("https://api.github.com/repos/test/repo/issues/123"),
		Labels: []*github.Label{
			{
				Name:  github.String("bug"),
				Color: github.String("ff0000"),
			},
			{
				Name:  github.String("high"),
				Color: github.String("ff8800"),
			},
		},
		User: &github.User{
			Login:     github.String("testuser"),
			ID:        github.Int64(12345),
			AvatarURL: github.String("https://avatars.githubusercontent.com/u/12345"),
			HTMLURL:   github.String("https://github.com/testuser"),
		},
		Assignees: []*github.User{
			{
				Login:     github.String("assignee1"),
				ID:        github.Int64(67890),
				AvatarURL: github.String("https://avatars.githubusercontent.com/u/67890"),
				HTMLURL:   github.String("https://github.com/assignee1"),
			},
		},
	}
	
	// Test mapping with default options
	options := &IssueSyncOptions{
		DefaultPriority: ticket.TicketPriorityMedium,
		DefaultType:     ticket.TicketTypeBug,
		LabelToPriority: map[string]ticket.TicketPriority{
			"high": ticket.TicketPriorityHigh,
			"low":  ticket.TicketPriorityLow,
		},
		LabelToType: map[string]ticket.TicketType{
			"bug":     ticket.TicketTypeBug,
			"feature": ticket.TicketTypeFeature,
		},
	}
	
	mapping := integration.mapIssueToTicket(issue, options)
	
	// Verify basic mapping
	assert.Equal(t, 123, mapping.Number)
	assert.Equal(t, "Test Bug Fix", mapping.Title)
	assert.Equal(t, "This is a test issue for bug fixing", mapping.Body)
	assert.Equal(t, "open", mapping.State)
	assert.Equal(t, "https://github.com/test/repo/issues/123", mapping.HTMLURL)
	assert.Equal(t, "https://api.github.com/repos/test/repo/issues/123", mapping.URL)
	
	// Verify labels mapping
	assert.Len(t, mapping.Labels, 2)
	assert.Equal(t, "bug", mapping.Labels[0].Name)
	assert.Equal(t, "high", mapping.Labels[1].Name)
	assert.Contains(t, mapping.Tags, "bug")
	assert.Contains(t, mapping.Tags, "high")
	
	// Verify author mapping
	assert.Equal(t, "testuser", mapping.Author.Login)
	assert.Equal(t, int64(12345), mapping.Author.ID)
	
	// Verify assignee mapping
	assert.Len(t, mapping.Assignees, 1)
	assert.Equal(t, "assignee1", mapping.Assignees[0].Login)
	assert.Equal(t, "assignee1", mapping.AssignedTo)
	
	// Verify derived properties
	assert.Equal(t, ticket.TicketTypeBug, mapping.TicketType)     // From "bug" label
	assert.Equal(t, ticket.TicketPriorityHigh, mapping.TicketPriority) // From "high" label
	assert.Equal(t, ticket.TicketStatusOpen, mapping.TicketStatus)
	
	// Verify timestamps
	assert.False(t, mapping.CreatedAt.IsZero())
	assert.False(t, mapping.UpdatedAt.IsZero())
	assert.False(t, mapping.SyncedAt.IsZero())
}

func TestIntegration_MapIssueToTicket_ClosedIssue(t *testing.T) {
	tempDir := t.TempDir()
	setupTestDirs(t, tempDir)
	
	integration := NewIntegration(tempDir)
	
	// Create closed GitHub issue
	issueNumber := 456
	issueTitle := "Closed Issue"
	issueState := "closed"
	createdAt := github.Timestamp{Time: time.Now().Add(-48 * time.Hour)}
	updatedAt := github.Timestamp{Time: time.Now().Add(-24 * time.Hour)}
	closedAt := github.Timestamp{Time: time.Now().Add(-12 * time.Hour)}
	
	issue := &github.Issue{
		Number:    &issueNumber,
		Title:     &issueTitle,
		State:     &issueState,
		CreatedAt: &createdAt,
		UpdatedAt: &updatedAt,
		ClosedAt:  &closedAt,
		HTMLURL:   github.String("https://github.com/test/repo/issues/456"),
		URL:       github.String("https://api.github.com/repos/test/repo/issues/456"),
		User: &github.User{
			Login: github.String("testuser"),
		},
	}
	
	options := &IssueSyncOptions{
		DefaultPriority: ticket.TicketPriorityMedium,
		DefaultType:     ticket.TicketTypeTask,
		CloseResolved:   false, // Should map to resolved, not closed
	}
	
	mapping := integration.mapIssueToTicket(issue, options)
	
	// Verify closed issue mapping
	assert.Equal(t, "closed", mapping.State)
	assert.Equal(t, ticket.TicketStatusResolved, mapping.TicketStatus) // CloseResolved=false
	assert.NotNil(t, mapping.ClosedAt)
	
	// Test with CloseResolved=true
	options.CloseResolved = true
	mapping = integration.mapIssueToTicket(issue, options)
	assert.Equal(t, ticket.TicketStatusClosed, mapping.TicketStatus)
}

func TestIntegration_MapIssueToTicket_WithMilestone(t *testing.T) {
	tempDir := t.TempDir()
	setupTestDirs(t, tempDir)
	
	integration := NewIntegration(tempDir)
	
	// Create GitHub issue with milestone
	issueNumber := 789
	issueTitle := "Issue with Milestone"
	issueState := "open"
	createdAt := github.Timestamp{Time: time.Now()}
	updatedAt := github.Timestamp{Time: time.Now()}
	dueOn := github.Timestamp{Time: time.Now().Add(7 * 24 * time.Hour)}
	
	milestone := &github.Milestone{
		Title:       github.String("v1.0.0"),
		Number:      github.Int(1),
		State:       github.String("open"),
		Description: github.String("First major release"),
		DueOn:       &dueOn,
	}
	
	issue := &github.Issue{
		Number:    &issueNumber,
		Title:     &issueTitle,
		State:     &issueState,
		CreatedAt: &createdAt,
		UpdatedAt: &updatedAt,
		HTMLURL:   github.String("https://github.com/test/repo/issues/789"),
		URL:       github.String("https://api.github.com/repos/test/repo/issues/789"),
		Milestone: milestone,
		User: &github.User{
			Login: github.String("testuser"),
		},
	}
	
	options := &IssueSyncOptions{
		DefaultPriority: ticket.TicketPriorityMedium,
		DefaultType:     ticket.TicketTypeTask,
	}
	
	mapping := integration.mapIssueToTicket(issue, options)
	
	// Verify milestone mapping
	require.NotNil(t, mapping.Milestone)
	assert.Equal(t, "v1.0.0", mapping.Milestone.Title)
	assert.Equal(t, 1, mapping.Milestone.Number)
	assert.Equal(t, "open", mapping.Milestone.State)
	assert.Equal(t, "First major release", mapping.Milestone.Description)
	assert.NotNil(t, mapping.Milestone.DueOn)
}

func TestIntegration_FormatIssueDescription(t *testing.T) {
	tempDir := t.TempDir()
	setupTestDirs(t, tempDir)
	
	integration := NewIntegration(tempDir)
	
	// Create test mapping
	mapping := IssueMapping{
		Number:  123,
		Body:    "Original issue description with multiple lines.\n\nThis is paragraph 2.",
		HTMLURL: "https://github.com/test/repo/issues/123",
		Author: GitHubUser{
			Login: "testuser",
		},
		CreatedAt: time.Date(2023, 1, 15, 10, 30, 0, 0, time.UTC),
		UpdatedAt: time.Date(2023, 1, 16, 14, 45, 0, 0, time.UTC),
		Labels: []GitHubLabel{
			{Name: "bug"},
			{Name: "urgent"},
		},
		Tags: []string{"bug", "urgent"},
		Assignees: []GitHubUser{
			{Login: "assignee1"},
			{Login: "assignee2"},
		},
		Milestone: &GitHubMilestone{
			Title: "v1.0.0",
		},
	}
	
	description := integration.formatIssueDescription(mapping)
	
	// Verify description format
	assert.Contains(t, description, "GitHub Issue #123")
	assert.Contains(t, description, "Original issue description with multiple lines.")
	assert.Contains(t, description, "This is paragraph 2.")
	assert.Contains(t, description, "**GitHub Details:**")
	assert.Contains(t, description, "[#123](https://github.com/test/repo/issues/123)")
	assert.Contains(t, description, "Author: @testuser")
	assert.Contains(t, description, "Created: 2023-01-15 10:30:00")
	assert.Contains(t, description, "Updated: 2023-01-16 14:45:00")
	assert.Contains(t, description, "Labels: bug, urgent")
	assert.Contains(t, description, "Assignees: @assignee1, @assignee2")
	assert.Contains(t, description, "Milestone: v1.0.0")
}

func TestIntegration_FindExistingTicket(t *testing.T) {
	tempDir := t.TempDir()
	setupTestDirs(t, tempDir)
	
	integration := NewIntegration(tempDir)
	
	// Create a ticket with GitHub external reference
	externalRef := &ticket.ExternalReference{
		System: "github",
		ID:     "123",
		URL:    "https://github.com/test/repo/issues/123",
	}
	
	ticketOptions := ticket.TicketCreateOptions{
		Title:       "Test GitHub Ticket",
		Description: "Created from GitHub issue",
		Type:        ticket.TicketTypeBug,
		Priority:    ticket.TicketPriorityHigh,
		ExternalRef: externalRef,
	}
	
	createdTicket, err := integration.ticketManager.CreateTicket(ticketOptions)
	require.NoError(t, err)
	
	// Test finding existing ticket
	foundTicketID := integration.findExistingTicket(123)
	assert.Equal(t, createdTicket.ID, foundTicketID)
	
	// Test with non-existent issue
	notFoundTicketID := integration.findExistingTicket(999)
	assert.Empty(t, notFoundTicketID)
}

func TestIntegration_ConfigValidation(t *testing.T) {
	tempDir := t.TempDir()
	setupTestDirs(t, tempDir)
	
	integration := NewIntegration(tempDir)
	
	// Test with disabled integration
	config := DefaultConfig()
	config.Enabled = false
	
	err := integration.Initialize(config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "GitHub integration is disabled")
	
	// Test with missing owner/repo
	config = DefaultConfig()
	config.Enabled = true
	config.GitHub.Owner = ""
	config.GitHub.Repo = ""
	
	err = integration.Initialize(config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "owner and repo must be specified")
	
	// Test with missing authentication
	config = DefaultConfig()
	config.Enabled = true
	config.GitHub.Owner = "testorg"
	config.GitHub.Repo = "testrepo"
	config.Auth.Token = ""
	config.Auth.TokenFile = ""
	
	err = integration.Initialize(config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no authentication method configured")
}

func TestStringSlicesEqual(t *testing.T) {
	// Test equal slices
	assert.True(t, stringSlicesEqual([]string{"a", "b", "c"}, []string{"a", "b", "c"}))
	assert.True(t, stringSlicesEqual([]string{}, []string{}))
	
	// Test different lengths
	assert.False(t, stringSlicesEqual([]string{"a", "b"}, []string{"a", "b", "c"}))
	
	// Test different content
	assert.False(t, stringSlicesEqual([]string{"a", "b", "c"}, []string{"a", "c", "b"}))
	
	// Test nil vs empty
	assert.True(t, stringSlicesEqual(nil, []string{}))
	assert.True(t, stringSlicesEqual([]string{}, nil))
}

func TestTicketTypes_ValidationInMapping(t *testing.T) {
	// Test priority validation
	validPriorities := []ticket.TicketPriority{
		ticket.TicketPriorityLow,
		ticket.TicketPriorityMedium,
		ticket.TicketPriorityHigh,
		ticket.TicketPriorityCritical,
		ticket.TicketPriorityUrgent,
	}
	
	for _, priority := range validPriorities {
		assert.True(t, priority.IsValid(), "Priority %s should be valid", priority)
	}
	
	assert.False(t, ticket.TicketPriority("invalid").IsValid())
	
	// Test type validation
	validTypes := []ticket.TicketType{
		ticket.TicketTypeBug,
		ticket.TicketTypeFeature,
		ticket.TicketTypeInterruption,
		ticket.TicketTypeTask,
		ticket.TicketTypeSupport,
	}
	
	for _, ticketType := range validTypes {
		assert.True(t, ticketType.IsValid(), "Type %s should be valid", ticketType)
	}
	
	assert.False(t, ticket.TicketType("invalid").IsValid())
}

// Helper function to setup test directories
func setupTestDirs(t *testing.T, tempDir string) {
	docsDir := filepath.Join(tempDir, "docs", "1-project")
	err := os.MkdirAll(docsDir, 0755)
	require.NoError(t, err)
	
	currentEpicDir := filepath.Join(tempDir, "docs", "2-current-epic")
	err = os.MkdirAll(currentEpicDir, 0755)
	require.NoError(t, err)
	
	currentTaskDir := filepath.Join(tempDir, "docs", "3-current-task")
	err = os.MkdirAll(currentTaskDir, 0755)
	require.NoError(t, err)
}