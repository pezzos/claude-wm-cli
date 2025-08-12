package fsutil

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// EnsureDir creates a directory and all parent directories with 0755 permissions
func EnsureDir(path string) error {
	if err := os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", path, err)
	}
	return nil
}

// CopyTreeFS recursively copies files from an embedded filesystem to disk
func CopyTreeFS(src fs.FS, srcRoot string, dst string) error {
	return fs.WalkDir(src, srcRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Calculate relative path from srcRoot
		relPath := strings.TrimPrefix(path, srcRoot)
		relPath = strings.TrimPrefix(relPath, "/")
		if relPath == "" {
			return nil // Skip the root directory itself
		}

		// Calculate destination path
		dstPath := filepath.Join(dst, relPath)

		if d.IsDir() {
			// Create directory
			return EnsureDir(dstPath)
		}

		// Copy file
		return copyFile(src, path, dstPath)
	})
}

// copyFile copies a single file from the embedded filesystem to disk
func copyFile(src fs.FS, srcPath, dstPath string) error {
	// Ensure destination directory exists
	if err := EnsureDir(filepath.Dir(dstPath)); err != nil {
		return err
	}

	// Open source file
	srcFile, err := src.Open(srcPath)
	if err != nil {
		return fmt.Errorf("failed to open source file %s: %w", srcPath, err)
	}
	defer srcFile.Close()

	// Create destination file
	dstFile, err := os.OpenFile(dstPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to create destination file %s: %w", dstPath, err)
	}
	defer dstFile.Close()

	// Copy content
	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("failed to copy content from %s to %s: %w", srcPath, dstPath, err)
	}

	return nil
}