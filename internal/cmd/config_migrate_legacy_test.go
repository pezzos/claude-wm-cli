package cmd

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestAnalyzeFile(t *testing.T) {
	tests := []struct {
		name         string
		relPath      string
		size         int64
		wantType     string
		wantCategory string
		wantDisposition string
	}{
		{
			name:            "system config file",
			relPath:         "system/settings.json.template",
			size:            1024,
			wantType:        "config",
			wantCategory:    "system",
			wantDisposition: "migrate",
		},
		{
			name:            "user config file",
			relPath:         "user/settings.json",
			size:            512,
			wantType:        "config",
			wantCategory:    "user",
			wantDisposition: "migrate",
		},
		{
			name:            "runtime config file",
			relPath:         "runtime/settings.json",
			size:            800,
			wantType:        "config",
			wantCategory:    "runtime",
			wantDisposition: "ignore",
		},
		{
			name:            "meta file",
			relPath:         "meta.json",
			size:            256,
			wantType:        "config",
			wantCategory:    "meta",
			wantDisposition: "convert",
		},
		{
			name:            "hook script",
			relPath:         "system/hooks/post-write.sh",
			size:            2048,
			wantType:        "hook",
			wantCategory:    "system",
			wantDisposition: "migrate",
		},
		{
			name:            "template file",
			relPath:         "system/templates/task.md",
			size:            1536,
			wantType:        "template",
			wantCategory:    "system",
			wantDisposition: "migrate",
		},
		{
			name:            "cache file",
			relPath:         "cache/temp.dat",
			size:            4096,
			wantType:        "cache",
			wantCategory:    "other",
			wantDisposition: "ignore",
		},
		{
			name:            "backup file",
			relPath:         "backup/old-config.bak",
			size:            2048,
			wantType:        "backup",
			wantCategory:    "other",
			wantDisposition: "ignore",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := &testFileInfo{
				name:    filepath.Base(tt.relPath),
				size:    tt.size,
				modTime: time.Now(),
				isDir:   false,
			}

			result := analyzeFile(tt.relPath, info)

			if result.Type != tt.wantType {
				t.Errorf("analyzeFile().Type = %v, want %v", result.Type, tt.wantType)
			}

			if result.Category != tt.wantCategory {
				t.Errorf("analyzeFile().Category = %v, want %v", result.Category, tt.wantCategory)
			}

			if result.Disposition != tt.wantDisposition {
				t.Errorf("analyzeFile().Disposition = %v, want %v", result.Disposition, tt.wantDisposition)
			}

			if result.Path != tt.relPath {
				t.Errorf("analyzeFile().Path = %v, want %v", result.Path, tt.relPath)
			}

			if result.Size != tt.size {
				t.Errorf("analyzeFile().Size = %v, want %v", result.Size, tt.size)
			}
		})
	}
}

func TestMapLegacyPathToNew(t *testing.T) {
	tests := []struct {
		name     string
		relPath  string
		analysis FileAnalysis
		want     string
	}{
		{
			name:    "system file to baseline",
			relPath: "system/commands/template.md",
			analysis: FileAnalysis{Category: "system"},
			want:    "baseline/commands/template.md",
		},
		{
			name:    "user file stays in user structure",
			relPath: "user/custom.json",
			analysis: FileAnalysis{Category: "user"},
			want:    "user/custom.json",
		},
		{
			name:    "meta.json to root",
			relPath: "meta.json",
			analysis: FileAnalysis{Category: "meta"},
			want:    "meta.json",
		},
		{
			name:    "config.json stays as is",
			relPath: "config.json",
			analysis: FileAnalysis{Category: "meta"},
			want:    "config.json",
		},
		{
			name:    "other files keep structure",
			relPath: "hooks/custom.sh",
			analysis: FileAnalysis{Category: "other"},
			want:    "hooks/custom.sh",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapLegacyPathToNew(tt.relPath, tt.analysis)
			if result != tt.want {
				t.Errorf("mapLegacyPathToNew() = %v, want %v", result, tt.want)
			}
		})
	}
}

func TestDetermineMigrationAction(t *testing.T) {
	tests := []struct {
		name       string
		relPath    string
		analysis   FileAnalysis
		targetPath string
		wantType   string
		wantTarget string
	}{
		{
			name:    "migrate disposition creates copy action",
			relPath: "system/config.json",
			analysis: FileAnalysis{
				Type:        "config",
				Category:    "system",
				Disposition: "migrate",
			},
			targetPath: "/tmp/test",
			wantType:   "copy",
			wantTarget: "baseline/config.json",
		},
		{
			name:    "convert disposition creates convert action",
			relPath: "meta.json",
			analysis: FileAnalysis{
				Type:        "config",
				Category:    "meta",
				Disposition: "convert",
			},
			targetPath: "/tmp/test",
			wantType:   "convert",
			wantTarget: "meta.json",
		},
		{
			name:    "ignore disposition creates ignore action",
			relPath: "runtime/temp.json",
			analysis: FileAnalysis{
				Type:        "config",
				Category:    "runtime",
				Disposition: "ignore",
				Reason:      "Runtime files will be regenerated",
			},
			targetPath: "/tmp/test",
			wantType:   "ignore",
			wantTarget: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := determineMigrationAction(tt.relPath, tt.analysis, tt.targetPath)

			if result.Type != tt.wantType {
				t.Errorf("determineMigrationAction().Type = %v, want %v", result.Type, tt.wantType)
			}

			if result.TargetPath != tt.wantTarget {
				t.Errorf("determineMigrationAction().TargetPath = %v, want %v", result.TargetPath, tt.wantTarget)
			}

			if result.SourcePath != tt.relPath {
				t.Errorf("determineMigrationAction().SourcePath = %v, want %v", result.SourcePath, tt.relPath)
			}

			if result.Status != "planned" {
				t.Errorf("determineMigrationAction().Status = %v, want %v", result.Status, "planned")
			}
		})
	}
}

func TestValidateTargetDirectory(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name      string
		targetDir string
		setup     func() error
		wantError bool
	}{
		{
			name:      "non-existent directory is valid",
			targetDir: filepath.Join(tempDir, "new-dir"),
			setup:     func() error { return nil },
			wantError: false,
		},
		{
			name:      "empty directory is valid",
			targetDir: filepath.Join(tempDir, "empty-dir"),
			setup: func() error {
				return os.MkdirAll(filepath.Join(tempDir, "empty-dir"), 0755)
			},
			wantError: false,
		},
		{
			name:      "non-empty directory is invalid",
			targetDir: filepath.Join(tempDir, "non-empty-dir"),
			setup: func() error {
				dir := filepath.Join(tempDir, "non-empty-dir")
				if err := os.MkdirAll(dir, 0755); err != nil {
					return err
				}
				return os.WriteFile(filepath.Join(dir, "existing.txt"), []byte("content"), 0644)
			},
			wantError: true,
		},
		{
			name:      "file instead of directory is invalid",
			targetDir: filepath.Join(tempDir, "file-not-dir"),
			setup: func() error {
				return os.WriteFile(filepath.Join(tempDir, "file-not-dir"), []byte("content"), 0644)
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.setup(); err != nil {
				t.Fatalf("Setup failed: %v", err)
			}

			err := validateTargetDirectory(tt.targetDir)
			if (err != nil) != tt.wantError {
				t.Errorf("validateTargetDirectory() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestCopyFileWithDir(t *testing.T) {
	tempDir := t.TempDir()

	// Create source file
	srcDir := filepath.Join(tempDir, "source")
	srcFile := filepath.Join(srcDir, "test.txt")
	testContent := "Hello, Migration World!"

	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatalf("Failed to create source directory: %v", err)
	}

	if err := os.WriteFile(srcFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	// Copy to destination (which doesn't exist yet)
	dstFile := filepath.Join(tempDir, "target", "subdir", "test.txt")
	err := fsutil.CopyFileWithDir(srcFile, dstFile)
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

func TestAnalyzeLegacyStructure(t *testing.T) {
	tempDir := t.TempDir()

	// Create a mock legacy structure
	legacyDir := filepath.Join(tempDir, ".claude-wm")
	targetDir := filepath.Join(tempDir, ".wm")

	// Create directory structure and files
	structure := map[string]string{
		"system/settings.json.template": `{"version": "1.0.0"}`,
		"system/commands/task.md":       "# Task Template",
		"system/hooks/post-write.sh":    "#!/bin/bash\necho 'hook'",
		"user/settings.json":            `{"custom": true}`,
		"runtime/settings.json":         `{"generated": true}`,
		"meta.json":                     `{"version": "1.0"}`,
		"cache/temp.dat":                "temporary data",
		"backup/old.bak":                "old backup",
	}

	for relPath, content := range structure {
		fullPath := filepath.Join(legacyDir, relPath)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			t.Fatalf("Failed to create directory for %s: %v", relPath, err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", relPath, err)
		}
	}

	// Analyze the structure
	report, err := analyzeLegacyStructure(legacyDir, targetDir)
	if err != nil {
		t.Fatalf("analyzeLegacyStructure failed: %v", err)
	}

	// Verify basic report structure
	if report.LegacyPath != legacyDir {
		t.Errorf("LegacyPath = %v, want %v", report.LegacyPath, legacyDir)
	}

	if report.TargetPath != targetDir {
		t.Errorf("TargetPath = %v, want %v", report.TargetPath, targetDir)
	}

	if report.Summary.FilesAnalyzed != len(structure) {
		t.Errorf("FilesAnalyzed = %v, want %v", report.Summary.FilesAnalyzed, len(structure))
	}

	// Verify that some files are marked for migration
	if report.Summary.FilesToCopy == 0 {
		t.Error("Expected some files to be marked for copying")
	}

	// Verify that runtime files are ignored
	if report.Summary.FilesToIgnore == 0 {
		t.Error("Expected some files to be marked for ignoring")
	}

	// Check specific file analyses
	expectedAnalyses := map[string]struct {
		category    string
		disposition string
	}{
		"system/settings.json.template": {"system", "migrate"},
		"runtime/settings.json":         {"runtime", "ignore"},
		"meta.json":                     {"meta", "convert"},
		"cache/temp.dat":                {"other", "ignore"},
	}

	for path, expected := range expectedAnalyses {
		analysis, exists := report.AnalyzedFiles[path]
		if !exists {
			t.Errorf("File analysis missing for %s", path)
			continue
		}

		if analysis.Category != expected.category {
			t.Errorf("File %s: Category = %v, want %v", path, analysis.Category, expected.category)
		}

		if analysis.Disposition != expected.disposition {
			t.Errorf("File %s: Disposition = %v, want %v", path, analysis.Disposition, expected.disposition)
		}
	}
}

// Test helper: mock FileInfo implementation
type testFileInfo struct {
	name    string
	size    int64
	modTime time.Time
	isDir   bool
}

func (tfi *testFileInfo) Name() string       { return tfi.name }
func (tfi *testFileInfo) Size() int64        { return tfi.size }
func (tfi *testFileInfo) Mode() os.FileMode  { return 0644 }
func (tfi *testFileInfo) ModTime() time.Time { return tfi.modTime }
func (tfi *testFileInfo) IsDir() bool        { return tfi.isDir }
func (tfi *testFileInfo) Sys() interface{}   { return nil }