package types

type ProviderConfig struct {
	Name    string `mapstructure:"name"`
	APIKey  string `mapstructure:"apiKey"`
	Enabled bool   `mapstructure:"enabled"`
}

type ProviderExtra struct {
	Model     string
	MaxTokens int64
}

type ProviderConfigWithExtras struct {
	Name     string
	APIKey   string
	Enabled  bool
	Defaults *ProviderExtra
}
