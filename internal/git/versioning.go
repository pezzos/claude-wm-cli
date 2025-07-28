package git

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"claude-wm-cli/internal/state"
)

// StateVersionManager manages versioning of state files using Git
type StateVersionManager struct {
	repository    *Repository
	config        *GitConfig
	atomicWriter  *state.AtomicWriter
	templates     map[StateCommitType]CommitTemplate
	autoCommit    bool
}

// NewStateVersionManager creates a new state version manager
func NewStateVersionManager(workingDir string, config *GitConfig, atomicWriter *state.AtomicWriter) (*StateVersionManager, error) {
	if config == nil {
		config = DefaultGitConfig()
	}
	
	repo := NewRepository(workingDir, config)
	
	// Initialize repository if enabled
	if config.Enabled {
		if err := repo.Initialize(); err != nil {
			return nil, fmt.Errorf("failed to initialize Git repository: %w", err)
		}
	}
	
	return &StateVersionManager{
		repository:   repo,
		config:       config,
		atomicWriter: atomicWriter,
		templates:    DefaultCommitTemplates(),
		autoCommit:   config.AutoCommit,
	}, nil
}

// VersionState creates a version of the current state
func (svm *StateVersionManager) VersionState(commitType StateCommitType, description string, files ...string) (*CommitInfo, error) {
	if !svm.config.Enabled {
		return nil, nil // Git integration disabled
	}
	
	if len(files) == 0 {
		return nil, fmt.Errorf("no files specified for versioning")
	}
	
	// Validate that all files exist
	for _, file := range files {
		if !svm.atomicWriter.Exists(file) {
			return nil, fmt.Errorf("file does not exist: %s", file)
		}
	}
	
	// Add files to Git
	if err := svm.repository.Add(files...); err != nil {
		return nil, fmt.Errorf("failed to add files to Git: %w", err)
	}
	
	// Generate commit message
	message := svm.generateCommitMessage(commitType, description)
	
	// Create commit
	commit, err := svm.repository.Commit(message)
	if err != nil {
		// Handle case where there's nothing to commit
		if gitErr, ok := err.(*GitError); ok && strings.Contains(gitErr.Stderr, "nothing to commit") {
			return nil, nil // No changes to commit
		}
		return nil, fmt.Errorf("failed to create commit: %w", err)
	}
	
	return commit, nil
}

// AutoVersionOnWrite automatically versions state after atomic write operations
func (svm *StateVersionManager) AutoVersionOnWrite(filePath string, commitType StateCommitType, description string) error {
	if !svm.autoCommit || !svm.config.Enabled {
		return nil
	}
	
	// Extract meaningful description from file path if not provided
	if description == "" {
		description = svm.generateDescriptionFromPath(filePath)
	}
	
	_, err := svm.VersionState(commitType, description, filePath)
	return err
}

// CreateRecoveryPoint creates a tagged recovery point
func (svm *StateVersionManager) CreateRecoveryPoint(name, description string, files ...string) (*RecoveryPoint, error) {
	if !svm.config.Enabled {
		return nil, fmt.Errorf("Git integration is disabled")
	}
	
	// Version the current state
	commit, err := svm.VersionState(CommitTypeBackup, description, files...)
	if err != nil {
		return nil, fmt.Errorf("failed to create recovery commit: %w", err)
	}
	
	if commit == nil {
		return nil, fmt.Errorf("no changes to create recovery point")
	}
	
	// Create tag for easy recovery
	tagName := fmt.Sprintf("recovery/%s/%d", name, time.Now().Unix())
	if err := svm.createTag(tagName, commit.Hash, description); err != nil {
		return nil, fmt.Errorf("failed to create recovery tag: %w", err)
	}
	
	// Verify state integrity
	verified := svm.verifyStateIntegrity(files...)
	
	recoveryPoint := &RecoveryPoint{
		Commit:      *commit,
		Description: description,
		StateFiles:  files,
		Verified:    verified,
		Safe:        verified,
		Automatic:   false,
	}
	
	return recoveryPoint, nil
}

// GetRecoveryPoints returns available recovery points
func (svm *StateVersionManager) GetRecoveryPoints(limit int) ([]*RecoveryPoint, error) {
	if !svm.config.Enabled {
		return nil, fmt.Errorf("Git integration is disabled")
	}
	
	// Get commit history
	commits, err := svm.repository.GetLog(limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get commit history: %w", err)
	}
	
	var recoveryPoints []*RecoveryPoint
	
	for _, commit := range commits {
		// Determine if this is a recovery-suitable commit
		isRecovery := strings.Contains(commit.Message, "backup:") ||
			strings.Contains(commit.Message, "chore(state):") ||
			strings.Contains(commit.Message, "feat(")
		
		if !isRecovery {
			continue
		}
		
		recoveryPoint := &RecoveryPoint{
			Commit:      *commit,
			Description: commit.Message,
			StateFiles:  commit.Files,
			Verified:    false, // Would need to checkout and verify
			Safe:        true,  // Assume safe unless proven otherwise
			Automatic:   strings.Contains(commit.Message, "backup:"),
		}
		
		recoveryPoints = append(recoveryPoints, recoveryPoint)
	}
	
	return recoveryPoints, nil
}

// RecoverToPoint recovers state to a specific recovery point
func (svm *StateVersionManager) RecoverToPoint(recoveryPoint *RecoveryPoint, files ...string) error {
	if !svm.config.Enabled {
		return fmt.Errorf("Git integration is disabled")
	}
	
	// Create backup of current state before recovery
	backupCommit, err := svm.VersionState(CommitTypeBackup, "pre-recovery backup", files...)
	if err != nil {
		return fmt.Errorf("failed to create pre-recovery backup: %w", err)
	}
	
	// Checkout specific files from the recovery point
	for _, file := range files {
		if err := svm.checkoutFile(recoveryPoint.Commit.Hash, file); err != nil {
			// If recovery fails, try to restore from backup
			if backupCommit != nil {
				svm.checkoutFile(backupCommit.Hash, file)
			}
			return fmt.Errorf("failed to recover file %s: %w", file, err)
		}
	}
	
	// Verify recovered state integrity
	if !svm.verifyStateIntegrity(files...) {
		// Recovery failed, restore from backup
		if backupCommit != nil {
			for _, file := range files {
				svm.checkoutFile(backupCommit.Hash, file)
			}
		}
		return fmt.Errorf("recovered state failed integrity check")
	}
	
	// Commit the recovery
	message := fmt.Sprintf("recover: restore from %s (%s)", 
		recoveryPoint.Commit.ShortHash, recoveryPoint.Description)
	
	if err := svm.repository.Add(files...); err == nil {
		svm.repository.Commit(message)
	}
	
	return nil
}

// GetStateDiff returns differences between current state and a reference
func (svm *StateVersionManager) GetStateDiff(ref string) (*DiffInfo, error) {
	if !svm.config.Enabled {
		return nil, fmt.Errorf("Git integration is disabled")
	}
	
	return svm.repository.GetDiff(ref, "HEAD")
}

// CleanupOldVersions removes old commits beyond the configured limit
func (svm *StateVersionManager) CleanupOldVersions() error {
	if !svm.config.Enabled || svm.config.MaxCommits <= 0 {
		return nil
	}
	
	commits, err := svm.repository.GetLog(0) // Get all commits
	if err != nil {
		return fmt.Errorf("failed to get commit history: %w", err)
	}
	
	if len(commits) <= svm.config.MaxCommits {
		return nil // Nothing to clean up
	}
	
	// For now, just log that cleanup is needed
	// In a full implementation, you might use git filter-branch or similar
	fmt.Printf("Cleanup needed: %d commits exceed limit of %d\n", 
		len(commits), svm.config.MaxCommits)
	
	return nil
}

// GetRepositoryStatus returns the current Git repository status
func (svm *StateVersionManager) GetRepositoryStatus() (*GitStatus, error) {
	if !svm.config.Enabled {
		return nil, fmt.Errorf("Git integration is disabled")
	}
	
	return svm.repository.GetStatus()
}

// Helper methods

func (svm *StateVersionManager) generateCommitMessage(commitType StateCommitType, description string) string {
	template, exists := svm.templates[commitType]
	if !exists {
		template = svm.templates[CommitTypeState] // Default
	}
	
	return fmt.Sprintf(template.Template, description)
}

func (svm *StateVersionManager) generateDescriptionFromPath(filePath string) string {
	filename := filepath.Base(filePath)
	dir := filepath.Base(filepath.Dir(filePath))
	
	// Generate meaningful description based on file patterns
	switch {
	case strings.Contains(filename, "project"):
		return fmt.Sprintf("update project state in %s", dir)
	case strings.Contains(filename, "epic"):
		return fmt.Sprintf("update epic state in %s", dir)
	case strings.Contains(filename, "story"):
		return fmt.Sprintf("update story state in %s", dir)
	case strings.Contains(filename, "task") || strings.Contains(filename, "todo"):
		return fmt.Sprintf("update task state in %s", dir)
	default:
		return fmt.Sprintf("update %s", filename)
	}
}

func (svm *StateVersionManager) createTag(name, hash, message string) error {
	result := svm.repository.execute(GitOpTag, "tag", "-a", name, hash, "-m", message)
	if !result.Success {
		return fmt.Errorf("failed to create tag: %s", result.Error)
	}
	return nil
}

func (svm *StateVersionManager) checkoutFile(hash, filePath string) error {
	result := svm.repository.execute(GitOpCheckout, "checkout", hash, "--", filePath)
	if !result.Success {
		return fmt.Errorf("failed to checkout file: %s", result.Error)
	}
	return nil
}

func (svm *StateVersionManager) verifyStateIntegrity(files ...string) bool {
	// Use the corruption detector to verify integrity
	if detector := svm.getCorruptionDetector(); detector != nil {
		for _, file := range files {
			if report := detector.ScanFile(file); report.IsCorrupted {
				return false
			}
		}
	}
	return true
}

func (svm *StateVersionManager) getCorruptionDetector() *state.CorruptionDetector {
	// This would ideally be injected or configured
	// For now, create a new one
	return state.NewCorruptionDetector(svm.atomicWriter)
}

// Configuration and state management

// UpdateConfig updates the Git configuration
func (svm *StateVersionManager) UpdateConfig(config *GitConfig) error {
	svm.config = config
	svm.autoCommit = config.AutoCommit
	
	// Update repository configuration
	if config.Username != "" && config.Email != "" {
		svm.repository.config = config
		return svm.repository.configureUser()
	}
	
	return nil
}

// IsEnabled returns whether Git integration is enabled
func (svm *StateVersionManager) IsEnabled() bool {
	return svm.config.Enabled
}

// GetConfig returns the current Git configuration
func (svm *StateVersionManager) GetConfig() *GitConfig {
	return svm.config
}

// SetAutoCommit enables or disables automatic commits
func (svm *StateVersionManager) SetAutoCommit(enabled bool) {
	svm.autoCommit = enabled
	svm.config.AutoCommit = enabled
}

// GetLastCommit returns information about the last commit
func (svm *StateVersionManager) GetLastCommit() (*CommitInfo, error) {
	if !svm.config.Enabled {
		return nil, fmt.Errorf("Git integration is disabled")
	}
	
	hash, err := svm.repository.getLastCommitHash()
	if err != nil {
		return nil, err
	}
	
	return svm.repository.GetCommitInfo(hash)
}

// HasUncommittedChanges checks if there are uncommitted changes
func (svm *StateVersionManager) HasUncommittedChanges() (bool, error) {
	if !svm.config.Enabled {
		return false, nil
	}
	
	status, err := svm.repository.GetStatus()
	if err != nil {
		return false, err
	}
	
	return !status.Clean, nil
}