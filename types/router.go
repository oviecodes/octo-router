package types

type RoutingData struct {
	Strategy  string         `mapstructure:"strategy"`
	Weights   map[string]int `mapstructure:"weights"`
	Fallbacks []string       `mapstructure:"fallbacks"`
}

type RouterConfig struct {
	Providers            []ProviderConfigWithExtras
	CircuitBreakerConfig map[string]int
	RetryConfig          map[string]int
}
