package state

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"

	"claude-wm-cli/internal/errors"
)

// Performance thresholds
const (
	// File size thresholds for different optimization strategies
	SmallFileThreshold  = 1 * 1024 * 1024    // 1MB - load entirely into memory
	MediumFileThreshold = 10 * 1024 * 1024   // 10MB - use streaming with chunking
	LargeFileThreshold  = 100 * 1024 * 1024  // 100MB - use aggressive lazy loading

	// Memory usage limits
	MaxMemoryUsage     = 50 * 1024 * 1024  // 50MB max for state operations
	ChunkSize          = 64 * 1024         // 64KB chunks for streaming
	GCInterval         = 5 * time.Second   // How often to check memory and trigger GC

	// Performance targets (from task requirements)
	SmallFileMaxDuration = 1 * time.Second // Sub-second for normal files
	LargeFileMaxDuration = 5 * time.Second // <5s for large files
)

// MemoryStats tracks memory usage for state operations
type MemoryStats struct {
	AllocatedBytes   uint64        `json:"allocated_bytes"`
	TotalAllocBytes  uint64        `json:"total_alloc_bytes"`
	NumGC            uint32        `json:"num_gc"`
	PauseTime        time.Duration `json:"pause_time"`
	LastCheckTime    time.Time     `json:"last_check_time"`
	PeakUsage        uint64        `json:"peak_usage"`
	OperationCount   int64         `json:"operation_count"`
}

// PerformanceMetrics tracks performance of state operations
type PerformanceMetrics struct {
	ReadOperations   int64         `json:"read_operations"`
	WriteOperations  int64         `json:"write_operations"`
	TotalReadTime    time.Duration `json:"total_read_time"`
	TotalWriteTime   time.Duration `json:"total_write_time"`
	AverageReadTime  time.Duration `json:"average_read_time"`
	AverageWriteTime time.Duration `json:"average_write_time"`
	LargestFileSize  int64         `json:"largest_file_size"`
	SmallestFileSize int64         `json:"smallest_file_size"`
	MemoryStats      MemoryStats   `json:"memory_stats"`
	LastReset        time.Time     `json:"last_reset"`
}

// LazyStateSection represents a section of state that can be loaded on demand
type LazyStateSection struct {
	Key           string      `json:"key"`
	Offset        int64       `json:"offset"`
	Length        int64       `json:"length"`
	Loaded        bool        `json:"loaded"`
	Data          interface{} `json:"data,omitempty"`
	LastAccessed  time.Time   `json:"last_accessed"`
	AccessCount   int64       `json:"access_count"`
	Checksum      string      `json:"checksum"`
}

// OptimizedStateManager provides performance-optimized state operations
type OptimizedStateManager struct {
	atomic        *AtomicWriter
	metrics       *PerformanceMetrics
	memoryLimit   uint64
	gcEnabled     bool
	gcTicker      *time.Ticker
	mu            sync.RWMutex
	lazySections  map[string]*LazyStateSection
	memoryMonitor *MemoryMonitor
	ctx           context.Context
	cancel        context.CancelFunc
}

// MemoryMonitor tracks and controls memory usage
type MemoryMonitor struct {
	limit        uint64
	checkTicker  *time.Ticker
	forceGC      bool
	alertThresh  float64 // Percentage of limit that triggers alerts
	mu           sync.RWMutex
	stats        MemoryStats
	alerts       []MemoryAlert
}

// MemoryAlert represents a memory usage alert
type MemoryAlert struct {
	Timestamp   time.Time `json:"timestamp"`
	UsageBytes  uint64    `json:"usage_bytes"`
	LimitBytes  uint64    `json:"limit_bytes"`
	Percentage  float64   `json:"percentage"`
	Operation   string    `json:"operation"`
	Severity    string    `json:"severity"` // info, warning, critical
}

// OptimizedReadOptions configures optimized read operations
type OptimizedReadOptions struct {
	UseStreaming     bool          `json:"use_streaming"`
	EnableLazyLoad   bool          `json:"enable_lazy_load"`
	ChunkSize        int           `json:"chunk_size"`
	MemoryLimit      uint64        `json:"memory_limit"`
	Timeout          time.Duration `json:"timeout"`
	MaxSections      int           `json:"max_sections"`
	PreloadSections  []string      `json:"preload_sections"`
	TrackPerformance bool          `json:"track_performance"`
}

// OptimizedWriteOptions configures optimized write operations
type OptimizedWriteOptions struct {
	UseStreaming     bool          `json:"use_streaming"`
	ChunkSize        int           `json:"chunk_size"`
	MemoryLimit      uint64        `json:"memory_limit"`
	Timeout          time.Duration `json:"timeout"`
	CompressLarge    bool          `json:"compress_large"`
	BackgroundWrite  bool          `json:"background_write"`
	TrackPerformance bool          `json:"track_performance"`
}

// NewOptimizedStateManager creates a new performance-optimized state manager
func NewOptimizedStateManager(atomic *AtomicWriter) *OptimizedStateManager {
	ctx, cancel := context.WithCancel(context.Background())
	
	memMonitor := &MemoryMonitor{
		limit:       MaxMemoryUsage,
		checkTicker: time.NewTicker(GCInterval),
		forceGC:     true,
		alertThresh: 0.8, // Alert at 80% of limit
		alerts:      make([]MemoryAlert, 0),
	}

	osm := &OptimizedStateManager{
		atomic:        atomic,
		metrics:       &PerformanceMetrics{LastReset: time.Now()},
		memoryLimit:   MaxMemoryUsage,
		gcEnabled:     true,
		lazySections:  make(map[string]*LazyStateSection),
		memoryMonitor: memMonitor,
		ctx:           ctx,
		cancel:        cancel,
	}

	// Start memory monitoring
	go osm.monitorMemory()

	return osm
}

// ReadJSONOptimized reads JSON with performance optimizations
func (osm *OptimizedStateManager) ReadJSONOptimized(filePath string, target interface{}, opts *OptimizedReadOptions) error {
	startTime := time.Now()
	defer func() {
		if opts != nil && opts.TrackPerformance {
			osm.trackReadOperation(time.Since(startTime))
		}
	}()

	// Check file size to determine strategy
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return errors.ErrFileNotFound(filePath)
	}

	fileSize := fileInfo.Size()
	
	// Set default options if not provided
	if opts == nil {
		opts = osm.getDefaultReadOptions(fileSize)
	}

	// Apply timeout if specified
	ctx := osm.ctx
	if opts.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, opts.Timeout)
		defer cancel()
	}

	// Choose optimization strategy based on file size
	switch {
	case fileSize <= SmallFileThreshold:
		return osm.readSmallFile(ctx, filePath, target, opts)
	case fileSize <= MediumFileThreshold:
		return osm.readMediumFile(ctx, filePath, target, opts)
	default:
		return osm.readLargeFile(ctx, filePath, target, opts)
	}
}

// WriteJSONOptimized writes JSON with performance optimizations
func (osm *OptimizedStateManager) WriteJSONOptimized(filePath string, data interface{}, opts *OptimizedWriteOptions) error {
	startTime := time.Now()
	defer func() {
		if opts != nil && opts.TrackPerformance {
			osm.trackWriteOperation(time.Since(startTime))
		}
	}()

	// Set default options if not provided
	if opts == nil {
		opts = osm.getDefaultWriteOptions()
	}

	// Check memory usage before operation
	if err := osm.checkMemoryUsage("write"); err != nil {
		return err
	}

	// Apply timeout if specified
	ctx := osm.ctx
	if opts.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, opts.Timeout)
		defer cancel()
	}

	// Estimate data size to choose strategy
	dataSize := osm.estimateDataSize(data)

	switch {
	case dataSize <= SmallFileThreshold:
		return osm.writeSmallData(ctx, filePath, data, opts)
	case dataSize <= MediumFileThreshold:
		return osm.writeMediumData(ctx, filePath, data, opts)
	default:
		return osm.writeLargeData(ctx, filePath, data, opts)
	}
}

// readSmallFile handles small files with standard JSON parsing
func (osm *OptimizedStateManager) readSmallFile(ctx context.Context, filePath string, target interface{}, opts *OptimizedReadOptions) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Use atomic reader for small files
	return osm.atomic.ReadJSON(filePath, target)
}

// readMediumFile handles medium files with chunked streaming
func (osm *OptimizedStateManager) readMediumFile(ctx context.Context, filePath string, target interface{}, opts *OptimizedReadOptions) error {
	file, err := os.Open(filePath)
	if err != nil {
		return errors.ErrFileNotFound(filePath)
	}
	defer file.Close()

	// Use buffered reader for better performance
	chunkSize := opts.ChunkSize
	if chunkSize <= 0 {
		chunkSize = ChunkSize
	}

	reader := bufio.NewReaderSize(file, chunkSize)
	decoder := json.NewDecoder(reader)

	// Stream decode with memory monitoring
	return osm.streamDecode(ctx, decoder, target, opts)
}

// readLargeFile handles large files with lazy loading
func (osm *OptimizedStateManager) readLargeFile(ctx context.Context, filePath string, target interface{}, opts *OptimizedReadOptions) error {
	if !opts.EnableLazyLoad {
		// Fallback to medium file strategy but with larger chunks
		opts.ChunkSize = ChunkSize * 4
		return osm.readMediumFile(ctx, filePath, target, opts)
	}

	// Implement lazy loading strategy
	return osm.readWithLazyLoading(ctx, filePath, target, opts)
}

// writeSmallData handles small data with standard JSON marshaling
func (osm *OptimizedStateManager) writeSmallData(ctx context.Context, filePath string, data interface{}, opts *OptimizedWriteOptions) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Use atomic writer for small data
	atomicOpts := &AtomicWriteOptions{
		Permissions: 0644,
		Backup:      true,
		Verify:      true,
	}

	return osm.atomic.WriteJSON(filePath, data, atomicOpts)
}

// writeMediumData handles medium data with streaming
func (osm *OptimizedStateManager) writeMediumData(ctx context.Context, filePath string, data interface{}, opts *OptimizedWriteOptions) error {
	// Create temporary file for streaming write
	tempFile := filePath + ".tmp"
	
	file, err := os.Create(tempFile)
	if err != nil {
		return errors.ErrPermissionDenied(tempFile)
	}
	defer func() {
		file.Close()
		if err != nil {
			os.Remove(tempFile)
		}
	}()

	// Use buffered writer
	writer := bufio.NewWriterSize(file, opts.ChunkSize)
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")

	// Stream encode with memory monitoring
	if err := osm.streamEncode(ctx, encoder, data, opts); err != nil {
		return err
	}

	if err := writer.Flush(); err != nil {
		return fmt.Errorf("failed to flush write buffer: %w", err)
	}

	if err := file.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	// Atomic rename
	return os.Rename(tempFile, filePath)
}

// writeLargeData handles large data with advanced optimizations
func (osm *OptimizedStateManager) writeLargeData(ctx context.Context, filePath string, data interface{}, opts *OptimizedWriteOptions) error {
	// For very large data, break it into sections and write progressively
	return osm.writeWithSectioning(ctx, filePath, data, opts)
}

// streamDecode decodes JSON with memory monitoring
func (osm *OptimizedStateManager) streamDecode(ctx context.Context, decoder *json.Decoder, target interface{}, opts *OptimizedReadOptions) error {
	// Monitor memory usage during decode
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	done := make(chan error, 1)
	
	go func() {
		defer func() {
			if r := recover(); r != nil {
				done <- fmt.Errorf("JSON decode panic: %v", r)
			}
		}()
		done <- decoder.Decode(target)
	}()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := osm.checkMemoryUsage("read"); err != nil {
				return err
			}
		case err := <-done:
			return err
		}
	}
}

// streamEncode encodes JSON with memory monitoring
func (osm *OptimizedStateManager) streamEncode(ctx context.Context, encoder *json.Encoder, data interface{}, opts *OptimizedWriteOptions) error {
	// Monitor memory usage during encode
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	done := make(chan error, 1)
	
	go func() {
		defer func() {
			if r := recover(); r != nil {
				done <- fmt.Errorf("JSON encode panic: %v", r)
			}
		}()
		done <- encoder.Encode(data)
	}()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := osm.checkMemoryUsage("write"); err != nil {
				return err
			}
		case err := <-done:
			return err
		}
	}
}

// readWithLazyLoading implements lazy loading for large JSON files
func (osm *OptimizedStateManager) readWithLazyLoading(ctx context.Context, filePath string, target interface{}, opts *OptimizedReadOptions) error {
	// This is a simplified implementation of lazy loading
	// In a real implementation, we would need to parse the JSON structure
	// and create lazy sections for large arrays or objects
	
	file, err := os.Open(filePath)
	if err != nil {
		return errors.ErrFileNotFound(filePath)
	}
	defer file.Close()

	// For now, fall back to streaming with aggressive GC
	reader := bufio.NewReaderSize(file, ChunkSize*2)
	decoder := json.NewDecoder(reader)

	// Force garbage collection before and after
	runtime.GC()
	defer runtime.GC()

	return osm.streamDecode(ctx, decoder, target, opts)
}

// writeWithSectioning breaks large data into sections for writing
func (osm *OptimizedStateManager) writeWithSectioning(ctx context.Context, filePath string, data interface{}, opts *OptimizedWriteOptions) error {
	// Simplified implementation - in practice this would intelligently
	// break the data structure into logical sections
	return osm.writeMediumData(ctx, filePath, data, opts)
}

// GetPerformanceMetrics returns current performance metrics
func (osm *OptimizedStateManager) GetPerformanceMetrics() *PerformanceMetrics {
	osm.mu.RLock()
	defer osm.mu.RUnlock()

	// Create a copy to avoid race conditions
	metrics := *osm.metrics
	return &metrics
}

// ResetPerformanceMetrics resets all performance counters
func (osm *OptimizedStateManager) ResetPerformanceMetrics() {
	osm.mu.Lock()
	defer osm.mu.Unlock()

	osm.metrics = &PerformanceMetrics{LastReset: time.Now()}
}

// GetMemoryStats returns current memory statistics
func (osm *OptimizedStateManager) GetMemoryStats() MemoryStats {
	osm.memoryMonitor.mu.RLock()
	defer osm.memoryMonitor.mu.RUnlock()

	return osm.memoryMonitor.stats
}

// SetMemoryLimit sets the memory limit for state operations
func (osm *OptimizedStateManager) SetMemoryLimit(limit uint64) {
	osm.mu.Lock()
	defer osm.mu.Unlock()

	osm.memoryLimit = limit
	osm.memoryMonitor.limit = limit
}

// EnableGarbageCollection enables or disables automatic garbage collection
func (osm *OptimizedStateManager) EnableGarbageCollection(enabled bool) {
	osm.mu.Lock()
	defer osm.mu.Unlock()

	osm.gcEnabled = enabled
	osm.memoryMonitor.forceGC = enabled
}

// Close shuts down the optimized state manager
func (osm *OptimizedStateManager) Close() error {
	osm.cancel()
	
	if osm.memoryMonitor.checkTicker != nil {
		osm.memoryMonitor.checkTicker.Stop()
	}

	return nil
}

// Helper methods

func (osm *OptimizedStateManager) getDefaultReadOptions(fileSize int64) *OptimizedReadOptions {
	opts := &OptimizedReadOptions{
		ChunkSize:        ChunkSize,
		MemoryLimit:      osm.memoryLimit,
		Timeout:          SmallFileMaxDuration,
		TrackPerformance: true,
	}

	if fileSize > MediumFileThreshold {
		opts.UseStreaming = true
		opts.EnableLazyLoad = true
		opts.Timeout = LargeFileMaxDuration
		opts.ChunkSize = ChunkSize * 2
	} else if fileSize > SmallFileThreshold {
		opts.UseStreaming = true
		opts.ChunkSize = ChunkSize
	}

	return opts
}

func (osm *OptimizedStateManager) getDefaultWriteOptions() *OptimizedWriteOptions {
	return &OptimizedWriteOptions{
		ChunkSize:        ChunkSize,
		MemoryLimit:      osm.memoryLimit,
		Timeout:          SmallFileMaxDuration,
		TrackPerformance: true,
	}
}

func (osm *OptimizedStateManager) estimateDataSize(data interface{}) int64 {
	// Quick estimation by marshaling to JSON
	// In practice, this could be optimized with reflection-based size estimation
	jsonData, err := json.Marshal(data)
	if err != nil {
		return 0
	}
	return int64(len(jsonData))
}

func (osm *OptimizedStateManager) trackReadOperation(duration time.Duration) {
	osm.mu.Lock()
	defer osm.mu.Unlock()

	osm.metrics.ReadOperations++
	osm.metrics.TotalReadTime += duration
	osm.metrics.AverageReadTime = osm.metrics.TotalReadTime / time.Duration(osm.metrics.ReadOperations)
}

func (osm *OptimizedStateManager) trackWriteOperation(duration time.Duration) {
	osm.mu.Lock()
	defer osm.mu.Unlock()

	osm.metrics.WriteOperations++
	osm.metrics.TotalWriteTime += duration
	osm.metrics.AverageWriteTime = osm.metrics.TotalWriteTime / time.Duration(osm.metrics.WriteOperations)
}

func (osm *OptimizedStateManager) checkMemoryUsage(operation string) error {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	if m.Alloc > osm.memoryLimit {
		// Try to free memory with garbage collection
		runtime.GC()
		runtime.ReadMemStats(&m)

		if m.Alloc > osm.memoryLimit {
			return errors.NewCLIError("Memory limit exceeded", 1).
				WithDetails(fmt.Sprintf("Current usage: %d bytes, limit: %d bytes", m.Alloc, osm.memoryLimit)).
				WithContext("operation", operation).
				WithSuggestion("Reduce the size of state files or increase memory limit")
		}
	}

	// Update memory stats
	osm.memoryMonitor.mu.Lock()
	osm.memoryMonitor.stats.AllocatedBytes = m.Alloc
	osm.memoryMonitor.stats.TotalAllocBytes = m.TotalAlloc
	osm.memoryMonitor.stats.NumGC = m.NumGC
	osm.memoryMonitor.stats.LastCheckTime = time.Now()
	osm.memoryMonitor.stats.OperationCount++

	if m.Alloc > osm.memoryMonitor.stats.PeakUsage {
		osm.memoryMonitor.stats.PeakUsage = m.Alloc
	}

	// Check if we should create an alert
	percentage := float64(m.Alloc) / float64(osm.memoryLimit)
	if percentage > osm.memoryMonitor.alertThresh {
		severity := "warning"
		if percentage > 0.95 {
			severity = "critical"
		}

		alert := MemoryAlert{
			Timestamp:  time.Now(),
			UsageBytes: m.Alloc,
			LimitBytes: osm.memoryLimit,
			Percentage: percentage,
			Operation:  operation,
			Severity:   severity,
		}

		osm.memoryMonitor.alerts = append(osm.memoryMonitor.alerts, alert)

		// Keep only last 100 alerts
		if len(osm.memoryMonitor.alerts) > 100 {
			osm.memoryMonitor.alerts = osm.memoryMonitor.alerts[1:]
		}
	}

	osm.memoryMonitor.mu.Unlock()

	return nil
}

func (osm *OptimizedStateManager) monitorMemory() {
	ticker := time.NewTicker(GCInterval)
	defer ticker.Stop()

	for {
		select {
		case <-osm.ctx.Done():
			return
		case <-ticker.C:
			if osm.gcEnabled {
				var m runtime.MemStats
				runtime.ReadMemStats(&m)

				// Force GC if memory usage is high
				if m.Alloc > osm.memoryLimit/2 {
					runtime.GC()
				}
			}
		}
	}
}

// Benchmark functions for performance validation

// BenchmarkReadPerformance tests read performance against targets
func (osm *OptimizedStateManager) BenchmarkReadPerformance(filePath string, iterations int) (*PerformanceBenchmark, error) {
	benchmark := &PerformanceBenchmark{
		Operation:   "read",
		FilePath:    filePath,
		Iterations:  iterations,
		StartTime:   time.Now(),
	}

	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}

	benchmark.FileSize = fileInfo.Size()
	
	var totalDuration time.Duration
	var target interface{}

	for i := 0; i < iterations; i++ {
		start := time.Now()
		
		err := osm.ReadJSONOptimized(filePath, &target, &OptimizedReadOptions{
			TrackPerformance: false, // Don't skew metrics during benchmark
		})
		
		duration := time.Since(start)
		totalDuration += duration

		if err != nil {
			benchmark.Errors = append(benchmark.Errors, err.Error())
		} else {
			benchmark.Durations = append(benchmark.Durations, duration)
		}
	}

	benchmark.EndTime = time.Now()
	benchmark.TotalDuration = totalDuration
	benchmark.AverageDuration = totalDuration / time.Duration(iterations)

	// Check against performance targets
	target_duration := SmallFileMaxDuration
	if benchmark.FileSize > MediumFileThreshold {
		target_duration = LargeFileMaxDuration
	}

	benchmark.MeetsTarget = benchmark.AverageDuration <= target_duration
	benchmark.TargetDuration = target_duration

	return benchmark, nil
}

// PerformanceBenchmark contains benchmark results
type PerformanceBenchmark struct {
	Operation       string          `json:"operation"`
	FilePath        string          `json:"file_path"`
	FileSize        int64           `json:"file_size"`
	Iterations      int             `json:"iterations"`
	StartTime       time.Time       `json:"start_time"`
	EndTime         time.Time       `json:"end_time"`
	TotalDuration   time.Duration   `json:"total_duration"`
	AverageDuration time.Duration   `json:"average_duration"`
	Durations       []time.Duration `json:"durations"`
	Errors          []string        `json:"errors"`
	MeetsTarget     bool            `json:"meets_target"`
	TargetDuration  time.Duration   `json:"target_duration"`
}