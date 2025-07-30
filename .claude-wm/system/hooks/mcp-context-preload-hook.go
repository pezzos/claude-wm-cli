package main

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type MCPCache struct {
	Key       string    `json:"key"`
	Command   string    `json:"command"`
	Result    string    `json:"result"`
	Timestamp time.Time `json:"timestamp"`
	TTL       int       `json:"ttl_hours"`
}

type PreloadContext struct {
	Command     string            `json:"command"`
	ProjectType string            `json:"project_type"`
	Context     map[string]string `json:"context"`
}

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Usage: %s <tool_name> <tool_params>\n", os.Args[0])
		os.Exit(1)
	}

	toolName := os.Args[1]
	toolParams := strings.Join(os.Args[2:], " ")

	// Only handle Task tool calls that might benefit from MCP preloading
	if toolName != "Task" {
		os.Exit(0)
	}

	// Detect command context from tool parameters
	context := detectCommandContext(toolParams)
	if context == nil {
		// No preloading needed
		os.Exit(0)
	}

	// Execute MCP preloading
	err := preloadMCPContext(context)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: MCP preload failed: %v\n", err)
		// Don't fail the hook, just warn
		os.Exit(0)
	}

	fmt.Printf("✅ MCP context preloaded for %s\n", context.Command)
}

func detectCommandContext(params string) *PreloadContext {
	// Patterns that benefit from MCP preloading
	patterns := map[string]*PreloadContext{
		`2-Challenge|Challenge`: {
			Command:     "challenge",
			ProjectType: detectProjectType(),
			Context: map[string]string{
				"mem0_query":     "similar challenges and outcomes",
				"consult7_pattern": `".*\.(js|ts|py|java|md)$"`,
				"context7_topic":   "project documentation best practices",
			},
		},
		`Plan-stories|2-epic:1-start:2-Plan-stories`: {
			Command:     "plan-stories", 
			ProjectType: detectProjectType(),
			Context: map[string]string{
				"mem0_query":       "story planning patterns",
				"consult7_pattern": `".*\.(js|ts|py|java)$"`,
				"context7_topic":   "agile story writing",
			},
		},
		`3-Implement|2-execute:3-Implement`: {
			Command:     "implement",
			ProjectType: detectProjectType(),
			Context: map[string]string{
				"mem0_query":       "implementation patterns",
				"consult7_pattern": `".*\.(js|ts|py|java)$"`, 
				"context7_topic":   "coding best practices",
			},
		},
		`Architecture-Review|1-Architecture-Review`: {
			Command:     "architecture-review",
			ProjectType: detectProjectType(),
			Context: map[string]string{
				"mem0_query":       "architecture analysis patterns",
				"consult7_pattern": `".*\.(js|ts|py|java|md)$"`,
				"context7_topic":   "software architecture",
			},
		},
	}

	for pattern, context := range patterns {
		matched, _ := regexp.MatchString(pattern, params)
		if matched {
			return context
		}
	}

	return nil
}

func preloadMCPContext(context *PreloadContext) error {
	cacheDir := "hooks/cache"
	err := os.MkdirAll(cacheDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create cache directory: %v", err)
	}

	// Preload mem0 search if needed
	if query, ok := context.Context["mem0_query"]; ok && query != "" {
		err := preloadMem0Search(cacheDir, query, context.Command)
		if err != nil {
			fmt.Printf("Warning: mem0 preload failed: %v\n", err)
		}
	}

	// Preload consult7 analysis if needed
	if pattern, ok := context.Context["consult7_pattern"]; ok && pattern != "" {
		err := preloadConsult7Analysis(cacheDir, pattern, context.Command)
		if err != nil {
			fmt.Printf("Warning: consult7 preload failed: %v\n", err)
		}
	}

	// Preload context7 documentation if needed
	if topic, ok := context.Context["context7_topic"]; ok && topic != "" {
		err := preloadContext7Docs(cacheDir, topic, context.Command)
		if err != nil {
			fmt.Printf("Warning: context7 preload failed: %v\n", err)
		}
	}

	return nil
}

func preloadMem0Search(cacheDir, query, command string) error {
	// Create cache key
	key := createCacheKey(fmt.Sprintf("mem0_%s_%s", command, query))
	cachePath := filepath.Join(cacheDir, fmt.Sprintf("mem0_%s.json", key))

	// Check if cache exists and is valid
	if isCacheValid(cachePath, 24) { // 24 hour TTL
		return nil
	}

	// Execute mem0 search (simulated - in real implementation would call claude with mcp__mem0__search_coding_preferences)
	result, err := executeMCPCommand("mcp__mem0__search_coding_preferences", map[string]string{
		"query": query,
	})
	if err != nil {
		return fmt.Errorf("mem0 search failed: %v", err)
	}

	// Cache result
	cache := MCPCache{
		Key:       key,
		Command:   "mem0_search",
		Result:    result,
		Timestamp: time.Now(),
		TTL:       24,
	}

	return saveCacheEntry(cachePath, &cache)
}

func preloadConsult7Analysis(cacheDir, pattern, command string) error {
	// Get current working directory for analysis
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %v", err)
	}

	// Create cache key
	key := createCacheKey(fmt.Sprintf("consult7_%s_%s_%s", command, cwd, pattern))
	cachePath := filepath.Join(cacheDir, fmt.Sprintf("consult7_%s.json", key))

	// Check if cache exists and is valid
	if isCacheValid(cachePath, 12) { // 12 hour TTL
		return nil
	}

	// Execute consult7 analysis (simulated - in real implementation would call claude with mcp__consult7__consultation)
	result, err := executeMCPCommand("mcp__consult7__consultation", map[string]string{
		"path":    cwd,
		"pattern": pattern,
		"query":   fmt.Sprintf("Analyze codebase for %s command", command),
	})
	if err != nil {
		return fmt.Errorf("consult7 analysis failed: %v", err)
	}

	// Cache result
	cache := MCPCache{
		Key:       key,
		Command:   "consult7_analysis",
		Result:    result,
		Timestamp: time.Now(),
		TTL:       12,
	}

	return saveCacheEntry(cachePath, &cache)
}

func preloadContext7Docs(cacheDir, topic, command string) error {
	// Detect relevant libraries based on project type
	libraries := detectRelevantLibraries()
	
	for _, library := range libraries {
		// Create cache key
		key := createCacheKey(fmt.Sprintf("context7_%s_%s_%s", command, library, topic))
		cachePath := filepath.Join(cacheDir, fmt.Sprintf("context7_%s.json", key))

		// Check if cache exists and is valid
		if isCacheValid(cachePath, 168) { // 1 week TTL for docs
			continue
		}

		// Execute context7 lookup (simulated - in real implementation would call claude with mcp__context7__)
		result, err := executeMCPCommand("mcp__context7__get-library-docs", map[string]string{
			"context7CompatibleLibraryID": library,
			"topic": topic,
		})
		if err != nil {
			fmt.Printf("Warning: context7 lookup failed for %s: %v\n", library, err)
			continue
		}

		// Cache result
		cache := MCPCache{
			Key:       key,
			Command:   "context7_docs",
			Result:    result,
			Timestamp: time.Now(),
			TTL:       168,
		}

		err = saveCacheEntry(cachePath, &cache)
		if err != nil {
			fmt.Printf("Warning: failed to cache context7 result for %s: %v\n", library, err)
		}
	}

	return nil
}

func detectProjectType() string {
	// Check for common project files
	projectFiles := map[string]string{
		"package.json":      "javascript",
		"requirements.txt":  "python",
		"Cargo.toml":       "rust", 
		"go.mod":           "go",
		"pom.xml":          "java",
		"composer.json":    "php",
	}

	for file, projectType := range projectFiles {
		if _, err := os.Stat(file); err == nil {
			return projectType
		}
	}

	return "unknown"
}

func detectRelevantLibraries() []string {
	libraries := []string{}
	
	// Check package.json for JavaScript/Node.js projects
	if packageData, err := os.ReadFile("package.json"); err == nil {
		var packageJSON map[string]interface{}
		if json.Unmarshal(packageData, &packageJSON) == nil {
			// Check for common frameworks
			deps := make(map[string]interface{})
			if dependencies, ok := packageJSON["dependencies"].(map[string]interface{}); ok {
				for k, v := range dependencies {
					deps[k] = v
				}
			}
			if devDependencies, ok := packageJSON["devDependencies"].(map[string]interface{}); ok {
				for k, v := range devDependencies {
					deps[k] = v
				}
			}
			
			// Map dependencies to context7 library IDs
			if _, hasReact := deps["react"]; hasReact {
				libraries = append(libraries, "/facebook/react")
			}
			if _, hasNext := deps["next"]; hasNext {
				libraries = append(libraries, "/vercel/next.js")
			}
			if _, hasVue := deps["vue"]; hasVue {
				libraries = append(libraries, "/vuejs/vue")
			}
		}
	}

	// Check requirements.txt for Python projects
	if reqData, err := os.ReadFile("requirements.txt"); err == nil {
		reqText := string(reqData)
		if strings.Contains(reqText, "fastapi") {
			libraries = append(libraries, "/tiangolo/fastapi")
		}
		if strings.Contains(reqText, "flask") {
			libraries = append(libraries, "/pallets/flask")
		}
		if strings.Contains(reqText, "django") {
			libraries = append(libraries, "/django/django")
		}
	}

	// Default to common libraries if none detected
	if len(libraries) == 0 {
		libraries = append(libraries, "/nodejs/node") // Generic Node.js docs
	}

	return libraries
}

func executeMCPCommand(command string, params map[string]string) (string, error) {
	// This is a simplified simulation of MCP command execution
	// In a real implementation, this would integrate with Claude Code's MCP system
	
	// For now, return a placeholder result
	switch command {
	case "mcp__mem0__search_coding_preferences":
		return fmt.Sprintf("Found 3 similar patterns for query: %s", params["query"]), nil
	case "mcp__consult7__consultation":
		return fmt.Sprintf("Analyzed %s with pattern %s - found 15 files", params["path"], params["pattern"]), nil
	case "mcp__context7__get-library-docs":
		return fmt.Sprintf("Retrieved docs for %s on topic: %s", params["context7CompatibleLibraryID"], params["topic"]), nil
	default:
		return "", fmt.Errorf("unknown MCP command: %s", command)
	}
}

func createCacheKey(input string) string {
	hash := md5.Sum([]byte(input))
	return fmt.Sprintf("%x", hash)[:8]
}

func isCacheValid(cachePath string, ttlHours int) bool {
	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		return false
	}

	data, err := os.ReadFile(cachePath)
	if err != nil {
		return false
	}

	var cache MCPCache
	if json.Unmarshal(data, &cache) != nil {
		return false
	}

	// Check if cache is still valid
	expiry := cache.Timestamp.Add(time.Duration(ttlHours) * time.Hour)
	return time.Now().Before(expiry)
}

func saveCacheEntry(cachePath string, cache *MCPCache) error {
	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal cache entry: %v", err)
	}

	err = os.WriteFile(cachePath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write cache file: %v", err)
	}

	return nil
}