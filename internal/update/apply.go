package update

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"claude-wm-cli/internal/fsutil"
)

// ApplyPlan applies the update plan to disk with atomic operations
func ApplyPlan(plan *Plan, upstream fs.FS, upstreamRoot string, baselineDir, localDir string) error {
	if plan == nil {
		return fmt.Errorf("plan cannot be nil")
	}

	// Only apply merge actions (skip upstream/local change tracking)
	if len(plan.Merge) == 0 {
		return nil // No actions to apply
	}

	// Create temporary directory for atomic operations
	tempDir := filepath.Join(filepath.Dir(localDir), fmt.Sprintf("tmp-update-%d", time.Now().Unix()))
	if err := os.RemoveAll(tempDir); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to clean existing temp directory: %w", err)
	}

	// Ensure cleanup of temp directory
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			fmt.Printf("Warning: failed to clean up temp directory %s: %v\n", tempDir, err)
		}
	}()

	// Copy current local directory to temp directory as starting point
	if err := fsutil.CopyDirectory(localDir, tempDir); err != nil {
		return fmt.Errorf("failed to create temporary working directory: %w", err)
	}

	// Track conflicts for reporting
	var conflicts []string
	var applied []string

	// Apply each merge action to the temporary directory
	for _, action := range plan.Merge {
		switch action.Action {
		case ActApply:
			if err := applyFile(upstream, upstreamRoot, tempDir, action.Path); err != nil {
				return fmt.Errorf("failed to apply %s: %w", action.Path, err)
			}
			applied = append(applied, action.Path)

		case ActPreserveLocal:
			// No action needed - file stays as is in temp directory
			applied = append(applied, action.Path)

		case ActDelete:
			if err := deleteFile(tempDir, action.Path); err != nil {
				return fmt.Errorf("failed to delete %s: %w", action.Path, err)
			}
			applied = append(applied, action.Path)

		case ActConflict:
			if err := createConflictFiles(upstream, upstreamRoot, localDir, tempDir, action.Path); err != nil {
				return fmt.Errorf("failed to create conflict files for %s: %w", action.Path, err)
			}
			conflicts = append(conflicts, action.Path)

		case ActKeep:
			// No action needed
			applied = append(applied, action.Path)

		default:
			return fmt.Errorf("unknown action type: %s for file %s", action.Action, action.Path)
		}
	}

	// If there are unresolved conflicts, report them but don't fail
	if len(conflicts) > 0 {
		fmt.Printf("⚠️  Created conflict files for %d files that need manual resolution:\n", len(conflicts))
		for _, path := range conflicts {
			fmt.Printf("   %s (see %s.conflict, %s.upstream)\n", path, path, path)
		}
	}

	// Atomic switch: replace local directory with temp directory
	backupLocal := localDir + ".backup-" + fmt.Sprintf("%d", time.Now().Unix())
	if err := os.Rename(localDir, backupLocal); err != nil {
		return fmt.Errorf("failed to backup current local directory: %w", err)
	}

	if err := os.Rename(tempDir, localDir); err != nil {
		// Try to restore backup
		if restoreErr := os.Rename(backupLocal, localDir); restoreErr != nil {
			return fmt.Errorf("failed to apply update AND failed to restore backup: apply error: %w, restore error: %v", err, restoreErr)
		}
		return fmt.Errorf("failed to apply update (backup restored): %w", err)
	}

	// Remove the backup now that update succeeded
	if err := os.RemoveAll(backupLocal); err != nil {
		fmt.Printf("Warning: failed to clean up backup directory %s: %v\n", backupLocal, err)
	}

	// Update baseline to reflect new upstream state
	if err := updateBaseline(upstream, upstreamRoot, baselineDir); err != nil {
		fmt.Printf("Warning: update succeeded but failed to update baseline: %v\n", err)
	}

	// Report results
	fmt.Printf("✅ Successfully applied %d changes", len(applied))
	if len(conflicts) > 0 {
		fmt.Printf(" (%d conflicts need manual resolution)", len(conflicts))
	}
	fmt.Println()

	return nil
}

// applyFile copies a file from upstream to the target directory
func applyFile(upstream fs.FS, upstreamRoot, targetDir, relPath string) error {
	upstreamPath := filepath.Join(upstreamRoot, relPath)
	targetPath := filepath.Join(targetDir, relPath)

	// Ensure target directory exists
	if err := fsutil.EnsureDir(filepath.Dir(targetPath)); err != nil {
		return err
	}

	// Open upstream file
	srcFile, err := upstream.Open(upstreamPath)
	if err != nil {
		return fmt.Errorf("failed to open upstream file: %w", err)
	}
	defer srcFile.Close()

	// Create target file
	targetFile, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to create target file: %w", err)
	}
	defer targetFile.Close()

	// Copy content
	if _, err := io.Copy(targetFile, srcFile); err != nil {
		return fmt.Errorf("failed to copy file content: %w", err)
	}

	return nil
}

// deleteFile removes a file from the target directory
func deleteFile(targetDir, relPath string) error {
	targetPath := filepath.Join(targetDir, relPath)
	if err := os.Remove(targetPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

// createConflictFiles creates conflict resolution files
func createConflictFiles(upstream fs.FS, upstreamRoot, localDir, targetDir, relPath string) error {
	upstreamPath := filepath.Join(upstreamRoot, relPath)
	localPath := filepath.Join(localDir, relPath)
	targetPath := filepath.Join(targetDir, relPath)

	// Ensure target directory exists
	if err := fsutil.EnsureDir(filepath.Dir(targetPath)); err != nil {
		return err
	}

	// Create .upstream file with upstream version
	upstreamFile, err := upstream.Open(upstreamPath)
	if err != nil {
		return fmt.Errorf("failed to open upstream file: %w", err)
	}
	defer upstreamFile.Close()

	upstreamTargetPath := targetPath + ".upstream"
	upstreamTarget, err := os.Create(upstreamTargetPath)
	if err != nil {
		return fmt.Errorf("failed to create upstream conflict file: %w", err)
	}
	defer upstreamTarget.Close()

	if _, err := io.Copy(upstreamTarget, upstreamFile); err != nil {
		return fmt.Errorf("failed to write upstream conflict file: %w", err)
	}

	// Create .conflict file with local version (if it exists)
	if _, err := os.Stat(localPath); err == nil {
		localFile, err := os.Open(localPath)
		if err != nil {
			return fmt.Errorf("failed to open local file: %w", err)
		}
		defer localFile.Close()

		conflictTargetPath := targetPath + ".conflict"
		conflictTarget, err := os.Create(conflictTargetPath)
		if err != nil {
			return fmt.Errorf("failed to create conflict file: %w", err)
		}
		defer conflictTarget.Close()

		if _, err := io.Copy(conflictTarget, localFile); err != nil {
			return fmt.Errorf("failed to write conflict file: %w", err)
		}
	}

	// Create a README file explaining the conflict
	readmePath := targetPath + ".README"
	readmeContent := fmt.Sprintf(`CONFLICT RESOLUTION NEEDED: %s

This file has changes in both upstream and local versions that cannot be automatically merged.

Files created for manual resolution:
- %s.conflict  - Your local version
- %s.upstream  - New upstream version  
- %s          - Current version (same as .conflict)

To resolve:
1. Compare the .conflict and .upstream versions
2. Edit %s with your desired merged content
3. Delete the .conflict, .upstream, and .README files
4. Run 'claude-wm-cli config status' to verify resolution

`, relPath, relPath, relPath, relPath, relPath)

	if err := os.WriteFile(readmePath, []byte(readmeContent), 0644); err != nil {
		return fmt.Errorf("failed to create conflict README: %w", err)
	}

	return nil
}

// updateBaseline copies the current upstream to the baseline directory
func updateBaseline(upstream fs.FS, upstreamRoot, baselineDir string) error {
	// Remove existing baseline
	if err := os.RemoveAll(baselineDir); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove existing baseline: %w", err)
	}

	// Copy upstream to baseline
	if err := fsutil.CopyTreeFS(upstream, upstreamRoot, baselineDir); err != nil {
		return fmt.Errorf("failed to copy upstream to baseline: %w", err)
	}

	return nil
}

