package config

import (
	"fmt"
	"llm-router/types"

	"github.com/spf13/viper"
)

type Config struct {
	Providers []types.ProviderConfig `mapstructure:"providers"`
	MaxTokens int64                  `mapstructure:"models.maxTokens"`
	Model     string                 `mapstructure:"models.default"`
}

// LoadConfig reads the config.yaml file from the project root
func LoadConfig() (*Config, error) {

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./")
	viper.AddConfigPath("../")
	viper.AddConfigPath("../../")

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}

// GetEnabledProviders returns only the enabled providers
func (c *Config) GetEnabledProviders() []types.ProviderConfig {
	var enabled []types.ProviderConfig
	for _, provider := range c.Providers {
		if provider.Enabled {
			enabled = append(enabled, provider)
		}
	}
	return enabled
}

// GetProviderByName returns a specific provider by name
func (c *Config) GetProviderByName(name string) *types.ProviderConfig {
	for _, provider := range c.Providers {
		if provider.Name == name {
			return &provider
		}
	}
	return nil
}
