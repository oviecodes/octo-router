package providers

import (
	"context"
	"llm-router/types"
	"log"
)

type Provider interface {
	Complete(ctx context.Context, messages []types.Message) (*types.Message, error)
	CountTokens(ctx context.Context, messages []types.Message) (int, error)
}

// var configMap = map[string] Provider{
// 	"openai": &OpenAIProvider{},
// }

func ConfigureProviders(configs []types.ProviderConfig, extra types.ProviderExtra) []Provider {

	var providers []Provider

	for _, config := range configs {
		switch config.Name {
		case "openai":
			provider, err := NewOpenAIProvider(OpenAIConfig{
				APIKey:    config.APIKey,
				MaxTokens: extra.MaxTokens,
				Model:     extra.Model,
			})

			if err != nil {
				log.Printf("cannot configure %v provider", config.Name)
				continue
			}

			providers = append(providers, provider)

		case "anthropic":
			provider, err := NewAnthropicProvider(AnthropicConfig{
				APIKey:    config.APIKey,
				MaxTokens: extra.MaxTokens,
				Model:     extra.Model,
			})

			if err != nil {
				log.Printf("cannot configure %v provider", config.Name)
				continue
			}

			providers = append(providers, provider)

		default:
			log.Printf("Warning: unknown provider %s, skipping", config.Name)
		}

	}
	return providers
}
