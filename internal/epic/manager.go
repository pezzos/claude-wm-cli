package epic

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	EpicsFileName = "epics.json"
	EpicsVersion  = "1.0.0"
)

// Manager handles epic operations and state management
type Manager struct {
	rootPath string
	tracker  *EpicTracker
}

// NewManager creates a new epic manager
func NewManager(rootPath string) *Manager {
	manager := &Manager{
		rootPath: rootPath,
	}
	// Initialize tracker after manager is created
	manager.tracker = NewEpicTracker(manager)
	return manager
}

// GetTracker returns the epic tracker for advanced state management
func (m *Manager) GetTracker() *EpicTracker {
	return m.tracker
}

// CreateEpic creates a new epic with the given options
func (m *Manager) CreateEpic(options EpicCreateOptions) (*Epic, error) {
	// Validate inputs
	if strings.TrimSpace(options.Title) == "" {
		return nil, fmt.Errorf("epic title cannot be empty")
	}

	if options.Priority != "" && !options.Priority.IsValid() {
		return nil, fmt.Errorf("invalid priority: %s", options.Priority)
	}

	// Set default priority if not specified
	if options.Priority == "" {
		options.Priority = PriorityMedium
	}

	// Load existing epics
	collection, err := m.loadEpicCollection()
	if err != nil {
		return nil, fmt.Errorf("failed to load epic collection: %w", err)
	}

	// Generate unique ID
	epicID := m.generateEpicID(options.Title, collection)

	// Create the epic
	now := time.Now()
	epic := &Epic{
		ID:           epicID,
		Title:        strings.TrimSpace(options.Title),
		Description:  strings.TrimSpace(options.Description),
		Priority:     options.Priority,
		Status:       StatusPlanned,
		Duration:     options.Duration,
		Tags:         options.Tags,
		Dependencies: options.Dependencies,
		UserStories:  []UserStory{},
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	epic.CalculateProgress()

	// Add to collection
	collection.Epics[epicID] = epic
	collection.Metadata.TotalEpics = len(collection.Epics)
	collection.Metadata.LastUpdated = now

	// Save collection
	if err := m.saveEpicCollection(collection); err != nil {
		return nil, fmt.Errorf("failed to save epic collection: %w", err)
	}

	// Notify tracker of the new epic
	if m.tracker != nil {
		go m.tracker.UpdateEpicBasedOnStories(epic.ID)
	}

	return epic, nil
}

// ListEpics returns a list of epics based on the given options
func (m *Manager) ListEpics(options EpicListOptions) ([]*Epic, error) {
	collection, err := m.loadEpicCollection()
	if err != nil {
		return nil, fmt.Errorf("failed to load epic collection: %w", err)
	}

	var epics []*Epic
	for _, epic := range collection.Epics {
		// Apply filters
		if options.Status != "" && epic.Status != options.Status {
			continue
		}
		if options.Priority != "" && epic.Priority != options.Priority {
			continue
		}

		epics = append(epics, epic)
	}

	// Sort by creation date (newest first)
	sort.Slice(epics, func(i, j int) bool {
		return epics[i].CreatedAt.After(epics[j].CreatedAt)
	})

	return epics, nil
}

// UpdateEpic updates an existing epic with the given options
func (m *Manager) UpdateEpic(epicID string, options EpicUpdateOptions) (*Epic, error) {
	collection, err := m.loadEpicCollection()
	if err != nil {
		return nil, fmt.Errorf("failed to load epic collection: %w", err)
	}

	epic, exists := collection.Epics[epicID]
	if !exists {
		return nil, fmt.Errorf("epic not found: %s", epicID)
	}

	// Apply updates
	now := time.Now()

	if options.Title != nil {
		if strings.TrimSpace(*options.Title) == "" {
			return nil, fmt.Errorf("epic title cannot be empty")
		}
		epic.Title = strings.TrimSpace(*options.Title)
	}

	if options.Description != nil {
		epic.Description = strings.TrimSpace(*options.Description)
	}

	if options.Priority != nil {
		if !options.Priority.IsValid() {
			return nil, fmt.Errorf("invalid priority: %s", *options.Priority)
		}
		epic.Priority = *options.Priority
	}

	if options.Status != nil {
		if !options.Status.IsValid() {
			return nil, fmt.Errorf("invalid status: %s", *options.Status)
		}

		// Handle status transitions
		if err := m.validateStatusTransition(epic, *options.Status); err != nil {
			return nil, err
		}

		epic.Status = *options.Status

		// Set timestamps for status changes
		if *options.Status == StatusInProgress && epic.StartDate == nil {
			epic.StartDate = &now
		}
		if *options.Status == StatusCompleted && epic.EndDate == nil {
			epic.EndDate = &now
		}
	}

	if options.Duration != nil {
		epic.Duration = *options.Duration
	}

	if options.Tags != nil {
		epic.Tags = *options.Tags
	}

	if options.Dependencies != nil {
		epic.Dependencies = *options.Dependencies
	}

	epic.UpdatedAt = now
	epic.CalculateProgress()

	// Update metadata
	collection.Metadata.LastUpdated = now

	// Save collection
	if err := m.saveEpicCollection(collection); err != nil {
		return nil, fmt.Errorf("failed to save epic collection: %w", err)
	}

	// Notify tracker of the update
	if m.tracker != nil {
		go m.tracker.UpdateEpicBasedOnStories(epic.ID)
	}

	return epic, nil
}

// SelectEpic sets the given epic as the current active epic
func (m *Manager) SelectEpic(epicID string) (*Epic, error) {
	collection, err := m.loadEpicCollection()
	if err != nil {
		return nil, fmt.Errorf("failed to load epic collection: %w", err)
	}

	epic, exists := collection.Epics[epicID]
	if !exists {
		return nil, fmt.Errorf("epic not found: %s", epicID)
	}

	// Can only select planned or in-progress epics
	if epic.Status != StatusPlanned && epic.Status != StatusInProgress {
		return nil, fmt.Errorf("cannot select epic with status: %s", epic.Status)
	}

	// Set as current epic
	collection.CurrentEpic = epicID
	collection.Metadata.LastUpdated = time.Now()

	// Start the epic if it's planned
	if epic.Status == StatusPlanned {
		now := time.Now()
		epic.Status = StatusInProgress
		epic.StartDate = &now
		epic.UpdatedAt = now
	}

	// Save collection
	if err := m.saveEpicCollection(collection); err != nil {
		return nil, fmt.Errorf("failed to save epic collection: %w", err)
	}

	return epic, nil
}

// GetEpic returns a specific epic by ID
func (m *Manager) GetEpic(epicID string) (*Epic, error) {
	collection, err := m.loadEpicCollection()
	if err != nil {
		return nil, fmt.Errorf("failed to load epic collection: %w", err)
	}

	epic, exists := collection.Epics[epicID]
	if !exists {
		return nil, fmt.Errorf("epic not found: %s", epicID)
	}

	return epic, nil
}

// GetCurrentEpic returns the currently active epic
func (m *Manager) GetCurrentEpic() (*Epic, error) {
	collection, err := m.loadEpicCollection()
	if err != nil {
		return nil, fmt.Errorf("failed to load epic collection: %w", err)
	}

	if collection.CurrentEpic == "" {
		return nil, fmt.Errorf("no epic is currently active")
	}

	epic, exists := collection.Epics[collection.CurrentEpic]
	if !exists {
		// Current epic was deleted, clear the reference
		collection.CurrentEpic = ""
		m.saveEpicCollection(collection)
		return nil, fmt.Errorf("current epic no longer exists")
	}

	return epic, nil
}

// GetEpicCollection returns the entire epic collection
func (m *Manager) GetEpicCollection() (*EpicCollection, error) {
	return m.loadEpicCollection()
}

// UpdateEpicWithStoryTracking updates an epic and triggers story-based tracking
func (m *Manager) UpdateEpicWithStoryTracking(epicID string, options EpicUpdateOptions) (*Epic, error) {
	epic, err := m.UpdateEpic(epicID, options)
	if err != nil {
		return nil, err
	}

	// Use tracker to validate and potentially auto-transition based on stories
	if m.tracker != nil {
		if err := m.tracker.UpdateEpicBasedOnStories(epicID); err != nil {
			// Log error but don't fail the update
			fmt.Printf("Warning: Failed to update epic tracking: %v\n", err)
		}
	}

	return epic, nil
}

// TransitionEpicStatus uses the tracker to safely transition epic status
func (m *Manager) TransitionEpicStatus(epicID string, newStatus Status, reason string) (*Epic, error) {
	if m.tracker == nil {
		// Fallback to direct update
		statusPtr := newStatus
		return m.UpdateEpic(epicID, EpicUpdateOptions{Status: &statusPtr})
	}

	// Use tracker for validated transition
	var transitionReason TransitionReason
	switch reason {
	case "manual":
		transitionReason = ReasonManual
	case "auto":
		transitionReason = ReasonAutoStoryComplete
	default:
		transitionReason = ReasonManual
	}

	if err := m.tracker.ValidateAndTransitionState(epicID, newStatus, transitionReason, "manual"); err != nil {
		return nil, err
	}

	// Return updated epic
	return m.GetEpic(epicID)
}

// GetEpicStateHistory returns the state transition history for an epic
func (m *Manager) GetEpicStateHistory(epicID string) []StateTransition {
	if m.tracker == nil {
		return []StateTransition{}
	}
	return m.tracker.GetStateHistory(epicID)
}

// GetEpicAdvancedMetrics returns advanced metrics for an epic
func (m *Manager) GetEpicAdvancedMetrics(epicID string) (*AdvancedMetrics, error) {
	if m.tracker == nil {
		return nil, fmt.Errorf("tracker not available")
	}
	return m.tracker.CalculateAdvancedMetrics(epicID)
}

// DeleteEpic removes an epic from the collection
func (m *Manager) DeleteEpic(epicID string) error {
	collection, err := m.loadEpicCollection()
	if err != nil {
		return fmt.Errorf("failed to load epic collection: %w", err)
	}

	_, exists := collection.Epics[epicID]
	if !exists {
		return fmt.Errorf("epic not found: %s", epicID)
	}

	// Clear current epic if it's the one being deleted
	if collection.CurrentEpic == epicID {
		collection.CurrentEpic = ""
	}

	// Remove from collection
	delete(collection.Epics, epicID)
	collection.Metadata.TotalEpics = len(collection.Epics)
	collection.Metadata.LastUpdated = time.Now()

	// Save collection
	return m.saveEpicCollection(collection)
}

// loadEpicCollection loads the epic collection from disk
func (m *Manager) loadEpicCollection() (*EpicCollection, error) {
	epicsPath := filepath.Join(m.rootPath, "docs", "1-project", EpicsFileName)

	// Check if file exists
	if _, err := os.Stat(epicsPath); os.IsNotExist(err) {
		// Create default collection
		return &EpicCollection{
			ProjectID:   "default",
			Epics:       make(map[string]*Epic),
			CurrentEpic: "",
			Metadata: CollectionMetadata{
				Version:     EpicsVersion,
				LastUpdated: time.Now(),
				TotalEpics:  0,
			},
		}, nil
	}

	// Read file
	data, err := os.ReadFile(epicsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read epics file: %w", err)
	}

	var collection EpicCollection

	// Try to unmarshal as the new format first
	if err := json.Unmarshal(data, &collection); err != nil {
		// If that fails, try to parse the old format and migrate
		var oldFormat struct {
			Epics []struct {
				ID          string `json:"id"`
				Title       string `json:"title"`
				Priority    string `json:"priority"`
				Status      string `json:"status"`
				Description string `json:"description"`
				// Add other fields as needed for migration
			} `json:"epics"`
		}

		if err := json.Unmarshal(data, &oldFormat); err != nil {
			return nil, fmt.Errorf("failed to parse epics file: %w", err)
		}

		// Migrate from old format to new format
		collection = EpicCollection{
			ProjectID:   "default",
			Epics:       make(map[string]*Epic),
			CurrentEpic: "",
			Metadata: CollectionMetadata{
				Version:     EpicsVersion,
				LastUpdated: time.Now(),
				TotalEpics:  len(oldFormat.Epics),
			},
		}

		// Convert old epics to new format
		for _, oldEpic := range oldFormat.Epics {
			// Map old priority values to new ones
			var priority Priority
			switch oldEpic.Priority {
			case "low":
				priority = PriorityLow
			case "medium":
				priority = PriorityMedium
			case "high":
				priority = PriorityHigh
			case "critical":
				priority = PriorityCritical
			default:
				priority = PriorityMedium
			}

			// Map old status values to new ones
			var status Status
			switch oldEpic.Status {
			case "todo", "backlog":
				status = StatusPlanned
			case "in_progress", "🚧 In Progress":
				status = StatusInProgress
			case "completed", "✅ Completed":
				status = StatusCompleted
			case "cancelled":
				status = StatusCancelled
			default:
				status = StatusPlanned
			}

			now := time.Now()
			epic := &Epic{
				ID:          oldEpic.ID,
				Title:       oldEpic.Title,
				Description: oldEpic.Description,
				Priority:    priority,
				Status:      status,
				UserStories: []UserStory{}, // Will be populated later if needed
				CreatedAt:   now,
				UpdatedAt:   now,
			}

			epic.CalculateProgress()
			collection.Epics[oldEpic.ID] = epic
		}

		// Save the migrated format back to disk
		if err := m.saveEpicCollection(&collection); err != nil {
			return nil, fmt.Errorf("failed to save migrated epic collection: %w", err)
		}
	}

	// Validate and migrate if needed
	if err := m.validateAndMigrateCollection(&collection); err != nil {
		return nil, fmt.Errorf("failed to validate epic collection: %w", err)
	}

	return &collection, nil
}

// saveEpicCollection saves the epic collection to disk
func (m *Manager) saveEpicCollection(collection *EpicCollection) error {
	epicsPath := filepath.Join(m.rootPath, "docs", "1-project", EpicsFileName)

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(epicsPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Update metadata
	collection.Metadata.LastUpdated = time.Now()
	collection.Metadata.Version = EpicsVersion

	// Marshal to JSON
	data, err := json.MarshalIndent(collection, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal epic collection: %w", err)
	}

	// Write file atomically using temp file + rename
	tempPath := epicsPath + ".tmp"
	if err := os.WriteFile(tempPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write temp epics file: %w", err)
	}

	if err := os.Rename(tempPath, epicsPath); err != nil {
		os.Remove(tempPath) // cleanup
		return fmt.Errorf("failed to replace epics file: %w", err)
	}

	return nil
}

// generateEpicID generates a unique ID for a new epic
func (m *Manager) generateEpicID(title string, collection *EpicCollection) string {
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
		baseID = "EPIC"
	}

	// Ensure uniqueness
	counter := 1
	epicID := fmt.Sprintf("EPIC-%03d-%s", counter, baseID)

	for {
		if _, exists := collection.Epics[epicID]; !exists {
			break
		}
		counter++
		epicID = fmt.Sprintf("EPIC-%03d-%s", counter, baseID)
	}

	return epicID
}

// validateStatusTransition checks if a status transition is valid
func (m *Manager) validateStatusTransition(epic *Epic, newStatus Status) error {
	currentStatus := epic.Status

	// Define valid transitions
	validTransitions := map[Status][]Status{
		StatusPlanned:    {StatusInProgress, StatusOnHold, StatusCancelled},
		StatusInProgress: {StatusOnHold, StatusCompleted, StatusCancelled},
		StatusOnHold:     {StatusInProgress, StatusCancelled},
		StatusCompleted:  {},              // Cannot transition from completed
		StatusCancelled:  {StatusPlanned}, // Can restart cancelled epics
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

// validateAndMigrateCollection validates and migrates the collection if needed
func (m *Manager) validateAndMigrateCollection(collection *EpicCollection) error {
	// Initialize maps if nil
	if collection.Epics == nil {
		collection.Epics = make(map[string]*Epic)
	}

	// Set default metadata if missing
	if collection.Metadata.Version == "" {
		collection.Metadata.Version = EpicsVersion
	}

	// Update epic count
	collection.Metadata.TotalEpics = len(collection.Epics)

	// Validate each epic
	for id, epic := range collection.Epics {
		if epic == nil {
			delete(collection.Epics, id)
			continue
		}

		// Ensure required fields
		if epic.ID == "" {
			epic.ID = id
		}
		if epic.CreatedAt.IsZero() {
			epic.CreatedAt = time.Now()
		}
		if epic.UpdatedAt.IsZero() {
			epic.UpdatedAt = epic.CreatedAt
		}
		if epic.Priority == "" {
			epic.Priority = PriorityMedium
		}
		if epic.Status == "" {
			epic.Status = StatusPlanned
		}
		if epic.UserStories == nil {
			epic.UserStories = []UserStory{}
		}

		// Recalculate progress
		epic.CalculateProgress()
	}

	return nil
}
