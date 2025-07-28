//go:build windows

package locking

import (
	"os"
	"syscall"
	"time"
	"unsafe"
)

var (
	kernel32         = syscall.NewLazyDLL("kernel32.dll")
	procLockFileEx   = kernel32.NewProc("LockFileEx")
	procUnlockFileEx = kernel32.NewProc("UnlockFileEx")
)

const (
	LOCKFILE_EXCLUSIVE_LOCK   = 0x00000002
	LOCKFILE_FAIL_IMMEDIATELY = 0x00000001
	ERROR_LOCK_VIOLATION      = 33
)

// platformLock performs platform-specific file locking on Windows
func (fl *FileLock) platformLock(file *os.File) error {
	var flags uint32

	if fl.options.Type == LockExclusive {
		flags |= LOCKFILE_EXCLUSIVE_LOCK
	}

	if fl.options.NonBlocking {
		flags |= LOCKFILE_FAIL_IMMEDIATELY
	}

	// Create overlapped structure
	overlapped := &syscall.Overlapped{}

	// Lock the entire file (0 to max)
	ret, _, err := procLockFileEx.Call(
		uintptr(file.Fd()),
		uintptr(flags),
		uintptr(0),          // reserved
		uintptr(0xFFFFFFFF), // number of bytes to lock (low)
		uintptr(0xFFFFFFFF), // number of bytes to lock (high)
		uintptr(unsafe.Pointer(overlapped)),
	)

	if ret == 0 {
		if errno, ok := err.(syscall.Errno); ok {
			if errno == ERROR_LOCK_VIOLATION {
				return &LockConflictError{
					RequestedType: fl.options.Type,
					FilePath:      fl.filePath,
					Timestamp:     time.Now(),
				}
			}
		}
		return err
	}

	return nil
}

// platformUnlock performs platform-specific file unlocking on Windows
func (fl *FileLock) platformUnlock(file *os.File) error {
	// Create overlapped structure
	overlapped := &syscall.Overlapped{}

	// Unlock the entire file
	ret, _, err := procUnlockFileEx.Call(
		uintptr(file.Fd()),
		uintptr(0),          // reserved
		uintptr(0xFFFFFFFF), // number of bytes to unlock (low)
		uintptr(0xFFFFFFFF), // number of bytes to unlock (high)
		uintptr(unsafe.Pointer(overlapped)),
	)

	if ret == 0 {
		return err
	}

	return nil
}

// processExists checks if a process with the given PID exists on Windows
func (fl *FileLock) processExists(pid int) bool {
	if pid <= 0 {
		return false
	}

	// Try to open the process handle
	handle, err := syscall.OpenProcess(syscall.PROCESS_QUERY_INFORMATION, false, uint32(pid))
	if err != nil {
		return false
	}
	defer syscall.CloseHandle(handle)

	// Check if process is still running
	var exitCode uint32
	err = syscall.GetExitCodeProcess(handle, &exitCode)
	if err != nil {
		return false
	}

	// STILL_ACTIVE = 259
	return exitCode == 259
}
