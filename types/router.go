package types

type CostOptions struct {
	DefaultTier  string `mapstructure:"defaultTier"`  // Default tier when not specified in request
	MinimumTier  string `mapstructure:"minimumTier"`  // Absolute minimum tier to use
	TierStrategy string `mapstructure:"tierStrategy"` // "same-tier", "allow-downgrade", "cheapest"
}

type SemanticGroup struct {
	Name               string   `mapstructure:"name"`
	IntentKeywords     []string `mapstructure:"intent_keywords"`
	IntentDescription  string   `mapstructure:"intent_description"`
	Examples           []string `mapstructure:"examples"` // Few-shot examples for better matching
	AllowProviders     []string `mapstructure:"allow_providers"`
	RequiredCapability string   `mapstructure:"required_capability"`
}

type SemanticPolicy struct {
	Enabled      bool            `mapstructure:"enabled"`
	Engine       string          `mapstructure:"engine"` // "keyword", "embedding"
	Threshold    float64         `mapstructure:"threshold"`
	ModelPath    string          `mapstructure:"model_path"`
	DefaultGroup string          `mapstructure:"default_group"`
	Groups       []SemanticGroup `mapstructure:"groups"`
}

type Policies struct {
	Semantic *SemanticPolicy `mapstructure:"semantic"`
}

type RoutingData struct {
	Strategy    string         `mapstructure:"strategy"`
	Weights     map[string]int `mapstructure:"weights"`
	Fallbacks   []string       `mapstructure:"fallbacks"`
	Policies    *Policies      `mapstructure:"policies"`
	CostOptions *CostOptions   `mapstructure:"costOptions"`
}

type RouterConfig struct {
	Providers     []ProviderConfigWithExtras
	FallbackChain []string
}

type SelectProviderInput struct {
	Circuits   map[string]CircuitBreaker
	Messages   []Message
	Tier       string     // Requested tier (optional)
	Candidates []Provider // Optional: Pre-filtered list of providers (e.g. from Semantic Router)
}

type SelectedProviderOutput struct {
	Provider   Provider
	Model      string
	Candidates []Provider // The filtered pool of candidates
}
