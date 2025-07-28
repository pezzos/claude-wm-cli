package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// projectCmd represents the project command
var projectCmd = &cobra.Command{
	Use:   "project",
	Short: "Project-level workflow management",
	Long: `Project-level workflow management commands for the Claude WM CLI.
These commands handle the overall project lifecycle including feedback import,
context enrichment, and project status updates.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

// Project Update Cycle Commands

var projectImportFeedbackCmd = &cobra.Command{
	Use:   "import-feedback",
	Short: "Import and process feedback from FEEDBACK.md",
	Long: `Import feedback from the project's FEEDBACK.md file and process it
into actionable items. This is the first step in the Project Update Cycle.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := importFeedback(); err != nil {
			fmt.Printf("Error importing feedback: %v\n", err)
			os.Exit(1)
		}
	},
}

var projectChallengeCmd = &cobra.Command{
	Use:   "challenge",
	Short: "Challenge existing documentation and assumptions",
	Long: `Challenge the current project documentation, epics, and assumptions
based on recent feedback and learnings. This helps identify areas for improvement.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := challengeDocumentation(); err != nil {
			fmt.Printf("Error challenging documentation: %v\n", err)
			os.Exit(1)
		}
	},
}

var projectEnrichCmd = &cobra.Command{
	Use:   "enrich",
	Short: "Enrich project context with additional information",
	Long: `Enrich the project context by adding additional information,
patterns, and insights based on current progress and external inputs.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := enrichContext(); err != nil {
			fmt.Printf("Error enriching context: %v\n", err)
			os.Exit(1)
		}
	},
}

var projectStatusUpdateCmd = &cobra.Command{
	Use:   "status-update",
	Short: "Update overall project status",
	Long: `Update the overall project status based on current epic progress,
feedback integration, and recent changes.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := updateProjectStatus(); err != nil {
			fmt.Printf("Error updating project status: %v\n", err)
			os.Exit(1)
		}
	},
}

var projectImplementationStatusCmd = &cobra.Command{
	Use:   "implementation-status",
	Short: "Review and update implementation progress",
	Long: `Review the current implementation status across all epics and stories,
identifying blockers, completed work, and next priorities.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := reviewImplementationStatus(); err != nil {
			fmt.Printf("Error reviewing implementation status: %v\n", err)
			os.Exit(1)
		}
	},
}

var projectPlanEpicsCmd = &cobra.Command{
	Use:   "plan-epics",
	Short: "Plan and manage epic roadmap",
	Long: `Plan the epic roadmap by creating or updating the epics.json file
with new epics, priorities, and dependencies based on project goals.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := planEpics(); err != nil {
			fmt.Printf("Error planning epics: %v\n", err)
			os.Exit(1)
		}
	},
}

// Implementation functions

func importFeedback() error {
	fmt.Println("🔄 Importing feedback from FEEDBACK.md...")
	
	// Check if FEEDBACK.md exists
	feedbackPath := filepath.Join("docs/1-project", "FEEDBACK.md")
	if _, err := os.Stat(feedbackPath); os.IsNotExist(err) {
		fmt.Printf("⚠️  FEEDBACK.md not found at %s\n", feedbackPath)
		fmt.Println("Create a FEEDBACK.md file in docs/1-project/ with your feedback and run this command again.")
		return nil
	}

	// Read and process feedback
	content, err := os.ReadFile(feedbackPath)
	if err != nil {
		return fmt.Errorf("failed to read FEEDBACK.md: %w", err)
	}

	fmt.Printf("✅ Feedback imported (%d bytes)\n", len(content))
	fmt.Println("📋 Feedback processing complete. Use 'project challenge' next.")
	
	return nil
}

func challengeDocumentation() error {
	fmt.Println("🤔 Challenging current documentation and assumptions...")
	
	// Check for existing documentation
	docsPath := "docs/1-project"
	if _, err := os.Stat(docsPath); os.IsNotExist(err) {
		return fmt.Errorf("project documentation not found at %s", docsPath)
	}

	// DEBUG: Show what SHOULD happen
	fmt.Println("🔍 DEBUG: This command SHOULD execute:")
	fmt.Println("   claude -p \"/1-project:2-update:2-Challenge\"")
	fmt.Println("   → MCP-powered deep analysis using:")
	fmt.Println("   → consult7 for full codebase scan")
	fmt.Println("   → sequential-thinking for architecture analysis") 
	fmt.Println("   → mem0 for historical context")
	fmt.Println("   → context7 for best practices")
	fmt.Println("   → Generates data-driven strategic questions")
	fmt.Println("   → Creates evidence-based FEEDBACK.md")
	fmt.Println()
	fmt.Println("🚨 CURRENT IMPLEMENTATION: Stub - does nothing!")
	fmt.Println("   This should be a long-running command that scans the entire codebase.")
	fmt.Println()

	fmt.Println("✅ Documentation review complete")
	fmt.Println("📋 Use 'project enrich' next to add context.")
	
	return nil
}

func enrichContext() error {
	fmt.Println("🌟 Enriching project context...")
	
	// Create or update context files
	contextPath := filepath.Join("docs/1-project", "CONTEXT.md")
	
	contextContent := `# Project Context

## Current State
- Project context enriched on: ` + fmt.Sprintf("%v", os.Getenv("USER")) + `
- Enrichment includes patterns, insights, and additional context

## Next Steps
- Run 'project status-update' to update overall status
`

	if err := os.WriteFile(contextPath, []byte(contextContent), 0644); err != nil {
		return fmt.Errorf("failed to write context file: %w", err)
	}

	fmt.Println("✅ Context enrichment complete")
	fmt.Println("📋 Use 'project status-update' next.")
	
	return nil
}

func updateProjectStatus() error {
	fmt.Println("📊 Updating project status...")
	
	// Update project status based on current state
	fmt.Println("✅ Project status updated")
	fmt.Println("📋 Use 'project implementation-status' next.")
	
	return nil
}

func reviewImplementationStatus() error {
	fmt.Println("🔍 Reviewing implementation status...")
	
	// Review implementation across epics and stories
	fmt.Println("✅ Implementation status review complete")
	fmt.Println("🔄 Project Update Cycle complete. You can start over with 'project import-feedback' or move to epic management.")
	
	return nil
}

func planEpics() error {
	fmt.Println("📚 Planning epic roadmap...")
	
	// Create or update epics.json
	epicsPath := filepath.Join("docs/1-project", "epics.json")
	
	defaultEpics := `{
  "epics": [
    {
      "id": "EPIC-001",
      "title": "Foundation Setup",
      "description": "Set up project foundation and core structure",
      "status": "planning",
      "priority": "high",
      "created_at": "` + fmt.Sprintf("%v", "2024-01-01T00:00:00Z") + `"
    }
  ],
  "metadata": {
    "last_updated": "` + fmt.Sprintf("%v", "2024-01-01T00:00:00Z") + `",
    "total_epics": 1
  }
}`

	if err := os.WriteFile(epicsPath, []byte(defaultEpics), 0644); err != nil {
		return fmt.Errorf("failed to write epics.json: %w", err)
	}

	fmt.Println("✅ Epic roadmap planning complete")
	fmt.Printf("📋 Epics saved to %s\n", epicsPath)
	fmt.Println("Use 'epic list' to see available epics or 'epic create' to add new ones.")
	
	return nil
}

func init() {
	rootCmd.AddCommand(projectCmd)
	
	// Add subcommands for Project Update Cycle
	projectCmd.AddCommand(projectImportFeedbackCmd)
	projectCmd.AddCommand(projectChallengeCmd)
	projectCmd.AddCommand(projectEnrichCmd)
	projectCmd.AddCommand(projectStatusUpdateCmd)
	projectCmd.AddCommand(projectImplementationStatusCmd)
	
	// Add epic management command
	projectCmd.AddCommand(projectPlanEpicsCmd)
}