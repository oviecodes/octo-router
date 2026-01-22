package router_test

import (
	"context"
	"fmt"
	"llm-router/cmd/internal/providers"
	"llm-router/cmd/internal/router"
	"llm-router/types"
	"math/rand"
	"testing"
	"time"
)

type SimulatedProvider struct {
	name        string
	baseLatency time.Duration
}

func (p *SimulatedProvider) Complete(ctx context.Context, input *types.CompletionInput) (*types.CompletionResponse, error) {
	jitter := time.Duration(rand.Intn(20)) * time.Millisecond
	time.Sleep(p.baseLatency + jitter)
	return &types.CompletionResponse{}, nil
}

func (p *SimulatedProvider) CompleteStream(ctx context.Context, input *types.StreamCompletionInput) (<-chan *types.StreamChunk, error) {
	return nil, nil
}

func (p *SimulatedProvider) CountTokens(ctx context.Context, messages []types.Message) (int, error) {
	return 0, nil
}

func (p *SimulatedProvider) GetProviderName() string {
	return p.name
}

func TestLatencyRouter_EndToEnd_Convergence(t *testing.T) {
	tracker := router.NewLatencyTracker()

	fastProv := &SimulatedProvider{name: "fast-ai", baseLatency: 20 * time.Millisecond}
	medProv := &SimulatedProvider{name: "medium-ai", baseLatency: 100 * time.Millisecond}
	slowProv := &SimulatedProvider{name: "slow-ai", baseLatency: 300 * time.Millisecond}

	rawProviders := []types.Provider{fastProv, medProv, slowProv}

	var wrappedProviders []types.Provider
	for _, p := range rawProviders {
		wrappedProviders = append(wrappedProviders, router.NewLatencyMonitoringProvider(p, tracker))
	}

	factory := providers.NewProviderFactory()
	manager := providers.NewProviderManager(factory)
	manager.SetProviders(wrappedProviders)

	lr, err := router.NewLatencyRouter(manager, tracker, nil, nil, nil)
	if err != nil {
		t.Fatalf("Failed to create router: %v", err)
	}

	fmt.Println("\n--- Starting Latency Routing Verification ---")

	selections := make(map[string]int)

	for i := 0; i < 20; i++ {
		out, err := lr.SelectProvider(context.Background(), &types.SelectProviderInput{
			Circuits: make(map[string]types.CircuitBreaker),
		})
		if err != nil {
			t.Fatalf("Selection failed: %v", err)
		}

		providerName := out.Provider.GetProviderName()
		selections[providerName]++

		_, _ = out.Provider.Complete(context.Background(), nil)

		score := tracker.GetLatencyScore(providerName)
		fmt.Printf("Iter %02d: Selected [%s] (Latency Score: %.2fms)\n", i+1, providerName, score)
	}

	fmt.Println("--- Summary ---")
	for name, count := range selections {
		fmt.Printf("%s: selected %d times\n", name, count)
	}

	if selections["fast-ai"] <= selections["slow-ai"] {
		t.Errorf("Router failed to favor fast provider! Stats: %v", selections)
	}
}
