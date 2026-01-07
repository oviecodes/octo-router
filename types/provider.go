package types

import (
	"context"
)

type Provider interface {
	Complete(ctx context.Context, messages []Message) (*Message, error)
	CountTokens(ctx context.Context, messages []Message) (int, error)
	CompleteStream(ctx context.Context, messages []Message) (<-chan *StreamChunk, error)
	GetProviderName() string
}

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
	Timeout  int
	Limits   ProviderLimits
}

type ProviderWithModel struct {
	Provider Provider
	Model    string
}
