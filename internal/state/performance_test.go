package state

import (
	"fmt"
	"os"
	"testing"
	"time"
)

// Test data structures for performance testing
type SmallTestData struct {
	ID      int      `json:"id"`
	Name    string   `json:"name"`
	Tags    []string `json:"tags"`
	Created time.Time `json:"created"`
}

type MediumTestData struct {
	Projects []ProjectData `json:"projects"`
	Meta     MetaData      `json:"meta"`
}

type ProjectData struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Tasks       []TaskData `json:"tasks"`
	Created     time.Time  `json:"created"`
	Modified    time.Time  `json:"modified"`
}

type TaskData struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	Priority    string    `json:"priority"`
	Tags        []string  `json:"tags"`
	Created     time.Time `json:"created"`
	Modified    time.Time `json:"modified"`
}

type MetaData struct {
	Version     string    `json:"version"`
	Created     time.Time `json:"created"`
	LastUpdated time.Time `json:"last_updated"`
	Statistics  StatsData `json:"statistics"`
}

type StatsData struct {
	TotalProjects int64 `json:"total_projects"`
	TotalTasks    int64 `json:"total_tasks"`
	CompletedTasks int64 `json:"completed_tasks"`
}

// Test fixtures
func createSmallTestData() *SmallTestData {
	return &SmallTestData{
		ID:      1,
		Name:    "Test Project",
		Tags:    []string{"test", "development", "go"},
		Created: time.Now(),
	}
}

func createMediumTestData() *MediumTestData {
	projects := make([]ProjectData, 100)
	for i := 0; i < 100; i++ {
		tasks := make([]TaskData, 50)
		for j := 0; j < 50; j++ {
			tasks[j] = TaskData{
				ID:          fmt.Sprintf("task-%d-%d", i, j),
				Title:       fmt.Sprintf("Task %d in Project %d", j, i),
				Description: fmt.Sprintf("This is a detailed description for task %d in project %d. It contains multiple sentences to increase the data size. The task involves implementing features, writing tests, and documenting the work.", j, i),
				Status:      []string{"todo", "in_progress", "completed"}[j%3],
				Priority:    []string{"low", "medium", "high"}[j%3],
				Tags:        []string{"frontend", "backend", "testing", "documentation", "bug", "feature"},
				Created:     time.Now().Add(-time.Duration(j) * time.Hour),
				Modified:    time.Now().Add(-time.Duration(j/2) * time.Minute),
			}
		}

		projects[i] = ProjectData{
			ID:          fmt.Sprintf("project-%d", i),
			Title:       fmt.Sprintf("Project %d", i),
			Description: fmt.Sprintf("This is project %d with detailed description. It involves multiple components, extensive testing, and comprehensive documentation. The project aims to deliver high-quality software solutions.", i),
			Tasks:       tasks,
			Created:     time.Now().Add(-time.Duration(i) * 24 * time.Hour),
			Modified:    time.Now().Add(-time.Duration(i) * time.Hour),
		}
	}

	return &MediumTestData{
		Projects: projects,
		Meta: MetaData{
			Version:     "1.0.0",
			Created:     time.Now().Add(-30 * 24 * time.Hour),
			LastUpdated: time.Now(),
			Statistics: StatsData{
				TotalProjects:  100,
				TotalTasks:     5000,
				CompletedTasks: 1667,
			},
		},
	}
}

func createLargeTestData() *MediumTestData {
	// Create larger dataset for large file testing (reduced size for memory efficiency)
	projects := make([]ProjectData, 200)
	for i := 0; i < 200; i++ {
		tasks := make([]TaskData, 50)
		for j := 0; j < 50; j++ {
			tasks[j] = TaskData{
				ID:          fmt.Sprintf("task-%d-%d", i, j),
				Title:       fmt.Sprintf("Task %d in Project %d", j, i),
				Description: fmt.Sprintf("This is a very detailed description for task %d in project %d. It contains multiple sentences and paragraphs to significantly increase the data size. The task involves complex implementations, extensive testing, comprehensive documentation, and thorough code reviews. Additionally, it requires integration with multiple systems, performance optimization, security considerations, and user experience enhancements. The description continues with more details about technical requirements, business logic, edge cases, error handling, logging, monitoring, and deployment strategies.", j, i),
				Status:      []string{"todo", "in_progress", "completed"}[j%3],
				Priority:    []string{"low", "medium", "high"}[j%3],
				Tags:        []string{"frontend", "backend", "testing", "documentation", "bug", "feature", "performance", "security", "ui", "api"},
				Created:     time.Now().Add(-time.Duration(j) * time.Hour),
				Modified:    time.Now().Add(-time.Duration(j/2) * time.Minute),
			}
		}

		projects[i] = ProjectData{
			ID:          fmt.Sprintf("project-%d", i),
			Title:       fmt.Sprintf("Large Scale Project %d", i),
			Description: fmt.Sprintf("This is project %d with very comprehensive description. It involves multiple components, extensive testing, comprehensive documentation, performance optimization, security hardening, scalability planning, and maintainability considerations. The project aims to deliver enterprise-grade software solutions with high availability, fault tolerance, and exceptional user experience. The implementation requires careful architecture design, code quality standards, automated testing strategies, continuous integration pipelines, monitoring and alerting systems, and robust deployment procedures.", i),
			Tasks:       tasks,
			Created:     time.Now().Add(-time.Duration(i) * 24 * time.Hour),
			Modified:    time.Now().Add(-time.Duration(i) * time.Hour),
		}
	}

	return &MediumTestData{
		Projects: projects,
		Meta: MetaData{
			Version:     "2.0.0",
			Created:     time.Now().Add(-365 * 24 * time.Hour),
			LastUpdated: time.Now(),
			Statistics: StatsData{
				TotalProjects:  200,
				TotalTasks:     10000,
				CompletedTasks: 3333,
			},
		},
	}
}

// Test performance with small files (should be sub-second)
func TestSmallFilePerformance(t *testing.T) {
	// Setup
	atomic := NewAtomicWriter("")
	osm := NewOptimizedStateManager(atomic)
	defer osm.Close()

	// Create test file
	testData := createSmallTestData()
	testFile := "test_small.json"
	defer os.Remove(testFile)

	// Write test data
	err := osm.WriteJSONOptimized(testFile, testData, nil)
	if err != nil {
		t.Fatalf("Failed to write test data: %v", err)
	}

	// Test read performance
	start := time.Now()
	var readData SmallTestData
	err = osm.ReadJSONOptimized(testFile, &readData, nil)
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("Failed to read test data: %v", err)
	}

	// Verify performance target (sub-second for small files)
	if duration > SmallFileMaxDuration {
		t.Errorf("Small file read took %v, expected less than %v", duration, SmallFileMaxDuration)
	}

	t.Logf("Small file read performance: %v (target: <%v)", duration, SmallFileMaxDuration)

	// Verify data integrity
	if readData.ID != testData.ID || readData.Name != testData.Name {
		t.Error("Data integrity check failed")
	}
}

// Test performance with medium files
func TestMediumFilePerformance(t *testing.T) {
	// Setup
	atomic := NewAtomicWriter("")
	osm := NewOptimizedStateManager(atomic)
	defer osm.Close()

	// Create test file
	testData := createMediumTestData()
	testFile := "test_medium.json"
	defer os.Remove(testFile)

	// Write test data
	start := time.Now()
	err := osm.WriteJSONOptimized(testFile, testData, &OptimizedWriteOptions{
		UseStreaming: true,
		ChunkSize:    ChunkSize,
	})
	writeDuration := time.Since(start)

	if err != nil {
		t.Fatalf("Failed to write test data: %v", err)
	}

	t.Logf("Medium file write performance: %v", writeDuration)

	// Test read performance
	start = time.Now()
	var readData MediumTestData
	err = osm.ReadJSONOptimized(testFile, &readData, &OptimizedReadOptions{
		UseStreaming: true,
		ChunkSize:    ChunkSize,
	})
	readDuration := time.Since(start)

	if err != nil {
		t.Fatalf("Failed to read test data: %v", err)
	}

	// Verify performance target
	if readDuration > LargeFileMaxDuration {
		t.Errorf("Medium file read took %v, expected less than %v", readDuration, LargeFileMaxDuration)
	}

	t.Logf("Medium file read performance: %v (target: <%v)", readDuration, LargeFileMaxDuration)

	// Verify data integrity
	if len(readData.Projects) != len(testData.Projects) {
		t.Error("Data integrity check failed: project count mismatch")
	}
}

// Test performance with large files (should be <5s)
func TestLargeFilePerformance(t *testing.T) {
	// Skip if in short mode
	if testing.Short() {
		t.Skip("Skipping large file test in short mode")
	}

	// Setup
	atomic := NewAtomicWriter("")
	osm := NewOptimizedStateManager(atomic)
	defer osm.Close()

	// Set higher memory limit for large file test
	osm.SetMemoryLimit(200 * 1024 * 1024) // 200MB

	// Create test file
	testData := createLargeTestData()
	testFile := "test_large.json"
	defer os.Remove(testFile)

	// Write test data
	start := time.Now()
	err := osm.WriteJSONOptimized(testFile, testData, &OptimizedWriteOptions{
		UseStreaming: true,
		ChunkSize:    ChunkSize * 2,
	})
	writeDuration := time.Since(start)

	if err != nil {
		t.Fatalf("Failed to write test data: %v", err)
	}

	t.Logf("Large file write performance: %v", writeDuration)

	// Check file size
	fileInfo, err := os.Stat(testFile)
	if err != nil {
		t.Fatalf("Failed to stat test file: %v", err)
	}
	t.Logf("Large file size: %d bytes (%.2f MB)", fileInfo.Size(), float64(fileInfo.Size())/1024/1024)

	// Test read performance with lazy loading
	start = time.Now()
	var readData MediumTestData
	err = osm.ReadJSONOptimized(testFile, &readData, &OptimizedReadOptions{
		UseStreaming:   true,
		EnableLazyLoad: true,
		ChunkSize:      ChunkSize * 2,
	})
	readDuration := time.Since(start)

	if err != nil {
		t.Fatalf("Failed to read test data: %v", err)
	}

	// Verify performance target (<5s for large files)
	if readDuration > LargeFileMaxDuration {
		t.Errorf("Large file read took %v, expected less than %v", readDuration, LargeFileMaxDuration)
	}

	t.Logf("Large file read performance: %v (target: <%v)", readDuration, LargeFileMaxDuration)

	// Verify data integrity
	if len(readData.Projects) != len(testData.Projects) {
		t.Error("Data integrity check failed: project count mismatch")
	}
}

// Test memory usage monitoring
func TestMemoryUsageMonitoring(t *testing.T) {
	// Setup
	atomic := NewAtomicWriter("")
	osm := NewOptimizedStateManager(atomic)
	defer osm.Close()

	// Set a lower memory limit for testing
	osm.SetMemoryLimit(10 * 1024 * 1024) // 10MB

	// Create medium test data
	testData := createMediumTestData()
	testFile := "test_memory.json"
	defer os.Remove(testFile)

	// Write and read data
	err := osm.WriteJSONOptimized(testFile, testData, nil)
	if err != nil {
		t.Fatalf("Failed to write test data: %v", err)
	}

	var readData MediumTestData
	err = osm.ReadJSONOptimized(testFile, &readData, nil)
	if err != nil {
		t.Fatalf("Failed to read test data: %v", err)
	}

	// Check memory stats
	stats := osm.GetMemoryStats()
	if stats.OperationCount == 0 {
		t.Error("Memory monitoring should have tracked operations")
	}

	t.Logf("Memory stats - Operations: %d, Peak usage: %d bytes", 
		stats.OperationCount, stats.PeakUsage)
}

// Test performance metrics tracking
func TestPerformanceMetricsTracking(t *testing.T) {
	// Setup
	atomic := NewAtomicWriter("")
	osm := NewOptimizedStateManager(atomic)
	defer osm.Close()

	// Reset metrics
	osm.ResetPerformanceMetrics()

	// Create test data
	testData := createSmallTestData()
	testFile := "test_metrics.json"
	defer os.Remove(testFile)

	// Perform operations
	err := osm.WriteJSONOptimized(testFile, testData, &OptimizedWriteOptions{
		TrackPerformance: true,
	})
	if err != nil {
		t.Fatalf("Failed to write test data: %v", err)
	}

	var readData SmallTestData
	err = osm.ReadJSONOptimized(testFile, &readData, &OptimizedReadOptions{
		TrackPerformance: true,
	})
	if err != nil {
		t.Fatalf("Failed to read test data: %v", err)
	}

	// Check metrics
	metrics := osm.GetPerformanceMetrics()
	if metrics.ReadOperations != 1 {
		t.Errorf("Expected 1 read operation, got %d", metrics.ReadOperations)
	}
	if metrics.WriteOperations != 1 {
		t.Errorf("Expected 1 write operation, got %d", metrics.WriteOperations)
	}
	if metrics.AverageReadTime == 0 {
		t.Error("Average read time should be tracked")
	}

	t.Logf("Performance metrics - Reads: %d (avg: %v), Writes: %d (avg: %v)",
		metrics.ReadOperations, metrics.AverageReadTime,
		metrics.WriteOperations, metrics.AverageWriteTime)
}

// Benchmark read performance
func BenchmarkOptimizedRead(b *testing.B) {
	// Setup
	atomic := NewAtomicWriter("")
	osm := NewOptimizedStateManager(atomic)
	defer osm.Close()

	// Create test file
	testData := createMediumTestData()
	testFile := "bench_read.json"
	defer os.Remove(testFile)

	err := osm.WriteJSONOptimized(testFile, testData, nil)
	if err != nil {
		b.Fatalf("Failed to write test data: %v", err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var readData MediumTestData
		err := osm.ReadJSONOptimized(testFile, &readData, &OptimizedReadOptions{
			TrackPerformance: false, // Don't skew benchmark
		})
		if err != nil {
			b.Fatalf("Failed to read test data: %v", err)
		}
	}
}

// Benchmark write performance
func BenchmarkOptimizedWrite(b *testing.B) {
	// Setup
	atomic := NewAtomicWriter("")
	osm := NewOptimizedStateManager(atomic)
	defer osm.Close()

	testData := createMediumTestData()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		testFile := fmt.Sprintf("bench_write_%d.json", i)
		err := osm.WriteJSONOptimized(testFile, testData, &OptimizedWriteOptions{
			TrackPerformance: false, // Don't skew benchmark
		})
		if err != nil {
			b.Fatalf("Failed to write test data: %v", err)
		}
		os.Remove(testFile)
	}
}

// Test benchmark functionality
func TestBenchmarkReadPerformance(t *testing.T) {
	// Setup
	atomic := NewAtomicWriter("")
	osm := NewOptimizedStateManager(atomic)
	defer osm.Close()

	// Create test file
	testData := createSmallTestData()
	testFile := "test_benchmark.json"
	defer os.Remove(testFile)

	err := osm.WriteJSONOptimized(testFile, testData, nil)
	if err != nil {
		t.Fatalf("Failed to write test data: %v", err)
	}

	// Run benchmark
	benchmark, err := osm.BenchmarkReadPerformance(testFile, 5)
	if err != nil {
		t.Fatalf("Failed to run benchmark: %v", err)
	}

	// Verify benchmark results
	if benchmark.Iterations != 5 {
		t.Errorf("Expected 5 iterations, got %d", benchmark.Iterations)
	}

	if len(benchmark.Durations) == 0 {
		t.Error("No duration measurements recorded")
	}

	if benchmark.AverageDuration == 0 {
		t.Error("Average duration should be calculated")
	}

	t.Logf("Benchmark results - Average: %v, Target: %v, Meets target: %v",
		benchmark.AverageDuration, benchmark.TargetDuration, benchmark.MeetsTarget)
}

// Test file size thresholds
func TestFileSizeThresholds(t *testing.T) {
	tests := []struct {
		name      string
		size      int64
		expected  string
	}{
		{"Small file", 500 * 1024, "small"},          // 500KB
		{"Medium file", 5 * 1024 * 1024, "medium"},   // 5MB
		{"Large file", 50 * 1024 * 1024, "large"},    // 50MB
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var strategy string
			switch {
			case tt.size <= SmallFileThreshold:
				strategy = "small"
			case tt.size <= MediumFileThreshold:
				strategy = "medium"
			default:
				strategy = "large"
			}

			if strategy != tt.expected {
				t.Errorf("Expected strategy %s for size %d, got %s", 
					tt.expected, tt.size, strategy)
			}
		})
	}
}

// Test error handling in performance operations
func TestPerformanceErrorHandling(t *testing.T) {
	// Setup
	atomic := NewAtomicWriter("")
	osm := NewOptimizedStateManager(atomic)
	defer osm.Close()

	// Test reading non-existent file
	var readData SmallTestData
	err := osm.ReadJSONOptimized("nonexistent.json", &readData, nil)
	if err == nil {
		t.Error("Expected error reading non-existent file")
	}

	// Test invalid JSON structure
	invalidFile := "invalid.json"
	os.WriteFile(invalidFile, []byte("invalid json content"), 0644)
	defer os.Remove(invalidFile)

	err = osm.ReadJSONOptimized(invalidFile, &readData, nil)
	if err == nil {
		t.Error("Expected error reading invalid JSON")
	}
}