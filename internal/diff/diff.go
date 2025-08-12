package diff

import (
	"crypto/sha256"
	"fmt"
	"io"
	"io/fs"
	"sort"
	"strings"
)

// ChangeType represents the type of change between two file trees
type ChangeType string

const (
	ChangeNew ChangeType = "new" // file exists in A but not in B
	ChangeMod ChangeType = "mod" // file exists in both but with different content
	ChangeDel ChangeType = "del" // file exists in B but not in A
)

// Change represents a single file difference between two trees
type Change struct {
	Path string
	Type ChangeType
}

// DiffTrees compares two file trees and returns a list of changes
// Changes are reported from the perspective of A compared to B:
// - "new": file exists in A but not in B
// - "del": file exists in B but not in A  
// - "mod": file exists in both but with different content
func DiffTrees(a fs.FS, aRoot string, b fs.FS, bRoot string) ([]Change, error) {
	// Collect all files and their hashes from both trees
	aFiles, err := collectFiles(a, aRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to collect files from A: %w", err)
	}

	bFiles, err := collectFiles(b, bRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to collect files from B: %w", err)
	}

	var changes []Change

	// Find files in A but not in B (new)
	for path := range aFiles {
		if _, exists := bFiles[path]; !exists {
			changes = append(changes, Change{
				Path: path,
				Type: ChangeNew,
			})
		}
	}

	// Find files in B but not in A (deleted)
	for path := range bFiles {
		if _, exists := aFiles[path]; !exists {
			changes = append(changes, Change{
				Path: path,
				Type: ChangeDel,
			})
		}
	}

	// Find files in both with different content (modified)
	for path, aHash := range aFiles {
		if bHash, exists := bFiles[path]; exists && aHash != bHash {
			changes = append(changes, Change{
				Path: path,
				Type: ChangeMod,
			})
		}
	}

	// Sort changes by path for consistent output
	sort.Slice(changes, func(i, j int) bool {
		return changes[i].Path < changes[j].Path
	})

	return changes, nil
}

// collectFiles walks a filesystem and returns a map of relative paths to their SHA256 hashes
func collectFiles(fsys fs.FS, root string) (map[string]string, error) {
	files := make(map[string]string)

	err := fs.WalkDir(fsys, root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if d.IsDir() {
			return nil
		}

		// Calculate relative path from root
		relPath := strings.TrimPrefix(path, root)
		relPath = strings.TrimPrefix(relPath, "/")
		if relPath == "" {
			return nil // Skip if somehow we get an empty path
		}

		// Calculate file hash
		hash, err := hashFile(fsys, path)
		if err != nil {
			return fmt.Errorf("failed to hash file %s: %w", path, err)
		}

		files[relPath] = hash
		return nil
	})

	if err != nil {
		return nil, err
	}

	return files, nil
}

// hashFile calculates the SHA256 hash of a file's content
func hashFile(fsys fs.FS, path string) (string, error) {
	file, err := fsys.Open(path)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", fmt.Errorf("failed to read file content: %w", err)
	}

	return fmt.Sprintf("%x", hasher.Sum(nil)), nil
}