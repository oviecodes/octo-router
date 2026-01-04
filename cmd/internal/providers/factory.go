package providers

import (
	"fmt"
	"llm-router/types"
	"llm-router/utils"
	"strings"
	"time"

	"go.uber.org/zap"
)

var factoryLogger = utils.SetUpLogger()

type ProviderFactory struct{}

func NewProviderFactory() *ProviderFactory {
	return &ProviderFactory{}
}

func (f *ProviderFactory) CreateProvider(config types.ProviderConfigWithExtras) (types.Provider, error) {
	if err := f.validateConfig(config); err != nil {
		return nil, err
	}

	if err := f.validateModelCompatibility(config); err != nil {
		return nil, err
	}

	timeout := time.Duration(config.Timeout) * time.Millisecond

	switch strings.ToLower(config.Name) {
	case ProviderOpenAI:
		return NewOpenAIProvider(OpenAIConfig{
			APIKey:    config.APIKey,
			MaxTokens: config.Defaults.MaxTokens,
			Model:     config.Defaults.Model,
			Timeout:   timeout,
		})

	case ProviderAnthropic:
		return NewAnthropicProvider(AnthropicConfig{
			APIKey:    config.APIKey,
			MaxTokens: config.Defaults.MaxTokens,
			Model:     config.Defaults.Model,
			Timeout:   timeout,
		})

	case ProviderGemini:
		return NewGeminiProvider(GeminiConfig{
			APIKey:    config.APIKey,
			MaxTokens: config.Defaults.MaxTokens,
			Model:     config.Defaults.Model,
			Timeout:   timeout,
		})

	default:
		return nil, fmt.Errorf("unknown provider: %s", config.Name)
	}
}

func (f *ProviderFactory) validateConfig(config types.ProviderConfigWithExtras) error {
	if config.APIKey == "" {
		return fmt.Errorf("API key is required for provider %s", config.Name)
	}

	if !config.Enabled {
		return fmt.Errorf("provider %s is disabled", config.Name)
	}

	if config.Defaults == nil {
		return fmt.Errorf("default configuration is required for provider %s", config.Name)
	}

	if config.Defaults.Model == "" {
		return fmt.Errorf("model is required for provider %s", config.Name)
	}

	return nil
}

func (f *ProviderFactory) validateModelCompatibility(config types.ProviderConfigWithExtras) error {
	modelProvider, err := ValidateModelID(config.Defaults.Model)
	if err != nil {
		return fmt.Errorf("invalid model for provider %s: %w", config.Name, err)
	}

	expectedProvider := strings.ToLower(config.Name)
	if modelProvider != expectedProvider {
		return fmt.Errorf("model %s is not compatible with provider %s (expected %s)",
			config.Defaults.Model, config.Name, modelProvider)
	}

	return nil
}

func (f *ProviderFactory) CreateProviders(configs []types.ProviderConfigWithExtras) []types.Provider {
	var providers []types.Provider

	for _, config := range configs {
		provider, err := f.CreateProvider(config)
		if err != nil {
			factoryLogger.Warn("Failed to create provider",
				zap.String("provider", config.Name),
				zap.Error(err),
			)
			continue
		}

		providers = append(providers, provider)
		factoryLogger.Info("Provider created successfully",
			zap.String("provider", config.Name),
			zap.String("model", config.Defaults.Model),
		)
	}

	return providers
}
