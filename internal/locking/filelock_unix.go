//go:build unix || linux || darwin

package locking

import (
	"os"
	"syscall"
)

// platformLock performs platform-specific file locking on Unix systems
func (fl *FileLock) platformLock(file *os.File) error {
	var lockType int
	
	switch fl.options.Type {
	case LockExclusive:
		lockType = syscall.LOCK_EX
	case LockShared:
		lockType = syscall.LOCK_SH
	default:
		lockType = syscall.LOCK_EX
	}
	
	if fl.options.NonBlocking {
		lockType |= syscall.LOCK_NB
	}
	
	return syscall.Flock(int(file.Fd()), lockType)
}

// platformUnlock performs platform-specific file unlocking on Unix systems
func (fl *FileLock) platformUnlock(file *os.File) error {
	return syscall.Flock(int(file.Fd()), syscall.LOCK_UN)
}

// processExists checks if a process with the given PID exists on Unix systems
func (fl *FileLock) processExists(pid int) bool {
	if pid <= 0 {
		return false
	}
	
	// On Unix, we can send signal 0 to check if process exists
	err := syscall.Kill(pid, 0)
	if err == nil {
		return true
	}
	
	// ESRCH means no such process
	if err == syscall.ESRCH {
		return false
	}
	
	// EPERM means process exists but we don't have permission to signal it
	if err == syscall.EPERM {
		return true
	}
	
	// Other errors, assume process doesn't exist
	return false
}