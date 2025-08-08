package subagents

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// RoutingDecision represents the result of subagent routing
type RoutingDecision struct {
	SubagentName     string        `json:"subagent_name"`
	Confidence       float64       `json:"confidence"`
	Reason           string        `json:"reason"`
	EstimatedSavings TokenSavings  `json:"estimated_savings"`
	FallbackRequired bool          `json:"fallback_required"`
	Context          RouteContext  `json:"context"`
}

// TokenSavings represents the estimated token savings from using a subagent
type TokenSavings struct {
	OriginalTokens  int64   `json:"original_tokens"`
	SubagentTokens  int64   `json:"subagent_tokens"`
	SavedTokens     int64   `json:"saved_tokens"`
	SavingsPercent  float64 `json:"savings_percent"`
}

// RouteContext contains metadata about the routing decision
type RouteContext struct {
	CommandPath     string            `json:"command_path"`
	TaskType        string            `json:"task_type"`
	ContextSize     int64             `json:"context_size"`
	Complexity      ComplexityLevel   `json:"complexity"`
	Priority        Priority          `json:"priority"`
	Metadata        map[string]string `json:"metadata"`
}

type ComplexityLevel string

const (
	ComplexityLow    ComplexityLevel = "low"
	ComplexityMedium ComplexityLevel = "medium"
	ComplexityHigh   ComplexityLevel = "high"
)

type Priority string

const (
	PriorityLow      Priority = "low"
	PriorityMedium   Priority = "medium"
	PriorityHigh     Priority = "high"
	PriorityCritical Priority = "critical"
)

// SubagentRouter handles intelligent routing to specialized subagents
type SubagentRouter struct {
	manager   *SubagentManager
	metrics   *RoutingMetrics
	enabled   bool
	threshold float64 // Minimum confidence for subagent routing
}

// NewSubagentRouter creates a new subagent router
func NewSubagentRouter(manager *SubagentManager) *SubagentRouter {
	return &SubagentRouter{
		manager:   manager,
		metrics:   NewRoutingMetrics(),
		enabled:   true,
		threshold: 0.3, // 30% confidence minimum - more aggressive routing
	}
}

// Route determines the optimal subagent for a given task
func (sr *SubagentRouter) Route(ctx context.Context, commandPath string, contextData map[string]interface{}) (*RoutingDecision, error) {
	startTime := time.Now()
	
	// Analyze the routing context
	routeContext := sr.analyzeContext(commandPath, contextData)
	
	// Find the best subagent match
	subagent, confidence := sr.manager.MatchSubagent(commandPath)
	
	decision := &RoutingDecision{
		Context:    routeContext,
		Confidence: confidence,
	}

	// If no good subagent match, use main agent
	if subagent == nil || confidence < sr.threshold {
		decision.SubagentName = "main"
		decision.FallbackRequired = true
		decision.Reason = fmt.Sprintf("no_suitable_subagent_confidence_%.2f", confidence)
		sr.metrics.RecordFallback("no_match", time.Since(startTime))
		return decision, nil
	}

	// Calculate estimated savings
	savings := sr.calculateTokenSavings(routeContext, subagent)
	
	decision.SubagentName = subagent.Name
	decision.Reason = fmt.Sprintf("matched_patterns_confidence_%.2f", confidence)
	decision.EstimatedSavings = savings
	decision.FallbackRequired = false

	// Record successful routing
	sr.metrics.RecordRouting(subagent.Name, confidence, savings, time.Since(startTime))
	
	return decision, nil
}

// analyzeContext extracts routing context from command and data
func (sr *SubagentRouter) analyzeContext(commandPath string, contextData map[string]interface{}) RouteContext {
	context := RouteContext{
		CommandPath: commandPath,
		Metadata:    make(map[string]string),
		ContextSize: sr.estimateContextSize(contextData),
	}

	// Determine task type from command path
	context.TaskType = sr.extractTaskType(commandPath)
	
	// Assess complexity based on task type and context
	context.Complexity = sr.assessComplexity(context.TaskType, context.ContextSize)
	
	// Determine priority
	context.Priority = sr.determinePriority(commandPath, context.TaskType)
	
	// Extract metadata
	if projectName, ok := contextData["project_name"].(string); ok {
		context.Metadata["project_name"] = projectName
	}
	if taskType, ok := contextData["task_type"].(string); ok {
		context.Metadata["task_type"] = taskType
	}

	return context
}

// extractTaskType determines the task type from command path
func (sr *SubagentRouter) extractTaskType(commandPath string) string {
	commandLower := strings.ToLower(commandPath)
	
	// Template patterns
	if strings.Contains(commandLower, "template") || 
	   strings.Contains(commandLower, "architecture.md") ||
	   strings.Contains(commandLower, "prd.md") ||
	   strings.Contains(commandLower, "technical.md") {
		return "template"
	}
	
	// Status/monitoring patterns  
	if strings.Contains(commandLower, "status") ||
	   strings.Contains(commandLower, "dashboard") ||
	   strings.Contains(commandLower, "debug") ||
	   strings.Contains(commandLower, "metrics") {
		return "status"
	}
	
	// Planning patterns
	if strings.Contains(commandLower, "plan") ||
	   strings.Contains(commandLower, "decompose") ||
	   strings.Contains(commandLower, "estimate") {
		return "planning"
	}
	
	// Review patterns
	if strings.Contains(commandLower, "review") ||
	   strings.Contains(commandLower, "validate") ||
	   strings.Contains(commandLower, "architecture-review") {
		return "review"
	}

	return "general"
}

// assessComplexity determines task complexity
func (sr *SubagentRouter) assessComplexity(taskType string, contextSize int64) ComplexityLevel {
	// Base complexity by task type
	baseComplexity := map[string]ComplexityLevel{
		"template":  ComplexityLow,
		"status":    ComplexityLow,
		"planning":  ComplexityHigh,
		"review":    ComplexityMedium,
		"general":   ComplexityMedium,
	}

	base := baseComplexity[taskType]
	
	// Adjust based on context size
	switch {
	case contextSize > 100000: // >100KB context
		if base == ComplexityLow {
			return ComplexityMedium
		}
		return ComplexityHigh
	case contextSize > 50000: // >50KB context
		if base == ComplexityLow {
			return ComplexityMedium
		}
		return base
	default:
		return base
	}
}

// determinePriority assesses task priority
func (sr *SubagentRouter) determinePriority(commandPath, taskType string) Priority {
	commandLower := strings.ToLower(commandPath)
	
	// Critical patterns
	if strings.Contains(commandLower, "implement") ||
	   strings.Contains(commandLower, "security") ||
	   strings.Contains(commandLower, "production") {
		return PriorityCritical
	}
	
	// High priority patterns
	if strings.Contains(commandLower, "review") ||
	   strings.Contains(commandLower, "validate") {
		return PriorityHigh
	}
	
	// Medium priority for planning and templates
	if taskType == "planning" || taskType == "template" {
		return PriorityMedium
	}
	
	// Low priority for status/debug
	if taskType == "status" {
		return PriorityLow
	}

	return PriorityMedium
}

// calculateTokenSavings estimates token savings from using a subagent
func (sr *SubagentRouter) calculateTokenSavings(context RouteContext, subagent *SubagentConfig) TokenSavings {
	// Baseline estimates for full-context operations
	originalTokens := sr.estimateOriginalTokens(context.TaskType, context.ContextSize)
	
	// Subagent token usage based on context limits
	subagentTokens := int64(subagent.ContextLimit) + 1000 // Base overhead
	
	savedTokens := originalTokens - subagentTokens
	if savedTokens < 0 {
		savedTokens = 0
	}
	
	savingsPercent := 0.0
	if originalTokens > 0 {
		savingsPercent = float64(savedTokens) / float64(originalTokens) * 100
	}

	return TokenSavings{
		OriginalTokens:  originalTokens,
		SubagentTokens:  subagentTokens,
		SavedTokens:     savedTokens,
		SavingsPercent:  savingsPercent,
	}
}

// estimateOriginalTokens estimates tokens for full-context operations
func (sr *SubagentRouter) estimateOriginalTokens(taskType string, contextSize int64) int64 {
	// Base token estimates for different task types (full context)
	baseTokens := map[string]int64{
		"template": 70000,  // Full project context for templates
		"status":   45000,  // Full state + history
		"planning": 100000, // Full codebase analysis  
		"review":   120000, // Full project + diff context
		"general":  80000,  // Average full context
	}

	base := baseTokens[taskType]
	if base == 0 {
		base = 80000 // Default
	}

	// Adjust based on actual context size
	if contextSize > 0 {
		// Use actual context size as a factor
		estimated := contextSize * 4 // Rough token estimation (4 chars per token)
		if estimated > base {
			return estimated
		}
	}

	return base
}

// estimateContextSize estimates context size from data
func (sr *SubagentRouter) estimateContextSize(contextData map[string]interface{}) int64 {
	totalSize := int64(0)
	
	for _, value := range contextData {
		if str, ok := value.(string); ok {
			totalSize += int64(len(str))
		}
	}
	
	// Add base overhead for system prompts, etc.
	totalSize += 5000
	
	return totalSize
}