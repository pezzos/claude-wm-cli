/*
Copyright © 2025 Claude WM CLI Team
*/
package cmd

import (
	"fmt"
	"os"
	"sort"
	"text/tabwriter"

	"claude-wm-cli/internal/metrics"

	"github.com/spf13/cobra"
)

// metricsCmd represents the metrics command
var metricsCmd = &cobra.Command{
	Use:   "metrics",
	Short: "Performance metrics and analysis",
	Long: `Performance metrics and analysis for claude-wm-cli commands.

This command provides detailed performance statistics to help identify
slow operations and optimization opportunities.

Features:
  • Command-level performance statistics (min/avg/max durations)
  • Step-level profiling within commands
  • Project-wide performance comparison
  • Historical trend analysis
  • Slow command identification

The metrics are stored in $HOME/.claude-wm/metrics/performance.db

Examples:
  claude-wm-cli metrics status               # Overall metrics status
  claude-wm-cli metrics commands            # List all command statistics
  claude-wm-cli metrics command "Start Story" --days 7  # Specific command stats
  claude-wm-cli metrics steps "Start Story" # Step-level profiling
  claude-wm-cli metrics slow --threshold 5000  # Commands slower than 5s
  claude-wm-cli metrics projects            # Performance by project`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return showMetricsStatus()
	},
}

// Subcommands
var (
	metricsStatusCmd = &cobra.Command{
		Use:   "status",
		Short: "Show metrics collection status",
		Long:  `Display the current status of metrics collection and basic statistics.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return showMetricsStatus()
		},
	}

	metricsCommandsCmd = &cobra.Command{
		Use:   "commands",
		Short: "List performance statistics for all commands",
		Long:  `Display performance statistics for all commands with min/avg/max durations.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return showCommandMetrics(metricsDays)
		},
	}

	metricsCommandCmd = &cobra.Command{
		Use:   "command <command-name>",
		Short: "Show detailed statistics for a specific command",
		Long:  `Display detailed performance statistics for a specific command including step-level analysis.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return showCommandDetails(args[0], metricsDays)
		},
	}

	metricsStepsCmd = &cobra.Command{
		Use:   "steps <command-name>",
		Short: "Show step-level profiling for a command",
		Long:  `Display step-level performance profiling for a specific command.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return showStepMetrics(args[0], metricsDays)
		},
	}

	metricsSlowCmd = &cobra.Command{
		Use:   "slow",
		Short: "Show slowest commands",
		Long:  `List commands that are slower than the specified threshold.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return showSlowCommands(metricsThreshold, metricsDays)
		},
	}

	metricsProjectsCmd = &cobra.Command{
		Use:   "projects",
		Short: "Compare performance across projects",
		Long:  `Display performance comparison across different projects.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return showProjectComparison(metricsDays)
		},
	}

	metricsCleanCmd = &cobra.Command{
		Use:   "clean",
		Short: "Clean metrics database",
		Long:  `Clean old metrics data from the database. Use --force to confirm deletion.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cleanMetrics(metricsForce, metricsDays)
		},
	}
)

// Command flags
var (
	metricsDays      int
	metricsThreshold int64
	metricsForce     bool
)

func init() {
	rootCmd.AddCommand(metricsCmd)

	// Add subcommands
	metricsCmd.AddCommand(metricsStatusCmd)
	metricsCmd.AddCommand(metricsCommandsCmd)
	metricsCmd.AddCommand(metricsCommandCmd)
	metricsCmd.AddCommand(metricsStepsCmd)
	metricsCmd.AddCommand(metricsSlowCmd)
	metricsCmd.AddCommand(metricsProjectsCmd)
	metricsCmd.AddCommand(metricsCleanCmd)

	// Add flags
	metricsCmd.PersistentFlags().IntVar(&metricsDays, "days", 30, "Number of days to analyze")
	metricsSlowCmd.Flags().Int64Var(&metricsThreshold, "threshold", 3000, "Threshold in milliseconds for slow commands")
	metricsCleanCmd.Flags().BoolVar(&metricsForce, "force", false, "Force deletion without confirmation")
}

// showMetricsStatus displays the current metrics collection status
func showMetricsStatus() error {
	collector := metrics.GetCollector()
	
	fmt.Printf("🔍 Claude WM CLI Performance Metrics\n")
	fmt.Printf("=====================================\n\n")
	
	if !collector.IsEnabled() {
		fmt.Printf("❌ Metrics collection is DISABLED\n")
		fmt.Printf("   This usually means the SQLite database could not be initialized.\n")
		fmt.Printf("   Check that $HOME/.claude-wm/metrics/ is writable.\n\n")
		return nil
	}
	
	fmt.Printf("✅ Metrics collection is ENABLED\n")
	fmt.Printf("   Database: %s\n", getDatabasePath())
	fmt.Printf("   Tool version: %s\n\n", metrics.GetToolVersion())
	
	// Show basic statistics
	stats, err := collector.GetStats("", metricsDays)
	if err != nil {
		fmt.Printf("⚠️  Could not retrieve statistics: %v\n", err)
		return nil
	}
	
	if stats.Count == 0 {
		fmt.Printf("📊 No metrics data available for the last %d days\n", metricsDays)
		fmt.Printf("   Run some commands to collect performance data!\n\n")
		return nil
	}
	
	fmt.Printf("📊 Metrics Summary (last %d days):\n", metricsDays)
	fmt.Printf("   Total commands executed: %d\n", stats.Count)
	fmt.Printf("   Average execution time: %.0fms\n", stats.AvgDuration)
	fmt.Printf("   Fastest command: %.0fms\n", stats.MinDuration)
	fmt.Printf("   Slowest command: %.0fms\n", stats.MaxDuration)
	if stats.P95Duration > 0 {
		fmt.Printf("   95th percentile: %.0fms\n", stats.P95Duration)
	}
	fmt.Printf("\n")
	
	fmt.Printf("💡 Available commands:\n")
	fmt.Printf("   • claude-wm-cli metrics commands     # List all command statistics\n")
	fmt.Printf("   • claude-wm-cli metrics slow         # Find slow commands\n")
	fmt.Printf("   • claude-wm-cli metrics projects     # Compare project performance\n")
	fmt.Printf("   • claude-wm-cli metrics command \"Start Story\"  # Detailed command analysis\n")
	
	return nil
}

// showCommandMetrics displays performance statistics for all commands
func showCommandMetrics(days int) error {
	collector := metrics.GetCollector()
	if !collector.IsEnabled() {
		return fmt.Errorf("metrics collection is disabled")
	}
	
	fmt.Printf("📊 Command Performance Statistics (last %d days)\n", days)
	fmt.Printf("==================================================\n\n")
	
	commands, err := collector.GetAllCommandStats(days)
	if err != nil {
		return fmt.Errorf("failed to get command statistics: %w", err)
	}
	
	if len(commands) == 0 {
		fmt.Printf("📊 No command data available for the last %d days\n", days)
		fmt.Printf("   Run some commands to collect performance data!\n")
		return nil
	}
	
	// Create table
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "COMMAND\tEXECUTIONS\tMIN\tAVG\tMAX\tPERFORMANCE\n")
	fmt.Fprintf(w, "───────\t──────────\t───\t───\t───\t───────────\n")
	
	for _, cmd := range commands {
		performance := getPerformanceIcon(cmd.AvgDuration)
		fmt.Fprintf(w, "%s\t%d\t%.0fms\t%.0fms\t%.0fms\t%s\n",
			truncateMetricsString(cmd.CommandName, 35),
			cmd.Count,
			cmd.MinDuration,
			cmd.AvgDuration,
			cmd.MaxDuration,
			performance)
	}
	
	w.Flush()
	
	fmt.Printf("\n💡 Next steps:\n")
	fmt.Printf("   • claude-wm-cli metrics command \"<name>\"  # Detailed analysis\n")
	fmt.Printf("   • claude-wm-cli metrics steps \"<name>\"    # Step-level profiling\n")
	fmt.Printf("   • claude-wm-cli metrics slow              # Find slow commands\n")
	
	return nil
}

// showCommandDetails displays detailed statistics for a specific command
func showCommandDetails(commandName string, days int) error {
	collector := metrics.GetCollector()
	if !collector.IsEnabled() {
		return fmt.Errorf("metrics collection is disabled")
	}
	
	fmt.Printf("🔍 Command Analysis: %s (last %d days)\n", commandName, days)
	fmt.Printf("===========================================\n\n")
	
	// Get command statistics
	stats, err := collector.GetStats(commandName, days)
	if err != nil {
		return fmt.Errorf("failed to get command statistics: %w", err)
	}
	
	if stats.Count == 0 {
		fmt.Printf("📊 No data available for command '%s' in the last %d days\n", commandName, days)
		fmt.Printf("   Make sure the command name is correct and you've run it recently.\n")
		return nil
	}
	
	fmt.Printf("📈 Execution Statistics:\n")
	fmt.Printf("   Executions: %d\n", stats.Count)
	fmt.Printf("   Min time:   %.0fms\n", stats.MinDuration)
	fmt.Printf("   Avg time:   %.0fms\n", stats.AvgDuration)
	fmt.Printf("   Max time:   %.0fms\n", stats.MaxDuration)
	if stats.P95Duration > 0 {
		fmt.Printf("   95th %%ile:  %.0fms\n", stats.P95Duration)
	}
	fmt.Printf("\n")
	
	// Performance assessment
	avgSeconds := stats.AvgDuration / 1000.0
	if avgSeconds < 1.0 {
		fmt.Printf("⚡ Performance: FAST (< 1 second average)\n")
	} else if avgSeconds < 5.0 {
		fmt.Printf("✅ Performance: GOOD (< 5 seconds average)\n")  
	} else if avgSeconds < 15.0 {
		fmt.Printf("⚠️  Performance: SLOW (> 5 seconds average)\n")
	} else {
		fmt.Printf("🐌 Performance: VERY SLOW (> 15 seconds average)\n")
	}
	
	// Show step-level analysis if available
	steps, err := collector.GetStepStats(commandName, days)
	if err == nil && len(steps) > 0 {
		fmt.Printf("\n🔬 Step-level Analysis:\n")
		fmt.Printf("   Use 'claude-wm-cli metrics steps \"%s\"' for detailed breakdown\n", commandName)
	}
	
	return nil
}

// showStepMetrics displays step-level profiling for a command
func showStepMetrics(commandName string, days int) error {
	collector := metrics.GetCollector()
	if !collector.IsEnabled() {
		return fmt.Errorf("metrics collection is disabled")
	}
	
	fmt.Printf("🔬 Step-level Profiling: %s (last %d days)\n", commandName, days)
	fmt.Printf("============================================\n\n")
	
	steps, err := collector.GetStepStats(commandName, days)
	if err != nil {
		return fmt.Errorf("failed to get step statistics: %w", err)
	}
	
	if len(steps) == 0 {
		fmt.Printf("📊 No step-level data available for command '%s'\n", commandName)
		fmt.Printf("   Step-level profiling may not be implemented for this command yet.\n")
		return nil
	}
	
	// Sort steps by average duration (slowest first)
	sort.Slice(steps, func(i, j int) bool {
		return steps[i].AvgDuration > steps[j].AvgDuration
	})
	
	// Create table
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "STEP\tEXECUTIONS\tMIN\tAVG\tMAX\tIMPACT\n")
	fmt.Fprintf(w, "────\t──────────\t───\t───\t───\t──────\n")
	
	// Calculate total average for impact percentage
	var totalAvg float64
	for _, step := range steps {
		totalAvg += step.AvgDuration
	}
	
	for _, step := range steps {
		impactPercent := (step.AvgDuration / totalAvg) * 100
		impactIcon := getImpactIcon(impactPercent)
		
		fmt.Fprintf(w, "%s\t%d\t%.0fms\t%.0fms\t%.0fms\t%s %.1f%%\n",
			truncateMetricsString(step.StepName, 25),
			step.Count,
			step.MinDuration,
			step.AvgDuration,
			step.MaxDuration,
			impactIcon,
			impactPercent)
	}
	
	w.Flush()
	
	// Performance recommendations
	fmt.Printf("\n💡 Performance Insights:\n")
	if len(steps) > 0 {
		slowestStep := steps[0]
		if slowestStep.AvgDuration > 2000 {
			fmt.Printf("   • '%s' is the slowest step (%.0fms avg) - consider optimization\n", 
				slowestStep.StepName, slowestStep.AvgDuration)
		}
		
		if len(steps) >= 3 {
			topThreeTime := steps[0].AvgDuration + steps[1].AvgDuration + steps[2].AvgDuration
			topThreePercent := (topThreeTime / totalAvg) * 100
			if topThreePercent > 80 {
				fmt.Printf("   • Top 3 steps account for %.1f%% of execution time\n", topThreePercent)
			}
		}
	}
	
	return nil
}

// showSlowCommands displays commands slower than threshold
func showSlowCommands(thresholdMs int64, days int) error {
	collector := metrics.GetCollector()
	if !collector.IsEnabled() {
		return fmt.Errorf("metrics collection is disabled")
	}
	
	fmt.Printf("🐌 Slow Commands (> %dms, last %d days)\n", thresholdMs, days)
	fmt.Printf("=====================================\n\n")
	
	slowCommands, err := collector.GetSlowCommands(thresholdMs, days)
	if err != nil {
		return fmt.Errorf("failed to get slow commands: %w", err)
	}
	
	if len(slowCommands) == 0 {
		fmt.Printf("🎉 No commands slower than %dms found!\n", thresholdMs)
		fmt.Printf("   Your performance looks good.\n")
		return nil
	}
	
	// Sort by average duration (slowest first)
	sort.Slice(slowCommands, func(i, j int) bool {
		return slowCommands[i].AvgDuration > slowCommands[j].AvgDuration
	})
	
	// Create table
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "COMMAND\tEXECUTIONS\tAVG TIME\tMAX TIME\tSEVERITY\n")
	fmt.Fprintf(w, "───────\t──────────\t────────\t────────\t────────\n")
	
	for _, cmd := range slowCommands {
		severity := getSeverityIcon(cmd.AvgDuration)
		fmt.Fprintf(w, "%s\t%d\t%.0fms\t%.0fms\t%s\n",
			truncateMetricsString(cmd.CommandName, 30),
			cmd.Count,
			cmd.AvgDuration,
			cmd.MaxDuration,
			severity)
	}
	
	w.Flush()
	
	fmt.Printf("\n💡 Optimization suggestions:\n")
	fmt.Printf("   • Use 'claude-wm-cli metrics steps <command>' for detailed analysis\n")
	fmt.Printf("   • Focus on commands with highest execution count and duration\n")
	
	return nil
}

// showProjectComparison displays performance comparison across projects  
func showProjectComparison(days int) error {
	collector := metrics.GetCollector()
	if !collector.IsEnabled() {
		return fmt.Errorf("metrics collection is disabled")
	}
	
	fmt.Printf("📊 Project Performance Comparison (last %d days)\n", days)
	fmt.Printf("==============================================\n\n")
	
	projects, err := collector.GetProjectComparison(days)
	if err != nil {
		return fmt.Errorf("failed to get project comparison: %w", err)
	}
	
	if len(projects) == 0 {
		fmt.Printf("📊 No project data available\n")
		return nil
	}
	
	// Sort by average duration (slowest first)
	sort.Slice(projects, func(i, j int) bool {
		return projects[i].AvgDuration > projects[j].AvgDuration
	})
	
	// Create table
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "PROJECT\tCOMMANDS\tAVG TIME\tMAX TIME\tCOMPLEXITY\n")
	fmt.Fprintf(w, "───────\t────────\t────────\t────────\t──────────\n")
	
	for _, project := range projects {
		complexity := getComplexityIcon(project.TotalCommands, project.AvgDuration)
		fmt.Fprintf(w, "%s\t%d\t%.0fms\t%.0fms\t%s\n",
			truncateMetricsString(project.ProjectName, 30),
			project.TotalCommands,
			project.AvgDuration,
			project.MaxDuration,
			complexity)
	}
	
	w.Flush()
	
	return nil
}

// cleanMetrics cleans old metrics data
func cleanMetrics(force bool, olderThanDays int) error {
	if !force {
		fmt.Printf("⚠️  This will delete metrics data older than %d days\n", olderThanDays)
		fmt.Printf("   Use --force to confirm deletion\n")
		return nil
	}
	
	fmt.Printf("🧹 Cleaning metrics data older than %d days...\n", olderThanDays)
	fmt.Printf("   (Implementation pending)\n")
	
	return nil
}

// Helper functions

func getDatabasePath() string {
	homeDir, _ := os.UserHomeDir()
	return fmt.Sprintf("%s/.claude-wm/metrics/performance.db", homeDir)
}

func getImpactIcon(percentage float64) string {
	if percentage > 50 {
		return "🔴" // High impact
	} else if percentage > 25 {
		return "🟡" // Medium impact  
	} else if percentage > 10 {
		return "🟢" // Low impact
	}
	return "⚪" // Minimal impact
}

func getSeverityIcon(avgDurationMs float64) string {
	seconds := avgDurationMs / 1000.0
	if seconds > 30 {
		return "🔴 CRITICAL"
	} else if seconds > 15 {
		return "🟠 HIGH"
	} else if seconds > 5 {
		return "🟡 MEDIUM"
	}
	return "🟢 LOW"
}

func getComplexityIcon(commandCount int, avgDuration float64) string {
	score := float64(commandCount) * (avgDuration / 1000.0)
	if score > 1000 {
		return "🔴 HIGH"
	} else if score > 300 {
		return "🟡 MEDIUM"
	}
	return "🟢 LOW" 
}

func getPerformanceIcon(avgDurationMs float64) string {
	seconds := avgDurationMs / 1000.0
	if seconds < 1.0 {
		return "⚡ FAST"
	} else if seconds < 5.0 {
		return "✅ GOOD"
	} else if seconds < 15.0 {
		return "⚠️  SLOW"
	}
	return "🐌 VERY SLOW"
}

func truncateMetricsString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}