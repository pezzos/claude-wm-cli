package backup

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"sync"
	"time"
)

// Manager manages backup and recovery operations for state files
type Manager struct {
	config        *BackupConfig
	retention     *RetentionPolicy
	backupDir     string
	metadataFile  string
	backups       map[string]*BackupMetadata
	events        []BackupEvent
	stats         *BackupStats
	mu            sync.RWMutex
	eventHandlers []func(BackupEvent)
}

// NewManager creates a new backup manager with the given configuration
func NewManager(config *BackupConfig) (*Manager, error) {
	if config == nil {
		config = DefaultBackupConfig()
	}

	backupDir := config.BackupDirectory
	if !filepath.IsAbs(backupDir) {
		workingDir, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get working directory: %w", err)
		}
		backupDir = filepath.Join(workingDir, backupDir)
	}

	// Ensure backup directory exists
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create backup directory %s: %w", backupDir, err)
	}

	metadataFile := filepath.Join(backupDir, "backups.json")

	manager := &Manager{
		config:       config,
		retention:    DefaultRetentionPolicy(),
		backupDir:    backupDir,
		metadataFile: metadataFile,
		backups:      make(map[string]*BackupMetadata),
		events:       make([]BackupEvent, 0),
		stats:        &BackupStats{},
		mu:           sync.RWMutex{},
	}

	// Load existing metadata
	if err := manager.loadMetadata(); err != nil {
		return nil, fmt.Errorf("failed to load backup metadata: %w", err)
	}

	return manager, nil
}

// CreateBackup creates a backup of the specified file
func (m *Manager) CreateBackup(request *BackupRequest) (*BackupResult, error) {
	if !m.config.Enabled {
		return &BackupResult{
			Success:   false,
			Skipped:   true,
			Reason:    "backup disabled",
			Timestamp: time.Now(),
		}, nil
	}

	startTime := time.Now()
	backupID := m.generateBackupID(request.SourceFile)

	// Check if source file exists
	if _, err := os.Stat(request.SourceFile); err != nil {
		return &BackupResult{
			Success:   false,
			Error:     fmt.Errorf("source file does not exist: %w", err),
			Timestamp: time.Now(),
		}, nil
	}

	// Check if we should skip this backup
	if !request.Force && m.shouldSkipBackup(request.SourceFile, request.Type) {
		return &BackupResult{
			Success:   false,
			Skipped:   true,
			Reason:    "recent backup exists",
			Timestamp: time.Now(),
		}, nil
	}

	// Emit start event
	m.emitEvent(BackupEvent{
		Type:       EventBackupStarted,
		SourceFile: request.SourceFile,
		BackupID:   backupID,
		Message:    fmt.Sprintf("Starting %s backup", request.Type),
		Timestamp:  startTime,
	})

	// Create backup metadata
	metadata := &BackupMetadata{
		ID:         backupID,
		SourceFile: request.SourceFile,
		BackupFile: m.generateBackupPath(request.SourceFile, backupID),
		Type:       request.Type,
		Reason:     request.Reason,
		Status:     BackupStatusCreating,
		CreatedAt:  startTime,
		Tags:       request.Tags,
		CreatedBy:  "claude-wm-cli",
		Version:    "1.0",
		Compressed: request.Compress,
	}

	// Calculate source file checksum and size
	sourceChecksum, sourceSize, err := m.calculateFileInfo(request.SourceFile)
	if err != nil {
		m.emitFailureEvent(request.SourceFile, backupID, err)
		return &BackupResult{
			Success:   false,
			Error:     fmt.Errorf("failed to calculate source file info: %w", err),
			Duration:  time.Since(startTime),
			Timestamp: time.Now(),
		}, nil
	}

	metadata.SourceChecksum = sourceChecksum
	metadata.SourceSize = sourceSize

	// Perform the actual backup
	backupChecksum, backupSize, err := m.performBackup(request.SourceFile, metadata.BackupFile, request.Compress)
	if err != nil {
		// Clean up partial backup file
		os.Remove(metadata.BackupFile)
		m.emitFailureEvent(request.SourceFile, backupID, err)
		return &BackupResult{
			Success:   false,
			Error:     fmt.Errorf("backup operation failed: %w", err),
			Duration:  time.Since(startTime),
			Timestamp: time.Now(),
		}, nil
	}

	metadata.BackupChecksum = backupChecksum
	metadata.BackupSize = backupSize
	metadata.Duration = time.Since(startTime)

	// Verify integrity if requested
	if request.Verify {
		if err := m.verifyBackupIntegrity(metadata); err != nil {
			os.Remove(metadata.BackupFile)
			m.emitFailureEvent(request.SourceFile, backupID, err)
			return &BackupResult{
				Success:   false,
				Error:     fmt.Errorf("backup verification failed: %w", err),
				Duration:  time.Since(startTime),
				Timestamp: time.Now(),
			}, nil
		}
		metadata.IntegrityCheck = true
		metadata.Status = BackupStatusVerified
	} else {
		metadata.Status = BackupStatusCompleted
	}

	completedAt := time.Now()
	metadata.CompletedAt = &completedAt
	metadata.Duration = completedAt.Sub(startTime)

	// Store metadata
	m.mu.Lock()
	m.backups[backupID] = metadata
	m.updateStats(metadata, true)
	m.mu.Unlock()

	// Save metadata to disk
	if err := m.saveMetadata(); err != nil {
		// Don't fail the backup, but log the error
		m.emitEvent(BackupEvent{
			Type:       EventBackupCompleted,
			SourceFile: request.SourceFile,
			BackupID:   backupID,
			Message:    "Backup completed but metadata save failed",
			Error:      err.Error(),
			Duration:   metadata.Duration,
			Timestamp:  completedAt,
		})
	}

	// Emit success event
	m.emitEvent(BackupEvent{
		Type:       EventBackupCompleted,
		SourceFile: request.SourceFile,
		BackupID:   backupID,
		Message:    fmt.Sprintf("Backup completed successfully (size: %d bytes)", backupSize),
		Duration:   metadata.Duration,
		Timestamp:  completedAt,
	})

	// Schedule cleanup if needed
	go m.cleanupOldBackups(request.SourceFile)

	return &BackupResult{
		Success:    true,
		Metadata:   metadata,
		Duration:   metadata.Duration,
		BytesTotal: backupSize,
		Timestamp:  completedAt,
	}, nil
}

// RecoverFromBackup recovers a file from backup
func (m *Manager) RecoverFromBackup(request *RecoveryRequest) (*RecoveryResult, error) {
	startTime := time.Now()

	// Find the backup to restore from
	var backup *BackupMetadata
	var err error

	if request.BackupID != "" {
		backup, err = m.GetBackup(request.BackupID)
		if err != nil {
			return &RecoveryResult{
				Success:   false,
				Error:     fmt.Errorf("backup not found: %w", err),
				Duration:  time.Since(startTime),
				Timestamp: time.Now(),
			}, nil
		}
	} else if request.BackupTime != nil {
		backup, err = m.findBackupByTime(request.SourceFile, *request.BackupTime)
		if err != nil {
			return &RecoveryResult{
				Success:   false,
				Error:     fmt.Errorf("no backup found for time %v: %w", request.BackupTime, err),
				Duration:  time.Since(startTime),
				Timestamp: time.Now(),
			}, nil
		}
	} else {
		backup, err = m.getLatestBackup(request.SourceFile)
		if err != nil {
			return &RecoveryResult{
				Success:   false,
				Error:     fmt.Errorf("no backup found for file %s: %w", request.SourceFile, err),
				Duration:  time.Since(startTime),
				Timestamp: time.Now(),
			}, nil
		}
	}

	// Emit recovery start event
	m.emitEvent(BackupEvent{
		Type:       EventRecoveryStarted,
		SourceFile: request.SourceFile,
		BackupID:   backup.ID,
		Message:    fmt.Sprintf("Starting recovery from backup %s", backup.ID),
		Timestamp:  startTime,
	})

	// Verify backup before recovery if requested
	if request.VerifyBefore {
		if err := m.verifyBackupIntegrity(backup); err != nil {
			m.emitEvent(BackupEvent{
				Type:       EventRecoveryFailed,
				SourceFile: request.SourceFile,
				BackupID:   backup.ID,
				Message:    "Recovery failed: backup verification failed",
				Error:      err.Error(),
				Duration:   time.Since(startTime),
				Timestamp:  time.Now(),
			})
			return &RecoveryResult{
				Success:   false,
				Error:     fmt.Errorf("backup verification failed: %w", err),
				Duration:  time.Since(startTime),
				Timestamp: time.Now(),
			}, nil
		}
	}

	result := &RecoveryResult{
		BackupUsed: backup,
		Changes:    make([]string, 0),
		Warnings:   make([]string, 0),
	}

	// Create backup of current file if requested
	if request.CreateBackup {
		if _, err := os.Stat(request.SourceFile); err == nil {
			backupReq := &BackupRequest{
				SourceFile: request.SourceFile,
				Type:       BackupTypeEmergency,
				Reason:     ReasonPreRecovery,
				Tags:       []string{"pre-recovery"},
				Compress:   false,
				Verify:     true,
				Async:      false,
				Force:      true,
			}
			backupResult, err := m.CreateBackup(backupReq)
			if err != nil {
				result.Warnings = append(result.Warnings, fmt.Sprintf("Failed to create pre-recovery backup: %v", err))
			} else if backupResult.Success {
				result.BackupCreated = backupResult.Metadata
				result.Changes = append(result.Changes, "Created pre-recovery backup")
			}
		}
	}

	// Determine restore path
	restorePath := request.SourceFile
	if request.RestorePath != "" {
		restorePath = request.RestorePath
	}

	// Handle restore mode
	switch request.RestoreMode {
	case RestoreModeReplace:
		err = m.performRestore(backup.BackupFile, restorePath, backup.Compressed)
		if err == nil {
			result.Changes = append(result.Changes, "Replaced existing file")
		}
	case RestoreModeRename:
		// Rename existing file
		if _, err := os.Stat(restorePath); err == nil {
			renamedPath := restorePath + ".backup." + strconv.FormatInt(time.Now().Unix(), 10)
			if err := os.Rename(restorePath, renamedPath); err != nil {
				result.Warnings = append(result.Warnings, fmt.Sprintf("Failed to rename existing file: %v", err))
			} else {
				result.Changes = append(result.Changes, fmt.Sprintf("Renamed existing file to %s", renamedPath))
			}
		}
		err = m.performRestore(backup.BackupFile, restorePath, backup.Compressed)
		if err == nil {
			result.Changes = append(result.Changes, "Restored from backup")
		}
	case RestoreModePreview:
		// Just return what would be restored without actually doing it
		result.Success = true
		result.RestoredFile = restorePath
		result.Changes = append(result.Changes, fmt.Sprintf("Would restore from backup %s to %s", backup.ID, restorePath))
		result.Duration = time.Since(startTime)
		result.Timestamp = time.Now()
		return result, nil
	default:
		return &RecoveryResult{
			Success:   false,
			Error:     fmt.Errorf("unsupported restore mode: %s", request.RestoreMode),
			Duration:  time.Since(startTime),
			Timestamp: time.Now(),
		}, nil
	}

	if err != nil {
		m.emitEvent(BackupEvent{
			Type:       EventRecoveryFailed,
			SourceFile: request.SourceFile,
			BackupID:   backup.ID,
			Message:    "Recovery failed during restore operation",
			Error:      err.Error(),
			Duration:   time.Since(startTime),
			Timestamp:  time.Now(),
		})
		return &RecoveryResult{
			Success:    false,
			Error:      fmt.Errorf("restore operation failed: %w", err),
			BackupUsed: backup,
			Duration:   time.Since(startTime),
			Timestamp:  time.Now(),
			Changes:    result.Changes,
			Warnings:   result.Warnings,
		}, nil
	}

	result.Success = true
	result.RestoredFile = restorePath
	result.BytesRestored = backup.SourceSize
	result.Duration = time.Since(startTime)
	result.Timestamp = time.Now()

	// Verify restored file if requested
	if request.VerifyAfter {
		if restoredChecksum, _, err := m.calculateFileInfo(restorePath); err != nil {
			result.Warnings = append(result.Warnings, fmt.Sprintf("Failed to verify restored file: %v", err))
		} else if restoredChecksum == backup.SourceChecksum {
			result.IntegrityCheck = true
			result.Changes = append(result.Changes, "Verified restored file integrity")
		} else {
			result.Warnings = append(result.Warnings, "Restored file checksum does not match backup")
		}
	}

	// Emit success event
	m.emitEvent(BackupEvent{
		Type:       EventRecoveryCompleted,
		SourceFile: request.SourceFile,
		BackupID:   backup.ID,
		Message:    fmt.Sprintf("Recovery completed successfully to %s", restorePath),
		Duration:   result.Duration,
		Timestamp:  result.Timestamp,
	})

	return result, nil
}

// GetBackup retrieves backup metadata by ID
func (m *Manager) GetBackup(backupID string) (*BackupMetadata, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	backup, exists := m.backups[backupID]
	if !exists {
		return nil, fmt.Errorf("backup %s not found", backupID)
	}

	return backup, nil
}

// ListBackups returns a list of backups matching the filter
func (m *Manager) ListBackups(filter *BackupFilter) ([]*BackupMetadata, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var results []*BackupMetadata

	for _, backup := range m.backups {
		if m.matchesFilter(backup, filter) {
			results = append(results, backup)
		}
	}

	// Sort results
	if filter != nil && filter.SortBy != "" {
		m.sortBackups(results, filter.SortBy, filter.SortOrder)
	} else {
		// Default sort by creation time, newest first
		sort.Slice(results, func(i, j int) bool {
			return results[i].CreatedAt.After(results[j].CreatedAt)
		})
	}

	// Apply limit
	if filter != nil && filter.Limit > 0 && len(results) > filter.Limit {
		results = results[:filter.Limit]
	}

	return results, nil
}

// DeleteBackup removes a backup and its associated files
func (m *Manager) DeleteBackup(backupID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	backup, exists := m.backups[backupID]
	if !exists {
		return fmt.Errorf("backup %s not found", backupID)
	}

	// Remove backup file
	if err := os.Remove(backup.BackupFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove backup file: %w", err)
	}

	// Remove from memory
	delete(m.backups, backupID)

	// Update stats
	m.updateStats(backup, false)

	// Save metadata
	if err := m.saveMetadata(); err != nil {
		return fmt.Errorf("failed to save metadata after deletion: %w", err)
	}

	return nil
}

// GetStats returns backup statistics
func (m *Manager) GetStats() *BackupStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Create a copy to avoid race conditions
	stats := *m.stats
	return &stats
}

// OnEvent adds an event handler for backup events
func (m *Manager) OnEvent(handler func(BackupEvent)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.eventHandlers = append(m.eventHandlers, handler)
}

// Cleanup performs maintenance operations (cleanup old backups, verify integrity, etc.)
func (m *Manager) Cleanup() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	startTime := time.Now()

	m.emitEvent(BackupEvent{
		Type:      EventCleanupStarted,
		Message:   "Starting backup cleanup",
		Timestamp: startTime,
	})

	removed := 0
	errors := make([]error, 0)

	// Group backups by source file
	fileBackups := make(map[string][]*BackupMetadata)
	for _, backup := range m.backups {
		fileBackups[backup.SourceFile] = append(fileBackups[backup.SourceFile], backup)
	}

	// Clean up each file's backups according to retention policy
	for _, backups := range fileBackups {
		toRemove := m.selectBackupsForRemoval(backups)
		for _, backup := range toRemove {
			if err := os.Remove(backup.BackupFile); err != nil && !os.IsNotExist(err) {
				errors = append(errors, fmt.Errorf("failed to remove backup %s: %w", backup.ID, err))
			} else {
				delete(m.backups, backup.ID)
				removed++
			}
		}
	}

	// Save updated metadata
	if err := m.saveMetadata(); err != nil {
		errors = append(errors, fmt.Errorf("failed to save metadata: %w", err))
	}

	duration := time.Since(startTime)
	message := fmt.Sprintf("Cleanup completed: removed %d backups", removed)
	if len(errors) > 0 {
		message += fmt.Sprintf(" with %d errors", len(errors))
	}

	m.emitEvent(BackupEvent{
		Type:      EventCleanupCompleted,
		Message:   message,
		Duration:  duration,
		Timestamp: time.Now(),
	})

	if len(errors) > 0 {
		return fmt.Errorf("cleanup completed with errors: %v", errors)
	}

	return nil
}

// Helper methods

func (m *Manager) generateBackupID(sourceFile string) string {
	// Create a deterministic but unique ID based on source file and timestamp
	hash := sha256.Sum256([]byte(fmt.Sprintf("%s-%d", sourceFile, time.Now().UnixNano())))
	return fmt.Sprintf("backup-%s", hex.EncodeToString(hash[:8]))
}

func (m *Manager) generateBackupPath(sourceFile, backupID string) string {
	fileName := filepath.Base(sourceFile)
	timestamp := time.Now().Format("20060102-150405")
	backupFileName := fmt.Sprintf("%s.%s.%s.backup", fileName, timestamp, backupID[:8])
	return filepath.Join(m.backupDir, backupFileName)
}

func (m *Manager) calculateFileInfo(filePath string) (checksum string, size int64, err error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", 0, err
	}
	defer file.Close()

	hash := sha256.New()
	size, err = io.Copy(hash, file)
	if err != nil {
		return "", 0, err
	}

	checksum = hex.EncodeToString(hash.Sum(nil))
	return checksum, size, nil
}

func (m *Manager) performBackup(sourceFile, backupFile string, compress bool) (checksum string, size int64, err error) {
	// For now, implement simple file copy (compression can be added later)
	source, err := os.Open(sourceFile)
	if err != nil {
		return "", 0, err
	}
	defer source.Close()

	// Ensure backup directory exists
	if err := os.MkdirAll(filepath.Dir(backupFile), 0755); err != nil {
		return "", 0, err
	}

	dest, err := os.Create(backupFile)
	if err != nil {
		return "", 0, err
	}
	defer dest.Close()

	hash := sha256.New()
	writer := io.MultiWriter(dest, hash)

	size, err = io.Copy(writer, source)
	if err != nil {
		return "", 0, err
	}

	checksum = hex.EncodeToString(hash.Sum(nil))
	return checksum, size, nil
}

func (m *Manager) performRestore(backupFile, targetFile string, compressed bool) error {
	// For now, implement simple file copy (decompression can be added later)
	source, err := os.Open(backupFile)
	if err != nil {
		return err
	}
	defer source.Close()

	// Ensure target directory exists
	if err := os.MkdirAll(filepath.Dir(targetFile), 0755); err != nil {
		return err
	}

	dest, err := os.Create(targetFile)
	if err != nil {
		return err
	}
	defer dest.Close()

	_, err = io.Copy(dest, source)
	return err
}

func (m *Manager) verifyBackupIntegrity(metadata *BackupMetadata) error {
	backupChecksum, _, err := m.calculateFileInfo(metadata.BackupFile)
	if err != nil {
		return fmt.Errorf("failed to calculate backup file checksum: %w", err)
	}

	if backupChecksum != metadata.BackupChecksum {
		return fmt.Errorf("backup file checksum mismatch: expected %s, got %s", metadata.BackupChecksum, backupChecksum)
	}

	return nil
}

func (m *Manager) shouldSkipBackup(sourceFile string, backupType BackupType) bool {
	// Don't skip emergency or manual backups
	if backupType == BackupTypeEmergency || backupType == BackupTypeManual {
		return false
	}

	// Check if there's a recent backup
	m.mu.RLock()
	defer m.mu.RUnlock()

	var latestBackup *BackupMetadata
	for _, backup := range m.backups {
		if backup.SourceFile == sourceFile && backup.IsCompleted() {
			if latestBackup == nil || backup.CreatedAt.After(latestBackup.CreatedAt) {
				latestBackup = backup
			}
		}
	}

	if latestBackup == nil {
		return false
	}

	// Skip if backup is less than 5 minutes old
	return time.Since(latestBackup.CreatedAt) < 5*time.Minute
}

func (m *Manager) getLatestBackup(sourceFile string) (*BackupMetadata, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var latestBackup *BackupMetadata
	for _, backup := range m.backups {
		if backup.SourceFile == sourceFile && backup.IsCompleted() {
			if latestBackup == nil || backup.CreatedAt.After(latestBackup.CreatedAt) {
				latestBackup = backup
			}
		}
	}

	if latestBackup == nil {
		return nil, fmt.Errorf("no completed backup found for file %s", sourceFile)
	}

	return latestBackup, nil
}

func (m *Manager) findBackupByTime(sourceFile string, targetTime time.Time) (*BackupMetadata, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var closestBackup *BackupMetadata
	var closestDiff time.Duration

	for _, backup := range m.backups {
		if backup.SourceFile == sourceFile && backup.IsCompleted() {
			diff := targetTime.Sub(backup.CreatedAt)
			if diff >= 0 && (closestBackup == nil || diff < closestDiff) {
				closestBackup = backup
				closestDiff = diff
			}
		}
	}

	if closestBackup == nil {
		return nil, fmt.Errorf("no backup found for file %s before time %v", sourceFile, targetTime)
	}

	return closestBackup, nil
}

func (m *Manager) matchesFilter(backup *BackupMetadata, filter *BackupFilter) bool {
	if filter == nil {
		return true
	}

	if filter.SourceFile != "" && backup.SourceFile != filter.SourceFile {
		return false
	}

	if filter.Type != "" && backup.Type != filter.Type {
		return false
	}

	if filter.Reason != "" && backup.Reason != filter.Reason {
		return false
	}

	if filter.Status != "" && backup.Status != filter.Status {
		return false
	}

	if filter.CreatedAfter != nil && backup.CreatedAt.Before(*filter.CreatedAfter) {
		return false
	}

	if filter.CreatedBefore != nil && backup.CreatedAt.After(*filter.CreatedBefore) {
		return false
	}

	if filter.MinSize > 0 && backup.BackupSize < filter.MinSize {
		return false
	}

	if filter.MaxSize > 0 && backup.BackupSize > filter.MaxSize {
		return false
	}

	if filter.Verified != nil && backup.IntegrityCheck != *filter.Verified {
		return false
	}

	if len(filter.Tags) > 0 {
		hasAllTags := true
		for _, filterTag := range filter.Tags {
			found := false
			for _, backupTag := range backup.Tags {
				if backupTag == filterTag {
					found = true
					break
				}
			}
			if !found {
				hasAllTags = false
				break
			}
		}
		if !hasAllTags {
			return false
		}
	}

	return true
}

func (m *Manager) sortBackups(backups []*BackupMetadata, sortBy, sortOrder string) {
	ascending := sortOrder != "desc"

	sort.Slice(backups, func(i, j int) bool {
		var result bool

		switch sortBy {
		case "created_at":
			result = backups[i].CreatedAt.Before(backups[j].CreatedAt)
		case "size":
			result = backups[i].BackupSize < backups[j].BackupSize
		case "type":
			result = string(backups[i].Type) < string(backups[j].Type)
		case "status":
			result = string(backups[i].Status) < string(backups[j].Status)
		default:
			result = backups[i].CreatedAt.Before(backups[j].CreatedAt)
		}

		if !ascending {
			result = !result
		}

		return result
	})
}

func (m *Manager) selectBackupsForRemoval(backups []*BackupMetadata) []*BackupMetadata {
	if len(backups) <= m.retention.MaxCount {
		return nil
	}

	// Sort by creation time, oldest first
	sort.Slice(backups, func(i, j int) bool {
		return backups[i].CreatedAt.Before(backups[j].CreatedAt)
	})

	var toRemove []*BackupMetadata

	// Remove excess backups beyond MaxCount
	if len(backups) > m.retention.MaxCount {
		toRemove = append(toRemove, backups[:len(backups)-m.retention.MaxCount]...)
	}

	// Remove backups older than MaxAge
	cutoffTime := time.Now().Add(-m.retention.MaxAge)
	for _, backup := range backups {
		if backup.CreatedAt.Before(cutoffTime) {
			// Don't add duplicates
			found := false
			for _, existing := range toRemove {
				if existing.ID == backup.ID {
					found = true
					break
				}
			}
			if !found {
				toRemove = append(toRemove, backup)
			}
		}
	}

	return toRemove
}

func (m *Manager) updateStats(backup *BackupMetadata, isAdd bool) {
	if isAdd {
		m.stats.TotalBackups++
		m.stats.TotalSize += backup.BackupSize
		if backup.IsCompleted() {
			m.stats.SuccessfulBackups++
		} else {
			m.stats.FailedBackups++
		}
		m.stats.LastBackupTime = backup.CreatedAt

		// Update average backup time
		if m.stats.TotalBackups > 0 {
			totalDuration := m.stats.AverageBackupTime * time.Duration(m.stats.TotalBackups-1)
			totalDuration += backup.Duration
			m.stats.AverageBackupTime = totalDuration / time.Duration(m.stats.TotalBackups)
		}
	} else {
		m.stats.TotalBackups--
		m.stats.TotalSize -= backup.BackupSize
		if backup.IsCompleted() {
			m.stats.SuccessfulBackups--
		} else {
			m.stats.FailedBackups--
		}
	}
}

func (m *Manager) emitEvent(event BackupEvent) {
	m.events = append(m.events, event)

	// Call event handlers
	for _, handler := range m.eventHandlers {
		go handler(event)
	}
}

func (m *Manager) emitFailureEvent(sourceFile, backupID string, err error) {
	m.emitEvent(BackupEvent{
		Type:       EventBackupFailed,
		SourceFile: sourceFile,
		BackupID:   backupID,
		Message:    "Backup operation failed",
		Error:      err.Error(),
		Timestamp:  time.Now(),
	})
}

func (m *Manager) cleanupOldBackups(sourceFile string) {
	filter := &BackupFilter{
		SourceFile: sourceFile,
	}
	backups, err := m.ListBackups(filter)
	if err != nil {
		return
	}

	toRemove := m.selectBackupsForRemoval(backups)
	for _, backup := range toRemove {
		m.DeleteBackup(backup.ID)
	}
}

func (m *Manager) loadMetadata() error {
	if _, err := os.Stat(m.metadataFile); os.IsNotExist(err) {
		// Metadata file doesn't exist yet, that's okay
		return nil
	}

	data, err := os.ReadFile(m.metadataFile)
	if err != nil {
		return err
	}

	var backupList []*BackupMetadata
	if err := json.Unmarshal(data, &backupList); err != nil {
		return err
	}

	m.backups = make(map[string]*BackupMetadata)
	for _, backup := range backupList {
		m.backups[backup.ID] = backup
		m.updateStats(backup, true)
	}

	return nil
}

func (m *Manager) saveMetadata() error {
	var backupList []*BackupMetadata
	for _, backup := range m.backups {
		backupList = append(backupList, backup)
	}

	data, err := json.MarshalIndent(backupList, "", "  ")
	if err != nil {
		return err
	}

	// Use atomic write (temp file + rename)
	tempFile := m.metadataFile + ".tmp"
	if err := os.WriteFile(tempFile, data, 0644); err != nil {
		return err
	}

	return os.Rename(tempFile, m.metadataFile)
}
