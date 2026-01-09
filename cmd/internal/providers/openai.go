package providers

import (
	"context"
	"errors"
	"fmt"
	"llm-router/cmd/internal/metrics"
	providererrors "llm-router/cmd/internal/provider_errors"
	"llm-router/types"
	"strconv"
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
}

type OpenAIProvider struct {
	maxTokens       int64
	client          openai.Client
	model           string // SDK model string
	standardModelID string // Standardized model ID for cost calculation
	timeout         time.Duration
}

func (o *OpenAIProvider) Complete(ctx context.Context, input *types.CompletionInput) (*types.CompletionResponse, error) {
	start := time.Now()
	providerName := o.GetProviderName()

	ctx, cancel := context.WithTimeout(ctx, o.timeout)
	defer cancel()

	modelToUse := o.model
	standardModelID := o.standardModelID
	if input.Model != "" {
		sdkModel, err := MapToOpenAIModel(input.Model)
		if err != nil {
			return nil, fmt.Errorf("invalid openai model: %w", err)
		}
		modelToUse = sdkModel
		standardModelID = input.Model
	}

	openAIMessages := o.convertMessages(input.Messages)

	chatCompletion, err := o.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages:            openAIMessages,
		Model:               openai.ChatModel(modelToUse),
		MaxCompletionTokens: openai.Opt(o.maxTokens),
	})

	duration := time.Since(start).Seconds()
	status := "success"

	if err != nil {
		status = "error"
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

		metrics.ProviderRequestsTotal.WithLabelValues(providerName, status).Inc()
		metrics.ProviderRequestDuration.WithLabelValues(providerName).Observe(duration)
		return nil, translatedErr
	}

	metrics.ProviderRequestsTotal.WithLabelValues(providerName, status).Inc()
	metrics.ProviderRequestDuration.WithLabelValues(providerName).Observe(duration)

	inputTokens := int(chatCompletion.Usage.PromptTokens)
	outputTokens := int(chatCompletion.Usage.CompletionTokens)

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

	response := o.convertToRouterMessage(chatCompletion)

	return &types.CompletionResponse{
		Message: response,
		Headers: map[string]string{
			"cost": strconv.FormatFloat(cost, 'f', -1, 64),
		},
	}, nil
}

func (o *OpenAIProvider) CompleteStream(ctx context.Context, input *types.StreamCompletionInput) (<-chan *types.StreamChunk, error) {
	ctx, cancel := context.WithTimeout(ctx, o.timeout)

	modelToUse := o.model
	// standardModelID := o.standardModelID
	if input.Model != "" {
		sdkModel, err := MapToOpenAIModel(input.Model)
		if err != nil {
			defer cancel()
			return nil, fmt.Errorf("invalid openai model: %w", err)
		}
		modelToUse = sdkModel
		// standardModelID = input.Model
	}

	fmt.Printf("model to use %v \n", modelToUse)

	openAIMessages := o.convertMessages(input.Messages)

	stream := o.client.Chat.Completions.NewStreaming(ctx, openai.ChatCompletionNewParams{
		Messages:            openAIMessages,
		Model:               modelToUse,
		MaxCompletionTokens: openai.Opt(o.maxTokens),
	})

	acc := openai.ChatCompletionAccumulator{}
	chunks := make(chan *types.StreamChunk)

	go func() {
		defer cancel()
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

	encoding, err := tiktoken.GetEncoding("cl100k_base")
	if err != nil {
		return 0, fmt.Errorf("failed to get tiktoken encoding: %w", err)
	}

	totalTokens := 0
	for _, msg := range messages {

		tokens := encoding.Encode(msg.Content, nil, nil)
		totalTokens += len(tokens)

		// Add overhead for role and message formatting
		// OpenAI format: ~4 tokens per message for structure
		totalTokens += 4
	}

	return totalTokens, nil
}

func (o *OpenAIProvider) GetProviderName() string {
	return ProviderOpenAI
}

func NewOpenAIProvider(config OpenAIConfig) (*OpenAIProvider, error) {
	client := openai.NewClient(
		option.WithAPIKey(config.APIKey),
	)

	// Map standardized model ID to OpenAI SDK model
	model, err := MapToOpenAIModel(config.Model)
	if err != nil {
		return nil, fmt.Errorf("invalid OpenAI model: %w", err)
	}

	timeout := config.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	return &OpenAIProvider{
		client:          client,
		model:           model,
		standardModelID: config.Model,
		maxTokens:       int64(config.MaxTokens),
		timeout:         timeout,
	}, nil
}
