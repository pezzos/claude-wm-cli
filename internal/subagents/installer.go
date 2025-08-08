package subagents

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
)

//go:embed agents/*.md
var embeddedAgents embed.FS

// AgentInstaller handles installation of subagents to target projects
type AgentInstaller struct {
	sourceAgentsPath string // Where to find the source agents (for development)
}

// NewAgentInstaller creates a new agent installer
func NewAgentInstaller() *AgentInstaller {
	return &AgentInstaller{}
}

// InstallAgents installs all claude-wm agents to the target project's .claude/agents directory
func (ai *AgentInstaller) InstallAgents(projectPath string) error {
	targetDir := filepath.Join(projectPath, ".claude", "agents")
	
	// Ensure target directory exists
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create agents directory: %w", err)
	}

	// Try to install from embedded agents first
	if err := ai.installEmbeddedAgents(targetDir); err != nil {
		// Fallback to copying from source directory if embedded fails
		return ai.installFromSource(targetDir)
	}

	return nil
}

// installEmbeddedAgents installs agents from embedded filesystem
func (ai *AgentInstaller) installEmbeddedAgents(targetDir string) error {
	agentFiles, err := embeddedAgents.ReadDir("agents")
	if err != nil {
		return fmt.Errorf("failed to read embedded agents: %w", err)
	}

	installedCount := 0
	for _, file := range agentFiles {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".md" {
			// Read embedded file
			content, err := embeddedAgents.ReadFile(filepath.Join("agents", file.Name()))
			if err != nil {
				continue // Skip this file and try others
			}

			// Write to target directory
			targetFile := filepath.Join(targetDir, file.Name())
			if err := os.WriteFile(targetFile, content, 0644); err != nil {
				continue // Skip this file and try others
			}

			installedCount++
		}
	}

	if installedCount == 0 {
		return fmt.Errorf("no agents were installed from embedded filesystem")
	}

	return nil
}

// installFromSource installs agents by copying from a source directory (fallback)
func (ai *AgentInstaller) installFromSource(targetDir string) error {
	// Try to find agents in common locations
	possibleSources := []string{
		".claude/agents",           // Current project
		"../claude-wm-cli/.claude/agents", // Relative to current project
		filepath.Join(os.Getenv("HOME"), ".claude-wm-cli", "agents"), // User home
	}

	var sourceDir string
	for _, source := range possibleSources {
		if _, err := os.Stat(source); err == nil {
			sourceDir = source
			break
		}
	}

	if sourceDir == "" {
		return fmt.Errorf("could not find source agents directory")
	}

	return ai.copyAgentsFromDirectory(sourceDir, targetDir)
}

// copyAgentsFromDirectory copies agent files from source to target directory
func (ai *AgentInstaller) copyAgentsFromDirectory(sourceDir, targetDir string) error {
	files, err := filepath.Glob(filepath.Join(sourceDir, "claude-wm-*.md"))
	if err != nil {
		return fmt.Errorf("failed to find agent files: %w", err)
	}

	if len(files) == 0 {
		return fmt.Errorf("no claude-wm agent files found in %s", sourceDir)
	}

	for _, sourceFile := range files {
		content, err := os.ReadFile(sourceFile)
		if err != nil {
			return fmt.Errorf("failed to read agent file %s: %w", sourceFile, err)
		}

		targetFile := filepath.Join(targetDir, filepath.Base(sourceFile))
		if err := os.WriteFile(targetFile, content, 0644); err != nil {
			return fmt.Errorf("failed to write agent file %s: %w", targetFile, err)
		}
	}

	return nil
}

// ListInstalledAgents returns a list of installed agents in a project
func (ai *AgentInstaller) ListInstalledAgents(projectPath string) ([]string, error) {
	agentsDir := filepath.Join(projectPath, ".claude", "agents")
	
	files, err := filepath.Glob(filepath.Join(agentsDir, "claude-wm-*.md"))
	if err != nil {
		return nil, fmt.Errorf("failed to list agent files: %w", err)
	}

	agents := make([]string, 0, len(files))
	for _, file := range files {
		agents = append(agents, filepath.Base(file))
	}

	return agents, nil
}

// VerifyAgentsInstalled checks if all required agents are installed
func (ai *AgentInstaller) VerifyAgentsInstalled(projectPath string) (bool, []string, error) {
	requiredAgents := []string{
		"claude-wm-templates.md",
		"claude-wm-status.md",
		"claude-wm-planner.md", 
		"claude-wm-reviewer.md",
	}

	installed, err := ai.ListInstalledAgents(projectPath)
	if err != nil {
		return false, nil, err
	}

	installedMap := make(map[string]bool)
	for _, agent := range installed {
		installedMap[agent] = true
	}

	missing := make([]string, 0)
	for _, required := range requiredAgents {
		if !installedMap[required] {
			missing = append(missing, required)
		}
	}

	return len(missing) == 0, missing, nil
}

// GetAgentInstallationInfo returns information about agent installation
func (ai *AgentInstaller) GetAgentInstallationInfo(projectPath string) (*InstallationInfo, error) {
	allInstalled, missing, err := ai.VerifyAgentsInstalled(projectPath)
	if err != nil {
		return nil, err
	}

	installed, err := ai.ListInstalledAgents(projectPath)
	if err != nil {
		return nil, err
	}

	return &InstallationInfo{
		ProjectPath:      projectPath,
		AllInstalled:     allInstalled,
		InstalledAgents:  installed,
		MissingAgents:    missing,
		AgentsDirectory:  filepath.Join(projectPath, ".claude", "agents"),
	}, nil
}

// InstallationInfo contains information about agent installation status
type InstallationInfo struct {
	ProjectPath      string   `json:"project_path"`
	AllInstalled     bool     `json:"all_installed"`
	InstalledAgents  []string `json:"installed_agents"`
	MissingAgents    []string `json:"missing_agents"`
	AgentsDirectory  string   `json:"agents_directory"`
}

// GetInstallationSummary returns a human-readable summary of installation status
func (info *InstallationInfo) GetInstallationSummary() string {
	if info.AllInstalled {
		return fmt.Sprintf("✅ All %d claude-wm agents are installed in %s", 
			len(info.InstalledAgents), info.AgentsDirectory)
	}

	return fmt.Sprintf("⚠️  %d/%d agents installed. Missing: %v", 
		len(info.InstalledAgents), 
		len(info.InstalledAgents)+len(info.MissingAgents),
		info.MissingAgents)
}