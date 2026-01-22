package router_test

import (
	"context"
	"fmt"
	"llm-router/cmd/internal/providers"
	"llm-router/cmd/internal/router"
	"llm-router/types"
	"testing"
)

type MockProvider struct {
	name string
}

func (p *MockProvider) Complete(ctx context.Context, input *types.CompletionInput) (*types.CompletionResponse, error) {
	return &types.CompletionResponse{}, nil
}
func (p *MockProvider) CompleteStream(ctx context.Context, input *types.StreamCompletionInput) (<-chan *types.StreamChunk, error) {
	return nil, nil
}
func (p *MockProvider) CountTokens(ctx context.Context, messages []types.Message) (int, error) {
	return 0, nil
}
func (p *MockProvider) GetProviderName() string {
	return p.name
}

func TestWeightedRouter_Distribution(t *testing.T) {
	provA := &MockProvider{name: "provider-a"}
	provB := &MockProvider{name: "provider-b"}
	provC := &MockProvider{name: "provider-c"} // Zero weight

	weights := map[string]int{
		"provider-a": 10,
		"provider-b": 90,
		"provider-c": 0,
	}

	factory := providers.NewProviderFactory()
	manager := providers.NewProviderManager(factory)
	manager.SetProviders([]types.Provider{provA, provB, provC})

	wr, err := router.NewWeightedRouter(manager, weights, nil, nil)
	if err != nil {
		t.Fatalf("Failed to create weighted router: %v", err)
	}

	iterations := 1000
	counts := make(map[string]int)

	fmt.Println("\n--- Starting Weighted Router Verification (1000 requests) ---")

	for i := 0; i < iterations; i++ {
		out, err := wr.SelectProvider(context.Background(), &types.SelectProviderInput{
			Circuits: make(map[string]types.CircuitBreaker),
		})
		if err != nil {
			t.Fatalf("Selection failed: %v", err)
		}
		counts[out.Provider.GetProviderName()]++
	}

	fmt.Printf("Counts: A=%d, B=%d, C=%d\n", counts["provider-a"], counts["provider-b"], counts["provider-c"])

	if counts["provider-c"] != 0 {
		t.Errorf("Provider C (weight 0) selected %d times", counts["provider-c"])
	}
	if counts["provider-a"] < 50 || counts["provider-a"] > 150 {
		t.Errorf("Provider A count %d out of expected range (~100)", counts["provider-a"])
	}

	if counts["provider-b"] < 850 || counts["provider-b"] > 950 {
		t.Errorf("Provider B count %d out of expected range (~900)", counts["provider-b"])
	}
}
