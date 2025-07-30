package profiler

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewHookProfiler(t *testing.T) {
	profiler := NewHookProfiler()
	
	if profiler == nil {
		t.Fatal("NewHookProfiler returned nil")
	}
	
	if profiler.metrics == nil {
		t.Error("metrics slice not initialized")
	}
	
	if profiler.aggregatedMetrics == nil {
		t.Error("aggregatedMetrics map not initialized")
	}
	
	if profiler.maxMetrics != 1000 {
		t.Errorf("Expected maxMetrics to be 1000, got %d", profiler.maxMetrics)
	}
}

func TestGetProfiler(t *testing.T) {
	profiler1 := GetProfiler()
	profiler2 := GetProfiler()
	
	if profiler1 != profiler2 {
		t.Error("GetProfiler should return the same instance (singleton)")
	}
}

func TestStartTimer(t *testing.T) {
	profiler := NewHookProfiler()
	timer := profiler.StartTimer("test-hook")
	
	if timer == nil {
		t.Fatal("StartTimer returned nil")
	}
	
	if timer.hookName != "test-hook" {
		t.Errorf("Expected hookName to be 'test-hook', got %s", timer.hookName)
	}
	
	if timer.profiler != profiler {
		t.Error("Timer should reference the correct profiler")
	}
	
	if timer.customData == nil {
		t.Error("customData map not initialized")
	}
	
	if time.Since(timer.startTime) > time.Millisecond {
		t.Error("Timer should be started recently")
	}
}

func TestTimerSetters(t *testing.T) {
	profiler := NewHookProfiler()
	timer := profiler.StartTimer("test-hook")
	
	// Test SetFilesProcessed
	timer.SetFilesProcessed(5)
	if timer.filesCount != 5 {
		t.Errorf("Expected filesCount to be 5, got %d", timer.filesCount)
	}
	
	// Test SetCacheStats
	timer.SetCacheStats(10, 3)
	if timer.cacheHits != 10 {
		t.Errorf("Expected cacheHits to be 10, got %d", timer.cacheHits)
	}
	if timer.cacheMisses != 3 {
		t.Errorf("Expected cacheMisses to be 3, got %d", timer.cacheMisses)
	}
	
	// Test SetCustomMetric
	timer.SetCustomMetric("test_key", "test_value")
	if timer.customData["test_key"] != "test_value" {
		t.Error("Custom metric not set correctly")
	}
}

func TestTimerStop(t *testing.T) {
	profiler := NewHookProfiler()
	timer := profiler.StartTimer("test-hook")
	
	timer.SetFilesProcessed(3)
	timer.SetCacheStats(5, 2)
	timer.SetCustomMetric("test_metric", 42)
	
	// Add small delay to ensure non-zero execution time
	time.Sleep(1 * time.Millisecond)
	
	timer.Stop("success")
	
	// Check that metrics were recorded
	if len(profiler.metrics) != 1 {
		t.Errorf("Expected 1 metric recorded, got %d", len(profiler.metrics))
	}
	
	metric := profiler.metrics[0]
	if metric.HookName != "test-hook" {
		t.Errorf("Expected hook name 'test-hook', got %s", metric.HookName)
	}
	
	if metric.Status != "success" {
		t.Errorf("Expected status 'success', got %s", metric.Status)
	}
	
	if metric.FilesProcessed != 3 {
		t.Errorf("Expected 3 files processed, got %d", metric.FilesProcessed)
	}
	
	if metric.CacheHits != 5 {
		t.Errorf("Expected 5 cache hits, got %d", metric.CacheHits)
	}
	
	if metric.CacheMisses != 2 {
		t.Errorf("Expected 2 cache misses, got %d", metric.CacheMisses)
	}
	
	if metric.CustomMetrics["test_metric"] != 42 {
		t.Error("Custom metric not preserved")
	}
	
	if metric.ExecutionTimeMs <= 0 {
		t.Error("Execution time should be positive")
	}
}

func TestTimerStopWithError(t *testing.T) {
	profiler := NewHookProfiler()
	timer := profiler.StartTimer("test-hook")
	
	timer.Stop("error", "Test error message")
	
	metric := profiler.metrics[0]
	if metric.Status != "error" {
		t.Errorf("Expected status 'error', got %s", metric.Status)
	}
	
	if metric.ErrorDetails != "Test error message" {
		t.Errorf("Expected error details 'Test error message', got %s", metric.ErrorDetails)
	}
}

func TestRecordMetrics(t *testing.T) {
	profiler := NewHookProfiler()
	
	metrics := HookMetrics{
		HookName:         "test-hook",
		ExecutionTimeMs:  100,
		FilesProcessed:   5,
		CacheHits:        3,
		CacheMisses:      2,
		Status:           "success",
		Timestamp:        time.Now(),
	}
	
	profiler.RecordMetrics(metrics)
	
	// Check raw metrics
	if len(profiler.metrics) != 1 {
		t.Errorf("Expected 1 metric, got %d", len(profiler.metrics))
	}
	
	// Check aggregated metrics
	agg, exists := profiler.aggregatedMetrics["test-hook"]
	if !exists {
		t.Fatal("Aggregated metrics not created")
	}
	
	if agg.TotalExecutions != 1 {
		t.Errorf("Expected 1 total execution, got %d", agg.TotalExecutions)
	}
	
	if agg.SuccessfulRuns != 1 {
		t.Errorf("Expected 1 successful run, got %d", agg.SuccessfulRuns)
	}
	
	if agg.FailedRuns != 0 {
		t.Errorf("Expected 0 failed runs, got %d", agg.FailedRuns)
	}
	
	if agg.AverageTimeMs != 100.0 {
		t.Errorf("Expected average time 100.0, got %f", agg.AverageTimeMs)
	}
}

func TestAggregatedMetricsMultipleRuns(t *testing.T) {
	profiler := NewHookProfiler()
	
	// Record multiple metrics for the same hook
	metrics1 := HookMetrics{
		HookName:        "test-hook",
		ExecutionTimeMs: 100,
		FilesProcessed:  5,
		CacheHits:       3,
		CacheMisses:     2,
		Status:          "success",
		Timestamp:       time.Now(),
	}
	
	metrics2 := HookMetrics{
		HookName:        "test-hook",
		ExecutionTimeMs: 200,
		FilesProcessed:  10,
		CacheHits:       5,
		CacheMisses:     1,
		Status:          "success",
		Timestamp:       time.Now(),
	}
	
	metrics3 := HookMetrics{
		HookName:        "test-hook",
		ExecutionTimeMs: 150,
		FilesProcessed:  8,
		CacheHits:       2,
		CacheMisses:     3,
		Status:          "error",
		Timestamp:       time.Now(),
	}
	
	profiler.RecordMetrics(metrics1)
	profiler.RecordMetrics(metrics2)
	profiler.RecordMetrics(metrics3)
	
	agg := profiler.aggregatedMetrics["test-hook"]
	
	if agg.TotalExecutions != 3 {
		t.Errorf("Expected 3 total executions, got %d", agg.TotalExecutions)
	}
	
	if agg.SuccessfulRuns != 2 {
		t.Errorf("Expected 2 successful runs, got %d", agg.SuccessfulRuns)
	}
	
	if agg.FailedRuns != 1 {
		t.Errorf("Expected 1 failed run, got %d", agg.FailedRuns)
	}
	
	expectedAvg := (100.0 + 200.0 + 150.0) / 3.0
	if agg.AverageTimeMs != expectedAvg {
		t.Errorf("Expected average time %f, got %f", expectedAvg, agg.AverageTimeMs)
	}
	
	if agg.MinTimeMs != 100 {
		t.Errorf("Expected min time 100, got %d", agg.MinTimeMs)
	}
	
	if agg.MaxTimeMs != 200 {
		t.Errorf("Expected max time 200, got %d", agg.MaxTimeMs)
	}
	
	if agg.TotalFilesProcessed != 23 {
		t.Errorf("Expected total files processed 23, got %d", agg.TotalFilesProcessed)
	}
	
	// Cache hit rate calculation: (3+5+2) / (3+5+2+2+1+3) = 10/16 = 0.625
	expectedCacheHitRate := 10.0 / 16.0
	if agg.CacheHitRate != expectedCacheHitRate {
		t.Errorf("Expected cache hit rate %f, got %f", expectedCacheHitRate, agg.CacheHitRate)
	}
}

func TestGetMetrics(t *testing.T) {
	profiler := NewHookProfiler()
	
	// Add multiple metrics for different hooks
	for i := 0; i < 5; i++ {
		metrics := HookMetrics{
			HookName:        "test-hook",
			ExecutionTimeMs: int64(100 + i*10),
			Status:          "success",
			Timestamp:       time.Now(),
		}
		profiler.RecordMetrics(metrics)
	}
	
	// Add metrics for different hook
	metrics := HookMetrics{
		HookName:        "other-hook",
		ExecutionTimeMs: 50,
		Status:          "success",
		Timestamp:       time.Now(),
	}
	profiler.RecordMetrics(metrics)
	
	// Get metrics for specific hook
	testHookMetrics := profiler.GetMetrics("test-hook", 3)
	
	if len(testHookMetrics) != 3 {
		t.Errorf("Expected 3 metrics, got %d", len(testHookMetrics))
	}
	
	// Should return most recent first
	if testHookMetrics[0].ExecutionTimeMs != 140 {
		t.Errorf("Expected first metric to be most recent (140ms), got %dms", testHookMetrics[0].ExecutionTimeMs)
	}
	
	// Get metrics for other hook
	otherHookMetrics := profiler.GetMetrics("other-hook", 10)
	if len(otherHookMetrics) != 1 {
		t.Errorf("Expected 1 metric for other-hook, got %d", len(otherHookMetrics))
	}
}

func TestGetTopSlowestHooks(t *testing.T) {
	profiler := NewHookProfiler()
	
	// Add metrics for different hooks with different average times
	hooks := []struct {
		name string
		times []int64
	}{
		{"fast-hook", []int64{10, 20, 30}},
		{"slow-hook", []int64{200, 300, 400}},
		{"medium-hook", []int64{100, 150, 200}},
	}
	
	for _, hook := range hooks {
		for _, execTime := range hook.times {
			metrics := HookMetrics{
				HookName:        hook.name,
				ExecutionTimeMs: execTime,
				Status:          "success",
				Timestamp:       time.Now(),
			}
			profiler.RecordMetrics(metrics)
		}
	}
	
	slowest := profiler.GetTopSlowestHooks(2)
	
	if len(slowest) != 2 {
		t.Errorf("Expected 2 slowest hooks, got %d", len(slowest))
	}
	
	// Should be sorted by average time descending
	if slowest[0].HookName != "slow-hook" {
		t.Errorf("Expected first slowest to be 'slow-hook', got %s", slowest[0].HookName)
	}
	
	if slowest[1].HookName != "medium-hook" {
		t.Errorf("Expected second slowest to be 'medium-hook', got %s", slowest[1].HookName)
	}
}

func TestMetricsRotation(t *testing.T) {
	profiler := NewHookProfiler()
	profiler.maxMetrics = 3 // Set small limit for testing
	
	// Add more metrics than the limit
	for i := 0; i < 5; i++ {
		metrics := HookMetrics{
			HookName:        "test-hook",
			ExecutionTimeMs: int64(100 + i*10),
			Status:          "success",
			Timestamp:       time.Now(),
		}
		profiler.RecordMetrics(metrics)
	}
	
	// Should only keep the last 3 metrics
	if len(profiler.metrics) != 3 {
		t.Errorf("Expected 3 metrics after rotation, got %d", len(profiler.metrics))
	}
	
	// Should keep the most recent ones
	if profiler.metrics[0].ExecutionTimeMs != 120 {
		t.Errorf("Expected first metric to be 120ms, got %dms", profiler.metrics[0].ExecutionTimeMs)
	}
	
	if profiler.metrics[2].ExecutionTimeMs != 140 {
		t.Errorf("Expected last metric to be 140ms, got %dms", profiler.metrics[2].ExecutionTimeMs)
	}
}

func TestSaveMetricsToFile(t *testing.T) {
	// Create temporary directory for testing
	tempDir := t.TempDir()
	
	profiler := NewHookProfiler()
	profiler.metricsFile = filepath.Join(tempDir, "test-metrics.json")
	
	// Add some metrics
	metrics := HookMetrics{
		HookName:        "test-hook",
		ExecutionTimeMs: 100,
		Status:          "success",
		Timestamp:       time.Now(),
	}
	profiler.RecordMetrics(metrics)
	
	// Check that file was created
	if _, err := os.Stat(profiler.metricsFile); os.IsNotExist(err) {
		t.Error("Metrics file was not created")
	}
	
	// Check file content
	data, err := os.ReadFile(profiler.metricsFile)
	if err != nil {
		t.Fatalf("Error reading metrics file: %v", err)
	}
	
	var jsonData map[string]interface{}
	if err := json.Unmarshal(data, &jsonData); err != nil {
		t.Fatalf("Error parsing metrics JSON: %v", err)
	}
	
	if _, exists := jsonData["aggregated_metrics"]; !exists {
		t.Error("Metrics file missing aggregated_metrics")
	}
	
	if _, exists := jsonData["recent_metrics"]; !exists {
		t.Error("Metrics file missing recent_metrics")
	}
	
	if _, exists := jsonData["generated_at"]; !exists {
		t.Error("Metrics file missing generated_at")
	}
}

func TestClearMetrics(t *testing.T) {
	profiler := NewHookProfiler()
	
	// Add some metrics
	metrics := HookMetrics{
		HookName:        "test-hook",
		ExecutionTimeMs: 100,
		Status:          "success",
		Timestamp:       time.Now(),
	}
	profiler.RecordMetrics(metrics)
	
	// Verify metrics exist
	if len(profiler.metrics) != 1 {
		t.Error("Expected 1 metric before clearing")
	}
	
	if len(profiler.aggregatedMetrics) != 1 {
		t.Error("Expected 1 aggregated metric before clearing")
	}
	
	// Clear metrics
	profiler.ClearMetrics()
	
	// Verify metrics are cleared
	if len(profiler.metrics) != 0 {
		t.Error("Expected 0 metrics after clearing")
	}
	
	if len(profiler.aggregatedMetrics) != 0 {
		t.Error("Expected 0 aggregated metrics after clearing")
	}
}

func TestConcurrentAccess(t *testing.T) {
	profiler := NewHookProfiler()
	
	// Start multiple goroutines to record metrics concurrently
	done := make(chan bool)
	
	for i := 0; i < 10; i++ {
		go func(id int) {
			timer := profiler.StartTimer("concurrent-hook")
			time.Sleep(time.Millisecond)
			timer.Stop("success")
			done <- true
		}(i)
	}
	
	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
	
	// Verify all metrics were recorded
	if len(profiler.metrics) != 10 {
		t.Errorf("Expected 10 metrics from concurrent access, got %d", len(profiler.metrics))
	}
	
	agg := profiler.aggregatedMetrics["concurrent-hook"]
	if agg.TotalExecutions != 10 {
		t.Errorf("Expected 10 total executions, got %d", agg.TotalExecutions)
	}
}

func BenchmarkRecordMetrics(b *testing.B) {
	profiler := NewHookProfiler()
	
	metrics := HookMetrics{
		HookName:        "benchmark-hook",
		ExecutionTimeMs: 100,
		Status:          "success",
		Timestamp:       time.Now(),
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		profiler.RecordMetrics(metrics)
	}
}

func BenchmarkTimerStartStop(b *testing.B) {
	profiler := NewHookProfiler()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		timer := profiler.StartTimer("benchmark-hook")
		timer.Stop("success")
	}
}