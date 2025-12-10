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

func (m *mockProvider) Complete(ctx context.Context, messages []types.Message) (*types.Message, error) {
	return &types.Message{
		Role:    "assistant",
		Content: "test response",
	}, nil
}

func (m *mockProvider) CompleteStream(ctx context.Context, messages []types.Message) (<-chan *types.StreamChunk, error) {
	return nil, nil
}

func (m *mockProvider) CountTokens(ctx context.Context, messages []types.Message) (int, error) {
	return 100, nil
}

func TestRoundRobinSelection(t *testing.T) {
	// Create mock providers
	allProviders := []providers.Provider{
		&mockProvider{name: "provider1"},
		&mockProvider{name: "provider2"},
		&mockProvider{name: "provider3"},
	}

	router := &RoundRobinRouter{
		providers: allProviders,
		current:   0,
	}

	// Test round-robin distribution
	selectedProviders := make([]providers.Provider, 6)

	for i := range 6 {
		selectedProviders[i] = router.SelectProvider(context.Background())
	}

	// Should cycle through providers
	if selectedProviders[0] != allProviders[0] {
		t.Error("First selection should be provider 0")
	}
	if selectedProviders[1] != allProviders[1] {
		t.Error("Second selection should be provider 1")
	}
	if selectedProviders[2] != allProviders[2] {
		t.Error("Third selection should be provider 2")
	}
	if selectedProviders[3] != allProviders[0] {
		t.Error("Fourth selection should wrap to provider 0")
	}
}

func TestRoundRobinWithSingleProvider(t *testing.T) {
	providers := []providers.Provider{
		&mockProvider{name: "solo"},
	}

	router := &RoundRobinRouter{
		providers: providers,
		current:   0,
	}

	// Should always return the same provider
	for i := range 5 {
		selected := router.SelectProvider(context.Background())
		if selected != providers[0] {
			t.Errorf("Expected solo provider, got different provider on iteration %d", i)
		}
	}
}
