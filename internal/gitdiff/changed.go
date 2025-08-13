package gitdiff

import (
	"bufio"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ChangedPaths returns a list of file paths that have been modified (staged or unstaged)
// in the git repository. If git is not available or not a git repo, returns empty slice
// without error (graceful degradation).
func ChangedPaths(root string) ([]string, error) {
	// Check if this is a git repository
	if !isGitRepository(root) {
		// Not a git repo - return empty list gracefully
		return []string{}, nil
	}

	// Try to get git status using porcelain format
	paths, err := getGitStatus(root)
	if err != nil {
		// Git command failed - return empty list gracefully 
		return []string{}, nil
	}

	return paths, nil
}

// isGitRepository checks if the given directory is a git repository
func isGitRepository(root string) bool {
	gitDir := filepath.Join(root, ".git")
	if stat, err := os.Stat(gitDir); err == nil && stat.IsDir() {
		return true
	}
	return false
}

// getGitStatus executes git status --porcelain and parses the output
func getGitStatus(root string) ([]string, error) {
	// Execute git status --porcelain to get a parseable format
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = root

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	// Parse the porcelain output
	var changedFiles []string
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) < 3 {
			continue // Invalid line format
		}

		// Git status --porcelain format: "XY filename"
		// Where X is staged status, Y is unstaged status
		// We want both staged and unstaged files
		filename := line[3:] // Skip the two status chars and space

		// Handle quoted filenames (contain spaces or special chars)
		if strings.HasPrefix(filename, "\"") && strings.HasSuffix(filename, "\"") {
			// Git quotes filenames with special characters
			// For simplicity, just remove the quotes - this handles most cases
			filename = filename[1 : len(filename)-1]
			// Note: A full implementation would need to handle escape sequences
			// but for this use case, removing quotes is sufficient
		}

		changedFiles = append(changedFiles, filename)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return changedFiles, nil
}