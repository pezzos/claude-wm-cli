// Package model defines core interfaces used throughout the application.
// This documentation serves as a reference for implementing and using these interfaces.
package model

import (
	"context"
	"time"
)

// ═══════════════════════════════════════════════════════════════════════════════════
// PERSISTENCE INTERFACES
// ═══════════════════════════════════════════════════════════════════════════════════

// Repository defines the generic interface for data persistence operations.
// This interface follows the Repository pattern to abstract data access.
//
// Type Parameters:
//   T: The entity type managed by this repository
//
// Usage Example:
//   type EpicRepository Repository[Epic]
//   repo := NewRepository[Epic]()
//   epic, err := repo.Read(ctx, "epic-123")
type Repository[T any] interface {
	// Create persists a new entity and returns an error if the operation fails.
	// The entity ID should be set before calling this method.
	Create(ctx context.Context, entity T) error

	// Read retrieves an entity by its ID.
	// Returns ErrCodeNotFound if the entity doesn't exist.
	Read(ctx context.Context, id string) (T, error)

	// Update modifies an existing entity.
	// Returns ErrCodeNotFound if the entity doesn't exist.
	// Returns ErrCodeConflict if there's a concurrent modification.
	Update(ctx context.Context, id string, entity T) error

	// Delete removes an entity by its ID.
	// Returns ErrCodeNotFound if the entity doesn't exist.
	Delete(ctx context.Context, id string) error

	// List retrieves entities based on filter criteria.
	// Returns empty slice if no entities match the filter.
	List(ctx context.Context, filter Filter) ([]T, error)

	// Exists checks if an entity with the given ID exists.
	Exists(ctx context.Context, id string) (bool, error)

	// Count returns the total number of entities matching the filter.
	Count(ctx context.Context, filter Filter) (int, error)
}

// Filter defines criteria for querying entities.
// Implementations should support common filtering operations.
type Filter interface {
	// Apply returns true if the entity matches the filter criteria
	Apply(entity interface{}) bool

	// ToSQL converts the filter to SQL WHERE clause (for SQL repositories)
	ToSQL() (string, []interface{})

	// Validate checks if the filter criteria are valid
	Validate() error
}

// Transaction defines the interface for transactional operations.
// Ensures ACID properties for multi-operation workflows.
//
// Usage Example:
//   err := txManager.WithTransaction(ctx, func(tx Transaction) error {
//       return tx.Repository().Create(ctx, entity)
//   })
type Transaction interface {
	// Commit applies all changes made within the transaction
	Commit() error

	// Rollback discards all changes made within the transaction
	Rollback() error

	// Repository returns a repository instance bound to this transaction
	Repository() Repository[any]

	// IsClosed returns true if the transaction has been committed or rolled back
	IsClosed() bool
}

// ═══════════════════════════════════════════════════════════════════════════════════
// WORKFLOW INTERFACES
// ═══════════════════════════════════════════════════════════════════════════════════

// WorkflowEngine defines the interface for managing entity workflows.
// Handles state transitions, validation, and lifecycle management.
//
// Usage Example:
//   engine := NewWorkflowEngine()
//   err := engine.TransitionTo(ctx, entity, StatusInProgress)
type WorkflowEngine interface {
	// TransitionTo attempts to transition an entity to a new status
	// Returns ErrCodeWorkflowViolation if the transition is invalid
	TransitionTo(ctx context.Context, entity WorkflowState, newStatus Status) error

	// CanTransition checks if a transition is valid without performing it
	CanTransition(entity WorkflowState, newStatus Status) bool

	// GetValidTransitions returns all valid next statuses for an entity
	GetValidTransitions(entity WorkflowState) []Status

	// ValidateState checks if an entity's current state is valid
	ValidateState(entity WorkflowState) error

	// RegisterHook adds a callback for state transition events
	RegisterHook(hook WorkflowHook) error

	// UnregisterHook removes a previously registered hook
	UnregisterHook(hookID string) error
}

// WorkflowHook defines callbacks for workflow events.
// Allows customization of workflow behavior without modifying core logic.
//
// Usage Example:
//   hook := &LoggingHook{}
//   engine.RegisterHook(hook)
type WorkflowHook interface {
	// ID returns a unique identifier for this hook
	ID() string

	// BeforeTransition is called before a state transition
	// Return an error to prevent the transition
	BeforeTransition(ctx context.Context, entity WorkflowState, newStatus Status) error

	// AfterTransition is called after a successful state transition
	AfterTransition(ctx context.Context, entity WorkflowState, oldStatus, newStatus Status) error

	// OnTransitionError is called when a transition fails
	OnTransitionError(ctx context.Context, entity WorkflowState, newStatus Status, err error)
}

// StateChangeSubscriber defines the interface for components that want to be notified
// of state changes. This enables event-driven architecture patterns.
//
// Usage Example:
//   subscriber := &NotificationService{}
//   tracker.Subscribe(subscriber)
type StateChangeSubscriber interface {
	// OnEpicStateChange is called when an epic's state changes
	OnEpicStateChange(epicID string, transition StateTransition) error

	// OnStoryStateChange is called when a story's state changes  
	OnStoryStateChange(storyID string, transition StateTransition) error

	// OnTicketStateChange is called when a ticket's state changes
	OnTicketStateChange(ticketID string, transition StateTransition) error
}

// StateTransition represents a state change event with metadata.
type StateTransition struct {
	EntityID    string    `json:"entity_id"`    // ID of the entity that changed
	EntityType  string    `json:"entity_type"`  // Type of entity (epic, story, ticket)
	OldStatus   Status    `json:"old_status"`   // Previous status
	NewStatus   Status    `json:"new_status"`   // New status
	Timestamp   time.Time `json:"timestamp"`    // When the transition occurred
	TriggerBy   string    `json:"trigger_by"`   // Who or what triggered the change
	Reason      string    `json:"reason"`       // Reason for the change
	Metadata    map[string]interface{} `json:"metadata"` // Additional context
}

// ═══════════════════════════════════════════════════════════════════════════════════
// EXTERNAL INTEGRATION INTERFACES
// ═══════════════════════════════════════════════════════════════════════════════════

// GitVersionManager defines the interface for Git integration.
// Provides version control capabilities for state management.
//
// Usage Example:
//   gitManager := NewGitVersionManager()
//   err := gitManager.AutoVersionOnWrite("state.json", "update", "Updated epic status")
type GitVersionManager interface {
	// AutoVersionOnWrite creates a Git commit after file modifications
	AutoVersionOnWrite(filePath string, commitType interface{}, description string) error

	// IsEnabled returns true if Git integration is active
	IsEnabled() bool

	// GetCurrentBranch returns the name of the current Git branch
	GetCurrentBranch() (string, error)

	// CreateBranch creates a new Git branch
	CreateBranch(branchName string) error

	// SwitchBranch switches to an existing Git branch
	SwitchBranch(branchName string) error

	// GetCommitHistory returns recent commits for a file
	GetCommitHistory(filePath string, limit int) ([]GitCommit, error)

	// RestoreFile restores a file to a specific commit
	RestoreFile(filePath string, commitHash string) error
}

// GitCommit represents a Git commit with metadata.
type GitCommit struct {
	Hash      string    `json:"hash"`       // Commit SHA hash
	Message   string    `json:"message"`    // Commit message
	Author    string    `json:"author"`     // Commit author
	Timestamp time.Time `json:"timestamp"`  // Commit timestamp
	Files     []string  `json:"files"`      // Modified files
}

// LockManager defines the interface for file locking to prevent concurrent access.
// Ensures data integrity in multi-process environments.
//
// Usage Example:
//   lockManager := NewLockManager()
//   release, err := lockManager.LockFile("state.json", LockOptions{Timeout: 30*time.Second})
//   defer release()
type LockManager interface {
	// LockFile acquires an exclusive lock on a file
	// Returns a release function and any error
	LockFile(filePath string, options interface{}) (interface{}, interface{})

	// UnlockFile releases a lock on a file
	UnlockFile(filePath string) error

	// IsLocked checks if a file is currently locked
	IsLocked(filePath string) bool

	// ListLocks returns all currently held locks
	ListLocks() ([]string, error)

	// CleanupStale removes locks that are no longer valid
	CleanupStale() error
}

// ═══════════════════════════════════════════════════════════════════════════════════
// OBSERVABILITY INTERFACES  
// ═══════════════════════════════════════════════════════════════════════════════════

// MetricsCollector defines the interface for collecting application metrics.
// Enables monitoring and observability of application behavior.
//
// Usage Example:
//   metrics := NewMetricsCollector()
//   metrics.IncrementCounter("operations.created", map[string]string{"type": "epic"})
type MetricsCollector interface {
	// IncrementCounter increments a counter metric
	IncrementCounter(name string, labels map[string]string)

	// RecordHistogram records a value in a histogram metric
	RecordHistogram(name string, value float64, labels map[string]string)

	// SetGauge sets the value of a gauge metric
	SetGauge(name string, value float64, labels map[string]string)

	// RecordDuration records the duration of an operation
	RecordDuration(name string, duration time.Duration, labels map[string]string)

	// Flush ensures all metrics are sent to the backend
	Flush() error
}

// Logger defines the interface for structured logging.
// Provides consistent logging across all application components.
//
// Usage Example:
//   logger := NewLogger()
//   logger.Info("Operation completed", "operation", "create_epic", "duration", "1.2s")
type Logger interface {
	// Debug logs debug-level messages
	Debug(msg string, keysAndValues ...interface{})

	// Info logs info-level messages  
	Info(msg string, keysAndValues ...interface{})

	// Warn logs warning-level messages
	Warn(msg string, keysAndValues ...interface{})

	// Error logs error-level messages
	Error(msg string, keysAndValues ...interface{})

	// WithFields returns a logger with additional fields
	WithFields(fields map[string]interface{}) Logger

	// WithContext returns a logger that includes context information
	WithContext(ctx context.Context) Logger
}

// HealthChecker defines the interface for health check operations.
// Enables monitoring of service health and dependencies.
//
// Usage Example:
//   checker := NewHealthChecker()
//   status := checker.CheckHealth(ctx)
type HealthChecker interface {
	// CheckHealth performs a health check and returns the current status
	CheckHealth(ctx context.Context) HealthStatus

	// RegisterCheck adds a new health check
	RegisterCheck(name string, check HealthCheck) error

	// UnregisterCheck removes a health check
	UnregisterCheck(name string) error

	// ListChecks returns all registered health checks
	ListChecks() []string
}

// HealthCheck defines a single health check operation.
type HealthCheck interface {
	// Name returns the name of this health check
	Name() string

	// Check performs the health check operation
	Check(ctx context.Context) HealthResult
}

// HealthResult represents the result of a health check.
type HealthResult struct {
	Name      string                 `json:"name"`       // Check name
	Status    HealthStatus          `json:"status"`     // Check status
	Message   string                `json:"message"`    // Status message
	Details   map[string]interface{} `json:"details"`   // Additional details
	Duration  time.Duration         `json:"duration"`   // Check duration
	Timestamp time.Time             `json:"timestamp"`  // Check timestamp
}

// HealthStatus represents the overall health status.
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"   // All checks passing
	HealthStatusDegraded  HealthStatus = "degraded"  // Some non-critical checks failing
	HealthStatusUnhealthy HealthStatus = "unhealthy" // Critical checks failing
)

// ═══════════════════════════════════════════════════════════════════════════════════
// PLUGIN INTERFACES
// ═══════════════════════════════════════════════════════════════════════════════════

// Plugin defines the interface for application plugins.
// Enables extensibility through third-party integrations.
//
// Usage Example:
//   plugin := &MyCustomPlugin{}
//   registry.Register(plugin)
type Plugin interface {
	// Name returns the plugin name
	Name() string

	// Version returns the plugin version
	Version() string

	// Description returns a human-readable description
	Description() string

	// Initialize initializes the plugin with configuration
	Initialize(config map[string]interface{}) error

	// Execute executes the plugin with given arguments
	Execute(ctx context.Context, args map[string]interface{}) (interface{}, error)

	// Shutdown cleanly shuts down the plugin
	Shutdown(ctx context.Context) error

	// HealthCheck returns the plugin's health status
	HealthCheck(ctx context.Context) HealthResult
}

// PluginRegistry defines the interface for managing plugins.
// Provides plugin lifecycle management and discovery.
//
// Usage Example:
//   registry := NewPluginRegistry()
//   err := registry.Register(plugin)
//   result, err := registry.Execute(ctx, "my-plugin", args)
type PluginRegistry interface {
	// Register registers a new plugin
	Register(plugin Plugin) error

	// Unregister removes a plugin
	Unregister(name string) error

	// Get retrieves a plugin by name
	Get(name string) (Plugin, error)

	// List returns all registered plugins
	List() []Plugin

	// Execute executes a plugin by name
	Execute(ctx context.Context, name string, args map[string]interface{}) (interface{}, error)

	// Shutdown shuts down all plugins
	Shutdown(ctx context.Context) error
}

// ═══════════════════════════════════════════════════════════════════════════════════
// CONFIGURATION INTERFACES
// ═══════════════════════════════════════════════════════════════════════════════════

// ConfigProvider defines the interface for configuration management.
// Provides hierarchical configuration with hot-reload support.
//
// Usage Example:
//   config := NewConfigProvider()
//   timeout := config.GetDuration("database.timeout", 30*time.Second)
type ConfigProvider interface {
	// Get retrieves a configuration value
	Get(key string) interface{}

	// GetString retrieves a string configuration value
	GetString(key string, defaultValue string) string

	// GetInt retrieves an integer configuration value
	GetInt(key string, defaultValue int) int

	// GetBool retrieves a boolean configuration value
	GetBool(key string, defaultValue bool) bool

	// GetDuration retrieves a duration configuration value
	GetDuration(key string, defaultValue time.Duration) time.Duration

	// Set sets a configuration value
	Set(key string, value interface{}) error

	// Watch watches for changes to a configuration key
	Watch(key string, callback func(interface{})) error

	// Reload reloads configuration from sources
	Reload() error

	// Validate validates the current configuration
	Validate() error
}

// ═══════════════════════════════════════════════════════════════════════════════════
// INTERFACE IMPLEMENTATION GUIDELINES
// ═══════════════════════════════════════════════════════════════════════════════════

/*
IMPLEMENTATION GUIDELINES:

1. Error Handling:
   - Always return CLIError types from interface methods
   - Use appropriate error codes (ErrCodeNotFound, ErrCodeValidation, etc.)
   - Include context and suggestions in errors

2. Context Usage:
   - Always accept context.Context as the first parameter
   - Respect context cancellation and timeouts
   - Pass context through to downstream operations

3. Thread Safety:
   - All interface implementations should be thread-safe
   - Use appropriate synchronization primitives
   - Document any non-thread-safe behavior clearly

4. Resource Management:
   - Implement proper cleanup in Shutdown() methods
   - Use defer statements for resource cleanup
   - Handle partial failures gracefully

5. Observability:
   - Log important operations and errors
   - Emit metrics for performance monitoring
   - Include tracing information where appropriate

6. Testing:
   - Create mock implementations for testing
   - Use dependency injection for testability
   - Write comprehensive unit tests

7. Documentation:
   - Document all public methods and types
   - Include usage examples in comments
   - Specify error conditions and return values

8. Versioning:
   - Use semantic versioning for interface changes
   - Maintain backward compatibility when possible
   - Document breaking changes clearly
*/