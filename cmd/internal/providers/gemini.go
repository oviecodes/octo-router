package providers

import (
	"context"
	"errors"
	"fmt"
	"llm-router/cmd/internal/metrics"
	providererrors "llm-router/cmd/internal/provider_errors"
	"llm-router/types"
	"time"

	"github.com/pkoukk/tiktoken-go"
	"go.uber.org/zap"
	"google.golang.org/genai"
)

type GeminiConfig struct {
	APIKey    string
	MaxTokens int64
	Model     string
	Timeout   time.Duration
}

type GeminiProvider struct {
	client          *genai.Client
	maxTokens       int64
	model           string
	standardModelID string
	timeout         time.Duration
}

func (g *GeminiProvider) Complete(ctx context.Context, input *types.CompletionInput) (*types.CompletionResponse, error) {
	start := time.Now()
	providerName := g.GetProviderName()

	ctx, cancel := context.WithTimeout(ctx, g.timeout)
	defer cancel()

	modelToUse := g.model
	standardModelID := g.standardModelID
	if input.Model != "" {

		sdkModel, err := MapToGeminiModel(input.Model)
		if err != nil {
			return nil, fmt.Errorf("invalid gemini model: %w", err)
		}
		modelToUse = sdkModel
		standardModelID = input.Model
	}

	geminiMessages, currentMessage := g.convertMessages(input.Messages)

	chat, err := g.client.Chats.Create(
		ctx,
		modelToUse,
		nil,
		geminiMessages,
	)

	status := "success"

	if err != nil {
		status = "error"

		translatedErr := providererrors.TranslateGeminiError(err)

		var providerErr *providererrors.ProviderError
		if errors.As(translatedErr, &providerErr) {
			logger.Error("Gemini chat creation failed",
				zap.String("error_type", providerErr.Type.String()),
				zap.Int("status_code", providerErr.StatusCode),
				zap.Bool("retryable", providerErr.Retryable),
				zap.Error(providerErr.OriginalError),
			)
		}

		return nil, translatedErr
	}

	duration := time.Since(start).Seconds()

	res, err := chat.SendMessage(ctx, genai.Part{Text: currentMessage})

	if err != nil {
		status = "error"

		translatedErr := providererrors.TranslateGeminiError(err)

		var providerErr *providererrors.ProviderError
		if errors.As(translatedErr, &providerErr) {
			logger.Error("Gemini send message failed",
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

	var response *types.Message

	if len(res.Candidates) == 0 {
		return nil, fmt.Errorf("no response candidates from Gemini")
	}

	metrics.ProviderRequestsTotal.WithLabelValues(providerName, status).Inc()
	metrics.ProviderRequestDuration.WithLabelValues(providerName).Observe(duration)

	usage := &types.Usage{}
	var costUSD float64

	if res.UsageMetadata != nil {
		usage.PromptTokens = int(res.UsageMetadata.PromptTokenCount)
		usage.CompletionTokens = int(res.UsageMetadata.CandidatesTokenCount)
		usage.TotalTokens = usage.PromptTokens + usage.CompletionTokens

		metrics.ProviderTokensUsed.WithLabelValues(providerName, "input").Add(float64(usage.PromptTokens))
		metrics.ProviderTokensUsed.WithLabelValues(providerName, "output").Add(float64(usage.CompletionTokens))

		cost, err := CalculateCost(standardModelID, usage.PromptTokens, usage.CompletionTokens)
		if err != nil {
			logger.Warn("Failed to calculate cost",
				zap.String("provider", providerName),
				zap.String("model", standardModelID),
				zap.Error(err),
			)
		} else {
			costUSD = cost
			metrics.ProviderCostTotal.WithLabelValues(providerName).Add(cost)
			logger.Debug("Request cost calculated",
				zap.String("provider", providerName),
				zap.String("model", standardModelID),
				zap.Int("input_tokens", usage.PromptTokens),
				zap.Int("output_tokens", usage.CompletionTokens),
				zap.Float64("cost_usd", cost),
			)
		}
	}

	response = g.convertToRouterMessage(res.Candidates[0].Content.Parts[0].Text)
	return &types.CompletionResponse{
		Message: *response,
		Usage:   *usage,
		CostUSD: costUSD,
	}, nil
}

func (g *GeminiProvider) CompleteStream(ctx context.Context, input *types.StreamCompletionInput) (<-chan *types.StreamChunk, error) {
	ctx, cancel := context.WithTimeout(ctx, g.timeout)

	modelToUse := g.model
	standardModelID := g.standardModelID
	if input.Model != "" {

		sdkModel, err := MapToGeminiModel(input.Model)
		if err != nil {
			defer cancel()
			return nil, fmt.Errorf("invalid gemini model: %w", err)
		}
		modelToUse = sdkModel
		standardModelID = input.Model
	}

	geminiMessages, currentMessage := g.convertMessages(input.Messages)

	chat, err := g.client.Chats.Create(
		ctx,
		modelToUse,
		nil,
		geminiMessages,
	)

	if err != nil {
		defer cancel()
		return nil, err
	}

	stream := chat.SendMessageStream(ctx, genai.Part{Text: currentMessage})

	chunks := make(chan *types.StreamChunk)

	go func() {
		defer cancel()
		defer close(chunks)

		var finalUsage types.Usage
		for chunk, err := range stream {
			if err != nil {
				logger.Error("Streaming error occurred", zap.Error(err))
				providerErr := providererrors.TranslateGeminiError(err)

				chunks <- &types.StreamChunk{
					Content: "",
					Done:    true,
					Error:   providerErr,
				}
				return
			}

			if chunk.UsageMetadata != nil {
				finalUsage.PromptTokens = int(chunk.UsageMetadata.PromptTokenCount)
				finalUsage.CompletionTokens = int(chunk.UsageMetadata.CandidatesTokenCount)
				finalUsage.TotalTokens = finalUsage.PromptTokens + finalUsage.CompletionTokens
			}

			if len(chunk.Candidates) == 0 ||
				len(chunk.Candidates[0].Content.Parts) == 0 {
				continue
			}

			part := chunk.Candidates[0].Content.Parts[0]

			chunks <- &types.StreamChunk{
				Content: part.Text,
				Done:    false,
			}
		}

		cost, _ := CalculateCost(standardModelID, finalUsage.PromptTokens, finalUsage.CompletionTokens)
		chunks <- &types.StreamChunk{
			Done:    true,
			Usage:   finalUsage,
			CostUSD: cost,
		}
	}()

	return chunks, nil
}

func (g *GeminiProvider) convertMessages(messages []types.Message) ([]*genai.Content, string) {
	var geminiMessages []*genai.Content
	var lastMessage string

	for i, message := range messages {
		var toAppend *genai.Content

		if i == len(messages)-1 {
			lastMessage = message.Content
			break
		}

		if message.Role == "user" || message.Role == "system" {
			toAppend = genai.NewContentFromText(message.Content, genai.RoleUser)
		} else {
			toAppend = genai.NewContentFromText(message.Content, genai.RoleModel)
		}

		geminiMessages = append(geminiMessages, toAppend)
	}

	return geminiMessages, lastMessage
}

func (g *GeminiProvider) convertToRouterMessage(text string) *types.Message {
	return &types.Message{
		Content: text,
		Role:    "assistant",
	}
}

func (g *GeminiProvider) CountTokens(ctx context.Context, messages []types.Message) (int, error) {
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

func (g *GeminiProvider) GetProviderName() string {
	return ProviderGemini
}

func NewGeminiProvider(config GeminiConfig) (*GeminiProvider, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  config.APIKey,
		Backend: genai.BackendGeminiAPI,
	})

	if err != nil {
		return nil, err
	}

	model, err := MapToGeminiModel(config.Model)
	if err != nil {
		return nil, fmt.Errorf("invalid Gemini model: %w", err)
	}

	timeout := config.Timeout
	if config.Timeout == 0 {
		timeout = 30 * time.Second
	}

	return &GeminiProvider{
		client:          client,
		maxTokens:       config.MaxTokens,
		model:           model,
		standardModelID: config.Model,
		timeout:         timeout,
	}, nil
}
