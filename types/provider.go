package types

import (
	"context"
)

type Provider interface {
	Complete(ctx context.Context, input *CompletionInput) (*CompletionResponse, error)
	CountTokens(ctx context.Context, messages []Message) (int, error)
	CompleteStream(ctx context.Context, data *StreamCompletionInput) (<-chan *StreamChunk, error)
	GetProviderName() string
}

type CompletionInput struct {
	Model    string
	Messages []Message
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type CompletionResponse struct {
	Message Message           `json:"message"`
	Usage   Usage             `json:"usage"`
	CostUSD float64           `json:"cost_usd"`
	Headers map[string]string `json:"-"`
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

type StreamCompletionInput struct {
	Model    string
	Messages []Message
}
