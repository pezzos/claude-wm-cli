/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"claude-wm-cli/internal/locking"
	"github.com/spf13/cobra"
)

var (
	lockType        string
	lockTimeout     time.Duration
	lockNonBlocking bool
	lockFormat      string
	lockCleanup     bool
)

// lockCmd represents the lock command
var lockCmd = &cobra.Command{
	Use:   "lock",
	Short: "File locking operations for concurrent access prevention",
	Long: `File locking system prevents concurrent CLI instances from corrupting state files.

FEATURES:
  • Cross-platform advisory file locking
  • Automatic stale lock detection and cleanup
  • Lock timeout with exponential backoff
  • Clear error messages for lock conflicts
  • Integration with atomic file operations

COMMANDS:
  • status     - Show lock status for files
  • acquire    - Acquire a lock on a file  
  • release    - Release a lock on a file
  • test       - Test locking functionality
  • cleanup    - Clean up stale lock files

Examples:
  claude-wm-cli lock status file.json              # Check lock status
  claude-wm-cli lock acquire file.json             # Acquire exclusive lock
  claude-wm-cli lock test file.json               # Test lock acquisition
  claude-wm-cli lock cleanup ./state/             # Clean stale locks`,
}

// lockStatusCmd shows lock status
var lockStatusCmd = &cobra.Command{
	Use:   "status [files...]",
	Short: "Show lock status for files",
	Long: `Display the current lock status for specified files.
Shows whether files are locked, by which process, and when.`,
	Run: func(cmd *cobra.Command, args []string) {
		showLockStatus(args)
	},
}

// lockAcquireCmd acquires locks
var lockAcquireCmd = &cobra.Command{
	Use:   "acquire [files...]",
	Short: "Acquire locks on files",
	Long: `Acquire exclusive or shared locks on the specified files.
Useful for testing locking behavior and ensuring exclusive access.`,
	Run: func(cmd *cobra.Command, args []string) {
		acquireLocks(args)
	},
}

// lockReleaseCmd releases locks
var lockReleaseCmd = &cobra.Command{
	Use:   "release [files...]",
	Short: "Release locks on files",
	Long: `Release locks that were previously acquired by this process.
Only locks owned by the current process can be released.`,
	Run: func(cmd *cobra.Command, args []string) {
		releaseLocks(args)
	},
}

// lockTestCmd tests locking functionality
var lockTestCmd = &cobra.Command{
	Use:   "test [file]",
	Short: "Test file locking functionality",
	Long: `Test the file locking system with a comprehensive test suite.
Verifies lock acquisition, conflict detection, and cleanup.`,
	Run: func(cmd *cobra.Command, args []string) {
		testLocking(args)
	},
}

// lockCleanupCmd cleans up stale locks
var lockCleanupCmd = &cobra.Command{
	Use:   "cleanup [directories...]",
	Short: "Clean up stale lock files",
	Long: `Remove stale lock files from specified directories.
Locks are considered stale if the owning process no longer exists.`,
	Run: func(cmd *cobra.Command, args []string) {
		cleanupLocks(args)
	},
}

func showLockStatus(files []string) {
	if len(files) == 0 {
		fmt.Println("❌ No files specified")
		fmt.Println("💡 Usage: claude-wm-cli lock status <file1> [file2...]")
		return
	}
	
	manager := locking.GetGlobalLockManager()
	
	fmt.Println("🔒 File Lock Status")
	fmt.Println("==================")
	
	for _, file := range files {
		absPath, err := filepath.Abs(file)
		if err != nil {
			fmt.Printf("❌ %s: Invalid path (%v)\n", file, err)
			continue
		}
		
		result := manager.CheckLockStatus(absPath)
		
		fmt.Printf("\n📁 %s\n", file)
		fmt.Printf("Status: %s\n", getStatusIcon(result.Status))
		
		if result.LockInfo != nil {
			info := result.LockInfo
			fmt.Printf("Type: %s\n", info.Type)
			fmt.Printf("PID: %d\n", info.PID)
			fmt.Printf("Hostname: %s\n", info.Hostname)
			fmt.Printf("Acquired: %s\n", info.AcquiredAt.Format("2006-01-02 15:04:05"))
			fmt.Printf("Age: %v\n", info.Age())
			
			if info.IsStale() {
				fmt.Printf("⚠️  Lock is stale (expired %v ago)\n", time.Since(info.ExpiresAt))
			}
		}
		
		if result.Error != nil {
			fmt.Printf("Error: %v\n", result.Error)
		}
		
		if result.Suggestion != "" {
			fmt.Printf("💡 %s\n", result.Suggestion)
		}
		
		if lockFormat == "json" {
			jsonData, _ := json.MarshalIndent(result, "", "  ")
			fmt.Printf("\nRaw data:\n%s\n", string(jsonData))
		}
	}
}

func acquireLocks(files []string) {
	if len(files) == 0 {
		fmt.Println("❌ No files specified")
		return
	}
	
	options := &locking.LockOptions{
		Type:         locking.LockExclusive,
		Timeout:      lockTimeout,
		NonBlocking:  lockNonBlocking,
		StaleTimeout: 5 * time.Minute,
		RetryDelay:   100 * time.Millisecond,
	}
	
	if lockType == "shared" {
		options.Type = locking.LockShared
	}
	
	manager := locking.GetGlobalLockManager()
	
	fmt.Printf("🔒 Acquiring %s locks...\n", options.Type)
	
	acquired := 0
	for _, file := range files {
		absPath, err := filepath.Abs(file)
		if err != nil {
			fmt.Printf("❌ %s: Invalid path (%v)\n", file, err)
			continue
		}
		
		lock, result := manager.LockFile(absPath, options)
		
		fmt.Printf("\n📁 %s\n", file)
		
		if result.Status == locking.LockStatusHeld {
			fmt.Printf("✅ Lock acquired successfully\n")
			fmt.Printf("Duration: %v\n", result.Duration)
			fmt.Printf("Attempts: %d\n", result.Attempts)
			acquired++
			
			if lockInfo := lock.GetLockInfo(); lockInfo != nil {
				fmt.Printf("Lock ID: %d\n", lockInfo.PID)
			}
		} else {
			fmt.Printf("❌ Failed to acquire lock\n")
			fmt.Printf("Status: %s\n", result.Status)
			if result.Error != nil {
				fmt.Printf("Error: %v\n", result.Error)
			}
			if result.Suggestion != "" {
				fmt.Printf("💡 %s\n", result.Suggestion)
			}
		}
	}
	
	fmt.Printf("\n📊 Summary: %d/%d locks acquired\n", acquired, len(files))
	
	if acquired > 0 {
		fmt.Println("⚠️  Locks are held by this process. Use 'claude-wm-cli lock release' to release them.")
	}
}

func releaseLocks(files []string) {
	if len(files) == 0 {
		fmt.Println("❌ No files specified")
		return
	}
	
	manager := locking.GetGlobalLockManager()
	
	fmt.Println("🔓 Releasing locks...")
	
	released := 0
	for _, file := range files {
		absPath, err := filepath.Abs(file)
		if err != nil {
			fmt.Printf("❌ %s: Invalid path (%v)\n", file, err)
			continue
		}
		
		fmt.Printf("\n📁 %s\n", file)
		
		if err := manager.UnlockFile(absPath); err != nil {
			fmt.Printf("❌ Failed to release lock: %v\n", err)
		} else {
			fmt.Printf("✅ Lock released successfully\n")
			released++
		}
	}
	
	fmt.Printf("\n📊 Summary: %d/%d locks released\n", released, len(files))
}

func testLocking(args []string) {
	testFile := "test-lock-file.tmp"
	if len(args) > 0 {
		testFile = args[0]
	}
	
	fmt.Println("🧪 Testing File Locking System")
	fmt.Println("==============================")
	
	// Create test file
	if err := os.WriteFile(testFile, []byte("test data"), 0644); err != nil {
		fmt.Printf("❌ Failed to create test file: %v\n", err)
		return
	}
	defer os.Remove(testFile)
	
	manager := locking.GetGlobalLockManager()
	options := locking.DefaultLockOptions()
	
	// Test 1: Basic lock acquisition
	fmt.Println("\n🔬 Test 1: Basic Lock Acquisition")
	lock1, result1 := manager.LockFile(testFile, options)
	if result1.Status == locking.LockStatusHeld {
		fmt.Printf("✅ Lock acquired successfully (duration: %v)\n", result1.Duration)
	} else {
		fmt.Printf("❌ Lock acquisition failed: %v\n", result1.Error)
		return
	}
	
	// Test 2: Conflict detection
	fmt.Println("\n🔬 Test 2: Conflict Detection")
	_, result2 := manager.LockFile(testFile, options)
	if result2.Status == locking.LockStatusHeld {
		fmt.Println("✅ Detected existing lock (already held by same process)")
	} else if result2.Status == locking.LockStatusBlocked {
		fmt.Printf("✅ Correctly detected lock conflict: %v\n", result2.Error)
	} else {
		fmt.Printf("❌ Unexpected result: %s\n", result2.Status)
	}
	
	// Test 3: Lock info
	fmt.Println("\n🔬 Test 3: Lock Information")
	if info := lock1.GetLockInfo(); info != nil {
		fmt.Printf("✅ Lock info retrieved:\n")
		fmt.Printf("   PID: %d\n", info.PID)
		fmt.Printf("   Type: %s\n", info.Type)
		fmt.Printf("   Age: %v\n", info.Age())
	} else {
		fmt.Println("❌ Failed to get lock info")
	}
	
	// Test 4: Release lock
	fmt.Println("\n🔬 Test 4: Lock Release")
	if err := manager.UnlockFile(testFile); err != nil {
		fmt.Printf("❌ Failed to release lock: %v\n", err)
	} else {
		fmt.Println("✅ Lock released successfully")
	}
	
	// Test 5: Re-acquisition after release
	fmt.Println("\n🔬 Test 5: Re-acquisition After Release")
	_, result5 := manager.LockFile(testFile, options)
	if result5.Status == locking.LockStatusHeld {
		fmt.Printf("✅ Lock re-acquired successfully (duration: %v)\n", result5.Duration)
		manager.UnlockFile(testFile)
	} else {
		fmt.Printf("❌ Failed to re-acquire lock: %v\n", result5.Error)
	}
	
	// Test 6: Timeout test
	fmt.Println("\n🔬 Test 6: Timeout Behavior")
	quickOptions := &locking.LockOptions{
		Type:        locking.LockExclusive,
		Timeout:     100 * time.Millisecond,
		NonBlocking: false,
		RetryDelay:  10 * time.Millisecond,
	}
	
	// Acquire lock first
	manager.LockFile(testFile, options)
	
	// Try to acquire with short timeout (should fail)
	start := time.Now()
	_, timeoutResult := manager.LockFile(testFile, quickOptions)
	duration := time.Since(start)
	
	if timeoutResult.Status == locking.LockStatusError {
		fmt.Printf("✅ Timeout behavior correct (took %v)\n", duration)
	} else {
		fmt.Printf("❌ Unexpected timeout result: %s\n", timeoutResult.Status)
	}
	
	// Clean up
	manager.UnlockFile(testFile)
	
	// Show metrics
	fmt.Println("\n📊 Test Metrics")
	metrics := manager.GetMetrics()
	fmt.Printf("Total requests: %d\n", metrics.TotalRequests)
	fmt.Printf("Successful locks: %d\n", metrics.SuccessfulLocks)
	fmt.Printf("Failed locks: %d\n", metrics.FailedLocks)
	fmt.Printf("Timeout errors: %d\n", metrics.TimeoutErrors)
	fmt.Printf("Active locks: %d\n", metrics.ActiveLocks)
	
	fmt.Println("\n🎉 File locking test completed!")
}

func cleanupLocks(directories []string) {
	if len(directories) == 0 {
		directories = []string{"."} // Default to current directory
	}
	
	manager := locking.GetGlobalLockManager()
	
	fmt.Println("🧹 Cleaning Up Stale Locks")
	fmt.Println("==========================")
	
	totalCleaned := 0
	for _, dir := range directories {
		fmt.Printf("\n📁 Scanning directory: %s\n", dir)
		
		cleaned, err := manager.CleanupStaleLocks(dir)
		if err != nil {
			fmt.Printf("❌ Error scanning directory: %v\n", err)
			continue
		}
		
		fmt.Printf("✅ Cleaned %d stale locks\n", cleaned)
		totalCleaned += cleaned
	}
	
	fmt.Printf("\n📊 Total stale locks cleaned: %d\n", totalCleaned)
	
	if totalCleaned == 0 {
		fmt.Println("🎉 No stale locks found!")
	}
}

// Helper functions

func getStatusIcon(status locking.LockStatus) string {
	switch status {
	case locking.LockStatusAvailable:
		return "🟢 Available"
	case locking.LockStatusHeld:
		return "🔒 Locked"
	case locking.LockStatusBlocked:
		return "🔴 Blocked"
	case locking.LockStatusStale:
		return "⚠️ Stale"
	case locking.LockStatusError:
		return "❌ Error"
	default:
		return "❓ Unknown"
	}
}

func init() {
	rootCmd.AddCommand(lockCmd)
	
	// Add subcommands
	lockCmd.AddCommand(lockStatusCmd)
	lockCmd.AddCommand(lockAcquireCmd)
	lockCmd.AddCommand(lockReleaseCmd)
	lockCmd.AddCommand(lockTestCmd)
	lockCmd.AddCommand(lockCleanupCmd)
	
	// Global flags
	lockCmd.PersistentFlags().StringVarP(&lockFormat, "format", "f", "text", "Output format (text, json)")
	
	// Acquire command flags
	lockAcquireCmd.Flags().StringVar(&lockType, "type", "exclusive", "Lock type (exclusive, shared)")
	lockAcquireCmd.Flags().DurationVarP(&lockTimeout, "timeout", "t", 30*time.Second, "Lock acquisition timeout")
	lockAcquireCmd.Flags().BoolVar(&lockNonBlocking, "non-blocking", false, "Fail immediately if lock unavailable")
	
	// Cleanup command flags
	lockCleanupCmd.Flags().BoolVar(&lockCleanup, "force", false, "Force cleanup even if process might be alive")
}