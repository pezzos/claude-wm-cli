package hooks

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/claude-wm-cli/internal/formatting"
	"github.com/claude-wm-cli/internal/git"
	"github.com/claude-wm-cli/internal/validation"
)

// ToolInput represents input from Claude Code hooks
type ToolInput struct {
	ToolName  string                 `json:"tool_name"`
	ToolInput map[string]interface{} `json:"tool_input"`
}

// HookHandler handles hook execution for claude-wm-cli
type HookHandler struct {
	projectRoot string
}

// NewHookHandler creates a new hook handler
func NewHookHandler(projectRoot string) *HookHandler {
	return &HookHandler{
		projectRoot: projectRoot,
	}
}

// HandleGitValidation handles git validation hooks
func (h *HookHandler) HandleGitValidation() error {
	// Read input from stdin
	inputBytes, err := os.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("error reading input: %v", err)
	}

	var input ToolInput
	if err := json.Unmarshal(inputBytes, &input); err != nil {
		return fmt.Errorf("error parsing JSON input: %v", err)
	}

	// Initialize git validator
	validator, err := git.NewValidator()
	if err != nil {
		return fmt.Errorf("error initializing git validator: %v", err)
	}

	// Run validation
	success := validator.ValidateTool(input.ToolName, input.ToolInput)

	// Print results
	validator.PrintResults()

	// Exit with appropriate code
	if success {
		return nil
	} else {
		os.Exit(2)
		return nil
	}
}

// HandleAutoFormat handles auto-formatting hooks
func (h *HookHandler) HandleAutoFormat() error {
	formatter := formatting.NewFormatter(h.projectRoot)
	return formatter.FormatAll()
}

// HandleDuplicateDetection handles duplicate detection hooks
func (h *HookHandler) HandleDuplicateDetection() error {
	// Read input from stdin
	inputBytes, err := os.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("error reading input: %v", err)
	}

	var input ToolInput
	if err := json.Unmarshal(inputBytes, &input); err != nil {
		return fmt.Errorf("error parsing JSON input: %v", err)
	}

	// Only process Write tool for new files
	if input.ToolName != "Write" {
		return nil
	}

	filePath, ok := input.ToolInput["file_path"].(string)
	if !ok || filePath == "" {
		return nil
	}

	// Initialize detector
	detector := validation.NewDuplicateDetector(h.projectRoot)

	// Skip if not relevant
	if !detector.IsRelevantForDetection(filePath) {
		return nil
	}

	// Run detection
	result := detector.DetectDuplicates(filePath)

	// Output results
	if len(result.Errors) > 0 {
		for _, error := range result.Errors {
			fmt.Fprintf(os.Stderr, "❌ %s\n", error)
		}
		fmt.Fprintf(os.Stderr, "\nCheck existing files before creating new ones.\n")
	}

	if len(result.Warnings) > 0 {
		for _, warning := range result.Warnings {
			fmt.Fprintf(os.Stderr, "⚠️  %s\n", warning)
		}
	}

	// Exit with appropriate code
	if result.Success {
		return nil
	} else {
		os.Exit(2)
		return nil
	}
}