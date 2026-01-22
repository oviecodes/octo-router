package providers

import (
	"context"
	"errors"
	"fmt"
	"llm-router/cmd/internal/metrics"
	providererrors "llm-router/cmd/internal/provider_errors"
	"llm-router/types"
	"llm-router/utils"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/pkoukk/tiktoken-go"
	"go.uber.org/zap"
)

var logger = utils.SetUpLogger()

type AnthropicConfig struct {
	APIKey    string
	MaxTokens int64
	Model     string
	Timeout   time.Duration
}

type AnthropicProvider struct {
	client          anthropic.Client
	model           anthropic.Model // SDK model
	standardModelID string          // Standardized model ID for cost calculation
	maxTokens       int64
	timeout         time.Duration
}

func (a *AnthropicProvider) Complete(ctx context.Context, input *types.CompletionInput) (*types.CompletionResponse, error) {
	start := time.Now()
	providerName := a.GetProviderName()

	ctx, cancel := context.WithTimeout(ctx, a.timeout)
	defer cancel()

	modelToUse := a.model
	standardModelID := a.standardModelID
	if input.Model != "" {
		sdkModel, err := MapToAnthropicModel(input.Model)
		if err != nil {
			return nil, fmt.Errorf("invalid anthropic model: %w", err)
		}
		modelToUse = sdkModel
		standardModelID = input.Model
	}

	anthropicMessages := a.convertMessages(input.Messages)

	message, err := a.client.Messages.New(ctx, anthropic.MessageNewParams{
		MaxTokens: a.maxTokens,
		Messages:  anthropicMessages,
		Model:     modelToUse,
	})

	duration := time.Since(start).Seconds()
	status := "success"

	if err != nil {
		status = "error"
		translatedErr := providererrors.TranslateAnthropicError(err)

		var providerErr *providererrors.ProviderError
		if errors.As(translatedErr, &providerErr) {
			logger.Error("Anthropic request failed",
				zap.String("error_type", providerErr.Type.String()),
				zap.Int("status_code", providerErr.StatusCode),
				zap.Bool("retryable", providerErr.Retryable),
				zap.Error(providerErr.OriginalError),
			)
		}

		metrics.ProviderRequestsTotal.WithLabelValues(providerName, status).Inc()
		metrics.ProviderRequestDuration.WithLabelValues(providerName).Observe(duration)

		return nil, translatedErr
	}

	metrics.ProviderRequestsTotal.WithLabelValues(providerName, status).Inc()
	metrics.ProviderRequestDuration.WithLabelValues(providerName).Observe(duration)

	inputTokens := int(message.Usage.InputTokens)
	outputTokens := int(message.Usage.OutputTokens)

	metrics.ProviderTokensUsed.WithLabelValues(providerName, "input").Add(float64(inputTokens))
	metrics.ProviderTokensUsed.WithLabelValues(providerName, "output").Add(float64(outputTokens))

	cost, err := CalculateCost(standardModelID, inputTokens, outputTokens)
	if err != nil {
		logger.Warn("Failed to calculate cost",
			zap.String("provider", providerName),
			zap.String("model", standardModelID),
			zap.Error(err),
		)
	} else {
		metrics.ProviderCostTotal.WithLabelValues(providerName).Add(cost)
		logger.Debug("Request cost calculated",
			zap.String("provider", providerName),
			zap.String("model", standardModelID),
			zap.Int("input_tokens", inputTokens),
			zap.Int("output_tokens", outputTokens),
			zap.Float64("cost_usd", cost),
		)
	}

	response := a.convertToRouterMessage(message)
	return &types.CompletionResponse{
		Message: *response,
		Usage: types.Usage{
			PromptTokens:     inputTokens,
			CompletionTokens: outputTokens,
			TotalTokens:      inputTokens + outputTokens,
		},
		CostUSD: cost,
	}, nil
}

func (a *AnthropicProvider) CompleteStream(ctx context.Context, input *types.StreamCompletionInput) (<-chan *types.StreamChunk, error) {
	ctx, cancel := context.WithTimeout(ctx, a.timeout)

	modelToUse := a.model
	standardModelID := a.standardModelID
	if input.Model != "" {
		sdkModel, err := MapToAnthropicModel(input.Model)
		if err != nil {
			cancel()
			return nil, fmt.Errorf("invalid anthropic model: %w", err)
		}
		modelToUse = sdkModel
		standardModelID = input.Model
	}

	anthropicMessages := a.convertMessages(input.Messages)

	stream := a.client.Messages.NewStreaming(ctx, anthropic.MessageNewParams{
		MaxTokens: a.maxTokens,
		Messages:  anthropicMessages,
		Model:     modelToUse,
	})

	chunks := make(chan *types.StreamChunk)
	message := anthropic.Message{}
	go func() {
		defer cancel()
		defer close(chunks)

		for stream.Next() {
			event := stream.Current()

			err := message.Accumulate(event)
			if err != nil {
				logger.Error("Failed to accumulate message event", zap.Error(err))
				providerErr := providererrors.TranslateOpenAIError(err)
				chunks <- &types.StreamChunk{
					Content: "",
					Done:    true,
					Error:   providerErr,
				}

				return
			}

			switch eventVariant := event.AsAny().(type) {
			case anthropic.ContentBlockDeltaEvent:
				switch deltaVariant := eventVariant.Delta.AsAny().(type) {
				case anthropic.TextDelta:
					chunks <- &types.StreamChunk{
						Content: deltaVariant.Text,
						Done:    false,
					}
				}

			case anthropic.MessageStopEvent:
				inputTokens := int(message.Usage.InputTokens)
				outputTokens := int(message.Usage.OutputTokens)
				cost, _ := CalculateCost(standardModelID, inputTokens, outputTokens)

				chunks <- &types.StreamChunk{
					Done: true,
					Usage: types.Usage{
						PromptTokens:     inputTokens,
						CompletionTokens: outputTokens,
						TotalTokens:      inputTokens + outputTokens,
					},
					CostUSD: cost,
				}
			}
		}

		if err := stream.Err(); err != nil {
			logger.Sugar().Errorf("An error occurred while streaming: %v", err)
			providerErr := providererrors.TranslateAnthropicError(err)
			chunks <- &types.StreamChunk{
				Content: "",
				Done:    true,
				Error:   providerErr,
			}
		}
	}()

	return chunks, nil
}

func (a *AnthropicProvider) convertMessages(messages []types.Message) []anthropic.MessageParam {
	var anthropicMessages []anthropic.MessageParam

	for _, message := range messages {
		var toAppend anthropic.MessageParam

		if message.Role == "user" {
			toAppend = anthropic.NewUserMessage(anthropic.NewTextBlock(message.Content))
		} else {
			toAppend = anthropic.NewAssistantMessage(anthropic.NewTextBlock(message.Content))
		}

		anthropicMessages = append(anthropicMessages, toAppend)
	}
	return anthropicMessages
}

func (a *AnthropicProvider) convertToRouterMessage(message *anthropic.Message) *types.Message {
	var content string

	for _, block := range message.Content {
		switch block.Type {
		case "text":
			content = block.Text
		case "thinking":
			content = block.Thinking
		}

		if content != "" {
			break
		}
	}

	return &types.Message{
		Role:    string(message.Role),
		Content: content,
	}
}

func (a *AnthropicProvider) CountTokens(ctx context.Context, messages []types.Message) (int, error) {
	encoding, err := tiktoken.GetEncoding("cl100k_base")
	if err != nil {
		return 0, fmt.Errorf("failed to get tiktoken encoding: %w", err)
	}

	totalTokens := 0
	for _, msg := range messages {
		tokens := encoding.Encode(msg.Content, nil, nil)
		totalTokens += len(tokens)

		// Add overhead for role and message formatting
		// Anthropic format: ~4 tokens per message for structure
		totalTokens += 4
	}

	return totalTokens, nil
}

func (a *AnthropicProvider) GetProviderName() string {
	return ProviderAnthropic
}

func NewAnthropicProvider(config AnthropicConfig) (*AnthropicProvider, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("API key cannot be empty")
	}

	client := anthropic.NewClient(
		option.WithAPIKey(config.APIKey),
	)

	model, err := MapToAnthropicModel(config.Model)
	if err != nil {
		return nil, fmt.Errorf("invalid Anthropic model: %w", err)
	}

	timeout := config.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	return &AnthropicProvider{
		client:          client,
		model:           model,
		standardModelID: config.Model,
		maxTokens:       int64(config.MaxTokens),
		timeout:         timeout,
	}, nil
}
