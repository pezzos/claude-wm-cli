package debug

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// DebugEnabled indicates if debug mode is enabled
var DebugEnabled bool

// DevMode indicates if we're in development mode (disables timeouts)
var DevMode = true // Set to true for development

// SetDebugMode enables or disables debug mode
func SetDebugMode(enabled bool) {
	DebugEnabled = enabled
}

// LogCommand logs a command that is about to be executed
func LogCommand(category, description, fullCommand string) {
	if !DebugEnabled {
		return
	}
	
	timestamp := time.Now().Format("15:04:05.000")
	fmt.Fprintf(os.Stderr, "🔍 [%s] DEBUG [%s]: %s\n", timestamp, category, description)
	fmt.Fprintf(os.Stderr, "   ↳ Command: %s\n", fullCommand)
}

// LogCommandWithArgs logs a command with its arguments separately
func LogCommandWithArgs(category, description, command string, args []string) {
	if !DebugEnabled {
		return
	}
	
	timestamp := time.Now().Format("15:04:05.000")
	fmt.Fprintf(os.Stderr, "🔍 [%s] DEBUG [%s]: %s\n", timestamp, category, description)
	fmt.Fprintf(os.Stderr, "   ↳ Command: %s\n", command)
	if len(args) > 0 {
		fmt.Fprintf(os.Stderr, "   ↳ Args: [%s]\n", strings.Join(args, ", "))
	}
}

// LogClaudeCommand specifically logs Claude command executions
func LogClaudeCommand(prompt, description string) {
	if !DebugEnabled {
		return
	}
	
	timestamp := time.Now().Format("15:04:05.000")
	fmt.Fprintf(os.Stderr, "🤖 [%s] DEBUG [CLAUDE]: %s\n", timestamp, description)
	fmt.Fprintf(os.Stderr, "   ↳ Prompt: %s\n", prompt)
	fmt.Fprintf(os.Stderr, "   ↳ Full Command: claude -p \"%s\"\n", prompt)
}

// LogExecution logs the start and expected behavior of a command
func LogExecution(category, action, expectedBehavior string) {
	if !DebugEnabled {
		return
	}
	
	timestamp := time.Now().Format("15:04:05.000")
	fmt.Fprintf(os.Stderr, "⚡ [%s] DEBUG [%s]: Starting %s\n", timestamp, category, action)
	fmt.Fprintf(os.Stderr, "   ↳ Expected: %s\n", expectedBehavior)
}

// LogResult logs the result of a command execution
func LogResult(category, action, result string, success bool) {
	if !DebugEnabled {
		return
	}
	
	timestamp := time.Now().Format("15:04:05.000")
	status := "✅"
	if !success {
		status = "❌"
	}
	
	fmt.Fprintf(os.Stderr, "%s [%s] DEBUG [%s]: %s completed\n", status, timestamp, category, action)
	fmt.Fprintf(os.Stderr, "   ↳ Result: %s\n", result)
}

// LogStub logs when a stub function is called (should not happen in production)
func LogStub(category, functionName, shouldDo string) {
	if !DebugEnabled {
		return
	}
	
	timestamp := time.Now().Format("15:04:05.000")
	fmt.Fprintf(os.Stderr, "🚨 [%s] DEBUG [%s]: STUB CALLED: %s\n", timestamp, category, functionName)
	fmt.Fprintf(os.Stderr, "   ↳ Should do: %s\n", shouldDo)
	fmt.Fprintf(os.Stderr, "   ↳ Current: Does nothing (stub implementation)\n")
}