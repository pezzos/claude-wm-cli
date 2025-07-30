package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

func main() {
	// Parse command line arguments from Claude Code hook
	if len(os.Args) < 2 {
		// Silent exit if no arguments
		os.Exit(0)
	}

	userPrompt := strings.Join(os.Args[1:], " ")

	// Detect last executed command type
	suggestion := getNextStepSuggestion(userPrompt)
	if suggestion != "" {
		fmt.Printf("Next suggested step(s): %s\n", suggestion)
	}
}

func getNextStepSuggestion(prompt string) string {
	// Command patterns and their next step suggestions
	suggestions := map[string]string{
		`1-project:1-start:1-Init-Project`:           "/1-project:3-epics:1-Plan-Epics if ready to plan epics or /1-project:2-update:1-README if need to update project info",
		`1-project:3-epics:1-Plan-Epics`:             "/2-epic:1-start:1-Select-Stories if ready to start an epic or /1-project:2-update:3-Architecture if need architecture review",
		`1-project:3-epics:2-Update-Implementation`:  "/2-epic:1-start:1-Select-Stories if implementation complete or /1-project:2-update:4-Metrics if need status update",
		`2-epic:1-start:1-Select-Stories`:            "/2-epic:1-start:2-Plan-stories if epic selected or /1-project:3-epics:1-Plan-Epics if need different epic",
		`2-epic:1-start:2-Plan-stories`:              "/3-story:1-manage:1-Start-Story if stories planned or /2-epic:1-start:1-Select-Stories if need different stories",
		`3-story:1-manage:1-Start-Story`:             "/4-task:1-start:1-From-story if story started or /3-story:1-manage:2-Complete-Story if story done",
		`3-story:1-manage:2-Complete-Story`:          "/2-epic:2-manage:1-Complete-Epic if epic done or /3-story:1-manage:1-Start-Story if more stories",
		`4-task:1-start:1-From-story`:                "/4-task:2-execute:1-Plan-Task if task created or /3-story:1-manage:1-Start-Story if need different story",
		`4-task:2-execute:1-Plan-Task`:               "/4-task:2-execute:2-Test-design if ready to implement or /4-task:1-start:1-From-story if need replanning",
		`4-task:2-execute:2-Test-design`:             "/4-task:2-execute:3-Execute-Task if design approved or /4-task:2-execute:1-Plan-Task if need redesign",
		`4-task:2-execute:3-Execute-Task`:            "/4-task:2-execute:4-Validate-Task if implementation done or continue with /4-task:2-execute:3-Execute-Task",
		`4-task:2-execute:4-Validate-Task`:           "/4-task:3-complete:1-Archive-Task if validation passed or /4-task:2-execute:3-Execute-Task if fixes needed",
		`4-task:3-complete:1-Archive-Task`:           "/4-task:1-start:1-From-story if more tasks or /3-story:1-manage:2-Complete-Story if story complete",
		`2-epic:2-manage:1-Complete-Epic`:            "/1-project:3-epics:2-Update-Implementation then /1-project:3-epics:1-Plan-Epics if more epics needed",
		`2-epic:2-manage:2-Status-Epic`:              "/2-epic:1-start:2-Plan-stories if behind schedule or /3-story:1-manage:1-Start-Story if on track",
		`1-project:2-update:1-README`:                "/1-project:3-epics:1-Plan-Epics if ready for epics or /1-project:2-update:3-Architecture if need architecture",
		`1-project:2-update:2-Challenge`:             "/1-project:2-update:1-README if challenges addressed or /1-project:3-epics:1-Plan-Epics if ready to proceed",
		`1-project:2-update:3-Architecture`:          "/1-project:3-epics:1-Plan-Epics if architecture finalized or /1-project:2-update:2-Challenge if issues found",
		`1-project:2-update:4-Metrics`:               "/1-project:3-epics:1-Plan-Epics if metrics good or /2-epic:2-manage:2-Status-Epic if need epic review",
		`1-project:2-update:5-Implementation-Status`: "/1-project:3-epics:2-Update-Implementation if status updated or /2-epic:1-start:1-Select-Stories if ready for new work",
	}

	for pattern, suggestion := range suggestions {
		matched, _ := regexp.MatchString(pattern, prompt)
		if matched {
			return suggestion
		}
	}

	// Default suggestion for unmatched commands
	if strings.Contains(prompt, "commands/") {
		return "/1-project:2-update:4-Metrics if need status check or continue with current workflow"
	}

	return ""
}
