package providers

import (
	"context"
	"errors"
	"net"
	"net/url"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/openai/openai-go/v3"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// TranslateOpenAIError converts OpenAI SDK errors to ProviderError
func TranslateOpenAIError(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, context.Canceled) {
		return &ProviderError{
			Type:          ErrorTypeCanceled,
			ProviderName:  "openai",
			Message:       "request canceled",
			OriginalError: err,
			Retryable:     false,
		}
	}

	if errors.Is(err, context.DeadlineExceeded) {
		return &ProviderError{
			Type:          ErrorTypeTimeout,
			ProviderName:  "openai",
			Message:       "request timeout",
			OriginalError: err,
			Retryable:     false,
		}
	}

	var apiErr *openai.Error
	if errors.As(err, &apiErr) {
		statusCode := apiErr.StatusCode

		switch statusCode {
		case 401:
			return NewAuthenticationError("openai", err)

		case 403:
			return &ProviderError{
				Type:          ErrorTypeAuthentication,
				ProviderName:  "openai",
				StatusCode:    403,
				Message:       "forbidden - check API key permissions or model access",
				OriginalError: err,
				Retryable:     false,
			}

		case 404:
			return &ProviderError{
				Type:          ErrorTypeNotFound,
				ProviderName:  "openai",
				StatusCode:    404,
				Message:       "model or resource not found",
				OriginalError: err,
				Retryable:     false,
			}

		case 429:
			return NewRateLimitError("openai", 429, 0, err)

		case 400:
			return NewValidationError("openai", "invalid request parameters", err)

		case 413:
			return &ProviderError{
				Type:          ErrorTypeValidation,
				ProviderName:  "openai",
				StatusCode:    413,
				Message:       "request too large - reduce message size or tokens",
				OriginalError: err,
				Retryable:     false,
			}

		case 422:
			return NewValidationError("openai", "unprocessable entity - validation failed", err)

		case 500, 502, 503, 504:
			return NewServerError("openai", statusCode, err)

		default:
			// Unknown status code - retry if 5xx
			return &ProviderError{
				Type:          ErrorTypeUnknown,
				ProviderName:  "openai",
				StatusCode:    statusCode,
				Message:       "unknown error",
				OriginalError: err,
				Retryable:     statusCode >= 500,
			}
		}
	}

	// Check for network errors
	if isNetworkError(err) {
		return NewNetworkError("openai", err)
	}

	// Unknown error type - don't retry for safety
	return &ProviderError{
		Type:          ErrorTypeUnknown,
		ProviderName:  "openai",
		Message:       err.Error(),
		OriginalError: err,
		Retryable:     false,
	}
}

// TranslateAnthropicError converts Anthropic SDK errors to ProviderError
func TranslateAnthropicError(err error) error {
	if err == nil {
		return nil
	}

	// Check for context errors first
	if errors.Is(err, context.Canceled) {
		return &ProviderError{
			Type:          ErrorTypeCanceled,
			ProviderName:  "anthropic",
			Message:       "request canceled",
			OriginalError: err,
			Retryable:     false,
		}
	}

	if errors.Is(err, context.DeadlineExceeded) {
		return &ProviderError{
			Type:          ErrorTypeTimeout,
			ProviderName:  "anthropic",
			Message:       "request timeout",
			OriginalError: err,
			Retryable:     false,
		}
	}

	var apiErr *anthropic.Error
	if errors.As(err, &apiErr) {
		statusCode := apiErr.StatusCode

		switch statusCode {
		case 401:
			return NewAuthenticationError("anthropic", err)

		case 403:
			return &ProviderError{
				Type:          ErrorTypeAuthentication,
				ProviderName:  "anthropic",
				StatusCode:    403,
				Message:       "forbidden - check API key permissions or model access",
				OriginalError: err,
				Retryable:     false,
			}

		case 404:
			return &ProviderError{
				Type:          ErrorTypeNotFound,
				ProviderName:  "anthropic",
				StatusCode:    404,
				Message:       "resource not found",
				OriginalError: err,
				Retryable:     false,
			}

		case 429:
			return NewRateLimitError("anthropic", 429, 0, err)

		case 400:
			return NewValidationError("anthropic", "invalid request parameters", err)

		case 413:
			return &ProviderError{
				Type:          ErrorTypeValidation,
				ProviderName:  "anthropic",
				StatusCode:    413,
				Message:       "request too large - reduce message size or tokens",
				OriginalError: err,
				Retryable:     false,
			}

		case 422:
			return NewValidationError("anthropic", "unprocessable entity - validation failed", err)

		case 529:
			// Anthropic-specific: overloaded error (retryable)
			return &ProviderError{
				Type:          ErrorTypeUnavailable,
				ProviderName:  "anthropic",
				StatusCode:    529,
				Message:       "service overloaded - temporarily unavailable",
				OriginalError: err,
				Retryable:     true,
			}

		case 500, 502, 503, 504:
			return NewServerError("anthropic", statusCode, err)

		default:
			return &ProviderError{
				Type:          ErrorTypeUnknown,
				ProviderName:  "anthropic",
				StatusCode:    statusCode,
				Message:       "unknown error",
				OriginalError: err,
				Retryable:     statusCode >= 500,
			}
		}
	}

	// Check for network errors
	if isNetworkError(err) {
		return NewNetworkError("anthropic", err)
	}

	return &ProviderError{
		Type:          ErrorTypeUnknown,
		ProviderName:  "anthropic",
		Message:       err.Error(),
		OriginalError: err,
		Retryable:     false,
	}
}

// TranslateGeminiError converts Gemini gRPC errors to ProviderError
func TranslateGeminiError(err error) error {
	if err == nil {
		return nil
	}

	// Check for context errors first
	if errors.Is(err, context.Canceled) {
		return &ProviderError{
			Type:          ErrorTypeCanceled,
			ProviderName:  "gemini",
			Message:       "request canceled",
			OriginalError: err,
			Retryable:     false,
		}
	}

	if errors.Is(err, context.DeadlineExceeded) {
		return &ProviderError{
			Type:          ErrorTypeTimeout,
			ProviderName:  "gemini",
			Message:       "request timeout",
			OriginalError: err,
			Retryable:     false,
		}
	}

	// Check for gRPC status errors (Gemini uses gRPC)
	if st, ok := status.FromError(err); ok {
		switch st.Code() {
		case codes.Unauthenticated:
			return NewAuthenticationError("gemini", err)

		case codes.PermissionDenied:
			return &ProviderError{
				Type:          ErrorTypeAuthentication,
				ProviderName:  "gemini",
				StatusCode:    403,
				Message:       "permission denied - check API key or model access",
				OriginalError: err,
				Retryable:     false,
			}

		case codes.InvalidArgument:
			return NewValidationError("gemini", st.Message(), err)

		case codes.NotFound:
			return &ProviderError{
				Type:          ErrorTypeNotFound,
				ProviderName:  "gemini",
				StatusCode:    404,
				Message:       "resource not found - check model name",
				OriginalError: err,
				Retryable:     false,
			}

		case codes.ResourceExhausted:
			// gRPC equivalent of rate limiting
			return NewRateLimitError("gemini", 429, 0, err)

		case codes.DeadlineExceeded:
			return &ProviderError{
				Type:          ErrorTypeTimeout,
				ProviderName:  "gemini",
				StatusCode:    504,
				Message:       "deadline exceeded",
				OriginalError: err,
				Retryable:     false,
			}

		case codes.Unavailable:
			return &ProviderError{
				Type:          ErrorTypeUnavailable,
				ProviderName:  "gemini",
				StatusCode:    503,
				Message:       "service unavailable",
				OriginalError: err,
				Retryable:     true,
			}

		case codes.Internal:
			return &ProviderError{
				Type:          ErrorTypeServerError,
				ProviderName:  "gemini",
				StatusCode:    500,
				Message:       "internal server error",
				OriginalError: err,
				Retryable:     true,
			}

		case codes.Unknown:
			return &ProviderError{
				Type:          ErrorTypeServerError,
				ProviderName:  "gemini",
				StatusCode:    500,
				Message:       "unknown server error",
				OriginalError: err,
				Retryable:     true,
			}

		case codes.Canceled:
			return &ProviderError{
				Type:          ErrorTypeCanceled,
				ProviderName:  "gemini",
				Message:       "request canceled",
				OriginalError: err,
				Retryable:     false,
			}

		default:
			return &ProviderError{
				Type:          ErrorTypeUnknown,
				ProviderName:  "gemini",
				Message:       st.Message(),
				OriginalError: err,
				Retryable:     false,
			}
		}
	}

	// Check for network errors
	if isNetworkError(err) {
		return NewNetworkError("gemini", err)
	}

	return &ProviderError{
		Type:          ErrorTypeUnknown,
		ProviderName:  "gemini",
		Message:       err.Error(),
		OriginalError: err,
		Retryable:     false,
	}
}

// isNetworkError checks if error is a network-related error
func isNetworkError(err error) bool {
	// Check for url.Error (which wraps network errors from HTTP clients)
	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		return true
	}

	// Check for net.Error (DNS, connection failures, etc.)
	var netErr net.Error
	return errors.As(err, &netErr)
}
