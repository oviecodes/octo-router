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
			Model:     "openai/gpt-4o-mini",
			MaxTokens: 1024,
		},
	},
	{
		Name:    "gemini",
		APIKey:  "AIza-****",
		Enabled: true,
		Defaults: &types.ProviderExtra{
			Model:     "gemini/gemini-2.5-flash",
			MaxTokens: 1024,
		},
	},
	{
		Name:    "anthropic",
		APIKey:  "sk-****",
		Enabled: true,
		Defaults: &types.ProviderExtra{
			Model:     "anthropic/claude-haiku-3",
			MaxTokens: 1024,
		},
	},
}

// Test provider factory
func TestProviderFactory(t *testing.T) {
	factory := NewProviderFactory()
	providers := factory.CreateProviders(providerConfigs)

	// should return all providers properly
	if len(providers) != len(providerConfigs) {
		t.Fatalf("expected length of %T to be equal to %v", providers, len(providerConfigs))
	}
}

func TestProviderFactoryWithUnknownProvider(t *testing.T) {
	configs := []types.ProviderConfigWithExtras{
		{
			Name:    "unknown-provider",
			APIKey:  "test-key",
			Enabled: true,
			Defaults: &types.ProviderExtra{
				Model:     "test-model",
				MaxTokens: 1024,
			},
		},
	}

	factory := NewProviderFactory()
	providers := factory.CreateProviders(configs)

	// Should skip unknown providers
	if len(providers) != 0 {
		t.Errorf("expected 0 providers for unknown provider type, got %d", len(providers))
	}
}

func TestProviderFactoryWithInvalidAPIKey(t *testing.T) {
	configs := []types.ProviderConfigWithExtras{
		{
			Name:    "openai",
			APIKey:  "",
			Enabled: true,
			Defaults: &types.ProviderExtra{
				Model:     "openai/gpt-4o-mini",
				MaxTokens: 1024,
			},
		},
	}

	factory := NewProviderFactory()
	providers := factory.CreateProviders(configs)

	// Should skip invalid providers
	if len(providers) != 0 {
		t.Errorf("expected 0 providers with invalid API key, got %d", len(providers))
	}
}

// Test model selection functions
// func TestSelectOpenAIModel(t *testing.T) {
// 	tests := []struct {
// 		input    string
// 		expected string
// 	}{
// 		{"gpt-4o", "chatgpt-4o-latest"},
// 		{"gpt-5", "gpt-5-2025-08-07"},
// 		{"", "gpt-4o-mini"},
// 		{"invalid", "gpt-4o-mini"},
// 	}

// 	for _, tt := range tests {
// 		result := selectOpenAIModel(tt.input)
// 		if result != tt.expected {
// 			t.Errorf("selectOpenAIModel(%q) = %q, want %q", tt.input, result, tt.expected)
// 		}
// 	}
// }

// func TestSelectAnthropicModel(t *testing.T) {
// 	tests := []struct {
// 		input    string
// 		expected string
// 	}{
// 		{"opus", "claude-opus-4-5-20251101"},
// 		{"haiku", "claude-haiku-4-5-20251001"},
// 		{"sonnet", "claude-4-sonnet-20250514"},
// 		{"", "claude-3-haiku-20240307"},
// 	}

// 	for _, tt := range tests {
// 		result := selectAnthropicModel(tt.input)
// 		if string(result) != tt.expected {
// 			t.Errorf("selectAnthropicModel(%q) = %q, want %q", tt.input, result, tt.expected)
// 		}
// 	}
// }

// func TestSelectGeminiModel(t *testing.T) {
// 	tests := []struct {
// 		input    string
// 		expected string
// 	}{
// 		{"gemini-2.5", "gemini-2.5-flash"},
// 		{"gemini-3", "gemini-3-pro"},
// 		{"", "gemini-1.5-flash"},
// 	}

// 	for _, tt := range tests {
// 		result := selectGeminiModel(tt.input)
// 		if result != tt.expected {
// 			t.Errorf("selectGeminiModel(%q) = %q, want %q", tt.input, result, tt.expected)
// 		}
// 	}
// }
