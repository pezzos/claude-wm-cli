package locking

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// LockManager manages multiple file locks and provides high-level locking operations
type LockManager struct {
	config   *LockConfig
	locks    map[string]*FileLock
	mu       sync.RWMutex
	metrics  *LockMetrics
	cleanup  *time.Ticker
	stopChan chan struct{}
	events   chan LockEvent
}

// NewLockManager creates a new lock manager with the specified configuration
func NewLockManager(config *LockConfig) *LockManager {
	if config == nil {
		config = DefaultLockConfig()
	}

	lm := &LockManager{
		config:   config,
		locks:    make(map[string]*FileLock),
		metrics:  &LockMetrics{},
		stopChan: make(chan struct{}),
		events:   make(chan LockEvent, 100), // Buffer for events
	}

	// Start cleanup routine if enabled
	if config.Enabled && config.CleanupInterval > 0 {
		lm.cleanup = time.NewTicker(config.CleanupInterval)
		go lm.cleanupRoutine()
	}

	return lm
}

// LockFile acquires a lock on the specified file
func (lm *LockManager) LockFile(filePath string, options *LockOptions) (*FileLock, *LockResult) {
	if !lm.config.Enabled {
		return nil, &LockResult{
			Status:    LockStatusError,
			Error:     fmt.Errorf("file locking is disabled"),
			Timestamp: time.Now(),
		}
	}

	if options == nil {
		options = DefaultLockOptions()
		options.Timeout = lm.config.DefaultTimeout
	}

	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, &LockResult{
			Status:    LockStatusError,
			Error:     fmt.Errorf("failed to resolve absolute path: %w", err),
			Timestamp: time.Now(),
		}
	}

	lm.mu.Lock()
	defer lm.mu.Unlock()

	// Check if we already have a lock on this file
	if existingLock, exists := lm.locks[absPath]; exists {
		if existingLock.IsLocked() {
			return existingLock, &LockResult{
				Status:    LockStatusHeld,
				LockInfo:  existingLock.GetLockInfo(),
				Message:   "Lock already held by current process",
				Timestamp: time.Now(),
			}
		}
		// Remove stale lock entry
		delete(lm.locks, absPath)
	}

	// Create new lock
	lock := NewFileLock(absPath, options)
	result := lock.Lock()

	// Store lock if successful
	if result.Status == LockStatusHeld {
		lm.locks[absPath] = lock
		lm.emitEvent(LockEventAcquired, absPath, options.Type, result.Duration, "")
	} else {
		lm.emitEvent(LockEventError, absPath, options.Type, result.Duration, result.Error.Error())
	}

	// Update metrics
	lm.updateMetrics(result)

	return lock, result
}

// UnlockFile releases a lock on the specified file
func (lm *LockManager) UnlockFile(filePath string) error {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("failed to resolve absolute path: %w", err)
	}

	lm.mu.Lock()
	defer lm.mu.Unlock()

	lock, exists := lm.locks[absPath]
	if !exists {
		return fmt.Errorf("no lock found for file: %s", filePath)
	}

	if err := lock.Unlock(); err != nil {
		lm.emitEvent(LockEventError, absPath, lock.options.Type, 0, err.Error())
		return err
	}

	delete(lm.locks, absPath)
	lm.emitEvent(LockEventReleased, absPath, lock.options.Type, 0, "")

	return nil
}

// UnlockAll releases all locks held by this manager
func (lm *LockManager) UnlockAll() error {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	var errors []error

	for path, lock := range lm.locks {
		if err := lock.Unlock(); err != nil {
			errors = append(errors, fmt.Errorf("failed to unlock %s: %w", path, err))
		}
	}

	// Clear all locks
	lm.locks = make(map[string]*FileLock)

	if len(errors) > 0 {
		return fmt.Errorf("failed to unlock some files: %v", errors)
	}

	return nil
}

// IsLocked checks if a file is currently locked by this manager
func (lm *LockManager) IsLocked(filePath string) bool {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return false
	}

	lm.mu.RLock()
	defer lm.mu.RUnlock()

	if lock, exists := lm.locks[absPath]; exists {
		return lock.IsLocked()
	}

	return false
}

// GetLockInfo returns information about a locked file
func (lm *LockManager) GetLockInfo(filePath string) *LockInfo {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil
	}

	lm.mu.RLock()
	defer lm.mu.RUnlock()

	if lock, exists := lm.locks[absPath]; exists {
		return lock.GetLockInfo()
	}

	return nil
}

// CheckLockStatus checks the status of a file lock without acquiring it
func (lm *LockManager) CheckLockStatus(filePath string) *LockResult {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return &LockResult{
			Status:    LockStatusError,
			Error:     fmt.Errorf("failed to resolve absolute path: %w", err),
			Timestamp: time.Now(),
		}
	}

	// Create a temporary lock to check status
	lock := NewFileLock(absPath, DefaultLockOptions())
	return lock.CheckLockStatus()
}

// ListActiveLocks returns information about all active locks
func (lm *LockManager) ListActiveLocks() []*LockInfo {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	var locks []*LockInfo
	for _, lock := range lm.locks {
		if info := lock.GetLockInfo(); info != nil {
			locks = append(locks, info)
		}
	}

	return locks
}

// GetMetrics returns aggregated locking metrics
func (lm *LockManager) GetMetrics() *LockMetrics {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	// Aggregate metrics from all locks
	aggregated := *lm.metrics
	aggregated.ActiveLocks = len(lm.locks)

	return &aggregated
}

// CleanupStaleLocks removes stale lock files from the filesystem
func (lm *LockManager) CleanupStaleLocks(directory string) (int, error) {
	if !lm.config.Enabled {
		return 0, nil
	}

	cleaned := 0

	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Continue walking
		}

		// Look for lock files (files starting with . and ending with .lock)
		if !info.IsDir() && isLockFile(path) {
			targetFile := getLockTarget(path)

			// Create temporary lock to check if it's stale
			tempLock := NewFileLock(targetFile, DefaultLockOptions())
			if existingInfo, exists := tempLock.checkExistingLock(); exists {
				if existingInfo.IsStale() || !tempLock.validateLockInfo(existingInfo) {
					if err := os.Remove(path); err == nil {
						cleaned++
						lm.emitEvent(LockEventStale, targetFile, LockExclusive, 0, "cleaned")
					}
				}
			}
		}

		return nil
	})

	lm.metrics.StaleLocksCleared += int64(cleaned)
	lm.metrics.LastCleanup = time.Now()

	return cleaned, err
}

// WithLock executes a function while holding a lock on the specified file
func (lm *LockManager) WithLock(filePath string, options *LockOptions, fn func() error) error {
	_, result := lm.LockFile(filePath, options)
	if result.Status != LockStatusHeld {
		return result.Error
	}

	defer func() {
		if err := lm.UnlockFile(filePath); err != nil {
			// Log error but don't override function error
			lm.emitEvent(LockEventError, filePath, options.Type, 0,
				fmt.Sprintf("failed to unlock: %v", err))
		}
	}()

	return fn()
}

// WithLockContext executes a function while holding a lock, with context cancellation
func (lm *LockManager) WithLockContext(ctx context.Context, filePath string, options *LockOptions,
	fn func(context.Context) error) error {

	// Create channel to coordinate lock acquisition
	lockChan := make(chan *LockResult, 1)

	go func() {
		_, result := lm.LockFile(filePath, options)
		lockChan <- result
	}()

	// Wait for lock or context cancellation
	select {
	case result := <-lockChan:
		if result.Status != LockStatusHeld {
			return result.Error
		}

		defer func() {
			if err := lm.UnlockFile(filePath); err != nil {
				lm.emitEvent(LockEventError, filePath, options.Type, 0,
					fmt.Sprintf("failed to unlock: %v", err))
			}
		}()

		return fn(ctx)

	case <-ctx.Done():
		return ctx.Err()
	}
}

// SetConfig updates the manager configuration
func (lm *LockManager) SetConfig(config *LockConfig) {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	lm.config = config

	// Restart cleanup routine if needed
	if lm.cleanup != nil {
		lm.cleanup.Stop()
		lm.cleanup = nil
	}

	if config.Enabled && config.CleanupInterval > 0 {
		lm.cleanup = time.NewTicker(config.CleanupInterval)
		go lm.cleanupRoutine()
	}
}

// GetConfig returns the current configuration
func (lm *LockManager) GetConfig() *LockConfig {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	// Return a copy
	config := *lm.config
	return &config
}

// Close shuts down the lock manager and releases all locks
func (lm *LockManager) Close() error {
	// Stop cleanup routine
	close(lm.stopChan)
	if lm.cleanup != nil {
		lm.cleanup.Stop()
	}

	// Release all locks
	return lm.UnlockAll()
}

// Events returns a channel for lock events
func (lm *LockManager) Events() <-chan LockEvent {
	return lm.events
}

// Helper methods

func (lm *LockManager) cleanupRoutine() {
	for {
		select {
		case <-lm.cleanup.C:
			// Get all directories that might contain lock files
			directories := lm.getActiveLockDirectories()
			for _, dir := range directories {
				lm.CleanupStaleLocks(dir)
			}

		case <-lm.stopChan:
			return
		}
	}
}

func (lm *LockManager) getActiveLockDirectories() []string {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	dirSet := make(map[string]bool)
	for path := range lm.locks {
		dir := filepath.Dir(path)
		dirSet[dir] = true
	}

	var directories []string
	for dir := range dirSet {
		directories = append(directories, dir)
	}

	return directories
}

func (lm *LockManager) updateMetrics(result *LockResult) {
	switch result.Status {
	case LockStatusHeld:
		lm.metrics.SuccessfulLocks++
	case LockStatusError:
		lm.metrics.FailedLocks++
		if _, isTimeout := result.Error.(*LockTimeoutError); isTimeout {
			lm.metrics.TimeoutErrors++
		}
		if _, isConflict := result.Error.(*LockConflictError); isConflict {
			lm.metrics.ConflictErrors++
		}
	}

	// Update average lock time
	if result.Duration > 0 {
		if lm.metrics.AverageLockTime == 0 {
			lm.metrics.AverageLockTime = result.Duration
		} else {
			// Simple moving average
			lm.metrics.AverageLockTime = (lm.metrics.AverageLockTime + result.Duration) / 2
		}

		if result.Duration > lm.metrics.MaxLockTime {
			lm.metrics.MaxLockTime = result.Duration
		}
	}
}

func (lm *LockManager) emitEvent(eventType LockEventType, filePath string, lockType LockType,
	duration time.Duration, details string) {

	event := LockEvent{
		Type:      eventType,
		FilePath:  filePath,
		LockType:  lockType,
		PID:       os.Getpid(),
		Duration:  duration,
		Details:   details,
		Timestamp: time.Now(),
	}

	// Non-blocking send
	select {
	case lm.events <- event:
	default:
		// Event buffer full, drop event
	}
}

func isLockFile(path string) bool {
	base := filepath.Base(path)
	return len(base) > 5 && base[0] == '.' && filepath.Ext(base) == ".lock"
}

func getLockTarget(lockPath string) string {
	dir := filepath.Dir(lockPath)
	base := filepath.Base(lockPath)

	// Remove leading dot and .lock extension
	if len(base) > 5 && base[0] == '.' && filepath.Ext(base) == ".lock" {
		targetBase := base[1 : len(base)-5] // Remove . prefix and .lock suffix
		return filepath.Join(dir, targetBase)
	}

	return lockPath
}

// Global lock manager instance
var globalLockManager *LockManager
var globalLockManagerOnce sync.Once

// GetGlobalLockManager returns the global lock manager instance
func GetGlobalLockManager() *LockManager {
	globalLockManagerOnce.Do(func() {
		globalLockManager = NewLockManager(DefaultLockConfig())
	})
	return globalLockManager
}

// Convenience functions using global manager

// GlobalLockFile acquires a lock using the global manager
func GlobalLockFile(filePath string, options *LockOptions) (*FileLock, *LockResult) {
	return GetGlobalLockManager().LockFile(filePath, options)
}

// GlobalUnlockFile releases a lock using the global manager
func GlobalUnlockFile(filePath string) error {
	return GetGlobalLockManager().UnlockFile(filePath)
}

// GlobalWithLock executes a function with a lock using the global manager
func GlobalWithLock(filePath string, options *LockOptions, fn func() error) error {
	return GetGlobalLockManager().WithLock(filePath, options, fn)
}
