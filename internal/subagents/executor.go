package subagents

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// SubagentExecutor handles the execution of tasks through specialized subagents
type SubagentExecutor struct {
	router       *SubagentRouter
	claudePath   string
	timeout      time.Duration
	fallbackEnabled bool
}

// ExecutionResult represents the result of a subagent execution
type ExecutionResult struct {
	SubagentName     string           `json:"subagent_name"`
	Success          bool             `json:"success"`
	Output           string           `json:"output"`
	Error            string           `json:"error,omitempty"`
	Duration         time.Duration    `json:"duration"`
	TokenSavings     TokenSavings     `json:"token_savings"`
	RoutingDecision  *RoutingDecision `json:"routing_decision"`
	FallbackUsed     bool             `json:"fallback_used"`
}

// NewSubagentExecutor creates a new subagent executor
func NewSubagentExecutor(router *SubagentRouter, claudePath string) *SubagentExecutor {
	return &SubagentExecutor{
		router:          router,
		claudePath:      claudePath,
		timeout:         120 * time.Second, // 2 minutes default
		fallbackEnabled: true,
	}
}

// Execute runs a task through the appropriate subagent or fallback
func (se *SubagentExecutor) Execute(ctx context.Context, commandPath string, prompt string, contextData map[string]interface{}) (*ExecutionResult, error) {
	startTime := time.Now()

	// Route to appropriate subagent
	decision, err := se.router.Route(ctx, commandPath, contextData)
	if err != nil {
		return nil, fmt.Errorf("routing failed: %w", err)
	}

	result := &ExecutionResult{
		RoutingDecision: decision,
		Duration:       time.Since(startTime),
		TokenSavings:   decision.EstimatedSavings,
	}

	// Execute with chosen subagent or fallback
	if decision.FallbackRequired || decision.SubagentName == "main" {
		return se.executeWithMainAgent(ctx, commandPath, prompt, contextData, result)
	}

	return se.executeWithSubagent(ctx, decision, prompt, contextData, result)
}

// executeWithSubagent executes using a specialized subagent
func (se *SubagentExecutor) executeWithSubagent(ctx context.Context, decision *RoutingDecision, prompt string, contextData map[string]interface{}, result *ExecutionResult) (*ExecutionResult, error) {
	result.SubagentName = decision.SubagentName
	
	// Get subagent configuration
	subagent, err := se.router.manager.GetSubagent(decision.SubagentName)
	if err != nil {
		if se.fallbackEnabled {
			return se.executeWithMainAgent(ctx, decision.Context.CommandPath, prompt, contextData, result)
		}
		return nil, fmt.Errorf("subagent not found: %w", err)
	}

	// Prepare context for subagent (limited context based on configuration)
	limitedContext := se.prepareLimitedContext(contextData, subagent.ContextLimit)
	
	// Create subagent-specific prompt
	subagentPrompt := se.buildSubagentPrompt(subagent, prompt, limitedContext)
	
	// Execute using Claude Code Task tool with subagent
	output, err := se.executeClaudeTask(ctx, decision.SubagentName, subagentPrompt)
	if err != nil {
		result.Error = err.Error()
		
		// Fallback to main agent if enabled
		if se.fallbackEnabled {
			result.FallbackUsed = true
			return se.executeWithMainAgent(ctx, decision.Context.CommandPath, prompt, contextData, result)
		}
		
		result.Success = false
		return result, nil
	}

	result.Success = true
	result.Output = output
	return result, nil
}

// executeWithMainAgent executes using the main Claude Code agent
func (se *SubagentExecutor) executeWithMainAgent(ctx context.Context, commandPath string, prompt string, contextData map[string]interface{}, result *ExecutionResult) (*ExecutionResult, error) {
	result.SubagentName = "main"
	
	// Execute with full context using main agent
	output, err := se.executeMainClaudeTask(ctx, commandPath, prompt, contextData)
	if err != nil {
		result.Error = err.Error()
		result.Success = false
		return result, nil
	}

	result.Success = true
	result.Output = output
	return result, nil
}

// prepareLimitedContext creates a context limited to the subagent's requirements
func (se *SubagentExecutor) prepareLimitedContext(contextData map[string]interface{}, contextLimit int) map[string]interface{} {
	limitedContext := make(map[string]interface{})
	currentSize := 0

	// Priority order for context inclusion
	priorities := []string{"task_type", "project_name", "template_type", "command_path"}
	
	// Include high-priority context first
	for _, key := range priorities {
		if value, exists := contextData[key]; exists {
			if str, ok := value.(string); ok {
				if currentSize + len(str) < contextLimit {
					limitedContext[key] = value
					currentSize += len(str)
				}
			} else {
				limitedContext[key] = value
			}
		}
	}

	// Include other context until limit reached
	for key, value := range contextData {
		// Skip if already included
		if _, exists := limitedContext[key]; exists {
			continue
		}

		if str, ok := value.(string); ok {
			if currentSize + len(str) < contextLimit {
				limitedContext[key] = value
				currentSize += len(str)
			}
		} else {
			limitedContext[key] = value
		}

		if currentSize >= contextLimit {
			break
		}
	}

	return limitedContext
}

// buildSubagentPrompt constructs a prompt optimized for the specific subagent
func (se *SubagentExecutor) buildSubagentPrompt(subagent *SubagentConfig, originalPrompt string, limitedContext map[string]interface{}) string {
	var promptBuilder strings.Builder
	
	// Include subagent's system prompt
	promptBuilder.WriteString(fmt.Sprintf("# Subagent: %s\n\n", subagent.Name))
	promptBuilder.WriteString(fmt.Sprintf("%s\n\n", subagent.SystemPrompt))
	
	// Add limited context
	promptBuilder.WriteString("## Context\n\n")
	for key, value := range limitedContext {
		promptBuilder.WriteString(fmt.Sprintf("- %s: %v\n", key, value))
	}
	
	// Add original prompt
	promptBuilder.WriteString(fmt.Sprintf("\n## Task\n\n%s\n", originalPrompt))
	
	// Add subagent-specific constraints
	promptBuilder.WriteString("\n## Constraints\n\n")
	promptBuilder.WriteString("- Work within your specialized domain only\n")
	promptBuilder.WriteString("- Use minimal context as provided above\n")
	promptBuilder.WriteString("- Focus on efficiency and token optimization\n")
	if len(subagent.Tools) > 0 {
		promptBuilder.WriteString(fmt.Sprintf("- Available tools: %s\n", strings.Join(subagent.Tools, ", ")))
	}

	return promptBuilder.String()
}

// executeClaudeTask executes a task using the Claude Code Task tool with subagent
func (se *SubagentExecutor) executeClaudeTask(ctx context.Context, subagentName string, prompt string) (string, error) {
	// Create a timeout context
	timeoutCtx, cancel := context.WithTimeout(ctx, se.timeout)
	defer cancel()

	// Use the Claude Code Task tool to execute with the specified subagent
	taskPrompt := fmt.Sprintf("Use %s subagent: %s", subagentName, prompt)
	
	// Execute using claude CLI with the task prompt
	cmd := exec.CommandContext(timeoutCtx, se.claudePath)
	
	// Write the task prompt to stdin
	cmd.Stdin = strings.NewReader(taskPrompt)
	
	// Capture output
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("claude task execution failed: %w, output: %s", err, string(output))
	}

	return string(output), nil
}

// executeMainClaudeTask executes using the main Claude agent
func (se *SubagentExecutor) executeMainClaudeTask(ctx context.Context, commandPath string, prompt string, contextData map[string]interface{}) (string, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, se.timeout)
	defer cancel()

	// For main agent, use full context (simulated)
	fullPrompt := fmt.Sprintf("Command: %s\n\nContext: %v\n\nPrompt: %s", commandPath, contextData, prompt)
	
	// Simulate main agent execution
	cmd := exec.CommandContext(timeoutCtx, se.claudePath, "execute")
	cmd.Stdin = strings.NewReader(fullPrompt)
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("main claude execution failed: %w, output: %s", err, string(output))
	}

	return string(output), nil
}

// ExecuteTemplate is a specialized method for template generation
func (se *SubagentExecutor) ExecuteTemplate(ctx context.Context, templateType string, variables map[string]string) (*ExecutionResult, error) {
	commandPath := fmt.Sprintf("templates/%s.md", strings.ToUpper(templateType))
	prompt := fmt.Sprintf("Generate a %s template using the provided variables", templateType)
	
	contextData := map[string]interface{}{
		"task_type":     "template",
		"template_type": templateType,
		"command_path":  commandPath,
	}
	
	// Add variables to context
	for key, value := range variables {
		contextData[key] = value
	}
	
	return se.Execute(ctx, commandPath, prompt, contextData)
}

// ExecuteStatus is a specialized method for status reporting
func (se *SubagentExecutor) ExecuteStatus(ctx context.Context, statusType string, stateData map[string]interface{}) (*ExecutionResult, error) {
	commandPath := fmt.Sprintf("status/%s", statusType)
	prompt := fmt.Sprintf("Generate a %s status report from the provided state data", statusType)
	
	contextData := map[string]interface{}{
		"task_type":    "status",
		"status_type":  statusType,
		"command_path": commandPath,
	}
	
	// Add state data to context
	for key, value := range stateData {
		contextData[key] = value
	}
	
	return se.Execute(ctx, commandPath, prompt, contextData)
}

// ExecutePlanning is a specialized method for task planning
func (se *SubagentExecutor) ExecutePlanning(ctx context.Context, storyDescription string, technicalContext map[string]string) (*ExecutionResult, error) {
	commandPath := "planning/decompose-story"
	prompt := fmt.Sprintf("Decompose the following user story into implementable tasks: %s", storyDescription)
	
	contextData := map[string]interface{}{
		"task_type":         "planning",
		"story_description": storyDescription,
		"command_path":      commandPath,
	}
	
	// Add technical context
	for key, value := range technicalContext {
		contextData[key] = value
	}
	
	return se.Execute(ctx, commandPath, prompt, contextData)
}

// GetMetrics returns current routing and execution metrics
func (se *SubagentExecutor) GetMetrics() *RoutingMetrics {
	return se.router.metrics
}