package config

import (
	"fmt"
	"llm-router/types"

	"github.com/spf13/viper"
)

type Config struct {
	Providers []types.ProviderConfig `mapstructure:"providers"`
	Routing   any                    `mapstructure:"routing"`
	Models    ModelData              `mapstructure:"models"`
}

type ModelData struct {
	Model    string `mapstructure:"default"`
	MaxToken int64  `mapstructure:"maxTokens"`
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

func (c *Config) GetModelData() ModelData {
	return ModelData{
		Model:    c.Models.Model,
		MaxToken: c.Models.MaxToken,
	}
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
