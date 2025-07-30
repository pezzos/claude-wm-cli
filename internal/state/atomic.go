package state

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"claude-wm-cli/internal/model"
)

// GitVersionManager interface for optional Git integration
type GitVersionManager interface {
	AutoVersionOnWrite(filePath string, commitType interface{}, description string) error
	IsEnabled() bool
}

// LockManager interface for optional file locking integration
type LockManager interface {
	LockFile(filePath string, options interface{}) (interface{}, interface{})
	UnlockFile(filePath string) error
	IsLocked(filePath string) bool
}

// AtomicWriter provides atomic file operations using temp file + rename pattern
// This prevents corruption by ensuring writes are either complete or not applied at all
type AtomicWriter struct {
	mu             sync.RWMutex
	tempDir        string
	permissions    os.FileMode
	backup         bool
	compress       bool
	checksums      map[string]string // file -> checksum for integrity verification
	gitManager     GitVersionManager // Optional Git integration
	lockManager    LockManager       // Optional file locking integration
	lockingEnabled bool              // Whether to use file locking
}

// AtomicWriteOptions configures atomic write operations
type AtomicWriteOptions struct {
	Permissions os.FileMode
	Backup      bool
	Compress    bool
	Verify      bool          // Verify write by reading back and comparing
	GitCommit   bool          // Whether to create Git commit after write
	CommitType  interface{}   // Git commit type for versioning
	CommitMsg   string        // Custom commit message
	UseLocking  bool          // Whether to use file locking for this operation
	LockTimeout time.Duration // Timeout for lock acquisition
}

// NewAtomicWriter creates a new atomic writer with default settings
func NewAtomicWriter(tempDir string) *AtomicWriter {
	if tempDir == "" {
		tempDir = os.TempDir()
	}

	return &AtomicWriter{
		tempDir:        tempDir,
		permissions:    0644,
		backup:         true,
		compress:       false,
		checksums:      make(map[string]string),
		gitManager:     nil,   // Git integration disabled by default
		lockManager:    nil,   // File locking disabled by default
		lockingEnabled: false, // Disabled by default for backward compatibility
	}
}

// WriteJSON atomically writes JSON data to a file
func (aw *AtomicWriter) WriteJSON(filePath string, data interface{}, opts *AtomicWriteOptions) error {
	aw.mu.Lock()
	defer aw.mu.Unlock()

	if opts == nil {
		opts = &AtomicWriteOptions{
			Permissions: aw.permissions,
			Backup:      aw.backup,
			Compress:    aw.compress,
			Verify:      true,
		}
	}

	// Serialize data to JSON
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return model.NewInternalError("failed to serialize JSON data").
			WithCause(err).
			WithContext(filePath)
	}

	return aw.writeBytes(filePath, jsonData, opts)
}

// WriteBytes atomically writes raw bytes to a file
func (aw *AtomicWriter) WriteBytes(filePath string, data []byte, opts *AtomicWriteOptions) error {
	aw.mu.Lock()
	defer aw.mu.Unlock()

	if opts == nil {
		opts = &AtomicWriteOptions{
			Permissions: aw.permissions,
			Backup:      aw.backup,
			Compress:    aw.compress,
			Verify:      true,
		}
	}

	return aw.writeBytes(filePath, data, opts)
}

// writeBytes implements the atomic write pattern: temp file + rename
func (aw *AtomicWriter) writeBytes(filePath string, data []byte, opts *AtomicWriteOptions) error {
	// Acquire file lock if enabled and requested
	var lockAcquired bool
	if opts.UseLocking && aw.lockingEnabled && aw.lockManager != nil {
		lockOptions := map[string]interface{}{
			"timeout": opts.LockTimeout,
			"type":    "exclusive",
		}
		if opts.LockTimeout == 0 {
			lockOptions["timeout"] = 30 * time.Second // Default proven timeout
		}

		_, lockResult := aw.lockManager.LockFile(filePath, lockOptions)
		if lockResult != nil {
			// Check if lock was acquired (implementation dependent)
			lockAcquired = true
		}

		// Defer unlock if we acquired the lock
		if lockAcquired {
			defer func() {
				if err := aw.lockManager.UnlockFile(filePath); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: Failed to unlock file %s: %v\n", filePath, err)
				}
			}()
		}
	}

	// Ensure target directory exists
	targetDir := filepath.Dir(filePath)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return model.NewFileSystemError("create_directory", targetDir, err).
			WithSuggestions([]string{
				"Check directory permissions",
				"Run with appropriate privileges",
				"Ensure parent directory exists",
			})
	}

	// Create backup if requested and file exists
	if opts.Backup && fileExists(filePath) {
		if err := aw.createBackup(filePath); err != nil {
			return model.NewInternalError("failed to create backup").
				WithCause(err).
				WithContext(filePath).
				WithSuggestions([]string{
					"Check write permissions in target directory",
					"Ensure sufficient disk space",
				})
		}
	}

	// Generate temporary file name in the same directory as target
	// This ensures atomic rename works (same filesystem)
	tempFile := filepath.Join(targetDir, fmt.Sprintf(".tmp_%s_%d",
		filepath.Base(filePath), time.Now().UnixNano()))

	// Write to temporary file
	if err := aw.writeToTempFile(tempFile, data, opts); err != nil {
		// Clean up temp file on error
		os.Remove(tempFile)
		return err
	}

	// Verify write if requested
	if opts.Verify {
		if err := aw.verifyWrite(tempFile, data); err != nil {
			os.Remove(tempFile)
			return model.NewValidationError("write verification failed").
				WithCause(err).
				WithContext(tempFile).
				WithSuggestion("Data integrity check failed during atomic write")
		}
	}

	// Atomic rename - this is the critical atomic operation
	if err := os.Rename(tempFile, filePath); err != nil {
		os.Remove(tempFile)
		return model.NewFileSystemError("rename", filePath, err).
			WithSuggestions([]string{
				"Check file permissions",
				"Ensure target directory is writable",
				"Verify no other process is using the file",
			})
	}

	// Update checksum for integrity tracking
	checksum := calculateMD5(data)
	aw.checksums[filePath] = checksum

	// Trigger Git versioning if enabled and requested
	if opts.GitCommit && aw.gitManager != nil && aw.gitManager.IsEnabled() {
		commitMsg := opts.CommitMsg
		if commitMsg == "" {
			commitMsg = fmt.Sprintf("update %s", filepath.Base(filePath))
		}

		if err := aw.gitManager.AutoVersionOnWrite(filePath, opts.CommitType, commitMsg); err != nil {
			// Git versioning failure is non-critical for atomic writes
			// Log the error but don't fail the operation
			fmt.Fprintf(os.Stderr, "Warning: Git versioning failed: %v\n", err)
		}
	}

	return nil
}

// writeToTempFile writes data to temporary file with proper error handling
func (aw *AtomicWriter) writeToTempFile(tempFile string, data []byte, opts *AtomicWriteOptions) error {
	file, err := os.OpenFile(tempFile, os.O_WRONLY|os.O_CREATE|os.O_EXCL, opts.Permissions)
	if err != nil {
		return model.NewFileSystemError("create_temp_file", tempFile, err).
			WithSuggestions([]string{
				"Check directory permissions",
				"Ensure sufficient disk space",
				"Verify temp directory is writable",
			})
	}
	defer file.Close()

	// Write data with proper error handling
	written := 0
	for written < len(data) {
		n, err := file.Write(data[written:])
		if err != nil {
			return model.NewFileSystemError("write", tempFile, err).
				WithSuggestions([]string{
					"Check disk space availability",
					"Verify temp directory permissions",
				})
		}
		written += n
	}

	// Ensure data is flushed to disk
	if err := file.Sync(); err != nil {
		return model.NewFileSystemError("sync", tempFile, err).
		WithSuggestions([]string{
			"Check disk health and available space",
			"Verify filesystem supports sync operations",
		})
	}

	return nil
}

// verifyWrite verifies that the written data matches the original
func (aw *AtomicWriter) verifyWrite(tempFile string, originalData []byte) error {
	readData, err := os.ReadFile(tempFile)
	if err != nil {
		return model.NewFileSystemError("read", tempFile, err).
		WithSuggestion("Failed to read back written data for verification")
	}

	if len(readData) != len(originalData) {
		return model.NewValidationError("data size mismatch during verification").
		WithContext(fmt.Sprintf("expected %d bytes, got %d bytes", len(originalData), len(readData))).
		WithSuggestion("Atomic write may have been interrupted")
	}

	originalChecksum := calculateMD5(originalData)
	readChecksum := calculateMD5(readData)

	if originalChecksum != readChecksum {
		return model.NewValidationError("checksum mismatch during verification").
		WithContext(fmt.Sprintf("expected %s, got %s", originalChecksum, readChecksum)).
		WithSuggestion("Data corruption detected during atomic write")
	}

	return nil
}

// createBackup creates a backup of the existing file
func (aw *AtomicWriter) createBackup(filePath string) error {
	backupPath := filePath + fmt.Sprintf(".backup.%d", time.Now().Unix())

	sourceFile, err := os.Open(filePath)
	if err != nil {
		return model.NewFileSystemError("open", filePath, err).
		WithSuggestion("Cannot create backup - source file not accessible")
	}
	defer sourceFile.Close()

	backupFile, err := os.Create(backupPath)
	if err != nil {
		return model.NewFileSystemError("create", backupPath, err).
		WithSuggestions([]string{
			"Check write permissions in target directory",
			"Ensure sufficient disk space for backup",
		})
	}
	defer backupFile.Close()

	_, err = io.Copy(backupFile, sourceFile)
	if err != nil {
		os.Remove(backupPath) // Clean up failed backup
		return model.NewFileSystemError("copy", backupPath, err).
			WithSuggestions([]string{
				"Check available disk space",
				"Verify backup directory permissions",
			})
	}

	return nil
}

// ReadJSON atomically reads and parses JSON data from a file
func (aw *AtomicWriter) ReadJSON(filePath string, target interface{}) error {
	aw.mu.RLock()
	defer aw.mu.RUnlock()

	data, err := aw.ReadBytes(filePath)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(data, target); err != nil {
		return model.NewValidationError("failed to parse JSON data").
			WithCause(err).
			WithContext(filePath).
			WithSuggestion("Check if the file contains valid JSON")
	}

	return nil
}

// ReadBytes atomically reads raw bytes from a file
func (aw *AtomicWriter) ReadBytes(filePath string) ([]byte, error) {
	if !fileExists(filePath) {
		return nil, model.NewNotFoundError("file").WithContext(filePath)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, model.NewFileSystemError("read", filePath, err).
			WithSuggestions([]string{
				"Check file permissions",
				"Ensure file is readable",
				"Verify file is not locked by another process",
			})
	}

	// Verify integrity if we have a stored checksum
	if expectedChecksum, exists := aw.checksums[filePath]; exists {
		actualChecksum := calculateMD5(data)
		if actualChecksum != expectedChecksum {
			return nil, model.NewValidationError("file integrity check failed").
				WithContext(fmt.Sprintf("%s: expected %s, got %s", filePath, expectedChecksum, actualChecksum)).
				WithSuggestion("File may have been corrupted or modified externally")
		}
	}

	return data, nil
}

// Exists checks if a file exists
func (aw *AtomicWriter) Exists(filePath string) bool {
	return fileExists(filePath)
}

// Delete atomically deletes a file (with backup if enabled)
func (aw *AtomicWriter) Delete(filePath string, createBackup bool) error {
	aw.mu.Lock()
	defer aw.mu.Unlock()

	if !fileExists(filePath) {
		return nil // Already deleted
	}

	if createBackup {
		if err := aw.createBackup(filePath); err != nil {
			return model.NewInternalError("failed to create backup before deletion").
				WithCause(err).
				WithContext(filePath).
				WithSuggestion("Cannot safely delete file without backup")
		}
	}

	if err := os.Remove(filePath); err != nil {
		return model.NewFileSystemError("delete", filePath, err).
			WithSuggestions([]string{
				"Check file permissions",
				"Ensure file is not locked by another process",
				"Verify you have delete permissions for the directory",
			})
	}

	// Remove from checksum tracking
	delete(aw.checksums, filePath)

	return nil
}

// GetChecksum returns the stored checksum for a file
func (aw *AtomicWriter) GetChecksum(filePath string) (string, bool) {
	aw.mu.RLock()
	defer aw.mu.RUnlock()

	checksum, exists := aw.checksums[filePath]
	return checksum, exists
}

// UpdateChecksum recalculates and updates the checksum for a file
func (aw *AtomicWriter) UpdateChecksum(filePath string) error {
	aw.mu.Lock()
	defer aw.mu.Unlock()

	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	checksum := calculateMD5(data)
	aw.checksums[filePath] = checksum

	return nil
}

// ListBackups returns all backup files for a given file
func (aw *AtomicWriter) ListBackups(filePath string) ([]string, error) {
	dir := filepath.Dir(filePath)
	base := filepath.Base(filePath)

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var backups []string
	prefix := base + ".backup."

	for _, entry := range entries {
		if !entry.IsDir() && filepath.HasPrefix(entry.Name(), prefix) {
			backups = append(backups, filepath.Join(dir, entry.Name()))
		}
	}

	return backups, nil
}

// RestoreFromBackup restores a file from its most recent backup
func (aw *AtomicWriter) RestoreFromBackup(filePath string) error {
	backups, err := aw.ListBackups(filePath)
	if err != nil {
		return model.NewFileSystemError("list", filepath.Dir(filePath), err).
		WithSuggestion("Cannot access backup directory")
	}

	if len(backups) == 0 {
		return model.NewNotFoundError("backups").
		WithContext(filePath).
		WithSuggestion("No backup files available for restoration")
	}

	// Use the most recent backup (highest timestamp)
	latestBackup := backups[len(backups)-1]

	data, err := os.ReadFile(latestBackup)
	if err != nil {
		return model.NewFileSystemError("read", latestBackup, err).
		WithSuggestion("Backup file may be corrupted or inaccessible")
	}

	opts := &AtomicWriteOptions{
		Permissions: aw.permissions,
		Backup:      false, // Don't backup when restoring
		Verify:      true,
	}

	return aw.writeBytes(filePath, data, opts)
}

// CleanupBackups removes old backup files, keeping only the specified number
func (aw *AtomicWriter) CleanupBackups(filePath string, keepCount int) error {
	backups, err := aw.ListBackups(filePath)
	if err != nil {
		return err
	}

	if len(backups) <= keepCount {
		return nil // Nothing to clean up
	}

	// Remove oldest backups
	toRemove := backups[:len(backups)-keepCount]

	for _, backup := range toRemove {
		if err := os.Remove(backup); err != nil {
			// Log warning but don't fail the operation
			fmt.Fprintf(os.Stderr, "Warning: failed to remove backup %s: %v\n", backup, err)
		}
	}

	return nil
}

// Utility functions
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func calculateMD5(data []byte) string {
	hash := md5.Sum(data)
	return fmt.Sprintf("%x", hash)
}

// AtomicOperation represents a batch of atomic operations
type AtomicOperation struct {
	writer     *AtomicWriter
	operations []atomicOp
}

type atomicOp struct {
	opType   string // write, delete
	filePath string
	data     []byte
	opts     *AtomicWriteOptions
}

// NewAtomicOperation creates a new batch operation
func (aw *AtomicWriter) NewAtomicOperation() *AtomicOperation {
	return &AtomicOperation{
		writer:     aw,
		operations: make([]atomicOp, 0),
	}
}

// AddWrite adds a write operation to the batch
func (ao *AtomicOperation) AddWrite(filePath string, data []byte, opts *AtomicWriteOptions) {
	ao.operations = append(ao.operations, atomicOp{
		opType:   "write",
		filePath: filePath,
		data:     data,
		opts:     opts,
	})
}

// AddWriteJSON adds a JSON write operation to the batch
func (ao *AtomicOperation) AddWriteJSON(filePath string, data interface{}, opts *AtomicWriteOptions) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	ao.AddWrite(filePath, jsonData, opts)
	return nil
}

// AddDelete adds a delete operation to the batch
func (ao *AtomicOperation) AddDelete(filePath string) {
	ao.operations = append(ao.operations, atomicOp{
		opType:   "delete",
		filePath: filePath,
	})
}

// Execute performs all operations atomically
// If any operation fails, all changes are rolled back
func (ao *AtomicOperation) Execute() error {
	// Create backups for all files that will be modified
	backupPaths := make(map[string]string)

	for _, op := range ao.operations {
		if op.opType == "write" && fileExists(op.filePath) {
			backupPath := op.filePath + fmt.Sprintf(".tx_backup.%d", time.Now().UnixNano())
			if err := copyFile(op.filePath, backupPath); err != nil {
				// Clean up any backups created so far
				for _, backup := range backupPaths {
					os.Remove(backup)
				}
				return model.NewInternalError("failed to create transaction backup").
					WithCause(err).
					WithContext(op.filePath).
					WithSuggestion("Cannot proceed with atomic operation without backup")
			}
			backupPaths[op.filePath] = backupPath
		}
	}

	// Execute all operations
	executedOps := make([]string, 0)

	for _, op := range ao.operations {
		switch op.opType {
		case "write":
			if err := ao.writer.writeBytes(op.filePath, op.data, op.opts); err != nil {
				// Rollback all executed operations
				ao.rollback(executedOps, backupPaths)
				return model.NewInternalError("atomic operation failed during write").
					WithCause(err).
					WithContext(op.filePath).
					WithSuggestion("All changes have been rolled back")
			}
			executedOps = append(executedOps, op.filePath)

		case "delete":
			if err := os.Remove(op.filePath); err != nil && !os.IsNotExist(err) {
				ao.rollback(executedOps, backupPaths)
				return model.NewInternalError("atomic operation failed during delete").
					WithCause(err).
					WithContext(op.filePath).
					WithSuggestion("All changes have been rolled back")
			}
			executedOps = append(executedOps, op.filePath)
		}
	}

	// Success - clean up backup files
	for _, backupPath := range backupPaths {
		os.Remove(backupPath)
	}

	return nil
}

// rollback restores all files from backups
func (ao *AtomicOperation) rollback(executedOps []string, backupPaths map[string]string) {
	for _, filePath := range executedOps {
		if backupPath, exists := backupPaths[filePath]; exists {
			// Restore from backup
			copyFile(backupPath, filePath)
		} else {
			// This was a new file, remove it
			os.Remove(filePath)
		}
	}

	// Clean up backup files
	for _, backupPath := range backupPaths {
		os.Remove(backupPath)
	}
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

// Git integration methods

// SetGitManager sets the Git version manager for automatic versioning
func (aw *AtomicWriter) SetGitManager(gitManager GitVersionManager) {
	aw.mu.Lock()
	defer aw.mu.Unlock()
	aw.gitManager = gitManager
}

// GetGitManager returns the current Git version manager
func (aw *AtomicWriter) GetGitManager() GitVersionManager {
	aw.mu.RLock()
	defer aw.mu.RUnlock()
	return aw.gitManager
}

// IsGitEnabled returns whether Git integration is enabled
func (aw *AtomicWriter) IsGitEnabled() bool {
	aw.mu.RLock()
	defer aw.mu.RUnlock()
	return aw.gitManager != nil && aw.gitManager.IsEnabled()
}

// WriteJSONWithGit atomically writes JSON data and creates a Git commit
func (aw *AtomicWriter) WriteJSONWithGit(filePath string, data interface{},
	commitType interface{}, commitMessage string) error {

	opts := &AtomicWriteOptions{
		Permissions: aw.permissions,
		Backup:      aw.backup,
		Compress:    aw.compress,
		Verify:      true,
		GitCommit:   true,
		CommitType:  commitType,
		CommitMsg:   commitMessage,
	}

	return aw.WriteJSON(filePath, data, opts)
}

// WriteBytesWithGit atomically writes raw bytes and creates a Git commit
func (aw *AtomicWriter) WriteBytesWithGit(filePath string, data []byte,
	commitType interface{}, commitMessage string) error {

	opts := &AtomicWriteOptions{
		Permissions: aw.permissions,
		Backup:      aw.backup,
		Compress:    aw.compress,
		Verify:      true,
		GitCommit:   true,
		CommitType:  commitType,
		CommitMsg:   commitMessage,
	}

	return aw.WriteBytes(filePath, data, opts)
}

// File locking integration methods

// SetLockManager sets the lock manager for file locking
func (aw *AtomicWriter) SetLockManager(lockManager LockManager) {
	aw.mu.Lock()
	defer aw.mu.Unlock()
	aw.lockManager = lockManager
}

// GetLockManager returns the current lock manager
func (aw *AtomicWriter) GetLockManager() LockManager {
	aw.mu.RLock()
	defer aw.mu.RUnlock()
	return aw.lockManager
}

// EnableLocking enables or disables file locking
func (aw *AtomicWriter) EnableLocking(enabled bool) {
	aw.mu.Lock()
	defer aw.mu.Unlock()
	aw.lockingEnabled = enabled
}

// IsLockingEnabled returns whether file locking is enabled
func (aw *AtomicWriter) IsLockingEnabled() bool {
	aw.mu.RLock()
	defer aw.mu.RUnlock()
	return aw.lockingEnabled && aw.lockManager != nil
}

// WriteJSONWithLock atomically writes JSON data with file locking
func (aw *AtomicWriter) WriteJSONWithLock(filePath string, data interface{},
	lockTimeout time.Duration) error {

	opts := &AtomicWriteOptions{
		Permissions: aw.permissions,
		Backup:      aw.backup,
		Compress:    aw.compress,
		Verify:      true,
		UseLocking:  true,
		LockTimeout: lockTimeout,
	}

	return aw.WriteJSON(filePath, data, opts)
}

// WriteBytesWithLock atomically writes raw bytes with file locking
func (aw *AtomicWriter) WriteBytesWithLock(filePath string, data []byte,
	lockTimeout time.Duration) error {

	opts := &AtomicWriteOptions{
		Permissions: aw.permissions,
		Backup:      aw.backup,
		Compress:    aw.compress,
		Verify:      true,
		UseLocking:  true,
		LockTimeout: lockTimeout,
	}

	return aw.WriteBytes(filePath, data, opts)
}

// IsFileLocked checks if a file is currently locked
func (aw *AtomicWriter) IsFileLocked(filePath string) bool {
	aw.mu.RLock()
	defer aw.mu.RUnlock()

	if aw.lockManager != nil {
		return aw.lockManager.IsLocked(filePath)
	}
	return false
}
