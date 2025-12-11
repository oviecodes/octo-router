package providers

import (
	"context"
	"fmt"
	"llm-router/types"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/pkoukk/tiktoken-go"
	"go.uber.org/zap"
)

type AnthropicConfig struct {
	APIKey    string
	MaxTokens int64
	Model     string
	Timeout   time.Duration
}

type AnthropicProvider struct {
	client    anthropic.Client
	model     anthropic.Model
	maxTokens int64
	timeout   time.Duration
}

func (a *AnthropicProvider) Complete(ctx context.Context, messages []types.Message) (*types.Message, error) {

	ctx, cancel := context.WithTimeout(ctx, a.timeout)
	defer cancel()

	// Send all messages in the conversation
	anthropicMessages := a.convertMessages(messages)

	message, err := a.client.Messages.New(ctx, anthropic.MessageNewParams{
		MaxTokens: a.maxTokens,
		Messages:  anthropicMessages,
		Model:     a.model,
	})

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("anthropic provider timeout after %v: %w", a.timeout, err)
		}
		return nil, err
	}

	response := a.convertToRouterMessage(message)
	return response, nil
}

func (a *AnthropicProvider) CompleteStream(ctx context.Context, messages []types.Message) (<-chan *types.StreamChunk, error) {
	ctx, cancel := context.WithTimeout(ctx, a.timeout)
	defer cancel()

	anthropicMessages := a.convertMessages(messages)

	stream := a.client.Messages.NewStreaming(ctx, anthropic.MessageNewParams{
		MaxTokens: a.maxTokens,
		Messages:  anthropicMessages,
		Model:     a.model,
	})

	chunks := make(chan *types.StreamChunk)
	message := anthropic.Message{}
	go func() {
		defer close(chunks)

		for stream.Next() {
			event := stream.Current()

			err := message.Accumulate(event)
			if err != nil {
				logger.Error("Failed to accumulate message event", zap.Error(err))
				chunks <- &types.StreamChunk{
					Content: "",
					Done:    true,
					Error:   &err,
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
				chunks <- &types.StreamChunk{
					Content: "",
					Done:    true,
				}
			}
		}

		if err := stream.Err(); err != nil {
			logger.Sugar().Errorf("An error occurred while streaming: %v", err)
			chunks <- &types.StreamChunk{
				Content: "",
				Done:    true,
				Error:   &err,
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
	// Use tiktoken to estimate tokens locally (fast, no API calls, no rate limits)
	encoding, err := tiktoken.GetEncoding("cl100k_base")
	if err != nil {
		return 0, fmt.Errorf("failed to get tiktoken encoding: %w", err)
	}

	totalTokens := 0
	for _, msg := range messages {
		// Count tokens in content
		tokens := encoding.Encode(msg.Content, nil, nil)
		totalTokens += len(tokens)

		// Add overhead for role and message formatting
		// Anthropic format: ~4 tokens per message for structure
		totalTokens += 4
	}

	return totalTokens, nil
}

func NewAnthropicProvider(config AnthropicConfig) (*AnthropicProvider, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("API key cannot be empty")
	}

	client := anthropic.NewClient(
		option.WithAPIKey(config.APIKey),
	)

	model := selectAnthropicModel(config.Model)

	timeout := config.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	return &AnthropicProvider{
		client:    client,
		model:     model,
		maxTokens: int64(config.MaxTokens),
		timeout:   timeout,
	}, nil
}

func selectAnthropicModel(model string) anthropic.Model {
	switch model {
	case "opus":
		return anthropic.ModelClaudeOpus4_5_20251101
	case "sonnet":
		return anthropic.ModelClaude4Sonnet20250514
	case "haiku":
		return anthropic.ModelClaudeHaiku4_5_20251001
	default:
		return anthropic.ModelClaude_3_Haiku_20240307
	}
}
