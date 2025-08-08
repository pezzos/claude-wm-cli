package subagents

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// SerenaContextProcessor handles intelligent context preprocessing using Serena
type SerenaContextProcessor struct {
	enabled       bool
	serverEnabled bool
	timeout       time.Duration
}

// SerenaAnalysis represents the result of Serena context analysis
type SerenaAnalysis struct {
	RelevantFiles   []string          `json:"relevant_files"`
	KeySymbols      []string          `json:"key_symbols"`
	Dependencies    []string          `json:"dependencies"`
	ContextSize     int64             `json:"context_size"`
	ProcessedData   map[string]string `json:"processed_data"`
	AnalysisDuration time.Duration    `json:"analysis_duration"`
	TokenSavings    int64             `json:"token_savings"`
}

// SerenaIntegratedRouter extends the existing router with Serena preprocessing
type SerenaIntegratedRouter struct {
	baseRouter       *SubagentRouter
	serenaProcessor  *SerenaContextProcessor
	enabled          bool
	fallbackToBase   bool
}

// NewSerenaContextProcessor creates a new Serena context processor
func NewSerenaContextProcessor() *SerenaContextProcessor {
	return &SerenaContextProcessor{
		enabled:       false, // Will be enabled when Serena MCP is detected
		serverEnabled: false,
		timeout:       30 * time.Second,
	}
}

// NewSerenaIntegratedRouter creates a new Serena-integrated router
func NewSerenaIntegratedRouter(baseRouter *SubagentRouter) *SerenaIntegratedRouter {
	return &SerenaIntegratedRouter{
		baseRouter:      baseRouter,
		serenaProcessor: NewSerenaContextProcessor(),
		enabled:         false, // Will be auto-enabled when Serena is available
		fallbackToBase:  true,
	}
}

// EnableSerena enables Serena integration if the MCP server is available
func (sir *SerenaIntegratedRouter) EnableSerena() error {
	// Check if Serena MCP server is available by attempting to call a tool
	if sir.checkSerenaMCPAvailability() {
		sir.enabled = true
		sir.serenaProcessor.enabled = true
		sir.serenaProcessor.serverEnabled = true
		return nil
	}
	return fmt.Errorf("Serena MCP server not available - falling back to base router")
}

// checkSerenaMCPAvailability checks if Serena MCP tools are available
func (sir *SerenaIntegratedRouter) checkSerenaMCPAvailability() bool {
	// This would be implemented to check for Serena MCP tools
	// For now, we'll assume it's available if explicitly configured
	return false // Will be set to true when Serena MCP is properly configured
}

// Route performs intelligent routing with optional Serena preprocessing
func (sir *SerenaIntegratedRouter) Route(ctx context.Context, commandPath string, contextData map[string]interface{}) (*RoutingDecision, error) {
	startTime := time.Now()

	// If Serena is not enabled, use base router
	if !sir.enabled || !sir.serenaProcessor.enabled {
		return sir.baseRouter.Route(ctx, commandPath, contextData)
	}

	// Step 1: Serena preprocessing
	analysis, err := sir.performSerenaAnalysis(ctx, commandPath, contextData)
	if err != nil {
		// Fallback to base router on Serena failure
		if sir.fallbackToBase {
			sir.baseRouter.metrics.RecordFallback("serena_analysis_failed", time.Since(startTime))
			return sir.baseRouter.Route(ctx, commandPath, contextData)
		}
		return nil, fmt.Errorf("Serena analysis failed: %w", err)
	}

	// Step 2: Enhanced context data with Serena insights
	enhancedContext := sir.mergeSerenaAnalysis(contextData, analysis)

	// Step 3: Route with base router using enhanced context
	decision, err := sir.baseRouter.Route(ctx, commandPath, enhancedContext)
	if err != nil {
		return nil, err
	}

	// Step 4: Update decision with Serena benefits
	sir.enhanceDecisionWithSerena(decision, analysis, time.Since(startTime))

	return decision, nil
}

// performSerenaAnalysis uses Serena MCP tools to analyze and optimize context
func (sir *SerenaIntegratedRouter) performSerenaAnalysis(ctx context.Context, commandPath string, contextData map[string]interface{}) (*SerenaAnalysis, error) {
	startTime := time.Now()

	// Determine analysis strategy based on task type
	analysisType := sir.determineAnalysisType(commandPath, contextData)
	
	analysis := &SerenaAnalysis{
		RelevantFiles:   make([]string, 0),
		KeySymbols:      make([]string, 0),
		Dependencies:    make([]string, 0),
		ProcessedData:   make(map[string]string),
		AnalysisDuration: 0,
		TokenSavings:    0,
	}

	// Simulate Serena analysis (will be replaced with actual MCP calls)
	switch analysisType {
	case "code_review":
		analysis = sir.performCodeReviewAnalysis(ctx, contextData)
	case "template_generation":
		analysis = sir.performTemplateAnalysis(ctx, contextData)
	case "status_reporting":
		analysis = sir.performStatusAnalysis(ctx, contextData)
	case "planning":
		analysis = sir.performPlanningAnalysis(ctx, contextData)
	default:
		analysis = sir.performGeneralAnalysis(ctx, contextData)
	}

	analysis.AnalysisDuration = time.Since(startTime)
	return analysis, nil
}

// determineAnalysisType determines the type of Serena analysis needed
func (sir *SerenaIntegratedRouter) determineAnalysisType(commandPath string, contextData map[string]interface{}) string {
	commandLower := strings.ToLower(commandPath)
	
	// Map command patterns to analysis types
	if strings.Contains(commandLower, "review") || strings.Contains(commandLower, "validate") {
		return "code_review"
	}
	if strings.Contains(commandLower, "template") || strings.Contains(commandLower, "architecture.md") {
		return "template_generation"
	}
	if strings.Contains(commandLower, "status") || strings.Contains(commandLower, "dashboard") {
		return "status_reporting"
	}
	if strings.Contains(commandLower, "plan") || strings.Contains(commandLower, "decompose") {
		return "planning"
	}
	
	return "general"
}

// performCodeReviewAnalysis optimizes context for code review tasks
func (sir *SerenaIntegratedRouter) performCodeReviewAnalysis(ctx context.Context, contextData map[string]interface{}) *SerenaAnalysis {
	// This would use Serena MCP tools to:
	// 1. Identify changed files and their dependencies
	// 2. Extract relevant code symbols and interfaces
	// 3. Focus on security and architecture concerns
	
	return &SerenaAnalysis{
		RelevantFiles: []string{"auth/", "api/handlers.go", "security/"},
		KeySymbols:    []string{"AuthHandler", "SecurityMiddleware", "ValidateToken"},
		Dependencies:  []string{"jwt-go", "bcrypt"},
		ContextSize:   3000, // Reduced from full context
		ProcessedData: map[string]string{
			"focus_area": "security_review",
			"scope":      "authentication_changes",
		},
		TokenSavings: 25000, // Estimated savings vs full context
	}
}

// performTemplateAnalysis optimizes context for template generation
func (sir *SerenaIntegratedRouter) performTemplateAnalysis(ctx context.Context, contextData map[string]interface{}) *SerenaAnalysis {
	// This would use Serena MCP tools to:
	// 1. Extract project metadata and structure
	// 2. Identify technology stack and patterns
	// 3. Focus on architectural components
	
	return &SerenaAnalysis{
		RelevantFiles: []string{"go.mod", "main.go", "cmd/", "internal/"},
		KeySymbols:    []string{"main", "CLI", "Router", "Config"},
		Dependencies:  []string{"cobra", "viper", "yaml"},
		ContextSize:   2000,
		ProcessedData: map[string]string{
			"tech_stack":    "Go CLI application",
			"architecture":  "clean_architecture",
			"domain":       "window_management",
		},
		TokenSavings: 35000,
	}
}

// performStatusAnalysis optimizes context for status reporting
func (sir *SerenaIntegratedRouter) performStatusAnalysis(ctx context.Context, contextData map[string]interface{}) *SerenaAnalysis {
	return &SerenaAnalysis{
		RelevantFiles: []string{"metrics/", "logs/", "state/"},
		KeySymbols:    []string{"Metrics", "Status", "Health"},
		Dependencies:  []string{},
		ContextSize:   1500,
		ProcessedData: map[string]string{
			"data_type": "metrics_and_state",
			"format":    "structured_data",
		},
		TokenSavings: 20000,
	}
}

// performPlanningAnalysis optimizes context for task planning
func (sir *SerenaIntegratedRouter) performPlanningAnalysis(ctx context.Context, contextData map[string]interface{}) *SerenaAnalysis {
	return &SerenaAnalysis{
		RelevantFiles: []string{"docs/", "internal/", "cmd/"},
		KeySymbols:    []string{"Controller", "Service", "Repository"},
		Dependencies:  []string{},
		ContextSize:   8000,
		ProcessedData: map[string]string{
			"scope":        "feature_planning",
			"complexity":   "medium",
			"dependencies": "minimal",
		},
		TokenSavings: 15000,
	}
}

// performGeneralAnalysis provides general context optimization
func (sir *SerenaIntegratedRouter) performGeneralAnalysis(ctx context.Context, contextData map[string]interface{}) *SerenaAnalysis {
	return &SerenaAnalysis{
		RelevantFiles: []string{"main.go", "internal/"},
		KeySymbols:    []string{},
		Dependencies:  []string{},
		ContextSize:   5000,
		ProcessedData: map[string]string{
			"analysis_type": "general",
		},
		TokenSavings: 10000,
	}
}

// mergeSerenaAnalysis combines original context with Serena insights
func (sir *SerenaIntegratedRouter) mergeSerenaAnalysis(originalContext map[string]interface{}, analysis *SerenaAnalysis) map[string]interface{} {
	enhancedContext := make(map[string]interface{})
	
	// Copy original context
	for key, value := range originalContext {
		enhancedContext[key] = value
	}
	
	// Add Serena insights
	enhancedContext["serena_relevant_files"] = analysis.RelevantFiles
	enhancedContext["serena_key_symbols"] = analysis.KeySymbols
	enhancedContext["serena_dependencies"] = analysis.Dependencies
	enhancedContext["serena_context_size"] = analysis.ContextSize
	enhancedContext["serena_processed_data"] = analysis.ProcessedData
	
	// Override context size estimation
	enhancedContext["estimated_context_size"] = analysis.ContextSize
	
	return enhancedContext
}

// enhanceDecisionWithSerena updates routing decision with Serena benefits
func (sir *SerenaIntegratedRouter) enhanceDecisionWithSerena(decision *RoutingDecision, analysis *SerenaAnalysis, totalDuration time.Duration) {
	// Update token savings with Serena benefits
	additionalSavings := analysis.TokenSavings
	decision.EstimatedSavings.SavedTokens += additionalSavings
	decision.EstimatedSavings.SubagentTokens = int64(analysis.ContextSize) + 1000 // Base overhead
	
	// Recalculate savings percentage
	if decision.EstimatedSavings.OriginalTokens > 0 {
		decision.EstimatedSavings.SavingsPercent = float64(decision.EstimatedSavings.SavedTokens) / float64(decision.EstimatedSavings.OriginalTokens) * 100
	}
	
	// Update reason to include Serena
	decision.Reason = fmt.Sprintf("serena_enhanced_%s_analysis_%.1fs", decision.Reason, analysis.AnalysisDuration.Seconds())
	
	// Add Serena metadata
	decision.Context.Metadata["serena_enabled"] = "true"
	decision.Context.Metadata["serena_analysis_duration"] = fmt.Sprintf("%.1fs", analysis.AnalysisDuration.Seconds())
	decision.Context.Metadata["serena_token_savings"] = fmt.Sprintf("%d", additionalSavings)
}

// GetSerenaStatus returns the current status of Serena integration
func (sir *SerenaIntegratedRouter) GetSerenaStatus() map[string]interface{} {
	return map[string]interface{}{
		"enabled":         sir.enabled,
		"processor_ready": sir.serenaProcessor.enabled,
		"server_available": sir.serenaProcessor.serverEnabled,
		"fallback_enabled": sir.fallbackToBase,
		"timeout":         sir.serenaProcessor.timeout.String(),
	}
}

// DisableSerena disables Serena integration and falls back to base router
func (sir *SerenaIntegratedRouter) DisableSerena() {
	sir.enabled = false
	sir.serenaProcessor.enabled = false
}