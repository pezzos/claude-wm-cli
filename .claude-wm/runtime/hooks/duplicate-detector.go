package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"
)

// ToolInput represents the input from Claude Code
type ToolInput struct {
	ToolName  string                 `json:"tool_name"`
	ToolInput map[string]interface{} `json:"tool_input"`
}

// FileInfo represents information about a file for duplicate detection
type FileInfo struct {
	Path         string    `json:"path"`
	RelativePath string    `json:"relative_path"`
	Name         string    `json:"name"`
	Extension    string    `json:"extension"`
	Size         int64     `json:"size"`
	Hash         string    `json:"hash,omitempty"`
	ModTime      time.Time `json:"mod_time"`
	Route        string    `json:"route,omitempty"`
	Type         string    `json:"type"` // page, api, component, utility
}

// DuplicateDetector handles concurrent duplicate detection with optimizations
type DuplicateDetector struct {
	projectRoot   string
	cacheEnabled  bool
	workerCount   int
	maxFileSize   int64 // Files larger than this use mmap
	fileCache     map[string]*FileInfo
	cacheMutex    sync.RWMutex
	startTime     time.Time
}

// Result represents the detection result
type Result struct {
	Success   bool     `json:"success"`
	Errors    []string `json:"errors"`
	Warnings  []string `json:"warnings"`
	Duration  int64    `json:"duration_ms"`
	FilesScanned int   `json:"files_scanned"`
	CacheHits int      `json:"cache_hits"`
}

// NewDuplicateDetector creates a new optimized duplicate detector
func NewDuplicateDetector() *DuplicateDetector {
	return &DuplicateDetector{
		workerCount:  runtime.NumCPU(),
		maxFileSize:  10 * 1024 * 1024, // 10MB
		fileCache:    make(map[string]*FileInfo),
		cacheEnabled: os.Getenv("CACHE_ENABLED") == "true",
		startTime:    time.Now(),
	}
}

// FindProjectRoot finds the project root by looking for package.json
func (dd *DuplicateDetector) FindProjectRoot(startPath string) (string, error) {
	path, err := filepath.Abs(startPath)
	if err != nil {
		return "", err
	}

	for {
		packageJSON := filepath.Join(path, "package.json")
		if _, err := os.Stat(packageJSON); err == nil {
			return path, nil
		}

		parent := filepath.Dir(path)
		if parent == path {
			// Reached root directory
			break
		}
		path = parent
	}

	// Fallback to the starting path
	return filepath.Abs(startPath)
}

// ShouldSkipFile determines if a file should be skipped for performance
func (dd *DuplicateDetector) ShouldSkipFile(path string) bool {
	// Skip patterns for better performance
	skipPatterns := []string{
		"/node_modules/",
		"/.git/",
		"/.next/",
		"/dist/",
		"/build/",
		"/.nuxt/",
		"/vendor/",
		"/__pycache__/",
		".min.js",
		".min.css",
		".map",
	}

	for _, pattern := range skipPatterns {
		if strings.Contains(path, pattern) {
			return true
		}
	}

	// Skip binary files by extension
	ext := strings.ToLower(filepath.Ext(path))
	binaryExts := []string{
		".jpg", ".jpeg", ".png", ".gif", ".ico", ".svg",
		".mp4", ".mp3", ".avi", ".mov", ".pdf", ".zip",
		".tar", ".gz", ".exe", ".dll", ".so", ".dylib",
	}

	for _, binExt := range binaryExts {
		if ext == binExt {
			return true
		}
	}

	return false
}

// GetFileHash computes SHA256 hash of a file with optimizations
func (dd *DuplicateDetector) GetFileHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return "", err
	}

	hash := sha256.New()

	// For large files, use memory mapping if available
	if stat.Size() > dd.maxFileSize {
		// Try memory mapping for large files
		if data, err := dd.mmapFile(filePath); err == nil {
			hash.Write(data)
			return fmt.Sprintf("%x", hash.Sum(nil)), nil
		}
		// Fallback to regular read for large files
	}

	// Standard file reading
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

// mmapFile attempts to memory-map a file for reading
func (dd *DuplicateDetector) mmapFile(filePath string) ([]byte, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}

	// Memory map the file
	data, err := syscall.Mmap(int(file.Fd()), 0, int(stat.Size()), syscall.PROT_READ, syscall.MAP_PRIVATE)
	if err != nil {
		return nil, err
	}

	// Note: In a real implementation, we'd need to handle unmapping
	// For this hook's short-lived nature, we'll let the OS clean up
	return data, nil
}

// ScanFilesConcurrently scans files using worker pools for better performance
func (dd *DuplicateDetector) ScanFilesConcurrently(patterns []string) ([]*FileInfo, error) {
	// Channel for file paths to process
	filePaths := make(chan string, 100)
	// Channel for results
	results := make(chan *FileInfo, 100)
	// Channel for errors
	errors := make(chan error, 10)

	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < dd.workerCount; i++ {
		wg.Add(1)
		go dd.fileWorker(filePaths, results, errors, &wg)
	}

	// Start file discovery goroutine
	go func() {
		defer close(filePaths)
		dd.discoverFiles(patterns, filePaths)
	}()

	// Collect results
	go func() {
		wg.Wait()
		close(results)
		close(errors)
	}()

	var files []*FileInfo
	var scanErrors []error

	// Collect results and errors
	done := false
	for !done {
		select {
		case file, ok := <-results:
			if !ok {
				results = nil
			} else {
				files = append(files, file)
			}
		case err, ok := <-errors:
			if !ok {
				errors = nil
			} else {
				scanErrors = append(scanErrors, err)
			}
		}
		done = results == nil && errors == nil
	}

	if len(scanErrors) > 0 {
		return files, fmt.Errorf("scan errors: %v", scanErrors[0])
	}

	return files, nil
}

// fileWorker processes files in a worker pool
func (dd *DuplicateDetector) fileWorker(filePaths <-chan string, results chan<- *FileInfo, errors chan<- error, wg *sync.WaitGroup) {
	defer wg.Done()

	for filePath := range filePaths {
		if dd.ShouldSkipFile(filePath) {
			continue
		}

		fileInfo, err := dd.processFile(filePath)
		if err != nil {
			select {
			case errors <- err:
			default:
				// Error channel full, skip this error
			}
			continue
		}

		if fileInfo != nil {
			select {
			case results <- fileInfo:
			default:
				// Results channel full, this shouldn't happen with proper buffering
			}
		}
	}
}

// discoverFiles finds files matching patterns
func (dd *DuplicateDetector) discoverFiles(patterns []string, filePaths chan<- string) {
	for _, pattern := range patterns {
		searchPath := filepath.Join(dd.projectRoot, pattern)
		
		err := filepath.Walk(filepath.Dir(searchPath), func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil // Skip errors, don't stop the walk
			}

			if info.IsDir() {
				return nil
			}

			// Check if file matches any pattern
			if dd.matchesPatterns(path, patterns) {
				select {
				case filePaths <- path:
				case <-time.After(100 * time.Millisecond):
					// Timeout sending to channel, skip this file
				}
			}

			return nil
		})

		if err != nil {
			// Continue with other patterns even if one fails
			continue
		}
	}
}

// matchesPatterns checks if a file path matches any of the given patterns
func (dd *DuplicateDetector) matchesPatterns(filePath string, patterns []string) bool {
	for _, pattern := range patterns {
		if matched, _ := filepath.Match(pattern, filepath.Base(filePath)); matched {
			return true
		}
		
		// Also check full path for directory patterns
		if strings.Contains(filePath, strings.Replace(pattern, "*", "", -1)) {
			return true
		}
	}
	return false
}

// processFileForWriteHook processes a file that may not exist yet (for Write hook)
func (dd *DuplicateDetector) processFileForWriteHook(filePath string) (*FileInfo, error) {
	relPath, err := filepath.Rel(dd.projectRoot, filePath)
	if err != nil {
		relPath = filePath
	}

	fileInfo := &FileInfo{
		Path:         filePath,
		RelativePath: relPath,
		Name:         strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath)),
		Extension:    filepath.Ext(filePath),
		Size:         0, // File doesn't exist yet
		ModTime:      time.Now(),
	}

	// Determine file type and extract route information
	dd.classifyFile(fileInfo)

	return fileInfo, nil
}

// processFile processes a single file and extracts information
func (dd *DuplicateDetector) processFile(filePath string) (*FileInfo, error) {
	// Check cache first
	if dd.cacheEnabled {
		if cached := dd.getCachedFileInfo(filePath); cached != nil {
			return cached, nil
		}
	}

	stat, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}

	relPath, err := filepath.Rel(dd.projectRoot, filePath)
	if err != nil {
		relPath = filePath
	}

	fileInfo := &FileInfo{
		Path:         filePath,
		RelativePath: relPath,
		Name:         strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath)),
		Extension:    filepath.Ext(filePath),
		Size:         stat.Size(),
		ModTime:      stat.ModTime(),
	}

	// Determine file type and extract route information
	dd.classifyFile(fileInfo)

	// Compute hash for duplicate detection (only for small files by default)
	if stat.Size() < dd.maxFileSize {
		if hash, err := dd.GetFileHash(filePath); err == nil {
			fileInfo.Hash = hash
		}
	}

	// Cache the result
	if dd.cacheEnabled {
		dd.cacheFileInfo(filePath, fileInfo)
	}

	return fileInfo, nil
}

// classifyFile determines the type of file and extracts relevant information
func (dd *DuplicateDetector) classifyFile(fileInfo *FileInfo) {
	path := fileInfo.RelativePath
	
	// Next.js page detection
	if strings.Contains(path, "page.tsx") || strings.Contains(path, "page.jsx") {
		fileInfo.Type = "page"
		fileInfo.Route = dd.extractRouteFromPath(path)
	} else if strings.Contains(path, "route.ts") || strings.Contains(path, "route.js") {
		fileInfo.Type = "api"
		fileInfo.Route = dd.extractAPIRouteFromPath(path)
	} else if strings.Contains(path, "components") && (strings.HasSuffix(path, ".tsx") || strings.HasSuffix(path, ".jsx")) {
		fileInfo.Type = "component"
	} else if (strings.Contains(path, "utils") || strings.Contains(path, "lib") || strings.Contains(path, "helpers")) &&
		(strings.HasSuffix(path, ".ts") || strings.HasSuffix(path, ".js")) {
		fileInfo.Type = "utility"
	}
}

// extractRouteFromPath extracts route from Next.js app directory structure
func (dd *DuplicateDetector) extractRouteFromPath(path string) string {
	parts := strings.Split(filepath.ToSlash(path), "/")
	
	// Find app directory index
	appIndex := -1
	for i, part := range parts {
		if part == "app" {
			appIndex = i
			break
		}
	}
	
	if appIndex == -1 {
		return ""
	}
	
	// Build route from directory structure
	var routeParts []string
	for i := appIndex + 1; i < len(parts)-1; i++ { // Skip 'app' and filename
		part := parts[i]
		if strings.HasPrefix(part, "(") && strings.HasSuffix(part, ")") {
			// Route groups don't affect the URL
			continue
		}
		routeParts = append(routeParts, part)
	}
	
	if len(routeParts) == 0 {
		return "/"
	}
	return "/" + strings.Join(routeParts, "/")
}

// isRelevantForDuplicateDetection checks if a file is relevant for duplicate detection
func (dd *DuplicateDetector) isRelevantForDuplicateDetection(filePath string) bool {
	// Skip non-code files
	ext := strings.ToLower(filepath.Ext(filePath))
	relevantExtensions := []string{".ts", ".tsx", ".js", ".jsx", ".vue", ".svelte", ".php", ".py", ".rb", ".go", ".java", ".cs", ".cpp", ".c", ".h"}
	
	isRelevantExt := false
	for _, relevantExt := range relevantExtensions {
		if ext == relevantExt {
			isRelevantExt = true
			break
		}
	}
	
	if !isRelevantExt {
		return false
	}

	// Check if file is in a relevant directory or matches patterns
	lowerPath := strings.ToLower(filePath)
	relevantPatterns := []string{
		"page.tsx", "page.jsx", "page.ts", "page.js",
		"route.ts", "route.js", "route.tsx", "route.jsx",
		"components", "utils", "lib", "helpers", "services", "hooks",
		"pages", "app", "src", "controllers", "models", "views",
		"handlers", "middleware", "routes", "api", "endpoints",
	}

	for _, pattern := range relevantPatterns {
		if strings.Contains(lowerPath, pattern) {
			return true
		}
	}

	return false
}

// extractAPIRouteFromPath extracts API route from file path
func (dd *DuplicateDetector) extractAPIRouteFromPath(path string) string {
	parts := strings.Split(filepath.ToSlash(path), "/")
	
	// Find app directory index
	appIndex := -1
	for i, part := range parts {
		if part == "app" {
			appIndex = i
			break
		}
	}
	
	if appIndex == -1 {
		return ""
	}
	
	// Build API route
	var routeParts []string
	for i := appIndex + 1; i < len(parts)-1; i++ { // Skip 'app' and filename
		part := parts[i]
		if !(strings.HasPrefix(part, "(") && strings.HasSuffix(part, ")")) {
			routeParts = append(routeParts, part)
		}
	}
	
	return strings.Join(routeParts, "/")
}

// DetectDuplicates performs the main duplicate detection logic
func (dd *DuplicateDetector) DetectDuplicates(filePath string) *Result {
	result := &Result{
		Success:  true,
		Errors:   []string{},
		Warnings: []string{},
	}

	// Find project root
	var err error
	dd.projectRoot, err = dd.FindProjectRoot(filepath.Dir(filePath))
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to find project root: %v", err))
		result.Success = false
		return result
	}

	// Early exit: Skip if file is not relevant for duplicate detection
	if !dd.isRelevantForDuplicateDetection(filePath) {
		result.Duration = time.Since(dd.startTime).Milliseconds()
		return result
	}

	// Define search patterns based on file type
	var patterns []string
	
	if strings.Contains(filePath, "page.tsx") || strings.Contains(filePath, "page.jsx") {
		patterns = []string{"app/**/page.tsx", "app/**/page.jsx", "pages/**/*.tsx", "pages/**/*.jsx"}
	} else if strings.Contains(filePath, "route.ts") || strings.Contains(filePath, "route.js") {
		patterns = []string{"app/**/route.ts", "app/**/route.js", "pages/api/**/*.ts", "pages/api/**/*.js"}
	} else if strings.Contains(filePath, "components") {
		patterns = []string{"components/**/*.tsx", "components/**/*.jsx", "src/components/**/*.tsx", "src/components/**/*.jsx", "app/components/**/*.tsx", "app/components/**/*.jsx"}
	} else if strings.Contains(filePath, "utils") || strings.Contains(filePath, "lib") || strings.Contains(filePath, "helpers") {
		patterns = []string{"utils/**/*.ts", "utils/**/*.js", "lib/**/*.ts", "lib/**/*.js", "helpers/**/*.ts", "helpers/**/*.js", "src/utils/**/*.ts", "src/utils/**/*.js", "src/lib/**/*.ts", "src/lib/**/*.js"}
	} else {
		// Not a file type we care about
		result.Duration = time.Since(dd.startTime).Milliseconds()
		return result
	}

	// Scan files concurrently
	files, err := dd.ScanFilesConcurrently(patterns)
	if err != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf("Scan warning: %v", err))
	}

	result.FilesScanned = len(files)

	// Early exit: Skip if no files to compare against
	if len(files) < 1 {
		result.Duration = time.Since(dd.startTime).Milliseconds()
		return result
	}

	// Process the target file (create a virtual file info for non-existent files)
	targetFile, err := dd.processFileForWriteHook(filePath)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to process target file: %v", err))
		result.Success = false
		return result
	}

	// Early exit: Skip if target file has no route/name to compare
	if targetFile.Route == "" && targetFile.Name == "" {
		result.Duration = time.Since(dd.startTime).Milliseconds()
		return result
	}

	// Check for duplicates based on file type
	dd.checkForDuplicates(targetFile, files, result)

	result.Duration = time.Since(dd.startTime).Milliseconds()
	return result
}

// checkForDuplicates checks for various types of duplicates
func (dd *DuplicateDetector) checkForDuplicates(targetFile *FileInfo, files []*FileInfo, result *Result) {
	switch targetFile.Type {
	case "page":
		dd.checkPageDuplicates(targetFile, files, result)
	case "api":
		dd.checkAPIDuplicates(targetFile, files, result)
	case "component":
		dd.checkComponentDuplicates(targetFile, files, result)
	case "utility":
		dd.checkUtilityDuplicates(targetFile, files, result)
	}
}

// checkPageDuplicates checks for duplicate page routes
func (dd *DuplicateDetector) checkPageDuplicates(targetFile *FileInfo, files []*FileInfo, result *Result) {
	if targetFile.Route == "" {
		return
	}

	for _, file := range files {
		if file.Path == targetFile.Path {
			continue
		}
		if file.Type == "page" && file.Route == targetFile.Route {
			result.Errors = append(result.Errors, fmt.Sprintf("Duplicate route detected! Route '%s' already exists at: %s", targetFile.Route, file.RelativePath))
			result.Success = false
		}
	}
}

// checkAPIDuplicates checks for duplicate API routes
func (dd *DuplicateDetector) checkAPIDuplicates(targetFile *FileInfo, files []*FileInfo, result *Result) {
	if targetFile.Route == "" {
		return
	}

	for _, file := range files {
		if file.Path == targetFile.Path {
			continue
		}
		if file.Type == "api" && file.Route == targetFile.Route {
			result.Errors = append(result.Errors, fmt.Sprintf("Duplicate API route detected! API route '%s' already exists at: %s", targetFile.Route, file.RelativePath))
			result.Success = false
		}
	}
}

// checkComponentDuplicates checks for duplicate or similar components
func (dd *DuplicateDetector) checkComponentDuplicates(targetFile *FileInfo, files []*FileInfo, result *Result) {
	var exactMatches []string
	var similarMatches []string

	targetName := strings.ToLower(targetFile.Name)

	for _, file := range files {
		if file.Path == targetFile.Path {
			continue
		}
		if file.Type == "component" {
			fileName := strings.ToLower(file.Name)
			
			// Exact match (case-insensitive)
			if fileName == targetName {
				exactMatches = append(exactMatches, file.RelativePath)
			} else if strings.Contains(fileName, targetName) || strings.Contains(targetName, fileName) {
				// Similar match
				similarMatches = append(similarMatches, file.RelativePath)
			}
		}
	}

	if len(exactMatches) > 0 {
		result.Errors = append(result.Errors, fmt.Sprintf("Duplicate component detected! Component '%s' already exists at: %s", targetFile.Name, strings.Join(exactMatches, ", ")))
		result.Success = false
	}

	if len(similarMatches) > 0 {
		result.Warnings = append(result.Warnings, fmt.Sprintf("Similar components found: %s", strings.Join(similarMatches, ", ")))
		result.Warnings = append(result.Warnings, "Consider if you can use or extend an existing component")
	}
}

// checkUtilityDuplicates checks for duplicate utility files
func (dd *DuplicateDetector) checkUtilityDuplicates(targetFile *FileInfo, files []*FileInfo, result *Result) {
	targetName := strings.ToLower(targetFile.Name)

	for _, file := range files {
		if file.Path == targetFile.Path {
			continue
		}
		if file.Type == "utility" && strings.ToLower(file.Name) == targetName {
			result.Errors = append(result.Errors, fmt.Sprintf("Duplicate utility file detected! Utility '%s' already exists at: %s", targetFile.Name, file.RelativePath))
			result.Success = false
		}
	}
}

// Cache methods
func (dd *DuplicateDetector) getCachedFileInfo(filePath string) *FileInfo {
	if !dd.cacheEnabled {
		return nil
	}

	dd.cacheMutex.RLock()
	defer dd.cacheMutex.RUnlock()

	if cached, exists := dd.fileCache[filePath]; exists {
		// Validate cache by checking modification time
		if stat, err := os.Stat(filePath); err == nil {
			if stat.ModTime().Equal(cached.ModTime) {
				return cached
			}
		}
	}
	return nil
}

func (dd *DuplicateDetector) cacheFileInfo(filePath string, info *FileInfo) {
	if !dd.cacheEnabled {
		return
	}

	dd.cacheMutex.Lock()
	defer dd.cacheMutex.Unlock()

	dd.fileCache[filePath] = info
}

func main() {
	// Read input from stdin
	inputBytes, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
		os.Exit(1)
	}

	var input ToolInput
	if err := json.Unmarshal(inputBytes, &input); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing JSON input: %v\n", err)
		os.Exit(1)
	}

	// Only process Write tool for new files
	if input.ToolName != "Write" {
		os.Exit(0)
	}

	filePath, ok := input.ToolInput["file_path"].(string)
	if !ok || filePath == "" {
		os.Exit(0)
	}

	// Initialize detector
	detector := NewDuplicateDetector()

	// Run detection
	result := detector.DetectDuplicates(filePath)

	// Output results
	if len(result.Errors) > 0 {
		for _, error := range result.Errors {
			fmt.Fprintf(os.Stderr, "❌ %s\n", error)
		}
		fmt.Fprintf(os.Stderr, "\nCheck existing files before creating new ones.\n")
	}

	if len(result.Warnings) > 0 {
		for _, warning := range result.Warnings {
			fmt.Fprintf(os.Stderr, "⚠️  %s\n", warning)
		}
	}

	// Output JSON result for orchestrator
	resultJSON, _ := json.Marshal(result)
	fmt.Printf("DETECTION_RESULT: %s\n", resultJSON)

	// Exit with appropriate code
	if result.Success {
		os.Exit(0)
	} else {
		os.Exit(2)
	}
}