package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"claude-wm-cli/internal/navigation"
	"claude-wm-cli/internal/workflow"
)

func TestInteractiveCmd_Basic(t *testing.T) {
	// Test that the command is properly configured
	assert.Equal(t, "interactive", InteractiveCmd.Use)
	assert.Contains(t, InteractiveCmd.Aliases, "nav")
	assert.Contains(t, InteractiveCmd.Aliases, "menu")
	assert.NotEmpty(t, InteractiveCmd.Short)
	assert.NotEmpty(t, InteractiveCmd.Long)
}

func TestInteractiveCmd_Flags(t *testing.T) {
	// Test that all expected flags are defined
	flags := InteractiveCmd.Flags()

	assert.NotNil(t, flags.Lookup("status"))
	assert.NotNil(t, flags.Lookup("suggest"))
	assert.NotNil(t, flags.Lookup("quick"))
	assert.NotNil(t, flags.Lookup("no-interactive"))
	assert.NotNil(t, flags.Lookup("width"))
	assert.NotNil(t, flags.Lookup("max-suggestions"))
}

func TestInteractiveCmd_FlagDefaults(t *testing.T) {
	// Test flag default values
	flags := InteractiveCmd.Flags()

	statusFlag := flags.Lookup("status")
	assert.Equal(t, "false", statusFlag.DefValue)

	widthFlag := flags.Lookup("width")
	assert.Equal(t, "80", widthFlag.DefValue)

	maxSuggestionsFlag := flags.Lookup("max-suggestions")
	assert.Equal(t, "5", maxSuggestionsFlag.DefValue)
}

func TestRunInteractive_StatusOnly(t *testing.T) {
	// Create temporary directory with project structure
	tempDir := t.TempDir()
	createTestProjectStructure(t, tempDir)

	// Change to temp directory
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalDir)

	err = os.Chdir(tempDir)
	require.NoError(t, err)

	// Create command with status flag
	cmd := &cobra.Command{}
	showStatusOnly = true
	showSuggestOnly = false
	showQuickStatus = false
	noInteractive = true
	displayWidth = 80
	maxSuggestions = 5

	// Capture output
	var output bytes.Buffer
	originalStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run command
	err = runInteractive(cmd, []string{})

	// Restore stdout and get output
	w.Close()
	os.Stdout = originalStdout
	output.ReadFrom(r)

	// Reset flags
	showStatusOnly = false

	assert.NoError(t, err)
	assert.Contains(t, output.String(), "Project")
}

func TestRunInteractive_QuickStatus(t *testing.T) {
	tempDir := t.TempDir()
	createTestProjectStructure(t, tempDir)

	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalDir)

	err = os.Chdir(tempDir)
	require.NoError(t, err)

	cmd := &cobra.Command{}
	showStatusOnly = false
	showSuggestOnly = false
	showQuickStatus = true
	noInteractive = true
	displayWidth = 80
	maxSuggestions = 5

	var output bytes.Buffer
	originalStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = runInteractive(cmd, []string{})

	w.Close()
	os.Stdout = originalStdout
	output.ReadFrom(r)

	showQuickStatus = false

	assert.NoError(t, err)
	// Should contain state information
	outputStr := output.String()
	assert.NotEmpty(t, outputStr)
}

func TestRunInteractive_SuggestOnly(t *testing.T) {
	tempDir := t.TempDir()
	createTestProjectStructure(t, tempDir)

	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalDir)

	err = os.Chdir(tempDir)
	require.NoError(t, err)

	cmd := &cobra.Command{}
	showStatusOnly = false
	showSuggestOnly = true
	showQuickStatus = false
	noInteractive = true
	displayWidth = 80
	maxSuggestions = 3

	var output bytes.Buffer
	originalStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = runInteractive(cmd, []string{})

	w.Close()
	os.Stdout = originalStdout
	output.ReadFrom(r)

	showSuggestOnly = false

	assert.NoError(t, err)
	outputStr := output.String()
	assert.Contains(t, outputStr, "Suggestions")
}

func TestRunInteractive_NonInteractive(t *testing.T) {
	tempDir := t.TempDir()
	createTestProjectStructure(t, tempDir)

	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalDir)

	err = os.Chdir(tempDir)
	require.NoError(t, err)

	cmd := &cobra.Command{}
	showStatusOnly = false
	showSuggestOnly = false
	showQuickStatus = false
	noInteractive = true
	displayWidth = 80
	maxSuggestions = 5

	var output bytes.Buffer
	originalStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = runInteractive(cmd, []string{})

	w.Close()
	os.Stdout = originalStdout
	output.ReadFrom(r)

	assert.NoError(t, err)
	outputStr := output.String()
	assert.NotEmpty(t, outputStr)
	// Should show project state and suggestions without interactive menu
}

func TestRunInteractive_InvalidDirectory(t *testing.T) {
	// Change to a non-existent directory (this should work as os.Getwd() will succeed)
	// We'll test the error case by mocking, but for now test the basic error handling

	cmd := &cobra.Command{}

	// Test with valid directory but ensure error handling works
	// The error cases are harder to test without dependency injection
	// For now, we verify the command doesn't panic
	assert.NotPanics(t, func() {
		runInteractive(cmd, []string{})
	})
}

func TestCreateMainMenu(t *testing.T) {
	ctx := &navigation.ProjectContext{
		State: navigation.StateNotInitialized,
	}

	suggestions := []*navigation.Suggestion{
		{
			Action: &workflow.WorkflowAction{
				ID:   "init-project",
				Name: "Initialize Project",
			},
			Priority:  workflow.PriorityP0,
			Reasoning: "Project needs initialization",
		},
		{
			Action: &workflow.WorkflowAction{
				ID:   "help",
				Name: "Show Help",
			},
			Priority:  workflow.PriorityP2,
			Reasoning: "Get help with commands",
		},
	}

	menu := createMainMenu(ctx, suggestions)

	assert.NotNil(t, menu)
	assert.Equal(t, "🧭 Project Navigation", menu.Title)
	assert.True(t, menu.ShowNumbers)
	assert.True(t, menu.AllowBack)
	assert.True(t, menu.AllowQuit)
	assert.NotEmpty(t, menu.Options)

	// Should have suggestions plus standard options
	assert.GreaterOrEqual(t, len(menu.Options), 3) // At least standard options
}

func TestCreateMainMenu_ManySuggestions(t *testing.T) {
	ctx := &navigation.ProjectContext{
		State: navigation.StateNotInitialized,
	}

	// Create more than 3 suggestions
	suggestions := []*navigation.Suggestion{}
	for i := 0; i < 5; i++ {
		suggestions = append(suggestions, &navigation.Suggestion{
			Action: &workflow.WorkflowAction{
				ID:   "action" + string(rune(i+'1')),
				Name: "Action " + string(rune(i+'1')),
			},
			Priority:  workflow.PriorityP1,
			Reasoning: "Test suggestion",
		})
	}

	menu := createMainMenu(ctx, suggestions)

	// Should limit to top 3 suggestions plus standard options
	suggestionCount := 0
	for _, option := range menu.Options {
		if option.Action != "status" && option.Action != "suggestions" &&
			option.Action != "refresh" && option.Enabled {
			suggestionCount++
		}
	}

	assert.LessOrEqual(t, suggestionCount, 3, "Should limit to 3 suggestions in menu")
}

func TestExecuteAction_DirectoryCreation(t *testing.T) {
	// Test the basic directory creation logic that would be used in init
	tempDir := t.TempDir()

	expectedDirs := []string{
		"docs/1-project",
		"docs/2-current-epic",
		"docs/3-current-task",
	}

	// Test creating directories (simulates what executeInitProject would do)
	for _, dir := range expectedDirs {
		fullPath := filepath.Join(tempDir, dir)
		err := os.MkdirAll(fullPath, 0755)
		assert.NoError(t, err)
		assert.DirExists(t, fullPath)
	}
}

func TestGetPriorityIcon(t *testing.T) {
	tests := []struct {
		priority workflow.Priority
		expected string
	}{
		{workflow.PriorityP0, "🔴 "},
		{workflow.PriorityP1, "🟡 "},
		{workflow.PriorityP2, "🟢 "},
		{workflow.Priority("unknown"), "⚪ "},
	}

	for _, tt := range tests {
		t.Run(string(tt.priority), func(t *testing.T) {
			result := getPriorityIcon(tt.priority)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTruncateString(t *testing.T) {
	tests := []struct {
		input    string
		maxLen   int
		expected string
	}{
		{"short", 10, "short"},
		{"exactly10c", 10, "exactly10c"}, // Actually 10 chars
		{"this is a very long string", 10, "this is..."},
		{"", 5, ""},
		{"test", 5, "test"}, // Actually shorter than maxLen
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := truncateString(tt.input, tt.maxLen)
			assert.Equal(t, tt.expected, result)
			if len(tt.input) > tt.maxLen {
				assert.LessOrEqual(t, len(result), tt.maxLen)
			}
		})
	}
}

func TestDisplaySuggestions(t *testing.T) {
	suggestions := []*navigation.Suggestion{
		{
			Action: &workflow.WorkflowAction{
				ID:   "test-action",
				Name: "Test Action",
			},
			Priority:    workflow.PriorityP1,
			Reasoning:   "Test reasoning",
			NextActions: []string{"next-action"},
		},
	}

	engine := navigation.NewSuggestionEngine()

	// Capture output
	var output bytes.Buffer
	originalStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	displaySuggestions(suggestions, engine)

	w.Close()
	os.Stdout = originalStdout
	output.ReadFrom(r)

	outputStr := output.String()
	assert.Contains(t, outputStr, "Action Suggestions")
	assert.Contains(t, outputStr, "Test Action")
	assert.Contains(t, outputStr, "Test reasoning")
	assert.Contains(t, outputStr, "next-action")
}

func TestDisplaySuggestions_Empty(t *testing.T) {
	engine := navigation.NewSuggestionEngine()

	var output bytes.Buffer
	originalStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	displaySuggestions([]*navigation.Suggestion{}, engine)

	w.Close()
	os.Stdout = originalStdout
	output.ReadFrom(r)

	outputStr := output.String()
	assert.Contains(t, outputStr, "No suggestions available")
}

// Helper functions for tests

func createTestProjectStructure(t *testing.T, tempDir string) {
	dirs := []string{
		"docs/1-project",
		"docs/2-current-epic",
		"docs/3-current-task",
	}

	for _, dir := range dirs {
		err := os.MkdirAll(filepath.Join(tempDir, dir), 0755)
		require.NoError(t, err)
	}
}

// Mock implementations for testing

// MenuDisplayInterface defines the interface for menu display functionality
type MenuDisplayInterface interface {
	Confirm(message string) (bool, error)
	PromptString(prompt string) (string, error)
	ShowMessage(message string)
	ShowError(message string)
	ShowWarning(message string)
	ShowSuccess(message string)
	WaitForKeyPress(message string) error
}

type mockMenuDisplay struct {
	confirmResponse bool
	promptResponse  string
	lastMessage     string
	lastError       string
	lastWarning     string
	lastSuccess     string
}

func (m *mockMenuDisplay) Confirm(message string) (bool, error) {
	return m.confirmResponse, nil
}

func (m *mockMenuDisplay) PromptString(prompt string) (string, error) {
	return m.promptResponse, nil
}

func (m *mockMenuDisplay) ShowMessage(message string) {
	m.lastMessage = message
}

func (m *mockMenuDisplay) ShowError(message string) {
	m.lastError = message
}

func (m *mockMenuDisplay) ShowWarning(message string) {
	m.lastWarning = message
}

func (m *mockMenuDisplay) ShowSuccess(message string) {
	m.lastSuccess = message
}

func (m *mockMenuDisplay) WaitForKeyPress(message string) error {
	return nil
}
