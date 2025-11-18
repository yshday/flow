package errors

import (
	"errors"
	"fmt"
)

// Sentinel errors
var (
	ErrNotFound           = errors.New("not found")
	ErrConflict           = errors.New("conflict")
	ErrForbidden          = errors.New("forbidden")
	ErrValidation         = errors.New("validation error")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUnauthorized       = errors.New("unauthorized")
)

// AppError represents an application-specific error with HTTP status code
type AppError struct {
	Message    string
	StatusCode int
	Err        error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// NewValidationError creates a 400 Bad Request error
func NewValidationError(message string) *AppError {
	return &AppError{
		Message:    message,
		StatusCode: 400,
	}
}

// NewNotFoundError creates a 404 Not Found error
func NewNotFoundError(message string) *AppError {
	return &AppError{
		Message:    message,
		StatusCode: 404,
	}
}

// NewPermissionError creates a 403 Forbidden error
func NewPermissionError(message string) *AppError {
	return &AppError{
		Message:    message,
		StatusCode: 403,
	}
}

// NewFileTooLargeError creates a 413 Request Entity Too Large error
func NewFileTooLargeError(message string) *AppError {
	return &AppError{
		Message:    message,
		StatusCode: 413,
	}
}

// NewInternalError creates a 500 Internal Server Error
func NewInternalError(message string, err error) *AppError {
	return &AppError{
		Message:    message,
		StatusCode: 500,
		Err:        err,
	}
}
