package providers

import (
	"context"
	"fmt"
	"llm-router/types"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/pkoukk/tiktoken-go"
)

type OpenAIConfig struct {
	APIKey    string
	MaxTokens int
	Model     string
}

type OpenAIProvider struct {
	maxTokens int64
	client    openai.Client
	model     string
}

func (o *OpenAIProvider) Complete(ctx context.Context, messages []types.Message) (*types.Message, error) {

	openAIMessages := o.convertMessages(messages)

	chatCompletion, err := o.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages:            openAIMessages,
		Model:               o.model,
		MaxCompletionTokens: openai.Opt(o.maxTokens),
	})

	if err != nil {
		return nil, err
	}

	response := o.convertToRouterMessage(chatCompletion)

	return &response, nil
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

func NewOpenAIProvider(config OpenAIConfig) (*OpenAIProvider, error) {
	client := openai.NewClient(
		option.WithAPIKey(config.APIKey),
	)

	model := selectOpenAIModel(config.Model)

	return &OpenAIProvider{
		client:    client,
		model:     model,
		maxTokens: int64(config.MaxTokens),
	}, nil
}

func selectOpenAIModel(model string) string {
	switch model {
	case "gpt-5":
		return openai.ChatModelGPT5_2025_08_07
	case "gpt-5.1":
		return openai.ChatModelGPT5_1ChatLatest
	case "gpt-40":
		return openai.ChatModelChatgpt4oLatest
	case "4o-mini":
		return openai.ChatModelGPT4oMini
	case "gpt-3.5":
		return openai.ChatModelGPT3_5Turbo
	default:
		return openai.ChatModelGPT4
	}
}
