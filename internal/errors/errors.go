package errors

import (
	"fmt"
	"net/http"
)

// ErrorCode represents standardized error codes
type ErrorCode string

const (
	// Authentication errors
	ErrUnauthorized          ErrorCode = "UNAUTHORIZED"
	ErrInvalidToken         ErrorCode = "INVALID_TOKEN"
	ErrTokenExpired         ErrorCode = "TOKEN_EXPIRED"
	ErrMissingAuth          ErrorCode = "MISSING_AUTH"

	// Validation errors
	ErrInvalidRequest       ErrorCode = "INVALID_REQUEST"
	ErrValidationFailed     ErrorCode = "VALIDATION_FAILED"
	ErrMissingRequired      ErrorCode = "MISSING_REQUIRED"
	ErrInvalidFormat        ErrorCode = "INVALID_FORMAT"

	// Resource errors
	ErrNotFound             ErrorCode = "NOT_FOUND"
	ErrAlreadyExists        ErrorCode = "ALREADY_EXISTS"
	ErrResourceExhausted    ErrorCode = "RESOURCE_EXHAUSTED"

	// Permission errors
	ErrForbidden           ErrorCode = "FORBIDDEN"
	ErrPermissionDenied    ErrorCode = "PERMISSION_DENIED"

	// Business logic errors
	ErrInsufficientFunds    ErrorCode = "INSUFFICIENT_FUNDS"
	ErrInvalidState         ErrorCode = "INVALID_STATE"
	ErrBusinessRule         ErrorCode = "BUSINESS_RULE"

	// System errors
	ErrInternal             ErrorCode = "INTERNAL"
	ErrDatabase             ErrorCode = "DATABASE"
	ErrNetwork              ErrorCode = "NETWORK"
	ErrTimeout              ErrorCode = "TIMEOUT"
)

// APIError represents a structured API error
type APIError struct {
	Code       ErrorCode              `json:"code"`
	Message    string                 `json:"message"`
	Details    map[string]interface{} `json:"details,omitempty"`
	HTTPStatus int                    `json:"-"`
}

// Error implements the error interface
func (e *APIError) Error() string {
	if e.Details != nil {
		return fmt.Sprintf("%s: %s (details: %+v)", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// NewAPIError creates a new API error
func NewAPIError(code ErrorCode, message string) *APIError {
	return &APIError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatusForCode(code),
	}
}

// NewAPIErrorWithDetails creates a new API error with details
func NewAPIErrorWithDetails(code ErrorCode, message string, details map[string]interface{}) *APIError {
	return &APIError{
		Code:       code,
		Message:    message,
		Details:    details,
		HTTPStatus: httpStatusForCode(code),
	}
}

// WrapError wraps an existing error with API error context
func WrapError(err error, code ErrorCode, message string) *APIError {
	return &APIError{
		Code:       code,
		Message:    message,
		Details:    map[string]interface{}{"original_error": err.Error()},
		HTTPStatus: httpStatusForCode(code),
	}
}

// httpStatusForCode returns appropriate HTTP status for error code
func httpStatusForCode(code ErrorCode) int {
	switch code {
	case ErrUnauthorized, ErrInvalidToken, ErrTokenExpired, ErrMissingAuth:
		return http.StatusUnauthorized
	case ErrForbidden, ErrPermissionDenied:
		return http.StatusForbidden
	case ErrNotFound:
		return http.StatusNotFound
	case ErrAlreadyExists:
		return http.StatusConflict
	case ErrInvalidRequest, ErrValidationFailed, ErrMissingRequired, ErrInvalidFormat:
		return http.StatusBadRequest
	case ErrResourceExhausted:
		return http.StatusTooManyRequests
	case ErrInsufficientFunds, ErrInvalidState, ErrBusinessRule:
		return http.StatusUnprocessableEntity
	default:
		return http.StatusInternalServerError
	}
}

// Common error constructors
func NewUnauthorizedError(message string) *APIError {
	return NewAPIError(ErrUnauthorized, message)
}

func NewForbiddenError(message string) *APIError {
	return NewAPIError(ErrForbidden, message)
}

func NewNotFoundError(resource string) *APIError {
	return NewAPIError(ErrNotFound, fmt.Sprintf("%s not found", resource))
}

func NewValidationError(message string, details map[string]interface{}) *APIError {
	return NewAPIErrorWithDetails(ErrValidationFailed, message, details)
}

func NewInternalError(message string) *APIError {
	return NewAPIError(ErrInternal, message)
}

func NewDatabaseError(err error) *APIError {
	return WrapError(err, ErrDatabase, "Database operation failed")
}

func NewNetworkError(err error) *APIError {
	return WrapError(err, ErrNetwork, "Network operation failed")
}

func NewTimeoutError(operation string) *APIError {
	return NewAPIError(ErrTimeout, fmt.Sprintf("%s timed out", operation))
}

// IsErrorCode checks if error matches specific error code
func IsErrorCode(err error, code ErrorCode) bool {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.Code == code
	}
	return false
}

// GetErrorCode extracts error code from error
func GetErrorCode(err error) ErrorCode {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.Code
	}
	return ErrInternal
}