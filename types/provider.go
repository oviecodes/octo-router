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
