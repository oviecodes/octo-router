package providers

import (
	"context"
	"errors"
	"fmt"
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
	client    *genai.Client
	maxTokens int64
	model     string
	timeout   time.Duration
}

func (g *GeminiProvider) Complete(ctx context.Context, messages []types.Message) (*types.Message, error) {

	ctx, cancel := context.WithTimeout(ctx, g.timeout)
	defer cancel()
	//convert to geminiMessageformat
	geminiMessages, currentMessage := g.convertMessages(messages)

	chat, err := g.client.Chats.Create(
		ctx,
		g.model,
		nil,
		geminiMessages,
	)

	if err != nil {
		translatedErr := TranslateGeminiError(err)

		var providerErr *ProviderError
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

	res, err := chat.SendMessage(ctx, genai.Part{Text: currentMessage})

	if err != nil {
		translatedErr := TranslateGeminiError(err)

		var providerErr *ProviderError
		if errors.As(translatedErr, &providerErr) {
			logger.Error("Gemini send message failed",
				zap.String("error_type", providerErr.Type.String()),
				zap.Int("status_code", providerErr.StatusCode),
				zap.Bool("retryable", providerErr.Retryable),
				zap.Error(providerErr.OriginalError),
			)
		}

		return nil, translatedErr
	}

	var response *types.Message

	// convert to octo-router message type
	if len(res.Candidates) == 0 {
		return nil, fmt.Errorf("no response candidates from Gemini")
	}

	response = g.convertToRouterMessage(res.Candidates[0].Content.Parts[0].Text)

	return response, nil
}

func (g *GeminiProvider) CompleteStream(ctx context.Context, messages []types.Message) (<-chan *types.StreamChunk, error) {
	ctx, cancel := context.WithTimeout(ctx, g.timeout)
	defer cancel()

	geminiMessages, currentMessage := g.convertMessages(messages)

	chat, err := g.client.Chats.Create(
		ctx,
		g.model,
		nil,
		geminiMessages,
	)

	if err != nil {
		return nil, err
	}

	stream := chat.SendMessageStream(ctx, genai.Part{Text: currentMessage})

	chunks := make(chan *types.StreamChunk)

	go func() {
		defer close(chunks)

		for chunk, err := range stream {
			if err != nil {
				logger.Error("Streaming error occurred", zap.Error(err))
				chunks <- &types.StreamChunk{
					Content: "",
					Done:    true,
					Error:   &err,
				}
				return
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

		chunks <- &types.StreamChunk{
			Content: "",
			Done:    true,
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

func NewGeminiProvider(config GeminiConfig) (*GeminiProvider, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  config.APIKey,
		Backend: genai.BackendGeminiAPI,
	})

	if err != nil {
		return nil, err
	}

	model := selectGeminiModel(config.Model)

	timeout := config.Timeout
	if config.Timeout == 0 {
		timeout = 30 * time.Second
	}

	return &GeminiProvider{
		client:    client,
		maxTokens: config.MaxTokens,
		model:     model,
		timeout:   timeout,
	}, nil
}

func selectGeminiModel(model string) string {
	switch model {
	case "gemini-2.5":
		return "gemini-2.5-flash"
	case "gemini-3":
		return "gemini-3-pro"
	case "gemini-2.5-flash-lite":
		return "gemini-2.5-flash-lite"
	default:
		return "gemini-2.5-flash-lite"
	}
}
