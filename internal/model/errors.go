// Package model provides standardized error types and handling.
// This module centralizes error management to ensure consistent error reporting across the application.
package model

import (
	"fmt"
	"strings"
)

// ErrorCode represents standardized error codes similar to HTTP status codes.
// This provides consistent error categorization across the application.
type ErrorCode int

const (
	// Client errors (4xxx range)
	ErrCodeBadRequest      ErrorCode = 4000 // Invalid request format or parameters
	ErrCodeUnauthorized    ErrorCode = 4001 // Authentication required
	ErrCodeForbidden       ErrorCode = 4003 // Access denied
	ErrCodeNotFound        ErrorCode = 4004 // Requested resource not found
	ErrCodeMethodNotAllowed ErrorCode = 4005 // Operation not allowed on resource
	ErrCodeConflict        ErrorCode = 4009 // Resource conflict (e.g., concurrent modification)
	ErrCodeValidation      ErrorCode = 4022 // Input validation failed
	ErrCodeLocked          ErrorCode = 4023 // Resource is locked
	ErrCodeTooManyRequests ErrorCode = 4029 // Rate limit exceeded

	// Server errors (5xxx range)
	ErrCodeInternal         ErrorCode = 5000 // Internal server error
	ErrCodeNotImplemented   ErrorCode = 5001 // Feature not implemented
	ErrCodeBadGateway      ErrorCode = 5002 // External service error
	ErrCodeServiceUnavailable ErrorCode = 5003 // Service temporarily unavailable
	ErrCodeTimeout         ErrorCode = 5004 // Operation timeout
	ErrCodeInsufficientStorage ErrorCode = 5007 // Insufficient storage space

	// Application-specific errors (6xxx range)
	ErrCodeWorkflowViolation ErrorCode = 6001 // Workflow state transition violation
	ErrCodeDependencyMissing ErrorCode = 6002 // Required dependency not found
	ErrCodeConfigurationError ErrorCode = 6003 // Configuration error
	ErrCodeFileSystemError   ErrorCode = 6004 // File system operation failed
	ErrCodeGitError          ErrorCode = 6005 // Git operation failed
	ErrCodeGitHubError       ErrorCode = 6006 // GitHub API error
)

// ErrorSeverity indicates the severity level of an error.
type ErrorSeverity string

const (
	SeverityInfo     ErrorSeverity = "info"     // Informational - no action required
	SeverityWarning  ErrorSeverity = "warning"  // Warning - may need attention
	SeverityError    ErrorSeverity = "error"    // Error - action required
	SeverityCritical ErrorSeverity = "critical" // Critical - immediate action required
)

// CLIError represents a standardized error with rich context information.
// This provides consistent error reporting across all CLI operations.
type CLIError struct {
	Code        ErrorCode     `json:"code"`                  // Standardized error code
	Message     string        `json:"message"`               // Human-readable error message
	Context     string        `json:"context,omitempty"`     // Additional context about the error
	Details     interface{}   `json:"details,omitempty"`     // Structured error details
	Suggestions []string      `json:"suggestions,omitempty"` // Suggested actions to resolve the error
	Severity    ErrorSeverity `json:"severity"`              // Error severity level
	Cause       error         `json:"-"`                     // Underlying error cause (not serialized)
}

// Error implements the error interface.
func (e CLIError) Error() string {
	parts := []string{fmt.Sprintf("[%d]", e.Code), e.Message}
	if e.Context != "" {
		parts = append(parts, fmt.Sprintf("(%s)", e.Context))
	}
	return strings.Join(parts, " ")
}

// Unwrap returns the underlying error cause for error wrapping support.
func (e CLIError) Unwrap() error {
	return e.Cause
}

// Is implements error comparison for errors.Is support.
func (e CLIError) Is(target error) bool {
	if t, ok := target.(CLIError); ok {
		return e.Code == t.Code
	}
	return false
}

// WithContext adds context information to the error.
func (e CLIError) WithContext(context string) CLIError {
	e.Context = context
	return e
}

// WithDetails adds structured details to the error.
func (e CLIError) WithDetails(details interface{}) CLIError {
	e.Details = details
	return e
}

// WithSuggestion adds a suggestion for resolving the error.
func (e CLIError) WithSuggestion(suggestion string) CLIError {
	e.Suggestions = append(e.Suggestions, suggestion)
	return e
}

// WithSuggestions adds multiple suggestions for resolving the error.
func (e CLIError) WithSuggestions(suggestions []string) CLIError {
	e.Suggestions = append(e.Suggestions, suggestions...)
	return e
}

// WithCause wraps an underlying error as the cause.
func (e CLIError) WithCause(cause error) CLIError {
	e.Cause = cause
	return e
}

// IsClientError returns true if the error is a client-side error (4xxx range).
func (e CLIError) IsClientError() bool {
	return e.Code >= 4000 && e.Code < 5000
}

// IsServerError returns true if the error is a server-side error (5xxx range).
func (e CLIError) IsServerError() bool {
	return e.Code >= 5000 && e.Code < 6000
}

// IsApplicationError returns true if the error is application-specific (6xxx range).
func (e CLIError) IsApplicationError() bool {
	return e.Code >= 6000 && e.Code < 7000
}

// ExitCode returns the appropriate exit code for CLI applications.
func (e CLIError) ExitCode() int {
	switch {
	case e.IsClientError():
		return 2 // Client error
	case e.IsServerError():
		return 3 // Server error
	case e.IsApplicationError():
		return 4 // Application error
	default:
		return 1 // General error
	}
}

// Error constructors for common error patterns

// NewBadRequestError creates a new bad request error.
func NewBadRequestError(message string) CLIError {
	return CLIError{
		Code:     ErrCodeBadRequest,
		Message:  message,
		Severity: SeverityError,
	}
}

// NewNotFoundError creates a new not found error.
func NewNotFoundError(resource string) CLIError {
	return CLIError{
		Code:     ErrCodeNotFound,
		Message:  fmt.Sprintf("%s not found", resource),
		Severity: SeverityError,
	}.WithSuggestion("Check that the resource exists and you have access to it")
}

// NewValidationError creates a new validation error.
func NewValidationError(message string) CLIError {
	return CLIError{
		Code:     ErrCodeValidation,
		Message:  message,
		Severity: SeverityError,
	}
}

// NewConflictError creates a new conflict error.
func NewConflictError(message string) CLIError {
	return CLIError{
		Code:     ErrCodeConflict,
		Message:  message,
		Severity: SeverityError,
	}.WithSuggestion("Try refreshing and retrying the operation")
}

// NewInternalError creates a new internal error.
func NewInternalError(message string) CLIError {
	return CLIError{
		Code:     ErrCodeInternal,
		Message:  message,
		Severity: SeverityCritical,
	}.WithSuggestion("This is an internal error. Please report this issue")
}

// NewTimeoutError creates a new timeout error.
func NewTimeoutError(operation string) CLIError {
	return CLIError{
		Code:     ErrCodeTimeout,
		Message:  fmt.Sprintf("operation '%s' timed out", operation),
		Severity: SeverityError,
	}.WithSuggestions([]string{
		"Try increasing the timeout value",
		"Check network connectivity",
		"Retry the operation",
	})
}

// NewWorkflowViolationError creates a new workflow violation error.
func NewWorkflowViolationError(from, to Status) CLIError {
	return CLIError{
		Code:    ErrCodeWorkflowViolation,
		Message: fmt.Sprintf("invalid status transition from '%s' to '%s'", from, to),
		Context: "workflow state machine",
		Severity: SeverityError,
	}.WithSuggestion(fmt.Sprintf("Valid transitions from '%s' are: %s", from, getValidTransitions(from)))
}

// NewFileSystemError creates a new file system error.
func NewFileSystemError(operation, path string, cause error) CLIError {
	return CLIError{
		Code:     ErrCodeFileSystemError,
		Message:  fmt.Sprintf("file system operation '%s' failed on path '%s'", operation, path),
		Severity: SeverityError,
		Cause:    cause,
	}.WithSuggestions([]string{
		"Check file permissions",
		"Ensure the directory exists",
		"Check available disk space",
	})
}

// NewGitError creates a new Git operation error.
func NewGitError(operation string, cause error) CLIError {
	return CLIError{
		Code:     ErrCodeGitError,
		Message:  fmt.Sprintf("git operation '%s' failed", operation),
		Severity: SeverityError,
		Cause:    cause,
	}.WithSuggestions([]string{
		"Check git repository status",
		"Ensure you have proper git credentials",
		"Try running git commands manually to diagnose",
	})
}

// NewGitHubError creates a new GitHub API error.
func NewGitHubError(operation string, statusCode int, cause error) CLIError {
	err := CLIError{
		Code:     ErrCodeGitHubError,
		Message:  fmt.Sprintf("GitHub API operation '%s' failed", operation),
		Context:  fmt.Sprintf("HTTP %d", statusCode),
		Severity: SeverityError,
		Cause:    cause,
	}

	// Add specific suggestions based on status code
	switch statusCode {
	case 401:
		err = err.WithSuggestion("Check your GitHub authentication token")
	case 403:
		err = err.WithSuggestions([]string{
			"Check repository permissions",
			"Verify API rate limits",
		})
	case 404:
		err = err.WithSuggestion("Check that the repository exists and you have access")
	case 422:
		err = err.WithSuggestion("Check the request parameters and format")
	default:
		err = err.WithSuggestion("Check GitHub service status and try again")
	}

	return err
}

// Helper function to get valid transitions for workflow errors
func getValidTransitions(from Status) string {
	transitions := map[Status][]Status{
		StatusPlanned:    {StatusInProgress, StatusOnHold, StatusCancelled},
		StatusInProgress: {StatusBlocked, StatusOnHold, StatusCompleted, StatusCancelled},
		StatusBlocked:    {StatusInProgress, StatusOnHold, StatusCancelled},
		StatusOnHold:     {StatusPlanned, StatusInProgress, StatusCancelled},
		StatusCompleted:  {},
		StatusCancelled:  {StatusPlanned},
	}

	valid, exists := transitions[from]
	if !exists || len(valid) == 0 {
		return "none"
	}

	var strs []string
	for _, status := range valid {
		strs = append(strs, string(status))
	}
	return strings.Join(strs, ", ")
}

// ErrorHandler defines the interface for handling errors consistently across the application.
type ErrorHandler interface {
	HandleError(error) error
	LogError(error)
	FormatError(error) string
}

// ExitCodes defines standard exit codes for CLI applications.
var ExitCodes = struct {
	Success        int
	GeneralError   int
	ClientError    int
	ServerError    int
	ApplicationError int
}{
	Success:         0,
	GeneralError:    1,
	ClientError:     2,
	ServerError:     3,
	ApplicationError: 4,
}