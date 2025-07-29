package navigation

import (
	"fmt"
	"math"
	"strings"
	"time"
)

// ProjectStateDisplay handles the visual representation of project state
type ProjectStateDisplay struct {
	width int // Terminal width for formatting
}

// NewProjectStateDisplay creates a new project state display
func NewProjectStateDisplay() *ProjectStateDisplay {
	return &ProjectStateDisplay{
		width: 80, // Default terminal width
	}
}

// SetWidth sets the display width for formatting
func (psd *ProjectStateDisplay) SetWidth(width int) {
	psd.width = width
}

// DisplayProjectOverview shows a comprehensive overview of the project state
func (psd *ProjectStateDisplay) DisplayProjectOverview(ctx *ProjectContext) {
	psd.displayCompactHeader(ctx)

	if len(ctx.Issues) > 0 {
		psd.displayIssues(ctx)
	}
}

// displayCompactHeader shows a compact overview of the project state
func (psd *ProjectStateDisplay) displayCompactHeader(ctx *ProjectContext) {
	fmt.Println()
	
	// Top separator with title
	projectName := psd.getProjectName(ctx)
	title := fmt.Sprintf("  🚀 %s - %s  ", projectName, ctx.State.String())
	separatorWidth := 65
	titleWidth := len(title)
	leftPadding := (separatorWidth - titleWidth) / 2
	
	fmt.Print("═")
	for i := 0; i < leftPadding; i++ {
		fmt.Print("═")
	}
	fmt.Print(title)
	for i := leftPadding + titleWidth; i < separatorWidth; i++ {
		fmt.Print("═")
	}
	fmt.Println()
	
	// Project path
	if ctx.ProjectPath != "" {
		fmt.Printf("📂 Project Path: %s\n", ctx.ProjectPath)
	}
	
	// Epic status
	if ctx.CurrentEpic != nil {
		epic := ctx.CurrentEpic
		statusIcon := psd.getStatusIcon(epic.Status)
		priorityIcon := psd.getPriorityIcon(epic.Priority)
		fmt.Printf("📚 Current epic status: %s %s %s (%d/%d stories)\n", 
			statusIcon, priorityIcon, epic.Title, epic.CompletedStories, epic.TotalStories)
	} else {
		fmt.Printf("📚 Current epic status: No active epic\n")
	}
	
	// Story status  
	if ctx.CurrentStory != nil {
		story := ctx.CurrentStory
		statusIcon := psd.getStatusIcon(story.Status)
		priorityIcon := psd.getPriorityIcon(story.Priority)
		fmt.Printf("📖 Current story status: %s %s %s (%d/%d tasks)\n", 
			statusIcon, priorityIcon, story.Title, story.CompletedTasks, story.TotalTasks)
	} else if ctx.State >= StateStoryInProgress {
		fmt.Printf("📖 Current story status: No active story\n")
	}
	
	// Current step
	fmt.Printf("📍 Current step: %s\n", ctx.State.String())
	
	// Timestamp
	fmt.Printf("🕐 Last updated: %s\n", time.Now().Format("15:04:05"))
	
	// Bottom separator
	for i := 0; i < separatorWidth; i++ {
		fmt.Print("═")
	}
	fmt.Println()
	fmt.Println()
}

// displayIssues shows any project issues or warnings
func (psd *ProjectStateDisplay) displayIssues(ctx *ProjectContext) {
	if len(ctx.Issues) == 0 {
		return
	}

	fmt.Printf("⚠️  Issues (%d):\n", len(ctx.Issues))
	for i, issue := range ctx.Issues {
		if i >= 5 { // Limit to first 5 issues
			fmt.Printf("   ... and %d more\n", len(ctx.Issues)-5)
			break
		}
		fmt.Printf("   • %s\n", issue)
	}
	fmt.Println()
}

// Helper functions for visual formatting

// createProgressBar creates an ASCII progress bar
func (psd *ProjectStateDisplay) createProgressBar(progress float64, width int) string {
	if progress < 0 {
		progress = 0
	}
	if progress > 1 {
		progress = 1
	}

	filled := int(math.Round(progress * float64(width)))
	empty := width - filled

	filledBar := strings.Repeat("█", filled)
	emptyBar := strings.Repeat("░", empty)

	return fmt.Sprintf("[%s%s]", filledBar, emptyBar)
}

// getStateIcon returns an icon for the current state
func (psd *ProjectStateDisplay) getStateIcon(state WorkflowState) string {
	switch state {
	case StateNotInitialized:
		return "🆕 "
	case StateProjectInitialized:
		return "📁 "
	case StateHasEpics:
		return "📚 "
	case StateEpicInProgress:
		return "🚧 "
	case StateStoryInProgress:
		return "📖 "
	case StateTaskInProgress:
		return "⚡ "
	default:
		return "❓ "
	}
}

// getStatusIcon returns an icon for status strings
func (psd *ProjectStateDisplay) getStatusIcon(status string) string {
	status = strings.ToLower(status)
	switch {
	case strings.Contains(status, "completed") || strings.Contains(status, "done"):
		return "✅"
	case strings.Contains(status, "in_progress") || strings.Contains(status, "progress"):
		return "🚧"
	case strings.Contains(status, "todo") || strings.Contains(status, "pending"):
		return "⏳"
	case strings.Contains(status, "blocked"):
		return "🚫"
	case strings.Contains(status, "review"):
		return "👀"
	default:
		return "📋"
	}
}

// getPriorityIcon returns an icon for priority levels
func (psd *ProjectStateDisplay) getPriorityIcon(priority string) string {
	priority = strings.ToLower(priority)
	switch {
	case strings.Contains(priority, "high") || priority == "p0":
		return "🔴 "
	case strings.Contains(priority, "medium") || priority == "p1":
		return "🟡 "
	case strings.Contains(priority, "low") || priority == "p2":
		return "🟢 "
	default:
		return "⚪ "
	}
}

// getProjectName extracts or generates a project name
func (psd *ProjectStateDisplay) getProjectName(ctx *ProjectContext) string {
	// Try to extract from path
	if ctx.ProjectPath != "" {
		parts := strings.Split(strings.Trim(ctx.ProjectPath, "/"), "/")
		if len(parts) > 0 && parts[len(parts)-1] != "" {
			return parts[len(parts)-1]
		}
	}

	// Fallback name
	return "Claude WM Project"
}

// printSeparator prints a line separator
func (psd *ProjectStateDisplay) printSeparator(char string) {
	fmt.Println(strings.Repeat(char, psd.width))
}

// printCentered prints text centered within the display width
func (psd *ProjectStateDisplay) printCentered(text string) {
	textLen := len(text)
	if textLen >= psd.width {
		fmt.Println(text)
		return
	}

	padding := (psd.width - textLen) / 2
	fmt.Printf("%s%s%s\n",
		strings.Repeat(" ", padding),
		text,
		strings.Repeat(" ", psd.width-textLen-padding))
}

// DisplayQuickStatus shows a compact one-line status
func (psd *ProjectStateDisplay) DisplayQuickStatus(ctx *ProjectContext) {
	icon := psd.getStateIcon(ctx.State)
	fmt.Printf("%s%s", icon, ctx.State.String())

	if ctx.CurrentEpic != nil {
		fmt.Printf(" | Epic: %s (%.0f%%)", ctx.CurrentEpic.Title, ctx.CurrentEpic.Progress*100)
	}

	if ctx.CurrentStory != nil {
		fmt.Printf(" | Story: %s", ctx.CurrentStory.Title)
	}

	if ctx.CurrentTask != nil {
		fmt.Printf(" | Task: %s", ctx.CurrentTask.Title)
	}

	fmt.Println()
}

// DisplayProgressSummary shows a summary of progress across all levels
func (psd *ProjectStateDisplay) DisplayProgressSummary(ctx *ProjectContext) {
	fmt.Println("\n📊 Progress Summary:")

	if ctx.CurrentEpic != nil {
		epic := ctx.CurrentEpic
		fmt.Printf("   Epic: %s %.1f%% complete\n",
			psd.createProgressBar(epic.Progress, 20), epic.Progress*100)
	}

	if ctx.CurrentStory != nil && ctx.CurrentStory.TotalTasks > 0 {
		story := ctx.CurrentStory
		progress := float64(story.CompletedTasks) / float64(story.TotalTasks)
		fmt.Printf("   Story: %s %.1f%% complete\n",
			psd.createProgressBar(progress, 20), progress*100)
	}

	// Show next milestone
	if ctx.State < StateTaskInProgress {
		fmt.Printf("   Next: %s\n", psd.getNextMilestone(ctx))
	}

	fmt.Println()
}

// getNextMilestone suggests the next major milestone
func (psd *ProjectStateDisplay) getNextMilestone(ctx *ProjectContext) string {
	switch ctx.State {
	case StateNotInitialized:
		return "Initialize project structure"
	case StateProjectInitialized:
		return "Create first epic"
	case StateHasEpics:
		return "Start working on an epic"
	case StateEpicInProgress:
		return "Complete current epic"
	default:
		return "Continue current work"
	}
}

// DisplayActionSummary shows available actions with formatting
func (psd *ProjectStateDisplay) DisplayActionSummary(ctx *ProjectContext) {
	if len(ctx.AvailableActions) == 0 {
		fmt.Println("No actions available")
		return
	}

	fmt.Printf("\n💡 Available Actions (%d):\n", len(ctx.AvailableActions))

	for i, action := range ctx.AvailableActions {
		if i >= 8 { // Limit display
			fmt.Printf("   ... and %d more (use 'interactive' to see all)\n",
				len(ctx.AvailableActions)-8)
			break
		}
		fmt.Printf("   • %s\n", action)
	}

	fmt.Println()
}

// DisplayWithSuggestions combines state display with suggestions
func (psd *ProjectStateDisplay) DisplayWithSuggestions(ctx *ProjectContext, suggestions []*Suggestion) {
	psd.DisplayProjectOverview(ctx)

	if len(suggestions) > 0 {
		fmt.Println("🎯 Recommended Actions:")

		for i, suggestion := range suggestions {
			if i >= 3 { // Show top 3 suggestions
				break
			}

			icon := psd.getPriorityIcon(string(suggestion.Priority))
			fmt.Printf("   %d. %s%s\n", i+1, icon, suggestion.Action.Name)

			if suggestion.Reasoning != "" {
				fmt.Printf("      %s\n", suggestion.Reasoning)
			}
		}

		if len(suggestions) > 3 {
			fmt.Printf("   ... and %d more suggestions\n", len(suggestions)-3)
		}

		fmt.Println()
	}
}
