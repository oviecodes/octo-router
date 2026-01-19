package router

import (
	"context"
	"llm-router/types"
	"time"
)

type LatencyMonitoringProvider struct {
	types.Provider
	tracker *LatencyTracker
}

func NewLatencyMonitoringProvider(provider types.Provider, tracker *LatencyTracker) *LatencyMonitoringProvider {
	return &LatencyMonitoringProvider{
		Provider: provider,
		tracker:  tracker,
	}
}

func (p *LatencyMonitoringProvider) Complete(ctx context.Context, input *types.CompletionInput) (*types.CompletionResponse, error) {
	start := time.Now()
	resp, err := p.Provider.Complete(ctx, input)
	duration := time.Since(start).Seconds() * 1000

	if err == nil {
		p.tracker.RecordLatency(p.Provider.GetProviderName(), duration)
	}

	return resp, err
}

// CompleteStream wraps the streaming completion to measure time to first byte purely?
// Or full duration? For routing, "Time to First Token" (TTFT) is usually most important for perception,
// but total throughput matters too. For simplicity, let's track the start of the stream call.
// Tracking full stream duration requires wrapping the channel which is complex.
// For now, we will track the API call overhead.
func (p *LatencyMonitoringProvider) CompleteStream(ctx context.Context, input *types.StreamCompletionInput) (<-chan *types.StreamChunk, error) {
	start := time.Now()
	stream, err := p.Provider.CompleteStream(ctx, input)
	duration := time.Since(start).Seconds() * 1000

	if err == nil {
		p.tracker.RecordLatency(p.Provider.GetProviderName(), duration)
	}

	return stream, err
}
