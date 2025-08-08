package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// SerenaConfig represents the configuration for Serena MCP integration
type SerenaConfig struct {
	Enabled        bool              `yaml:"enabled"`
	MCPServerPath  string            `yaml:"mcp_server_path"`
	ServerArgs     []string          `yaml:"server_args"`
	Timeout        int               `yaml:"timeout_seconds"`
	ContextLimits  map[string]int    `yaml:"context_limits"`
	AnalysisTypes  map[string]string `yaml:"analysis_types"`
	FallbackEnabled bool             `yaml:"fallback_enabled"`
	AutoDetect     bool              `yaml:"auto_detect"`
}

// DefaultSerenaConfig returns the default Serena configuration
func DefaultSerenaConfig() *SerenaConfig {
	return &SerenaConfig{
		Enabled:        false, // Disabled by default until properly configured
		MCPServerPath:  "serena-mcp-server", // Assumes it's in PATH
		ServerArgs:     []string{},
		Timeout:        30,
		ContextLimits: map[string]int{
			"code_review":        3000,
			"template_generation": 2000,
			"status_reporting":   1500,
			"planning":           8000,
			"general":            5000,
		},
		AnalysisTypes: map[string]string{
			"review":     "code_review",
			"template":   "template_generation",
			"status":     "status_reporting",
			"dashboard":  "status_reporting",
			"plan":       "planning",
			"decompose":  "planning",
		},
		FallbackEnabled: true,
		AutoDetect:      true,
	}
}

// LoadSerenaConfig loads Serena configuration from file or creates default
func LoadSerenaConfig(configDir string) (*SerenaConfig, error) {
	configPath := filepath.Join(configDir, "serena.yaml")
	
	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Create default config file
		defaultConfig := DefaultSerenaConfig()
		if err := SaveSerenaConfig(configPath, defaultConfig); err != nil {
			return nil, fmt.Errorf("failed to create default serena config: %w", err)
		}
		return defaultConfig, nil
	}
	
	// Load existing config
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read serena config: %w", err)
	}
	
	var config SerenaConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse serena config: %w", err)
	}
	
	// Merge with defaults for missing values
	defaultConfig := DefaultSerenaConfig()
	if config.ContextLimits == nil {
		config.ContextLimits = defaultConfig.ContextLimits
	}
	if config.AnalysisTypes == nil {
		config.AnalysisTypes = defaultConfig.AnalysisTypes
	}
	if config.MCPServerPath == "" {
		config.MCPServerPath = defaultConfig.MCPServerPath
	}
	if config.Timeout == 0 {
		config.Timeout = defaultConfig.Timeout
	}
	
	return &config, nil
}

// SaveSerenaConfig saves Serena configuration to file
func SaveSerenaConfig(configPath string, config *SerenaConfig) error {
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}
	
	// Marshal config to YAML
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	
	// Write to file
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}
	
	return nil
}

// ValidateSerenaConfig validates the Serena configuration
func (sc *SerenaConfig) ValidateSerenaConfig() error {
	if !sc.Enabled {
		return nil // No validation needed if disabled
	}
	
	// Check if MCP server path is accessible
	if sc.MCPServerPath == "" {
		return fmt.Errorf("mcp_server_path cannot be empty when enabled")
	}
	
	// Check timeout is reasonable
	if sc.Timeout < 5 || sc.Timeout > 300 {
		return fmt.Errorf("timeout must be between 5 and 300 seconds")
	}
	
	// Validate context limits
	for analysisType, limit := range sc.ContextLimits {
		if limit < 100 || limit > 50000 {
			return fmt.Errorf("context limit for %s must be between 100 and 50000", analysisType)
		}
	}
	
	return nil
}

// GetContextLimitForType returns the context limit for a specific analysis type
func (sc *SerenaConfig) GetContextLimitForType(analysisType string) int {
	if limit, exists := sc.ContextLimits[analysisType]; exists {
		return limit
	}
	
	// Default fallback
	return sc.ContextLimits["general"]
}

// GetAnalysisTypeForCommand maps a command pattern to an analysis type
func (sc *SerenaConfig) GetAnalysisTypeForCommand(commandPath string) string {
	for pattern, analysisType := range sc.AnalysisTypes {
		if contains(commandPath, pattern) {
			return analysisType
		}
	}
	
	return "general"
}

// contains is a case-insensitive string contains check
func contains(str, substr string) bool {
	str = strings.ToLower(str)
	substr = strings.ToLower(substr)
	return strings.Contains(str, substr)
}

// SerenaConfigManager manages Serena configuration lifecycle
type SerenaConfigManager struct {
	config     *SerenaConfig
	configPath string
}

// NewSerenaConfigManager creates a new Serena configuration manager
func NewSerenaConfigManager(configDir string) (*SerenaConfigManager, error) {
	config, err := LoadSerenaConfig(configDir)
	if err != nil {
		return nil, err
	}
	
	if err := config.ValidateSerenaConfig(); err != nil {
		return nil, fmt.Errorf("invalid serena configuration: %w", err)
	}
	
	return &SerenaConfigManager{
		config:     config,
		configPath: filepath.Join(configDir, "serena.yaml"),
	}, nil
}

// GetConfig returns the current Serena configuration
func (scm *SerenaConfigManager) GetConfig() *SerenaConfig {
	return scm.config
}

// UpdateConfig updates the Serena configuration
func (scm *SerenaConfigManager) UpdateConfig(newConfig *SerenaConfig) error {
	if err := newConfig.ValidateSerenaConfig(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}
	
	scm.config = newConfig
	return SaveSerenaConfig(scm.configPath, newConfig)
}

// EnableSerena enables Serena integration
func (scm *SerenaConfigManager) EnableSerena() error {
	scm.config.Enabled = true
	return SaveSerenaConfig(scm.configPath, scm.config)
}

// DisableSerena disables Serena integration
func (scm *SerenaConfigManager) DisableSerena() error {
	scm.config.Enabled = false
	return SaveSerenaConfig(scm.configPath, scm.config)
}

// IsSerenaAvailable checks if Serena MCP server is available
func (scm *SerenaConfigManager) IsSerenaAvailable() bool {
	if !scm.config.Enabled {
		return false
	}
	
	// Check if MCP server executable exists
	if scm.config.MCPServerPath != "" {
		// This is a simplified check - in reality, we'd want to ping the server
		return true
	}
	
	return false
}