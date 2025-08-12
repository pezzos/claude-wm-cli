package ziputil

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// CreateBackup creates a ZIP archive of the source directory
func CreateBackup(sourceDir, backupPath string) error {
	// Create the backup directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(backupPath), 0755); err != nil {
		return fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Create the ZIP file
	zipFile, err := os.Create(backupPath)
	if err != nil {
		return fmt.Errorf("failed to create backup file %s: %w", backupPath, err)
	}
	defer zipFile.Close()

	// Create ZIP writer
	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// Walk through the source directory
	err = filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Get relative path from source directory
		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %w", err)
		}

		// Normalize path separators for ZIP format (always use forward slashes)
		relPath = filepath.ToSlash(relPath)

		// Create file in ZIP
		zipEntry, err := zipWriter.Create(relPath)
		if err != nil {
			return fmt.Errorf("failed to create ZIP entry for %s: %w", relPath, err)
		}

		// Open source file
		srcFile, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("failed to open source file %s: %w", path, err)
		}
		defer srcFile.Close()

		// Copy file content to ZIP
		if _, err := io.Copy(zipEntry, srcFile); err != nil {
			return fmt.Errorf("failed to write file %s to ZIP: %w", relPath, err)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	return nil
}

// CreateTimestampedBackup creates a backup with a timestamp in the filename
func CreateTimestampedBackup(sourceDir, backupDir string, timestamp string) (string, error) {
	// Generate backup filename
	backupName := fmt.Sprintf("backup-%s.zip", timestamp)
	backupPath := filepath.Join(backupDir, backupName)

	// Create the backup
	if err := CreateBackup(sourceDir, backupPath); err != nil {
		return "", err
	}

	return backupPath, nil
}

// RestoreFromBackup extracts a ZIP backup to a target directory
func RestoreFromBackup(backupPath, targetDir string) error {
	// Open ZIP file
	zipReader, err := zip.OpenReader(backupPath)
	if err != nil {
		return fmt.Errorf("failed to open backup file %s: %w", backupPath, err)
	}
	defer zipReader.Close()

	// Create target directory
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	// Extract files
	for _, file := range zipReader.File {
		// Skip directories (they'll be created as needed)
		if strings.HasSuffix(file.Name, "/") {
			continue
		}

		// Calculate target path
		targetPath := filepath.Join(targetDir, file.Name)

		// Create directory for file if needed
		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return fmt.Errorf("failed to create directory for %s: %w", targetPath, err)
		}

		// Open file in ZIP
		srcFile, err := file.Open()
		if err != nil {
			return fmt.Errorf("failed to open file %s in ZIP: %w", file.Name, err)
		}

		// Create target file
		targetFile, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			srcFile.Close()
			return fmt.Errorf("failed to create target file %s: %w", targetPath, err)
		}

		// Copy content
		if _, err := io.Copy(targetFile, srcFile); err != nil {
			srcFile.Close()
			targetFile.Close()
			return fmt.Errorf("failed to extract file %s: %w", file.Name, err)
		}

		srcFile.Close()
		targetFile.Close()
	}

	return nil
}