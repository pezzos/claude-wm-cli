package executor

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"claude-wm-cli/internal/debug"
)

// ClaudeExecutor handles execution of Claude commands
type ClaudeExecutor struct {
	timeout time.Duration
}

// NewClaudeExecutor creates a new Claude command executor
func NewClaudeExecutor() *ClaudeExecutor {
	return &ClaudeExecutor{
		timeout: 30 * time.Minute, // Default 30 minutes timeout for dev - should be enough for any command
	}
}

// SetTimeout sets the timeout for Claude command execution
func (ce *ClaudeExecutor) SetTimeout(timeout time.Duration) {
	ce.timeout = timeout
}

// ExecutePrompt executes a Claude prompt command
func (ce *ClaudeExecutor) ExecutePrompt(prompt, description string) error {
	debug.LogClaudeCommand(prompt, description)
	debug.LogExecution("CLAUDE", "execute prompt", fmt.Sprintf("Long-running Claude analysis with MCP tools (timeout: %v)", ce.timeout))
	
	// Build the command
	cmd := exec.Command("claude", "-p", prompt)
	
	// Set up environment and output
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	
	// In development mode, run without timeout to avoid interrupting long analyses
	if debug.DevMode {
		debug.LogExecution("CLAUDE", "dev mode", "Running without timeout - kill manually if needed (Ctrl+C)")
		err := cmd.Run()
		if err != nil {
			debug.LogResult("CLAUDE", "execute prompt", fmt.Sprintf("Command failed: %v", err), false)
			return fmt.Errorf("claude command failed: %w", err)
		}
		debug.LogResult("CLAUDE", "execute prompt", "Command completed successfully", true)
		return nil
	}
	
	// Production mode with timeout
	done := make(chan error, 1)
	go func() {
		done <- cmd.Run()
	}()
	
	select {
	case err := <-done:
		if err != nil {
			debug.LogResult("CLAUDE", "execute prompt", fmt.Sprintf("Command failed: %v", err), false)
			return fmt.Errorf("claude command failed: %w", err)
		}
		debug.LogResult("CLAUDE", "execute prompt", "Command completed successfully", true)
		return nil
		
	case <-time.After(ce.timeout):
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
		debug.LogResult("CLAUDE", "execute prompt", fmt.Sprintf("Command timed out after %v", ce.timeout), false)
		return fmt.Errorf("claude command timed out after %v", ce.timeout)
	}
}

// ExecuteSlashCommand executes a Claude slash command
func (ce *ClaudeExecutor) ExecuteSlashCommand(slashCommand, description string) error {
	// Slash commands are passed directly as prompts
	return ce.ExecutePrompt(slashCommand, description)
}

// ExecuteSlashCommandWithExitCode executes a Claude slash command and returns the exit code
func (ce *ClaudeExecutor) ExecuteSlashCommandWithExitCode(slashCommand, description string) (int, error) {
	debug.LogClaudeCommand(slashCommand, description)
	debug.LogExecution("CLAUDE", "execute slash command with exit code", fmt.Sprintf("Claude command with exit code tracking (timeout: %v)", ce.timeout))
	
	// Build the command
	cmd := exec.Command("claude", "-p", slashCommand)
	
	// Set up environment and output
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	
	// In development mode, run without timeout
	if debug.DevMode {
		debug.LogExecution("CLAUDE", "dev mode", "Running without timeout - kill manually if needed (Ctrl+C)")
		err := cmd.Run()
		exitCode := getExitCode(err)
		
		debug.LogResult("CLAUDE", "execute slash command with exit code", 
			fmt.Sprintf("Command completed with exit code: %d", exitCode), err == nil)
		
		return exitCode, nil
	}
	
	// Run with timeout in production mode
	done := make(chan error, 1)
	go func() {
		done <- cmd.Run()
	}()
	
	select {
	case err := <-done:
		exitCode := getExitCode(err)
		debug.LogResult("CLAUDE", "execute slash command with exit code", 
			fmt.Sprintf("Command completed with exit code: %d", exitCode), err == nil)
		return exitCode, nil
		
	case <-time.After(ce.timeout):
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
		debug.LogResult("CLAUDE", "execute slash command with exit code", 
			fmt.Sprintf("Command timed out after %v", ce.timeout), false)
		return -1, fmt.Errorf("claude command timed out after %v", ce.timeout)
	}
}

// getExitCode extracts exit code from error
func getExitCode(err error) int {
	if err == nil {
		return 0
	}
	
	if exitError, ok := err.(*exec.ExitError); ok {
		return exitError.ExitCode()
	}
	
	// If we can't determine the exit code, assume failure
	return 1
}

// ValidateClaudeAvailable checks if Claude CLI is available
func (ce *ClaudeExecutor) ValidateClaudeAvailable() error {
	debug.LogExecution("CLAUDE", "validate availability", "Check if claude command is in PATH")
	
	cmd := exec.Command("claude", "--version")
	output, err := cmd.Output()
	
	if err != nil {
		debug.LogResult("CLAUDE", "validate availability", "Claude CLI not found in PATH", false)
		return fmt.Errorf("claude CLI not found: %w", err)
	}
	
	version := strings.TrimSpace(string(output))
	debug.LogResult("CLAUDE", "validate availability", fmt.Sprintf("Claude CLI found: %s", version), true)
	return nil
}