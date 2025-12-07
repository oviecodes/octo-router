package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type ProviderConfig struct {
	Name    string `mapstructure:"name"`
	APIKey  string `mapstructure:"apiKey"`
	Enabled bool   `mapstructure:"enabled"`
}

type Config struct {
	Providers []ProviderConfig `mapstructure:"providers"`
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
func (c *Config) GetEnabledProviders() []ProviderConfig {
	var enabled []ProviderConfig
	for _, provider := range c.Providers {
		if provider.Enabled {
			enabled = append(enabled, provider)
		}
	}
	return enabled
}

// GetProviderByName returns a specific provider by name
func (c *Config) GetProviderByName(name string) *ProviderConfig {
	for _, provider := range c.Providers {
		if provider.Name == name {
			return &provider
		}
	}
	return nil
}
