package metrics

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// InstrumentCommand instruments a command with performance monitoring
func InstrumentCommand(commandName string) *Timer {
	collector := GetCollector()
	if !collector.enabled {
		return &Timer{} // Return dummy timer
	}
	
	return collector.StartCommand(commandName)
}

// InstrumentCommandInteractive instruments an interactive command with user context
func InstrumentCommandInteractive(commandName string) *Timer {
	collector := GetCollector()
	if !collector.enabled {
		return &Timer{}
	}
	
	timer := collector.StartCommand(commandName)
	
	// Add interactive context
	timer.SetContext("execution_mode", "interactive")
	timer.SetContext("terminal_width", getTerminalWidth())
	
	// Add project state context if available
	if projectContext := getProjectState(); projectContext != nil {
		timer.SetContext("project_state", projectContext)
	}
	
	return timer
}

// InstrumentStoryStart instruments the "Start Story" command with detailed profiling
func InstrumentStoryStart() *Timer {
	collector := GetCollector()
	if !collector.enabled {
		return &Timer{}
	}
	
	timer := collector.StartCommand("Start Story")
	
	// Add story-specific context
	timer.SetContext("execution_mode", "interactive")
	timer.SetContext("command_type", "story_management")
	timer.SetContext("workflow_phase", "story_start")
	
	return timer
}

// InstrumentClaudeCommand instruments Claude slash commands
func InstrumentClaudeCommand(slashCommand string) *Timer {
	collector := GetCollector()
	if !collector.enabled {
		return &Timer{}
	}
	
	timer := collector.StartCommand(fmt.Sprintf("Claude: %s", slashCommand))
	
	// Add Claude-specific context
	timer.SetContext("execution_mode", "claude_integration")
	timer.SetContext("slash_command", slashCommand)
	timer.SetContext("ai_workflow", true)
	
	return timer
}

// Common step names for consistent profiling across commands
const (
	StepJSONValidation      = "json_validation"
	StepFileDiscovery      = "file_discovery"
	StepParsingStories     = "parsing_stories"
	StepStorySelection     = "story_selection"
	StepClaudePreparation  = "claude_preparation"
	StepClaudeExecution    = "claude_execution"
	StepResponseProcessing = "response_processing"
	StepFileWrites         = "file_writes"
	StepGitOperations      = "git_operations"
	StepUserInput          = "user_input"
	StepValidation         = "validation"
	StepPreprocessing      = "preprocessing"
	StepPostprocessing     = "postprocessing"
	StepContextDetection   = "context_detection"
	StepMenuDisplay        = "menu_display"
	StepTemplateProcessing = "template_processing"
	StepConfigSync         = "config_sync"
)

// ProfileStep creates and starts a step timer with common metadata
func (t *Timer) ProfileStep(stepName string) *StepTimer {
	if t == nil || t.collector == nil || !t.collector.enabled {
		return &StepTimer{stepName: stepName, startTime: time.Now()} // Return dummy timer
	}
	
	step := t.StartStep(stepName)
	
	// Add common step metadata
	step.SetMetadata("pid", os.Getpid())
	step.SetMetadata("start_time", time.Now().Unix())
	
	return step
}

// ProfileJSONValidation profiles JSON validation steps with file details
func (t *Timer) ProfileJSONValidation(jsonType string) *StepTimer {
	step := t.ProfileStep(StepJSONValidation)
	if step != nil {
		step.SetMetadata("json_type", jsonType)
		step.SetMetadata("validation_target", jsonType)
	}
	return step
}

// ProfileFileDiscovery profiles file discovery operations
func (t *Timer) ProfileFileDiscovery(searchPattern string) *StepTimer {
	step := t.ProfileStep(StepFileDiscovery)
	if step != nil {
		step.SetMetadata("search_pattern", searchPattern)
		step.SetMetadata("working_directory", getWorkingDirectory())
	}
	return step
}

// ProfileStoryParsing profiles story parsing operations
func (t *Timer) ProfileStoryParsing(filePath string) *StepTimer {
	step := t.ProfileStep(StepParsingStories)
	if step != nil {
		step.SetMetadata("file_path", filePath)
		if fileInfo, err := os.Stat(filePath); err == nil {
			step.SetMetadata("file_size_bytes", fileInfo.Size())
		}
	}
	return step
}

// ProfileStorySelection profiles story selection logic
func (t *Timer) ProfileStorySelection(storyCount int) *StepTimer {
	step := t.ProfileStep(StepStorySelection)
	if step != nil {
		step.SetMetadata("available_stories", storyCount)
		step.SetMetadata("selection_mode", "highest_priority")
	}
	return step
}

// ProfileClaudePreparation profiles Claude command preparation
func (t *Timer) ProfileClaudePreparation(command string, contextSize int) *StepTimer {
	step := t.ProfileStep(StepClaudePreparation)
	if step != nil {
		step.SetMetadata("claude_command", command)
		step.SetMetadata("context_size_chars", contextSize)
		step.SetMetadata("preparation_type", "slash_command")
	}
	return step
}

// ProfileClaudeExecution profiles Claude command execution
func (t *Timer) ProfileClaudeExecution(command string) *StepTimer {
	step := t.ProfileStep(StepClaudeExecution)
	if step != nil {
		step.SetMetadata("claude_command", command)
		step.SetMetadata("execution_start", time.Now().Unix())
	}
	return step
}

// ProfileResponseProcessing profiles response processing
func (t *Timer) ProfileResponseProcessing(responseSize int) *StepTimer {
	step := t.ProfileStep(StepResponseProcessing)
	if step != nil {
		step.SetMetadata("response_size_chars", responseSize)
		step.SetMetadata("processing_type", "claude_response")
	}
	return step
}

// ProfileFileWrites profiles file write operations
func (t *Timer) ProfileFileWrites(fileCount int, totalSize int64) *StepTimer {
	step := t.ProfileStep(StepFileWrites)
	if step != nil {
		step.SetMetadata("files_written", fileCount)
		step.SetMetadata("total_bytes_written", totalSize)
	}
	return step
}

// ProfileGitOperations profiles git operations
func (t *Timer) ProfileGitOperations(operation string) *StepTimer {
	step := t.ProfileStep(StepGitOperations)
	if step != nil {
		step.SetMetadata("git_operation", operation)
		step.SetMetadata("repository_path", getWorkingDirectory())
	}
	return step
}

// ProfileUserInput profiles user input operations
func (t *Timer) ProfileUserInput(inputType string) *StepTimer {
	step := t.ProfileStep(StepUserInput)
	if step != nil {
		step.SetMetadata("input_type", inputType)
		step.SetMetadata("interactive_mode", true)
	}
	return step
}

// ProfileContextDetection profiles project context detection
func (t *Timer) ProfileContextDetection(workDir string) *StepTimer {
	step := t.ProfileStep(StepContextDetection)
	if step != nil {
		step.SetMetadata("working_directory", workDir)
		step.SetMetadata("detection_type", "project_context")
	}
	return step
}

// ProfileMenuDisplay profiles menu display operations
func (t *Timer) ProfileMenuDisplay(menuType string, optionCount int) *StepTimer {
	step := t.ProfileStep(StepMenuDisplay)
	if step != nil {
		step.SetMetadata("menu_type", menuType)
		step.SetMetadata("option_count", optionCount)
		step.SetMetadata("display_mode", "interactive")
	}
	return step
}

// Helper functions

func getTerminalWidth() int {
	// Basic terminal width detection (could be enhanced)
	return 80 // Default width
}

func getWorkingDirectory() string {
	wd, err := os.Getwd()
	if err != nil {
		return "unknown"
	}
	return wd
}

func getProjectState() map[string]interface{} {
	wd, err := os.Getwd()
	if err != nil {
		return nil
	}
	
	state := make(map[string]interface{})
	
	// Check for key project files
	state["has_stories_json"] = fileExists(filepath.Join(wd, "docs/2-current-epic/stories.json"))
	state["has_epic_docs"] = dirExists(filepath.Join(wd, "docs/2-current-epic"))
	state["has_current_task"] = dirExists(filepath.Join(wd, "docs/3-current-task"))
	state["is_git_repo"] = dirExists(filepath.Join(wd, ".git"))
	state["has_project_docs"] = dirExists(filepath.Join(wd, "docs/1-project"))
	
	return state
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}