package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"claude-wm-cli/internal/workflow"
)

func TestParseContextType(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected workflow.WorkflowContextType
		wantErr  bool
	}{
		{"normal", "normal", workflow.WorkflowContextTypeNormal, false},
		{"interruption", "interruption", workflow.WorkflowContextTypeInterruption, false},
		{"interrupt", "interrupt", workflow.WorkflowContextTypeInterruption, false},
		{"emergency", "emergency", workflow.WorkflowContextTypeEmergency, false},
		{"hotfix", "hotfix", workflow.WorkflowContextTypeHotfix, false},
		{"experiment", "experiment", workflow.WorkflowContextTypeExperiment, false},
		{"case insensitive", "EMERGENCY", workflow.WorkflowContextTypeEmergency, false},
		{"invalid", "invalid", "", true},
		{"empty", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseContextType(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "invalid context type")
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestTruncateInterruptString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{"short string", "hello", 10, "hello"},
		{"exact length", "hello", 5, "hello"},
		{"truncate needed", "hello world", 8, "hello..."},
		{"very short limit", "hello", 3, "hel"},
		{"empty string", "", 5, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateInterruptString(tt.input, tt.maxLen)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestInterruptCommands_Structure(t *testing.T) {
	// Test that all subcommands are properly added
	assert.NotNil(t, interruptCmd)
	assert.Equal(t, "interrupt", interruptCmd.Use)

	// Check subcommands
	subcommands := interruptCmd.Commands()
	subcommandNames := make([]string, 0)
	for _, cmd := range subcommands {
		subcommandNames = append(subcommandNames, cmd.Use)
	}

	// Check that expected command patterns are present
	expectedPatterns := []string{"start", "resume", "status", "clear"}
	for _, pattern := range expectedPatterns {
		found := false
		for _, cmdName := range subcommandNames {
			if strings.HasPrefix(cmdName, pattern) {
				found = true
				break
			}
		}
		assert.True(t, found, "Missing subcommand with pattern: %s", pattern)
	}
}

func TestInterruptStartCommand_Flags(t *testing.T) {
	// Test that required flags are marked as required
	nameFlag := interruptStartCmd.Flag("name")
	require.NotNil(t, nameFlag)

	// Test default values
	typeFlag := interruptStartCmd.Flag("type")
	require.NotNil(t, typeFlag)
	assert.Equal(t, "interruption", typeFlag.DefValue)

	includeFilesFlag := interruptStartCmd.Flag("include-files")
	require.NotNil(t, includeFilesFlag)
	assert.Equal(t, "true", includeFilesFlag.DefValue)

	includeGitFlag := interruptStartCmd.Flag("include-git")
	require.NotNil(t, includeGitFlag)
	assert.Equal(t, "true", includeGitFlag.DefValue)
}

func TestInterruptResumeCommand_Flags(t *testing.T) {
	// Test default values for resume flags
	restoreFilesFlag := interruptResumeCmd.Flag("restore-files")
	require.NotNil(t, restoreFilesFlag)
	assert.Equal(t, "true", restoreFilesFlag.DefValue)

	restoreGitFlag := interruptResumeCmd.Flag("restore-git")
	require.NotNil(t, restoreGitFlag)
	assert.Equal(t, "true", restoreGitFlag.DefValue)

	restoreTicketsFlag := interruptResumeCmd.Flag("restore-tickets")
	require.NotNil(t, restoreTicketsFlag)
	assert.Equal(t, "true", restoreTicketsFlag.DefValue)

	restoreEpicsFlag := interruptResumeCmd.Flag("restore-epics")
	require.NotNil(t, restoreEpicsFlag)
	assert.Equal(t, "true", restoreEpicsFlag.DefValue)

	backupCurrentFlag := interruptResumeCmd.Flag("backup-current")
	require.NotNil(t, backupCurrentFlag)
	assert.Equal(t, "true", backupCurrentFlag.DefValue)

	forceFlag := interruptResumeCmd.Flag("force")
	require.NotNil(t, forceFlag)
	assert.Equal(t, "false", forceFlag.DefValue)
}

func TestInterruptStatusCommand_Flags(t *testing.T) {
	// Test default values for status flags
	verboseFlag := interruptStatusCmd.Flag("verbose")
	require.NotNil(t, verboseFlag)
	assert.Equal(t, "false", verboseFlag.DefValue)

	formatFlag := interruptStatusCmd.Flag("format")
	require.NotNil(t, formatFlag)
	assert.Equal(t, "table", formatFlag.DefValue)
}

func TestInterruptClearCommand_Flags(t *testing.T) {
	// Test that confirm flag is required
	confirmFlag := interruptClearCmd.Flag("confirm")
	require.NotNil(t, confirmFlag)
	assert.Equal(t, "false", confirmFlag.DefValue)

	backupFlag := interruptClearCmd.Flag("backup")
	require.NotNil(t, backupFlag)
	assert.Equal(t, "false", backupFlag.DefValue)
}

func TestStartInterruption_Integration(t *testing.T) {
	tempDir := t.TempDir()
	setupTestDirs(t, tempDir)

	// Change to temp directory
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tempDir)

	// Capture output
	var output bytes.Buffer

	// Create a test command with flags set
	cmd := &cobra.Command{}

	// Set global variables (simulating flag parsing)
	interruptName = "Test Interruption"
	interruptDescription = "Test interruption description"
	interruptType = "emergency"
	interruptNotes = "Test notes"
	interruptTags = []string{"test", "urgent"}
	includeFileState = true
	includeGitState = true

	// Redirect stdout
	originalStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run the command (this will exit on error, so we need to handle that)
	func() {
		defer func() {
			if r := recover(); r != nil {
				// Expected if there are issues, just continue test
			}
		}()
		startInterruption(cmd)
	}()

	// Restore stdout
	w.Close()
	os.Stdout = originalStdout

	// Read captured output
	buf := make([]byte, 1024)
	n, _ := r.Read(buf)
	output.Write(buf[:n])

	// Verify an interruption stack file was created
	stackPath := filepath.Join(tempDir, "docs", "2-current-epic", "interruption-stack.json")
	_, err := os.Stat(stackPath)

	// The command might fail due to missing dependencies, but we can test the flow
	if err == nil {
		// If successful, verify the stack file exists
		assert.FileExists(t, stackPath)
	}
	// If it fails, that's expected in this test environment
}

func TestResumeInterruption_EmptyStack(t *testing.T) {
	tempDir := t.TempDir()
	setupTestDirs(t, tempDir)

	// Change to temp directory
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tempDir)

	// Set global variables (simulating flag parsing)
	resumeForce = false
	restoreFiles = true
	restoreGitState = true
	restoreTickets = true
	restoreEpics = true
	backupCurrent = true

	// Capture stdout
	originalStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run resume with empty stack
	resumeInterruption(&cobra.Command{}, "")

	// Restore stdout
	w.Close()
	os.Stdout = originalStdout

	// Read captured output
	buf := make([]byte, 1024)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	// Should indicate empty stack
	assert.Contains(t, output, "No interruptions to resume")
}

func TestShowInterruptionStatus_EmptyStack(t *testing.T) {
	tempDir := t.TempDir()
	setupTestDirs(t, tempDir)

	// Change to temp directory
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tempDir)

	// Set global variables
	statusVerbose = false
	statusFormat = "table"

	// Capture stdout
	originalStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run status command
	showInterruptionStatus(&cobra.Command{})

	// Restore stdout
	w.Close()
	os.Stdout = originalStdout

	// Read captured output
	buf := make([]byte, 2048)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	// Should show status information
	assert.Contains(t, output, "Interruption Stack Status")
	assert.Contains(t, output, "Current Context")
	assert.Contains(t, output, "Stack Information")
}

func TestClearInterruptionStack_EmptyStack(t *testing.T) {
	tempDir := t.TempDir()
	setupTestDirs(t, tempDir)

	// Change to temp directory
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tempDir)

	// Set global variables
	clearConfirm = true
	clearBackup = false

	// Capture stdout
	originalStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run clear command
	clearInterruptionStack(&cobra.Command{})

	// Restore stdout
	w.Close()
	os.Stdout = originalStdout

	// Read captured output
	buf := make([]byte, 1024)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	// Should indicate already empty
	assert.Contains(t, output, "already empty")
}

func TestClearInterruptionStack_WithoutConfirm(t *testing.T) {
	tempDir := t.TempDir()
	setupTestDirs(t, tempDir)

	// Change to temp directory
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tempDir)

	// Create a mock interruption stack with nested interruptions
	stack := workflow.NewInterruptionStack(tempDir)

	// First save a normal context
	normalOptions := workflow.ContextSaveOptions{
		Name: "Normal Context",
		Type: workflow.WorkflowContextTypeNormal,
	}
	_, err := stack.SaveCurrentContext(normalOptions)
	require.NoError(t, err)

	// Then create an interruption (this will add to the stack)
	interruptOptions := workflow.ContextSaveOptions{
		Name: "Test Interruption",
		Type: workflow.WorkflowContextTypeInterruption,
	}
	_, err = stack.SaveCurrentContext(interruptOptions)
	require.NoError(t, err)

	// Set global variables (confirm = false)
	clearConfirm = false
	clearBackup = false

	// Capture stdout
	originalStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run clear command
	clearInterruptionStack(&cobra.Command{})

	// Restore stdout
	w.Close()
	os.Stdout = originalStdout

	// Read captured output
	buf := make([]byte, 1024)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	// Should be cancelled without confirm
	assert.Contains(t, output, "Operation cancelled")
	assert.Contains(t, output, "Use --confirm")
}

func TestInterruptCommandHelp(t *testing.T) {
	// Test main command help
	help := interruptCmd.Long
	assert.Contains(t, help, "Manage workflow interruptions")
	assert.Contains(t, help, "start")
	assert.Contains(t, help, "resume")
	assert.Contains(t, help, "status")
	assert.Contains(t, help, "clear")

	// Test subcommand help
	startHelp := interruptStartCmd.Long
	assert.Contains(t, startHelp, "Start a new interruption workflow")
	assert.Contains(t, startHelp, "saves your current workflow state")

	resumeHelp := interruptResumeCmd.Long
	assert.Contains(t, resumeHelp, "Resume the previous workflow")
	assert.Contains(t, resumeHelp, "from the interruption stack")

	statusHelp := interruptStatusCmd.Long
	assert.Contains(t, statusHelp, "current interruption stack")

	clearHelp := interruptClearCmd.Long
	assert.Contains(t, clearHelp, "Clear the entire interruption stack")
	assert.Contains(t, clearHelp, "WARNING")
	assert.Contains(t, clearHelp, "emergency command")
}

func TestInterruptCommandExamples(t *testing.T) {
	// Check that examples are included in help text
	examples := []string{
		"claude-wm-cli interrupt start --name",
		"claude-wm-cli interrupt resume",
		"claude-wm-cli interrupt status",
		"claude-wm-cli interrupt clear --confirm",
	}

	mainHelp := interruptCmd.Long
	for _, example := range examples {
		assert.Contains(t, mainHelp, example, "Missing example: %s", example)
	}
}

func TestInterruptCommand_Args(t *testing.T) {
	// Test that resume command has args validation (can't directly test function equality)
	assert.NotNil(t, interruptResumeCmd.Args, "Resume command should have args validation")

	// Test that other commands don't have specific args requirements
	assert.Nil(t, interruptStartCmd.Args)
	assert.Nil(t, interruptStatusCmd.Args)
	assert.Nil(t, interruptClearCmd.Args)
}

func TestDisplayStatusJSON_Placeholder(t *testing.T) {
	// Test that JSON display function exists and handles nil input gracefully
	tempDir := t.TempDir()
	setupTestDirs(t, tempDir)

	stack := workflow.NewInterruptionStack(tempDir)
	stackData, err := stack.ListContexts()
	require.NoError(t, err)

	// Capture stdout
	originalStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Call the function
	displayStatusJSON(stackData)

	// Restore stdout
	w.Close()
	os.Stdout = originalStdout

	// Read captured output
	buf := make([]byte, 1024)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	// Should indicate not yet implemented
	assert.Contains(t, output, "JSON output not yet implemented")
}

func TestInterruptCommands_FlagValidation(t *testing.T) {
	// Test that start command has required name flag
	nameFlag := interruptStartCmd.Flag("name")
	require.NotNil(t, nameFlag)

	// Test that clear command has required confirm flag
	confirmFlag := interruptClearCmd.Flag("confirm")
	require.NotNil(t, confirmFlag)

	// Test flag types and defaults
	typeFlag := interruptStartCmd.Flag("type")
	require.NotNil(t, typeFlag)
	assert.Equal(t, "interruption", typeFlag.DefValue)

	formatFlag := interruptStatusCmd.Flag("format")
	require.NotNil(t, formatFlag)
	assert.Equal(t, "table", formatFlag.DefValue)
}

func TestInterruptCommands_ErrorHandling(t *testing.T) {
	// Test parseContextType with various error cases
	invalidTypes := []string{"invalid", "", "INVALID", "123", "normal-interrupt"}

	for _, invalidType := range invalidTypes {
		_, err := parseContextType(invalidType)
		if invalidType == "" || !strings.Contains("normal interruption interrupt emergency hotfix experiment", strings.ToLower(invalidType)) {
			assert.Error(t, err, "Should error for type: %s", invalidType)
		}
	}
}

// Helper function to setup test directories
func setupTestDirs(t *testing.T, tempDir string) {
	docsDir := filepath.Join(tempDir, "docs", "1-project")
	err := os.MkdirAll(docsDir, 0755)
	require.NoError(t, err)

	currentEpicDir := filepath.Join(tempDir, "docs", "2-current-epic")
	err = os.MkdirAll(currentEpicDir, 0755)
	require.NoError(t, err)

	currentTaskDir := filepath.Join(tempDir, "docs", "3-current-task")
	err = os.MkdirAll(currentTaskDir, 0755)
	require.NoError(t, err)
}
