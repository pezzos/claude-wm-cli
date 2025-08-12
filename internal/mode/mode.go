package mode

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// Env represents the environment context for mode detection
type Env struct {
	Root string // root directory of the current repository (.)
}

// lastResult caches the most recent detection result to support Reason() function
var lastResult struct {
	env    Env
	result bool
	reason string
}

// Self returns true if the current repository is in SELF mode (claude-wm-cli project itself)
// Uses OR logic across three detection criteria:
// 1. Presence of WM_SELF file at root
// 2. Environment variable CLAUDE_WM_SELF=1
// 3. Heuristic: presence of internal/config/system/ AND go.mod with claude-wm-cli module
func Self(e Env) bool {
	// Criterion 1: Check for WM_SELF file
	wmSelfPath := filepath.Join(e.Root, "WM_SELF")
	if _, err := os.Stat(wmSelfPath); err == nil {
		lastResult = struct {
			env    Env
			result bool
			reason string
		}{e, true, "WM_SELF file found at root"}
		return true
	}

	// Criterion 2: Check environment variable
	if os.Getenv("CLAUDE_WM_SELF") == "1" {
		lastResult = struct {
			env    Env
			result bool
			reason string
		}{e, true, "CLAUDE_WM_SELF=1 environment variable set"}
		return true
	}

	// Criterion 3: Check heuristic (internal/config/system + go.mod with claude-wm-cli)
	if checkHeuristic(e) {
		lastResult = struct {
			env    Env
			result bool
			reason string
		}{e, true, "heuristic: internal/config/system/ directory and go.mod with claude-wm-cli module found"}
		return true
	}

	// No criteria matched
	lastResult = struct {
		env    Env
		result bool
		reason string
	}{e, false, "no criteria matched: WM_SELF file not found, CLAUDE_WM_SELF!=1, heuristic failed"}
	return false
}

// Reason returns a debug string explaining why Self() returned true or false
// If called with a different Env than the last Self() call, it will recalculate
func Reason(e Env) string {
	if lastResult.env.Root != e.Root {
		// If called with different Env, recalculate
		Self(e)
	}
	return lastResult.reason
}

// checkHeuristic implements the heuristic detection:
// - internal/config/system/ directory must exist
// - go.mod file must exist with module name containing claude-wm-cli
func checkHeuristic(e Env) bool {
	// Check if internal/config/system directory exists
	systemDir := filepath.Join(e.Root, "internal", "config", "system")
	if _, err := os.Stat(systemDir); err != nil {
		return false // internal/config/system not found
	}

	// Check if go.mod exists and contains claude-wm-cli module
	goModPath := filepath.Join(e.Root, "go.mod")
	file, err := os.Open(goModPath)
	if err != nil {
		return false // go.mod not found or cannot be read
	}
	defer file.Close()

	// Parse go.mod to find module declaration
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "module ") {
			moduleName := strings.TrimSpace(strings.TrimPrefix(line, "module"))
			// Check if module name contains or ends with claude-wm-cli
			return strings.Contains(moduleName, "claude-wm-cli") || strings.HasSuffix(moduleName, "claude-wm-cli")
		}
	}

	return false // No matching module declaration found
}