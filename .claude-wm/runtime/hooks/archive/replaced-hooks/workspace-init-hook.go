package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type Config struct {
	ProjectRoot string `json:"project_root"`
	CurrentEpic string `json:"current_epic,omitempty"`
	CurrentTask string `json:"current_task,omitempty"`
}

type HookState struct {
	LastCommand  string    `json:"last_command"`
	Timestamp    time.Time `json:"timestamp"`
	ProjectName  string    `json:"project_name"`
	EpicName     string    `json:"epic_name,omitempty"`
	IterationNum int       `json:"iteration_num"`
}

func main() {
	// Parse command line arguments from Claude Code hook
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <user_prompt>\n", os.Args[0])
		os.Exit(1)
	}

	userPrompt := strings.Join(os.Args[1:], " ")

	// Detect command type from user prompt
	commandType := detectCommandType(userPrompt)
	if commandType == "" {
		// Not a command we handle, exit silently
		os.Exit(0)
	}

	// Load or create hook state
	state, err := loadHookState()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading hook state: %v\n", err)
		os.Exit(1)
	}

	// Execute initialization based on command type
	err = initializeWorkspace(commandType, state)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing workspace: %v\n", err)
		os.Exit(1)
	}

	// Save updated state
	err = saveHookState(state)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error saving hook state: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Workspace initialized for %s\n", commandType)
}

func detectCommandType(prompt string) string {
	patterns := map[string]string{
		`1-project:3-epics:1-Plan-Epics`:   "plan-epics",
		`4-task:2-execute:1-Plan-Task`:     "plan-task",
		`2-epic:1-start:1-Select-Stories`:  "select-epic",
		`3-story:1-manage:1-Start-Story`:   "start-story",
		`4-task:1-start:1-From-story`:      "create-task",
		`1-project:1-start:1-Init-Project`: "init-project",
	}

	for pattern, cmdType := range patterns {
		matched, _ := regexp.MatchString(pattern, prompt)
		if matched {
			return cmdType
		}
	}

	return ""
}

func initializeWorkspace(commandType string, state *HookState) error {
	state.LastCommand = commandType
	state.Timestamp = time.Now()

	switch commandType {
	case "plan-epics":
		return initializePlanEpics(state)
	case "plan-task":
		return initializePlanTask(state)
	case "select-epic":
		return initializeSelectEpic(state)
	case "start-story":
		return initializeStartStory(state)
	case "create-task":
		return initializeCreateTask(state)
	case "init-project":
		return initializeInitProject(state)
	default:
		return fmt.Errorf("unknown command type: %s", commandType)
	}
}

func initializePlanEpics(state *HookState) error {
	// Ensure project structure exists
	dirs := []string{
		"docs/1-project",
		"docs/2-current-epic",
		"docs/3-current-task",
		"docs/archive/epics",
		"docs/archive/stories",
		"docs/archive/tasks",
	}

	for _, dir := range dirs {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return fmt.Errorf("failed to create directory %s: %v", dir, err)
		}
	}

	// Initialize EPICS.md if it doesn't exist
	epicsPath := "docs/1-project/EPICS.md"
	if _, err := os.Stat(epicsPath); os.IsNotExist(err) {
		epicsTemplate := `# Project Epics

## Epic Backlog

### High Priority (P0)
- [ ] Epic 1: Description

### Medium Priority (P1) 
- [ ] Epic 2: Description

### Low Priority (P2)
- [ ] Epic 3: Description

## Completed Epics
<!-- Completed epics will be listed here -->

---
*Last updated: ` + time.Now().Format("2006-01-02 15:04:05") + `*
`
		err := os.WriteFile(epicsPath, []byte(epicsTemplate), 0644)
		if err != nil {
			return fmt.Errorf("failed to create EPICS.md: %v", err)
		}
	}

	return nil
}

func initializePlanTask(state *HookState) error {
	// Clear current task directory except CLAUDE.md
	taskDir := "docs/3-current-task"

	entries, err := os.ReadDir(taskDir)
	if err != nil {
		// Directory doesn't exist, create it
		err = os.MkdirAll(taskDir, 0755)
		if err != nil {
			return fmt.Errorf("failed to create task directory: %v", err)
		}
	} else {
		// Clean existing files except CLAUDE.md
		for _, entry := range entries {
			if entry.Name() == "CLAUDE.md" {
				continue
			}
			filePath := filepath.Join(taskDir, entry.Name())
			err = os.Remove(filePath)
			if err != nil {
				fmt.Printf("Warning: failed to remove %s: %v\n", filePath, err)
			}
		}
	}

	// Initialize ITERATIONS.md
	iterationsPath := filepath.Join(taskDir, "ITERATIONS.md")
	iterationsContent := `# Implementation Iterations

## Iteration 1/3

**Status**: 🚧 In Progress
**Started**: ` + time.Now().Format("2006-01-02 15:04:05") + `

### Goals
- [ ] Complete task planning
- [ ] Implement core functionality
- [ ] Pass all tests

### Progress
<!-- Track progress here -->

---
*Maximum 3 iterations allowed*
`
	err = os.WriteFile(iterationsPath, []byte(iterationsContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to create ITERATIONS.md: %v", err)
	}

	state.IterationNum = 1
	return nil
}

func initializeSelectEpic(state *HookState) error {
	// Clear current epic directory
	epicDir := "docs/2-current-epic"

	err := os.RemoveAll(epicDir)
	if err != nil {
		return fmt.Errorf("failed to clear epic directory: %v", err)
	}

	err = os.MkdirAll(epicDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to recreate epic directory: %v", err)
	}

	return nil
}

func initializeStartStory(state *HookState) error {
	// Ensure TODO.md exists in current epic
	todoPath := "docs/2-current-epic/TODO.md"

	if _, err := os.Stat(todoPath); os.IsNotExist(err) {
		todoTemplate := `# Story Tasks

## Current Story: [Story Name]

### Tasks
- [ ] Task 1: Description
- [ ] Task 2: Description  
- [ ] Task 3: Description

### Acceptance Criteria
- [ ] Criteria 1
- [ ] Criteria 2

### Notes
<!-- Implementation notes and decisions -->

---
*Created: ` + time.Now().Format("2006-01-02 15:04:05") + `*
`
		err := os.WriteFile(todoPath, []byte(todoTemplate), 0644)
		if err != nil {
			return fmt.Errorf("failed to create TODO.md: %v", err)
		}
	}

	return nil
}

func initializeCreateTask(state *HookState) error {
	// Ensure task directory is ready
	taskDir := "docs/3-current-task"
	err := os.MkdirAll(taskDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create task directory: %v", err)
	}

	return nil
}

func initializeInitProject(state *HookState) error {
	// Create the complete project structure for Init-Project
	dirs := []string{
		"docs/1-project",
		"docs/2-current-epic",
		"docs/3-current-task",
		"docs/archive/epics",
		"docs/archive/stories",
		"docs/archive/tasks",
	}

	for _, dir := range dirs {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return fmt.Errorf("failed to create directory %s: %v", dir, err)
		}
	}

	// Initialize FEEDBACK.md if it doesn't exist
	feedbackPath := "docs/1-project/FEEDBACK.md"
	if _, err := os.Stat(feedbackPath); os.IsNotExist(err) {
		feedbackTemplate := `# Feedback - ` + time.Now().Format("2006-01-02") + `

## Questions from Review
### Architecture
- Q: {Question about architecture}
  A: {User response}

### Technical Choices
- Q: {Question about tech stack}
  A: {User response}

### Scope & Requirements
- Q: {Clarification needed}
  A: {User response}

## New Information
### Features
- {New feature requirement}
- {Changed requirement}

### Constraints
- {New technical constraint}
- {Business constraint}

## Decisions Made
- ✅ {Decision taken based on feedback}
- ❌ {What was rejected and why}
- 🔄 {What needs rework}

## Next Actions
- [ ] Update ARCHITECTURE.md with {specific change}
- [ ] Update README.md with {specific change}
- [ ] Revise epic #{id} for {reason}
`
		err := os.WriteFile(feedbackPath, []byte(feedbackTemplate), 0644)
		if err != nil {
			return fmt.Errorf("failed to create FEEDBACK.md: %v", err)
		}
	}

	// Initialize git if not exists
	if _, err := os.Stat(".git"); os.IsNotExist(err) {
		// Git will be initialized by the actual command
		fmt.Println("ℹ Git repository will be initialized by Init-Project command")
	}

	state.ProjectName = "project-initialized"
	return nil
}

func loadHookState() (*HookState, error) {
	statePath := "hooks/state.json"

	if _, err := os.Stat(statePath); os.IsNotExist(err) {
		// Create default state
		return &HookState{
			ProjectName:  "unknown",
			IterationNum: 1,
		}, nil
	}

	data, err := os.ReadFile(statePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read state file: %v", err)
	}

	var state HookState
	err = json.Unmarshal(data, &state)
	if err != nil {
		return nil, fmt.Errorf("failed to parse state file: %v", err)
	}

	return &state, nil
}

func saveHookState(state *HookState) error {
	// Ensure hooks directory exists
	err := os.MkdirAll("hooks", 0755)
	if err != nil {
		return fmt.Errorf("failed to create hooks directory: %v", err)
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state: %v", err)
	}

	statePath := "hooks/state.json"
	err = os.WriteFile(statePath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write state file: %v", err)
	}

	return nil
}
