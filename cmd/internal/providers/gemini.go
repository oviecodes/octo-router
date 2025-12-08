package providers

import (
	"context"
	"fmt"
	"llm-router/types"
	"log"

	"github.com/pkoukk/tiktoken-go"
	"google.golang.org/genai"
)

type GeminiConfig struct {
	APIKey    string
	MaxTokens int64
	Model     string
}

type GeminiProvider struct {
	client    *genai.Client
	maxTokens int64
	model     string
}

func (g *GeminiProvider) Complete(ctx context.Context, messages []types.Message) types.Message {

	//convert to geminiMessageformat
	geminiMessages, currentMessage := g.convertMessages(messages)

	chat, err := g.client.Chats.Create(
		ctx,
		g.model,
		nil,
		geminiMessages,
	)

	if err != nil {
		log.Fatal(err)
	}

	res, err := chat.SendMessage(ctx, genai.Part{Text: currentMessage})

	if err != nil {
		log.Fatal(err)
	}

	if len(res.Candidates) > 0 {
		fmt.Println(res.Candidates[0].Content.Parts[0].Text)
	}

	//convert result to octo-router format

	return types.Message{}
}

func (g *GeminiProvider) convertMessages(messages []types.Message) ([]*genai.Content, string) {
	var geminiMessages []*genai.Content

	var lastMessage string

	for i, message := range messages {
		var toAppend *genai.Content

		if i == len(messages) {
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

	// for _, message := range latestMessageSlice {

	// }

	return geminiMessages, lastMessage
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

func NewGeminiProvider(config GeminiConfig) *GeminiProvider {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, nil)

	if err != nil {
		log.Fatal(err)
	}

	return &GeminiProvider{
		client:    client,
		maxTokens: config.MaxTokens,
		model:     config.Model,
	}
}
