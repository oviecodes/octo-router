package filters

import (
	"context"
	"llm-router/types"

	"go.uber.org/zap"
)

type BudgetManager interface {
	IsWithinBudget(provider string) bool
}

type BudgetFilter struct {
	manager BudgetManager
	logger  *zap.Logger
}

func NewBudgetFilter(manager BudgetManager, logger *zap.Logger) *BudgetFilter {
	return &BudgetFilter{
		manager: manager,
		logger:  logger,
	}
}

func (f *BudgetFilter) Name() string {
	return "budget"
}

func (f *BudgetFilter) Filter(ctx context.Context, input *types.FilterInput) (*types.FilterOutput, error) {
	var filtered []types.Provider

	for _, p := range input.Candidates {
		name := p.GetProviderName()
		if f.manager.IsWithinBudget(name) {
			filtered = append(filtered, p)
		} else {
			f.logger.Warn("Budget limit reached, skipping provider",
				zap.String("provider", name),
			)
		}
	}

	return &types.FilterOutput{
		Candidates: filtered,
	}, nil
}

// update: use redis for cost tracking, then we could also implement rate-limiting for both budgets and requests per-provider
