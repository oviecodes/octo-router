package config

import (
	"fmt"
	"llm-router/types"
	"llm-router/utils"
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	Providers   []types.ProviderConfig `mapstructure:"providers"`
	Routing     types.RoutingData      `mapstructure:"routing"`
	Models      types.ModelData        `mapstructure:"models"`
	Resilience  types.ResilienceData   `mapstructure:"resilience"`
	Limits      types.LimitsData       `mapstructure:"limits"`
	CacheConfig types.CacheData        `mapstructure:"cache"`
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

	config.DeduplicateProviders()

	return &config, nil
}

// GetEnabledProviders returns only the enabled providers
func (c *Config) GetEnabledProviders() []types.ProviderConfigWithExtras {
	var enabled []types.ProviderConfigWithExtras

	for _, provider := range c.Providers {
		if provider.Enabled {

			providerDefaults := c.GetDefaultModelConfigDataByName(provider.Name)
			providerWithExtras := types.ProviderConfigWithExtras{
				Name:    provider.Name,
				APIKey:  provider.APIKey,
				Enabled: provider.Enabled,
				Timeout: c.Resilience.Timeout,
				Limits:  c.GetLimitsConfigDataByName(provider.Name),
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

func (c *Config) GetRouterStrategy() *types.RoutingData {
	return &c.Routing
}

func (c *Config) DeduplicateProviders() {

	var dedupConfigs []types.ProviderConfig
	providerMap := map[string]types.ProviderConfig{}

	for _, provider := range c.Providers {
		providerMap[provider.Name] = provider
	}

	for _, provider := range providerMap {
		dedupConfigs = append(dedupConfigs, provider)
	}

	c.Providers = dedupConfigs
}

func (c *Config) GetDetailedModelConfigData() types.ModelData {
	return types.ModelData{
		DefaultModels: c.Models.DefaultModels,
	}
}

func (c *Config) GetResilienceConfigData() types.ResilienceData {
	return types.ResilienceData{
		Timeout:              c.Resilience.Timeout,
		RetriesConfig:        c.Resilience.RetriesConfig,
		CircuitBreakerConfig: c.Resilience.CircuitBreakerConfig,
	}
}

func (c *Config) GetLimitsConfigData() *types.LimitsData {
	return &types.LimitsData{
		RequestsPerMinute: c.Limits.RequestsPerMinute,
		RequestsPerDay:    c.Limits.RequestsPerDay,
		DailyBudget:       c.Limits.DailyBudget,
		AlertThreshold:    c.Limits.AlertThreshold,
		Providers:         c.Limits.Providers,
	}
}

func (c *Config) GetLimitsConfigDataByName(name string) types.ProviderLimits {
	return c.Limits.Providers[name]
}

func (c *Config) GetDefaultModelConfigDataByName(name string) *types.ProviderExtra {
	extra, ok := c.Models.DefaultModels[name]

	if !ok {
		return nil
	}
	return &types.ProviderExtra{
		Model:     extra.Model,
		MaxTokens: extra.MaxTokens,
	}
}

func (c *Config) GetProviderByName(name string) *types.ProviderConfig {
	for _, provider := range c.Providers {
		if provider.Name == name {
			return &provider
		}
	}
	return nil
}

func (c *Config) GetCacheConfigData() *types.CacheData {
	return &c.CacheConfig
}
