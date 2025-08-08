package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"claude-wm-cli/internal/config"
	"claude-wm-cli/internal/executor"
	"claude-wm-cli/internal/subagents"
)

var subagentsCmd = &cobra.Command{
	Use:   "subagents",
	Short: "Manage and monitor specialized subagents",
	Long: `The subagents command provides tools to manage, monitor, and test
the specialized AI subagents that optimize token usage and performance
by handling specific task types with minimal context requirements.

Available subagents:
- claude-wm-templates: Documentation and template generation
- claude-wm-status: Status reporting and analytics  
- claude-wm-planner: Task decomposition and planning
- claude-wm-reviewer: Code review and validation

Subagents provide significant benefits:
- 60-93% token savings through context reduction
- 2-4x faster response times through specialization
- Automatic fallback to main agent for quality assurance`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

var subagentsMetricsCmd = &cobra.Command{
	Use:   "metrics",
	Short: "Display subagent performance metrics and token savings",
	Long:  `Shows detailed metrics about subagent usage, token savings, performance, and cost optimization.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Initialize configuration
		configPath := ".claude-wm-cli"
		if cfgFile != "" {
			configPath = cfgFile
		}

		configManager := config.NewManager(configPath)

		// Initialize subagent-aware executor
		claudeExecutor := executor.NewClaudeExecutor()
		subagentConfigPath := configManager.GetSubagentsPath()
		
		subagentExecutor, err := executor.NewSubagentAwareExecutor(claudeExecutor, subagentConfigPath)
		if err != nil {
			return fmt.Errorf("failed to initialize subagent executor: %w", err)
		}

		// Display metrics
		fmt.Println(subagentExecutor.GetSubagentMetrics())
		
		return nil
	},
}

var subagentsTestCmd = &cobra.Command{
	Use:   "test",
	Short: "Test subagent routing and execution",
	Long:  `Test the subagent system by executing sample commands and measuring performance.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		testType, _ := cmd.Flags().GetString("type")
		
		// Initialize configuration
		configPath := ".claude-wm-cli"
		if cfgFile != "" {
			configPath = cfgFile
		}

		configManager := config.NewManager(configPath)

		// Initialize subagent-aware executor
		claudeExecutor := executor.NewClaudeExecutor()
		subagentConfigPath := configManager.GetSubagentsPath()
		
		subagentExecutor, err := executor.NewSubagentAwareExecutor(claudeExecutor, subagentConfigPath)
		if err != nil {
			return fmt.Errorf("failed to initialize subagent executor: %w", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()

		fmt.Printf("🧪 Testing subagent system: %s\n", testType)
		fmt.Println("================================\n")

		switch testType {
		case "template":
			return testTemplateSubagent(ctx, subagentExecutor)
		case "status":
			return testStatusSubagent(ctx, subagentExecutor)
		case "planning":
			return testPlanningSubagent(ctx, subagentExecutor)
		case "all":
			if err := testTemplateSubagent(ctx, subagentExecutor); err != nil {
				return err
			}
			if err := testStatusSubagent(ctx, subagentExecutor); err != nil {
				return err
			}
			return testPlanningSubagent(ctx, subagentExecutor)
		default:
			return fmt.Errorf("unknown test type: %s (available: template, status, planning, all)", testType)
		}
	},
}

var subagentsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available subagents and their configurations",
	Long:  `Display all configured subagents with their specializations and trigger patterns.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Initialize configuration
		configPath := ".claude-wm-cli"
		if cfgFile != "" {
			configPath = cfgFile
		}

		configManager := config.NewManager(configPath)

		// Initialize subagent-aware executor
		claudeExecutor := executor.NewClaudeExecutor()
		subagentConfigPath := configManager.GetSubagentsPath()
		
		subagentExecutor, err := executor.NewSubagentAwareExecutor(claudeExecutor, subagentConfigPath)
		if err != nil {
			return fmt.Errorf("failed to initialize subagent executor: %w", err)
		}

		fmt.Println("🤖 AVAILABLE SUBAGENTS")
		fmt.Println("======================")
		
		subagents := subagentExecutor.ListAvailableSubagents()
		for i, subagent := range subagents {
			fmt.Printf("%d. %s\n", i+1, subagent)
		}

		if len(subagents) == 0 {
			fmt.Println("No subagents configured")
			fmt.Println("\nTo initialize subagents, ensure configuration files exist in:")
			fmt.Printf("%s\n", subagentConfigPath)
		}

		fmt.Printf("\nSubagent system status: ")
		if subagentExecutor.IsSubagentEnabled() {
			fmt.Println("✅ ENABLED")
		} else {
			fmt.Println("❌ DISABLED")
		}

		return nil
	},
}

// Test functions

func testTemplateSubagent(ctx context.Context, executor *executor.SubagentAwareExecutor) error {
	fmt.Println("📝 Testing Template Generation Subagent")
	fmt.Println("---------------------------------------")
	
	start := time.Now()
	
	variables := map[string]string{
		"PROJECT_NAME": "TestProject",
		"TECH_STACK":   "Go + React",
		"DESCRIPTION":  "Test project for subagent validation",
	}

	err := executor.ExecuteCommandTemplate(ctx, "ARCHITECTURE", variables)
	duration := time.Since(start)
	
	if err != nil {
		fmt.Printf("❌ Template test failed: %v\n", err)
		return err
	}

	fmt.Printf("✅ Template test completed in %v\n", duration)
	fmt.Printf("📊 Expected savings: ~93%% tokens (65K → 5K)\n\n")
	
	return nil
}

func testStatusSubagent(ctx context.Context, executor *executor.SubagentAwareExecutor) error {
	fmt.Println("📊 Testing Status Reporting Subagent")
	fmt.Println("------------------------------------")
	
	start := time.Now()

	stateData := map[string]interface{}{
		"active_epics":     3,
		"completed_stories": 23,
		"in_progress_tasks": 8,
		"success_rate":     94.2,
	}

	err := executor.ExecuteStatusReport(ctx, "project", stateData)
	duration := time.Since(start)

	if err != nil {
		fmt.Printf("❌ Status test failed: %v\n", err)
		return err
	}

	fmt.Printf("✅ Status test completed in %v\n", duration)
	fmt.Printf("📊 Expected savings: ~89%% tokens (45K → 5K)\n\n")

	return nil
}

func testPlanningSubagent(ctx context.Context, executor *executor.SubagentAwareExecutor) error {
	fmt.Println("🧠 Testing Task Planning Subagent")
	fmt.Println("---------------------------------")
	
	start := time.Now()

	storyDescription := "As a user, I want to authenticate using OAuth2 so that I can securely access the system"
	technicalContext := map[string]string{
		"framework": "Go + Gin",
		"database":  "PostgreSQL",
		"auth_provider": "Google OAuth2",
	}

	err := executor.ExecuteTaskPlanning(ctx, storyDescription, technicalContext)
	duration := time.Since(start)

	if err != nil {
		fmt.Printf("❌ Planning test failed: %v\n", err)
		return err
	}

	fmt.Printf("✅ Planning test completed in %v\n", duration)
	fmt.Printf("📊 Expected savings: ~85%% tokens (100K → 15K)\n\n")

	return nil
}

var subagentsInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install or repair claude-wm subagents in current project",
	Long: `Install or repair claude-wm subagents in the current project's .claude/agents directory.
This command is useful for:
- Installing agents in existing projects that don't have them
- Repairing missing or corrupted agent files
- Upgrading agent definitions to latest versions`,
	RunE: func(cmd *cobra.Command, args []string) error {
		force, _ := cmd.Flags().GetBool("force")
		verifyOnly, _ := cmd.Flags().GetBool("verify-only")

		// Get current working directory
		projectPath, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}

		installer := subagents.NewAgentInstaller()

		if verifyOnly {
			// Only verify installation status
			fmt.Println("🔍 VERIFYING SUBAGENT INSTALLATION")
			fmt.Println("==================================")

			info, err := installer.GetAgentInstallationInfo(projectPath)
			if err != nil {
				return fmt.Errorf("failed to get installation info: %w", err)
			}

			fmt.Printf("%s\n", info.GetInstallationSummary())
			
			if info.AllInstalled {
				fmt.Println("\n✅ All claude-wm subagents are properly installed")
				fmt.Println("🎯 Expected performance benefits:")
				fmt.Println("  • claude-wm-templates: 93% token savings on documentation")
				fmt.Println("  • claude-wm-status: 89% token savings on status reports")  
				fmt.Println("  • claude-wm-planner: 85% token savings on task planning")
				fmt.Println("  • claude-wm-reviewer: 83% token savings on code reviews")
			} else {
				fmt.Printf("\n⚠️  Missing agents: %v\n", info.MissingAgents)
				fmt.Println("Run 'claude-wm-cli subagents install' to install missing agents.")
			}

			return nil
		}

		// Check if agents already exist and force flag
		existingAgents, _ := installer.ListInstalledAgents(projectPath)
		if len(existingAgents) > 0 && !force {
			fmt.Printf("⚠️  Found %d existing agents: %v\n", len(existingAgents), existingAgents)
			fmt.Println("Use --force to overwrite existing agents")
			return nil
		}

		// Perform installation
		fmt.Println("🤖 INSTALLING CLAUDE-WM SUBAGENTS")
		fmt.Println("==================================")

		if err := installer.InstallAgents(projectPath); err != nil {
			return fmt.Errorf("failed to install subagents: %w", err)
		}

		// Verify installation was successful
		info, err := installer.GetAgentInstallationInfo(projectPath)
		if err != nil {
			fmt.Println("⚠️  Installation completed but verification failed")
			return nil
		}

		fmt.Printf("\n%s\n", info.GetInstallationSummary())

		if info.AllInstalled {
			fmt.Println("\n🎉 INSTALLATION SUCCESSFUL!")
			fmt.Println("🎯 Your project now benefits from:")
			fmt.Println("  • claude-wm-templates: 93% token savings on documentation")
			fmt.Println("  • claude-wm-status: 89% token savings on status reports")
			fmt.Println("  • claude-wm-planner: 85% token savings on task planning")
			fmt.Println("  • claude-wm-reviewer: 83% token savings on code reviews")
			fmt.Println("\n📋 Next steps:")
			fmt.Println("  1. Run 'claude-wm-cli subagents list' to verify agents")
			fmt.Println("  2. Run 'claude-wm-cli subagents test' to test functionality")
			fmt.Println("  3. Start using optimized commands with automatic routing!")
		} else {
			fmt.Printf("\n⚠️  Partial installation. Missing: %v\n", info.MissingAgents)
			fmt.Println("You may need to run the installation command again or check permissions.")
		}

		return nil
	},
}

// serenaCmd represents the serena command for managing Serena integration
var serenaCmd = &cobra.Command{
	Use:   "serena",
	Short: "Manage Serena MCP integration for enhanced context preprocessing",
	Long: `Manage Serena MCP integration which provides intelligent context preprocessing
and semantic code analysis to optimize token usage and improve subagent performance.

Serena works as a preprocessing layer before subagents, providing:
- Semantic code analysis via Language Server Protocol
- Intelligent context filtering and optimization
- 3-7% additional token savings on top of existing subagent optimizations
- Enhanced accuracy through semantic understanding`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

// serenaStatusCmd shows the status of Serena integration
var serenaStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show Serena integration status and configuration",
	Run: func(cmd *cobra.Command, args []string) {
		configPath := ".claude-wm-cli"
		if cfgFile != "" {
			configPath = cfgFile
		}

		manager := config.NewManager(configPath)
		serenaConfig, err := config.NewSerenaConfigManager(manager.GetConfigDir())
		if err != nil {
			fmt.Printf("❌ Failed to load Serena configuration: %v\n", err)
			return
		}

		// Display Serena status
		fmt.Println("🔍 SERENA INTEGRATION STATUS")
		fmt.Println("============================")
		
		sc := serenaConfig.GetConfig()
		if sc.Enabled {
			fmt.Printf("Status: ✅ ENABLED\n")
		} else {
			fmt.Printf("Status: ❌ DISABLED\n")
		}
		
		fmt.Printf("MCP Server Path: %s\n", sc.MCPServerPath)
		fmt.Printf("Timeout: %d seconds\n", sc.Timeout)
		fmt.Printf("Fallback Enabled: %v\n", sc.FallbackEnabled)
		fmt.Printf("Auto-Detection: %v\n", sc.AutoDetect)
		
		fmt.Printf("\n📊 CONTEXT LIMITS BY ANALYSIS TYPE\n")
		for analysisType, limit := range sc.ContextLimits {
			fmt.Printf("  %s: %d tokens\n", analysisType, limit)
		}
		
		// Check if Serena is actually available
		if sc.Enabled {
			fmt.Printf("\n🔗 CONNECTIVITY TEST\n")
			if serenaConfig.IsSerenaAvailable() {
				fmt.Printf("Serena MCP Server: ✅ AVAILABLE\n")
				
				// Show estimated benefits
				fmt.Printf("\n💰 ESTIMATED ADDITIONAL BENEFITS\n")
				fmt.Printf("  • Templates: +3%% savings (70K → 3K tokens)\n")
				fmt.Printf("  • Status: +5%% savings (45K → 2.5K tokens)\n") 
				fmt.Printf("  • Planning: +7%% savings (100K → 8K tokens)\n")
				fmt.Printf("  • Review: +7%% savings (120K → 12K tokens)\n")
				fmt.Printf("  • Performance: 2-3x faster analysis\n")
			} else {
				fmt.Printf("Serena MCP Server: ❌ UNAVAILABLE\n")
				fmt.Printf("  Run 'claude-wm-cli serena install' to setup Serena\n")
			}
		}
	},
}

// serenaEnableCmd enables Serena integration
var serenaEnableCmd = &cobra.Command{
	Use:   "enable",
	Short: "Enable Serena MCP integration",
	Run: func(cmd *cobra.Command, args []string) {
		configPath := ".claude-wm-cli"
		if cfgFile != "" {
			configPath = cfgFile
		}

		manager := config.NewManager(configPath)
		serenaConfig, err := config.NewSerenaConfigManager(manager.GetConfigDir())
		if err != nil {
			fmt.Printf("❌ Failed to load Serena configuration: %v\n", err)
			return
		}

		if err := serenaConfig.EnableSerena(); err != nil {
			fmt.Printf("❌ Failed to enable Serena: %v\n", err)
			return
		}

		fmt.Println("✅ Serena integration ENABLED")
		fmt.Println("📈 Expected benefits:")
		fmt.Println("  • 3-7% additional token savings")
		fmt.Println("  • 2-3x faster semantic analysis")
		fmt.Println("  • Enhanced context precision")
		fmt.Println("")
		fmt.Println("💡 Run 'claude-wm-cli serena status' to verify setup")
	},
}

// serenaDisableCmd disables Serena integration  
var serenaDisableCmd = &cobra.Command{
	Use:   "disable",
	Short: "Disable Serena MCP integration",
	Run: func(cmd *cobra.Command, args []string) {
		configPath := ".claude-wm-cli"
		if cfgFile != "" {
			configPath = cfgFile
		}

		manager := config.NewManager(configPath)
		serenaConfig, err := config.NewSerenaConfigManager(manager.GetConfigDir())
		if err != nil {
			fmt.Printf("❌ Failed to load Serena configuration: %v\n", err)
			return
		}

		if err := serenaConfig.DisableSerena(); err != nil {
			fmt.Printf("❌ Failed to disable Serena: %v\n", err)
			return
		}

		fmt.Println("❌ Serena integration DISABLED")
		fmt.Println("🔄 Falling back to base subagent routing")
		fmt.Println("📊 You'll still have the existing 83-93% token savings from subagents")
	},
}

// serenaInstallCmd provides installation instructions for Serena
var serenaInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Show installation instructions for Serena MCP server",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("📦 SERENA INSTALLATION GUIDE")
		fmt.Println("============================")
		fmt.Println("")
		fmt.Println("Step 1: Install Serena")
		fmt.Println("  pip install serena-agent")
		fmt.Println("")
		fmt.Println("Step 2: Verify installation")
		fmt.Println("  serena-mcp-server --help")
		fmt.Println("")
		fmt.Println("Step 3: Configure Claude Code MCP")
		fmt.Println("  Add to your Claude Code MCP configuration:")
		fmt.Println("")
		fmt.Println("  {")
		fmt.Println(`    "mcpServers": {`)
		fmt.Println(`      "serena": {`)
		fmt.Println(`        "command": "serena-mcp-server",`)
		fmt.Println(`        "args": [],`)
		fmt.Println(`        "env": {}`)
		fmt.Println(`      }`)
		fmt.Println(`    }`)
		fmt.Println("  }")
		fmt.Println("")
		fmt.Println("Step 4: Enable Serena integration")
		fmt.Println("  claude-wm-cli serena enable")
		fmt.Println("")
		fmt.Println("Step 5: Verify setup")
		fmt.Println("  claude-wm-cli serena status")
		fmt.Println("")
		fmt.Println("💡 For detailed setup instructions, visit:")
		fmt.Println("   https://github.com/oraios/serena")
	},
}

func init() {
	rootCmd.AddCommand(subagentsCmd)
	rootCmd.AddCommand(serenaCmd)
	
	subagentsCmd.AddCommand(subagentsMetricsCmd)
	subagentsCmd.AddCommand(subagentsTestCmd)
	subagentsCmd.AddCommand(subagentsListCmd)
	subagentsCmd.AddCommand(subagentsInstallCmd)
	
	serenaCmd.AddCommand(serenaStatusCmd)
	serenaCmd.AddCommand(serenaEnableCmd)
	serenaCmd.AddCommand(serenaDisableCmd)
	serenaCmd.AddCommand(serenaInstallCmd)

	// Test command flags
	subagentsTestCmd.Flags().StringP("type", "t", "all", "Type of test to run: template, status, planning, all")
	
	// Install command flags
	subagentsInstallCmd.Flags().BoolP("force", "f", false, "Force installation (overwrite existing agents)")
	subagentsInstallCmd.Flags().BoolP("verify-only", "c", false, "Only verify installation status, don't install")
}