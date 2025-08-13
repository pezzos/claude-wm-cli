package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"

	"claude-wm-cli/internal/serena"
)

func main() {
	var rootPath string
	flag.StringVar(&rootPath, "root", ".", "Root directory to scan for docs")
	flag.Parse()

	// Convert to absolute path
	absRoot, err := filepath.Abs(rootPath)
	if err != nil {
		log.Fatalf("Failed to get absolute path: %v", err)
	}

	// Ensure docs/ directory exists
	docsDir := filepath.Join(absRoot, "docs")
	if _, err := os.Stat(docsDir); os.IsNotExist(err) {
		log.Fatalf("docs/ directory not found in %s", absRoot)
	}

	log.Printf("Running Serena incremental indexer for: %s", absRoot)

	// Run incremental indexing
	if err := serena.RunIncrementalIndex(absRoot); err != nil {
		log.Fatalf("Incremental indexing failed: %v", err)
	}
}