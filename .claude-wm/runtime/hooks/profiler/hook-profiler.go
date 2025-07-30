package profiler

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// HookMetrics represents performance metrics for a single hook execution
type HookMetrics struct {
	HookName         string            `json:"hook_name"`
	ExecutionTimeMs  int64             `json:"execution_time_ms"`
	MemoryUsageBytes int64             `json:"memory_usage_bytes"`
	FilesProcessed   int               `json:"files_processed"`
	CacheHits        int               `json:"cache_hits"`
	CacheMisses      int               `json:"cache_misses"`
	Timestamp        time.Time         `json:"timestamp"`
	Status           string            `json:"status"` // "success", "error", "warning"
	ErrorDetails     string            `json:"error_details,omitempty"`
	CustomMetrics    map[string]interface{} `json:"custom_metrics,omitempty"`
}

// AggregatedMetrics represents summary statistics for hook performance
type AggregatedMetrics struct {
	HookName           string    `json:"hook_name"`
	TotalExecutions    int       `json:"total_executions"`
	SuccessfulRuns     int       `json:"successful_runs"`
	FailedRuns         int       `json:"failed_runs"`
	AverageTimeMs      float64   `json:"average_time_ms"`
	MinTimeMs          int64     `json:"min_time_ms"`
	MaxTimeMs          int64     `json:"max_time_ms"`
	TotalFilesProcessed int      `json:"total_files_processed"`
	CacheHitRate       float64   `json:"cache_hit_rate"`
	LastExecuted       time.Time `json:"last_executed"`
}

// HookProfiler manages performance monitoring for hooks
type HookProfiler struct {
	mu                sync.RWMutex
	metrics           []HookMetrics
	aggregatedMetrics map[string]*AggregatedMetrics
	maxMetrics        int
	metricsFile       string
}

// Global profiler instance
var (
	globalProfiler *HookProfiler
	profilerOnce   sync.Once
)

// GetProfiler returns the global profiler instance
func GetProfiler() *HookProfiler {
	profilerOnce.Do(func() {
		globalProfiler = NewHookProfiler()
	})
	return globalProfiler
}

// NewHookProfiler creates a new hook profiler instance
func NewHookProfiler() *HookProfiler {
	homeDir, _ := os.UserHomeDir()
	metricsDir := filepath.Join(homeDir, ".claude", "hooks", "metrics")
	os.MkdirAll(metricsDir, 0755)
	
	return &HookProfiler{
		metrics:           make([]HookMetrics, 0),
		aggregatedMetrics: make(map[string]*AggregatedMetrics),
		maxMetrics:        1000, // Keep last 1000 metrics
		metricsFile:       filepath.Join(metricsDir, "performance-metrics.json"),
	}
}

// Timer represents a timing session for a hook
type Timer struct {
	hookName    string
	startTime   time.Time
	profiler    *HookProfiler
	customData  map[string]interface{}
	filesCount  int
	cacheHits   int
	cacheMisses int
}

// StartTimer begins timing a hook execution
func (p *HookProfiler) StartTimer(hookName string) *Timer {
	return &Timer{
		hookName:   hookName,
		startTime:  time.Now(),
		profiler:   p,
		customData: make(map[string]interface{}),
	}
}

// SetFilesProcessed records the number of files processed
func (t *Timer) SetFilesProcessed(count int) {
	t.filesCount = count
}

// SetCacheStats records cache hit/miss statistics
func (t *Timer) SetCacheStats(hits, misses int) {
	t.cacheHits = hits
	t.cacheMisses = misses
}

// SetCustomMetric adds a custom metric to the timer
func (t *Timer) SetCustomMetric(key string, value interface{}) {
	t.customData[key] = value
}

// Stop completes the timing and records the metrics
func (t *Timer) Stop(status string, errorDetails ...string) {
	duration := time.Since(t.startTime)
	
	var errorMsg string
	if len(errorDetails) > 0 {
		errorMsg = errorDetails[0]
	}
	
	metrics := HookMetrics{
		HookName:         t.hookName,
		ExecutionTimeMs:  duration.Milliseconds(),
		MemoryUsageBytes: getMemoryUsage(),
		FilesProcessed:   t.filesCount,
		CacheHits:        t.cacheHits,
		CacheMisses:      t.cacheMisses,
		Timestamp:        time.Now(),
		Status:           status,
		ErrorDetails:     errorMsg,
		CustomMetrics:    t.customData,
	}
	
	t.profiler.RecordMetrics(metrics)
}

// RecordMetrics adds metrics to the profiler
func (p *HookProfiler) RecordMetrics(metrics HookMetrics) {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	// Add to raw metrics (with rotation)
	p.metrics = append(p.metrics, metrics)
	if len(p.metrics) > p.maxMetrics {
		p.metrics = p.metrics[len(p.metrics)-p.maxMetrics:]
	}
	
	// Update aggregated metrics
	p.updateAggregatedMetrics(metrics)
	
	// Save metrics to file
	p.saveMetricsToFile()
}

// updateAggregatedMetrics updates the aggregated statistics
func (p *HookProfiler) updateAggregatedMetrics(metrics HookMetrics) {
	agg, exists := p.aggregatedMetrics[metrics.HookName]
	if !exists {
		agg = &AggregatedMetrics{
			HookName:    metrics.HookName,
			MinTimeMs:   metrics.ExecutionTimeMs,
			MaxTimeMs:   metrics.ExecutionTimeMs,
		}
		p.aggregatedMetrics[metrics.HookName] = agg
	}
	
	agg.TotalExecutions++
	agg.TotalFilesProcessed += metrics.FilesProcessed
	agg.LastExecuted = metrics.Timestamp
	
	if metrics.Status == "success" {
		agg.SuccessfulRuns++
	} else {
		agg.FailedRuns++
	}
	
	// Update timing statistics
	if metrics.ExecutionTimeMs < agg.MinTimeMs {
		agg.MinTimeMs = metrics.ExecutionTimeMs
	}
	if metrics.ExecutionTimeMs > agg.MaxTimeMs {
		agg.MaxTimeMs = metrics.ExecutionTimeMs
	}
	
	// Calculate average time
	totalTime := int64(0)
	count := 0
	for _, m := range p.metrics {
		if m.HookName == metrics.HookName {
			totalTime += m.ExecutionTimeMs
			count++
		}
	}
	if count > 0 {
		agg.AverageTimeMs = float64(totalTime) / float64(count)
	}
	
	// Calculate cache hit rate
	totalHits := 0
	totalRequests := 0
	for _, m := range p.metrics {
		if m.HookName == metrics.HookName {
			totalHits += m.CacheHits
			totalRequests += m.CacheHits + m.CacheMisses
		}
	}
	if totalRequests > 0 {
		agg.CacheHitRate = float64(totalHits) / float64(totalRequests)
	}
}

// GetMetrics returns recent metrics for a specific hook
func (p *HookProfiler) GetMetrics(hookName string, limit int) []HookMetrics {
	p.mu.RLock()
	defer p.mu.RUnlock()
	
	var result []HookMetrics
	count := 0
	
	// Return most recent metrics first
	for i := len(p.metrics) - 1; i >= 0 && count < limit; i-- {
		if p.metrics[i].HookName == hookName {
			result = append(result, p.metrics[i])
			count++
		}
	}
	
	return result
}

// GetAggregatedMetrics returns aggregated statistics
func (p *HookProfiler) GetAggregatedMetrics() map[string]*AggregatedMetrics {
	p.mu.RLock()
	defer p.mu.RUnlock()
	
	// Return a copy to avoid concurrent access issues
	result := make(map[string]*AggregatedMetrics)
	for k, v := range p.aggregatedMetrics {
		result[k] = v
	}
	
	return result
}

// GetTopSlowestHooks returns the hooks with highest average execution time
func (p *HookProfiler) GetTopSlowestHooks(limit int) []*AggregatedMetrics {
	p.mu.RLock()
	defer p.mu.RUnlock()
	
	var hooks []*AggregatedMetrics
	for _, agg := range p.aggregatedMetrics {
		hooks = append(hooks, agg)
	}
	
	// Sort by average time (descending)
	for i := 0; i < len(hooks)-1; i++ {
		for j := i + 1; j < len(hooks); j++ {
			if hooks[i].AverageTimeMs < hooks[j].AverageTimeMs {
				hooks[i], hooks[j] = hooks[j], hooks[i]
			}
		}
	}
	
	if len(hooks) > limit {
		hooks = hooks[:limit]
	}
	
	return hooks
}

// saveMetricsToFile saves current metrics to JSON file
func (p *HookProfiler) saveMetricsToFile() {
	data := map[string]interface{}{
		"aggregated_metrics": p.aggregatedMetrics,
		"recent_metrics":     p.metrics,
		"generated_at":       time.Now(),
	}
	
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling metrics: %v\n", err)
		return
	}
	
	err = os.WriteFile(p.metricsFile, jsonData, 0644)
	if err != nil {
		fmt.Printf("Error writing metrics file: %v\n", err)
	}
}

// PrintSummary prints a summary of hook performance
func (p *HookProfiler) PrintSummary() {
	p.mu.RLock()
	defer p.mu.RUnlock()
	
	fmt.Println("=== Hook Performance Summary ===")
	fmt.Printf("Total hooks monitored: %d\n", len(p.aggregatedMetrics))
	fmt.Printf("Total executions: %d\n", len(p.metrics))
	
	slowest := p.GetTopSlowestHooks(5)
	if len(slowest) > 0 {
		fmt.Println("\nTop 5 slowest hooks:")
		for i, hook := range slowest {
			fmt.Printf("%d. %s: %.2fms avg (min: %dms, max: %dms)\n",
				i+1, hook.HookName, hook.AverageTimeMs, hook.MinTimeMs, hook.MaxTimeMs)
		}
	}
}

// getMemoryUsage returns current memory usage (simplified implementation)
func getMemoryUsage() int64 {
	// This is a simplified implementation
	// In production, you might want to use runtime.ReadMemStats() or similar
	return 0
}

// ClearMetrics clears all stored metrics
func (p *HookProfiler) ClearMetrics() {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	p.metrics = make([]HookMetrics, 0)
	p.aggregatedMetrics = make(map[string]*AggregatedMetrics)
	p.saveMetricsToFile()
}