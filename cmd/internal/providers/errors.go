package providers

import (
	"errors"
	"fmt"
)

type ErrorType int

const (
	ErrorTypeAuthentication ErrorType = iota
	ErrorTypeValidation
	ErrorTypeNotFound
	ErrorTypeQuotaExceeded
	ErrorTypeCanceled
	ErrorTypeTimeout

	// Retryable errors
	ErrorTypeRateLimit
	ErrorTypeServerError
	ErrorTypeNetworkError
	ErrorTypeUnavailable
	ErrorTypeUnknown
)

type ProviderError struct {
	Type          ErrorType
	ProviderName  string
	StatusCode    int
	Message       string
	OriginalError error
	Retryable     bool
	RetryAfter    int
}

func (e *ProviderError) Error() string {
	return fmt.Sprintf("%s provider error [%s]: %s", e.ProviderName, e.Type.String(), e.Message)
}

func (e *ProviderError) Unwrap() error {
	return e.OriginalError
}

func (t ErrorType) String() string {
	switch t {
	case ErrorTypeAuthentication:
		return "authentication"
	case ErrorTypeValidation:
		return "validation"
	case ErrorTypeNotFound:
		return "not_found"
	case ErrorTypeQuotaExceeded:
		return "quota_exceeded"
	case ErrorTypeCanceled:
		return "canceled"
	case ErrorTypeTimeout:
		return "timeout"
	case ErrorTypeRateLimit:
		return "rate_limit"
	case ErrorTypeServerError:
		return "server_error"
	case ErrorTypeNetworkError:
		return "network_error"
	case ErrorTypeUnavailable:
		return "unavailable"
	default:
		return "unknown"
	}
}

func (e *ProviderError) IsRetryable() bool {
	return e.Retryable
}

// Helper constructors for common error types
func NewAuthenticationError(provider string, err error) *ProviderError {
	return &ProviderError{
		Type:          ErrorTypeAuthentication,
		ProviderName:  provider,
		StatusCode:    401,
		Message:       "authentication failed - check API key",
		OriginalError: err,
		Retryable:     false,
	}
}

func NewRateLimitError(provider string, statusCode int, retryAfter int, err error) *ProviderError {
	return &ProviderError{
		Type:          ErrorTypeRateLimit,
		ProviderName:  provider,
		StatusCode:    statusCode,
		Message:       "rate limit exceeded",
		OriginalError: err,
		Retryable:     true,
		RetryAfter:    retryAfter,
	}
}

func NewServerError(provider string, statusCode int, err error) *ProviderError {
	return &ProviderError{
		Type:          ErrorTypeServerError,
		ProviderName:  provider,
		StatusCode:    statusCode,
		Message:       fmt.Sprintf("server error (status %d)", statusCode),
		OriginalError: err,
		Retryable:     true,
	}
}

func NewValidationError(provider string, message string, err error) *ProviderError {
	return &ProviderError{
		Type:          ErrorTypeValidation,
		ProviderName:  provider,
		StatusCode:    400,
		Message:       message,
		OriginalError: err,
		Retryable:     false,
	}
}

func NewNetworkError(provider string, err error) *ProviderError {
	return &ProviderError{
		Type:          ErrorTypeNetworkError,
		ProviderName:  provider,
		Message:       "network error",
		OriginalError: err,
		Retryable:     true,
	}
}

func IsRetryableError(err error) bool {
	if err == nil {
		return false
	}

	var providerErr *ProviderError
	if errors.As(err, &providerErr) {
		return providerErr.IsRetryable()
	}

	return false
}
