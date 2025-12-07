package types

type RouterConfig struct {
	Providers []ProviderConfig
	MaxTokens int64
	Model     string
}
