package providers

import (
	"llm-router/types"
	"testing"
)

var providerConfigs = []types.ProviderConfigWithExtras{
	{
		Name:    "openai",
		APIKey:  "sk-****",
		Enabled: true,
		Defaults: &types.ProviderExtra{
			Model:     "",
			MaxTokens: 1024,
		},
	},
	{
		Name:    "gemini",
		APIKey:  "AIza-****",
		Enabled: true,
		Defaults: &types.ProviderExtra{
			Model:     "gemini-1.5-flash",
			MaxTokens: 1024,
		},
	},
	{
		Name:    "anthropic",
		APIKey:  "sk-****",
		Enabled: true,
		Defaults: &types.ProviderExtra{
			Model:     "haiku",
			MaxTokens: 1024,
		},
	},
}

// creation of providers
func TestConfigureProviders(t *testing.T) {
	// returns an array of providers
	providers := ConfigureProviders(providerConfigs)

	if len(providers) != len(providerConfigs) {
		t.Fatalf("expected length of %T to be equal to %v", providers, len(providerConfigs))
	}
}
