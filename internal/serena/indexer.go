package serena

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Manifest represents the SHA256 manifest for documentation files
type Manifest map[string]string // path -> sha256

const (
	SerenaDir     = ".serena"
	ManifestFile  = "docs-manifest.json"  // Different name to avoid conflict with existing manifest.json
	DocsPattern   = "docs"
)

// BuildDocsManifest scans docs/ directory and computes SHA256 for all .md files
func BuildDocsManifest(root string) (Manifest, error) {
	manifest := make(Manifest)
	docsPath := filepath.Join(root, DocsPattern)
	
	err := filepath.Walk(docsPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		// Skip directories and non-markdown files
		if info.IsDir() || !strings.HasSuffix(strings.ToLower(path), ".md") {
			return nil
		}
		
		// Skip .serena directory itself
		if strings.Contains(path, SerenaDir) {
			return nil
		}
		
		// Compute relative path from root
		relPath, err := filepath.Rel(root, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path for %s: %w", path, err)
		}
		
		// Compute SHA256
		hash, err := computeFileSHA256(path)
		if err != nil {
			return fmt.Errorf("failed to compute SHA256 for %s: %w", path, err)
		}
		
		manifest[relPath] = hash
		return nil
	})
	
	if err != nil {
		return nil, fmt.Errorf("failed to build docs manifest: %w", err)
	}
	
	return manifest, nil
}

// LoadPrevManifest loads the previous manifest from .serena/docs-manifest.json
func LoadPrevManifest(root string) (Manifest, error) {
	manifestPath := filepath.Join(root, SerenaDir, ManifestFile)
	
	// If manifest doesn't exist, return empty manifest
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		return make(Manifest), nil
	}
	
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest file: %w", err)
	}
	
	var manifest Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to unmarshal manifest: %w", err)
	}
	
	return manifest, nil
}

// SaveManifest saves the manifest to .serena/docs-manifest.json
func SaveManifest(root string, manifest Manifest) error {
	serenaDir := filepath.Join(root, SerenaDir)
	
	// Create .serena directory if it doesn't exist
	if err := os.MkdirAll(serenaDir, 0755); err != nil {
		return fmt.Errorf("failed to create serena directory: %w", err)
	}
	
	manifestPath := filepath.Join(serenaDir, ManifestFile)
	
	// Marshal manifest to JSON with indentation
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal manifest: %w", err)
	}
	
	// Write to file
	if err := os.WriteFile(manifestPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write manifest file: %w", err)
	}
	
	return nil
}

// DeltaResult represents the differences between two manifests
type DeltaResult struct {
	Added    []string
	Modified []string
	Removed  []string
}

// Delta compares previous and current manifests to find changes
func Delta(prev, cur Manifest) DeltaResult {
	result := DeltaResult{
		Added:    make([]string, 0),
		Modified: make([]string, 0),
		Removed:  make([]string, 0),
	}
	
	// Find added and modified files
	for path, curHash := range cur {
		if prevHash, exists := prev[path]; !exists {
			result.Added = append(result.Added, path)
		} else if prevHash != curHash {
			result.Modified = append(result.Modified, path)
		}
	}
	
	// Find removed files
	for path := range prev {
		if _, exists := cur[path]; !exists {
			result.Removed = append(result.Removed, path)
		}
	}
	
	return result
}

// IndexWithSerena performs indexing of changed files with Serena
// Currently implemented as a stub with logging - can be extended for real Serena integration
func IndexWithSerena(paths []string) error {
	if len(paths) == 0 {
		log.Printf("[SERENA] No files to index")
		return nil
	}
	
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	log.Printf("[SERENA] %s - Starting indexation of %d files", timestamp, len(paths))
	
	for i, path := range paths {
		log.Printf("[SERENA] Indexing [%d/%d]: %s", i+1, len(paths), path)
		
		// TODO: Replace with actual Serena API call
		// Example: serenaClient.IndexFile(path)
		
		// Simulate processing time
		time.Sleep(10 * time.Millisecond)
	}
	
	log.Printf("[SERENA] %s - Completed indexation of %d files", time.Now().Format("2006-01-02 15:04:05"), len(paths))
	return nil
}

// RunIncrementalIndex performs the complete incremental indexing workflow
func RunIncrementalIndex(root string) error {
	log.Printf("[SERENA] Starting incremental indexing for docs/")
	
	// Load previous manifest
	prevManifest, err := LoadPrevManifest(root)
	if err != nil {
		return fmt.Errorf("failed to load previous manifest: %w", err)
	}
	
	// Build current manifest
	curManifest, err := BuildDocsManifest(root)
	if err != nil {
		return fmt.Errorf("failed to build current manifest: %w", err)
	}
	
	// Calculate delta
	delta := Delta(prevManifest, curManifest)
	
	// Log summary
	log.Printf("[SERENA] Files to process: %d added, %d modified, %d removed", 
		len(delta.Added), len(delta.Modified), len(delta.Removed))
	
	// Combine added and modified files for indexing
	filesToIndex := append(delta.Added, delta.Modified...)
	
	if len(filesToIndex) == 0 && len(delta.Removed) == 0 {
		log.Printf("[SERENA] No changes detected - skipping indexation")
		return nil
	}
	
	// Log detailed changes
	if len(delta.Added) > 0 {
		log.Printf("[SERENA] Added files: %v", delta.Added)
	}
	if len(delta.Modified) > 0 {
		log.Printf("[SERENA] Modified files: %v", delta.Modified)
	}
	if len(delta.Removed) > 0 {
		log.Printf("[SERENA] Removed files: %v", delta.Removed)
	}
	
	// Index changed files
	if err := IndexWithSerena(filesToIndex); err != nil {
		return fmt.Errorf("failed to index files with Serena: %w", err)
	}
	
	// Save updated manifest
	if err := SaveManifest(root, curManifest); err != nil {
		return fmt.Errorf("failed to save manifest: %w", err)
	}
	
	log.Printf("[SERENA] Incremental indexing completed successfully")
	return nil
}

// computeFileSHA256 computes SHA256 hash of a file
func computeFileSHA256(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()
	
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	
	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}