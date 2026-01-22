package filters

import (
	"context"
	"fmt"
	"llm-router/types"

	"go.uber.org/zap"
)

type RateLimitManager interface {
	Allow(ctx context.Context, key string, limit int) (bool, error)
}

type RateLimitFilter struct {
	manager RateLimitManager
	limits  map[string]int // provider name -> limit (RPM)
	logger  *zap.Logger
}

func NewRateLimitFilter(manager RateLimitManager, limits map[string]int, logger *zap.Logger) *RateLimitFilter {
	return &RateLimitFilter{
		manager: manager,
		limits:  limits,
		logger:  logger,
	}
}

func (f *RateLimitFilter) Name() string {
	return "ratelimit"
}

func (f *RateLimitFilter) Filter(ctx context.Context, input *types.FilterInput) (*types.FilterOutput, error) {
	var filtered []types.Provider

	for _, p := range input.Candidates {
		name := p.GetProviderName()
		limit, exists := f.limits[name]

		if !exists || limit <= 0 {
			filtered = append(filtered, p)
			continue
		}

		key := fmt.Sprintf("provider:%s", name)
		allowed, err := f.manager.Allow(ctx, key, limit)
		if err != nil {
			f.logger.Error("Rate limit check failed, allowing anyway", zap.Error(err), zap.String("provider", name))
			filtered = append(filtered, p)
			continue
		}

		if allowed {
			filtered = append(filtered, p)
		} else {
			f.logger.Warn("Provider rate limit reached, skipping", zap.String("provider", name))
		}
	}

	return &types.FilterOutput{
		Candidates: filtered,
	}, nil
}
