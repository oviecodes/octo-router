package providers

import (
	"llm-router/types"
	"llm-router/utils"
	"strings"
	"time"

	"go.uber.org/zap"
)

var logger = utils.SetUpLogger()

func ConfigureProviders(configs []types.ProviderConfigWithExtras) []types.Provider {

	var providers []types.Provider

	for _, config := range configs {
		switch strings.ToLower(config.Name) {
		case "openai":

			if config.APIKey == "" || !config.Enabled {
				logger.Sugar().Errorf("No API Key for %v provider", config.Name)
				continue
			}

			// Validate model ID before provider initialization
			modelProvider, err := ValidateModelID(config.Defaults.Model)
			if err != nil {
				logger.Sugar().Errorf("Invalid model for %v provider: %v", config.Name, err)
				continue
			}
			if modelProvider != ProviderOpenAI {
				logger.Sugar().Errorf("Model %v is not compatible with provider %v (expected %v)", config.Defaults.Model, config.Name, modelProvider)
				continue
			}

			timeoutMs := config.Timeout
			timeout := time.Duration(timeoutMs) * (time.Millisecond)

			provider, err := NewOpenAIProvider(OpenAIConfig{
				APIKey:    config.APIKey,
				MaxTokens: config.Defaults.MaxTokens,
				Model:     config.Defaults.Model,
				Timeout:   timeout,
			})

			if err != nil {
				logger.Sugar().Errorf("Cannot set up %v provider", config.Name, zap.Error(err))
				continue
			}

			providers = append(providers, provider)

		case "anthropic":

			if config.APIKey == "" || !config.Enabled {
				logger.Sugar().Errorf("No API Key for %v provider", config.Name)
				continue
			}

			// Validate model ID before provider initialization
			modelProvider, err := ValidateModelID(config.Defaults.Model)
			if err != nil {
				logger.Sugar().Errorf("Invalid model for %v provider: %v", config.Name, err)
				continue
			}
			if modelProvider != ProviderAnthropic {
				logger.Sugar().Errorf("Model %v is not compatible with provider %v (expected %v)", config.Defaults.Model, config.Name, modelProvider)
				continue
			}

			timeoutMs := config.Timeout
			timeout := time.Duration(timeoutMs) * time.Millisecond

			provider, err := NewAnthropicProvider(AnthropicConfig{
				APIKey:    config.APIKey,
				MaxTokens: config.Defaults.MaxTokens,
				Model:     config.Defaults.Model,
				Timeout:   timeout,
			})

			if err != nil {
				logger.Sugar().Errorf("Cannot set up %v provider", config.Name, zap.Error(err))
				continue
			}

			providers = append(providers, provider)

		case "gemini":

			if config.APIKey == "" || !config.Enabled {
				logger.Sugar().Errorf("No API Key for %v provider", config.Name)
				continue
			}

			// Validate model ID before provider initialization
			modelProvider, err := ValidateModelID(config.Defaults.Model)
			if err != nil {
				logger.Sugar().Errorf("Invalid model for %v provider: %v", config.Name, err)
				continue
			}
			if modelProvider != ProviderGemini {
				logger.Sugar().Errorf("Model %v is not compatible with provider %v (expected %v)", config.Defaults.Model, config.Name, modelProvider)
				continue
			}

			timeoutMs := config.Timeout
			timeout := time.Duration(timeoutMs) * time.Millisecond

			provider, err := NewGeminiProvider(GeminiConfig{
				APIKey:    config.APIKey,
				MaxTokens: config.Defaults.MaxTokens,
				Model:     config.Defaults.Model,
				Timeout:   timeout,
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
