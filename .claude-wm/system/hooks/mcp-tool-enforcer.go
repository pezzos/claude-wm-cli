package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"

	"claude-hooks-orchestrator/patterns"
)

// ToolInput represents the input structure for Claude Code tools
type ToolInput struct {
	ToolName  string          `json:"tool_name"`
	ToolInput json.RawMessage `json:"tool_input"`
}

// BashInput represents Bash tool input
type BashInput struct {
	Command string `json:"command"`
}

// FileInput represents file operation tool input
type FileInput struct {
	FilePath string `json:"file_path"`
	Content  string `json:"content"`
}

// EditInput represents Edit tool input
type EditInput struct {
	FilePath  string `json:"file_path"`
	NewString string `json:"new_string"`
}

// MultiEditInput represents MultiEdit tool input
type MultiEditInput struct {
	FilePath string `json:"file_path"`
	Edits    []struct {
		NewString string `json:"new_string"`
	} `json:"edits"`
}

// MCPAlternative represents an MCP tool alternative
type MCPAlternative struct {
	Patterns    []string `json:"patterns"`
	MCPTools    []string `json:"mcp_tools"`
	Message     string   `json:"message"`
	CompiledPat []*regexp.Regexp
}

// Suggestion represents a tool suggestion
type Suggestion struct {
	Message      string   `json:"message"`
	Alternatives []string `json:"alternatives"`
}

// Output represents the final output structure
type Output struct {
	Decision    string       `json:"decision"`
	Reason      string       `json:"reason"`
	Suggestions []Suggestion `json:"suggestions"`
}

// MCPEnforcer manages MCP tool enforcement with compiled rules
type MCPEnforcer struct {
	alternatives map[string]*MCPAlternative
	fileMutex    sync.RWMutex
	initOnce     sync.Once
}

// Global enforcer instance
var enforcer *MCPEnforcer

// Initialize the enforcer with compiled patterns
func (e *MCPEnforcer) init() {
	e.alternatives = map[string]*MCPAlternative{
		"WebFetch": {
			Patterns: []string{
				`\b(fetch|curl|wget|axios|requests)\b`,
				`\b(http\.get|http\.post|urllib)\b`,
				`\b(XMLHttpRequest|fetch\()\b`,
			},
			MCPTools: []string{"mcp__fetch__fetch", "mcp__browser__navigate"},
			Message:  "Consider using MCP fetch tools instead of direct web requests",
		},
		"WebSearch": {
			Patterns: []string{
				`\b(google|search|bing|duckduckgo)\b`,
				`\b(search_web|web_search)\b`,
			},
			MCPTools: []string{"mcp__search__search"},
			Message:  "Consider using MCP search tools for web searches",
		},
		"Database": {
			Patterns: []string{
				`\b(mysql|postgres|sqlite|mongodb)\b`,
				`\b(SELECT|INSERT|UPDATE|DELETE)\b`,
				`\b(db\.query|database\.execute)\b`,
			},
			MCPTools: []string{"mcp__database__query", "mcp__sqlite__query"},
			Message:  "Consider using MCP database tools for database operations",
		},
		"FileSystem": {
			Patterns: []string{
				`\bfs\.(read|write|mkdir|rmdir)\b`,
				`\b(readFile|writeFile|mkdir|rmdir)\b`,
				`\b(os\.open|open\()\b`,
			},
			MCPTools: []string{"mcp__filesystem__read", "mcp__filesystem__write"},
			Message:  "Consider using MCP filesystem tools for file operations",
		},
		"Memory": {
			Patterns: []string{
				`\b(localStorage|sessionStorage|cache|store)\b`,
				`\b(memory|persist|save_state)\b`,
			},
			MCPTools: []string{"mcp__mem0__add_coding_preference", "mcp__mem0__search_coding_preference"},
			Message:  "Consider using MCP memory tools for data persistence",
		},
	}

	// Compile all regex patterns using patterns helper
	for category, alternative := range e.alternatives {
		alternative.CompiledPat = make([]*regexp.Regexp, len(alternative.Patterns))
		for i, pattern := range alternative.Patterns {
			// Use the patterns helper for case-insensitive compilation
			compiled, err := patterns.CompilePattern(`(?i)` + pattern)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error compiling pattern %s: %v\n", pattern, err)
				continue
			}
			alternative.CompiledPat[i] = compiled
		}
		e.alternatives[category] = alternative
	}
}

// checkBashCommand checks if a bash command could use MCP tools
func (e *MCPEnforcer) checkBashCommand(command string) []Suggestion {
	var suggestions []Suggestion
	seen := make(map[string]bool)

	for category, alternative := range e.alternatives {
		if seen[category] {
			continue
		}

		for _, pattern := range alternative.CompiledPat {
			if pattern != nil && pattern.MatchString(command) {
				suggestions = append(suggestions, Suggestion{
					Message:      alternative.Message,
					Alternatives: alternative.MCPTools,
				})
				seen[category] = true
				break
			}
		}
	}

	return suggestions
}

// checkCodeContent checks code content for operations that could use MCP tools
func (e *MCPEnforcer) checkCodeContent(content, fileType string) []Suggestion {
	var suggestions []Suggestion

	// Skip configuration files
	if isConfigFile(fileType) {
		return suggestions
	}

	// Use AST parsing for supported languages
	if fileType == "go" {
		return e.checkGoAST(content)
	}

	// Fallback to regex for other languages
	return e.checkContentRegex(content)
}

// checkGoAST uses Go AST parsing to analyze Go code
func (e *MCPEnforcer) checkGoAST(content string) []Suggestion {
	var suggestions []Suggestion

	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, "", content, parser.ParseComments)
	if err != nil {
		// Fallback to regex if AST parsing fails
		return e.checkContentRegex(content)
	}

	seen := make(map[string]bool)

	ast.Inspect(node, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.CallExpr:
			if ident, ok := x.Fun.(*ast.Ident); ok {
				funcName := ident.Name
				if suggestion := e.getSuggestionForFunction(funcName, seen); suggestion != nil {
					suggestions = append(suggestions, *suggestion)
				}
			}
		case *ast.SelectorExpr:
			if ident, ok := x.X.(*ast.Ident); ok {
				pkgName := ident.Name
				selector := x.Sel.Name
				fullName := pkgName + "." + selector
				if suggestion := e.getSuggestionForFunction(fullName, seen); suggestion != nil {
					suggestions = append(suggestions, *suggestion)
				}
			}
		}
		return true
	})

	return suggestions
}

// getSuggestionForFunction returns suggestion for a function call
func (e *MCPEnforcer) getSuggestionForFunction(funcName string, seen map[string]bool) *Suggestion {
	for category, alternative := range e.alternatives {
		if seen[category] {
			continue
		}

		for _, pattern := range alternative.CompiledPat {
			if pattern != nil && pattern.MatchString(funcName) {
				seen[category] = true
				return &Suggestion{
					Message:      alternative.Message,
					Alternatives: alternative.MCPTools,
				}
			}
		}
	}
	return nil
}

// checkContentRegex uses regex to analyze content for non-Go languages
func (e *MCPEnforcer) checkContentRegex(content string) []Suggestion {
	var suggestions []Suggestion
	seen := make(map[string]bool)

	lines := strings.Split(content, "\n")

	for category, alternative := range e.alternatives {
		if seen[category] {
			continue
		}

		for _, pattern := range alternative.CompiledPat {
			if pattern == nil {
				continue
			}

			// Check each line, but skip comments
			for _, line := range lines {
				trimmed := strings.TrimSpace(line)
				if trimmed == "" || isCommentLine(trimmed) {
					continue
				}

				if pattern.MatchString(line) {
					suggestions = append(suggestions, Suggestion{
						Message:      alternative.Message,
						Alternatives: alternative.MCPTools,
					})
					seen[category] = true
					break
				}
			}
		}
	}

	return suggestions
}

// isConfigFile checks if the file type is a configuration file
func isConfigFile(fileType string) bool {
	configTypes := []string{"json", "yaml", "yml", "toml", "ini", "conf", "config"}
	for _, configType := range configTypes {
		if fileType == configType {
			return true
		}
	}
	return false
}

// isCommentLine checks if a line is a comment
func isCommentLine(line string) bool {
	return strings.HasPrefix(line, "//") ||
		strings.HasPrefix(line, "#") ||
		strings.HasPrefix(line, "/*") ||
		strings.HasPrefix(line, "*") ||
		strings.HasPrefix(line, "<!--")
}

// processInput processes the input and returns suggestions
func (e *MCPEnforcer) processInput(input *ToolInput) []Suggestion {
	var suggestions []Suggestion

	switch input.ToolName {
	case "Bash":
		var bashInput BashInput
		if err := json.Unmarshal(input.ToolInput, &bashInput); err != nil {
			return suggestions
		}
		suggestions = e.checkBashCommand(bashInput.Command)

	case "Write":
		var fileInput FileInput
		if err := json.Unmarshal(input.ToolInput, &fileInput); err != nil {
			return suggestions
		}
		fileType := getFileType(fileInput.FilePath)
		suggestions = e.checkCodeContent(fileInput.Content, fileType)

	case "Edit":
		var editInput EditInput
		if err := json.Unmarshal(input.ToolInput, &editInput); err != nil {
			return suggestions
		}
		fileType := getFileType(editInput.FilePath)
		suggestions = e.checkCodeContent(editInput.NewString, fileType)

	case "MultiEdit":
		var multiEditInput MultiEditInput
		if err := json.Unmarshal(input.ToolInput, &multiEditInput); err != nil {
			return suggestions
		}
		fileType := getFileType(multiEditInput.FilePath)
		
		var content strings.Builder
		for _, edit := range multiEditInput.Edits {
			content.WriteString(edit.NewString)
			content.WriteString("\n")
		}
		suggestions = e.checkCodeContent(content.String(), fileType)
	}

	return suggestions
}

// getFileType extracts file type from path
func getFileType(filePath string) string {
	parts := strings.Split(filePath, ".")
	if len(parts) > 1 {
		return strings.ToLower(parts[len(parts)-1])
	}
	return ""
}

// formatOutput formats suggestions into the output structure
func formatOutput(suggestions []Suggestion) *Output {
	if len(suggestions) == 0 {
		return nil
	}

	return &Output{
		Decision:    "approve",
		Reason:      "MCP tool alternatives available",
		Suggestions: suggestions,
	}
}

// processConcurrently processes multiple inputs concurrently
func (e *MCPEnforcer) processConcurrently(inputs []*ToolInput) []Suggestion {
	numWorkers := runtime.NumCPU()
	if numWorkers > len(inputs) {
		numWorkers = len(inputs)
	}

	inputChan := make(chan *ToolInput, len(inputs))
	resultChan := make(chan []Suggestion, len(inputs))

	// Start workers
	for i := 0; i < numWorkers; i++ {
		go func() {
			for input := range inputChan {
				suggestions := e.processInput(input)
				resultChan <- suggestions
			}
		}()
	}

	// Send inputs to workers
	for _, input := range inputs {
		inputChan <- input
	}
	close(inputChan)

	// Collect results
	var allSuggestions []Suggestion
	for i := 0; i < len(inputs); i++ {
		suggestions := <-resultChan
		allSuggestions = append(allSuggestions, suggestions...)
	}

	return allSuggestions
}

// main function
func main() {
	start := time.Now()
	
	// Initialize enforcer
	enforcer = &MCPEnforcer{}
	enforcer.initOnce.Do(func() {
		enforcer.init()
	})

	// Read input from stdin
	var input ToolInput
	if err := json.NewDecoder(os.Stdin).Decode(&input); err != nil {
		fmt.Fprintf(os.Stderr, "Error: Invalid JSON input: %v\n", err)
		os.Exit(1)
	}

	// Process input
	suggestions := enforcer.processInput(&input)

	// Format and output suggestions
	if len(suggestions) > 0 {
		output := formatOutput(suggestions)
		if output != nil {
			// Output structured JSON for Claude to process
			jsonOutput, err := json.Marshal(output)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error marshaling output: %v\n", err)
				os.Exit(1)
			}
			fmt.Println(string(jsonOutput))

			// Also log suggestions to stderr for visibility
			fmt.Fprintf(os.Stderr, "\nMCP Tool Suggestions:\n")
			for _, suggestion := range output.Suggestions {
				fmt.Fprintf(os.Stderr, "💡 %s\n", suggestion.Message)
				fmt.Fprintf(os.Stderr, "   Available tools: %s\n", strings.Join(suggestion.Alternatives, ", "))
			}
		}
	}

	// Log performance info to stderr
	duration := time.Since(start)
	fmt.Fprintf(os.Stderr, "MCP enforcer completed in %v\n", duration)
}