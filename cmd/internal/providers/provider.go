package providers

import (
	"context"
	"llm-router/cmd/internal/router"
)

type Provider interface {
	Complete(ctx context.Context, messages []router.Message) (*router.Message, error)
	CountTokens(ctx context.Context, messages []router.Message) (int, error)
}
