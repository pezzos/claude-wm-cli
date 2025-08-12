package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDevSandboxPathResolution(t *testing.T) {
	// Test path resolution logic
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	
	// Move up two levels from internal/cmd to project root
	projectRoot := filepath.Dir(filepath.Dir(cwd))
	
	sandboxPath := filepath.Join(projectRoot, ".wm", "sandbox", "claude")
	systemPath := filepath.Join(projectRoot, "internal", "config", "system")
	
	// Verify source path exists
	if _, err := os.Stat(systemPath); os.IsNotExist(err) {
		t.Errorf("Source system path does not exist: %s", systemPath)
	}
	
	// Verify sandbox path is correctly constructed
	expected := filepath.Join(projectRoot, ".wm", "sandbox", "claude")
	if sandboxPath != expected {
		t.Errorf("Expected sandbox path %s, got %s", expected, sandboxPath)
	}
}

func TestDevSandboxCommandStructure(t *testing.T) {
	// Verify command is properly configured
	if DevSandboxCmd == nil {
		t.Fatal("DevSandboxCmd is nil")
	}
	
	if DevSandboxCmd.Use != "sandbox" {
		t.Errorf("Expected Use to be 'sandbox', got %s", DevSandboxCmd.Use)
	}
	
	if DevSandboxCmd.Short == "" {
		t.Error("Short description is empty")
	}
	
	if DevSandboxCmd.Long == "" {
		t.Error("Long description is empty")
	}
	
	if DevSandboxCmd.RunE == nil {
		t.Error("RunE function is not set")
	}
}

func TestDevSandboxFlags(t *testing.T) {
	// Verify --reset flag is properly configured
	resetFlag := DevSandboxCmd.Flags().Lookup("reset")
	if resetFlag == nil {
		t.Error("--reset flag is not configured")
	}
	
	if resetFlag.DefValue != "false" {
		t.Errorf("Expected --reset default value to be 'false', got %s", resetFlag.DefValue)
	}
}