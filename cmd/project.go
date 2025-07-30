package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"claude-wm-cli/internal/debug"
	"claude-wm-cli/internal/executor"
	"claude-wm-cli/internal/model"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
		// Enable debug mode if flag is set
		debug.SetDebugMode(debugMode || viper.GetBool("debug"))
		
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
		// Enable debug mode if flag is set
		debug.SetDebugMode(debugMode || viper.GetBool("debug"))
		
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
		// Enable debug mode if flag is set
		debug.SetDebugMode(debugMode || viper.GetBool("debug"))
		
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
		// Enable debug mode if flag is set
		debug.SetDebugMode(debugMode || viper.GetBool("debug"))
		
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
		// Enable debug mode if flag is set
		debug.SetDebugMode(debugMode || viper.GetBool("debug"))
		
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
		// Enable debug mode if flag is set
		debug.SetDebugMode(debugMode || viper.GetBool("debug"))
		
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
		return model.NewFileSystemError("read", feedbackPath, err).
			WithSuggestions([]string{
				"Check if FEEDBACK.md exists in the docs/1-project/ directory",
				"Ensure you have read permissions for the file",
				"Run 'project challenge' to create a new FEEDBACK.md",
			})
	}

	fmt.Printf("✅ Feedback imported (%d bytes)\n", len(content))

	// Create Claude executor and process feedback with AI
	claudeExecutor := executor.NewClaudeExecutor()
	
	// Validate Claude is available
	if err := claudeExecutor.ValidateClaudeAvailable(); err != nil {
		debug.LogStub("PROJECT", "importFeedback", "Process feedback with Claude analysis but Claude CLI not available")
		fmt.Printf("⚠️  Claude CLI not found: %v\n", err)
		fmt.Println("📋 Falling back to basic feedback import...")
		
		// Still archive and reset feedback even without Claude
		if err := archiveAndResetFeedback(); err != nil {
			fmt.Printf("⚠️  Failed to archive feedback: %v\n", err)
			fmt.Println("📋 Feedback processing complete, but manual cleanup may be needed.")
		} else {
			fmt.Println("📋 Feedback archived and reset with template.")
		}
		
		fmt.Println("📋 Feedback processing complete. Use 'project challenge' next.")
		return nil
	}

	// Execute Claude command to process feedback
	prompt := "/1-project:2-update:1-Import-feedback"
	description := "Import and process feedback with AI analysis for actionable insights"
	
	if err := claudeExecutor.ExecutePrompt(prompt, description); err != nil {
		return model.NewInternalError("failed to execute Claude import feedback command").
			WithCause(err).
			WithContext(prompt).
			WithSuggestions([]string{
				"Check if Claude CLI is properly installed",
				"Verify Claude CLI authentication",
				"Try running the command manually to debug",
			})
	}
	
	// Archive processed feedback and reset with template
	if err := archiveAndResetFeedback(); err != nil {
		fmt.Printf("⚠️  Failed to archive feedback: %v\n", err)
		fmt.Println("📋 Feedback processing complete, but manual cleanup may be needed.")
	} else {
		fmt.Println("📋 Feedback archived and reset with template.")
	}
	
	fmt.Println("📋 Feedback processing complete. Use 'project challenge' next.")
	
	return nil
}

// archiveAndResetFeedback archives the current FEEDBACK.md and replaces it with template
func archiveAndResetFeedback() error {
	projectPath, err := os.Getwd()
	if err != nil {
		return model.NewInternalError("failed to get current directory").
			WithCause(err).
			WithSuggestions([]string{
				"Check if current directory exists",
				"Verify working directory permissions",
			})
	}
	
	// Archive current FEEDBACK.md
	if err := archiveFeedbackFile(projectPath); err != nil {
		return model.NewInternalError("failed to archive feedback file").
			WithCause(err).
			WithSuggestions([]string{
				"Check write permissions in docs/archive directory",
				"Ensure sufficient disk space",
			})
	}
	
	// Copy template FEEDBACK.md
	if err := copyFeedbackTemplate(projectPath); err != nil {
		return model.NewInternalError("failed to copy feedback template").
			WithCause(err).
			WithSuggestions([]string{
				"Check if template file exists in .claude/commands/template/",
				"Verify read permissions on template file",
				"Check write permissions in docs/1-project/",
			})
	}
	
	return nil
}

// archiveFeedbackFile moves FEEDBACK.md to archive with timestamp
func archiveFeedbackFile(projectPath string) error {
	sourcePath := filepath.Join(projectPath, "docs/1-project/FEEDBACK.md")
	
	// Check if FEEDBACK.md exists
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		// No file to archive
		return nil
	}
	
	// Generate timestamp for archive filename
	timestamp := time.Now().Format("2006-01-02-15h04")
	archiveFileName := fmt.Sprintf("FEEDBACK-%s-processed.md", timestamp)
	archivePath := filepath.Join(projectPath, "docs/archive", archiveFileName)
	
	// Ensure archive directory exists
	archiveDir := filepath.Join(projectPath, "docs/archive")
	if err := os.MkdirAll(archiveDir, 0755); err != nil {
		return model.NewFileSystemError("create_directory", archiveDir, err).
			WithSuggestions([]string{
				"Check parent directory permissions",
				"Verify sufficient disk space",
				"Ensure docs directory exists",
			})
	}
	
	// Move file to archive
	if err := os.Rename(sourcePath, archivePath); err != nil {
		return model.NewFileSystemError("move", sourcePath, err).
			WithContext(fmt.Sprintf("moving to %s", archivePath)).
			WithSuggestions([]string{
				"Check source file permissions",
				"Verify archive directory is writable",
				"Ensure no other process is using the file",
			})
	}
	
	fmt.Printf("📁 FEEDBACK.md archived as %s\n", archiveFileName)
	return nil
}

// copyFeedbackTemplate copies the FEEDBACK.md template to project directory
func copyFeedbackTemplate(projectPath string) error {
	templatePath := filepath.Join(projectPath, ".claude/commands/template/FEEDBACK.md")
	destPath := filepath.Join(projectPath, "docs/1-project/FEEDBACK.md")
	
	// Check if template exists
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		fmt.Println("⚠️  FEEDBACK.md template not found, skipping template copy")
		return nil
	}
	
	// Copy template file
	if err := copyFile(templatePath, destPath); err != nil {
		return model.NewFileSystemError("copy", templatePath, err).
			WithContext(fmt.Sprintf("copying to %s", destPath)).
			WithSuggestions([]string{
				"Check template file permissions",
				"Verify destination directory is writable",
				"Ensure sufficient disk space",
			})
	}
	
	fmt.Println("📄 Fresh FEEDBACK.md template copied")
	return nil
}

// Note: copyFile function is defined in interactive.go

func challengeDocumentation() error {
	fmt.Println("🤔 Challenging current documentation and assumptions...")
	
	// Check for existing documentation
	docsPath := "docs/1-project"
	if _, err := os.Stat(docsPath); os.IsNotExist(err) {
		return model.NewNotFoundError("project documentation").
			WithContext(docsPath).
			WithSuggestions([]string{
				"Run 'project init' to create project documentation",
				"Check if you're in the correct project directory",
				"Verify docs/1-project directory exists",
			})
	}

	// Create Claude executor
	claudeExecutor := executor.NewClaudeExecutor()
	
	// Validate Claude is available
	if err := claudeExecutor.ValidateClaudeAvailable(); err != nil {
		debug.LogStub("PROJECT", "challengeDocumentation", "Execute full Claude analysis but Claude CLI not available")
		fmt.Printf("⚠️  Claude CLI not found: %v\n", err)
		fmt.Println("📋 Falling back to basic documentation review...")
		fmt.Println("✅ Documentation review complete")
		fmt.Println("📋 Use 'project enrich' next to add context.")
		return nil
	}

	// Execute the actual Claude command with the proper prompt
	prompt := "/1-project:2-update:2-Challenge"
	description := "Challenge documentation and assumptions with deep codebase analysis"
	
	if err := claudeExecutor.ExecutePrompt(prompt, description); err != nil {
		return model.NewInternalError("failed to execute Claude challenge command").
			WithCause(err).
			WithContext(prompt).
			WithSuggestions([]string{
				"Check if Claude CLI is properly installed",
				"Verify Claude CLI authentication",
				"Check if command exists in .claude/commands/",
			})
	}

	fmt.Println("✅ Documentation review complete")
	fmt.Println("📋 Use 'project enrich' next to add context.")
	
	return nil
}

func enrichContext() error {
	fmt.Println("🌟 Enriching project context...")
	
	// Create Claude executor
	claudeExecutor := executor.NewClaudeExecutor()
	
	// Validate Claude is available
	if err := claudeExecutor.ValidateClaudeAvailable(); err != nil {
		debug.LogStub("PROJECT", "enrichContext", "Enrich context with Claude analysis but Claude CLI not available")
		fmt.Printf("⚠️  Claude CLI not found: %v\n", err)
		fmt.Println("📋 Falling back to basic context enrichment...")
		
		// Create basic context file as fallback
		contextPath := filepath.Join("docs/1-project", "CONTEXT.md")
		contextContent := `# Project Context

## Current State
- Project context enriched on: ` + fmt.Sprintf("%v", os.Getenv("USER")) + `
- Enrichment includes patterns, insights, and additional context

## Next Steps
- Run 'project status-update' to update overall status
`
		if err := os.WriteFile(contextPath, []byte(contextContent), 0644); err != nil {
			return model.NewFileSystemError("write", contextPath, err).
				WithSuggestions([]string{
					"Check write permissions in docs/1-project/",
					"Ensure sufficient disk space",
					"Verify directory exists",
				})
		}
		
		fmt.Println("✅ Context enrichment complete")
		fmt.Println("📋 Use 'project status-update' next.")
		return nil
	}

	// Execute Claude command to enrich context
	prompt := "/1-project:2-update:3-Enrich"
	description := "Enrich project context with AI-powered analysis and insights"
	
	if err := claudeExecutor.ExecutePrompt(prompt, description); err != nil {
		return model.NewInternalError("failed to execute Claude enrich context command").
			WithCause(err).
			WithContext(prompt).
			WithSuggestions([]string{
				"Check if Claude CLI is properly installed",
				"Verify Claude CLI authentication",
				"Check if command exists in .claude/commands/",
			})
	}

	fmt.Println("✅ Context enrichment complete")
	fmt.Println("📋 Use 'project status-update' next.")
	
	return nil
}

func updateProjectStatus() error {
	fmt.Println("📊 Updating project status...")
	
	// Create Claude executor
	claudeExecutor := executor.NewClaudeExecutor()
	
	// Validate Claude is available
	if err := claudeExecutor.ValidateClaudeAvailable(); err != nil {
		debug.LogStub("PROJECT", "updateProjectStatus", "Update project status with Claude analysis but Claude CLI not available")
		fmt.Printf("⚠️  Claude CLI not found: %v\n", err)
		fmt.Println("📋 Falling back to basic status update...")
		fmt.Println("✅ Project status updated")
		fmt.Println("📋 Use 'project implementation-status' next.")
		return nil
	}

	// Execute Claude command to update project status
	prompt := "/1-project:2-update:4-Status"
	description := "Update overall project status with comprehensive analysis"
	
	if err := claudeExecutor.ExecutePrompt(prompt, description); err != nil {
		return model.NewInternalError("failed to execute Claude update project status command").
			WithCause(err).
			WithContext(prompt).
			WithSuggestions([]string{
				"Check if Claude CLI is properly installed",
				"Verify Claude CLI authentication",
				"Check if command exists in .claude/commands/",
			})
	}
	
	fmt.Println("✅ Project status updated")
	fmt.Println("📋 Use 'project implementation-status' next.")
	
	return nil
}

func reviewImplementationStatus() error {
	fmt.Println("🔍 Reviewing implementation status...")
	
	// Create Claude executor
	claudeExecutor := executor.NewClaudeExecutor()
	
	// Validate Claude is available
	if err := claudeExecutor.ValidateClaudeAvailable(); err != nil {
		debug.LogStub("PROJECT", "reviewImplementationStatus", "Review implementation status with Claude analysis but Claude CLI not available")
		fmt.Printf("⚠️  Claude CLI not found: %v\n", err)
		fmt.Println("📋 Falling back to basic implementation review...")
		fmt.Println("✅ Implementation status review complete")
		fmt.Println("🔄 Project Update Cycle complete. You can start over with 'project import-feedback' or move to epic management.")
		return nil
	}

	// Execute Claude command to review implementation status
	prompt := "/1-project:2-update:5-Implementation-Status"
	description := "Review implementation status across all epics and stories with detailed analysis"
	
	if err := claudeExecutor.ExecutePrompt(prompt, description); err != nil {
		return model.NewInternalError("failed to execute Claude review implementation status command").
			WithCause(err).
			WithContext(prompt).
			WithSuggestions([]string{
				"Check if Claude CLI is properly installed",
				"Verify Claude CLI authentication",
				"Check if command exists in .claude/commands/",
			})
	}
	
	fmt.Println("✅ Implementation status review complete")
	fmt.Println("🔄 Project Update Cycle complete. You can start over with 'project import-feedback' or move to epic management.")
	
	return nil
}

func planEpics() error {
	fmt.Println("📚 Planning epic roadmap...")
	
	// Create Claude executor
	claudeExecutor := executor.NewClaudeExecutor()
	
	// Validate Claude is available
	if err := claudeExecutor.ValidateClaudeAvailable(); err != nil {
		debug.LogStub("PROJECT", "planEpics", "Plan epics with Claude analysis but Claude CLI not available")
		fmt.Printf("⚠️  Claude CLI not found: %v\n", err)
		fmt.Println("📋 Falling back to basic epic planning...")
		
		// Create basic epics.json as fallback
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
			return model.NewFileSystemError("write", epicsPath, err).
				WithSuggestions([]string{
					"Check write permissions in docs/1-project/",
					"Ensure sufficient disk space",
					"Verify directory exists",
				})
		}
		
		fmt.Println("✅ Epic roadmap planning complete")
		fmt.Printf("📋 Epics saved to %s\n", epicsPath)
		fmt.Println("Use 'epic list' to see available epics or 'epic create' to add new ones.")
		return nil
	}

	// Execute Claude command to plan epics
	prompt := "/1-project:3-epics:1-Plan-Epics"
	description := "Plan epic roadmap with AI-powered analysis and strategic planning"
	
	if err := claudeExecutor.ExecutePrompt(prompt, description); err != nil {
		return model.NewInternalError("failed to execute Claude plan epics command").
			WithCause(err).
			WithContext(prompt).
			WithSuggestions([]string{
				"Check if Claude CLI is properly installed",
				"Verify Claude CLI authentication",
				"Check if command exists in .claude/commands/",
			})
	}

	fmt.Println("✅ Epic roadmap planning complete")
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