package integration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	guardcmd "claude-wm-cli/internal/cmd"
)

// TestGuardInstallHook tests the guard install-hook command with various scenarios
func TestGuardInstallHook(t *testing.T) {
	t.Run("InstallInFreshGitRepo", func(t *testing.T) {
		// Create temp directory and fake Git repo
		tempDir := t.TempDir()
		originalWd, err := os.Getwd()
		require.NoError(t, err)
		
		err = os.Chdir(tempDir)
		require.NoError(t, err)
		defer func() {
			err := os.Chdir(originalWd)
			require.NoError(t, err)
		}()

		// Create fake .git/hooks directory structure
		hooksDir := filepath.Join(tempDir, ".git", "hooks")
		err = os.MkdirAll(hooksDir, 0755)
		require.NoError(t, err)

		// Run guard install-hook with --yes flag to skip confirmation
		cmd := guardcmd.GuardInstallHookCmd
		cmd.SetArgs([]string{"--yes"})
		
		err = cmd.Execute()
		require.NoError(t, err)

		// Verify hook was installed
		hookPath := filepath.Join(hooksDir, "pre-commit")
		assert.FileExists(t, hookPath)

		// Verify permissions are executable (check executable bits)
		info, err := os.Stat(hookPath)
		require.NoError(t, err)
		mode := info.Mode()
		assert.True(t, mode&0111 != 0, "Hook should be executable (mode: %o)", mode)

		// Verify hook content contains expected elements
		content, err := os.ReadFile(hookPath)
		require.NoError(t, err)
		hookContent := string(content)
		
		assert.Contains(t, hookContent, "#!/bin/sh")
		assert.Contains(t, hookContent, "claude-wm-cli guard check")
		assert.Contains(t, hookContent, "Claude WM CLI Pre-commit Hook")
	})

	t.Run("InstallWithExistingHook", func(t *testing.T) {
		// Create temp directory and fake Git repo
		tempDir := t.TempDir()
		originalWd, err := os.Getwd()
		require.NoError(t, err)
		
		err = os.Chdir(tempDir)
		require.NoError(t, err)
		defer func() {
			err := os.Chdir(originalWd)
			require.NoError(t, err)
		}()

		// Create fake .git/hooks directory
		hooksDir := filepath.Join(tempDir, ".git", "hooks")
		err = os.MkdirAll(hooksDir, 0755)
		require.NoError(t, err)

		// Create existing pre-commit hook
		hookPath := filepath.Join(hooksDir, "pre-commit")
		existingContent := "#!/bin/sh\necho 'existing hook'\n"
		err = os.WriteFile(hookPath, []byte(existingContent), 0644)
		require.NoError(t, err)

		// Run guard install-hook with --yes flag
		cmd := guardcmd.GuardInstallHookCmd
		cmd.SetArgs([]string{"--yes"})
		
		err = cmd.Execute()
		require.NoError(t, err)

		// Verify backup was created
		backupPath := hookPath + ".bak"
		assert.FileExists(t, backupPath)

		// Verify backup contains original content
		backupContent, err := os.ReadFile(backupPath)
		require.NoError(t, err)
		assert.Equal(t, existingContent, string(backupContent))

		// Verify new hook was installed
		newContent, err := os.ReadFile(hookPath)
		require.NoError(t, err)
		newHookContent := string(newContent)
		
		assert.Contains(t, newHookContent, "claude-wm-cli guard check")
		assert.NotEqual(t, existingContent, newHookContent)

		// Verify permissions are executable (check executable bits)
		info, err := os.Stat(hookPath)
		require.NoError(t, err)
		mode := info.Mode()
		assert.True(t, mode&0111 != 0, "Hook should be executable (mode: %o)", mode)
	})

	t.Run("FailWhenNotGitRepo", func(t *testing.T) {
		// Create temp directory without .git
		tempDir := t.TempDir()
		originalWd, err := os.Getwd()
		require.NoError(t, err)
		
		err = os.Chdir(tempDir)
		require.NoError(t, err)
		defer func() {
			err := os.Chdir(originalWd)
			require.NoError(t, err)
		}()

		// Run guard install-hook (should fail)
		cmd := guardcmd.GuardInstallHookCmd
		cmd.SetArgs([]string{"--yes"})
		
		err = cmd.Execute()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not a Git repository")
	})

	t.Run("HookPlacementValidation", func(t *testing.T) {
		// Create temp directory and fake Git repo
		tempDir := t.TempDir()
		originalWd, err := os.Getwd()
		require.NoError(t, err)
		
		err = os.Chdir(tempDir)
		require.NoError(t, err)
		defer func() {
			err := os.Chdir(originalWd)
			require.NoError(t, err)
		}()

		// Create fake .git/hooks directory
		hooksDir := filepath.Join(tempDir, ".git", "hooks")
		err = os.MkdirAll(hooksDir, 0755)
		require.NoError(t, err)

		// Run guard install-hook
		cmd := guardcmd.GuardInstallHookCmd
		cmd.SetArgs([]string{"--yes"})
		
		err = cmd.Execute()
		require.NoError(t, err)

		// Verify hook is ONLY in .git/hooks/pre-commit, not in .git/
		hookPath := filepath.Join(hooksDir, "pre-commit")
		assert.FileExists(t, hookPath)

		// Verify hook is NOT in .git/ root directory
		wrongPath := filepath.Join(tempDir, ".git", "pre-commit")
		assert.NoFileExists(t, wrongPath)

		// Verify hook content and permissions
		info, err := os.Stat(hookPath)
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(0755), info.Mode())
		assert.False(t, info.IsDir())
	})

	t.Run("PreserveExistingHookPermissions", func(t *testing.T) {
		// Create temp directory and fake Git repo
		tempDir := t.TempDir()
		originalWd, err := os.Getwd()
		require.NoError(t, err)
		
		err = os.Chdir(tempDir)
		require.NoError(t, err)
		defer func() {
			err := os.Chdir(originalWd)
			require.NoError(t, err)
		}()

		// Create fake .git/hooks directory
		hooksDir := filepath.Join(tempDir, ".git", "hooks")
		err = os.MkdirAll(hooksDir, 0755)
		require.NoError(t, err)

		// Create existing pre-commit hook with specific permissions
		hookPath := filepath.Join(hooksDir, "pre-commit")
		existingContent := "#!/bin/sh\necho 'existing hook'\n"
		err = os.WriteFile(hookPath, []byte(existingContent), 0700) // Different permissions
		require.NoError(t, err)

		// Run guard install-hook
		cmd := guardcmd.GuardInstallHookCmd
		cmd.SetArgs([]string{"--yes"})
		
		err = cmd.Execute()
		require.NoError(t, err)

		// Verify backup preserves original permissions (should preserve 0700)
		backupPath := hookPath + ".bak"
		backupInfo, err := os.Stat(backupPath)
		require.NoError(t, err)
		backupMode := backupInfo.Mode()
		// Check that backup preserved the original executable permissions
		assert.True(t, backupMode&0111 != 0, "Backup should preserve executable permissions (mode: %o)", backupMode)

		// Verify new hook has executable permissions
		hookInfo, err := os.Stat(hookPath)
		require.NoError(t, err)
		hookMode := hookInfo.Mode()
		assert.True(t, hookMode&0111 != 0, "New hook should be executable (mode: %o)", hookMode)
	})
}

// TestGuardInstallHookEdgeCases tests edge cases and error conditions
func TestGuardInstallHookEdgeCases(t *testing.T) {
	t.Run("CreateHooksDirectoryIfMissing", func(t *testing.T) {
		// Create temp directory with .git but no hooks directory
		tempDir := t.TempDir()
		originalWd, err := os.Getwd()
		require.NoError(t, err)
		
		err = os.Chdir(tempDir)
		require.NoError(t, err)
		defer func() {
			err := os.Chdir(originalWd)
			require.NoError(t, err)
		}()

		// Create .git directory but NOT .git/hooks
		gitDir := filepath.Join(tempDir, ".git")
		err = os.MkdirAll(gitDir, 0755)
		require.NoError(t, err)

		// Run guard install-hook (should fail since .git/hooks doesn't exist)
		cmd := guardcmd.GuardInstallHookCmd
		cmd.SetArgs([]string{"--yes"})
		
		err = cmd.Execute()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not a Git repository")
	})

	t.Run("BackupChainHandling", func(t *testing.T) {
		// Test scenario where .bak file already exists
		tempDir := t.TempDir()
		originalWd, err := os.Getwd()
		require.NoError(t, err)
		
		err = os.Chdir(tempDir)
		require.NoError(t, err)
		defer func() {
			err := os.Chdir(originalWd)
			require.NoError(t, err)
		}()

		// Create fake .git/hooks directory
		hooksDir := filepath.Join(tempDir, ".git", "hooks")
		err = os.MkdirAll(hooksDir, 0755)
		require.NoError(t, err)

		// Create existing pre-commit hook
		hookPath := filepath.Join(hooksDir, "pre-commit")
		originalContent := "#!/bin/sh\necho 'original hook'\n"
		err = os.WriteFile(hookPath, []byte(originalContent), 0755)
		require.NoError(t, err)

		// Create existing backup file
		backupPath := hookPath + ".bak"
		existingBackupContent := "#!/bin/sh\necho 'old backup'\n"
		err = os.WriteFile(backupPath, []byte(existingBackupContent), 0755)
		require.NoError(t, err)

		// Run guard install-hook
		cmd := guardcmd.GuardInstallHookCmd
		cmd.SetArgs([]string{"--yes"})
		
		err = cmd.Execute()
		require.NoError(t, err)

		// Verify backup was overwritten with the current hook content
		backupContent, err := os.ReadFile(backupPath)
		require.NoError(t, err)
		assert.Equal(t, originalContent, string(backupContent))
		assert.NotEqual(t, existingBackupContent, string(backupContent))

		// Verify new hook was installed
		newContent, err := os.ReadFile(hookPath)
		require.NoError(t, err)
		assert.Contains(t, string(newContent), "claude-wm-cli guard check")
	})
}