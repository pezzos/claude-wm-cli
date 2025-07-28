package git

import (
	"time"
)

// GitConfig contains configuration for Git integration
type GitConfig struct {
	Enabled          bool   `json:"enabled"`
	RepositoryPath   string `json:"repository_path"`
	AutoCommit       bool   `json:"auto_commit"`
	CommitMessage    string `json:"commit_message_template"`
	Branch           string `json:"branch"`
	RemoteURL        string `json:"remote_url,omitempty"`
	Username         string `json:"username,omitempty"`
	Email            string `json:"email,omitempty"`
	MaxCommits       int    `json:"max_commits"`        // Maximum commits to keep
	AutoPush         bool   `json:"auto_push"`          // Auto push to remote
	ConflictStrategy string `json:"conflict_strategy"`  // merge, rebase, manual
}

// DefaultGitConfig returns the default Git configuration
func DefaultGitConfig() *GitConfig {
	return &GitConfig{
		Enabled:          true,
		RepositoryPath:   ".git",
		AutoCommit:       true,
		CommitMessage:    "chore: update state - %s",
		Branch:           "main",
		MaxCommits:       100,
		AutoPush:         false,
		ConflictStrategy: "merge",
	}
}

// GitOperation represents the type of Git operation
type GitOperation string

const (
	GitOpInit       GitOperation = "init"
	GitOpAdd        GitOperation = "add"
	GitOpCommit     GitOperation = "commit"
	GitOpPush       GitOperation = "push"
	GitOpPull       GitOperation = "pull"
	GitOpCheckout   GitOperation = "checkout"
	GitOpReset      GitOperation = "reset"
	GitOpLog        GitOperation = "log"
	GitOpStatus     GitOperation = "status"
	GitOpDiff       GitOperation = "diff"
	GitOpBranch     GitOperation = "branch"
	GitOpMerge      GitOperation = "merge"
	GitOpRebase     GitOperation = "rebase"
	GitOpStash      GitOperation = "stash"
	GitOpTag        GitOperation = "tag"
)

// GitResult represents the result of a Git operation
type GitResult struct {
	Operation   GitOperation  `json:"operation"`
	Success     bool          `json:"success"`
	Output      string        `json:"output"`
	Error       string        `json:"error,omitempty"`
	ExitCode    int           `json:"exit_code"`
	Duration    time.Duration `json:"duration"`
	Command     string        `json:"command"`
	WorkingDir  string        `json:"working_dir"`
	Timestamp   time.Time     `json:"timestamp"`
}

// CommitInfo represents information about a Git commit
type CommitInfo struct {
	Hash        string    `json:"hash"`
	ShortHash   string    `json:"short_hash"`
	Message     string    `json:"message"`
	Author      string    `json:"author"`
	Email       string    `json:"email"`
	Date        time.Time `json:"date"`
	Files       []string  `json:"files"`
	Insertions  int       `json:"insertions"`
	Deletions   int       `json:"deletions"`
	Parent      string    `json:"parent,omitempty"`
	Tags        []string  `json:"tags,omitempty"`
}

// BranchInfo represents information about a Git branch
type BranchInfo struct {
	Name      string    `json:"name"`
	Current   bool      `json:"current"`
	Remote    string    `json:"remote,omitempty"`
	Upstream  string    `json:"upstream,omitempty"`
	LastCommit string   `json:"last_commit"`
	Behind    int       `json:"behind"`
	Ahead     int       `json:"ahead"`
	UpdatedAt time.Time `json:"updated_at"`
}

// FileStatus represents the status of a file in Git
type FileStatus struct {
	Path       string `json:"path"`
	Status     string `json:"status"`     // M, A, D, R, C, U, ?, !
	Staged     bool   `json:"staged"`
	Modified   bool   `json:"modified"`
	Untracked  bool   `json:"untracked"`
	Ignored    bool   `json:"ignored"`
	Conflicted bool   `json:"conflicted"`
}

// GitStatus represents the overall Git repository status
type GitStatus struct {
	Branch        string       `json:"branch"`
	Remote        string       `json:"remote,omitempty"`
	Behind        int          `json:"behind"`
	Ahead         int          `json:"ahead"`
	Clean         bool         `json:"clean"`
	Files         []FileStatus `json:"files"`
	Staged        int          `json:"staged"`
	Modified      int          `json:"modified"`
	Untracked     int          `json:"untracked"`
	Conflicted    int          `json:"conflicted"`
	LastCommit    string       `json:"last_commit,omitempty"`
	LastCommitMsg string       `json:"last_commit_message,omitempty"`
}

// RecoveryPoint represents a point in Git history that can be used for recovery
type RecoveryPoint struct {
	Commit      CommitInfo `json:"commit"`
	Description string     `json:"description"`
	StateFiles  []string   `json:"state_files"`
	Size        int64      `json:"size"`
	Verified    bool       `json:"verified"`     // Whether state integrity was verified
	Safe        bool       `json:"safe"`         // Whether recovery is considered safe
	Automatic   bool       `json:"automatic"`    // Whether this was an automatic checkpoint
}

// DiffInfo represents information about differences between Git references
type DiffInfo struct {
	FromRef     string         `json:"from_ref"`
	ToRef       string         `json:"to_ref"`
	Files       []FileDiff     `json:"files"`
	Insertions  int            `json:"insertions"`
	Deletions   int            `json:"deletions"`
	FilesChanged int           `json:"files_changed"`
	Binary      bool           `json:"binary"`
	Generated   time.Time      `json:"generated"`
}

// FileDiff represents differences in a single file
type FileDiff struct {
	Path        string   `json:"path"`
	Status      string   `json:"status"`     // A, M, D, R, C
	Insertions  int      `json:"insertions"`
	Deletions   int      `json:"deletions"`
	Binary      bool     `json:"binary"`
	Renamed     bool     `json:"renamed"`
	OldPath     string   `json:"old_path,omitempty"`
	Hunks       []string `json:"hunks,omitempty"`
}

// ConflictInfo represents information about merge conflicts
type ConflictInfo struct {
	Files       []string  `json:"files"`
	Count       int       `json:"count"`
	Resolvable  bool      `json:"resolvable"`
	Strategy    string    `json:"strategy"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

// GitError represents a Git-specific error
type GitError struct {
	Operation   GitOperation `json:"operation"`
	Command     string       `json:"command"`
	ExitCode    int          `json:"exit_code"`
	Stderr      string       `json:"stderr"`
	Stdout      string       `json:"stdout"`
	WorkingDir  string       `json:"working_dir"`
	Suggestion  string       `json:"suggestion,omitempty"`
	Recoverable bool         `json:"recoverable"`
	Timestamp   time.Time    `json:"timestamp"`
}

func (e GitError) Error() string {
	if e.Suggestion != "" {
		return e.Stderr + " (suggestion: " + e.Suggestion + ")"
	}
	return e.Stderr
}

// StateCommitType represents different types of state commits
type StateCommitType string

const (
	CommitTypeProject  StateCommitType = "project"
	CommitTypeEpic     StateCommitType = "epic"
	CommitTypeStory    StateCommitType = "story"
	CommitTypeTask     StateCommitType = "task"
	CommitTypeState    StateCommitType = "state"
	CommitTypeBackup   StateCommitType = "backup"
	CommitTypeRecovery StateCommitType = "recovery"
	CommitTypeMigration StateCommitType = "migration"
)

// CommitTemplate represents a template for generating commit messages
type CommitTemplate struct {
	Type        StateCommitType `json:"type"`
	Template    string          `json:"template"`
	Description string          `json:"description"`
	Example     string          `json:"example"`
}

// DefaultCommitTemplates returns the default commit message templates
func DefaultCommitTemplates() map[StateCommitType]CommitTemplate {
	return map[StateCommitType]CommitTemplate{
		CommitTypeProject: {
			Type:        CommitTypeProject,
			Template:    "feat(project): %s",
			Description: "Project-level changes",
			Example:     "feat(project): initialize new project 'claude-wm-cli'",
		},
		CommitTypeEpic: {
			Type:        CommitTypeEpic,
			Template:    "feat(epic): %s",
			Description: "Epic creation or updates",
			Example:     "feat(epic): add CLI foundation epic",
		},
		CommitTypeStory: {
			Type:        CommitTypeStory,
			Template:    "feat(story): %s",
			Description: "Story creation or updates",
			Example:     "feat(story): implement command execution",
		},
		CommitTypeTask: {
			Type:        CommitTypeTask,
			Template:    "fix(task): %s",
			Description: "Task updates and completion",
			Example:     "fix(task): complete atomic file operations",
		},
		CommitTypeState: {
			Type:        CommitTypeState,
			Template:    "chore(state): %s",
			Description: "General state updates",
			Example:     "chore(state): update progress metrics",
		},
		CommitTypeBackup: {
			Type:        CommitTypeBackup,
			Template:    "backup: %s",
			Description: "Automatic backup commits",
			Example:     "backup: automatic state backup before operation",
		},
		CommitTypeRecovery: {
			Type:        CommitTypeRecovery,
			Template:    "recover: %s",
			Description: "Recovery operations",
			Example:     "recover: restore state from corruption",
		},
		CommitTypeMigration: {
			Type:        CommitTypeMigration,
			Template:    "migrate: %s",
			Description: "Schema migrations",
			Example:     "migrate: upgrade schema from v1.0.0 to v1.1.0",
		},
	}
}

// TagInfo represents information about a Git tag
type TagInfo struct {
	Name        string    `json:"name"`
	Hash        string    `json:"hash"`
	Message     string    `json:"message,omitempty"`
	Tagger      string    `json:"tagger,omitempty"`
	Date        time.Time `json:"date"`
	Lightweight bool      `json:"lightweight"`
}

// RemoteInfo represents information about a Git remote
type RemoteInfo struct {
	Name      string `json:"name"`
	URL       string `json:"url"`
	FetchURL  string `json:"fetch_url"`
	PushURL   string `json:"push_url"`
	Type      string `json:"type"` // origin, upstream, etc.
}

// GitHookType represents different types of Git hooks
type GitHookType string

const (
	HookPreCommit   GitHookType = "pre-commit"
	HookPostCommit  GitHookType = "post-commit"
	HookPrePush     GitHookType = "pre-push"
	HookPostReceive GitHookType = "post-receive"
	HookPreReceive  GitHookType = "pre-receive"
)

// GitHook represents a Git hook configuration
type GitHook struct {
	Type        GitHookType `json:"type"`
	Script      string      `json:"script"`
	Enabled     bool        `json:"enabled"`
	Description string      `json:"description"`
}