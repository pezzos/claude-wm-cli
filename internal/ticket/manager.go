package ticket

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	
	"claude-wm-cli/internal/epic"
)

const (
	TicketsFileName = "tickets.json"
	TicketsVersion  = "1.0.0"
)

// Manager handles ticket operations and persistence
type Manager struct {
	rootPath    string
	epicManager *epic.Manager
}

// NewManager creates a new ticket manager
func NewManager(rootPath string) *Manager {
	return &Manager{
		rootPath:    rootPath,
		epicManager: epic.NewManager(rootPath),
	}
}

// CreateTicket creates a new ticket
func (m *Manager) CreateTicket(options TicketCreateOptions) (*Ticket, error) {
	// Validate inputs
	if strings.TrimSpace(options.Title) == "" {
		return nil, fmt.Errorf("ticket title cannot be empty")
	}
	
	if options.Type != "" && !options.Type.IsValid() {
		return nil, fmt.Errorf("invalid ticket type: %s", options.Type)
	}
	
	if options.Priority != "" && !options.Priority.IsValid() {
		return nil, fmt.Errorf("invalid ticket priority: %s", options.Priority)
	}
	
	// Validate epic/story references if provided
	if options.RelatedEpicID != "" {
		if _, err := m.epicManager.GetEpic(options.RelatedEpicID); err != nil {
			return nil, fmt.Errorf("related epic not found: %s", options.RelatedEpicID)
		}
	}
	
	// Load existing collection
	collection, err := m.loadTicketCollection()
	if err != nil {
		return nil, fmt.Errorf("failed to load ticket collection: %w", err)
	}
	
	// Generate unique ID
	ticketID := m.generateTicketID(options.Title, collection)
	
	// Set defaults
	if options.Type == "" {
		options.Type = TicketTypeTask
	}
	if options.Priority == "" {
		options.Priority = TicketPriorityMedium
	}
	
	// Create the ticket
	now := time.Now()
	ticket := &Ticket{
		ID:             ticketID,
		Title:          strings.TrimSpace(options.Title),
		Description:    strings.TrimSpace(options.Description),
		Type:           options.Type,
		Status:         TicketStatusOpen,
		Priority:       options.Priority,
		RelatedEpicID:  options.RelatedEpicID,
		RelatedStoryID: options.RelatedStoryID,
		AssignedTo:     options.AssignedTo,
		Estimations: TicketEstimation{
			EstimatedHours: options.EstimatedHours,
			StoryPoints:    options.StoryPoints,
		},
		Tags:           options.Tags,
		DueDate:        options.DueDate,
		ExternalRef:    options.ExternalRef,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	
	// Add to collection
	collection.Tickets[ticketID] = ticket
	
	// Update metadata
	m.updateCollectionMetadata(collection)
	
	// Save collection
	if err := m.saveTicketCollection(collection); err != nil {
		return nil, fmt.Errorf("failed to save ticket collection: %w", err)
	}
	
	return ticket, nil
}

// UpdateTicket updates an existing ticket
func (m *Manager) UpdateTicket(ticketID string, options TicketUpdateOptions) (*Ticket, error) {
	collection, err := m.loadTicketCollection()
	if err != nil {
		return nil, fmt.Errorf("failed to load ticket collection: %w", err)
	}
	
	ticket, exists := collection.Tickets[ticketID]
	if !exists {
		return nil, fmt.Errorf("ticket not found: %s", ticketID)
	}
	
	// Apply updates
	now := time.Now()
	
	if options.Title != nil {
		if strings.TrimSpace(*options.Title) == "" {
			return nil, fmt.Errorf("ticket title cannot be empty")
		}
		ticket.Title = strings.TrimSpace(*options.Title)
	}
	
	if options.Description != nil {
		ticket.Description = strings.TrimSpace(*options.Description)
	}
	
	if options.Type != nil {
		if !options.Type.IsValid() {
			return nil, fmt.Errorf("invalid ticket type: %s", *options.Type)
		}
		ticket.Type = *options.Type
	}
	
	if options.Status != nil {
		if !options.Status.IsValid() {
			return nil, fmt.Errorf("invalid ticket status: %s", *options.Status)
		}
		
		// Validate status transition
		if err := m.validateStatusTransition(ticket, *options.Status); err != nil {
			return nil, err
		}
		
		// Handle status change timestamps
		oldStatus := ticket.Status
		ticket.Status = *options.Status
		
		if *options.Status == TicketStatusInProgress && ticket.StartedAt == nil {
			ticket.StartedAt = &now
		}
		if *options.Status == TicketStatusResolved && ticket.ResolvedAt == nil {
			ticket.ResolvedAt = &now
		}
		if *options.Status == TicketStatusClosed && ticket.ClosedAt == nil {
			ticket.ClosedAt = &now
		}
		
		// Log activity
		m.logTicketActivity(collection, ticketID, "status_changed", oldStatus, *options.Status, now)
	}
	
	if options.Priority != nil {
		if !options.Priority.IsValid() {
			return nil, fmt.Errorf("invalid ticket priority: %s", *options.Priority)
		}
		ticket.Priority = *options.Priority
	}
	
	if options.RelatedEpicID != nil {
		if *options.RelatedEpicID != "" {
			if _, err := m.epicManager.GetEpic(*options.RelatedEpicID); err != nil {
				return nil, fmt.Errorf("related epic not found: %s", *options.RelatedEpicID)
			}
		}
		ticket.RelatedEpicID = *options.RelatedEpicID
	}
	
	if options.RelatedStoryID != nil {
		ticket.RelatedStoryID = *options.RelatedStoryID
	}
	
	if options.AssignedTo != nil {
		ticket.AssignedTo = *options.AssignedTo
	}
	
	if options.EstimatedHours != nil {
		ticket.Estimations.EstimatedHours = *options.EstimatedHours
	}
	
	if options.ActualHours != nil {
		ticket.Estimations.ActualHours = *options.ActualHours
	}
	
	if options.StoryPoints != nil {
		ticket.Estimations.StoryPoints = *options.StoryPoints
	}
	
	if options.Tags != nil {
		ticket.Tags = *options.Tags
	}
	
	if options.DueDate != nil {
		ticket.DueDate = options.DueDate
	}
	
	if options.ExternalRef != nil {
		ticket.ExternalRef = options.ExternalRef
	}
	
	ticket.UpdatedAt = now
	
	// Update metadata
	m.updateCollectionMetadata(collection)
	
	// Save collection
	if err := m.saveTicketCollection(collection); err != nil {
		return nil, fmt.Errorf("failed to save ticket collection: %w", err)
	}
	
	return ticket, nil
}

// GetTicket retrieves a specific ticket by ID
func (m *Manager) GetTicket(ticketID string) (*Ticket, error) {
	collection, err := m.loadTicketCollection()
	if err != nil {
		return nil, fmt.Errorf("failed to load ticket collection: %w", err)
	}
	
	ticket, exists := collection.Tickets[ticketID]
	if !exists {
		return nil, fmt.Errorf("ticket not found: %s", ticketID)
	}
	
	return ticket, nil
}

// ListTickets returns a filtered list of tickets
func (m *Manager) ListTickets(options TicketListOptions) ([]*Ticket, error) {
	collection, err := m.loadTicketCollection()
	if err != nil {
		return nil, fmt.Errorf("failed to load ticket collection: %w", err)
	}
	
	var tickets []*Ticket
	for _, ticket := range collection.Tickets {
		// Apply filters
		if options.Status != "" && ticket.Status != options.Status {
			continue
		}
		if options.Priority != "" && ticket.Priority != options.Priority {
			continue
		}
		if options.Type != "" && ticket.Type != options.Type {
			continue
		}
		if options.AssignedTo != "" && ticket.AssignedTo != options.AssignedTo {
			continue
		}
		if options.RelatedEpicID != "" && ticket.RelatedEpicID != options.RelatedEpicID {
			continue
		}
		if options.RelatedStoryID != "" && ticket.RelatedStoryID != options.RelatedStoryID {
			continue
		}
		if !options.ShowClosed && (ticket.Status == TicketStatusClosed) {
			continue
		}
		
		tickets = append(tickets, ticket)
	}
	
	// Sort by priority, then by creation date
	sort.Slice(tickets, func(i, j int) bool {
		// Priority order: urgent > critical > high > medium > low
		priorityOrder := map[TicketPriority]int{
			TicketPriorityUrgent:   5,
			TicketPriorityCritical: 4,
			TicketPriorityHigh:     3,
			TicketPriorityMedium:   2,
			TicketPriorityLow:      1,
		}
		
		if priorityOrder[tickets[i].Priority] != priorityOrder[tickets[j].Priority] {
			return priorityOrder[tickets[i].Priority] > priorityOrder[tickets[j].Priority]
		}
		
		// If same priority, sort by creation date (newest first)
		return tickets[i].CreatedAt.After(tickets[j].CreatedAt)
	})
	
	// Apply limit
	if options.Limit > 0 && len(tickets) > options.Limit {
		tickets = tickets[:options.Limit]
	}
	
	return tickets, nil
}

// SetCurrentTicket sets the active ticket
func (m *Manager) SetCurrentTicket(ticketID string) (*Ticket, error) {
	collection, err := m.loadTicketCollection()
	if err != nil {
		return nil, fmt.Errorf("failed to load ticket collection: %w", err)
	}
	
	if ticketID != "" {
		ticket, exists := collection.Tickets[ticketID]
		if !exists {
			return nil, fmt.Errorf("ticket not found: %s", ticketID)
		}
		
		// Auto-start ticket if it's open
		if ticket.Status == TicketStatusOpen {
			now := time.Now()
			ticket.Status = TicketStatusInProgress
			ticket.StartedAt = &now
			ticket.UpdatedAt = now
		}
		
		collection.CurrentTicket = ticketID
		
		// Save collection
		if err := m.saveTicketCollection(collection); err != nil {
			return nil, fmt.Errorf("failed to save ticket collection: %w", err)
		}
		
		return ticket, nil
	}
	
	// Clear current ticket
	collection.CurrentTicket = ""
	if err := m.saveTicketCollection(collection); err != nil {
		return nil, fmt.Errorf("failed to save ticket collection: %w", err)
	}
	
	return nil, nil
}

// GetCurrentTicket returns the currently active ticket
func (m *Manager) GetCurrentTicket() (*Ticket, error) {
	collection, err := m.loadTicketCollection()
	if err != nil {
		return nil, fmt.Errorf("failed to load ticket collection: %w", err)
	}
	
	if collection.CurrentTicket == "" {
		return nil, nil
	}
	
	ticket, exists := collection.Tickets[collection.CurrentTicket]
	if !exists {
		// Clear invalid current ticket
		collection.CurrentTicket = ""
		m.saveTicketCollection(collection)
		return nil, nil
	}
	
	return ticket, nil
}

// GetTicketStats returns analytics on the ticket collection
func (m *Manager) GetTicketStats() (*TicketStats, error) {
	collection, err := m.loadTicketCollection()
	if err != nil {
		return nil, fmt.Errorf("failed to load ticket collection: %w", err)
	}
	
	stats := &TicketStats{
		TotalTickets: len(collection.Tickets),
		ByStatus:     make(map[TicketStatus]int),
		ByPriority:   make(map[TicketPriority]int),
		ByType:       make(map[TicketType]int),
	}
	
	var resolutionTimes []time.Duration
	var oldestOpen *time.Time
	
	for _, ticket := range collection.Tickets {
		// Count by status
		stats.ByStatus[ticket.Status]++
		
		// Count by priority
		stats.ByPriority[ticket.Priority]++
		
		// Count by type
		stats.ByType[ticket.Type]++
		
		// Calculate resolution times
		if ticket.ResolvedAt != nil {
			duration := ticket.ResolvedAt.Sub(ticket.CreatedAt)
			resolutionTimes = append(resolutionTimes, duration)
		}
		
		// Track oldest open ticket
		if ticket.Status == TicketStatusOpen || ticket.Status == TicketStatusInProgress {
			if oldestOpen == nil || ticket.CreatedAt.Before(*oldestOpen) {
				oldestOpen = &ticket.CreatedAt
			}
		}
	}
	
	// Calculate average resolution time
	if len(resolutionTimes) > 0 {
		var total time.Duration
		for _, duration := range resolutionTimes {
			total += duration
		}
		stats.AverageResolutionTime = total / time.Duration(len(resolutionTimes))
	}
	
	stats.OldestOpenTicket = oldestOpen
	
	return stats, nil
}

// DeleteTicket removes a ticket from the collection
func (m *Manager) DeleteTicket(ticketID string) error {
	collection, err := m.loadTicketCollection()
	if err != nil {
		return fmt.Errorf("failed to load ticket collection: %w", err)
	}
	
	_, exists := collection.Tickets[ticketID]
	if !exists {
		return fmt.Errorf("ticket not found: %s", ticketID)
	}
	
	// Clear current ticket if it's the one being deleted
	if collection.CurrentTicket == ticketID {
		collection.CurrentTicket = ""
	}
	
	// Remove from collection
	delete(collection.Tickets, ticketID)
	
	// Update metadata
	m.updateCollectionMetadata(collection)
	
	// Save collection
	return m.saveTicketCollection(collection)
}

// Helper methods

func (m *Manager) loadTicketCollection() (*TicketCollection, error) {
	ticketsPath := filepath.Join(m.rootPath, "docs", "2-current-epic", TicketsFileName)
	
	// Check if file exists
	if _, err := os.Stat(ticketsPath); os.IsNotExist(err) {
		// Create default collection
		return &TicketCollection{
			Tickets: make(map[string]*Ticket),
			Metadata: TicketMetadata{
				Version:     TicketsVersion,
				LastUpdated: time.Now(),
			},
		}, nil
	}
	
	// Read file
	data, err := os.ReadFile(ticketsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read tickets file: %w", err)
	}
	
	var collection TicketCollection
	if err := json.Unmarshal(data, &collection); err != nil {
		return nil, fmt.Errorf("failed to parse tickets file: %w", err)
	}
	
	// Validate and migrate if needed
	if err := m.validateAndMigrateCollection(&collection); err != nil {
		return nil, fmt.Errorf("failed to validate ticket collection: %w", err)
	}
	
	return &collection, nil
}

func (m *Manager) saveTicketCollection(collection *TicketCollection) error {
	ticketsPath := filepath.Join(m.rootPath, "docs", "2-current-epic", TicketsFileName)
	
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(ticketsPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	
	// Update metadata
	collection.Metadata.LastUpdated = time.Now()
	collection.Metadata.Version = TicketsVersion
	
	// Marshal to JSON
	data, err := json.MarshalIndent(collection, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal ticket collection: %w", err)
	}
	
	// Write file atomically using temp file + rename
	tempPath := ticketsPath + ".tmp"
	if err := os.WriteFile(tempPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write temp tickets file: %w", err)
	}
	
	if err := os.Rename(tempPath, ticketsPath); err != nil {
		os.Remove(tempPath) // cleanup
		return fmt.Errorf("failed to replace tickets file: %w", err)
	}
	
	return nil
}

func (m *Manager) generateTicketID(title string, collection *TicketCollection) string {
	// Create base ID from title
	baseID := strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(title), " ", "-"))
	baseID = strings.ReplaceAll(baseID, "_", "-")
	
	// Remove special characters
	var cleaned strings.Builder
	for _, r := range baseID {
		if (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' {
			cleaned.WriteRune(r)
		}
	}
	
	baseID = cleaned.String()
	if baseID == "" {
		baseID = "TICKET"
	}
	
	// Ensure uniqueness
	counter := 1
	ticketID := fmt.Sprintf("TICKET-%03d-%s", counter, baseID)
	
	for {
		if _, exists := collection.Tickets[ticketID]; !exists {
			break
		}
		counter++
		ticketID = fmt.Sprintf("TICKET-%03d-%s", counter, baseID)
	}
	
	return ticketID
}

func (m *Manager) validateStatusTransition(ticket *Ticket, newStatus TicketStatus) error {
	currentStatus := ticket.Status
	
	// Define valid transitions
	validTransitions := map[TicketStatus][]TicketStatus{
		TicketStatusOpen:       {TicketStatusInProgress, TicketStatusClosed},
		TicketStatusInProgress: {TicketStatusResolved, TicketStatusOpen, TicketStatusClosed},
		TicketStatusResolved:   {TicketStatusClosed, TicketStatusInProgress}, // Can reopen
		TicketStatusClosed:     {TicketStatusOpen}, // Can reopen
	}
	
	allowedTransitions, exists := validTransitions[currentStatus]
	if !exists {
		return fmt.Errorf("unknown current status: %s", currentStatus)
	}
	
	// Check if transition is allowed
	for _, allowed := range allowedTransitions {
		if allowed == newStatus {
			return nil
		}
	}
	
	return fmt.Errorf("invalid status transition from %s to %s", currentStatus, newStatus)
}

func (m *Manager) updateCollectionMetadata(collection *TicketCollection) {
	collection.Metadata.TotalTickets = len(collection.Tickets)
	collection.Metadata.OpenTickets = 0
	collection.Metadata.ResolvedTickets = 0
	
	for _, ticket := range collection.Tickets {
		if ticket.Status == TicketStatusOpen || ticket.Status == TicketStatusInProgress {
			collection.Metadata.OpenTickets++
		}
		if ticket.Status == TicketStatusResolved || ticket.Status == TicketStatusClosed {
			collection.Metadata.ResolvedTickets++
		}
	}
}

func (m *Manager) validateAndMigrateCollection(collection *TicketCollection) error {
	// Initialize maps if nil
	if collection.Tickets == nil {
		collection.Tickets = make(map[string]*Ticket)
	}
	
	// Set default metadata if missing
	if collection.Metadata.Version == "" {
		collection.Metadata.Version = TicketsVersion
	}
	
	// Validate each ticket
	for id, ticket := range collection.Tickets {
		if ticket == nil {
			delete(collection.Tickets, id)
			continue
		}
		
		// Ensure required fields
		if ticket.ID == "" {
			ticket.ID = id
		}
		if ticket.CreatedAt.IsZero() {
			ticket.CreatedAt = time.Now()
		}
		if ticket.UpdatedAt.IsZero() {
			ticket.UpdatedAt = ticket.CreatedAt
		}
		if ticket.Priority == "" {
			ticket.Priority = TicketPriorityMedium
		}
		if ticket.Status == "" {
			ticket.Status = TicketStatusOpen
		}
		if ticket.Type == "" {
			ticket.Type = TicketTypeTask
		}
	}
	
	// Update metadata
	m.updateCollectionMetadata(collection)
	
	return nil
}

func (m *Manager) logTicketActivity(collection *TicketCollection, ticketID, action string, oldValue, newValue interface{}, timestamp time.Time) {
	// Note: In a full implementation, this would log to a separate activity log
	// For now, we'll keep it simple and just update the ticket's UpdatedAt time
	if ticket, exists := collection.Tickets[ticketID]; exists {
		ticket.UpdatedAt = timestamp
	}
}