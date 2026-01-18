package router

import "sync"

type LatencyTracker struct {
	mu                sync.Mutex
	slidingWindowSize int
	latencies         map[string][]float64
}

func NewLatencyTracker(slidingWindowSize int) *LatencyTracker {
	return &LatencyTracker{
		slidingWindowSize: slidingWindowSize,
		latencies:         make(map[string][]float64),
	}
}

func (lt *LatencyTracker) RecordLatency(provider string, latency float64) {
	lt.mu.Lock()
	defer lt.mu.Unlock()

	lt.latencies[provider] = append(lt.latencies[provider], latency)

	if len(lt.latencies[provider]) > lt.slidingWindowSize {
		lt.latencies[provider] = lt.latencies[provider][1:]
	}
}

func (lt *LatencyTracker) GetLatency(provider string) float64 {
	lt.mu.Lock()
	defer lt.mu.Unlock()

	return lt.latencies[provider][len(lt.latencies[provider])-1]
}

func (lt *LatencyTracker) GetAverageLatency(provider string) float64 {
	lt.mu.Lock()
	defer lt.mu.Unlock()

	var sum float64
	for _, latency := range lt.latencies[provider] {
		sum += latency
	}
	return sum / float64(len(lt.latencies[provider]))
}
