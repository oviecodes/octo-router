package types

type RoutingData struct {
	Strategy  string         `mapstructure:"strategy"`
	Weights   map[string]int `mapstructure:"weights"`
	Fallbacks []string       `mapstructure:"fallbacks"`
}

type RouterConfig struct {
	Providers     []ProviderConfigWithExtras
	FallbackChain []string
}

type SelectProviderInput struct {
	Circuits map[string]CircuitBreaker
	Messages []Message
}

type SelectedProviderOutput struct {
	Provider Provider
	Model    string
}
