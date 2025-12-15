package resilience

import (
	"context"
	"errors"
	"fmt"
	"llm-router/cmd/internal/providers"
	"math"
	"time"

	"go.uber.org/zap"
)

// type RetrySetup interface {
// 	Retry(handler func() (*types.Message, error))
// }

type Retry struct {
	config config
	logger *zap.Logger
}

type config struct {
	maxAttempts       int
	initialDelay      time.Duration
	maxDelay          time.Duration
	backoffMultiplier int
}

func Do[T any](ctx context.Context, r *Retry, handler func(context.Context) (T, error)) (T, error) {

	var result T
	var lastErr error

	for attempt := 0; attempt < r.config.maxAttempts; attempt++ {

		select {
		case <-ctx.Done():
			return result, fmt.Errorf("retry cancelled: %w", ctx.Err())
		default:
		}

		res, err := handler(ctx)

		if err == nil {
			return res, nil
		}

		lastErr = err

		// Check if error is retryable using our domain error system
		if !providers.IsRetryableError(err) {
			// Log additional details if it's a ProviderError
			var providerErr *providers.ProviderError
			if errors.As(err, &providerErr) {
				r.logger.Debug("Error not retryable, failing immediately",
					zap.String("error_type", providerErr.Type.String()),
					zap.String("provider", providerErr.ProviderName),
					zap.Int("status_code", providerErr.StatusCode),
					zap.Int("attempt", attempt+1),
					zap.Error(err),
				)
			} else {
				r.logger.Debug("Error not retryable, failing immediately",
					zap.Error(err),
					zap.Int("attempt", attempt+1),
				)
			}
			return result, fmt.Errorf("non-retryable error: %w", err)
		}

		if attempt < r.config.maxAttempts-1 {
			delay := r.calculateBackoff(attempt)

			// Log retry with provider error details if available
			var providerErr *providers.ProviderError
			if errors.As(err, &providerErr) {
				r.logger.Debug("Retrying after error",
					zap.String("error_type", providerErr.Type.String()),
					zap.String("provider", providerErr.ProviderName),
					zap.Int("status_code", providerErr.StatusCode),
					zap.Int("attempt", attempt+1),
					zap.Int("maxAttempts", r.config.maxAttempts),
					zap.Duration("delay", delay),
					zap.Error(err),
				)
			} else {
				r.logger.Debug("Retrying after error",
					zap.Error(err),
					zap.Int("attempt", attempt+1),
					zap.Int("maxAttempts", r.config.maxAttempts),
					zap.Duration("delay", delay),
				)
			}

			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return result, fmt.Errorf("retry cancelled during backoff: %w", ctx.Err())
			}
		}
	}

	return result, fmt.Errorf("max retry attempts (%d) exceeded: %w", r.config.maxAttempts, lastErr)
}

func (r *Retry) calculateBackoff(attempt int) time.Duration {
	backoff := float64(r.config.initialDelay) * math.Pow(float64(r.config.backoffMultiplier), float64(attempt))
	delay := min(time.Duration(backoff), time.Duration(r.config.maxDelay))

	return delay
}

func NewRetryHandler(configs map[string]int, logger *zap.Logger) *Retry {

	configStruct := config{
		maxAttempts:       getOrDefault(configs, "maxAttempts", 3),
		initialDelay:      time.Duration(getOrDefault(configs, "initialDelay", 1000)) * time.Millisecond,
		maxDelay:          time.Duration(getOrDefault(configs, "maxDelay", 10000)) * time.Millisecond,
		backoffMultiplier: getOrDefault(configs, "backoffMultiplier", 2),
	}

	return &Retry{
		config: configStruct,
		logger: logger,
	}
}

func getOrDefault(m map[string]int, key string, defaultValue int) int {
	if val, ok := m[key]; ok {
		return val
	}
	return defaultValue
}
