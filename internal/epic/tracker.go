package epic

import (
	"fmt"
	"sort"
	"sync"
	"time"
)

// StateTransition represents a change in epic state
type StateTransition struct {
	FromStatus  Status                 `json:"from_status"`
	ToStatus    Status                 `json:"to_status"`
	Timestamp   time.Time              `json:"timestamp"`
	Reason      TransitionReason       `json:"reason"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	TriggeredBy string                 `json:"triggered_by,omitempty"` // "auto", "manual", "system"
}

// TransitionReason indicates why a status transition occurred
type TransitionReason string

const (
	ReasonManual            TransitionReason = "manual"              // User-initiated change
	ReasonAutoStoryComplete TransitionReason = "auto_story_complete" // All stories completed
	ReasonAutoStoryStart    TransitionReason = "auto_story_start"    // First story started
	ReasonAutoTimeout       TransitionReason = "auto_timeout"        // Timeout-based transition
	ReasonAutoDependency    TransitionReason = "auto_dependency"     // Dependency resolution
	ReasonSystemMaintenance TransitionReason = "system_maintenance"  // System-triggered
)

// EpicStateEvent represents an event that occurred during epic state tracking
type EpicStateEvent struct {
	EpicID      string                 `json:"epic_id"`
	EventType   EventType              `json:"event_type"`
	Timestamp   time.Time              `json:"timestamp"`
	Description string                 `json:"description"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// EventType categorizes different types of state events
type EventType string

const (
	EventStatusChange    EventType = "status_change"
	EventProgressUpdate  EventType = "progress_update"
	EventDependencyCheck EventType = "dependency_check"
	EventValidationError EventType = "validation_error"
	EventAutoTransition  EventType = "auto_transition"
	EventMetricsUpdate   EventType = "metrics_update"
)

// TrackerConfig configures the behavior of the epic tracker
type TrackerConfig struct {
	AutoTransitionEnabled bool            `json:"auto_transition_enabled"`
	ProgressUpdateFreq    time.Duration   `json:"progress_update_freq"`
	MaxHistoryEntries     int             `json:"max_history_entries"`
	EnableEventLogging    bool            `json:"enable_event_logging"`
	ValidationRules       ValidationRules `json:"validation_rules"`
}

// ValidationRules defines the rules for epic state validation
type ValidationRules struct {
	RequireProgressForCompletion bool    `json:"require_progress_for_completion"`
	MinProgressForCompletion     float64 `json:"min_progress_for_completion"`
	MaxDurationDays              int     `json:"max_duration_days"`
	AllowBackwardTransitions     bool    `json:"allow_backward_transitions"`
}

// EpicTracker manages epic state tracking and automatic updates
type EpicTracker struct {
	mu          sync.RWMutex
	manager     *Manager
	config      TrackerConfig
	history     map[string][]StateTransition // epicID -> transitions
	events      []EpicStateEvent
	lastUpdate  map[string]time.Time // epicID -> last update time
	subscribers []StateChangeSubscriber
}

// StateChangeSubscriber interface for components that want to be notified of state changes
type StateChangeSubscriber interface {
	OnEpicStateChange(epicID string, transition StateTransition) error
}

// NewEpicTracker creates a new epic tracker with default configuration
func NewEpicTracker(manager *Manager) *EpicTracker {
	return &EpicTracker{
		manager: manager,
		config: TrackerConfig{
			AutoTransitionEnabled: true,
			ProgressUpdateFreq:    time.Minute * 5,
			MaxHistoryEntries:     100,
			EnableEventLogging:    true,
			ValidationRules: ValidationRules{
				RequireProgressForCompletion: true,
				MinProgressForCompletion:     100.0,
				MaxDurationDays:              365,
				AllowBackwardTransitions:     false,
			},
		},
		history:     make(map[string][]StateTransition),
		events:      make([]EpicStateEvent, 0),
		lastUpdate:  make(map[string]time.Time),
		subscribers: make([]StateChangeSubscriber, 0),
	}
}

// UpdateEpicBasedOnStories automatically updates epic status based on story completion
func (et *EpicTracker) UpdateEpicBasedOnStories(epicID string) error {
	et.mu.Lock()
	defer et.mu.Unlock()

	if !et.config.AutoTransitionEnabled {
		return nil
	}

	epic, err := et.manager.GetEpic(epicID)
	if err != nil {
		return fmt.Errorf("failed to get epic: %w", err)
	}

	// Calculate current progress
	previousProgress := epic.Progress.CompletionPercentage
	epic.CalculateProgress()
	currentProgress := epic.Progress.CompletionPercentage

	// Log progress update if changed
	if previousProgress != currentProgress {
		et.logEvent(EpicStateEvent{
			EpicID:      epicID,
			EventType:   EventProgressUpdate,
			Timestamp:   time.Now(),
			Description: fmt.Sprintf("Progress updated from %.1f%% to %.1f%%", previousProgress, currentProgress),
			Metadata: map[string]interface{}{
				"previous_progress": previousProgress,
				"current_progress":  currentProgress,
				"total_stories":     epic.Progress.TotalStories,
				"completed_stories": epic.Progress.CompletedStories,
			},
		})
	}

	// Determine if automatic status transition is needed
	newStatus, reason := et.determineAutoStatus(epic)
	if newStatus != epic.Status {
		return et.transitionEpicStatus(epic, newStatus, reason, "auto")
	}

	// Update last update time
	et.lastUpdate[epicID] = time.Now()
	return nil
}

// determineAutoStatus determines what status an epic should have based on its current state
func (et *EpicTracker) determineAutoStatus(epic *Epic) (Status, TransitionReason) {
	currentStatus := epic.Status
	progress := epic.Progress.CompletionPercentage

	switch currentStatus {
	case StatusPlanned:
		// Auto-start if any stories are in progress
		if epic.Progress.TotalStories > 0 && epic.Progress.CompletedStories > 0 {
			return StatusInProgress, ReasonAutoStoryStart
		}
		// Check if we have stories that are started but not completed
		for _, story := range epic.UserStories {
			if story.Status == StatusInProgress {
				return StatusInProgress, ReasonAutoStoryStart
			}
		}

	case StatusInProgress:
		// Auto-complete if all stories are done and we meet completion criteria
		if et.config.ValidationRules.RequireProgressForCompletion {
			if progress >= et.config.ValidationRules.MinProgressForCompletion {
				return StatusCompleted, ReasonAutoStoryComplete
			}
		} else if epic.Progress.TotalStories > 0 && epic.Progress.CompletedStories == epic.Progress.TotalStories {
			return StatusCompleted, ReasonAutoStoryComplete
		}

		// Check for timeout-based transitions if configured
		if et.config.ValidationRules.MaxDurationDays > 0 && epic.StartDate != nil {
			daysSinceStart := time.Since(*epic.StartDate).Hours() / 24
			if daysSinceStart > float64(et.config.ValidationRules.MaxDurationDays) {
				return StatusOnHold, ReasonAutoTimeout
			}
		}
	}

	return currentStatus, ReasonManual // No change needed
}

// ValidateAndTransitionState validates and performs a state transition
func (et *EpicTracker) ValidateAndTransitionState(epicID string, newStatus Status, reason TransitionReason, triggeredBy string) error {
	et.mu.Lock()
	defer et.mu.Unlock()

	epic, err := et.manager.GetEpic(epicID)
	if err != nil {
		return fmt.Errorf("failed to get epic: %w", err)
	}

	// Validate the transition
	if err := et.validateTransition(epic, newStatus); err != nil {
		et.logEvent(EpicStateEvent{
			EpicID:      epicID,
			EventType:   EventValidationError,
			Timestamp:   time.Now(),
			Description: fmt.Sprintf("Transition validation failed: %v", err),
			Metadata: map[string]interface{}{
				"from_status": epic.Status,
				"to_status":   newStatus,
				"error":       err.Error(),
			},
		})
		return err
	}

	return et.transitionEpicStatus(epic, newStatus, reason, triggeredBy)
}

// validateTransition validates if a status transition is allowed
func (et *EpicTracker) validateTransition(epic *Epic, newStatus Status) error {
	currentStatus := epic.Status

	// Basic validation using existing logic
	if err := et.manager.validateStatusTransition(epic, newStatus); err != nil {
		return err
	}

	// Additional validation based on tracker rules
	if newStatus == StatusCompleted {
		if et.config.ValidationRules.RequireProgressForCompletion {
			epic.CalculateProgress() // Ensure we have latest progress
			if epic.Progress.CompletionPercentage < et.config.ValidationRules.MinProgressForCompletion {
				return fmt.Errorf("cannot complete epic: progress is %.1f%%, required %.1f%%",
					epic.Progress.CompletionPercentage, et.config.ValidationRules.MinProgressForCompletion)
			}
		}
	}

	// Check for backward transitions if not allowed
	if !et.config.ValidationRules.AllowBackwardTransitions {
		if et.isBackwardTransition(currentStatus, newStatus) {
			return fmt.Errorf("backward transitions not allowed: %s -> %s", currentStatus, newStatus)
		}
	}

	return nil
}

// isBackwardTransition checks if a transition is considered backward
func (et *EpicTracker) isBackwardTransition(from, to Status) bool {
	statusOrder := map[Status]int{
		StatusPlanned:    1,
		StatusInProgress: 2,
		StatusOnHold:     2, // Same level as in_progress
		StatusCompleted:  3,
		StatusCancelled:  0, // Can move to cancelled from any state
	}

	fromOrder, fromExists := statusOrder[from]
	toOrder, toExists := statusOrder[to]

	if !fromExists || !toExists {
		return false
	}

	// Allow transitions to cancelled from any state
	if to == StatusCancelled {
		return false
	}

	// Allow transitions between in_progress and on_hold
	if (from == StatusInProgress && to == StatusOnHold) || (from == StatusOnHold && to == StatusInProgress) {
		return false
	}

	return toOrder < fromOrder
}

// transitionEpicStatus performs the actual status transition
func (et *EpicTracker) transitionEpicStatus(epic *Epic, newStatus Status, reason TransitionReason, triggeredBy string) error {
	oldStatus := epic.Status
	now := time.Now()

	// Create transition record
	transition := StateTransition{
		FromStatus:  oldStatus,
		ToStatus:    newStatus,
		Timestamp:   now,
		Reason:      reason,
		TriggeredBy: triggeredBy,
		Metadata: map[string]interface{}{
			"progress": epic.Progress.CompletionPercentage,
		},
	}

	// Update epic status using manager
	statusPtr := newStatus
	_, err := et.manager.UpdateEpic(epic.ID, EpicUpdateOptions{
		Status: &statusPtr,
	})
	if err != nil {
		return fmt.Errorf("failed to update epic status: %w", err)
	}

	// Record transition in history
	et.addTransitionToHistory(epic.ID, transition)

	// Log the transition event
	et.logEvent(EpicStateEvent{
		EpicID:      epic.ID,
		EventType:   EventStatusChange,
		Timestamp:   now,
		Description: fmt.Sprintf("Status changed from %s to %s (%s)", oldStatus, newStatus, reason),
		Metadata: map[string]interface{}{
			"transition": transition,
		},
	})

	// Notify subscribers
	for _, subscriber := range et.subscribers {
		if err := subscriber.OnEpicStateChange(epic.ID, transition); err != nil {
			et.logEvent(EpicStateEvent{
				EpicID:      epic.ID,
				EventType:   EventValidationError,
				Timestamp:   now,
				Description: fmt.Sprintf("Subscriber notification failed: %v", err),
			})
		}
	}

	return nil
}

// addTransitionToHistory adds a transition to the epic's history
func (et *EpicTracker) addTransitionToHistory(epicID string, transition StateTransition) {
	if et.history[epicID] == nil {
		et.history[epicID] = make([]StateTransition, 0)
	}

	et.history[epicID] = append(et.history[epicID], transition)

	// Trim history if it exceeds max entries
	if len(et.history[epicID]) > et.config.MaxHistoryEntries {
		// Keep the most recent entries
		et.history[epicID] = et.history[epicID][len(et.history[epicID])-et.config.MaxHistoryEntries:]
	}
}

// logEvent adds an event to the tracker's event log
func (et *EpicTracker) logEvent(event EpicStateEvent) {
	if !et.config.EnableEventLogging {
		return
	}

	et.events = append(et.events, event)

	// Trim events if they exceed reasonable limits
	maxEvents := et.config.MaxHistoryEntries * 5 // Allow more events than transitions
	if len(et.events) > maxEvents {
		et.events = et.events[len(et.events)-maxEvents:]
	}
}

// GetStateHistory returns the state transition history for an epic
func (et *EpicTracker) GetStateHistory(epicID string) []StateTransition {
	et.mu.RLock()
	defer et.mu.RUnlock()

	history, exists := et.history[epicID]
	if !exists {
		return []StateTransition{}
	}

	// Return a copy to prevent external modification
	result := make([]StateTransition, len(history))
	copy(result, history)
	return result
}

// GetRecentEvents returns recent events from the tracker
func (et *EpicTracker) GetRecentEvents(limit int) []EpicStateEvent {
	et.mu.RLock()
	defer et.mu.RUnlock()

	if limit <= 0 || limit > len(et.events) {
		limit = len(et.events)
	}

	// Return the most recent events
	start := len(et.events) - limit
	if start < 0 {
		start = 0
	}

	result := make([]EpicStateEvent, limit)
	copy(result, et.events[start:])
	return result
}

// CalculateAdvancedMetrics calculates advanced metrics for an epic
func (et *EpicTracker) CalculateAdvancedMetrics(epicID string) (*AdvancedMetrics, error) {
	et.mu.RLock()
	defer et.mu.RUnlock()

	epic, err := et.manager.GetEpic(epicID)
	if err != nil {
		return nil, fmt.Errorf("failed to get epic: %w", err)
	}

	metrics := &AdvancedMetrics{
		EpicID:       epicID,
		CalculatedAt: time.Now(),
		BasicMetrics: epic.Progress,
	}

	// Calculate duration metrics
	if epic.StartDate != nil {
		if epic.EndDate != nil {
			metrics.TotalDuration = epic.EndDate.Sub(*epic.StartDate)
		} else {
			metrics.TotalDuration = time.Since(*epic.StartDate)
		}
		metrics.DurationDays = int(metrics.TotalDuration.Hours() / 24)
	}

	// Calculate velocity and prediction metrics
	history := et.history[epicID]
	if len(history) > 0 {
		metrics.StateTransitions = len(history)
		metrics.LastTransition = &history[len(history)-1]

		// Calculate average time between transitions
		if len(history) > 1 {
			totalTime := history[len(history)-1].Timestamp.Sub(history[0].Timestamp)
			metrics.AvgTransitionTime = totalTime / time.Duration(len(history)-1)
		}
	}

	// Estimate completion if not yet completed
	if epic.Status != StatusCompleted && epic.Progress.CompletionPercentage > 0 && epic.StartDate != nil {
		elapsed := time.Since(*epic.StartDate)
		estimatedTotal := time.Duration(float64(elapsed) / (epic.Progress.CompletionPercentage / 100.0))
		estimatedEnd := epic.StartDate.Add(estimatedTotal)
		metrics.EstimatedCompletion = &estimatedEnd
	}

	return metrics, nil
}

// AdvancedMetrics contains detailed metrics about an epic
type AdvancedMetrics struct {
	EpicID              string           `json:"epic_id"`
	CalculatedAt        time.Time        `json:"calculated_at"`
	BasicMetrics        ProgressMetrics  `json:"basic_metrics"`
	TotalDuration       time.Duration    `json:"total_duration"`
	DurationDays        int              `json:"duration_days"`
	StateTransitions    int              `json:"state_transitions"`
	LastTransition      *StateTransition `json:"last_transition,omitempty"`
	AvgTransitionTime   time.Duration    `json:"avg_transition_time"`
	EstimatedCompletion *time.Time       `json:"estimated_completion,omitempty"`
}

// Subscribe adds a subscriber for state change notifications
func (et *EpicTracker) Subscribe(subscriber StateChangeSubscriber) {
	et.mu.Lock()
	defer et.mu.Unlock()
	et.subscribers = append(et.subscribers, subscriber)
}

// UpdateConfig updates the tracker configuration
func (et *EpicTracker) UpdateConfig(config TrackerConfig) {
	et.mu.Lock()
	defer et.mu.Unlock()
	et.config = config
}

// GetConfig returns the current tracker configuration
func (et *EpicTracker) GetConfig() TrackerConfig {
	et.mu.RLock()
	defer et.mu.RUnlock()
	return et.config
}

// RunPeriodicUpdates starts a background goroutine that periodically updates epics
func (et *EpicTracker) RunPeriodicUpdates(stopCh <-chan struct{}) {
	ticker := time.NewTicker(et.config.ProgressUpdateFreq)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			et.updateAllEpics()
		case <-stopCh:
			return
		}
	}
}

// updateAllEpics updates all epics based on their current state
func (et *EpicTracker) updateAllEpics() {
	collection, err := et.manager.GetEpicCollection()
	if err != nil {
		et.logEvent(EpicStateEvent{
			EpicID:      "",
			EventType:   EventValidationError,
			Timestamp:   time.Now(),
			Description: fmt.Sprintf("Failed to get epic collection for periodic update: %v", err),
		})
		return
	}

	for epicID := range collection.Epics {
		// Skip completed and cancelled epics
		epic := collection.Epics[epicID]
		if epic.Status == StatusCompleted || epic.Status == StatusCancelled {
			continue
		}

		// Update based on stories
		if err := et.UpdateEpicBasedOnStories(epicID); err != nil {
			et.logEvent(EpicStateEvent{
				EpicID:      epicID,
				EventType:   EventValidationError,
				Timestamp:   time.Now(),
				Description: fmt.Sprintf("Failed to update epic during periodic update: %v", err),
			})
		}
	}
}

// GetEpicsByStatus returns epics filtered by their current status
func (et *EpicTracker) GetEpicsByStatus(status Status) ([]*Epic, error) {
	collection, err := et.manager.GetEpicCollection()
	if err != nil {
		return nil, err
	}

	var epics []*Epic
	for _, epic := range collection.Epics {
		if epic.Status == status {
			epics = append(epics, epic)
		}
	}

	// Sort by last update time (most recently updated first)
	sort.Slice(epics, func(i, j int) bool {
		iTime, iExists := et.lastUpdate[epics[i].ID]
		jTime, jExists := et.lastUpdate[epics[j].ID]

		if !iExists && !jExists {
			return epics[i].UpdatedAt.After(epics[j].UpdatedAt)
		}
		if !iExists {
			return false
		}
		if !jExists {
			return true
		}
		return iTime.After(jTime)
	})

	return epics, nil
}
