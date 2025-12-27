package client

// Exit codes as specified in TEMPLATE.md PART 33
const (
	ExitSuccess      = 0  // Success
	ExitGeneralError = 1  // General error
	ExitConfigError  = 2  // Config error
	ExitConnError    = 3  // Connection error
	ExitAuthError    = 4  // Authentication error
	ExitNotFound     = 5  // Not found
	ExitUsageError   = 64 // Usage error
)

// ExitError represents an error with a specific exit code
type ExitError struct {
	Message string
	Code    int
}

// Error implements the error interface
func (e *ExitError) Error() string {
	return e.Message
}

// NewExitError creates a new ExitError
func NewExitError(message string, code int) *ExitError {
	return &ExitError{Message: message, Code: code}
}

// NewConfigError creates a config error (exit code 2)
func NewConfigError(message string) *ExitError {
	return &ExitError{Message: message, Code: ExitConfigError}
}

// NewConnectionError creates a connection error (exit code 3)
func NewConnectionError(message string) *ExitError {
	return &ExitError{Message: message, Code: ExitConnError}
}

// NewAuthError creates an auth error (exit code 4)
func NewAuthError(message string) *ExitError {
	return &ExitError{Message: message, Code: ExitAuthError}
}

// NewNotFoundError creates a not found error (exit code 5)
func NewNotFoundError(message string) *ExitError {
	return &ExitError{Message: message, Code: ExitNotFound}
}

// NewUsageError creates a usage error (exit code 64)
func NewUsageError(message string) *ExitError {
	return &ExitError{Message: message, Code: ExitUsageError}
}
