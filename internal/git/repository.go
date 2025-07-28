package git

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Repository provides a wrapper around Git operations with proven patterns
type Repository struct {
	config     *GitConfig
	workingDir string
	timeout    time.Duration
}

// NewRepository creates a new Git repository wrapper
func NewRepository(workingDir string, config *GitConfig) *Repository {
	if config == nil {
		config = DefaultGitConfig()
	}

	return &Repository{
		config:     config,
		workingDir: workingDir,
		timeout:    30 * time.Second, // Proven 30s timeout pattern
	}
}

// Initialize initializes a Git repository if it doesn't exist
func (r *Repository) Initialize() error {
	if r.IsRepository() {
		return nil // Already initialized
	}

	result := r.execute(GitOpInit, "init")
	if !result.Success {
		return &GitError{
			Operation:   GitOpInit,
			Command:     result.Command,
			ExitCode:    result.ExitCode,
			Stderr:      result.Error,
			WorkingDir:  r.workingDir,
			Suggestion:  "Check directory permissions and ensure Git is installed",
			Recoverable: true,
			Timestamp:   time.Now(),
		}
	}

	// Create initial .gitignore for temp files
	if err := r.createGitignore(); err != nil {
		return fmt.Errorf("failed to create .gitignore: %w", err)
	}

	// Configure user if provided
	if r.config.Username != "" && r.config.Email != "" {
		if err := r.configureUser(); err != nil {
			return fmt.Errorf("failed to configure Git user: %w", err)
		}
	}

	return nil
}

// IsRepository checks if the current directory is a Git repository
func (r *Repository) IsRepository() bool {
	gitDir := filepath.Join(r.workingDir, ".git")
	if stat, err := os.Stat(gitDir); err == nil {
		return stat.IsDir()
	}

	// Check if we're in a subdirectory of a Git repository
	result := r.execute(GitOpStatus, "rev-parse", "--git-dir")
	return result.Success
}

// Add stages files for commit
func (r *Repository) Add(files ...string) error {
	if len(files) == 0 {
		return fmt.Errorf("no files specified to add")
	}

	args := append([]string{"add"}, files...)
	result := r.execute(GitOpAdd, args...)

	if !result.Success {
		return &GitError{
			Operation:   GitOpAdd,
			Command:     result.Command,
			ExitCode:    result.ExitCode,
			Stderr:      result.Error,
			WorkingDir:  r.workingDir,
			Suggestion:  "Check if files exist and are not ignored",
			Recoverable: true,
			Timestamp:   time.Now(),
		}
	}

	return nil
}

// Commit creates a new commit with the specified message
func (r *Repository) Commit(message string) (*CommitInfo, error) {
	if message == "" {
		return nil, fmt.Errorf("commit message cannot be empty")
	}

	result := r.execute(GitOpCommit, "commit", "-m", message)
	if !result.Success {
		// Check if there's nothing to commit
		if strings.Contains(result.Error, "nothing to commit") {
			return nil, &GitError{
				Operation:   GitOpCommit,
				Command:     result.Command,
				ExitCode:    result.ExitCode,
				Stderr:      "Nothing to commit, working tree clean",
				WorkingDir:  r.workingDir,
				Suggestion:  "Make changes to files before committing",
				Recoverable: true,
				Timestamp:   time.Now(),
			}
		}

		return nil, &GitError{
			Operation:   GitOpCommit,
			Command:     result.Command,
			ExitCode:    result.ExitCode,
			Stderr:      result.Error,
			WorkingDir:  r.workingDir,
			Suggestion:  "Check if files are staged and user is configured",
			Recoverable: true,
			Timestamp:   time.Now(),
		}
	}

	// Get the commit hash from the output
	hash, err := r.getLastCommitHash()
	if err != nil {
		return nil, fmt.Errorf("failed to get commit hash: %w", err)
	}

	return r.GetCommitInfo(hash)
}

// GetStatus returns the current Git status
func (r *Repository) GetStatus() (*GitStatus, error) {
	result := r.execute(GitOpStatus, "status", "--porcelain", "-b")
	if !result.Success {
		return nil, &GitError{
			Operation:   GitOpStatus,
			Command:     result.Command,
			ExitCode:    result.ExitCode,
			Stderr:      result.Error,
			WorkingDir:  r.workingDir,
			Suggestion:  "Ensure you're in a Git repository",
			Recoverable: true,
			Timestamp:   time.Now(),
		}
	}

	return r.parseStatus(result.Output)
}

// GetLog returns commit history
func (r *Repository) GetLog(limit int) ([]*CommitInfo, error) {
	args := []string{"log", "--oneline", "--format=%H|%h|%s|%an|%ae|%ad", "--date=iso"}
	if limit > 0 {
		args = append(args, fmt.Sprintf("-%d", limit))
	}

	result := r.execute(GitOpLog, args...)
	if !result.Success {
		return nil, &GitError{
			Operation:   GitOpLog,
			Command:     result.Command,
			ExitCode:    result.ExitCode,
			Stderr:      result.Error,
			WorkingDir:  r.workingDir,
			Suggestion:  "Check if repository has commits",
			Recoverable: true,
			Timestamp:   time.Now(),
		}
	}

	return r.parseLog(result.Output)
}

// GetCommitInfo returns detailed information about a specific commit
func (r *Repository) GetCommitInfo(hash string) (*CommitInfo, error) {
	result := r.execute(GitOpLog, "show", "--format=%H|%h|%s|%an|%ae|%ad|%P", "--name-only", "--date=iso", hash)
	if !result.Success {
		return nil, &GitError{
			Operation:   GitOpLog,
			Command:     result.Command,
			ExitCode:    result.ExitCode,
			Stderr:      result.Error,
			WorkingDir:  r.workingDir,
			Suggestion:  "Check if the commit hash is valid",
			Recoverable: true,
			Timestamp:   time.Now(),
		}
	}

	return r.parseCommitInfo(result.Output)
}

// Checkout switches to a different branch or commit
func (r *Repository) Checkout(ref string) error {
	result := r.execute(GitOpCheckout, "checkout", ref)
	if !result.Success {
		return &GitError{
			Operation:   GitOpCheckout,
			Command:     result.Command,
			ExitCode:    result.ExitCode,
			Stderr:      result.Error,
			WorkingDir:  r.workingDir,
			Suggestion:  "Check if the reference exists and working directory is clean",
			Recoverable: true,
			Timestamp:   time.Now(),
		}
	}

	return nil
}

// Reset resets the repository to a specific commit
func (r *Repository) Reset(hash string, hard bool) error {
	args := []string{"reset"}
	if hard {
		args = append(args, "--hard")
	}
	args = append(args, hash)

	result := r.execute(GitOpReset, args...)
	if !result.Success {
		return &GitError{
			Operation:   GitOpReset,
			Command:     result.Command,
			ExitCode:    result.ExitCode,
			Stderr:      result.Error,
			WorkingDir:  r.workingDir,
			Suggestion:  "Check if the commit hash is valid",
			Recoverable: false, // Hard reset is not easily recoverable
			Timestamp:   time.Now(),
		}
	}

	return nil
}

// GetDiff returns differences between two references
func (r *Repository) GetDiff(from, to string) (*DiffInfo, error) {
	args := []string{"diff", "--stat"}
	if from != "" && to != "" {
		args = append(args, fmt.Sprintf("%s..%s", from, to))
	} else if from != "" {
		args = append(args, from)
	}

	result := r.execute(GitOpDiff, args...)
	if !result.Success {
		return nil, &GitError{
			Operation:   GitOpDiff,
			Command:     result.Command,
			ExitCode:    result.ExitCode,
			Stderr:      result.Error,
			WorkingDir:  r.workingDir,
			Suggestion:  "Check if the references are valid",
			Recoverable: true,
			Timestamp:   time.Now(),
		}
	}

	return r.parseDiff(result.Output, from, to)
}

// CreateBranch creates a new branch
func (r *Repository) CreateBranch(name string, fromRef string) error {
	args := []string{"checkout", "-b", name}
	if fromRef != "" {
		args = append(args, fromRef)
	}

	result := r.execute(GitOpBranch, args...)
	if !result.Success {
		return &GitError{
			Operation:   GitOpBranch,
			Command:     result.Command,
			ExitCode:    result.ExitCode,
			Stderr:      result.Error,
			WorkingDir:  r.workingDir,
			Suggestion:  "Check if branch name is valid and doesn't exist",
			Recoverable: true,
			Timestamp:   time.Now(),
		}
	}

	return nil
}

// execute runs a Git command with proven timeout and error handling patterns
func (r *Repository) execute(operation GitOperation, args ...string) *GitResult {
	start := time.Now()

	// Create context with timeout (proven 30s pattern)
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()

	// Prepare command
	cmdArgs := append([]string{"git"}, args...)
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = r.workingDir

	// Capture output
	output, err := cmd.CombinedOutput()
	duration := time.Since(start)

	result := &GitResult{
		Operation:  operation,
		Success:    err == nil,
		Command:    strings.Join(cmdArgs, " "),
		WorkingDir: r.workingDir,
		Duration:   duration,
		Timestamp:  start,
	}

	if err != nil {
		result.Error = string(output)
		if exitError, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitError.ExitCode()
		} else {
			result.ExitCode = 1
		}
	} else {
		result.Output = string(output)
		result.ExitCode = 0
	}

	return result
}

// Helper methods for parsing Git output

func (r *Repository) parseStatus(output string) (*GitStatus, error) {
	status := &GitStatus{
		Files: make([]FileStatus, 0),
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) == 0 {
		return status, nil
	}

	// Parse branch line (first line starts with ##)
	if len(lines) > 0 && strings.HasPrefix(lines[0], "##") {
		branchLine := strings.TrimPrefix(lines[0], "## ")
		if strings.Contains(branchLine, "...") {
			parts := strings.Split(branchLine, "...")
			status.Branch = parts[0]
			if len(parts) > 1 {
				remotePart := parts[1]
				if strings.Contains(remotePart, " ") {
					remoteInfo := strings.Fields(remotePart)
					status.Remote = remoteInfo[0]
					// Parse ahead/behind info if present
					for _, info := range remoteInfo[1:] {
						if strings.HasPrefix(info, "[ahead") {
							// Parse ahead count
							re := regexp.MustCompile(`ahead (\d+)`)
							if matches := re.FindStringSubmatch(info); len(matches) > 1 {
								status.Ahead, _ = strconv.Atoi(matches[1])
							}
						}
						if strings.HasPrefix(info, "behind") {
							// Parse behind count
							re := regexp.MustCompile(`behind (\d+)`)
							if matches := re.FindStringSubmatch(info); len(matches) > 1 {
								status.Behind, _ = strconv.Atoi(matches[1])
							}
						}
					}
				} else {
					status.Remote = remotePart
				}
			}
		} else {
			status.Branch = branchLine
		}
		lines = lines[1:] // Remove branch line
	}

	// Parse file status lines
	for _, line := range lines {
		if len(line) < 3 {
			continue
		}

		stagedStatus := line[0]
		modifiedStatus := line[1]
		path := strings.TrimSpace(line[3:])

		fileStatus := FileStatus{
			Path:       path,
			Staged:     stagedStatus != ' ' && stagedStatus != '?',
			Modified:   modifiedStatus != ' ',
			Untracked:  stagedStatus == '?' && modifiedStatus == '?',
			Conflicted: stagedStatus == 'U' || modifiedStatus == 'U',
		}

		// Determine overall status
		if fileStatus.Conflicted {
			fileStatus.Status = "U"
			status.Conflicted++
		} else if fileStatus.Untracked {
			fileStatus.Status = "?"
			status.Untracked++
		} else if fileStatus.Staged {
			fileStatus.Status = string(stagedStatus)
			status.Staged++
		} else if fileStatus.Modified {
			fileStatus.Status = string(modifiedStatus)
			status.Modified++
		}

		status.Files = append(status.Files, fileStatus)
	}

	status.Clean = len(status.Files) == 0
	return status, nil
}

func (r *Repository) parseLog(output string) ([]*CommitInfo, error) {
	var commits []*CommitInfo

	lines := strings.Split(strings.TrimSpace(output), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.Split(line, "|")
		if len(parts) < 6 {
			continue
		}

		date, err := time.Parse("2006-01-02 15:04:05 -0700", parts[5])
		if err != nil {
			date = time.Now() // Fallback
		}

		commit := &CommitInfo{
			Hash:      parts[0],
			ShortHash: parts[1],
			Message:   parts[2],
			Author:    parts[3],
			Email:     parts[4],
			Date:      date,
		}

		commits = append(commits, commit)
	}

	return commits, nil
}

func (r *Repository) parseCommitInfo(output string) (*CommitInfo, error) {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) == 0 {
		return nil, fmt.Errorf("no commit info found")
	}

	// First line contains commit metadata
	parts := strings.Split(lines[0], "|")
	if len(parts) < 6 {
		return nil, fmt.Errorf("invalid commit info format")
	}

	date, err := time.Parse("2006-01-02 15:04:05 -0700", parts[5])
	if err != nil {
		date = time.Now()
	}

	commit := &CommitInfo{
		Hash:      parts[0],
		ShortHash: parts[1],
		Message:   parts[2],
		Author:    parts[3],
		Email:     parts[4],
		Date:      date,
	}

	if len(parts) > 6 && parts[6] != "" {
		commit.Parent = parts[6]
	}

	// Remaining lines are file names
	files := make([]string, 0)
	for i := 1; i < len(lines); i++ {
		if line := strings.TrimSpace(lines[i]); line != "" {
			files = append(files, line)
		}
	}
	commit.Files = files

	return commit, nil
}

func (r *Repository) parseDiff(output, from, to string) (*DiffInfo, error) {
	diff := &DiffInfo{
		FromRef:   from,
		ToRef:     to,
		Files:     make([]FileDiff, 0),
		Generated: time.Now(),
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Parse file change line (format: "insertions(+), deletions(-), filename")
		parts := strings.Fields(line)
		if len(parts) >= 3 {
			filename := parts[len(parts)-1]

			// Extract insertions/deletions if present
			var insertions, deletions int
			for _, part := range parts[:len(parts)-1] {
				if strings.Contains(part, "+") {
					insertions, _ = strconv.Atoi(strings.TrimSuffix(part, "(+)"))
				}
				if strings.Contains(part, "-") {
					deletions, _ = strconv.Atoi(strings.TrimSuffix(part, "(-)"))
				}
			}

			fileDiff := FileDiff{
				Path:       filename,
				Insertions: insertions,
				Deletions:  deletions,
			}

			diff.Files = append(diff.Files, fileDiff)
			diff.Insertions += insertions
			diff.Deletions += deletions
			diff.FilesChanged++
		}
	}

	return diff, nil
}

func (r *Repository) getLastCommitHash() (string, error) {
	result := r.execute(GitOpLog, "rev-parse", "HEAD")
	if !result.Success {
		return "", fmt.Errorf("failed to get last commit hash: %s", result.Error)
	}

	return strings.TrimSpace(result.Output), nil
}

func (r *Repository) createGitignore() error {
	gitignorePath := filepath.Join(r.workingDir, ".gitignore")

	// Check if .gitignore already exists
	if _, err := os.Stat(gitignorePath); err == nil {
		return nil // Already exists
	}

	gitignoreContent := `# Claude WM CLI temporary files
.tmp_*
*.backup.*
*.tx_backup.*

# IDE and editor files
.vscode/
.idea/
*.swp
*.swo
*~

# OS generated files
.DS_Store
.DS_Store?
._*
.Spotlight-V100
.Trashes
ehthumbs.db
Thumbs.db

# Build artifacts
dist/
build/
*.exe
*.dll
*.so
*.dylib

# Logs
*.log
logs/

# Environment variables
.env
.env.local
.env.*.local
`

	return os.WriteFile(gitignorePath, []byte(gitignoreContent), 0644)
}

func (r *Repository) configureUser() error {
	// Set user name
	result := r.execute(GitOpInit, "config", "user.name", r.config.Username)
	if !result.Success {
		return fmt.Errorf("failed to set Git user name: %s", result.Error)
	}

	// Set user email
	result = r.execute(GitOpInit, "config", "user.email", r.config.Email)
	if !result.Success {
		return fmt.Errorf("failed to set Git user email: %s", result.Error)
	}

	return nil
}

// GetBranches returns information about all branches
func (r *Repository) GetBranches() ([]*BranchInfo, error) {
	result := r.execute(GitOpBranch, "branch", "-v")
	if !result.Success {
		return nil, &GitError{
			Operation:   GitOpBranch,
			Command:     result.Command,
			ExitCode:    result.ExitCode,
			Stderr:      result.Error,
			WorkingDir:  r.workingDir,
			Suggestion:  "Ensure you're in a Git repository",
			Recoverable: true,
			Timestamp:   time.Now(),
		}
	}

	return r.parseBranches(result.Output)
}

func (r *Repository) parseBranches(output string) ([]*BranchInfo, error) {
	var branches []*BranchInfo

	lines := strings.Split(strings.TrimSpace(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		branch := &BranchInfo{
			UpdatedAt: time.Now(),
		}

		// Check if this is the current branch (starts with *)
		if strings.HasPrefix(line, "*") {
			branch.Current = true
			line = strings.TrimPrefix(line, "* ")
		} else {
			line = strings.TrimPrefix(line, "  ")
		}

		// Parse branch name and last commit
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			branch.Name = parts[0]
			branch.LastCommit = parts[1]
		}

		branches = append(branches, branch)
	}

	return branches, nil
}
