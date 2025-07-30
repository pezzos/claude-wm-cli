package cmd

import (
	"fmt"
	"strings"

	"claude-wm-cli/internal/navigation"
)

// executeTaskListFromStory executes task list command for current story
func executeTaskListFromStory(ctx *navigation.ProjectContext, menuDisplay *navigation.MenuDisplay) error {
	wd := "." // Default to current directory
	
	fmt.Println("📋 Listing tasks from current story...")
	
	if err := displayTasksFromCurrentStory(wd, ""); err != nil {
		return fmt.Errorf("failed to display tasks: %w", err)
	}
	
	fmt.Println("✅ ✅ Task list completed successfully")
	return nil
}

// Helper functions for task display
func getTaskStatusIcon(status string) string {
	switch strings.ToLower(status) {
	case "todo":
		return "⏳"
	case "in_progress":
		return "🚧"
	case "done", "completed":
		return "✅"
	case "blocked":
		return "🚫"
	default:
		return "❓"
	}
}

func getTaskPriorityIcon(priority string) string {
	switch strings.ToUpper(priority) {
	case "P0", "CRITICAL":
		return "🔥"
	case "P1", "HIGH":
		return "⚡" 
	case "P2", "MEDIUM":
		return "📋"
	case "P3", "LOW":
		return "📝"
	default:
		return "❓"
	}
}

