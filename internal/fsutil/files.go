package fsutil

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// CopyFileWithDir copies a file from src to dst, creating destination directory if needed
func CopyFileWithDir(src, dst string) error {
	// Ensure destination directory exists
	if err := EnsureDir(filepath.Dir(dst)); err != nil {
		return err
	}

	// Open source file
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file %s: %w", src, err)
	}
	defer srcFile.Close()

	// Create destination file
	dstFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to create destination file %s: %w", dst, err)
	}
	defer dstFile.Close()

	// Copy content
	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("failed to copy content from %s to %s: %w", src, dst, err)
	}

	return nil
}

// CopyFile copies a file from src to dst (without creating directories)
func CopyFile(src, dst string) error {
	// Open source file
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file %s: %w", src, err)
	}
	defer srcFile.Close()

	// Create destination file
	dstFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to create destination file %s: %w", dst, err)
	}
	defer dstFile.Close()

	// Copy content
	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("failed to copy content from %s to %s: %w", src, dst, err)
	}

	return nil
}

// CopyDirectory recursively copies a directory from src to dst
func CopyDirectory(src, dst string) error {
	// Verify source exists
	if _, err := os.Stat(src); err != nil {
		return fmt.Errorf("failed to stat source directory %s: %w", src, err)
	}

	// Create destination directory
	if err := EnsureDir(dst); err != nil {
		return err
	}

	// Walk through source directory
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Calculate relative path
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %w", err)
		}

		// Skip root directory
		if relPath == "." {
			return nil
		}

		// Calculate destination path
		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			// Create directory
			return EnsureDir(dstPath)
		}

		// Copy file
		return CopyFile(path, dstPath)
	})
}