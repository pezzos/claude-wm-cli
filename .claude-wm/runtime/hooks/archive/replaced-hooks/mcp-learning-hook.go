package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type ConversationContext struct {
	SessionId      string    `json:"session_id"`
	Timestamp      time.Time `json:"timestamp"`
	ProjectInfo    ProjectInfo `json:"project_info"`
	ActionsPerformed []ActionInfo `json:"actions_performed"`
	TechnicalPatterns []TechnicalPattern `json:"technical_patterns"`
	LearningInsights []LearningInsight `json:"learning_insights"`
}

type ProjectInfo struct {
	WorkingDirectory string   `json:"working_directory"`
	GitBranch        string   `json:"git_branch,omitempty"`
	ProjectType      string   `json:"project_type,omitempty"`
	Languages        []string `json:"languages,omitempty"`
	Frameworks       []string `json:"frameworks,omitempty"`
}

type ActionInfo struct {
	Type        string    `json:"type"`
	Target      string    `json:"target,omitempty"`
	Description string    `json:"description"`
	Timestamp   time.Time `json:"timestamp"`
	Success     bool      `json:"success"`
}

type TechnicalPattern struct {
	Pattern     string `json:"pattern"`
	Context     string `json:"context"`
	Frequency   int    `json:"frequency"`
	Effectiveness string `json:"effectiveness"`
}

type LearningInsight struct {
	Type        string `json:"type"` // success_pattern, failure_pattern, best_practice, optimization
	Title       string `json:"title"`
	Description string `json:"description"`
	Context     string `json:"context"`
	Reusability string `json:"reusability"`
	Quality     int    `json:"quality"` // 1-10 rating
}

func main() {
	// Collect conversation context
	context, err := collectConversationContext()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error collecting context: %v\n", err)
		os.Exit(0) // Non-blocking: don't fail the main operation
	}

	// Generate learning insights
	insights, err := generateLearningInsights(context)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating insights: %v\n", err)
		os.Exit(0) // Non-blocking
	}
	context.LearningInsights = insights

	// Save to mem0 if context contains valuable information
	if shouldSaveToMem0(context) {
		err = saveToMem0(context)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to save to mem0: %v\n", err)
		} else {
			fmt.Printf("💡 Learning insights saved to mem0\n")
		}
	}

	// Save to local learning cache
	err = saveToLocalCache(context)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to save to local cache: %v\n", err)
	}
}

func collectConversationContext() (*ConversationContext, error) {
	context := &ConversationContext{
		SessionId: generateSessionId(),
		Timestamp: time.Now(),
	}

	// Get project info
	projectInfo, err := getProjectInfo()
	if err != nil {
		// Continue without project info if error
		projectInfo = &ProjectInfo{WorkingDirectory: getCurrentDir()}
	}
	context.ProjectInfo = *projectInfo

	// Detect actions performed (based on recent file changes and git history)
	actions, err := detectRecentActions()
	if err != nil {
		// Continue without action detection if error
		actions = []ActionInfo{}
	}
	context.ActionsPerformed = actions

	// Detect technical patterns used
	patterns, err := detectTechnicalPatterns()
	if err != nil {
		// Continue without pattern detection if error  
		patterns = []TechnicalPattern{}
	}
	context.TechnicalPatterns = patterns

	return context, nil
}

func getProjectInfo() (*ProjectInfo, error) {
	info := &ProjectInfo{
		WorkingDirectory: getCurrentDir(),
	}

	// Get git branch if available
	cmd := exec.Command("git", "branch", "--show-current")
	if output, err := cmd.Output(); err == nil {
		info.GitBranch = strings.TrimSpace(string(output))
	}

	// Detect project type and languages
	info.ProjectType, info.Languages, info.Frameworks = detectProjectDetails()

	return info, nil
}

func detectProjectDetails() (string, []string, []string) {
	var projectType string
	var languages []string
	var frameworks []string

	// Check for common project files
	if fileExists("package.json") {
		projectType = "Node.js"
		languages = append(languages, "JavaScript")
		
		// Check for framework-specific files
		if fileExists("next.config.js") || fileExists("next.config.ts") {
			frameworks = append(frameworks, "Next.js")
		}
		if fileExists("src/App.tsx") || fileExists("src/App.jsx") {
			frameworks = append(frameworks, "React")
		}
	}

	if fileExists("requirements.txt") || fileExists("setup.py") || fileExists("pyproject.toml") {
		if projectType != "" {
			projectType = "Multi-language"
		} else {
			projectType = "Python"
		}
		languages = append(languages, "Python")
		
		// Check for Python frameworks
		if fileExists("manage.py") {
			frameworks = append(frameworks, "Django")
		}
		if dirExists("app") && fileExists("app/__init__.py") {
			frameworks = append(frameworks, "Flask")
		}
	}

	if fileExists("go.mod") {
		if projectType != "" {
			projectType = "Multi-language"
		} else {
			projectType = "Go"
		}
		languages = append(languages, "Go")
	}

	if projectType == "" {
		projectType = "Unknown"
	}

	return projectType, languages, frameworks
}

func detectRecentActions() ([]ActionInfo, error) {
	var actions []ActionInfo

	// Check recent git commits (last 10 minutes)
	cutoff := time.Now().Add(-10 * time.Minute)
	cmd := exec.Command("git", "log", "--since=10 minutes ago", "--pretty=format:%H|%s|%ct")
	output, err := cmd.Output()
	if err == nil {
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if line == "" {
				continue
			}
			parts := strings.Split(line, "|")
			if len(parts) >= 3 {
				actions = append(actions, ActionInfo{
					Type:        "git_commit",
					Target:      parts[0][:8], // Short hash
					Description: parts[1],
					Success:     true,
				})
			}
		}
	}

	// Check for recently modified files
	err = filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		
		if info.ModTime().After(cutoff) && !strings.HasPrefix(path, ".") {
			actionType := "file_modified"
			if strings.HasSuffix(path, ".go") || strings.HasSuffix(path, ".py") || 
			   strings.HasSuffix(path, ".js") || strings.HasSuffix(path, ".ts") {
				actionType = "code_modified"
			}
			
			actions = append(actions, ActionInfo{
				Type:        actionType,
				Target:      path,
				Description: fmt.Sprintf("Modified %s", filepath.Base(path)),
				Timestamp:   info.ModTime(),
				Success:     true,
			})
		}
		return nil
	})

	return actions, err
}

func detectTechnicalPatterns() ([]TechnicalPattern, error) {
	var patterns []TechnicalPattern

	// Analyze code files for common patterns
	codeFiles := []string{}
	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		
		ext := filepath.Ext(path)
		if ext == ".go" || ext == ".py" || ext == ".js" || ext == ".ts" || ext == ".tsx" {
			codeFiles = append(codeFiles, path)
		}
		return nil
	})
	
	if err != nil {
		return patterns, err
	}

	// Look for common patterns in code files
	patternCounts := make(map[string]int)
	for _, file := range codeFiles {
		if len(codeFiles) > 50 { // Limit analysis to prevent slowdown
			break
		}
		
		content, err := os.ReadFile(file)
		if err != nil {
			continue
		}
		
		fileContent := string(content)
		
		// Detect common patterns
		if strings.Contains(fileContent, "func main()") {
			patternCounts["go_main_pattern"]++
		}
		if strings.Contains(fileContent, "import React") {
			patternCounts["react_import_pattern"]++
		}
		if strings.Contains(fileContent, "async/await") || strings.Contains(fileContent, "await ") {
			patternCounts["async_await_pattern"]++
		}
		if strings.Contains(fileContent, "error handling") || strings.Contains(fileContent, "try/catch") {
			patternCounts["error_handling_pattern"]++
		}
	}

	// Convert counts to patterns
	for pattern, count := range patternCounts {
		effectiveness := "medium"
		if count > 3 {
			effectiveness = "high"
		} else if count == 1 {
			effectiveness = "low"
		}
		
		patterns = append(patterns, TechnicalPattern{
			Pattern:       pattern,
			Context:       "codebase_analysis",
			Frequency:     count,
			Effectiveness: effectiveness,
		})
	}

	return patterns, nil
}

func generateLearningInsights(context *ConversationContext) ([]LearningInsight, error) {
	var insights []LearningInsight

	// Generate insights based on actions performed
	if len(context.ActionsPerformed) > 0 {
		successfulActions := 0
		for _, action := range context.ActionsPerformed {
			if action.Success {
				successfulActions++
			}
		}
		
		if successfulActions > 2 {
			insights = append(insights, LearningInsight{
				Type:        "success_pattern",
				Title:       "Productive Session",
				Description: fmt.Sprintf("Successfully completed %d actions in %s project", successfulActions, context.ProjectInfo.ProjectType),
				Context:     context.ProjectInfo.WorkingDirectory,
				Reusability: "high",
				Quality:     8,
			})
		}
	}

	// Generate insights based on technical patterns
	for _, pattern := range context.TechnicalPatterns {
		if pattern.Effectiveness == "high" && pattern.Frequency > 3 {
			insights = append(insights, LearningInsight{
				Type:        "best_practice",
				Title:       fmt.Sprintf("Common Pattern: %s", pattern.Pattern),
				Description: fmt.Sprintf("Pattern used %d times with high effectiveness", pattern.Frequency),
				Context:     pattern.Context,
				Reusability: "high",
				Quality:     9,
			})
		}
	}

	// Generate framework-specific insights
	for _, framework := range context.ProjectInfo.Frameworks {
		insights = append(insights, LearningInsight{
			Type:        "optimization",
			Title:       fmt.Sprintf("%s Development Session", framework),
			Description: fmt.Sprintf("Working on %s project with %s", framework, strings.Join(context.ProjectInfo.Languages, ", ")),
			Context:     context.ProjectInfo.WorkingDirectory,
			Reusability: "medium",
			Quality:     7,
		})
	}

	return insights, nil
}

func shouldSaveToMem0(context *ConversationContext) bool {
	// Save if we have valuable insights
	if len(context.LearningInsights) > 0 {
		// Check if any insights have high quality rating
		for _, insight := range context.LearningInsights {
			if insight.Quality >= 8 {
				return true
			}
		}
	}
	
	// Save if we have significant activity
	if len(context.ActionsPerformed) > 2 {
		return true
	}
	
	// Save if we have interesting technical patterns
	if len(context.TechnicalPatterns) > 1 {
		return true
	}
	
	return false
}

func saveToMem0(context *ConversationContext) error {
	// Create mem0 content
	memContent := generateMem0Content(context)
	
	// Use claude command to save to mem0
	cmd := exec.Command("claude", "--no-stream")
	cmd.Stdin = strings.NewReader(fmt.Sprintf("Save this coding session insight to memory: %s", memContent))
	
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to save to mem0: %v", err)
	}
	
	// Check if the output indicates successful mem0 save
	if !strings.Contains(string(output), "memory") && !strings.Contains(string(output), "saved") {
		return fmt.Errorf("mem0 save may have failed")
	}
	
	return nil
}

func generateMem0Content(context *ConversationContext) string {
	var content strings.Builder
	
	content.WriteString(fmt.Sprintf("CODING_SESSION_LEARNING: %s\n", context.Timestamp.Format("2006-01-02 15:04:05")))
	content.WriteString(fmt.Sprintf("Project: %s (%s)\n", context.ProjectInfo.ProjectType, strings.Join(context.ProjectInfo.Languages, ", ")))
	
	if len(context.ProjectInfo.Frameworks) > 0 {
		content.WriteString(fmt.Sprintf("Frameworks: %s\n", strings.Join(context.ProjectInfo.Frameworks, ", ")))
	}
	
	content.WriteString(fmt.Sprintf("Actions: %d successful operations\n", len(context.ActionsPerformed)))
	
	if len(context.LearningInsights) > 0 {
		content.WriteString("Key insights:\n")
		for _, insight := range context.LearningInsights {
			if insight.Quality >= 7 {
				content.WriteString(fmt.Sprintf("- %s: %s (Quality: %d/10)\n", insight.Title, insight.Description, insight.Quality))
			}
		}
	}
	
	if len(context.TechnicalPatterns) > 0 {
		content.WriteString("Technical patterns used:\n")
		for _, pattern := range context.TechnicalPatterns {
			if pattern.Effectiveness == "high" {
				content.WriteString(fmt.Sprintf("- %s (used %d times)\n", pattern.Pattern, pattern.Frequency))
			}
		}
	}
	
	return content.String()
}

func saveToLocalCache(context *ConversationContext) error {
	// Create cache directory
	cacheDir := "/Users/a.pezzotta/.claude/cache/learning"
	err := os.MkdirAll(cacheDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create cache directory: %v", err)
	}
	
	// Save context to JSON file
	filename := fmt.Sprintf("session_%s.json", context.Timestamp.Format("2006-01-02_15-04-05"))
	filepath := filepath.Join(cacheDir, filename)
	
	data, err := json.MarshalIndent(context, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal context: %v", err)
	}
	
	err = os.WriteFile(filepath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write cache file: %v", err)
	}
	
	return nil
}

func generateSessionId() string {
	return fmt.Sprintf("session_%d", time.Now().Unix())
}

func getCurrentDir() string {
	dir, _ := os.Getwd()
	return dir
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

func dirExists(dirname string) bool {
	info, err := os.Stat(dirname)
	return !os.IsNotExist(err) && info.IsDir()
}