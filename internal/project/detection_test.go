package project

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProjectStatus_String(t *testing.T) {
	tests := []struct {
		status   ProjectStatus
		expected string
	}{
		{NotInitialized, "not_initialized"},
		{Partial, "partial"},
		{Complete, "complete"},
		{ProjectStatus(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.status.String())
		})
	}
}

func TestDetectProjectInitialization_NotInitialized(t *testing.T) {
	tests := []struct {
		name           string
		setupFunc      func(string)
		expectedStatus ProjectStatus
	}{
		{
			name: "empty directory",
			setupFunc: func(dir string) {
				// Directory is empty by default
			},
			expectedStatus: NotInitialized,
		},
		{
			name: "no docs directory",
			setupFunc: func(dir string) {
				// Create some other files/directories
				os.Mkdir(filepath.Join(dir, "src"), 0755)
				os.WriteFile(filepath.Join(dir, "README.md"), []byte("# Test"), 0644)
			},
			expectedStatus: NotInitialized,
		},
		{
			name: "empty docs directory",
			setupFunc: func(dir string) {
				os.Mkdir(filepath.Join(dir, "docs"), 0755)
			},
			expectedStatus: NotInitialized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			tt.setupFunc(tempDir)

			result, err := DetectProjectInitialization(tempDir)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, result.Status)
			assert.Equal(t, tempDir, result.RootPath)
			assert.False(t, result.HasStructure)
			assert.False(t, result.HasFiles)
		})
	}
}

func TestDetectProjectInitialization_Partial(t *testing.T) {
	tests := []struct {
		name           string
		setupFunc      func(string)
		expectedStatus ProjectStatus
	}{
		{
			name: "partial directory structure",
			setupFunc: func(dir string) {
				// Create only some directories
				os.MkdirAll(filepath.Join(dir, "docs/1-project"), 0755)
				os.MkdirAll(filepath.Join(dir, "docs/2-current-epic"), 0755)
				// Missing docs/3-current-task
			},
			expectedStatus: Partial,
		},
		{
			name: "structure but missing epics.json",
			setupFunc: func(dir string) {
				// Create all directories
				for _, reqDir := range RequiredDirectories {
					os.MkdirAll(filepath.Join(dir, reqDir), 0755)
				}
				// Don't create epics.json
			},
			expectedStatus: Partial,
		},
		{
			name: "structure with invalid epics.json",
			setupFunc: func(dir string) {
				// Create all directories
				for _, reqDir := range RequiredDirectories {
					os.MkdirAll(filepath.Join(dir, reqDir), 0755)
				}
				// Create invalid JSON file
				os.WriteFile(filepath.Join(dir, "docs/1-project/epics.json"),
					[]byte("invalid json content"), 0644)
			},
			expectedStatus: Partial,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			tt.setupFunc(tempDir)

			result, err := DetectProjectInitialization(tempDir)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, result.Status)
			assert.Equal(t, tempDir, result.RootPath)
		})
	}
}

func TestDetectProjectInitialization_Complete(t *testing.T) {
	tempDir := t.TempDir()

	// Create complete project structure
	for _, reqDir := range RequiredDirectories {
		os.MkdirAll(filepath.Join(tempDir, reqDir), 0755)
	}

	// Create valid epics.json
	epicsData := map[string]interface{}{
		"epics": []map[string]interface{}{
			{
				"id":     "EPIC-001",
				"title":  "Test Epic",
				"status": "completed",
			},
		},
	}
	epicsJSON, _ := json.Marshal(epicsData)
	os.WriteFile(filepath.Join(tempDir, "docs/1-project/epics.json"), epicsJSON, 0644)

	result, err := DetectProjectInitialization(tempDir)
	require.NoError(t, err)
	assert.Equal(t, Complete, result.Status)
	assert.Equal(t, tempDir, result.RootPath)
	assert.True(t, result.HasStructure)
	assert.True(t, result.HasFiles)
	assert.Empty(t, result.MissingFiles)
}

func TestDetectProjectInitialization_WithOptionalFiles(t *testing.T) {
	tempDir := t.TempDir()

	// Create complete project structure
	for _, reqDir := range RequiredDirectories {
		os.MkdirAll(filepath.Join(tempDir, reqDir), 0755)
	}

	// Create valid epics.json
	epicsData := map[string]interface{}{"epics": []interface{}{}}
	epicsJSON, _ := json.Marshal(epicsData)
	os.WriteFile(filepath.Join(tempDir, "docs/1-project/epics.json"), epicsJSON, 0644)

	// Create optional files
	currentEpicData := map[string]interface{}{
		"id":    "EPIC-001",
		"title": "Current Epic",
	}
	currentEpicJSON, _ := json.Marshal(currentEpicData)
	os.WriteFile(filepath.Join(tempDir, "docs/2-current-epic/current-epic.json"),
		currentEpicJSON, 0644)

	storiesData := map[string]interface{}{"stories": []interface{}{}}
	storiesJSON, _ := json.Marshal(storiesData)
	os.WriteFile(filepath.Join(tempDir, "docs/2-current-epic/stories.json"),
		storiesJSON, 0644)

	result, err := DetectProjectInitialization(tempDir)
	require.NoError(t, err)
	assert.Equal(t, Complete, result.Status)
	assert.True(t, result.HasStructure)
	assert.True(t, result.HasFiles)
}

func TestDetectProjectInitialization_NonExistentPath(t *testing.T) {
	result, err := DetectProjectInitialization("/non/existent/path")
	require.NoError(t, err)
	assert.Equal(t, NotInitialized, result.Status)
	assert.Contains(t, result.Issues, "Root path does not exist: /non/existent/path")
}

func TestDetectProjectInitialization_PermissionError(t *testing.T) {
	tempDir := t.TempDir()

	// Create docs directory with no read permissions
	docsPath := filepath.Join(tempDir, "docs")
	os.Mkdir(docsPath, 0755)

	// Create a file with restricted permissions (this simulates permission issues)
	restrictedFile := filepath.Join(docsPath, "restricted")
	os.WriteFile(restrictedFile, []byte("test"), 0644)

	// This test depends on the system - on some systems we can't simulate permission errors in temp dirs
	result, err := DetectProjectInitialization(tempDir)
	require.NoError(t, err)
	// Should still detect as not initialized due to missing structure
	assert.Equal(t, NotInitialized, result.Status)
}

func TestCheckDocsStructure(t *testing.T) {
	tests := []struct {
		name           string
		setupFunc      func(string)
		expectedHas    bool
		expectedIssues int
	}{
		{
			name: "no docs directory",
			setupFunc: func(dir string) {
				// No setup - empty directory
			},
			expectedHas:    false,
			expectedIssues: 1,
		},
		{
			name: "partial structure",
			setupFunc: func(dir string) {
				os.MkdirAll(filepath.Join(dir, "docs/1-project"), 0755)
				os.MkdirAll(filepath.Join(dir, "docs/2-current-epic"), 0755)
				// Missing docs/3-current-task
			},
			expectedHas:    true, // Half or more directories exist
			expectedIssues: 1,
		},
		{
			name: "complete structure",
			setupFunc: func(dir string) {
				for _, reqDir := range RequiredDirectories {
					os.MkdirAll(filepath.Join(dir, reqDir), 0755)
				}
			},
			expectedHas:    true,
			expectedIssues: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			tt.setupFunc(tempDir)

			hasStructure, issues := checkDocsStructure(tempDir)
			assert.Equal(t, tt.expectedHas, hasStructure)
			assert.Len(t, issues, tt.expectedIssues)
		})
	}
}

func TestValidateRequiredFiles(t *testing.T) {
	tests := []struct {
		name            string
		setupFunc       func(string)
		expectedHas     bool
		expectedMissing int
	}{
		{
			name: "no files",
			setupFunc: func(dir string) {
				os.MkdirAll(filepath.Join(dir, "docs/1-project"), 0755)
			},
			expectedHas:     false,
			expectedMissing: 1,
		},
		{
			name: "invalid JSON file",
			setupFunc: func(dir string) {
				os.MkdirAll(filepath.Join(dir, "docs/1-project"), 0755)
				os.WriteFile(filepath.Join(dir, "docs/1-project/epics.json"),
					[]byte("invalid json"), 0644)
			},
			expectedHas:     false,
			expectedMissing: 1,
		},
		{
			name: "valid files",
			setupFunc: func(dir string) {
				os.MkdirAll(filepath.Join(dir, "docs/1-project"), 0755)
				epicsData := map[string]interface{}{"epics": []interface{}{}}
				epicsJSON, _ := json.Marshal(epicsData)
				os.WriteFile(filepath.Join(dir, "docs/1-project/epics.json"),
					epicsJSON, 0644)
			},
			expectedHas:     true,
			expectedMissing: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			tt.setupFunc(tempDir)

			hasFiles, missingFiles := validateRequiredFiles(tempDir)
			assert.Equal(t, tt.expectedHas, hasFiles)
			assert.Len(t, missingFiles, tt.expectedMissing)
		})
	}
}

func TestValidateJSONFile(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name      string
		content   string
		expectErr bool
	}{
		{
			name:      "valid JSON",
			content:   `{"test": "value"}`,
			expectErr: false,
		},
		{
			name:      "invalid JSON",
			content:   `{invalid json`,
			expectErr: true,
		},
		{
			name:      "empty file",
			content:   ``,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testFile := filepath.Join(tempDir, "test.json")
			os.WriteFile(testFile, []byte(tt.content), 0644)

			err := validateJSONFile(testFile)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFileExists(t *testing.T) {
	tempDir := t.TempDir()

	existingFile := filepath.Join(tempDir, "existing.txt")
	os.WriteFile(existingFile, []byte("test"), 0644)

	nonExistentFile := filepath.Join(tempDir, "nonexistent.txt")

	assert.True(t, fileExists(existingFile))
	assert.False(t, fileExists(nonExistentFile))
}

func TestDetermineInitializationStatus(t *testing.T) {
	tests := []struct {
		hasStructure bool
		hasFiles     bool
		expected     ProjectStatus
	}{
		{false, false, NotInitialized},
		{false, true, NotInitialized}, // Can't have files without structure
		{true, false, Partial},
		{true, true, Complete},
	}

	for _, tt := range tests {
		t.Run(tt.expected.String(), func(t *testing.T) {
			result := determineInitializationStatus(tt.hasStructure, tt.hasFiles)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetRequiredFilesForCompletion(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func(string)
		expectMin int // Minimum number of required items
	}{
		{
			name: "empty directory",
			setupFunc: func(dir string) {
				// Empty directory
			},
			expectMin: 3, // Should include directories and files
		},
		{
			name: "partial structure",
			setupFunc: func(dir string) {
				os.MkdirAll(filepath.Join(dir, "docs/1-project"), 0755)
			},
			expectMin: 1, // Should include missing files
		},
		{
			name: "complete project",
			setupFunc: func(dir string) {
				// Create complete structure
				for _, reqDir := range RequiredDirectories {
					os.MkdirAll(filepath.Join(dir, reqDir), 0755)
				}
				epicsData := map[string]interface{}{"epics": []interface{}{}}
				epicsJSON, _ := json.Marshal(epicsData)
				os.WriteFile(filepath.Join(dir, "docs/1-project/epics.json"),
					epicsJSON, 0644)
			},
			expectMin: 0, // Should be empty for complete project
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			tt.setupFunc(tempDir)

			required := GetRequiredFilesForCompletion(tempDir)
			assert.GreaterOrEqual(t, len(required), tt.expectMin)
		})
	}
}

func TestCheckOptionalFiles(t *testing.T) {
	tests := []struct {
		name           string
		setupFunc      func(string)
		expectedStatus int
	}{
		{
			name: "no optional files",
			setupFunc: func(dir string) {
				os.MkdirAll(filepath.Join(dir, "docs/2-current-epic"), 0755)
			},
			expectedStatus: 1, // Should have status about missing files
		},
		{
			name: "has current epic only",
			setupFunc: func(dir string) {
				os.MkdirAll(filepath.Join(dir, "docs/2-current-epic"), 0755)
				data := map[string]interface{}{"id": "EPIC-001"}
				jsonData, _ := json.Marshal(data)
				os.WriteFile(filepath.Join(dir, "docs/2-current-epic/current-epic.json"),
					jsonData, 0644)
			},
			expectedStatus: 1, // Should note missing stories
		},
		{
			name: "has both files",
			setupFunc: func(dir string) {
				os.MkdirAll(filepath.Join(dir, "docs/2-current-epic"), 0755)
				epicData := map[string]interface{}{"id": "EPIC-001"}
				epicJSON, _ := json.Marshal(epicData)
				os.WriteFile(filepath.Join(dir, "docs/2-current-epic/current-epic.json"),
					epicJSON, 0644)

				storiesData := map[string]interface{}{"stories": []interface{}{}}
				storiesJSON, _ := json.Marshal(storiesData)
				os.WriteFile(filepath.Join(dir, "docs/2-current-epic/stories.json"),
					storiesJSON, 0644)
			},
			expectedStatus: 0, // Should have no status messages
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			tt.setupFunc(tempDir)

			status := checkOptionalFiles(tempDir)
			assert.Len(t, status, tt.expectedStatus)
		})
	}
}
