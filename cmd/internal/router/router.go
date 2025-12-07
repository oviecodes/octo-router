package router

import (
	"context"
	"fmt"
	"llm-router/cmd/internal/providers"
	"sync"
)

type RouterConfig struct {
	AnthropicAPIKey string
	OpenAIAPIKey    string
	MaxTokens       int
	model           string
}

type RoundRobinRouter struct {
	providers []providers.Provider
	mu        sync.Mutex
	current   int
}

func (r *RoundRobinRouter) SelectProvider(ctx context.Context) providers.Provider {
	r.mu.Lock()
	defer r.mu.Unlock()

	provider := r.providers[r.current]
	r.current = (r.current + 1) % len(r.providers)

	return provider
}

func NewRoundRobinRouter(config RouterConfig) (*RoundRobinRouter, error) {
	anthropicProvider, err := providers.NewAnthropicProvider(providers.AnthropicConfig{
		APIKey:    config.AnthropicAPIKey,
		MaxTokens: config.MaxTokens,
		Model:     config.model,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create Anthropic provider: %w", err)
	}

	openaiProvider, err := providers.NewOpenAIProvider(providers.OpenAIConfig{
		APIKey:    config.OpenAIAPIKey,
		MaxTokens: config.MaxTokens,
		Model:     config.model,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create OpenAI provider: %w", err)
	}

	return &RoundRobinRouter{
		current: 0,
		providers: []providers.Provider{
			anthropicProvider,
			openaiProvider,
		},
		mu: sync.Mutex{},
	}, nil
}
