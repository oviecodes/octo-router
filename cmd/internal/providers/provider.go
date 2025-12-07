package providers

import (
	"context"
	"llm-router/cmd/internal/types"
)

type Provider interface {
	Complete(ctx context.Context, messages []types.Message) (*types.Message, error)
	CountTokens(ctx context.Context, messages []types.Message) (int, error)
}
