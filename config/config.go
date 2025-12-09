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
	DefaultModels []DefaultModels `mapstructure:"defaults"`
}

type DefaultModels struct {
	Name      string `mapstructure:"name"`
	Model     string `mapstructure:"model"`
	MaxTokens int64  `mapstructure:"maxTokens"`
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
func (c *Config) GetEnabledProviders() []types.ProviderConfigWithExtras {
	var enabled []types.ProviderConfigWithExtras
	for _, provider := range c.Providers {
		if provider.Enabled {
			// get defaults
			providerDefaults := c.GetDefaultModelDataByName(provider.Name)

			providerWithExtras := types.ProviderConfigWithExtras{
				Name:    provider.Name,
				APIKey:  provider.APIKey,
				Enabled: provider.Enabled,
				Defaults: &types.ProviderExtra{
					Model:     providerDefaults.Model,
					MaxTokens: providerDefaults.MaxTokens,
				},
			}
			enabled = append(enabled, providerWithExtras)
		}
	}
	return enabled
}

func (c *Config) GetDetailedModelData() ModelData {
	return ModelData{
		DefaultModels: c.Models.DefaultModels,
	}
}

func (c *Config) GetDefaultModelDataByName(name string) *types.ProviderExtra {
	for _, extra := range c.Models.DefaultModels {
		if extra.Name == name {
			return &types.ProviderExtra{
				Model:     extra.Model,
				MaxTokens: extra.MaxTokens,
			}
		}
	}
	return nil
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
