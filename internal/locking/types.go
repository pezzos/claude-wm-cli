package locking

import (
	"fmt"
	"time"
)

// LockType represents the type of file lock
type LockType int

const (
	LockExclusive LockType = iota // Exclusive lock - only one process can hold
	LockShared                    // Shared lock - multiple readers allowed
)

func (lt LockType) String() string {
	switch lt {
	case LockExclusive:
		return "exclusive"
	case LockShared:
		return "shared"
	default:
		return "unknown"
	}
}

// LockOptions configures file locking behavior
type LockOptions struct {
	Type         LockType      `json:"type"`          // Type of lock to acquire
	Timeout      time.Duration `json:"timeout"`       // Maximum time to wait for lock
	NonBlocking  bool          `json:"non_blocking"`  // Whether to fail immediately if lock unavailable
	StaleTimeout time.Duration `json:"stale_timeout"` // Time after which locks are considered stale
	RetryDelay   time.Duration `json:"retry_delay"`   // Delay between lock acquisition attempts
}

// DefaultLockOptions returns default locking configuration
func DefaultLockOptions() *LockOptions {
	return &LockOptions{
		Type:         LockExclusive,
		Timeout:      30 * time.Second, // Proven 30s timeout pattern
		NonBlocking:  false,
		StaleTimeout: 5 * time.Minute, // Locks older than 5 minutes are stale
		RetryDelay:   100 * time.Millisecond,
	}
}

// LockInfo contains information about an active lock
type LockInfo struct {
	FilePath    string    `json:"file_path"`    // Path to the locked file
	LockPath    string    `json:"lock_path"`    // Path to the lock file
	Type        LockType  `json:"type"`         // Type of lock
	PID         int       `json:"pid"`          // Process ID that owns the lock
	Hostname    string    `json:"hostname"`     // Hostname where lock was created
	AcquiredAt  time.Time `json:"acquired_at"`  // When the lock was acquired
	ExpiresAt   time.Time `json:"expires_at"`   // When the lock expires
	ProcessName string    `json:"process_name"` // Name of the process holding the lock
	UserInfo    string    `json:"user_info"`    // User information
}

// IsStale checks if the lock is considered stale
func (li *LockInfo) IsStale() bool {
	return time.Now().After(li.ExpiresAt)
}

// Age returns how long the lock has been held
func (li *LockInfo) Age() time.Duration {
	return time.Since(li.AcquiredAt)
}

// LockStatus represents the current status of a lock attempt
type LockStatus int

const (
	LockStatusAvailable LockStatus = iota // Lock is available and can be acquired
	LockStatusHeld                        // Lock is held by current process
	LockStatusBlocked                     // Lock is held by another process
	LockStatusStale                       // Lock exists but is stale
	LockStatusError                       // Error occurred during lock operation
)

func (ls LockStatus) String() string {
	switch ls {
	case LockStatusAvailable:
		return "available"
	case LockStatusHeld:
		return "held"
	case LockStatusBlocked:
		return "blocked"
	case LockStatusStale:
		return "stale"
	case LockStatusError:
		return "error"
	default:
		return "unknown"
	}
}

// LockResult contains the result of a lock operation
type LockResult struct {
	Status     LockStatus    `json:"status"`
	LockInfo   *LockInfo     `json:"lock_info,omitempty"`
	Error      error         `json:"error,omitempty"`
	Duration   time.Duration `json:"duration"`
	Attempts   int           `json:"attempts"`
	Message    string        `json:"message"`
	Suggestion string        `json:"suggestion,omitempty"`
	Timestamp  time.Time     `json:"timestamp"`
}

// LockError represents a locking-specific error
type LockError struct {
	Operation   string    `json:"operation"`              // Operation that failed
	FilePath    string    `json:"file_path"`              // File being locked
	LockPath    string    `json:"lock_path"`              // Lock file path
	LockType    LockType  `json:"lock_type"`              // Type of lock attempted
	Cause       error     `json:"cause"`                  // Underlying error
	CurrentLock *LockInfo `json:"current_lock,omitempty"` // Info about existing lock
	Suggestion  string    `json:"suggestion"`             // Suggested action
	Recoverable bool      `json:"recoverable"`            // Whether the error is recoverable
	Timestamp   time.Time `json:"timestamp"`
}

func (e *LockError) Error() string {
	msg := fmt.Sprintf("lock %s failed for %s", e.Operation, e.FilePath)
	if e.Cause != nil {
		msg += fmt.Sprintf(": %v", e.Cause)
	}
	if e.Suggestion != "" {
		msg += fmt.Sprintf(" (suggestion: %s)", e.Suggestion)
	}
	return msg
}

func (e *LockError) Unwrap() error {
	return e.Cause
}

// NewLockError creates a new lock error
func NewLockError(operation, filePath string, cause error) *LockError {
	return &LockError{
		Operation:   operation,
		FilePath:    filePath,
		Cause:       cause,
		Timestamp:   time.Now(),
		Recoverable: true,
	}
}

// LockConflictError represents a lock conflict
type LockConflictError struct {
	RequestedType LockType  `json:"requested_type"`
	ExistingLock  *LockInfo `json:"existing_lock"`
	FilePath      string    `json:"file_path"`
	Timestamp     time.Time `json:"timestamp"`
}

func (e *LockConflictError) Error() string {
	return fmt.Sprintf("cannot acquire %s lock on %s: %s lock held by PID %d since %s",
		e.RequestedType, e.FilePath, e.ExistingLock.Type,
		e.ExistingLock.PID, e.ExistingLock.AcquiredAt.Format("15:04:05"))
}

// LockTimeoutError represents a lock timeout
type LockTimeoutError struct {
	FilePath string        `json:"file_path"`
	LockType LockType      `json:"lock_type"`
	Timeout  time.Duration `json:"timeout"`
	Attempts int           `json:"attempts"`
}

func (e *LockTimeoutError) Error() string {
	return fmt.Sprintf("timeout acquiring %s lock on %s after %v (%d attempts)",
		e.LockType, e.FilePath, e.Timeout, e.Attempts)
}

// LockConfig contains global locking configuration
type LockConfig struct {
	Enabled         bool          `json:"enabled"`          // Whether file locking is enabled
	DefaultTimeout  time.Duration `json:"default_timeout"`  // Default lock timeout
	StaleTimeout    time.Duration `json:"stale_timeout"`    // When locks become stale
	RetryDelay      time.Duration `json:"retry_delay"`      // Delay between retries
	MaxRetries      int           `json:"max_retries"`      // Maximum retry attempts
	CleanupInterval time.Duration `json:"cleanup_interval"` // How often to clean stale locks
	LockDirectory   string        `json:"lock_directory"`   // Directory for lock files
}

// DefaultLockConfig returns default locking configuration
func DefaultLockConfig() *LockConfig {
	return &LockConfig{
		Enabled:         true,
		DefaultTimeout:  30 * time.Second,
		StaleTimeout:    5 * time.Minute,
		RetryDelay:      100 * time.Millisecond,
		MaxRetries:      300, // 30 seconds with 100ms delay
		CleanupInterval: 1 * time.Minute,
		LockDirectory:   "", // Use same directory as target file
	}
}

// Priority levels for lock operations
type LockPriority int

const (
	LockPriorityNormal LockPriority = iota
	LockPriorityHigh
	LockPriorityCritical
)

func (lp LockPriority) String() string {
	switch lp {
	case LockPriorityNormal:
		return "normal"
	case LockPriorityHigh:
		return "high"
	case LockPriorityCritical:
		return "critical"
	default:
		return "unknown"
	}
}

// LockRequest represents a lock acquisition request
type LockRequest struct {
	FilePath    string        `json:"file_path"`
	Type        LockType      `json:"type"`
	Priority    LockPriority  `json:"priority"`
	Timeout     time.Duration `json:"timeout"`
	NonBlocking bool          `json:"non_blocking"`
	Context     string        `json:"context"`      // Description of why lock is needed
	ProcessInfo string        `json:"process_info"` // Information about requesting process
}

// LockMetrics contains statistics about locking operations
type LockMetrics struct {
	TotalRequests     int64         `json:"total_requests"`
	SuccessfulLocks   int64         `json:"successful_locks"`
	FailedLocks       int64         `json:"failed_locks"`
	TimeoutErrors     int64         `json:"timeout_errors"`
	ConflictErrors    int64         `json:"conflict_errors"`
	StaleLocksCleared int64         `json:"stale_locks_cleared"`
	AverageLockTime   time.Duration `json:"average_lock_time"`
	MaxLockTime       time.Duration `json:"max_lock_time"`
	ActiveLocks       int           `json:"active_locks"`
	LastCleanup       time.Time     `json:"last_cleanup"`
}

// LockEventType represents different types of lock events
type LockEventType string

const (
	LockEventAcquired LockEventType = "acquired"
	LockEventReleased LockEventType = "released"
	LockEventTimeout  LockEventType = "timeout"
	LockEventConflict LockEventType = "conflict"
	LockEventStale    LockEventType = "stale"
	LockEventError    LockEventType = "error"
)

// LockEvent represents a lock-related event for logging/monitoring
type LockEvent struct {
	Type      LockEventType `json:"type"`
	FilePath  string        `json:"file_path"`
	LockType  LockType      `json:"lock_type"`
	PID       int           `json:"pid"`
	Duration  time.Duration `json:"duration"`
	Error     string        `json:"error,omitempty"`
	Details   string        `json:"details,omitempty"`
	Timestamp time.Time     `json:"timestamp"`
}
