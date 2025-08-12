package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type ManifestEntry struct {
	Path   string `json:"path"`
	Size   int64  `json:"size"`
	Sha256 string `json:"sha256"`
}

func main() {
	var entries []ManifestEntry

	err := filepath.WalkDir("internal/config/system", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		
		// Skip directories
		if d.IsDir() {
			return nil
		}
		
		// Skip hooks directory (different Go module)
		if strings.Contains(path, "/hooks/") {
			return nil
		}
		
		// Skip manifest.json itself to avoid circular dependency
		if strings.HasSuffix(path, "manifest.json") {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", path, err)
		}

		hash := sha256.Sum256(content)
		relPath := strings.TrimPrefix(path, "internal/config/")

		entries = append(entries, ManifestEntry{
			Path:   relPath,
			Size:   int64(len(content)),
			Sha256: hex.EncodeToString(hash[:]),
		})

		return nil
	})

	if err != nil {
		log.Fatalf("Error walking directory: %v", err)
	}

	// Sort by path for stability
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Path < entries[j].Path
	})

	// Generate JSON with proper formatting
	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		log.Fatalf("Error marshaling JSON: %v", err)
	}

	// Write to manifest file
	manifestPath := "internal/config/system/manifest.json"
	err = os.WriteFile(manifestPath, data, 0644)
	if err != nil {
		log.Fatalf("Error writing manifest: %v", err)
	}

	fmt.Printf("Generated manifest with %d entries at %s\n", len(entries), manifestPath)
}