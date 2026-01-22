package router_test

import (
	"context"
	"fmt"
	"llm-router/cmd/internal/providers"
	"llm-router/cmd/internal/router"
	"llm-router/types"
	"testing"
)

type SemanticMockProvider struct {
	name string
}

func (m *SemanticMockProvider) GetProviderName() string {
	return m.name
}

func (m *SemanticMockProvider) Complete(ctx context.Context, input *types.CompletionInput) (*types.CompletionResponse, error) {
	return &types.CompletionResponse{
		Message: types.Message{Role: "assistant", Content: "Mock response from " + m.name},
	}, nil
}
func (m *SemanticMockProvider) CompleteStream(ctx context.Context, data *types.StreamCompletionInput) (<-chan *types.StreamChunk, error) {
	return nil, nil
}
func (m *SemanticMockProvider) CountTokens(ctx context.Context, messages []types.Message) (int, error) {
	return 0, nil
}

type MockCircuitBreaker struct {
	canExecute bool
}

func (m *MockCircuitBreaker) Execute(err error) {
	// No-op for mock, or could track failure
}

func (m *MockCircuitBreaker) CanExecute() bool {
	return m.canExecute
}

func (m *MockCircuitBreaker) GetState() string {
	if m.canExecute {
		return "closed"
	}
	return "open"
}

func TestSemanticPipeline_KeywordFilter(t *testing.T) {
	provOpenAI := &SemanticMockProvider{name: "openai"}       // Good at code
	provAnthropic := &SemanticMockProvider{name: "anthropic"} // Good at code
	provGemini := &SemanticMockProvider{name: "gemini"}       // Not allowed involved for code in this test

	factory := providers.NewProviderFactory()
	manager := providers.NewProviderManager(factory)
	manager.SetProviders([]types.Provider{provOpenAI, provAnthropic, provGemini})

	policy := &types.SemanticPolicy{
		Enabled:      true,
		DefaultGroup: "general",
		Groups: []types.SemanticGroup{
			{
				Name:           "coding",
				IntentKeywords: []string{"code", "golang", "python"},
				AllowProviders: []string{"openai", "anthropic"},
			},
		},
	}

	routingData := &types.RoutingData{
		Strategy: "round-robin",
		Policies: &types.Policies{
			Semantic: policy,
		},
	}

	r, _, err := router.ConfigureRouterStrategy(routingData, manager, nil, nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("Failed to configure router: %v", err)
	}

	ctx := context.Background()

	circuits := make(map[string]types.CircuitBreaker)
	circuits["openai"] = &MockCircuitBreaker{canExecute: true}
	circuits["anthropic"] = &MockCircuitBreaker{canExecute: true}
	circuits["gemini"] = &MockCircuitBreaker{canExecute: true}

	inputCode := &types.SelectProviderInput{
		Circuits: circuits,
		Messages: []types.Message{
			{Role: "user", Content: "Write some golang code"},
		},
	}

	for i := 0; i < 20; i++ {
		out, err := r.SelectProvider(ctx, inputCode)
		if err != nil {
			t.Fatalf("Selection failed: %v", err)
		}
		if out.Provider.GetProviderName() == "gemini" {
			t.Errorf("Semantic Filter Failed! Selected 'gemini' for coding task")
		}
	}
	fmt.Println("PASS: Coding intent correctly excluded Gemini")

	inputGeneral := &types.SelectProviderInput{
		Circuits: circuits,
		Messages: []types.Message{
			{Role: "user", Content: "Tell me a joke"},
		},
	}

	geminiSelected := false
	for i := 0; i < 50; i++ {
		out, err := r.SelectProvider(ctx, inputGeneral)
		if err != nil {
			t.Fatalf("Selection failed: %v", err)
		}
		if out.Provider.GetProviderName() == "gemini" {
			geminiSelected = true
			break
		}
	}

	if !geminiSelected {
		t.Errorf("General intent should allow Gemini, but it was never selected in 50 tries (Round Robin)")
	} else {
		fmt.Println("PASS: General intent allowed Gemini")
	}
}
