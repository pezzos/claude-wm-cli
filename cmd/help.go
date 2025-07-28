package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// Custom help template with better formatting
const helpTemplate = `{{with (or .Long .Short)}}{{. | trimTrailingWhitespaces}}

{{end}}{{if or .Runnable .HasSubCommands}}{{.UsageString}}{{end}}`

// Custom usage template with examples
const usageTemplate = `USAGE:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

ALIASES:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

EXAMPLES:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}

AVAILABLE COMMANDS:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

FLAGS:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

GLOBAL FLAGS:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

ADDITIONAL HELP TOPICS:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`

// Initialize custom help templates
func init() {
	// Set custom templates
	rootCmd.SetHelpTemplate(helpTemplate)
	rootCmd.SetUsageTemplate(usageTemplate)

	// Add completion command help
	completionCmd := rootCmd.Commands()
	for _, cmd := range completionCmd {
		if cmd.Name() == "completion" {
			cmd.Long = `Generate shell completion scripts for claude-wm-cli.

The completion script needs to be sourced to enable completions.

BASH:
  # Install bash completion permanently:
  claude-wm-cli completion bash > /etc/bash_completion.d/claude-wm-cli
  
  # Or source it in your current session:
  source <(claude-wm-cli completion bash)

ZSH:
  # Add to your ~/.zshrc:
  autoload -U compinit; compinit
  source <(claude-wm-cli completion zsh)`
		}
	}
}

// helpCmd provides enhanced help functionality
var helpCmd = &cobra.Command{
	Use:   "help [command]",
	Short: "Help about any command",
	Long: `Get detailed help about claude-wm-cli commands and usage.

This command provides comprehensive help including examples, 
configuration options, and troubleshooting tips.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			rootCmd.Help()
			return
		}

		// Find the command and show its help
		cmdName := args[0]
		targetCmd, _, err := rootCmd.Find(args)
		if err != nil {
			fmt.Printf("❌ Unknown command: %s\n\n", cmdName)
			suggestSimilarCommands(cmdName)
			return
		}

		targetCmd.Help()
	},
}

// suggestSimilarCommands suggests commands similar to the mistyped one
func suggestSimilarCommands(input string) {
	fmt.Println("💡 Did you mean one of these?")

	commands := []string{"status", "execute", "init", "help", "completion"}
	suggestions := findSimilarCommands(input, commands)

	if len(suggestions) == 0 {
		fmt.Println("   No similar commands found.")
		fmt.Println("\n📋 Available commands:")
		for _, cmd := range commands {
			fmt.Printf("   %s\n", cmd)
		}
	} else {
		for _, suggestion := range suggestions {
			fmt.Printf("   %s\n", suggestion)
		}
	}

	fmt.Println("\n📖 Use 'claude-wm-cli help' to see all available commands.")
}

// findSimilarCommands finds commands similar to the input using basic string matching
func findSimilarCommands(input string, commands []string) []string {
	var suggestions []string
	input = strings.ToLower(input)

	// First pass: exact substring matches
	for _, cmd := range commands {
		if strings.Contains(strings.ToLower(cmd), input) || strings.Contains(input, strings.ToLower(cmd)) {
			suggestions = append(suggestions, cmd)
		}
	}

	// Second pass: similar starting characters (if no exact matches)
	if len(suggestions) == 0 && len(input) > 0 {
		firstChar := input[0:1]
		for _, cmd := range commands {
			if strings.HasPrefix(strings.ToLower(cmd), firstChar) {
				suggestions = append(suggestions, cmd)
			}
		}
	}

	return suggestions
}

// printQuickStart shows a quick start guide
func printQuickStart() {
	fmt.Println(`🚀 QUICK START GUIDE

1. Initialize a new project:
   claude-wm-cli init my-project

2. Check project status:
   claude-wm-cli status

3. Execute Claude commands:
   claude-wm-cli execute "claude --help"

4. Get help on any command:
   claude-wm-cli help [command]

📖 For detailed documentation, see: docs/README.md`)
}

func init() {
	rootCmd.AddCommand(helpCmd)
}
