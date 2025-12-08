package providers

import (
	"context"
	"llm-router/types"
	"os"

	"go.uber.org/zap"
)

type Provider interface {
	Complete(ctx context.Context, messages []types.Message) (*types.Message, error)
	CountTokens(ctx context.Context, messages []types.Message) (int, error)
}

var logger = setUpLogger()

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
				logger.Sugar().Infof("Cannot set up %v provider", config.Name)
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
				logger.Sugar().Infof("Cannot set up %v provider", config.Name)
				continue
			}

			providers = append(providers, provider)

		default:
			logger.Sugar().Infof("Warning: unknown provider %s, skipping", config.Name)
		}

	}
	return providers
}

func setUpLogger() *zap.Logger {
	switch os.Getenv("APP_ENV") {
	case "production":
		log, _ := zap.NewProduction()
		return log
	default:
		log, _ := zap.NewDevelopment()
		return log
	}
}
