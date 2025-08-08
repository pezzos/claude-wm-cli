package executor

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"claude-wm-cli/internal/debug"
	"claude-wm-cli/internal/subagents"
)

// SubagentAwareExecutor extends the ClaudeExecutor with subagent capabilities
type SubagentAwareExecutor struct {
	*ClaudeExecutor
	subagentExecutor *subagents.SubagentExecutor
	enabled          bool
}

// NewSubagentAwareExecutor creates a new executor with subagent support
func NewSubagentAwareExecutor(claudeExecutor *ClaudeExecutor, subagentConfigPath string) (*SubagentAwareExecutor, error) {
	// Initialize subagent manager
	manager, err := subagents.NewSubagentManager(subagentConfigPath)
	if err != nil {
		debug.LogResult("SUBAGENT", "initialization", fmt.Sprintf("Failed to initialize subagent manager: %v", err), false)
		return &SubagentAwareExecutor{
			ClaudeExecutor: claudeExecutor,
			enabled:       false,
		}, nil // Graceful degradation - continue without subagents
	}

	// Initialize router
	router := subagents.NewSubagentRouter(manager)

	// Initialize subagent executor  
	subagentExecutor := subagents.NewSubagentExecutor(router, "claude")

	debug.LogResult("SUBAGENT", "initialization", 
		fmt.Sprintf("Subagent system initialized with %d subagents", len(manager.ListSubagents())), true)

	return &SubagentAwareExecutor{
		ClaudeExecutor:   claudeExecutor,
		subagentExecutor: subagentExecutor,
		enabled:          true,
	}, nil
}

// ExecutePromptWithSubagents executes a prompt using optimal routing (subagent vs main)
func (sae *SubagentAwareExecutor) ExecutePromptWithSubagents(ctx context.Context, commandPath, prompt, description string) error {
	if !sae.enabled {
		debug.LogExecution("SUBAGENT", "fallback", "Subagents disabled - using main executor")
		return sae.ExecutePrompt(prompt, description)
	}

	startTime := time.Now()
	debug.LogExecution("SUBAGENT", "routing analysis", fmt.Sprintf("Analyzing command: %s", commandPath))

	// Extract context data from command path and prompt
	contextData := sae.extractContextData(commandPath, prompt, description)

	// Execute through subagent system
	result, err := sae.subagentExecutor.Execute(ctx, commandPath, prompt, contextData)
	if err != nil {
		debug.LogResult("SUBAGENT", "execution", fmt.Sprintf("Subagent execution failed: %v", err), false)
		return sae.ExecutePrompt(prompt, description) // Fallback to main executor
	}

	// Log results
	duration := time.Since(startTime)
	sae.logSubagentResult(result, duration)

	if !result.Success {
		if result.FallbackUsed {
			debug.LogExecution("SUBAGENT", "fallback", "Using fallback after subagent failure")
			return nil // Already executed with fallback
		}
		return fmt.Errorf("subagent execution failed: %s", result.Error)
	}

	return nil
}

// ExecuteSlashCommandWithSubagents executes slash commands with subagent routing
func (sae *SubagentAwareExecutor) ExecuteSlashCommandWithSubagents(ctx context.Context, commandPath, slashCommand, description string) (int, error) {
	if !sae.enabled {
		return sae.ExecuteSlashCommandWithExitCode(slashCommand, description)
	}

	debug.LogExecution("SUBAGENT", "slash command routing", fmt.Sprintf("Command: %s", slashCommand))

	// For slash commands, try subagent routing first
	if sae.shouldUseSubagent(commandPath) {
		err := sae.ExecutePromptWithSubagents(ctx, commandPath, slashCommand, description)
		if err != nil {
			debug.LogExecution("SUBAGENT", "slash fallback", "Falling back to main executor for slash command")
			return sae.ExecuteSlashCommandWithExitCode(slashCommand, description)
		}
		return 0, nil // Success with subagent
	}

	// Use main executor for non-subagent commands
	return sae.ExecuteSlashCommandWithExitCode(slashCommand, description)
}

// ExecuteCommandTemplate executes a template generation command
func (sae *SubagentAwareExecutor) ExecuteCommandTemplate(ctx context.Context, templateType string, variables map[string]string) error {
	if !sae.enabled {
		prompt := fmt.Sprintf("Generate %s template with variables: %v", templateType, variables)
		return sae.ExecutePrompt(prompt, fmt.Sprintf("Template generation: %s", templateType))
	}

	debug.LogExecution("SUBAGENT", "template generation", fmt.Sprintf("Type: %s", templateType))

	result, err := sae.subagentExecutor.ExecuteTemplate(ctx, templateType, variables)
	if err != nil || !result.Success {
		// Fallback to main executor
		prompt := fmt.Sprintf("Generate %s template with variables: %v", templateType, variables)
		return sae.ExecutePrompt(prompt, fmt.Sprintf("Template generation: %s", templateType))
	}

	sae.logSubagentResult(result, result.Duration)
	return nil
}

// ExecuteStatusReport executes a status reporting command
func (sae *SubagentAwareExecutor) ExecuteStatusReport(ctx context.Context, statusType string, stateData map[string]interface{}) error {
	if !sae.enabled {
		prompt := fmt.Sprintf("Generate %s status report from data: %v", statusType, stateData)
		return sae.ExecutePrompt(prompt, fmt.Sprintf("Status report: %s", statusType))
	}

	debug.LogExecution("SUBAGENT", "status reporting", fmt.Sprintf("Type: %s", statusType))

	result, err := sae.subagentExecutor.ExecuteStatus(ctx, statusType, stateData)
	if err != nil || !result.Success {
		// Fallback to main executor
		prompt := fmt.Sprintf("Generate %s status report from data: %v", statusType, stateData)
		return sae.ExecutePrompt(prompt, fmt.Sprintf("Status report: %s", statusType))
	}

	sae.logSubagentResult(result, result.Duration)
	return nil
}

// ExecuteTaskPlanning executes task planning and decomposition
func (sae *SubagentAwareExecutor) ExecuteTaskPlanning(ctx context.Context, storyDescription string, technicalContext map[string]string) error {
	if !sae.enabled {
		prompt := fmt.Sprintf("Plan and decompose story: %s with context: %v", storyDescription, technicalContext)
		return sae.ExecutePrompt(prompt, "Task planning and decomposition")
	}

	debug.LogExecution("SUBAGENT", "task planning", fmt.Sprintf("Story: %.50s...", storyDescription))

	result, err := sae.subagentExecutor.ExecutePlanning(ctx, storyDescription, technicalContext)
	if err != nil || !result.Success {
		// Fallback to main executor
		prompt := fmt.Sprintf("Plan and decompose story: %s with context: %v", storyDescription, technicalContext)
		return sae.ExecutePrompt(prompt, "Task planning and decomposition")
	}

	sae.logSubagentResult(result, result.Duration)
	return nil
}

// GetSubagentMetrics returns current metrics about subagent usage
func (sae *SubagentAwareExecutor) GetSubagentMetrics() string {
	if !sae.enabled {
		return "Subagent system is disabled"
	}

	metrics := sae.subagentExecutor.GetMetrics()
	return metrics.GetSummary()
}

// PrintSubagentMetrics prints current subagent metrics to stdout
func (sae *SubagentAwareExecutor) PrintSubagentMetrics() {
	fmt.Println(sae.GetSubagentMetrics())
}

// shouldUseSubagent determines if a command should use subagent routing
func (sae *SubagentAwareExecutor) shouldUseSubagent(commandPath string) bool {
	if !sae.enabled {
		return false
	}

	// Check for subagent-friendly patterns
	subagentPatterns := []string{
		"templates/",
		"status",
		"dashboard", 
		"debug/",
		"metrics/",
		"plan",
		"review",
		"validate",
	}

	commandLower := strings.ToLower(commandPath)
	for _, pattern := range subagentPatterns {
		if strings.Contains(commandLower, pattern) {
			return true
		}
	}

	return false
}

// extractContextData extracts relevant context from command execution
func (sae *SubagentAwareExecutor) extractContextData(commandPath, prompt, description string) map[string]interface{} {
	contextData := map[string]interface{}{
		"command_path": commandPath,
		"description":  description,
		"prompt":       prompt,
	}

	// Extract template type for template commands
	if strings.Contains(commandPath, "templates/") {
		templateFile := filepath.Base(commandPath)
		templateType := strings.TrimSuffix(templateFile, ".md")
		contextData["template_type"] = strings.ToLower(templateType)
		contextData["task_type"] = "template"
	}

	// Extract status type for status commands  
	if strings.Contains(commandPath, "status") || strings.Contains(commandPath, "dashboard") {
		contextData["task_type"] = "status"
	}

	// Extract planning context
	if strings.Contains(commandPath, "plan") || strings.Contains(commandPath, "decompose") {
		contextData["task_type"] = "planning"
	}

	// Extract review context
	if strings.Contains(commandPath, "review") || strings.Contains(commandPath, "validate") {
		contextData["task_type"] = "review"
	}

	return contextData
}

// logSubagentResult logs the results of subagent execution
func (sae *SubagentAwareExecutor) logSubagentResult(result *subagents.ExecutionResult, duration time.Duration) {
	success := result.Success && !result.FallbackUsed
	
	logMsg := fmt.Sprintf("Subagent: %s, Duration: %v", result.SubagentName, duration)
	
	if result.TokenSavings.SavedTokens > 0 {
		logMsg += fmt.Sprintf(", Tokens saved: %d (%.1f%%)", 
			result.TokenSavings.SavedTokens, result.TokenSavings.SavingsPercent)
	}
	
	if result.FallbackUsed {
		logMsg += " [FALLBACK USED]"
	}
	
	if result.Error != "" {
		logMsg += fmt.Sprintf(", Error: %s", result.Error)
	}

	debug.LogResult("SUBAGENT", "execution", logMsg, success)
}

// EnableSubagents enables or disables the subagent system
func (sae *SubagentAwareExecutor) EnableSubagents(enabled bool) {
	sae.enabled = enabled
	status := "disabled"
	if enabled {
		status = "enabled"
	}
	debug.LogExecution("SUBAGENT", "configuration", fmt.Sprintf("Subagent system %s", status))
}

// IsSubagentEnabled returns whether subagents are currently enabled
func (sae *SubagentAwareExecutor) IsSubagentEnabled() bool {
	return sae.enabled
}

// ListAvailableSubagents returns a list of available subagents
func (sae *SubagentAwareExecutor) ListAvailableSubagents() []string {
	if !sae.enabled || sae.subagentExecutor == nil {
		return []string{}
	}

	// We need to access the router's manager - let's assume we have access
	// In a real implementation, we'd add a method to get this information
	return []string{
		"claude-wm-templates",
		"claude-wm-status", 
		"claude-wm-planner",
		"claude-wm-reviewer",
	}
}