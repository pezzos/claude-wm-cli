package metrics

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Timer represents a hierarchical performance timer
type Timer struct {
	mu            sync.RWMutex
	commandName   string
	projectPath   string
	projectName   string
	startTime     time.Time
	endTime       *time.Time
	steps         []*StepTimer
	contextData   map[string]interface{}
	exitCode      int
	collector     *PerformanceCollector
}

// StepTimer represents a step within a command
type StepTimer struct {
	mu        sync.RWMutex
	stepName  string
	startTime time.Time
	endTime   *time.Time
	error     error
	metadata  map[string]interface{}
}

// NewTimer creates a new timer for a command
func NewTimer(commandName string, collector *PerformanceCollector) *Timer {
	projectPath, _ := os.Getwd()
	projectName := filepath.Base(projectPath)
	
	return &Timer{
		commandName: commandName,
		projectPath: projectPath,
		projectName: projectName,
		startTime:   time.Now(),
		steps:       make([]*StepTimer, 0),
		contextData: make(map[string]interface{}),
		collector:   collector,
	}
}

// StartStep starts timing a specific step within the command
func (t *Timer) StartStep(stepName string) *StepTimer {
	t.mu.Lock()
	defer t.mu.Unlock()
	
	step := &StepTimer{
		stepName:  stepName,
		startTime: time.Now(),
		metadata:  make(map[string]interface{}),
	}
	
	t.steps = append(t.steps, step)
	return step
}

// Stop stops the step timer
func (s *StepTimer) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	now := time.Now()
	s.endTime = &now
}

// StopWithError stops the step timer with an error
func (s *StepTimer) StopWithError(err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	now := time.Now()
	s.endTime = &now
	s.error = err
}

// SetMetadata adds metadata to the step
func (s *StepTimer) SetMetadata(key string, value interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.metadata[key] = value
}

// Duration returns the duration of the step
func (s *StepTimer) Duration() time.Duration {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	if s.endTime == nil {
		return time.Since(s.startTime)
	}
	return s.endTime.Sub(s.startTime)
}

// SetContext adds context data to the timer
func (t *Timer) SetContext(key string, value interface{}) {
	t.mu.Lock()
	defer t.mu.Unlock()
	
	t.contextData[key] = value
}

// SetExitCode sets the exit code for the command
func (t *Timer) SetExitCode(code int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	
	t.exitCode = code
}

// Stop stops the timer and saves metrics
func (t *Timer) Stop() {
	t.mu.Lock()
	now := time.Now()
	t.endTime = &now
	t.mu.Unlock()
	
	// Save metrics synchronously to ensure data is persisted before process exits
	t.saveMetrics()
}

// Duration returns the total duration of the command
func (t *Timer) Duration() time.Duration {
	t.mu.RLock()
	defer t.mu.RUnlock()
	
	if t.endTime == nil {
		return time.Since(t.startTime)
	}
	return t.endTime.Sub(t.startTime)
}

// saveMetrics saves all metrics to the database
func (t *Timer) saveMetrics() {
	t.mu.RLock()
	defer t.mu.RUnlock()
	
	if t.collector == nil {
		return
	}
	
	// Save main command metric
	contextJSON, _ := json.Marshal(t.contextData)
	err := t.collector.storage.SaveMetric(MetricEntry{
		Timestamp:    t.startTime,
		ProjectPath:  hashProjectPath(t.projectPath),
		ProjectName:  t.projectName,
		CommandName:  t.commandName,
		StepName:     "",
		DurationMs:   int64(t.Duration().Milliseconds()),
		ContextData:  string(contextJSON),
		ToolVersion:  t.collector.version,
		ExitCode:     t.exitCode,
	})
	
	if err != nil {
		// Log error but don't fail the command
		fmt.Fprintf(os.Stderr, "Warning: Failed to save command metric: %v\n", err)
	}
	
	// Save step metrics
	for _, step := range t.steps {
		step.mu.RLock()
		if step.endTime != nil {
			stepMetadata, _ := json.Marshal(step.metadata)
			stepExitCode := 0
			if step.error != nil {
				stepExitCode = 1
			}
			
			err := t.collector.storage.SaveMetric(MetricEntry{
				Timestamp:    step.startTime,
				ProjectPath:  hashProjectPath(t.projectPath),
				ProjectName:  t.projectName,
				CommandName:  t.commandName,
				StepName:     step.stepName,
				DurationMs:   int64(step.Duration().Milliseconds()),
				ContextData:  string(stepMetadata),
				ToolVersion:  t.collector.version,
				ExitCode:     stepExitCode,
			})
			
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: Failed to save step metric: %v\n", err)
			}
		}
		step.mu.RUnlock()
	}
}

// GetStepDurations returns a map of step names to their durations
func (t *Timer) GetStepDurations() map[string]time.Duration {
	t.mu.RLock()
	defer t.mu.RUnlock()
	
	durations := make(map[string]time.Duration)
	for _, step := range t.steps {
		durations[step.stepName] = step.Duration()
	}
	return durations
}

// PrintSummary prints a summary of the timer (for debugging)
func (t *Timer) PrintSummary() {
	t.mu.RLock()
	defer t.mu.RUnlock()
	
	fmt.Printf("Command: %s (Total: %v)\n", t.commandName, t.Duration())
	
	for _, step := range t.steps {
		step.mu.RLock()
		status := "✓"
		if step.error != nil {
			status = "✗"
		}
		fmt.Printf("  %s %s: %v\n", status, step.stepName, step.Duration())
		step.mu.RUnlock()
	}
}