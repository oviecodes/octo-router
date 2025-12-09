package config

import (
	"fmt"
	"llm-router/types"
	"llm-router/utils"
	"os"

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

var logger = utils.SetUpLogger()

// LoadConfig reads the config.yaml file from the project root
func LoadConfig() (*Config, error) {

	var configFile string

	if os.Getenv("APP_ENV") == "test" {
		configFile = "config_test"
	} else {
		configFile = "config"
	}

	// Enable environment variable substitution
	viper.AutomaticEnv()
	viper.SetEnvPrefix("")

	viper.SetConfigName(configFile)
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

	// Expand environment variables in provider API keys
	for i := range config.Providers {
		config.Providers[i].APIKey = os.ExpandEnv(config.Providers[i].APIKey)
	}

	return &config, nil
}

// GetEnabledProviders returns only the enabled providers
func (c *Config) GetEnabledProviders() []types.ProviderConfigWithExtras {
	var enabled []types.ProviderConfigWithExtras
	for _, provider := range c.Providers {
		if provider.Enabled {

			providerDefaults := c.GetDefaultModelDataByName(provider.Name)
			providerWithExtras := types.ProviderConfigWithExtras{
				Name:    provider.Name,
				APIKey:  provider.APIKey,
				Enabled: provider.Enabled,
			}

			if providerDefaults != nil {
				providerWithExtras.Defaults = &types.ProviderExtra{
					Model:     providerDefaults.Model,
					MaxTokens: providerDefaults.MaxTokens,
				}
			} else {
				logger.Sugar().Infof("Falling back to system defined defaults for the %v provider", provider.Name)
				providerWithExtras.Defaults = &types.ProviderExtra{
					Model:     "",
					MaxTokens: 4096,
				}
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
