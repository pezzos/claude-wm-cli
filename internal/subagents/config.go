package subagents

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// SubagentConfig represents the configuration for a specialized subagent
type SubagentConfig struct {
	Name             string            `yaml:"name"`
	Description      string            `yaml:"description"`
	SystemPrompt     string            `yaml:"system_prompt"`
	Tools            []string          `yaml:"tools"`
	Triggers         TriggerConfig     `yaml:"triggers"`
	ContextLimit     int              `yaml:"context_limit"`
	CostOptimization string           `yaml:"cost_optimization"`
}

// TriggerConfig defines when a subagent should be activated
type TriggerConfig struct {
	Patterns []string `yaml:"patterns"`
}

// SubagentManager manages the lifecycle and routing of specialized subagents
type SubagentManager struct {
	subagents map[string]*SubagentConfig
	configDir string
}

// NewSubagentManager creates a new subagent manager
func NewSubagentManager(configDir string) (*SubagentManager, error) {
	manager := &SubagentManager{
		subagents: make(map[string]*SubagentConfig),
		configDir: configDir,
	}

	// Ensure config directory exists
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create subagents config directory: %w", err)
	}

	// Load all subagent configurations
	if err := manager.loadSubagents(); err != nil {
		return nil, fmt.Errorf("failed to load subagents: %w", err)
	}

	return manager, nil
}

// loadSubagents loads all subagent configurations from Claude Code agent files (.md)
func (sm *SubagentManager) loadSubagents() error {
	// First try to load from .claude/agents/ (Claude Code format)
	claudeAgentsDir := filepath.Join(filepath.Dir(sm.configDir), ".claude", "agents")
	if err := sm.loadClaudeAgents(claudeAgentsDir); err == nil {
		return nil
	}

	// Fallback to YAML files in subagents directory
	yamlFiles, err := filepath.Glob(filepath.Join(sm.configDir, "*.yaml"))
	if err != nil {
		return fmt.Errorf("failed to find subagent config files: %w", err)
	}

	for _, yamlFile := range yamlFiles {
		config, err := sm.loadSubagentConfig(yamlFile)
		if err != nil {
			return fmt.Errorf("failed to load subagent config %s: %w", yamlFile, err)
		}
		sm.subagents[config.Name] = config
	}

	return nil
}

// loadClaudeAgents loads subagents from Claude Code agents directory
func (sm *SubagentManager) loadClaudeAgents(agentsDir string) error {
	mdFiles, err := filepath.Glob(filepath.Join(agentsDir, "claude-wm-*.md"))
	if err != nil {
		return fmt.Errorf("failed to find Claude agent files: %w", err)
	}

	if len(mdFiles) == 0 {
		return fmt.Errorf("no Claude agent files found")
	}

	for _, mdFile := range mdFiles {
		config, err := sm.loadClaudeAgentConfig(mdFile)
		if err != nil {
			return fmt.Errorf("failed to load Claude agent config %s: %w", mdFile, err)
		}
		sm.subagents[config.Name] = config
	}

	return nil
}

// loadSubagentConfig loads a single subagent configuration
func (sm *SubagentManager) loadSubagentConfig(filePath string) (*SubagentConfig, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config SubagentConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML: %w", err)
	}

	return &config, nil
}

// loadClaudeAgentConfig loads a Claude Code agent configuration from markdown file
func (sm *SubagentManager) loadClaudeAgentConfig(filePath string) (*SubagentConfig, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read agent file: %w", err)
	}

	content := string(data)
	
	// Parse frontmatter YAML
	frontmatter, systemPrompt, err := sm.parseFrontmatter(content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse frontmatter: %w", err)
	}

	// Extract name and description from frontmatter
	name, ok := frontmatter["name"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid name in frontmatter")
	}

	description, ok := frontmatter["description"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid description in frontmatter")
	}

	// Create subagent config from Claude Code agent
	config := &SubagentConfig{
		Name:         name,
		Description:  description,
		SystemPrompt: systemPrompt,
		Tools:        sm.extractToolsFromPrompt(systemPrompt),
		Triggers:     sm.extractTriggersFromName(name),
		ContextLimit: sm.getContextLimitForAgent(name),
		CostOptimization: "high",
	}

	return config, nil
}

// parseFrontmatter parses YAML frontmatter from markdown content
func (sm *SubagentManager) parseFrontmatter(content string) (map[string]interface{}, string, error) {
	// Check if content starts with frontmatter
	if !strings.HasPrefix(content, "---\n") {
		return nil, "", fmt.Errorf("no frontmatter found")
	}

	// Find the end of frontmatter
	parts := strings.SplitN(content[4:], "\n---\n", 2)
	if len(parts) != 2 {
		return nil, "", fmt.Errorf("invalid frontmatter format")
	}

	yamlContent := parts[0]
	markdownContent := strings.TrimSpace(parts[1])

	// Parse YAML frontmatter
	var frontmatter map[string]interface{}
	if err := yaml.Unmarshal([]byte(yamlContent), &frontmatter); err != nil {
		return nil, "", fmt.Errorf("failed to parse YAML frontmatter: %w", err)
	}

	return frontmatter, markdownContent, nil
}

// extractToolsFromPrompt extracts tool requirements from system prompt
func (sm *SubagentManager) extractToolsFromPrompt(prompt string) []string {
	// Default tools based on agent type analysis
	tools := []string{"Read", "Write"}

	// Add tools based on prompt content
	if strings.Contains(prompt, "mem0") {
		tools = append(tools, "mcp__mem0__search_coding_preferences", "mcp__mem0__add_coding_preference")
	}
	if strings.Contains(prompt, "context7") {
		tools = append(tools, "mcp__context7__resolve-library-id", "mcp__context7__get-library-docs")
	}
	if strings.Contains(prompt, "code review") || strings.Contains(prompt, "Grep") {
		tools = append(tools, "Grep", "Glob")
	}
	if strings.Contains(prompt, "Edit") {
		tools = append(tools, "Edit")
	}

	return tools
}

// extractTriggersFromName extracts trigger patterns based on agent name
func (sm *SubagentManager) extractTriggersFromName(name string) TriggerConfig {
	patterns := []string{}

	switch name {
	case "claude-wm-templates":
		patterns = []string{"templates/", "ARCHITECTURE.md", "PRD.md", "TECHNICAL.md", "generate"}
	case "claude-wm-status":
		patterns = []string{"dashboard.md", "status", "debug/", "learning/", "metrics/"}
	case "claude-wm-planner":
		patterns = []string{"Plan-Task", "decompose", "planning", "1-Plan-"}
	case "claude-wm-reviewer":
		patterns = []string{"Review-Task", "validate", "Architecture-Review", "review"}
	}

	return TriggerConfig{Patterns: patterns}
}

// getContextLimitForAgent returns appropriate context limit for agent type
func (sm *SubagentManager) getContextLimitForAgent(name string) int {
	limits := map[string]int{
		"claude-wm-templates": 8000,
		"claude-wm-status":    5000,
		"claude-wm-planner":   15000,
		"claude-wm-reviewer":  25000,
	}

	if limit, exists := limits[name]; exists {
		return limit
	}

	return 10000 // default
}

// GetSubagent retrieves a subagent configuration by name
func (sm *SubagentManager) GetSubagent(name string) (*SubagentConfig, error) {
	config, exists := sm.subagents[name]
	if !exists {
		return nil, fmt.Errorf("subagent %s not found", name)
	}
	return config, nil
}

// ListSubagents returns all available subagent names
func (sm *SubagentManager) ListSubagents() []string {
	names := make([]string, 0, len(sm.subagents))
	for name := range sm.subagents {
		names = append(names, name)
	}
	return names
}

// MatchSubagent finds the best subagent for a given command path or task
func (sm *SubagentManager) MatchSubagent(commandPath string) (*SubagentConfig, float64) {
	bestMatch := (*SubagentConfig)(nil)
	bestScore := 0.0

	commandLower := strings.ToLower(commandPath)

	for _, config := range sm.subagents {
		score := sm.calculateMatchScore(commandLower, config)
		if score > bestScore {
			bestScore = score
			bestMatch = config
		}
	}

	return bestMatch, bestScore
}

// calculateMatchScore calculates how well a subagent matches a command
func (sm *SubagentManager) calculateMatchScore(commandPath string, config *SubagentConfig) float64 {
	score := 0.0
	maxPatterns := float64(len(config.Triggers.Patterns))

	if maxPatterns == 0 {
		return 0.0
	}

	// Give bonus points for multiple pattern matches
	for _, pattern := range config.Triggers.Patterns {
		if strings.Contains(commandPath, strings.ToLower(pattern)) {
			// Base score for match
			score += 1.0
			
			// Bonus for exact filename matches
			if strings.HasSuffix(commandPath, strings.ToLower(pattern)) {
				score += 0.5
			}
			
			// Bonus for directory path matches
			if strings.Contains(commandPath, strings.ToLower(pattern)+"/") {
				score += 0.3
			}
		}
	}

	// Normalize score but allow scores > 1.0 for strong matches
	normalizedScore := score / maxPatterns
	
	// Cap at 1.0 but reward multiple strong matches
	if normalizedScore > 1.0 {
		return 1.0
	}
	
	return normalizedScore
}

// GetOptimalSubagentForTask returns the best subagent for a specific task type
func (sm *SubagentManager) GetOptimalSubagentForTask(taskType, commandPath string) (*SubagentConfig, string) {
	// Direct mapping based on task characteristics
	taskMappings := map[string]string{
		"template":     "claude-wm-templates",
		"status":       "claude-wm-status", 
		"dashboard":    "claude-wm-status",
		"debug":        "claude-wm-status",
		"metrics":      "claude-wm-status",
		"plan":         "claude-wm-planner",
		"decompose":    "claude-wm-planner",
		"estimate":     "claude-wm-planner",
		"review":       "claude-wm-reviewer",
		"validate":     "claude-wm-reviewer",
		"architecture": "claude-wm-reviewer",
	}

	// Check for direct task type mapping
	if subagentName, exists := taskMappings[strings.ToLower(taskType)]; exists {
		if config, exists := sm.subagents[subagentName]; exists {
			return config, fmt.Sprintf("direct_mapping_%s", taskType)
		}
	}

	// Fallback to pattern matching
	subagent, score := sm.MatchSubagent(commandPath)
	if score > 0.5 { // Confidence threshold
		return subagent, fmt.Sprintf("pattern_match_%.2f", score)
	}

	return nil, "no_match"
}