package config

import (
	"llm-router/types"
	"os"
	"testing"
)

func TestGetEnabledProviders(t *testing.T) {
	cfg := &Config{
		Providers: []types.ProviderConfig{
			{Name: "openai", APIKey: "key1", Enabled: true},
			{Name: "anthropic", APIKey: "key2", Enabled: false},
			{Name: "gemini", APIKey: "key3", Enabled: true},
		},
		Models: ModelData{

			DefaultModels: map[string]DefaultModels{
				"openai": {Model: "gpt-4", MaxTokens: 4096},
				"gemini": {Model: "gemini-2.5-flash", MaxTokens: 8192},
			},
		},
	}

	enabled := cfg.GetEnabledProviders()

	if len(enabled) != 2 {
		t.Errorf("Expected 2 enabled providers, got %d", len(enabled))
	}

	if enabled[0].Name != "openai" {
		t.Errorf("Expected first provider to be openai, got %s", enabled[0].Name)
	}
}

func TestGetDefaultModelDataByName(t *testing.T) {
	cfg := &Config{
		Models: ModelData{

			DefaultModels: map[string]DefaultModels{
				"openai": {Model: "gpt-4", MaxTokens: 4096},
				"gemini": {Model: "gemini-2.5-flash", MaxTokens: 8192},
			},
		},
	}

	// Test existing provider
	extra := cfg.GetDefaultModelDataByName("openai")
	if extra == nil {
		t.Fatal("Expected non-nil result for openai")
	}
	if extra.Model != "gpt-4" {
		t.Errorf("Expected model gpt-4, got %s", extra.Model)
	}

	// Test non-existing provider
	missing := cfg.GetDefaultModelDataByName("nonexistent")
	if missing != nil {
		t.Error("Expected nil for non-existent provider")
	}
}

func TestEnvironmentVariableSubstitution(t *testing.T) {

	// Set test environment variable
	os.Setenv("OPENAI_API_KEY", "test-key-123")
	os.Setenv("APP_ENV", "test")
	defer os.Unsetenv("TEST_API_KEY")
	defer os.Unsetenv("APP_ENV")

	cfg, err := LoadConfig()

	if err != nil {
		t.Error("Error invalid config")
	}

	enabled := cfg.GetEnabledProviders()

	if enabled[0].APIKey != "test-key-123" {
		t.Errorf("API keys not loaded from environment, expected %s but recieved %s", "test-key-123", enabled[0].APIKey)
	}

}
