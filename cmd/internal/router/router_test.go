package router

import (
	"context"
	"llm-router/cmd/internal/providers"
	"llm-router/types"
	"testing"
)

// Mock provider for testing
type mockProvider struct {
	name string
}

func (m *mockProvider) Complete(ctx context.Context, input *types.CompletionInput) (*types.CompletionResponse, error) {
	return &types.CompletionResponse{
		Message: types.Message{Role: "assistant",
			Content: "test response"},
	}, nil
}

func (m *mockProvider) CompleteStream(ctx context.Context, data *types.StreamCompletionInput) (<-chan *types.StreamChunk, error) {
	return nil, nil
}

func (m *mockProvider) CountTokens(ctx context.Context, messages []types.Message) (int, error) {
	return 100, nil
}

func (m *mockProvider) GetProviderName() string {
	return "mock-ai"
}

func TestRoundRobinSelection(t *testing.T) {
	// Create mock providers
	allProviders := []types.Provider{
		&mockProvider{name: "provider1"},
		&mockProvider{name: "provider2"},
		&mockProvider{name: "provider3"},
	}

	// Create manager and inject test providers
	factory := providers.NewProviderFactory()
	manager := providers.NewProviderManager(factory)
	manager.SetProvidersForTest(allProviders)

	router := &RoundRobinRouter{
		providerManager: manager,
		current:         0,
	}

	// Create circuit breakers that allow all mock providers
	circuitBreakers := map[string]types.CircuitBreaker{
		"mock-ai": &mockCircuitBreaker{canExecute: true},
	}

	// Test round-robin distribution
	selectedProviders := make([]types.SelectedProviderOutput, 6)

	for i := range 6 {
		providerOutput, _ := router.SelectProvider(context.Background(), &types.SelectProviderInput{
			Circuits: circuitBreakers,
		})

		selectedProviders[i] = *providerOutput
	}

	// Should cycle through providers
	if selectedProviders[0].Provider != allProviders[0] {
		t.Error("First selection should be provider 0")
	}
	if selectedProviders[1].Provider != allProviders[1] {
		t.Error("Second selection should be provider 1")
	}
	if selectedProviders[2].Provider != allProviders[2] {
		t.Error("Third selection should be provider 2")
	}
	if selectedProviders[3].Provider != allProviders[0] {
		t.Error("Fourth selection should wrap to provider 0")
	}
}

func TestRoundRobinWithSingleProvider(t *testing.T) {
	mockProviders := []types.Provider{
		&mockProvider{name: "solo"},
	}

	// Create manager and inject test providers
	factory := providers.NewProviderFactory()
	manager := providers.NewProviderManager(factory)
	manager.SetProvidersForTest(mockProviders)

	router := &RoundRobinRouter{
		providerManager: manager,
		current:         0,
	}

	// Create circuit breaker that allows the mock provider
	circuitBreakers := map[string]types.CircuitBreaker{
		"mock-ai": &mockCircuitBreaker{canExecute: true},
	}

	// Should always return the same provider
	for i := range 5 {
		selected, _ := router.SelectProvider(context.Background(), &types.SelectProviderInput{
			Circuits: circuitBreakers,
		})
		if selected.Provider != mockProviders[0] {
			t.Errorf("Expected solo provider, got different provider on iteration %d", i)
		}
	}
}

// Mock circuit breaker for testing
type mockCircuitBreaker struct {
	canExecute bool
}

func (m *mockCircuitBreaker) CanExecute() bool {
	return m.canExecute
}

func (m *mockCircuitBreaker) Execute(err error) {
	// Mock implementation - does nothing
}

func (m *mockCircuitBreaker) GetState() string {
	if m.canExecute {
		return "closed"
	}
	return "open"
}
