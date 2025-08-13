package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"claude-wm-cli/internal/diff"
)

func TestFilterChangesByPattern(t *testing.T) {
	tests := []struct {
		name     string
		changes  []diff.Change
		patterns []string
		wantLen  int
		wantPaths []string
	}{
		{
			name: "no patterns returns all",
			changes: []diff.Change{
				{Path: "agents/test.md", Type: diff.ChangeNew},
				{Path: "commands/help.md", Type: diff.ChangeMod},
			},
			patterns: []string{},
			wantLen:  2,
			wantPaths: []string{"agents/test.md", "commands/help.md"},
		},
		{
			name: "simple pattern matching",
			changes: []diff.Change{
				{Path: "agents/test.md", Type: diff.ChangeNew},
				{Path: "commands/help.md", Type: diff.ChangeMod},
				{Path: "agents/another.md", Type: diff.ChangeDel},
			},
			patterns: []string{"agents/*"},
			wantLen:  2,
			wantPaths: []string{"agents/test.md", "agents/another.md"},
		},
		{
			name: "double star pattern",
			changes: []diff.Change{
				{Path: "agents/test.md", Type: diff.ChangeNew},
				{Path: "agents/subdir/nested.md", Type: diff.ChangeMod},
				{Path: "commands/help.md", Type: diff.ChangeDel},
			},
			patterns: []string{"agents/**"},
			wantLen:  2,
			wantPaths: []string{"agents/test.md", "agents/subdir/nested.md"},
		},
		{
			name: "multiple patterns",
			changes: []diff.Change{
				{Path: "agents/test.md", Type: diff.ChangeNew},
				{Path: "commands/help.md", Type: diff.ChangeMod},
				{Path: "templates/base.md", Type: diff.ChangeDel},
				{Path: "other/file.txt", Type: diff.ChangeNew},
			},
			patterns: []string{"agents/**", "commands/**"},
			wantLen:  2,
			wantPaths: []string{"agents/test.md", "commands/help.md"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterChangesByPattern(tt.changes, tt.patterns)
			if len(result) != tt.wantLen {
				t.Errorf("filterChangesByPattern() returned %d changes, want %d", len(result), tt.wantLen)
			}

			// Check that expected paths are present
			for _, wantPath := range tt.wantPaths {
				found := false
				for _, change := range result {
					if change.Path == wantPath {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected path %s not found in results", wantPath)
				}
			}
		})
	}
}

func TestMatchesGlobPattern(t *testing.T) {
	tests := []struct {
		path    string
		pattern string
		want    bool
	}{
		{"agents/test.md", "agents/**", true},
		{"agents/subdir/nested.md", "agents/**", true},
		{"agents", "agents/**", true},
		{"other/file.txt", "agents/**", false},
		{"test.md", "**/test.md", true},
		{"subdir/test.md", "**/test.md", true},
		{"other.md", "**/test.md", false},
	}

	for _, tt := range tests {
		t.Run(tt.path+"|"+tt.pattern, func(t *testing.T) {
			if got := matchesGlobPattern(tt.path, tt.pattern); got != tt.want {
				t.Errorf("matchesGlobPattern(%q, %q) = %v, want %v", tt.path, tt.pattern, got, tt.want)
			}
		})
	}
}

func TestCreateDiffResult(t *testing.T) {
	changes := []diff.Change{
		{Path: "new.md", Type: diff.ChangeNew},
		{Path: "modified.md", Type: diff.ChangeMod},
		{Path: "deleted.md", Type: diff.ChangeDel},
	}

	t.Run("with delete allowed", func(t *testing.T) {
		result := createDiffResult(changes, true)
		if result.Summary.New != 1 {
			t.Errorf("Expected 1 new file, got %d", result.Summary.New)
		}
		if result.Summary.Modified != 1 {
			t.Errorf("Expected 1 modified file, got %d", result.Summary.Modified)
		}
		if result.Summary.Deleted != 1 {
			t.Errorf("Expected 1 deleted file, got %d", result.Summary.Deleted)
		}
		if result.Summary.Skipped != 0 {
			t.Errorf("Expected 0 skipped files, got %d", result.Summary.Skipped)
		}
		if len(result.Plan) != 3 {
			t.Errorf("Expected 3 actions, got %d", len(result.Plan))
		}
	})

	t.Run("with delete not allowed", func(t *testing.T) {
		result := createDiffResult(changes, false)
		if result.Summary.New != 1 {
			t.Errorf("Expected 1 new file, got %d", result.Summary.New)
		}
		if result.Summary.Modified != 1 {
			t.Errorf("Expected 1 modified file, got %d", result.Summary.Modified)
		}
		if result.Summary.Deleted != 0 {
			t.Errorf("Expected 0 deleted files, got %d", result.Summary.Deleted)
		}
		if result.Summary.Skipped != 1 {
			t.Errorf("Expected 1 skipped file, got %d", result.Summary.Skipped)
		}
		if len(result.Plan) != 3 {
			t.Errorf("Expected 3 actions, got %d", len(result.Plan))
		}
		// Check that delete action is marked as skipped
		deleteAction := result.Plan[2] // should be the delete action
		if deleteAction.Action != "del" || deleteAction.Status != "skipped" {
			t.Errorf("Expected delete action to be skipped, got action=%s status=%s", deleteAction.Action, deleteAction.Status)
		}
	})
}

func TestCopyFileWithDir(t *testing.T) {
	tempDir := t.TempDir()

	// Create source file
	srcDir := filepath.Join(tempDir, "src")
	err := os.MkdirAll(srcDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create source directory: %v", err)
	}

	srcFile := filepath.Join(srcDir, "test.txt")
	testContent := "Hello, World!"
	err = os.WriteFile(srcFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	// Copy to destination (which doesn't exist yet)
	dstFile := filepath.Join(tempDir, "dst", "subdir", "test.txt")
	err = fsutil.CopyFileWithDir(srcFile, dstFile)
	if err != nil {
		t.Fatalf("copyFileWithDir failed: %v", err)
	}

	// Verify destination file exists and has correct content
	dstContent, err := os.ReadFile(dstFile)
	if err != nil {
		t.Fatalf("Failed to read destination file: %v", err)
	}

	if string(dstContent) != testContent {
		t.Errorf("Content mismatch: got %q, want %q", string(dstContent), testContent)
	}

	// Verify permissions are preserved
	srcInfo, err := os.Stat(srcFile)
	if err != nil {
		t.Fatalf("Failed to stat source file: %v", err)
	}

	dstInfo, err := os.Stat(dstFile)
	if err != nil {
		t.Fatalf("Failed to stat destination file: %v", err)
	}

	if srcInfo.Mode() != dstInfo.Mode() {
		t.Errorf("Permission mismatch: src=%v, dst=%v", srcInfo.Mode(), dstInfo.Mode())
	}
}