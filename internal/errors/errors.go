package errors

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"
)

// CLIError represents an error with additional context
type CLIError struct {
	Message    string
	Code       int
	Suggestion string
	Details    string
	Timestamp  time.Time
	Context    map[string]interface{}
}

func (e *CLIError) Error() string {
	return e.Message
}

// NewCLIError creates a new CLI error
func NewCLIError(message string, code int) *CLIError {
	return &CLIError{
		Message:   message,
		Code:      code,
		Timestamp: time.Now(),
		Context:   make(map[string]interface{}),
	}
}

// WithSuggestion adds a suggestion to the error
func (e *CLIError) WithSuggestion(suggestion string) *CLIError {
	e.Suggestion = suggestion
	return e
}

// WithDetails adds details to the error
func (e *CLIError) WithDetails(details string) *CLIError {
	e.Details = details
	return e
}

// WithContext adds context information
func (e *CLIError) WithContext(key string, value interface{}) *CLIError {
	e.Context[key] = value
	return e
}

// HandleError handles errors with appropriate user feedback
func HandleError(err error, verbose bool) {
	if err == nil {
		return
	}

	if cliErr, ok := err.(*CLIError); ok {
		handleCLIError(cliErr, verbose)
	} else {
		handleGenericError(err, verbose)
	}
}

func handleCLIError(err *CLIError, verbose bool) {
	// Print main error message
	fmt.Fprintf(os.Stderr, "❌ %s\n", err.Message)

	// Print suggestion if available
	if err.Suggestion != "" {
		fmt.Fprintf(os.Stderr, "\n💡 %s\n", err.Suggestion)
	}

	// Print details if available and verbose mode is on
	if verbose && err.Details != "" {
		fmt.Fprintf(os.Stderr, "\n📋 Details:\n%s\n", err.Details)
	}

	// Print context in verbose mode
	if verbose && len(err.Context) > 0 {
		fmt.Fprintf(os.Stderr, "\n🔍 Context:\n")
		for key, value := range err.Context {
			fmt.Fprintf(os.Stderr, "  %s: %v\n", key, value)
		}
	}

	// Print timestamp in verbose mode
	if verbose {
		fmt.Fprintf(os.Stderr, "\n⏰ Time: %s\n", err.Timestamp.Format(time.RFC3339))
	}

	fmt.Fprintf(os.Stderr, "\n📖 Use --help for more information.\n")
	os.Exit(err.Code)
}

func handleGenericError(err error, verbose bool) {
	fmt.Fprintf(os.Stderr, "❌ Error: %s\n", err.Error())

	if verbose {
		// Print stack trace in verbose mode
		buf := make([]byte, 4096)
		n := runtime.Stack(buf, false)
		fmt.Fprintf(os.Stderr, "\n🔍 Stack trace:\n%s\n", buf[:n])
	}

	fmt.Fprintf(os.Stderr, "\n📖 Use --help for more information.\n")
	os.Exit(1)
}

// Common error constructors

// ErrInvalidInput creates an invalid input error
func ErrInvalidInput(field, value, message string) *CLIError {
	return NewCLIError(
		fmt.Sprintf("Invalid %s: %s", field, message),
		2,
	).WithContext("field", field).WithContext("value", value)
}

// ErrFileNotFound creates a file not found error
func ErrFileNotFound(path string) *CLIError {
	return NewCLIError(
		fmt.Sprintf("File not found: %s", path),
		3,
	).WithSuggestion("Check that the file path is correct and the file exists").
		WithContext("path", path)
}

// ErrPermissionDenied creates a permission denied error
func ErrPermissionDenied(path string) *CLIError {
	return NewCLIError(
		fmt.Sprintf("Permission denied: %s", path),
		4,
	).WithSuggestion("Check file permissions or run with appropriate privileges").
		WithContext("path", path)
}

// ErrTimeout creates a timeout error
func ErrTimeout(operation string, duration time.Duration) *CLIError {
	return NewCLIError(
		fmt.Sprintf("Operation timed out: %s", operation),
		5,
	).WithSuggestion(fmt.Sprintf("Try increasing the timeout (current: %v) or check your network connection", duration)).
		WithContext("operation", operation).
		WithContext("timeout", duration.String())
}

// ErrNetworkFailure creates a network failure error
func ErrNetworkFailure(operation string, cause error) *CLIError {
	return NewCLIError(
		fmt.Sprintf("Network failure during %s", operation),
		6,
	).WithSuggestion("Check your internet connection and try again").
		WithDetails(cause.Error()).
		WithContext("operation", operation)
}

// ErrCommandFailed creates a command execution failure error
func ErrCommandFailed(command string, exitCode int, stderr string) *CLIError {
	err := NewCLIError(
		fmt.Sprintf("Command failed with exit code %d", exitCode),
		1,
	).WithContext("command", command).
		WithContext("exit_code", exitCode)

	if stderr != "" {
		err = err.WithDetails(fmt.Sprintf("Command output:\n%s", stderr))
	}

	// Add suggestions based on common error patterns
	lowerStderr := strings.ToLower(stderr)
	if strings.Contains(lowerStderr, "permission denied") {
		err = err.WithSuggestion("Check file permissions or run with appropriate privileges")
	} else if strings.Contains(lowerStderr, "command not found") || strings.Contains(lowerStderr, "not recognized") {
		err = err.WithSuggestion("Make sure the command is installed and available in your PATH")
	} else if strings.Contains(lowerStderr, "network") || strings.Contains(lowerStderr, "connection") {
		err = err.WithSuggestion("Check your network connection and try again")
	} else {
		err = err.WithSuggestion("Review the command and its arguments, then try again")
	}

	return err
}

// PrintWarning prints a warning message
func PrintWarning(message string) {
	fmt.Fprintf(os.Stderr, "⚠️  Warning: %s\n", message)
}

// PrintInfo prints an informational message
func PrintInfo(message string) {
	fmt.Fprintf(os.Stdout, "ℹ️  %s\n", message)
}

// PrintSuccess prints a success message
func PrintSuccess(message string) {
	fmt.Fprintf(os.Stdout, "✅ %s\n", message)
}
