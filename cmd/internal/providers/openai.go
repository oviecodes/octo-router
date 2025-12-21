package providers

import (
	"context"
	"errors"
	"fmt"
	providererrors "llm-router/cmd/internal/provider_errors"
	"llm-router/types"
	"time"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/pkoukk/tiktoken-go"
	"go.uber.org/zap"
)

type OpenAIConfig struct {
	APIKey    string
	MaxTokens int64
	Model     string
	Timeout   time.Duration
	// Circuit *resilience.Circuit
}

type OpenAIProvider struct {
	maxTokens int64
	client    openai.Client
	model     string
	timeout   time.Duration
}

func (o *OpenAIProvider) Complete(ctx context.Context, messages []types.Message) (*types.Message, error) {

	ctx, cancel := context.WithTimeout(ctx, o.timeout)
	defer cancel()

	openAIMessages := o.convertMessages(messages)

	chatCompletion, err := o.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages:            openAIMessages,
		Model:               o.model,
		MaxCompletionTokens: openai.Opt(o.maxTokens),
	})

	if err != nil {
		// Translate to domain error
		translatedErr := providererrors.TranslateOpenAIError(err)

		var providerErr *providererrors.ProviderError
		if errors.As(translatedErr, &providerErr) {
			logger.Error("OpenAI request failed",
				zap.String("error_type", providerErr.Type.String()),
				zap.Int("status_code", providerErr.StatusCode),
				zap.Bool("retryable", providerErr.Retryable),
				zap.Error(providerErr.OriginalError),
			)
		}

		return nil, translatedErr
	}

	response := o.convertToRouterMessage(chatCompletion)

	return &response, nil
}

func (o *OpenAIProvider) CompleteStream(ctx context.Context, messages []types.Message) (<-chan *types.StreamChunk, error) {
	ctx, cancel := context.WithTimeout(ctx, o.timeout)
	defer cancel()

	openAIMessages := o.convertMessages(messages)

	stream := o.client.Chat.Completions.NewStreaming(ctx, openai.ChatCompletionNewParams{
		Messages:            openAIMessages,
		Model:               o.model,
		MaxCompletionTokens: openai.Opt(o.maxTokens),
	})

	acc := openai.ChatCompletionAccumulator{}
	chunks := make(chan *types.StreamChunk)

	go func() {
		defer close(chunks)

		for stream.Next() {
			chunk := stream.Current()

			acc.AddChunk(chunk)

			if _, ok := acc.JustFinishedContent(); ok {
				logger.Debug("Content streaming finished")
				chunks <- &types.StreamChunk{
					Content: "",
					Done:    true,
				}
			}

			if refusal, ok := acc.JustFinishedRefusal(); ok {
				logger.Warn("Content refused by OpenAI", zap.String("refusal", refusal))
			}

			// when I integrate tools calls
			if tool, ok := acc.JustFinishedToolCall(); ok {
				logger.Debug("Tool call finished",
					zap.Int("index", tool.Index),
					zap.String("name", tool.Name),
					zap.String("arguments", tool.Arguments),
				)

			}

			if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
				chunks <- &types.StreamChunk{
					Content: chunk.Choices[0].Delta.Content,
					Done:    false,
				}
			}
		}

		if err := stream.Err(); err != nil {
			logger.Error("Streaming error occurred", zap.Error(err))
			providerErr := providererrors.TranslateOpenAIError(err)

			chunks <- &types.StreamChunk{
				Content: "",
				Done:    true,
				Error:   providerErr,
			}

			return
		}

		// later for token usages stuffs
		if acc.Usage.TotalTokens > 0 {
			logger.Debug("Stream completed",
				zap.Int64("total_tokens", acc.Usage.TotalTokens),
				zap.Int64("prompt_tokens", acc.Usage.PromptTokens),
				zap.Int64("completion_tokens", acc.Usage.CompletionTokens),
			)
		}
	}()

	return chunks, nil
}

func (o *OpenAIProvider) convertMessages(messages []types.Message) []openai.ChatCompletionMessageParamUnion {
	var openAIMessages []openai.ChatCompletionMessageParamUnion

	for _, message := range messages {
		var toAppend openai.ChatCompletionMessageParamUnion
		if message.Role == "user" {
			toAppend = openai.UserMessage(message.Content)
		} else {
			toAppend = openai.AssistantMessage(message.Content)
		}

		openAIMessages = append(openAIMessages, toAppend)

	}

	return openAIMessages
}

func (o *OpenAIProvider) convertToRouterMessage(openAIMessage *openai.ChatCompletion) types.Message {

	return types.Message{
		Role:    string(openAIMessage.Choices[0].Message.Role),
		Content: openAIMessage.Choices[0].Message.Content,
	}
}

func (o *OpenAIProvider) CountTokens(ctx context.Context, messages []types.Message) (int, error) {
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
		// OpenAI format: ~4 tokens per message for structure
		totalTokens += 4
	}

	return totalTokens, nil
}

func (o *OpenAIProvider) GetProviderName(ctx context.Context) string {
	return "openai"
}

func NewOpenAIProvider(config OpenAIConfig) (*OpenAIProvider, error) {
	client := openai.NewClient(
		option.WithAPIKey(config.APIKey),
	)

	model := selectOpenAIModel(config.Model)

	timeout := config.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	return &OpenAIProvider{
		client:    client,
		model:     model,
		maxTokens: int64(config.MaxTokens),
		timeout:   timeout,
	}, nil
}

func selectOpenAIModel(model string) string {
	switch model {
	case "gpt-5":
		return openai.ChatModelGPT5_2025_08_07
	case "gpt-5.1":
		return openai.ChatModelGPT5_1ChatLatest
	case "gpt-4o":
		return openai.ChatModelChatgpt4oLatest
	case "4o-mini":
		return openai.ChatModelGPT4oMini
	case "gpt-3.5":
		return openai.ChatModelGPT3_5Turbo
	default:
		return openai.ChatModelGPT4oMini
	}
}
