package github

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-github/v57/github"
	"golang.org/x/oauth2"
	
	"claude-wm-cli/internal/ticket"
)

const (
	ConfigFileName = "github-integration.json"
	ConfigVersion  = "1.0.0"
)

// Integration manages GitHub issue synchronization
type Integration struct {
	rootPath      string
	config        GitHubIntegrationConfig
	client        *github.Client
	ticketManager *ticket.Manager
	ctx           context.Context
}

// NewIntegration creates a new GitHub integration
func NewIntegration(rootPath string) *Integration {
	return &Integration{
		rootPath:      rootPath,
		ticketManager: ticket.NewManager(rootPath),
		ctx:           context.Background(),
	}
}

// Initialize sets up the GitHub integration with authentication
func (gi *Integration) Initialize(config GitHubIntegrationConfig) error {
	gi.config = config
	
	if !gi.config.Enabled {
		return fmt.Errorf("GitHub integration is disabled")
	}
	
	// Validate required configuration
	if gi.config.GitHub.Owner == "" || gi.config.GitHub.Repo == "" {
		return fmt.Errorf("GitHub owner and repo must be specified")
	}
	
	// Setup authentication
	if err := gi.setupAuth(); err != nil {
		return fmt.Errorf("failed to setup GitHub authentication: %w", err)
	}
	
	// Test connection
	if err := gi.testConnection(); err != nil {
		return fmt.Errorf("failed to connect to GitHub: %w", err)
	}
	
	return nil
}

// LoadConfig loads GitHub integration configuration from file
func (gi *Integration) LoadConfig() error {
	configPath := filepath.Join(gi.rootPath, "docs", "2-current-epic", ConfigFileName)
	
	// Check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Create default configuration
		gi.config = DefaultConfig()
		return gi.SaveConfig()
	}
	
	// Read configuration file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}
	
	if err := json.Unmarshal(data, &gi.config); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}
	
	return nil
}

// SaveConfig saves the current configuration to file
func (gi *Integration) SaveConfig() error {
	configPath := filepath.Join(gi.rootPath, "docs", "2-current-epic", ConfigFileName)
	
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	
	// Marshal configuration
	data, err := json.MarshalIndent(gi.config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	
	// Write file atomically
	tempPath := configPath + ".tmp"
	if err := os.WriteFile(tempPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write temp config file: %w", err)
	}
	
	if err := os.Rename(tempPath, configPath); err != nil {
		os.Remove(tempPath)
		return fmt.Errorf("failed to replace config file: %w", err)
	}
	
	return nil
}

// SyncIssues fetches GitHub issues and creates/updates tickets
func (gi *Integration) SyncIssues(options *IssueSyncOptions) (*IssueSyncResult, error) {
	if gi.client == nil {
		return nil, fmt.Errorf("GitHub client not initialized")
	}
	
	// Use provided options or default config
	syncOptions := &gi.config.Sync
	if options != nil {
		syncOptions = options
	}
	
	result := &IssueSyncResult{
		SyncTime:        time.Now(),
		ProcessedIssues: make([]ProcessedIssue, 0),
		Errors:          make([]string, 0),
	}
	
	// Fetch GitHub issues
	issues, err := gi.fetchIssues(syncOptions)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to fetch issues: %v", err))
		result.ErrorCount++
		return result, err
	}
	
	result.TotalIssues = len(issues)
	
	// Process each issue
	for _, issue := range issues {
		if len(result.ProcessedIssues) >= syncOptions.MaxIssues && syncOptions.MaxIssues > 0 {
			break
		}
		
		processed := gi.processIssue(issue, syncOptions)
		result.ProcessedIssues = append(result.ProcessedIssues, processed)
		
		switch processed.Action {
		case "created":
			result.CreatedTickets++
		case "updated":
			result.UpdatedTickets++
		case "skipped":
			result.SkippedIssues++
		case "error":
			result.ErrorCount++
			result.Errors = append(result.Errors, processed.Error)
		}
	}
	
	// Get rate limit info
	rateLimit, _, err := gi.client.RateLimits(gi.ctx)
	if err == nil && rateLimit != nil && rateLimit.Core != nil {
		result.RateLimitInfo = &RateLimitInfo{
			Limit:     rateLimit.Core.Limit,
			Remaining: rateLimit.Core.Remaining,
			ResetTime: rateLimit.Core.Reset.Time,
		}
	}
	
	// Update sync history
	gi.updateSyncHistory(result)
	
	return result, nil
}

// GetIssueByNumber fetches a specific GitHub issue and creates/updates a ticket
func (gi *Integration) GetIssueByNumber(issueNumber int) (*ProcessedIssue, error) {
	if gi.client == nil {
		return nil, fmt.Errorf("GitHub client not initialized")
	}
	
	// Fetch the specific issue
	issue, _, err := gi.client.Issues.Get(gi.ctx, gi.config.GitHub.Owner, gi.config.GitHub.Repo, issueNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch issue #%d: %w", issueNumber, err)
	}
	
	// Process the issue
	processed := gi.processIssue(issue, &gi.config.Sync)
	return &processed, nil
}

// GetSyncHistory returns the synchronization history
func (gi *Integration) GetSyncHistory() SyncHistory {
	return gi.config.History
}

// UpdateConfig updates the integration configuration
func (gi *Integration) UpdateConfig(config GitHubIntegrationConfig) error {
	gi.config = config
	return gi.SaveConfig()
}

// Private methods

func (gi *Integration) setupAuth() error {
	var tc *oauth2.TokenSource
	
	// Use token authentication
	if gi.config.Auth.Token != "" {
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: gi.config.Auth.Token},
		)
		tc = &ts
	} else if gi.config.Auth.TokenFile != "" {
		// Read token from file
		tokenData, err := os.ReadFile(gi.config.Auth.TokenFile)
		if err != nil {
			return fmt.Errorf("failed to read token file: %w", err)
		}
		token := strings.TrimSpace(string(tokenData))
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: token},
		)
		tc = &ts
	} else {
		return fmt.Errorf("no authentication method configured")
	}
	
	var httpClient = oauth2.NewClient(gi.ctx, *tc)
	
	// Create GitHub client
	gi.client = github.NewClient(httpClient)
	
	// Set custom base URL for GitHub Enterprise
	if gi.config.GitHub.BaseURL != "" {
		var err error
		gi.client, err = gi.client.WithEnterpriseURLs(gi.config.GitHub.BaseURL, gi.config.GitHub.UploadURL)
		if err != nil {
			return fmt.Errorf("failed to set custom GitHub URLs: %w", err)
		}
	}
	
	return nil
}

func (gi *Integration) testConnection() error {
	// Test connection by getting repository information
	_, _, err := gi.client.Repositories.Get(gi.ctx, gi.config.GitHub.Owner, gi.config.GitHub.Repo)
	if err != nil {
		return fmt.Errorf("failed to access repository %s/%s: %w", 
			gi.config.GitHub.Owner, gi.config.GitHub.Repo, err)
	}
	
	return nil
}

func (gi *Integration) fetchIssues(options *IssueSyncOptions) ([]*github.Issue, error) {
	// Build GitHub issue list options
	listOptions := &github.IssueListByRepoOptions{
		State:     options.State,
		Labels:    options.Labels,
		Assignee:  options.Assignee,
		Creator:   options.Creator,
		Mentioned: options.Mentioned,
		Milestone: options.Milestone,
		ListOptions: github.ListOptions{
			PerPage: 100, // Maximum allowed by GitHub API
		},
	}
	
	var allIssues []*github.Issue
	
	// Fetch all pages
	for {
		issues, resp, err := gi.client.Issues.ListByRepo(gi.ctx, gi.config.GitHub.Owner, gi.config.GitHub.Repo, listOptions)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch issues page %d: %w", listOptions.Page, err)
		}
		
		// Filter out pull requests (GitHub API includes them in issues)
		for _, issue := range issues {
			if issue.PullRequestLinks == nil {
				allIssues = append(allIssues, issue)
			}
		}
		
		// Check if we have more pages
		if resp.NextPage == 0 {
			break
		}
		listOptions.Page = resp.NextPage
		
		// Respect rate limits
		if resp.Rate.Remaining < 10 {
			sleepTime := time.Until(resp.Rate.Reset.Time)
			if sleepTime > 0 {
				time.Sleep(sleepTime)
			}
		}
		
		// Limit total issues if specified
		if options.MaxIssues > 0 && len(allIssues) >= options.MaxIssues {
			break
		}
	}
	
	// Apply max issues limit
	if options.MaxIssues > 0 && len(allIssues) > options.MaxIssues {
		allIssues = allIssues[:options.MaxIssues]
	}
	
	return allIssues, nil
}

func (gi *Integration) processIssue(issue *github.Issue, options *IssueSyncOptions) ProcessedIssue {
	processed := ProcessedIssue{
		IssueNumber: issue.GetNumber(),
		IssueURL:    issue.GetHTMLURL(),
	}
	
	// Map GitHub issue to ticket
	mapping := gi.mapIssueToTicket(issue, options)
	
	// Check if ticket already exists
	existingTicketID := gi.findExistingTicket(issue.GetNumber())
	
	if existingTicketID != "" {
		// Update existing ticket
		if options.UpdateExisting {
			err := gi.updateExistingTicket(existingTicketID, mapping, options)
			if err != nil {
				processed.Action = "error"
				processed.Error = err.Error()
			} else {
				processed.Action = "updated"
				processed.TicketID = existingTicketID
			}
		} else {
			processed.Action = "skipped"
			processed.Reason = "existing ticket found, updates disabled"
			processed.TicketID = existingTicketID
		}
	} else {
		// Create new ticket
		if options.CreateNew {
			ticketID, err := gi.createNewTicket(mapping)
			if err != nil {
				processed.Action = "error"
				processed.Error = err.Error()
			} else {
				processed.Action = "created"
				processed.TicketID = ticketID
			}
		} else {
			processed.Action = "skipped"
			processed.Reason = "new ticket creation disabled"
		}
	}
	
	return processed
}

func (gi *Integration) mapIssueToTicket(issue *github.Issue, options *IssueSyncOptions) IssueMapping {
	mapping := IssueMapping{
		Number:    issue.GetNumber(),
		Title:     issue.GetTitle(),
		Body:      issue.GetBody(),
		State:     issue.GetState(),
		CreatedAt: issue.GetCreatedAt().Time,
		UpdatedAt: issue.GetUpdatedAt().Time,
		URL:       issue.GetURL(),
		HTMLURL:   issue.GetHTMLURL(),
		SyncedAt:  time.Now(),
	}
	
	if issue.ClosedAt != nil {
		closedAt := issue.GetClosedAt().Time
		mapping.ClosedAt = &closedAt
	}
	
	// Map labels
	mapping.Labels = make([]GitHubLabel, 0)
	mapping.Tags = make([]string, 0)
	for _, label := range issue.Labels {
		ghLabel := GitHubLabel{
			Name:        label.GetName(),
			Color:       label.GetColor(),
			Description: label.GetDescription(),
		}
		mapping.Labels = append(mapping.Labels, ghLabel)
		mapping.Tags = append(mapping.Tags, label.GetName())
	}
	
	// Map assignees
	mapping.Assignees = make([]GitHubUser, 0)
	if len(issue.Assignees) > 0 {
		for _, assignee := range issue.Assignees {
			ghUser := GitHubUser{
				Login:     assignee.GetLogin(),
				ID:        assignee.GetID(),
				AvatarURL: assignee.GetAvatarURL(),
				HTMLURL:   assignee.GetHTMLURL(),
			}
			mapping.Assignees = append(mapping.Assignees, ghUser)
		}
		// Use first assignee as the primary assignee
		mapping.AssignedTo = issue.Assignees[0].GetLogin()
	}
	
	// Map author
	if issue.User != nil {
		mapping.Author = GitHubUser{
			Login:     issue.User.GetLogin(),
			ID:        issue.User.GetID(),
			AvatarURL: issue.User.GetAvatarURL(),
			HTMLURL:   issue.User.GetHTMLURL(),
		}
	}
	
	// Map milestone
	if issue.Milestone != nil {
		milestone := GitHubMilestone{
			Title:       issue.Milestone.GetTitle(),
			Number:      issue.Milestone.GetNumber(),
			State:       issue.Milestone.GetState(),
			Description: issue.Milestone.GetDescription(),
		}
		if issue.Milestone.DueOn != nil {
			dueOn := issue.Milestone.GetDueOn().Time
			milestone.DueOn = &dueOn
		}
		mapping.Milestone = &milestone
	}
	
	// Determine ticket type from labels
	mapping.TicketType = options.DefaultType
	for _, label := range issue.Labels {
		if ticketType, exists := options.LabelToType[label.GetName()]; exists {
			mapping.TicketType = ticketType
			break
		}
	}
	
	// Determine ticket priority from labels
	mapping.TicketPriority = options.DefaultPriority
	for _, label := range issue.Labels {
		if priority, exists := options.LabelToPriority[label.GetName()]; exists {
			mapping.TicketPriority = priority
			break
		}
	}
	
	// Determine ticket status from issue state
	switch issue.GetState() {
	case "open":
		mapping.TicketStatus = ticket.TicketStatusOpen
	case "closed":
		if options.CloseResolved {
			mapping.TicketStatus = ticket.TicketStatusClosed
		} else {
			mapping.TicketStatus = ticket.TicketStatusResolved
		}
	default:
		mapping.TicketStatus = ticket.TicketStatusOpen
	}
	
	return mapping
}

func (gi *Integration) findExistingTicket(issueNumber int) string {
	// Look for tickets with GitHub external reference
	tickets, err := gi.ticketManager.ListTickets(ticket.TicketListOptions{
		ShowClosed: true,
	})
	if err != nil {
		return ""
	}
	
	issueNumberStr := strconv.Itoa(issueNumber)
	
	for _, t := range tickets {
		if t.ExternalRef != nil && 
		   t.ExternalRef.System == "github" && 
		   t.ExternalRef.ID == issueNumberStr {
			return t.ID
		}
	}
	
	return ""
}

func (gi *Integration) createNewTicket(mapping IssueMapping) (string, error) {
	// Create external reference
	externalRef := &ticket.ExternalReference{
		System: "github",
		ID:     strconv.Itoa(mapping.Number),
		URL:    mapping.HTMLURL,
		Metadata: map[string]interface{}{
			"author":     mapping.Author.Login,
			"created_at": mapping.CreatedAt,
			"updated_at": mapping.UpdatedAt,
		},
	}
	
	// Create ticket options
	options := ticket.TicketCreateOptions{
		Title:       mapping.Title,
		Description: gi.formatIssueDescription(mapping),
		Type:        mapping.TicketType,
		Priority:    mapping.TicketPriority,
		AssignedTo:  mapping.AssignedTo,
		Tags:        mapping.Tags,
		ExternalRef: externalRef,
	}
	
	// Create the ticket
	newTicket, err := gi.ticketManager.CreateTicket(options)
	if err != nil {
		return "", fmt.Errorf("failed to create ticket: %w", err)
	}
	
	// Update ticket status if needed
	if mapping.TicketStatus != ticket.TicketStatusOpen {
		_, err = gi.ticketManager.UpdateTicket(newTicket.ID, ticket.TicketUpdateOptions{
			Status: &mapping.TicketStatus,
		})
		if err != nil {
			// Don't fail if status update fails, ticket is still created
			fmt.Printf("Warning: failed to update ticket status: %v\n", err)
		}
	}
	
	return newTicket.ID, nil
}

func (gi *Integration) updateExistingTicket(ticketID string, mapping IssueMapping, options *IssueSyncOptions) error {
	// Get current ticket
	existingTicket, err := gi.ticketManager.GetTicket(ticketID)
	if err != nil {
		return fmt.Errorf("failed to get existing ticket: %w", err)
	}
	
	// Prepare update options
	updateOptions := ticket.TicketUpdateOptions{}
	
	// Update title if changed
	if existingTicket.Title != mapping.Title {
		updateOptions.Title = &mapping.Title
	}
	
	// Update description
	newDescription := gi.formatIssueDescription(mapping)
	if existingTicket.Description != newDescription {
		updateOptions.Description = &newDescription
	}
	
	// Update priority if changed
	if existingTicket.Priority != mapping.TicketPriority {
		updateOptions.Priority = &mapping.TicketPriority
	}
	
	// Update type if changed
	if existingTicket.Type != mapping.TicketType {
		updateOptions.Type = &mapping.TicketType
	}
	
	// Update assignee if changed
	if existingTicket.AssignedTo != mapping.AssignedTo {
		updateOptions.AssignedTo = &mapping.AssignedTo
	}
	
	// Update tags if changed
	if !stringSlicesEqual(existingTicket.Tags, mapping.Tags) {
		updateOptions.Tags = &mapping.Tags
	}
	
	// Update status if enabled and changed
	if options.CloseResolved && existingTicket.Status != mapping.TicketStatus {
		updateOptions.Status = &mapping.TicketStatus
	}
	
	// Update external reference metadata
	if existingTicket.ExternalRef != nil {
		existingTicket.ExternalRef.Metadata["updated_at"] = mapping.UpdatedAt
		updateOptions.ExternalRef = existingTicket.ExternalRef
	}
	
	// Apply updates if there are any changes
	if updateOptions.Title != nil || updateOptions.Description != nil || 
	   updateOptions.Priority != nil || updateOptions.Type != nil ||
	   updateOptions.AssignedTo != nil || updateOptions.Tags != nil ||
	   updateOptions.Status != nil || updateOptions.ExternalRef != nil {
		
		_, err = gi.ticketManager.UpdateTicket(ticketID, updateOptions)
		if err != nil {
			return fmt.Errorf("failed to update ticket: %w", err)
		}
	}
	
	return nil
}

func (gi *Integration) formatIssueDescription(mapping IssueMapping) string {
	description := fmt.Sprintf("GitHub Issue #%d\n\n", mapping.Number)
	
	if mapping.Body != "" {
		description += mapping.Body + "\n\n"
	}
	
	description += fmt.Sprintf("**GitHub Details:**\n")
	description += fmt.Sprintf("- Issue: [#%d](%s)\n", mapping.Number, mapping.HTMLURL)
	description += fmt.Sprintf("- Author: @%s\n", mapping.Author.Login)
	description += fmt.Sprintf("- Created: %s\n", mapping.CreatedAt.Format("2006-01-02 15:04:05"))
	description += fmt.Sprintf("- Updated: %s\n", mapping.UpdatedAt.Format("2006-01-02 15:04:05"))
	
	if len(mapping.Labels) > 0 {
		description += fmt.Sprintf("- Labels: %s\n", strings.Join(mapping.Tags, ", "))
	}
	
	if len(mapping.Assignees) > 0 {
		assignees := make([]string, 0)
		for _, assignee := range mapping.Assignees {
			assignees = append(assignees, "@"+assignee.Login)
		}
		description += fmt.Sprintf("- Assignees: %s\n", strings.Join(assignees, ", "))
	}
	
	if mapping.Milestone != nil {
		description += fmt.Sprintf("- Milestone: %s\n", mapping.Milestone.Title)
	}
	
	return description
}

func (gi *Integration) updateSyncHistory(result *IssueSyncResult) {
	gi.config.History.LastSync = result.SyncTime
	gi.config.History.TotalSyncs++
	
	if result.ErrorCount == 0 {
		gi.config.History.LastSuccessfulSync = result.SyncTime
		gi.config.History.SuccessfulSyncs++
	} else {
		gi.config.History.FailedSyncs++
		gi.config.History.LastErrors = result.Errors
	}
	
	// Keep only recent results (last 10)
	gi.config.History.RecentResults = append(gi.config.History.RecentResults, *result)
	if len(gi.config.History.RecentResults) > 10 {
		gi.config.History.RecentResults = gi.config.History.RecentResults[1:]
	}
	
	// Save updated configuration
	gi.SaveConfig()
}

// Helper functions

func stringSlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}