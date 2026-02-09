// Package errors defines error types and exit codes for the application.
package errors

import "fmt"

// Exit codes
const (
	ExitSuccess     = 0
	ExitRuntime     = 1
	ExitConfig      = 2
	ExitEnvironment = 3
	ExitUserAbort   = 4
)

// AppError represents an application error with an exit code.
type AppError struct {
	Message  string
	ExitCode int
	Cause    error
}

func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}

	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Cause
}

// Runtime creates a runtime error (exit code 1).
func Runtime(msg string, cause error) *AppError {
	return &AppError{Message: msg, ExitCode: ExitRuntime, Cause: cause}
}

// Config creates a configuration error (exit code 2).
func Config(msg string) *AppError {
	return &AppError{Message: msg, ExitCode: ExitConfig}
}

// Environment creates an environment error (exit code 3).
func Environment(msg string, cause error) *AppError {
	return &AppError{Message: msg, ExitCode: ExitEnvironment, Cause: cause}
}

// UserAbort creates a user abort error (exit code 4).
func UserAbort() *AppError {
	return &AppError{Message: "operation cancelled by user", ExitCode: ExitUserAbort}
}
