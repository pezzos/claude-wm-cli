package metrics

import (
	"fmt"
	"os"
	"sync"
)

// PerformanceCollector is the main collector for performance metrics
type PerformanceCollector struct {
	mu            sync.RWMutex
	storage       *Storage
	version       string
	enabled       bool
	currentTimers map[string]*Timer
}

var (
	globalCollector *PerformanceCollector
	collectorOnce   sync.Once
)

// GetCollector returns the global performance collector instance
func GetCollector() *PerformanceCollector {
	collectorOnce.Do(func() {
		storage, err := NewStorage()
		if err != nil {
			// If we can't initialize storage, create a disabled collector
			fmt.Fprintf(os.Stderr, "Warning: Failed to initialize metrics storage: %v\n", err)
			globalCollector = &PerformanceCollector{
				enabled: false,
			}
			return
		}
		
		globalCollector = &PerformanceCollector{
			storage:       storage,
			version:       GetToolVersion(),
			enabled:       true,
			currentTimers: make(map[string]*Timer),
		}
	})
	
	return globalCollector
}

// StartCommand starts timing a command
func (pc *PerformanceCollector) StartCommand(commandName string) *Timer {
	if !pc.enabled {
		return &Timer{} // Return dummy timer
	}
	
	pc.mu.Lock()
	defer pc.mu.Unlock()
	
	timer := NewTimer(commandName, pc)
	pc.currentTimers[commandName] = timer
	
	return timer
}

// StartCommandWithContext starts timing a command with additional context
func (pc *PerformanceCollector) StartCommandWithContext(commandName string, context map[string]interface{}) *Timer {
	timer := pc.StartCommand(commandName)
	
	if timer != nil {
		for k, v := range context {
			timer.SetContext(k, v)
		}
	}
	
	return timer
}

// GetCurrentTimer returns the currently active timer for a command
func (pc *PerformanceCollector) GetCurrentTimer(commandName string) *Timer {
	if !pc.enabled {
		return nil
	}
	
	pc.mu.RLock()
	defer pc.mu.RUnlock()
	
	return pc.currentTimers[commandName]
}

// StopCommand stops timing a command
func (pc *PerformanceCollector) StopCommand(commandName string) {
	if !pc.enabled {
		return
	}
	
	pc.mu.Lock()
	defer pc.mu.Unlock()
	
	if timer, exists := pc.currentTimers[commandName]; exists {
		timer.Stop()
		delete(pc.currentTimers, commandName)
	}
}

// IsEnabled returns whether metrics collection is enabled
func (pc *PerformanceCollector) IsEnabled() bool {
	return pc.enabled
}

// GetStats returns command statistics
func (pc *PerformanceCollector) GetStats(commandName string, days int) (*CommandStats, error) {
	if !pc.enabled {
		return nil, fmt.Errorf("metrics collection is disabled")
	}
	
	return pc.storage.GetCommandStats(commandName, days)
}

// GetStepStats returns step-level statistics
func (pc *PerformanceCollector) GetStepStats(commandName string, days int) ([]StepStats, error) {
	if !pc.enabled {
		return nil, fmt.Errorf("metrics collection is disabled")
	}
	
	return pc.storage.GetStepStats(commandName, days)
}

// GetSlowCommands returns commands slower than threshold
func (pc *PerformanceCollector) GetSlowCommands(thresholdMs int64, days int) ([]CommandStats, error) {
	if !pc.enabled {
		return nil, fmt.Errorf("metrics collection is disabled")
	}
	
	return pc.storage.GetSlowCommands(thresholdMs, days)
}

// GetProjectComparison returns performance comparison across projects
func (pc *PerformanceCollector) GetProjectComparison(days int) ([]ProjectStats, error) {
	if !pc.enabled {
		return nil, fmt.Errorf("metrics collection is disabled")
	}
	
	return pc.storage.GetProjectComparison(days)
}

// GetAllCommandStats returns statistics for all commands
func (pc *PerformanceCollector) GetAllCommandStats(days int) ([]CommandStats, error) {
	if !pc.enabled {
		return nil, fmt.Errorf("metrics collection is disabled")
	}
	
	return pc.storage.GetAllCommandStats(days)
}

// Close closes the collector and its storage
func (pc *PerformanceCollector) Close() error {
	if pc.storage != nil {
		return pc.storage.Close()
	}
	return nil
}

// GetToolVersion returns the current tool version
func GetToolVersion() string {
	// This would be set during build time
	version := os.Getenv("CLAUDE_WM_VERSION")
	if version == "" {
		version = "dev"
	}
	return version
}

// Helper function to create a timer with common context
func StartTimerWithProjectContext(commandName string) *Timer {
	collector := GetCollector()
	timer := collector.StartCommand(commandName)
	
	// Add common project context
	if wd, err := os.Getwd(); err == nil {
		timer.SetContext("working_directory", wd)
		
		// Add project complexity metrics
		if complexity := getProjectComplexity(wd); complexity != nil {
			timer.SetContext("project_complexity", complexity)
		}
	}
	
	return timer
}

// getProjectComplexity analyzes project complexity for context
func getProjectComplexity(projectPath string) map[string]interface{} {
	complexity := make(map[string]interface{})
	
	// Count JSON files
	jsonCount := 0
	if files, err := os.ReadDir(projectPath); err == nil {
		for _, file := range files {
			if !file.IsDir() && len(file.Name()) > 5 && file.Name()[len(file.Name())-5:] == ".json" {
				jsonCount++
			}
		}
	}
	
	complexity["json_files_count"] = jsonCount
	
	// Check for docs structure
	if _, err := os.Stat(fmt.Sprintf("%s/docs", projectPath)); err == nil {
		complexity["has_docs_structure"] = true
	} else {
		complexity["has_docs_structure"] = false
	}
	
	// Check for git
	if _, err := os.Stat(fmt.Sprintf("%s/.git", projectPath)); err == nil {
		complexity["is_git_repo"] = true
	} else {
		complexity["is_git_repo"] = false
	}
	
	return complexity
}