package update

import (
	"crypto/sha256"
	"fmt"
	"io"
	"io/fs"
	"sort"
	"strings"
)

// Action represents the type of action to take during an update
type Action string

const (
	ActKeep          Action = "keep"           // no changes needed
	ActApply         Action = "apply"          // fast-forward upstream changes to local
	ActPreserveLocal Action = "preserve_local" // upstream unchanged, keep local modifications
	ActConflict      Action = "conflict"       // both upstream and local changed, needs resolution
	ActDelete        Action = "delete"         // upstream deleted file, safe to delete locally
)

// PlanEntry represents a single file action in the update plan
type PlanEntry struct {
	Path   string `json:"path"`
	Action Action `json:"action"`
	Reason string `json:"reason"`
}

// Plan represents the complete update plan with upstream changes, local changes, and merge decisions
type Plan struct {
	UpstreamChanges []PlanEntry `json:"upstream_changes"` // changes from baseline to upstream
	LocalChanges    []PlanEntry `json:"local_changes"`    // changes from baseline to local
	Merge           []PlanEntry `json:"merge"`            // final merge decisions per file
}

// BuildPlan calculates the 3-way merge plan without writing to disk
func BuildPlan(upstream fs.FS, upstreamRoot string, baseline fs.FS, baselineRoot string, local fs.FS, localRoot string) (*Plan, error) {
	// Collect file hashes from all three spaces
	upstreamFiles, err := collectFileHashes(upstream, upstreamRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to collect upstream files: %w", err)
	}

	baselineFiles, err := collectFileHashes(baseline, baselineRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to collect baseline files: %w", err)
	}

	localFiles, err := collectFileHashes(local, localRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to collect local files: %w", err)
	}

	// Get all unique file paths
	allPaths := getAllPaths(upstreamFiles, baselineFiles, localFiles)

	plan := &Plan{
		UpstreamChanges: []PlanEntry{},
		LocalChanges:    []PlanEntry{},
		Merge:           []PlanEntry{},
	}

	// Calculate changes and merge decisions for each file
	for _, path := range allPaths {
		upstreamHash := getOptionalHash(upstreamFiles, path)
		baselineHash := getOptionalHash(baselineFiles, path)
		localHash := getOptionalHash(localFiles, path)

		// Record upstream changes (baseline -> upstream)
		if upstreamChange := getChangeEntry(path, baselineHash, upstreamHash, "upstream"); upstreamChange != nil {
			plan.UpstreamChanges = append(plan.UpstreamChanges, *upstreamChange)
		}

		// Record local changes (baseline -> local)
		if localChange := getChangeEntry(path, baselineHash, localHash, "local"); localChange != nil {
			plan.LocalChanges = append(plan.LocalChanges, *localChange)
		}

		// Calculate merge decision
		action, reason := decideAction(upstreamHash, baselineHash, localHash)
		if action != ActKeep { // Only record non-keep actions to reduce noise
			plan.Merge = append(plan.Merge, PlanEntry{
				Path:   path,
				Action: action,
				Reason: reason,
			})
		}
	}

	return plan, nil
}

// collectFileHashes walks a filesystem and returns a map of relative paths to their SHA256 hashes
func collectFileHashes(fsys fs.FS, root string) (map[string]string, error) {
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
		if relPath == "" || relPath == "." {
			return nil
		}

		// Calculate file hash
		hash, err := hashFile(fsys, path)
		if err != nil {
			return fmt.Errorf("failed to hash file %s: %w", path, err)
		}

		files[relPath] = hash
		return nil
	})

	return files, err
}

// getOptionalHash returns a pointer to hash string if the file exists, nil otherwise
func getOptionalHash(files map[string]string, path string) *string {
	if hash, exists := files[path]; exists {
		return &hash
	}
	return nil
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

// getAllPaths returns a sorted list of all unique file paths from the three file maps
func getAllPaths(upstream, baseline, local map[string]string) []string {
	pathSet := make(map[string]bool)

	for path := range upstream {
		pathSet[path] = true
	}
	for path := range baseline {
		pathSet[path] = true
	}
	for path := range local {
		pathSet[path] = true
	}

	var paths []string
	for path := range pathSet {
		paths = append(paths, path)
	}

	sort.Strings(paths)
	return paths
}

// getChangeEntry creates a PlanEntry for changes between two states, or nil if no change
func getChangeEntry(path string, fromHash *string, toHash *string, context string) *PlanEntry {
	if fromHash == nil && toHash == nil {
		return nil // Both missing
	}

	if fromHash == nil && toHash != nil {
		return &PlanEntry{
			Path:   path,
			Action: "new",
			Reason: fmt.Sprintf("new file in %s", context),
		}
	}

	if fromHash != nil && toHash == nil {
		return &PlanEntry{
			Path:   path,
			Action: "deleted",
			Reason: fmt.Sprintf("deleted in %s", context),
		}
	}

	if *fromHash != *toHash {
		return &PlanEntry{
			Path:   path,
			Action: "modified",
			Reason: fmt.Sprintf("modified in %s", context),
		}
	}

	return nil // No change
}

// decideAction implements 3-way merge decision logic
func decideAction(upstreamHash *string, baselineHash *string, localHash *string) (Action, string) {
	// File exists in all three spaces
	if upstreamHash != nil && baselineHash != nil && localHash != nil {
		localUnchanged := *localHash == *baselineHash
		upstreamUnchanged := *upstreamHash == *baselineHash

		switch {
		case localUnchanged && upstreamUnchanged:
			return ActKeep, "no changes in upstream or local"
		case localUnchanged && !upstreamUnchanged:
			return ActApply, "fast-forward upstream changes (local unchanged)"
		case !localUnchanged && upstreamUnchanged:
			return ActPreserveLocal, "preserve local modifications (upstream unchanged)"
		case !localUnchanged && !upstreamUnchanged:
			return ActConflict, "both upstream and local modified"
		}
	}

	// File added in upstream only
	if upstreamHash != nil && baselineHash == nil && localHash == nil {
		return ActApply, "new file from upstream"
	}

	// File deleted in upstream, existed in baseline
	if upstreamHash == nil && baselineHash != nil {
		if localHash != nil && *localHash == *baselineHash {
			return ActDelete, "upstream deleted, local unchanged"
		} else if localHash != nil && *localHash != *baselineHash {
			return ActConflict, "upstream deleted but local modified"
		} else {
			return ActDelete, "upstream deleted, local already deleted"
		}
	}

	// File exists only in local (user created)
	if upstreamHash == nil && baselineHash == nil && localHash != nil {
		return ActPreserveLocal, "user-created file"
	}

	// File deleted in local but exists in upstream
	if upstreamHash != nil && baselineHash != nil && localHash == nil {
		if *upstreamHash == *baselineHash {
			return ActKeep, "local deleted, upstream unchanged"
		} else {
			return ActConflict, "local deleted but upstream modified"
		}
	}

	// File exists in baseline and local but not upstream
	if upstreamHash == nil && baselineHash != nil && localHash != nil {
		if *localHash == *baselineHash {
			return ActDelete, "upstream removed, local unchanged"
		} else {
			return ActConflict, "upstream removed but local modified"
		}
	}

	// Default fallback
	return ActKeep, "no action needed"
}