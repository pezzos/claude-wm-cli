package validation

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ErrorCode represents standard CLI error codes
type ErrorCode int

const (
	ErrorSuccess         ErrorCode = 0
	ErrorGeneral         ErrorCode = 1
	ErrorInvalidInput    ErrorCode = 2
	ErrorFileNotFound    ErrorCode = 3
	ErrorPermissionDenied ErrorCode = 4
	ErrorTimeout         ErrorCode = 5
	ErrorNetworkFailure  ErrorCode = 6
)

// ValidationError represents a validation error with context
type ValidationError struct {
	Field   string
	Value   string
	Rule    string
	Message string
	Code    ErrorCode
}

func (e *ValidationError) Error() string {
	return e.Message
}

// NewValidationError creates a new validation error
func NewValidationError(field, value, rule, message string, code ErrorCode) *ValidationError {
	return &ValidationError{
		Field:   field,
		Value:   value,
		Rule:    rule,
		Message: message,
		Code:    code,
	}
}

// ValidateCommand validates a command string for execution
func ValidateCommand(command string) error {
	if strings.TrimSpace(command) == "" {
		return NewValidationError(
			"command",
			command,
			"required",
			"Command cannot be empty. Please provide a valid command to execute.",
			ErrorInvalidInput,
		)
	}

	if len(command) > 1000 {
		return NewValidationError(
			"command",
			command,
			"max_length",
			"Command is too long (max 1000 characters). Please shorten your command.",
			ErrorInvalidInput,
		)
	}

	// Check for potentially dangerous patterns
	dangerousPatterns := []string{
		"rm -rf",
		"sudo rm",
		"format c:",
		"del /f /s /q",
		"> /dev/null",
		"chmod -R 777",
	}

	lowerCommand := strings.ToLower(command)
	for _, pattern := range dangerousPatterns {
		if strings.Contains(lowerCommand, pattern) {
			return NewValidationError(
				"command",
				command,
				"safety",
				fmt.Sprintf("Command contains potentially dangerous pattern '%s'. Please review and modify.", pattern),
				ErrorInvalidInput,
			)
		}
	}

	return nil
}

// ValidateProjectName validates a project name
func ValidateProjectName(name string) error {
	if strings.TrimSpace(name) == "" {
		return NewValidationError(
			"project_name",
			name,
			"required",
			"Project name cannot be empty.",
			ErrorInvalidInput,
		)
	}

	// Check for invalid characters
	invalidChars := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	for _, char := range invalidChars {
		if strings.Contains(name, char) {
			return NewValidationError(
				"project_name",
				name,
				"invalid_chars",
				fmt.Sprintf("Project name contains invalid character '%s'. Use only letters, numbers, hyphens, and underscores.", char),
				ErrorInvalidInput,
			)
		}
	}

	if len(name) > 100 {
		return NewValidationError(
			"project_name",
			name,
			"max_length",
			"Project name is too long (max 100 characters).",
			ErrorInvalidInput,
		)
	}

	return nil
}

// ValidateTimeout validates timeout values
func ValidateTimeout(timeout int) error {
	if timeout < 1 {
		return NewValidationError(
			"timeout",
			fmt.Sprintf("%d", timeout),
			"min_value",
			"Timeout must be at least 1 second.",
			ErrorInvalidInput,
		)
	}

	if timeout > 3600 {
		return NewValidationError(
			"timeout",
			fmt.Sprintf("%d", timeout),
			"max_value",
			"Timeout cannot exceed 3600 seconds (1 hour).",
			ErrorInvalidInput,
		)
	}

	return nil
}

// ValidateRetries validates retry count
func ValidateRetries(retries int) error {
	if retries < 0 {
		return NewValidationError(
			"retries",
			fmt.Sprintf("%d", retries),
			"min_value",
			"Retry count cannot be negative.",
			ErrorInvalidInput,
		)
	}

	if retries > 10 {
		return NewValidationError(
			"retries",
			fmt.Sprintf("%d", retries),
			"max_value",
			"Retry count cannot exceed 10.",
			ErrorInvalidInput,
		)
	}

	return nil
}

// ValidateConfigFile validates configuration file path
func ValidateConfigFile(path string) error {
	if path == "" {
		return nil // Empty path is allowed (uses default)
	}

	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return NewValidationError(
			"config_file",
			path,
			"file_exists",
			fmt.Sprintf("Configuration file '%s' does not exist.", path),
			ErrorFileNotFound,
		)
	}

	// Check if it's readable
	if _, err := os.Open(path); err != nil {
		return NewValidationError(
			"config_file",
			path,
			"readable",
			fmt.Sprintf("Configuration file '%s' is not readable: %v", path, err),
			ErrorPermissionDenied,
		)
	}

	// Check file extension
	ext := filepath.Ext(path)
	validExts := []string{".yaml", ".yml", ".json"}
	isValid := false
	for _, validExt := range validExts {
		if strings.EqualFold(ext, validExt) {
			isValid = true
			break
		}
	}

	if !isValid {
		return NewValidationError(
			"config_file",
			path,
			"file_type",
			fmt.Sprintf("Configuration file must have .yaml, .yml, or .json extension. Got: %s", ext),
			ErrorInvalidInput,
		)
	}

	return nil
}

// HandleValidationError handles validation errors with user-friendly messages
func HandleValidationError(err error, suggestedCommand string) {
	if valErr, ok := err.(*ValidationError); ok {
		fmt.Fprintf(os.Stderr, "❌ %s\n", valErr.Message)
		
		if suggestedCommand != "" {
			fmt.Fprintf(os.Stderr, "\n💡 Try: %s\n", suggestedCommand)
		}
		
		fmt.Fprintf(os.Stderr, "\n📖 Use --help for more information.\n")
		os.Exit(int(valErr.Code))
	} else {
		fmt.Fprintf(os.Stderr, "❌ Error: %s\n", err.Error())
		os.Exit(int(ErrorGeneral))
	}
}