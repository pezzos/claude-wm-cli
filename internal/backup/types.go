package backup

import (
	"fmt"
	"time"
)

// BackupType represents the type of backup operation
type BackupType string

const (
	BackupTypeAutomatic BackupType = "automatic" // Automatic backup before state changes
	BackupTypeManual    BackupType = "manual"    // Manual backup requested by user
	BackupTypeEmergency BackupType = "emergency" // Emergency backup due to corruption
	BackupTypeSnapshot  BackupType = "snapshot"  // Periodic snapshot backup
)

func (bt BackupType) String() string {
	return string(bt)
}

// BackupReason represents why a backup was created
type BackupReason string

const (
	ReasonPreWrite    BackupReason = "pre_write"    // Before writing state
	ReasonUserRequest BackupReason = "user_request" // User explicitly requested
	ReasonCorruption  BackupReason = "corruption"   // Corruption detected
	ReasonMigration   BackupReason = "migration"    // Schema migration
	ReasonScheduled   BackupReason = "scheduled"    // Scheduled backup
	ReasonPreRecovery BackupReason = "pre_recovery" // Before recovery operation
)

func (br BackupReason) String() string {
	return string(br)
}

// BackupStatus represents the status of a backup
type BackupStatus string

const (
	BackupStatusCreating  BackupStatus = "creating"  // Backup in progress
	BackupStatusCompleted BackupStatus = "completed" // Backup successfully completed
	BackupStatusFailed    BackupStatus = "failed"    // Backup failed
	BackupStatusCorrupted BackupStatus = "corrupted" // Backup file is corrupted
	BackupStatusVerified  BackupStatus = "verified"  // Backup integrity verified
)

func (bs BackupStatus) String() string {
	return string(bs)
}

// BackupMetadata contains information about a backup
type BackupMetadata struct {
	ID             string        `json:"id"`              // Unique backup identifier
	SourceFile     string        `json:"source_file"`     // Original file path
	BackupFile     string        `json:"backup_file"`     // Backup file path
	Type           BackupType    `json:"type"`            // Type of backup
	Reason         BackupReason  `json:"reason"`          // Why backup was created
	Status         BackupStatus  `json:"status"`          // Current status
	CreatedAt      time.Time     `json:"created_at"`      // When backup was created
	CompletedAt    *time.Time    `json:"completed_at"`    // When backup completed
	Duration       time.Duration `json:"duration"`        // Time taken to create backup
	SourceSize     int64         `json:"source_size"`     // Original file size
	BackupSize     int64         `json:"backup_size"`     // Backup file size
	Compressed     bool          `json:"compressed"`      // Whether backup is compressed
	SourceChecksum string        `json:"source_checksum"` // Original file checksum
	BackupChecksum string        `json:"backup_checksum"` // Backup file checksum
	IntegrityCheck bool          `json:"integrity_check"` // Whether integrity was verified
	ErrorMessage   string        `json:"error_message"`   // Error message if failed
	Tags           []string      `json:"tags"`            // Additional tags
	CreatedBy      string        `json:"created_by"`      // Process/user that created backup
	Version        string        `json:"version"`         // Backup format version
}

// IsValid checks if the backup metadata is valid
func (bm *BackupMetadata) IsValid() bool {
	return bm.ID != "" && bm.SourceFile != "" && bm.BackupFile != "" &&
		!bm.CreatedAt.IsZero() && bm.SourceChecksum != ""
}

// Age returns how old the backup is
func (bm *BackupMetadata) Age() time.Duration {
	return time.Since(bm.CreatedAt)
}

// IsCompleted returns true if backup completed successfully
func (bm *BackupMetadata) IsCompleted() bool {
	return bm.Status == BackupStatusCompleted || bm.Status == BackupStatusVerified
}

// BackupConfig contains configuration for backup operations
type BackupConfig struct {
	Enabled          bool          `json:"enabled"`           // Whether backup is enabled
	BackupDirectory  string        `json:"backup_directory"`  // Directory to store backups
	MaxBackups       int           `json:"max_backups"`       // Maximum backups per file
	MaxAge           time.Duration `json:"max_age"`           // Maximum age of backups
	MaxTotalSize     int64         `json:"max_total_size"`    // Maximum total size of all backups
	CompressionLevel int           `json:"compression_level"` // Compression level (0-9)
	AutoBackup       bool          `json:"auto_backup"`       // Enable automatic backups
	VerifyIntegrity  bool          `json:"verify_integrity"`  // Verify backup integrity
	AsyncBackup      bool          `json:"async_backup"`      // Perform backups asynchronously
	CleanupInterval  time.Duration `json:"cleanup_interval"`  // How often to clean old backups
	BackupFormat     string        `json:"backup_format"`     // Backup format (copy, tar, etc.)
	IncludeMetadata  bool          `json:"include_metadata"`  // Include metadata in backup
}

// DefaultBackupConfig returns default backup configuration
func DefaultBackupConfig() *BackupConfig {
	return &BackupConfig{
		Enabled:          true,
		BackupDirectory:  ".backups",
		MaxBackups:       10,
		MaxAge:           30 * 24 * time.Hour, // 30 days
		MaxTotalSize:     100 * 1024 * 1024,   // 100MB
		CompressionLevel: 6,                   // Moderate compression
		AutoBackup:       true,
		VerifyIntegrity:  true,
		AsyncBackup:      false, // Synchronous by default for safety
		CleanupInterval:  24 * time.Hour,
		BackupFormat:     "copy",
		IncludeMetadata:  true,
	}
}

// RetentionPolicy defines how backups are retained and cleaned up
type RetentionPolicy struct {
	Strategy      RetentionStrategy `json:"strategy"`       // Retention strategy
	MaxCount      int               `json:"max_count"`      // Maximum number of backups
	MaxAge        time.Duration     `json:"max_age"`        // Maximum age of backups
	MaxSize       int64             `json:"max_size"`       // Maximum total size
	KeepDaily     int               `json:"keep_daily"`     // Keep daily backups for N days
	KeepWeekly    int               `json:"keep_weekly"`    // Keep weekly backups for N weeks
	KeepMonthly   int               `json:"keep_monthly"`   // Keep monthly backups for N months
	KeepImportant bool              `json:"keep_important"` // Keep backups tagged as important
}

// RetentionStrategy represents different retention strategies
type RetentionStrategy string

const (
	RetentionSimple       RetentionStrategy = "simple"       // Simple count/age based
	RetentionGenerational RetentionStrategy = "generational" // GFS (Grandfather-Father-Son)
	RetentionSizeBasedage RetentionStrategy = "size_based"   // Based on total size
	RetentionSmart        RetentionStrategy = "smart"        // Smart retention based on importance
)

// DefaultRetentionPolicy returns default retention policy
func DefaultRetentionPolicy() *RetentionPolicy {
	return &RetentionPolicy{
		Strategy:      RetentionSimple,
		MaxCount:      10,
		MaxAge:        30 * 24 * time.Hour,
		MaxSize:       100 * 1024 * 1024,
		KeepDaily:     7,
		KeepWeekly:    4,
		KeepMonthly:   3,
		KeepImportant: true,
	}
}

// BackupRequest represents a request to create a backup
type BackupRequest struct {
	SourceFile  string       `json:"source_file"` // File to backup
	Type        BackupType   `json:"type"`        // Type of backup
	Reason      BackupReason `json:"reason"`      // Reason for backup
	Tags        []string     `json:"tags"`        // Additional tags
	Compress    bool         `json:"compress"`    // Whether to compress
	Verify      bool         `json:"verify"`      // Whether to verify integrity
	Async       bool         `json:"async"`       // Whether to backup asynchronously
	Priority    int          `json:"priority"`    // Backup priority (0-10)
	Description string       `json:"description"` // Human-readable description
	Force       bool         `json:"force"`       // Force backup even if recent backup exists
}

// BackupResult contains the result of a backup operation
type BackupResult struct {
	Success    bool            `json:"success"`     // Whether backup succeeded
	Metadata   *BackupMetadata `json:"metadata"`    // Backup metadata
	Error      error           `json:"error"`       // Error if backup failed
	Duration   time.Duration   `json:"duration"`    // Time taken
	BytesTotal int64           `json:"bytes_total"` // Total bytes processed
	Skipped    bool            `json:"skipped"`     // Whether backup was skipped
	Reason     string          `json:"reason"`      // Reason for skip/failure
	Timestamp  time.Time       `json:"timestamp"`   // When operation completed
}

// RecoveryRequest represents a request to recover from backup
type RecoveryRequest struct {
	SourceFile   string        `json:"source_file"`   // File to recover
	BackupID     string        `json:"backup_id"`     // Specific backup to restore from
	BackupTime   *time.Time    `json:"backup_time"`   // Restore from backup at specific time
	VerifyBefore bool          `json:"verify_before"` // Verify backup before restore
	VerifyAfter  bool          `json:"verify_after"`  // Verify restored file
	CreateBackup bool          `json:"create_backup"` // Create backup before recovery
	Force        bool          `json:"force"`         // Force recovery even if current file is good
	RestorePath  string        `json:"restore_path"`  // Alternative restore path
	RestoreMode  RestoreMode   `json:"restore_mode"`  // How to restore
	Timeout      time.Duration `json:"timeout"`       // Recovery timeout
}

// RestoreMode represents different ways to restore from backup
type RestoreMode string

const (
	RestoreModeReplace RestoreMode = "replace" // Replace existing file
	RestoreModeRename  RestoreMode = "rename"  // Rename existing file and restore
	RestoreModeMerge   RestoreMode = "merge"   // Attempt to merge changes
	RestoreModePreview RestoreMode = "preview" // Preview what would be restored
)

// RecoveryResult contains the result of a recovery operation
type RecoveryResult struct {
	Success        bool            `json:"success"`         // Whether recovery succeeded
	RestoredFile   string          `json:"restored_file"`   // Path of restored file
	BackupUsed     *BackupMetadata `json:"backup_used"`     // Backup that was used
	BackupCreated  *BackupMetadata `json:"backup_created"`  // Backup created before recovery
	Error          error           `json:"error"`           // Error if recovery failed
	Duration       time.Duration   `json:"duration"`        // Time taken
	BytesRestored  int64           `json:"bytes_restored"`  // Bytes restored
	IntegrityCheck bool            `json:"integrity_check"` // Whether integrity was verified
	Changes        []string        `json:"changes"`         // List of changes made
	Warnings       []string        `json:"warnings"`        // Warnings during recovery
	Timestamp      time.Time       `json:"timestamp"`       // When operation completed
}

// BackupEvent represents an event in the backup system
type BackupEvent struct {
	Type       BackupEventType `json:"type"`        // Type of event
	SourceFile string          `json:"source_file"` // File involved
	BackupID   string          `json:"backup_id"`   // Backup involved
	Message    string          `json:"message"`     // Event message
	Details    string          `json:"details"`     // Additional details
	Error      string          `json:"error"`       // Error message if applicable
	Duration   time.Duration   `json:"duration"`    // Operation duration
	Timestamp  time.Time       `json:"timestamp"`   // When event occurred
}

// BackupEventType represents different types of backup events
type BackupEventType string

const (
	EventBackupStarted      BackupEventType = "backup_started"
	EventBackupCompleted    BackupEventType = "backup_completed"
	EventBackupFailed       BackupEventType = "backup_failed"
	EventBackupSkipped      BackupEventType = "backup_skipped"
	EventRecoveryStarted    BackupEventType = "recovery_started"
	EventRecoveryCompleted  BackupEventType = "recovery_completed"
	EventRecoveryFailed     BackupEventType = "recovery_failed"
	EventCleanupStarted     BackupEventType = "cleanup_started"
	EventCleanupCompleted   BackupEventType = "cleanup_completed"
	EventIntegrityCheck     BackupEventType = "integrity_check"
	EventCorruptionDetected BackupEventType = "corruption_detected"
)

// BackupStats contains statistics about backup operations
type BackupStats struct {
	TotalBackups      int64         `json:"total_backups"`       // Total backups created
	TotalRecoveries   int64         `json:"total_recoveries"`    // Total recoveries performed
	TotalSize         int64         `json:"total_size"`          // Total size of all backups
	SuccessfulBackups int64         `json:"successful_backups"`  // Successful backup operations
	FailedBackups     int64         `json:"failed_backups"`      // Failed backup operations
	SkippedBackups    int64         `json:"skipped_backups"`     // Skipped backup operations
	AverageBackupTime time.Duration `json:"average_backup_time"` // Average backup duration
	LastBackupTime    time.Time     `json:"last_backup_time"`    // When last backup was created
	LastCleanupTime   time.Time     `json:"last_cleanup_time"`   // When last cleanup was performed
	OldestBackup      time.Time     `json:"oldest_backup"`       // Oldest backup timestamp
	CompressionRatio  float64       `json:"compression_ratio"`   // Average compression ratio
}

// BackupError represents a backup-specific error
type BackupError struct {
	Operation   string    `json:"operation"`   // Operation that failed
	SourceFile  string    `json:"source_file"` // File being backed up
	BackupFile  string    `json:"backup_file"` // Backup file path
	BackupID    string    `json:"backup_id"`   // Backup identifier
	Cause       error     `json:"cause"`       // Underlying error
	Suggestion  string    `json:"suggestion"`  // Suggested action
	Recoverable bool      `json:"recoverable"` // Whether error is recoverable
	Timestamp   time.Time `json:"timestamp"`   // When error occurred
}

func (e *BackupError) Error() string {
	msg := fmt.Sprintf("backup %s failed for %s", e.Operation, e.SourceFile)
	if e.Cause != nil {
		msg += fmt.Sprintf(": %v", e.Cause)
	}
	if e.Suggestion != "" {
		msg += fmt.Sprintf(" (suggestion: %s)", e.Suggestion)
	}
	return msg
}

func (e *BackupError) Unwrap() error {
	return e.Cause
}

// NewBackupError creates a new backup error
func NewBackupError(operation, sourceFile string, cause error) *BackupError {
	return &BackupError{
		Operation:   operation,
		SourceFile:  sourceFile,
		Cause:       cause,
		Timestamp:   time.Now(),
		Recoverable: true,
	}
}

// BackupFilter contains criteria for filtering backups
type BackupFilter struct {
	SourceFile    string       `json:"source_file,omitempty"`    // Filter by source file
	Type          BackupType   `json:"type,omitempty"`           // Filter by backup type
	Reason        BackupReason `json:"reason,omitempty"`         // Filter by reason
	Status        BackupStatus `json:"status,omitempty"`         // Filter by status
	CreatedAfter  *time.Time   `json:"created_after,omitempty"`  // Filter by creation time
	CreatedBefore *time.Time   `json:"created_before,omitempty"` // Filter by creation time
	Tags          []string     `json:"tags,omitempty"`           // Filter by tags
	MinSize       int64        `json:"min_size,omitempty"`       // Minimum backup size
	MaxSize       int64        `json:"max_size,omitempty"`       // Maximum backup size
	Verified      *bool        `json:"verified,omitempty"`       // Filter by verification status
	Limit         int          `json:"limit,omitempty"`          // Limit number of results
	SortBy        string       `json:"sort_by,omitempty"`        // Sort field
	SortOrder     string       `json:"sort_order,omitempty"`     // Sort order (asc/desc)
}
