package locking

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// FileLock represents a file lock implementation
type FileLock struct {
	filePath string
	lockPath string
	lockFile *os.File
	lockInfo *LockInfo
	options  *LockOptions
	mu       sync.RWMutex
	acquired bool
	metrics  *LockMetrics
}

// NewFileLock creates a new file lock for the specified file
func NewFileLock(filePath string, options *LockOptions) *FileLock {
	if options == nil {
		options = DefaultLockOptions()
	}

	lockPath := generateLockPath(filePath)

	return &FileLock{
		filePath: filePath,
		lockPath: lockPath,
		options:  options,
		metrics:  &LockMetrics{},
	}
}

// Lock acquires the file lock
func (fl *FileLock) Lock() *LockResult {
	fl.mu.Lock()
	defer fl.mu.Unlock()

	start := time.Now()
	result := &LockResult{
		Timestamp: start,
	}

	if fl.acquired {
		result.Status = LockStatusHeld
		result.Message = "Lock already held by current process"
		result.Duration = time.Since(start)
		return result
	}

	// Check if lock file exists and analyze it
	if existingInfo, exists := fl.checkExistingLock(); exists {
		if existingInfo.IsStale() {
			// Clean up stale lock
			if err := fl.cleanupStaleLock(existingInfo); err != nil {
				result.Status = LockStatusError
				result.Error = fmt.Errorf("failed to cleanup stale lock: %w", err)
				result.Duration = time.Since(start)
				return result
			}
		} else {
			// Active lock exists
			result.Status = LockStatusBlocked
			result.LockInfo = existingInfo
			result.Error = &LockConflictError{
				RequestedType: fl.options.Type,
				ExistingLock:  existingInfo,
				FilePath:      fl.filePath,
				Timestamp:     time.Now(),
			}
			result.Message = fmt.Sprintf("Lock held by PID %d", existingInfo.PID)
			result.Suggestion = "Wait for lock to be released or kill process if stale"
			result.Duration = time.Since(start)
			return result
		}
	}

	// Attempt to acquire lock with retry logic
	attempts := 0
	maxAttempts := int(fl.options.Timeout / fl.options.RetryDelay)

	for attempts < maxAttempts {
		attempts++
		fl.metrics.TotalRequests++

		if err := fl.acquireLock(); err != nil {
			if fl.options.NonBlocking || attempts >= maxAttempts {
				result.Status = LockStatusError
				result.Error = err
				result.Attempts = attempts
				result.Duration = time.Since(start)
				fl.metrics.FailedLocks++
				return result
			}

			// Wait before retry
			time.Sleep(fl.options.RetryDelay)
			continue
		}

		// Lock acquired successfully
		fl.acquired = true
		result.Status = LockStatusHeld
		result.LockInfo = fl.lockInfo
		result.Attempts = attempts
		result.Duration = time.Since(start)
		result.Message = "Lock acquired successfully"

		fl.metrics.SuccessfulLocks++
		if result.Duration > fl.metrics.MaxLockTime {
			fl.metrics.MaxLockTime = result.Duration
		}

		return result
	}

	// Timeout reached
	result.Status = LockStatusError
	result.Error = &LockTimeoutError{
		FilePath: fl.filePath,
		LockType: fl.options.Type,
		Timeout:  fl.options.Timeout,
		Attempts: attempts,
	}
	result.Attempts = attempts
	result.Duration = time.Since(start)
	result.Suggestion = "Increase timeout or check for stale processes"

	fl.metrics.TimeoutErrors++

	return result
}

// Unlock releases the file lock
func (fl *FileLock) Unlock() error {
	fl.mu.Lock()
	defer fl.mu.Unlock()

	if !fl.acquired {
		return fmt.Errorf("lock not held by current process")
	}

	// Close the lock file
	if fl.lockFile != nil {
		fl.lockFile.Close()
		fl.lockFile = nil
	}

	// Remove the lock file
	if err := os.Remove(fl.lockPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove lock file: %w", err)
	}

	fl.acquired = false
	fl.lockInfo = nil

	return nil
}

// IsLocked returns true if the lock is currently held
func (fl *FileLock) IsLocked() bool {
	fl.mu.RLock()
	defer fl.mu.RUnlock()
	return fl.acquired
}

// GetLockInfo returns information about the current lock
func (fl *FileLock) GetLockInfo() *LockInfo {
	fl.mu.RLock()
	defer fl.mu.RUnlock()
	if fl.lockInfo != nil {
		// Return a copy to prevent modifications
		info := *fl.lockInfo
		return &info
	}
	return nil
}

// CheckLockStatus checks the status of the lock without acquiring it
func (fl *FileLock) CheckLockStatus() *LockResult {
	fl.mu.RLock()
	defer fl.mu.RUnlock()

	result := &LockResult{
		Timestamp: time.Now(),
	}

	if fl.acquired {
		result.Status = LockStatusHeld
		result.LockInfo = fl.lockInfo
		result.Message = "Lock held by current process"
		return result
	}

	if existingInfo, exists := fl.checkExistingLock(); exists {
		if existingInfo.IsStale() {
			result.Status = LockStatusStale
			result.LockInfo = existingInfo
			result.Message = "Stale lock detected"
			result.Suggestion = "Clean up stale lock and retry"
		} else {
			result.Status = LockStatusBlocked
			result.LockInfo = existingInfo
			result.Message = fmt.Sprintf("Lock held by PID %d", existingInfo.PID)
		}
	} else {
		result.Status = LockStatusAvailable
		result.Message = "Lock is available"
	}

	return result
}

// GetMetrics returns locking metrics
func (fl *FileLock) GetMetrics() *LockMetrics {
	fl.mu.RLock()
	defer fl.mu.RUnlock()

	// Return a copy
	metrics := *fl.metrics
	if fl.acquired {
		metrics.ActiveLocks = 1
	}
	return &metrics
}

// Helper methods

func (fl *FileLock) acquireLock() error {
	// Create lock directory if needed
	lockDir := filepath.Dir(fl.lockPath)
	if err := os.MkdirAll(lockDir, 0755); err != nil {
		return fmt.Errorf("failed to create lock directory: %w", err)
	}

	// Create lock info
	lockInfo := &LockInfo{
		FilePath:    fl.filePath,
		LockPath:    fl.lockPath,
		Type:        fl.options.Type,
		PID:         os.Getpid(),
		AcquiredAt:  time.Now(),
		ExpiresAt:   time.Now().Add(fl.options.StaleTimeout),
		ProcessName: getProcessName(),
		UserInfo:    getUserInfo(),
	}

	if hostname, err := os.Hostname(); err == nil {
		lockInfo.Hostname = hostname
	}

	// Try to create lock file exclusively
	lockFile, err := os.OpenFile(fl.lockPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if err != nil {
		if os.IsExist(err) {
			return fmt.Errorf("lock file already exists")
		}
		return fmt.Errorf("failed to create lock file: %w", err)
	}

	// Write lock info to file
	encoder := json.NewEncoder(lockFile)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(lockInfo); err != nil {
		lockFile.Close()
		os.Remove(fl.lockPath)
		return fmt.Errorf("failed to write lock info: %w", err)
	}

	// Perform platform-specific locking
	if err := fl.platformLock(lockFile); err != nil {
		lockFile.Close()
		os.Remove(fl.lockPath)
		return fmt.Errorf("failed to acquire platform lock: %w", err)
	}

	fl.lockFile = lockFile
	fl.lockInfo = lockInfo

	return nil
}

func (fl *FileLock) checkExistingLock() (*LockInfo, bool) {
	data, err := os.ReadFile(fl.lockPath)
	if err != nil {
		return nil, false
	}

	var lockInfo LockInfo
	if err := json.Unmarshal(data, &lockInfo); err != nil {
		// Invalid lock file, consider it stale
		return &LockInfo{
			FilePath:   fl.filePath,
			LockPath:   fl.lockPath,
			PID:        -1,
			AcquiredAt: time.Time{},
			ExpiresAt:  time.Time{},
		}, true
	}

	// Validate lock info
	if !fl.validateLockInfo(&lockInfo) {
		return &lockInfo, true // Consider invalid lock as stale
	}

	return &lockInfo, true
}

func (fl *FileLock) validateLockInfo(info *LockInfo) bool {
	// Check if the process still exists
	if !fl.processExists(info.PID) {
		return false
	}

	// Check if lock has expired
	if info.IsStale() {
		return false
	}

	return true
}

func (fl *FileLock) cleanupStaleLock(info *LockInfo) error {
	if err := os.Remove(fl.lockPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove stale lock file: %w", err)
	}

	fl.metrics.StaleLocksCleared++
	return nil
}

func generateLockPath(filePath string) string {
	dir := filepath.Dir(filePath)
	base := filepath.Base(filePath)
	lockName := fmt.Sprintf(".%s.lock", base)
	return filepath.Join(dir, lockName)
}

func getProcessName() string {
	if len(os.Args) > 0 {
		return filepath.Base(os.Args[0])
	}
	return "unknown"
}

func getUserInfo() string {
	if user := os.Getenv("USER"); user != "" {
		return user
	}
	if user := os.Getenv("USERNAME"); user != "" {
		return user
	}
	return "unknown"
}

// TryLock attempts to acquire the lock without waiting
func (fl *FileLock) TryLock() *LockResult {
	originalBlocking := fl.options.NonBlocking
	originalTimeout := fl.options.Timeout

	fl.options.NonBlocking = true
	fl.options.Timeout = 0

	result := fl.Lock()

	fl.options.NonBlocking = originalBlocking
	fl.options.Timeout = originalTimeout

	return result
}

// LockWithTimeout attempts to acquire the lock with a specific timeout
func (fl *FileLock) LockWithTimeout(timeout time.Duration) *LockResult {
	originalTimeout := fl.options.Timeout
	fl.options.Timeout = timeout

	result := fl.Lock()

	fl.options.Timeout = originalTimeout
	return result
}

// Refresh extends the lock expiration time
func (fl *FileLock) Refresh() error {
	fl.mu.Lock()
	defer fl.mu.Unlock()

	if !fl.acquired || fl.lockInfo == nil {
		return fmt.Errorf("no active lock to refresh")
	}

	// Update expiration time
	fl.lockInfo.ExpiresAt = time.Now().Add(fl.options.StaleTimeout)

	// Rewrite lock file with updated info
	if fl.lockFile != nil {
		if _, err := fl.lockFile.Seek(0, 0); err != nil {
			return fmt.Errorf("failed to seek lock file: %w", err)
		}

		if err := fl.lockFile.Truncate(0); err != nil {
			return fmt.Errorf("failed to truncate lock file: %w", err)
		}

		encoder := json.NewEncoder(fl.lockFile)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(fl.lockInfo); err != nil {
			return fmt.Errorf("failed to write updated lock info: %w", err)
		}
	}

	return nil
}

// Close releases the lock and closes the file
func (fl *FileLock) Close() error {
	if fl.acquired {
		return fl.Unlock()
	}
	return nil
}

// String returns a string representation of the lock
func (fl *FileLock) String() string {
	fl.mu.RLock()
	defer fl.mu.RUnlock()

	if fl.acquired && fl.lockInfo != nil {
		return fmt.Sprintf("FileLock{path=%s, type=%s, pid=%d, acquired=%s}",
			fl.filePath, fl.options.Type, fl.lockInfo.PID, fl.lockInfo.AcquiredAt.Format("15:04:05"))
	}

	return fmt.Sprintf("FileLock{path=%s, type=%s, acquired=false}", fl.filePath, fl.options.Type)
}
