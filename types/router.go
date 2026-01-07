package types

type CostOptions struct {
	DefaultTier  string `mapstructure:"defaultTier"`  // Default tier when not specified in request
	MinimumTier  string `mapstructure:"minimumTier"`  // Absolute minimum tier to use
	TierStrategy string `mapstructure:"tierStrategy"` // "same-tier", "allow-downgrade", "cheapest"
}

type RoutingData struct {
	Strategy    string         `mapstructure:"strategy"`
	Weights     map[string]int `mapstructure:"weights"`
	Fallbacks   []string       `mapstructure:"fallbacks"`
	CostOptions *CostOptions   `mapstructure:"costOptions"`
}

type RouterConfig struct {
	Providers     []ProviderConfigWithExtras
	FallbackChain []string
}

type SelectProviderInput struct {
	Circuits map[string]CircuitBreaker
	Messages []Message
	Tier     string // Requested tier (optional)
}

type SelectedProviderOutput struct {
	Provider Provider
	Model    string
}
