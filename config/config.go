package config

import (
	"fmt"
	"llm-router/types"
	"llm-router/utils"
	"os"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Providers   []types.ProviderConfig `mapstructure:"providers"`
	Routing     types.RoutingData      `mapstructure:"routing"`
	Models      types.ModelData        `mapstructure:"models"`
	Resilience  types.ResilienceData   `mapstructure:"resilience"`
	Limits      types.LimitsData       `mapstructure:"limits"`
	CacheConfig types.CacheData        `mapstructure:"cache"`
	Redis       types.RedisData        `mapstructure:"redis"`
	Security    types.SecurityData     `mapstructure:"security"`
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

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

// GetEnabledProviders returns only the enabled providers
func (c *Config) GetEnabledProviders() []types.ProviderConfigWithExtras {
	var enabled []types.ProviderConfigWithExtras

	for _, provider := range c.Providers {
		if provider.Enabled {

			providerDefaults := c.GetDefaultModelConfigDataByName(provider.Name)
			providerWithExtras := types.ProviderConfigWithExtras{
				Name:    strings.ToLower(provider.Name),
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

// GetProviderConfigWithExtras is an alias for GetEnabledProviders for clarity
func (c *Config) GetProviderConfigWithExtras() []types.ProviderConfigWithExtras {
	return c.GetEnabledProviders()
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

// Validate checks for configuration errors
func (c *Config) Validate() error {

	hasEnabled := false
	for _, p := range c.Providers {
		if p.Enabled {
			hasEnabled = true
			break
		}
	}
	if !hasEnabled {
		return fmt.Errorf("at least one provider must be enabled")
	}

	if c.Routing.Strategy == "weighted" {
		if len(c.Routing.Weights) == 0 {
			return fmt.Errorf("weighted routing strategy requires weights to be defined")
		}
		for name, w := range c.Routing.Weights {
			if w < 0 {
				return fmt.Errorf("weight for provider %s cannot be negative (got %d)", name, w)
			}
		}
	}

	if c.Resilience.Timeout <= 0 {
		return fmt.Errorf("resilience timeout must be greater than 0")
	}

	if c.Redis.Addr == "" && (c.CacheConfig.Enabled) {
		return fmt.Errorf("redis address is required when caching is enabled")
	}

	return nil
}
