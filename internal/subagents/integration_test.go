package subagents

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestSubagentManagerInitialization tests the basic initialization of subagent manager
func TestSubagentManagerInitialization(t *testing.T) {
	// Create temporary directory for test configuration
	tmpDir, err := os.MkdirTemp("", "subagent-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test configuration files
	if err := createTestSubagentConfigs(tmpDir); err != nil {
		t.Fatalf("Failed to create test configs: %v", err)
	}

	// Initialize manager
	manager, err := NewSubagentManager(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create subagent manager: %v", err)
	}

	// Test that subagents were loaded
	subagents := manager.ListSubagents()
	if len(subagents) == 0 {
		t.Error("No subagents were loaded")
	}

	expectedSubagents := []string{
		"claude-wm-templates",
		"claude-wm-status",
		"claude-wm-planner",
		"claude-wm-reviewer",
	}

	for _, expected := range expectedSubagents {
		found := false
		for _, actual := range subagents {
			if actual == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected subagent %s not found in loaded subagents", expected)
		}
	}
}

// TestSubagentRouting tests the routing logic for different command types
func TestSubagentRouting(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "subagent-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	if err := createTestSubagentConfigs(tmpDir); err != nil {
		t.Fatalf("Failed to create test configs: %v", err)
	}

	manager, err := NewSubagentManager(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create subagent manager: %v", err)
	}

	router := NewSubagentRouter(manager)

	testCases := []struct {
		commandPath       string
		expectedSubagent  string
		expectedConfidence float64
	}{
		{
			commandPath:      "templates/ARCHITECTURE.md",
			expectedSubagent: "claude-wm-templates",
			expectedConfidence: 0.5, // Realistic expectation for pattern matching
		},
		{
			commandPath:      "learning/dashboard.md",
			expectedSubagent: "claude-wm-status",
			expectedConfidence: 0.3, // Should match "dashboard.md" and "learning/"
		},
		{
			commandPath:      "4-task/2-execute/1-Plan-Task.md",
			expectedSubagent: "claude-wm-planner",
			expectedConfidence: 0.3, // Should match "Plan-Task"
		},
		{
			commandPath:      "validation/1-Architecture-Review.md",
			expectedSubagent: "claude-wm-reviewer",
			expectedConfidence: 0.4, // Should match "Architecture-Review"
		},
	}

	for _, tc := range testCases {
		t.Run(tc.commandPath, func(t *testing.T) {
			ctx := context.Background()
			contextData := map[string]interface{}{
				"command_path": tc.commandPath,
			}

			decision, err := router.Route(ctx, tc.commandPath, contextData)
			if err != nil {
				t.Fatalf("Routing failed: %v", err)
			}

			if decision.SubagentName != tc.expectedSubagent {
				t.Errorf("Expected subagent %s, got %s", tc.expectedSubagent, decision.SubagentName)
			}

			if decision.Confidence < tc.expectedConfidence {
				t.Errorf("Expected confidence >= %f, got %f", tc.expectedConfidence, decision.Confidence)
			}
		})
	}
}

// TestTokenSavingsCalculation tests the token savings estimation
func TestTokenSavingsCalculation(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "subagent-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	if err := createTestSubagentConfigs(tmpDir); err != nil {
		t.Fatalf("Failed to create test configs: %v", err)
	}

	manager, err := NewSubagentManager(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create subagent manager: %v", err)
	}

	router := NewSubagentRouter(manager)

	// Test template generation token savings
	subagent, err := manager.GetSubagent("claude-wm-templates")
	if err != nil {
		t.Fatalf("Failed to get template subagent: %v", err)
	}

	routeContext := RouteContext{
		TaskType:    "template",
		ContextSize: 5000,
	}

	savings := router.calculateTokenSavings(routeContext, subagent)

	// Template generation should save significant tokens
	if savings.SavingsPercent < 80 { // At least 80% savings expected
		t.Errorf("Expected at least 80%% token savings, got %.1f%%", savings.SavingsPercent)
	}

	if savings.SavedTokens <= 0 {
		t.Error("Expected positive token savings")
	}

	t.Logf("Template generation savings: %d tokens (%.1f%%)", 
		savings.SavedTokens, savings.SavingsPercent)
}

// TestSubagentMetrics tests the metrics collection and reporting
func TestSubagentMetrics(t *testing.T) {
	metrics := NewRoutingMetrics()

	// Simulate some routing decisions
	testSavings := TokenSavings{
		OriginalTokens:  70000,
		SubagentTokens:  5000,
		SavedTokens:     65000,
		SavingsPercent:  92.9,
	}

	// Record successful routing
	metrics.RecordRouting("claude-wm-templates", 0.95, testSavings, 2*time.Second)
	metrics.RecordRouting("claude-wm-status", 0.85, TokenSavings{
		OriginalTokens: 45000,
		SubagentTokens: 3000,
		SavedTokens:    42000,
		SavingsPercent: 93.3,
	}, 1500*time.Millisecond)

	// Record a fallback
	metrics.RecordFallback("no_match", 500*time.Millisecond)

	// Test metrics calculation
	if metrics.TotalRoutings != 3 {
		t.Errorf("Expected 3 total routings, got %d", metrics.TotalRoutings)
	}

	if metrics.SuccessfulRoutings != 2 {
		t.Errorf("Expected 2 successful routings, got %d", metrics.SuccessfulRoutings)
	}

	if metrics.FallbacksRequired != 1 {
		t.Errorf("Expected 1 fallback, got %d", metrics.FallbacksRequired)
	}

	// Test summary generation
	summary := metrics.GetSummary()
	if len(summary) == 0 {
		t.Error("Expected non-empty summary")
	}

	t.Logf("Metrics summary:\n%s", summary)
}

// TestRoutingContextAnalysis tests context analysis for routing decisions
func TestRoutingContextAnalysis(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "subagent-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	if err := createTestSubagentConfigs(tmpDir); err != nil {
		t.Fatalf("Failed to create test configs: %v", err)
	}

	manager, err := NewSubagentManager(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create subagent manager: %v", err)
	}

	router := NewSubagentRouter(manager)

	testCases := []struct {
		commandPath  string
		expectedType string
		expectedComplexity ComplexityLevel
	}{
		{
			commandPath:        "templates/PRD.md",
			expectedType:       "template",
			expectedComplexity: ComplexityLow,
		},
		{
			commandPath:        "4-task/2-execute/1-Plan-Task.md",
			expectedType:       "planning",
			expectedComplexity: ComplexityHigh,
		},
		{
			commandPath:        "debug/1-Check-state.md",
			expectedType:       "status",
			expectedComplexity: ComplexityLow,
		},
		{
			commandPath:        "validation/1-Architecture-Review.md",
			expectedType:       "review",
			expectedComplexity: ComplexityMedium,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.commandPath, func(t *testing.T) {
			contextData := map[string]interface{}{
				"command_path": tc.commandPath,
				"description":  "Test description",
			}

			context := router.analyzeContext(tc.commandPath, contextData)

			if context.TaskType != tc.expectedType {
				t.Errorf("Expected task type %s, got %s", tc.expectedType, context.TaskType)
			}

			if context.Complexity != tc.expectedComplexity {
				t.Errorf("Expected complexity %s, got %s", tc.expectedComplexity, context.Complexity)
			}
		})
	}
}

// Helper function to create test subagent configuration files
func createTestSubagentConfigs(dir string) error {
	configs := map[string]string{
		"template-generator.yaml": `name: claude-wm-templates
description: "Test template generator"
system_prompt: "Generate templates"
tools: ["Read", "Write"]
triggers:
  patterns: ["templates/", "ARCHITECTURE.md", "PRD.md", "TECHNICAL.md"]
context_limit: 8000
cost_optimization: high`,

		"status-reporter.yaml": `name: claude-wm-status
description: "Test status reporter"
system_prompt: "Generate status reports"
tools: ["Read"]
triggers:
  patterns: ["dashboard.md", "status", "debug/", "learning/", "metrics/"]
context_limit: 5000
cost_optimization: maximum`,

		"task-planner.yaml": `name: claude-wm-planner
description: "Test task planner"
system_prompt: "Plan tasks"
tools: ["Read", "Write"]
triggers:
  patterns: ["Plan-Task", "decompose", "planning", "1-Plan-"]
context_limit: 15000
cost_optimization: high`,

		"code-reviewer.yaml": `name: claude-wm-reviewer
description: "Test code reviewer"
system_prompt: "Review code"
tools: ["Read", "Edit"]
triggers:
  patterns: ["Review-Task", "validate", "Architecture-Review", "review"]
context_limit: 25000
cost_optimization: medium`,
	}

	for filename, content := range configs {
		path := filepath.Join(dir, filename)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return err
		}
	}

	return nil
}