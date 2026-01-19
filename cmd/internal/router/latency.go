package router

import (
	"sync"
)

const (
	DefaultEMAAlpha = 0.2
)

type LatencyTracker struct {
	mu     sync.RWMutex
	alpha  float64
	scores map[string]float64
}

func NewLatencyTracker() *LatencyTracker {
	return &LatencyTracker{
		alpha:  DefaultEMAAlpha,
		scores: make(map[string]float64),
	}
}

func (lt *LatencyTracker) RecordLatency(provider string, latencyMs float64) {
	lt.mu.Lock()
	defer lt.mu.Unlock()

	currentScore, exists := lt.scores[provider]
	if !exists {
		lt.scores[provider] = latencyMs
		return
	}

	// Exponential Moving Average
	// NewAverage = (Alpha * NewValue) + ((1 - Alpha) * OldAverage)
	newScore := (lt.alpha * latencyMs) + ((1 - lt.alpha) * currentScore)
	lt.scores[provider] = newScore
}

func (lt *LatencyTracker) GetLatencyScore(provider string) float64 {
	lt.mu.RLock()
	defer lt.mu.RUnlock()

	score, exists := lt.scores[provider]
	if !exists {
		return 0
	}
	return score
}
