package providers

import (
	"context"
	"llm-router/types"
	"llm-router/utils"

	"go.uber.org/zap"
)

type Provider interface {
	Complete(ctx context.Context, messages []types.Message) (*types.Message, error)
	CountTokens(ctx context.Context, messages []types.Message) (int, error)
}

var logger = utils.SetUpLogger()

func ConfigureProviders(configs []types.ProviderConfigWithExtras) []Provider {

	var providers []Provider

	for _, config := range configs {
		switch config.Name {
		case "openai":
			provider, err := NewOpenAIProvider(OpenAIConfig{
				APIKey:    config.APIKey,
				MaxTokens: config.Defaults.MaxTokens,
				Model:     config.Defaults.Model,
			})

			if err != nil {
				logger.Sugar().Errorf("Cannot set up %v provider", config.Name, zap.Error(err))
				continue
			}

			providers = append(providers, provider)

		case "anthropic":
			provider, err := NewAnthropicProvider(AnthropicConfig{
				APIKey:    config.APIKey,
				MaxTokens: config.Defaults.MaxTokens,
				Model:     config.Defaults.Model,
			})

			if err != nil {
				logger.Sugar().Errorf("Cannot set up %v provider", config.Name, zap.Error(err))
				continue
			}

			providers = append(providers, provider)

		case "gemini":
			provider, err := NewGeminiProvider(GeminiConfig{
				APIKey:    config.APIKey,
				MaxTokens: config.Defaults.MaxTokens,
				Model:     config.Defaults.Model,
			})

			if err != nil {
				logger.Sugar().Errorf("Cannot set up %v provider", config.Name, zap.Error(err))
				continue
			}

			providers = append(providers, provider)

		default:
			logger.Sugar().Infof("Warning: unknown provider %s, skipping", config.Name)
		}

	}
	return providers
}
