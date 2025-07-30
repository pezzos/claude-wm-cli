// Package model provides unified validation functions using the CLIError system.
// This replaces the old internal/validation package with rich error context.
package model

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ValidationEngine provides centralized validation with rich error context.
type ValidationEngine struct {
	strictMode bool
}

// NewValidationEngine creates a new validation engine.
func NewValidationEngine(strictMode bool) *ValidationEngine {
	return &ValidationEngine{
		strictMode: strictMode,
	}
}

// ValidateCommand validates a command string for execution.
func (v *ValidationEngine) ValidateCommand(command string) error {
	if strings.TrimSpace(command) == "" {
		return NewValidationError("command cannot be empty").
			WithContext(command).
			WithSuggestions([]string{
				"Provide a valid command to execute",
				"Check command syntax and spelling",
				"Use 'help' to see available commands",
			})
	}

	if len(command) > 1000 {
		return NewValidationError("command is too long").
			WithContext(fmt.Sprintf("current: %d chars, max: 1000", len(command))).
			WithSuggestions([]string{
				"Shorten your command to under 1000 characters",
				"Break complex commands into multiple steps",
				"Use command aliases or shorter argument names",
			})
	}

	// Check for potentially dangerous patterns
	dangerousPatterns := map[string]string{
		"rm -rf":         "recursive file deletion",
		"sudo rm":        "elevated file deletion",
		"format c:":      "disk formatting",
		"del /f /s /q":   "forced recursive deletion",
		"> /dev/null":    "output redirection to null",
		"chmod -R 777":   "recursive permission change",
		"dd if=":         "direct disk access",
		"mkfs":           "filesystem creation",
		"fdisk":          "disk partitioning",
		"killall":        "process termination",
	}

	lowerCommand := strings.ToLower(command)
	for pattern, description := range dangerousPatterns {
		if strings.Contains(lowerCommand, pattern) {
			if v.strictMode {
				return NewValidationError("command contains dangerous pattern").
					WithContext(fmt.Sprintf("pattern: '%s' (%s)", pattern, description)).
					WithSuggestions([]string{
						"Review the command for safety",
						"Use --force to override (not recommended)",
						"Consider a safer alternative approach",
						"Run the command manually if you're certain",
					})
			} else {
				// In non-strict mode, create a warning-level error
				return NewValidationError("potentially dangerous command detected").
					WithContext(fmt.Sprintf("pattern: '%s' (%s)", pattern, description)).
					WithSuggestions([]string{
						"Double-check this command before execution",
						"Consider using safer alternatives",
						"Enable strict mode for enhanced safety",
					})
			}
		}
	}

	return nil
}

// ValidateProjectName validates a project name with rich context.
func (v *ValidationEngine) ValidateProjectName(name string) error {
	if strings.TrimSpace(name) == "" {
		return NewValidationError("project name cannot be empty").
			WithSuggestions([]string{
				"Provide a descriptive project name",
				"Use letters, numbers, hyphens, and underscores",
				"Example: 'my-awesome-project'",
			})
	}

	// Check for invalid characters with detailed context
	invalidChars := map[string]string{
		"/":  "forward slash (use hyphens instead)",
		"\\": "backslash (use hyphens instead)",
		":":  "colon (reserved for paths)",
		"*":  "asterisk (wildcard character)",
		"?":  "question mark (wildcard character)",
		"\"": "double quote (reserved character)",
		"<":  "less than (reserved character)",
		">":  "greater than (reserved character)",
		"|":  "pipe (reserved character)",
		" ":  "space (use hyphens or underscores)",
	}

	for char, description := range invalidChars {
		if strings.Contains(name, char) {
			return NewValidationError("project name contains invalid character").
				WithContext(fmt.Sprintf("character: '%s' (%s)", char, description)).
				WithSuggestions([]string{
					"Use only letters, numbers, hyphens, and underscores",
					"Replace spaces with hyphens (-) or underscores (_)",
					"Example: 'my-project-name' or 'my_project_name'",
				})
		}
	}

	if len(name) > 100 {
		return NewValidationError("project name is too long").
			WithContext(fmt.Sprintf("current: %d chars, max: 100", len(name))).
			WithSuggestions([]string{
				"Shorten the project name to under 100 characters",
				"Use abbreviations or shorter words",
				"Remove unnecessary words or suffixes",
			})
	}

	// Additional validation: check for reserved names
	reservedNames := []string{
		"con", "prn", "aux", "nul", // Windows reserved
		"com1", "com2", "com3", "com4", "com5", "com6", "com7", "com8", "com9",
		"lpt1", "lpt2", "lpt3", "lpt4", "lpt5", "lpt6", "lpt7", "lpt8", "lpt9",
		"test", "temp", "tmp", // Common reserved names
	}

	lowerName := strings.ToLower(name)
	for _, reserved := range reservedNames {
		if lowerName == reserved {
			return NewValidationError("project name is reserved").
				WithContext(fmt.Sprintf("'%s' is a reserved name", name)).
				WithSuggestions([]string{
					"Choose a different project name",
					"Add a prefix or suffix to make it unique",
					"Use a more descriptive name",
				})
		}
	}

	return nil
}

// ValidateTimeout validates timeout values with enhanced context.
func (v *ValidationEngine) ValidateTimeout(timeout int) error {
	if timeout < 1 {
		return NewValidationError("timeout must be positive").
			WithContext(fmt.Sprintf("provided: %d seconds", timeout)).
			WithSuggestions([]string{
				"Use a timeout of at least 1 second",
				"For quick operations, use 5-10 seconds",
				"For long operations, use 60-300 seconds",
			})
	}

	if timeout > 3600 {
		return NewValidationError("timeout is too large").
			WithContext(fmt.Sprintf("provided: %d seconds, max: 3600 (1 hour)", timeout)).
			WithSuggestions([]string{
				"Use a timeout of 3600 seconds (1 hour) or less",
				"Consider breaking long operations into smaller parts",
				"Use background processing for very long tasks",
			})
	}

	// Add warning for very short timeouts
	if timeout < 5 && v.strictMode {
		return NewValidationError("timeout may be too short").
			WithContext(fmt.Sprintf("provided: %d seconds", timeout)).
			WithSuggestions([]string{
				"Consider using at least 5 seconds for network operations",
				"Use 1-3 seconds only for local file operations",
				"Increase timeout if operations frequently fail",
			})
	}

	return nil
}

// ValidateRetries validates retry count with enhanced logic.
func (v *ValidationEngine) ValidateRetries(retries int) error {
	if retries < 0 {
		return NewValidationError("retry count cannot be negative").
			WithContext(fmt.Sprintf("provided: %d", retries)).
			WithSuggestions([]string{
				"Use 0 for no retries",
				"Use 1-3 retries for most operations",
				"Use up to 5 retries for unreliable networks",
			})
	}

	if retries > 10 {
		return NewValidationError("retry count is too high").
			WithContext(fmt.Sprintf("provided: %d, max: 10", retries)).
			WithSuggestions([]string{
				"Use 10 or fewer retries to avoid infinite loops",
				"Check if the operation is fundamentally failing",
				"Consider exponential backoff for many retries",
			})
	}

	// Add guidance for optimal retry counts
	if retries > 5 && v.strictMode {
		return NewValidationError("high retry count detected").
			WithContext(fmt.Sprintf("provided: %d retries", retries)).
			WithSuggestions([]string{
				"Consider using 3-5 retries for most operations",
				"High retry counts may indicate a deeper issue",
				"Use exponential backoff with many retries",
			})
	}

	return nil
}

// ValidateConfigFile validates configuration file path with rich validation.
func (v *ValidationEngine) ValidateConfigFile(path string) error {
	if path == "" {
		return nil // Empty path is allowed (uses default)
	}

	// Normalize path for better error messages
	absPath, err := filepath.Abs(path)
	if err != nil {
		return NewValidationError("invalid file path").
			WithCause(err).
			WithContext(path).
			WithSuggestions([]string{
				"Check if the path contains valid characters",
				"Use absolute paths to avoid ambiguity",
				"Verify the directory structure exists",
			})
	}

	// Check if file exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return NewNotFoundError("configuration file").
			WithContext(absPath).
			WithSuggestions([]string{
				"Create the configuration file first",
				"Check if the path is correct",
				"Use 'config init' to create a default config",
				fmt.Sprintf("Ensure the directory '%s' exists", filepath.Dir(absPath)),
			})
	}

	// Check if it's readable
	file, err := os.Open(absPath)
	if err != nil {
		return NewFileSystemError("read", absPath, err).
			WithSuggestions([]string{
				"Check file permissions",
				"Ensure you have read access to the file",
				"Verify the file is not locked by another process",
			})
	}
	file.Close()

	// Check file extension with detailed guidance
	ext := filepath.Ext(absPath)
	validExts := map[string]string{
		".yaml": "YAML configuration format",
		".yml":  "YAML configuration format (short)",
		".json": "JSON configuration format",
		".toml": "TOML configuration format",
	}

	isValid := false
	var validExtList []string
	for validExt, description := range validExts {
		validExtList = append(validExtList, fmt.Sprintf("%s (%s)", validExt, description))
		if strings.EqualFold(ext, validExt) {
			isValid = true
			break
		}
	}

	if !isValid {
		return NewValidationError("unsupported configuration file format").
			WithContext(fmt.Sprintf("file: %s, extension: %s", absPath, ext)).
			WithSuggestions(append([]string{
				"Use one of the supported formats:",
			}, validExtList...))
	}

	// Additional validation: check file size
	if stat, err := os.Stat(absPath); err == nil {
		const maxSize = 10 * 1024 * 1024 // 10MB
		if stat.Size() > maxSize {
			return NewValidationError("configuration file is too large").
				WithContext(fmt.Sprintf("file: %s, size: %d bytes, max: %d", absPath, stat.Size(), maxSize)).
				WithSuggestions([]string{
					"Configuration files should be under 10MB",
					"Check if the file contains unexpected data",
					"Split large configurations into multiple files",
				})
		}

		if stat.Size() == 0 {
			return NewValidationError("configuration file is empty").
				WithContext(absPath).
				WithSuggestions([]string{
					"Add configuration data to the file",
					"Use 'config init' to create a default configuration",
					"Check if the file was corrupted",
				})
		}
	}

	return nil
}

// ValidateDirectory validates directory paths with comprehensive checks.
func (v *ValidationEngine) ValidateDirectory(path string) error {
	if path == "" {
		return NewValidationError("directory path cannot be empty").
			WithSuggestions([]string{
				"Provide a valid directory path",
				"Use '.' for current directory",
				"Use absolute paths to avoid ambiguity",
			})
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return NewValidationError("invalid directory path").
			WithCause(err).
			WithContext(path).
			WithSuggestions([]string{
				"Check if the path contains valid characters",
				"Use absolute paths when possible",
				"Verify the path structure is correct",
			})
	}

	// Check if directory exists
	stat, err := os.Stat(absPath)
	if os.IsNotExist(err) {
		return NewNotFoundError("directory").
			WithContext(absPath).
			WithSuggestions([]string{
				"Create the directory first: mkdir -p " + absPath,
				"Check if the path is correct",
				"Verify parent directories exist",
			})
	}

	if err != nil {
		return NewFileSystemError("access", absPath, err).
			WithSuggestions([]string{
				"Check directory permissions",
				"Ensure the path is accessible",
				"Verify the directory is not locked",
			})
	}

	// Check if it's actually a directory
	if !stat.IsDir() {
		return NewValidationError("path is not a directory").
			WithContext(fmt.Sprintf("path: %s (this is a file)", absPath)).
			WithSuggestions([]string{
				"Use a directory path, not a file path",
				"Check if you meant the parent directory",
				"Remove the file if you want to create a directory here",
			})
	}

	// Check if directory is writable (if in strict mode)
	if v.strictMode {
		testFile := filepath.Join(absPath, ".write_test_"+fmt.Sprintf("%d", os.Getpid()))
		if file, err := os.Create(testFile); err != nil {
			return NewFileSystemError("write", absPath, err).
				WithSuggestions([]string{
					"Check write permissions for the directory",
					"Ensure sufficient disk space",
					"Verify the directory is not read-only",
				})
		} else {
			file.Close()
			os.Remove(testFile) // Clean up test file
		}
	}

	return nil
}

// Composite validation functions

// ValidateProjectSetup validates a complete project setup.
func (v *ValidationEngine) ValidateProjectSetup(name, directory, configFile string) error {
	// Validate project name
	if err := v.ValidateProjectName(name); err != nil {
		return err
	}

	// Validate directory
	if err := v.ValidateDirectory(directory); err != nil {
		return err
	}

	// Validate config file if provided
	if configFile != "" {
		if err := v.ValidateConfigFile(configFile); err != nil {
			return err
		}
	}

	// Check if project already exists in directory
	projectPath := filepath.Join(directory, name)
	if _, err := os.Stat(projectPath); err == nil {
		return NewConflictError("project already exists").
			WithContext(projectPath).
			WithSuggestions([]string{
				"Choose a different project name",
				"Use a different directory",
				"Remove the existing project if you want to replace it",
			})
	}

	return nil
}

// ValidateExecutionEnvironment validates the environment for command execution.
func (v *ValidationEngine) ValidateExecutionEnvironment(command string, timeout, retries int, workingDir string) error {
	// Validate command
	if err := v.ValidateCommand(command); err != nil {
		return err
	}

	// Validate timeout
	if err := v.ValidateTimeout(timeout); err != nil {
		return err
	}

	// Validate retries
	if err := v.ValidateRetries(retries); err != nil {
		return err
	}

	// Validate working directory
	if workingDir != "" {
		if err := v.ValidateDirectory(workingDir); err != nil {
			return err
		}
	}

	return nil
}

// Global validation functions for backward compatibility

var defaultValidator = NewValidationEngine(false)
var strictValidator = NewValidationEngine(true)

// ValidateCommand validates a command using the default validator.
func ValidateCommand(command string) error {
	return defaultValidator.ValidateCommand(command)
}

// ValidateProjectName validates a project name using the default validator.
func ValidateProjectName(name string) error {
	return defaultValidator.ValidateProjectName(name)
}

// ValidateTimeout validates a timeout using the default validator.
func ValidateTimeout(timeout int) error {
	return defaultValidator.ValidateTimeout(timeout)
}

// ValidateRetries validates retry count using the default validator.
func ValidateRetries(retries int) error {
	return defaultValidator.ValidateRetries(retries)
}

// ValidateConfigFile validates a config file using the default validator.
func ValidateConfigFile(path string) error {
	return defaultValidator.ValidateConfigFile(path)
}

// Strict validation functions

// ValidateCommandStrict validates a command using strict mode.
func ValidateCommandStrict(command string) error {
	return strictValidator.ValidateCommand(command)
}

// ValidateProjectNameStrict validates a project name using strict mode.
func ValidateProjectNameStrict(name string) error {
	return strictValidator.ValidateProjectName(name)
}

// HandleValidationError provides backward compatibility for error handling.
// This function is deprecated - use the rich CLIError system instead.
func HandleValidationError(err error, suggestedCommand string) {
	if cliErr, ok := err.(*CLIError); ok {
		fmt.Fprintf(os.Stderr, "❌ %s\n", cliErr.Message)

		if cliErr.Context != "" {
			fmt.Fprintf(os.Stderr, "   Context: %s\n", cliErr.Context)
		}

		if len(cliErr.Suggestions) > 0 {
			fmt.Fprintf(os.Stderr, "\n💡 Suggestions:\n")
			for _, suggestion := range cliErr.Suggestions {
				fmt.Fprintf(os.Stderr, "   - %s\n", suggestion)
			}
		}

		if suggestedCommand != "" {
			fmt.Fprintf(os.Stderr, "\n💡 Try: %s\n", suggestedCommand)
		}

		fmt.Fprintf(os.Stderr, "\n📖 Use --help for more information.\n")
		os.Exit(cliErr.ExitCode())
	} else {
		fmt.Fprintf(os.Stderr, "❌ Error: %s\n", err.Error())
		os.Exit(1)
	}
}