package providers

import (
	"context"
	"llm-router/cmd/internal/types"
)

type ProviderConfig struct {
	// openAPIKey
}

type Provider interface {
	Complete(ctx context.Context, messages []types.Message) (*types.Message, error)
	CountTokens(ctx context.Context, messages []types.Message) (int, error)
}

func configureProviders(config ProviderConfig) []Provider {
	return []Provider{}
}
