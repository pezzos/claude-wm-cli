package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type TemplateContext struct {
	ProjectName   string `json:"project_name"`
	EpicName      string `json:"epic_name,omitempty"`
	StoryName     string `json:"story_name,omitempty"`
	TaskName      string `json:"task_name,omitempty"`
	IterationNum  int    `json:"iteration_num"`
	Timestamp     string `json:"timestamp"`
	Author        string `json:"author,omitempty"`
}

type TemplateConfig struct {
	Command      string            `json:"command"`
	TemplatePath string            `json:"template_path"`
	OutputPath   string            `json:"output_path"`
	Variables    map[string]string `json:"variables"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <user_prompt>\n", os.Args[0])
		os.Exit(1)
	}

	userPrompt := strings.Join(os.Args[1:], " ")
	
	// Detect template command
	templateConfig := detectTemplateCommand(userPrompt)
	if templateConfig == nil {
		// Not a template command we handle
		os.Exit(0)
	}

	// Load template context
	context, err := loadTemplateContext()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading template context: %v\n", err)
		os.Exit(1)
	}

	// Render template
	err = renderTemplate(templateConfig, context)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error rendering template: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Template rendered: %s -> %s\n", templateConfig.TemplatePath, templateConfig.OutputPath)
}

func detectTemplateCommand(prompt string) *TemplateConfig {
	patterns := map[string]*TemplateConfig{
		`2-epic:1-start:2-Plan-stories`: {
			Command:      "plan-stories",
			TemplatePath: "commands/templates/STORIES.md",
			OutputPath:   "docs/2-current-epic/STORIES.md",
			Variables:    make(map[string]string),
		},
		`4-task:2-execute:2-Test-design`: {
			Command:      "test-design",
			TemplatePath: "commands/templates/TEST.md", 
			OutputPath:   "docs/3-current-task/TEST.md",
			Variables:    make(map[string]string),
		},
		`4-task:2-execute:1-Plan-Task`: {
			Command:      "plan-task",
			TemplatePath: "commands/templates/TASK.md",
			OutputPath:   "docs/3-current-task/TASK.md", 
			Variables:    make(map[string]string),
		},
		`1-project:1-start:1-Init-Project`: {
			Command:      "init-project",
			TemplatePath: "commands/templates/README.md",
			OutputPath:   "README.md",
			Variables:    make(map[string]string),
		},
	}

	for pattern, config := range patterns {
		matched, _ := regexp.MatchString(pattern, prompt)
		if matched {
			return config
		}
	}

	return nil
}

func loadTemplateContext() (*TemplateContext, error) {
	context := &TemplateContext{
		Timestamp:    time.Now().Format("2006-01-02 15:04:05"),
		IterationNum: 1,
	}

	// Try to load from hook state
	if stateData, err := os.ReadFile("hooks/state.json"); err == nil {
		var state map[string]interface{}
		if json.Unmarshal(stateData, &state) == nil {
			if projectName, ok := state["project_name"].(string); ok {
				context.ProjectName = projectName
			}
			if epicName, ok := state["epic_name"].(string); ok {
				context.EpicName = epicName
			}
			if iterationNum, ok := state["iteration_num"].(float64); ok {
				context.IterationNum = int(iterationNum)
			}
		}
	}

	// Try to infer project name from directory or README
	if context.ProjectName == "" {
		if cwd, err := os.Getwd(); err == nil {
			context.ProjectName = filepath.Base(cwd)
		}
	}

	// Try to load epic name from PRD
	if prdData, err := os.ReadFile("docs/2-current-epic/PRD.md"); err == nil {
		lines := strings.Split(string(prdData), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "# ") {
				context.EpicName = strings.TrimPrefix(line, "# ")
				break
			}
		}
	}

	// Try to load story name from STORIES.md  
	if storiesData, err := os.ReadFile("docs/2-current-epic/STORIES.md"); err == nil {
		// Find current story (first uncompleted one)
		lines := strings.Split(string(storiesData), "\n")
		for _, line := range lines {
			if strings.Contains(line, "- [ ] Story") {
				// Extract story name
				parts := strings.Split(line, ":")
				if len(parts) > 1 {
					context.StoryName = strings.TrimSpace(parts[1])
					break
				}
			}
		}
	}

	// Try to get author from git
	if gitUser, err := getGitUser(); err == nil {
		context.Author = gitUser
	}

	return context, nil
}

func renderTemplate(config *TemplateConfig, context *TemplateContext) error {
	// Check if template exists
	if _, err := os.Stat(config.TemplatePath); os.IsNotExist(err) {
		return createDefaultTemplate(config, context)
	}

	// Read template
	templateData, err := os.ReadFile(config.TemplatePath)
	if err != nil {
		return fmt.Errorf("failed to read template %s: %v", config.TemplatePath, err)
	}

	// Render template with variables
	rendered := renderVariables(string(templateData), context, config.Variables)

	// Ensure output directory exists
	outputDir := filepath.Dir(config.OutputPath)
	err = os.MkdirAll(outputDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create output directory %s: %v", outputDir, err)
	}

	// Write rendered template
	err = os.WriteFile(config.OutputPath, []byte(rendered), 0644)
	if err != nil {
		return fmt.Errorf("failed to write output file %s: %v", config.OutputPath, err)
	}

	return nil
}

func renderVariables(template string, context *TemplateContext, extraVars map[string]string) string {
	// Define variable replacements
	vars := map[string]string{
		"{{PROJECT_NAME}}":   context.ProjectName,
		"{{EPIC_NAME}}":      context.EpicName,  
		"{{STORY_NAME}}":     context.StoryName,
		"{{TASK_NAME}}":      context.TaskName,
		"{{ITERATION_NUM}}":  fmt.Sprintf("%d", context.IterationNum),
		"{{TIMESTAMP}}":      context.Timestamp,
		"{{AUTHOR}}":         context.Author,
		"{{DATE}}":          time.Now().Format("2006-01-02"),
		"{{DATETIME}}":      time.Now().Format("2006-01-02 15:04:05"),
	}

	// Add extra variables
	for key, value := range extraVars {
		vars[key] = value
	}

	// Replace variables
	result := template
	for varName, varValue := range vars {
		result = strings.ReplaceAll(result, varName, varValue)
	}

	return result
}

func createDefaultTemplate(config *TemplateConfig, context *TemplateContext) error {
	var defaultContent string

	switch config.Command {
	case "plan-stories":
		defaultContent = `# Epic Stories: {{EPIC_NAME}}

## Story Breakdown

### High Priority Stories (P0)
- [ ] Story 1: Core functionality
  - **Acceptance Criteria**: Basic functionality works
  - **Complexity**: 5 points
  - **Dependencies**: None

### Medium Priority Stories (P1)
- [ ] Story 2: Enhanced features
  - **Acceptance Criteria**: Enhanced functionality works
  - **Complexity**: 3 points  
  - **Dependencies**: Story 1

### Low Priority Stories (P2)
- [ ] Story 3: Nice-to-have features
  - **Acceptance Criteria**: Additional functionality works
  - **Complexity**: 2 points
  - **Dependencies**: Story 2

## Completed Stories
<!-- Completed stories will be moved here -->

---
*Created: {{TIMESTAMP}} by {{AUTHOR}}*
*Project: {{PROJECT_NAME}}*
`

	case "test-design":
		defaultContent = `# Test Design: {{TASK_NAME}}

## Test Strategy

### Automated Tests
- [ ] Unit tests for core functions
- [ ] Integration tests for API endpoints
- [ ] UI tests with Playwright/Puppeteer (if applicable)

### Manual Tests
- [ ] User journey validation
- [ ] Edge case testing
- [ ] Cross-browser testing (if web UI)

### Performance Tests
- [ ] Load testing
- [ ] Response time validation
- [ ] Memory usage validation

### Security Tests
- [ ] Input validation testing
- [ ] Authentication testing
- [ ] Authorization testing

## Test Data
<!-- Define test data requirements -->

## Success Criteria
- [ ] All automated tests pass
- [ ] Manual test scenarios validated
- [ ] Performance benchmarks met
- [ ] Security requirements satisfied

---
*Created: {{TIMESTAMP}} for iteration {{ITERATION_NUM}}*
`

	case "plan-task":
		defaultContent = `# Task Implementation: {{TASK_NAME}}

## Overview
**Epic**: {{EPIC_NAME}}
**Story**: {{STORY_NAME}}
**Iteration**: {{ITERATION_NUM}}/3

## Implementation Approach
1. **Analysis Phase**
   - [ ] Requirements analysis
   - [ ] Technical investigation
   - [ ] Architecture planning

2. **Implementation Phase**  
   - [ ] Core functionality
   - [ ] Integration points
   - [ ] Error handling

3. **Testing Phase**
   - [ ] Unit tests
   - [ ] Integration tests
   - [ ] Manual validation

4. **Documentation Phase**
   - [ ] Code documentation
   - [ ] User documentation
   - [ ] Implementation notes

## File Changes
<!-- List files that will be modified or created -->

## Risks & Assumptions
<!-- Document potential risks and assumptions -->

## Definition of Done
- [ ] All acceptance criteria met
- [ ] Tests pass
- [ ] Code reviewed
- [ ] Documentation updated

---
*Created: {{TIMESTAMP}} by {{AUTHOR}}*
*Project: {{PROJECT_NAME}}*
`

	case "init-project":
		defaultContent = `# {{PROJECT_NAME}}

## Overview
Brief description of the project.

## Features
- Feature 1
- Feature 2
- Feature 3

## Getting Started

### Prerequisites
- Requirement 1
- Requirement 2

### Installation
\`\`\`bash
# Installation commands
\`\`\`

### Usage
\`\`\`bash
# Usage examples
\`\`\`

## Development
Describe development setup and workflow.

## Contributing
Guidelines for contributing to the project.

## License
License information.

---
*Created: {{TIMESTAMP}} by {{AUTHOR}}*
`

	default:
		defaultContent = `# {{PROJECT_NAME}}

## Content
Default content for {{TASK_NAME}}.

---
*Created: {{TIMESTAMP}} by {{AUTHOR}}*
`
	}

	// Render and write default template
	rendered := renderVariables(defaultContent, context, config.Variables)
	
	// Ensure template directory exists
	templateDir := filepath.Dir(config.TemplatePath)
	err := os.MkdirAll(templateDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create template directory %s: %v", templateDir, err)
	}

	// Create template file for future use
	err = os.WriteFile(config.TemplatePath, []byte(defaultContent), 0644)
	if err != nil {
		fmt.Printf("Warning: failed to create template file %s: %v\n", config.TemplatePath, err)
	}

	// Ensure output directory exists
	outputDir := filepath.Dir(config.OutputPath)
	err = os.MkdirAll(outputDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create output directory %s: %v", outputDir, err)
	}

	// Write rendered output
	err = os.WriteFile(config.OutputPath, []byte(rendered), 0644)
	if err != nil {
		return fmt.Errorf("failed to write output file %s: %v", config.OutputPath, err)
	}

	return nil
}

func getGitUser() (string, error) {
	// Try to get git user name
	if gitName, err := os.LookupEnv("GIT_AUTHOR_NAME"); err == false && gitName != "" {
		return gitName, nil
	}
	
	// Try reading from git config (simplified)
	if homeDir, err := os.UserHomeDir(); err == nil {
		gitConfigPath := filepath.Join(homeDir, ".gitconfig")
		if data, err := os.ReadFile(gitConfigPath); err == nil {
			lines := strings.Split(string(data), "\n")
			for _, line := range lines {
				if strings.Contains(line, "name =") {
					parts := strings.Split(line, "=")
					if len(parts) > 1 {
						return strings.TrimSpace(parts[1]), nil
					}
				}
			}
		}
	}
	
	return "Developer", nil
}