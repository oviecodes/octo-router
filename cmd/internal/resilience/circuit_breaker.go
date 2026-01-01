package resilience

import (
	"llm-router/cmd/internal/metrics"
	"llm-router/types"
	"sync"
	"time"
)

type Circuit struct {
	Mu           sync.Mutex
	State        string
	FailureCount int
	Threshold    int
	Timeout      time.Duration
	Provider     string
}

func (c *Circuit) GetState() string {
	return c.State
}

func (c *Circuit) CanExecute() bool {
	c.Mu.Lock()
	defer c.Mu.Unlock()

	if c.State == "CLOSED" {
		return true
	}

	if c.State == "HALF_OPEN" {
		return true
	}

	return false
}

func (c *Circuit) Execute(err error) {
	c.Mu.Lock()
	defer c.Mu.Unlock()

	if c.State == "OPEN" {
		return
	}

	if err != nil {
		c.FailureCount++
		if c.FailureCount >= c.Threshold {
			metrics.CircuitBreakerState.WithLabelValues(c.Provider).Set(1)
			metrics.CircuitBreakerTrips.WithLabelValues(c.Provider).Inc()

			c.State = "OPEN"
			go c.scheduleHalfOpen()
		}
	} else {
		if c.State == "HALF_OPEN" {
			c.resetState()
		} else {
			c.FailureCount = 0
		}
	}

}

func (c *Circuit) resetState() {
	c.FailureCount = 0
	metrics.CircuitBreakerState.WithLabelValues(c.Provider).Set(0)
	c.State = "CLOSED"
}

func (c *Circuit) scheduleHalfOpen() {
	time.Sleep(c.Timeout)
	c.Mu.Lock()
	defer c.Mu.Unlock()
	metrics.CircuitBreakerState.WithLabelValues(c.Provider).Set(2)
	c.State = "HALF_OPEN"
}

func NewCircuitBreakers(providers []string, config map[string]int) map[string]types.CircuitBreaker {

	failureThreshold := getCircuitPropsOrDefault(config, "failureThreshold", 5)
	resetTimeout := getCircuitPropsOrDefault(config, "resetTimeout", 60000)

	allCircuitBreakers := make(map[string]types.CircuitBreaker)

	for _, provider := range providers {
		allCircuitBreakers[provider] = &Circuit{
			State:     "CLOSED",
			Threshold: failureThreshold,
			Timeout:   time.Duration(resetTimeout) * time.Millisecond,
			Provider:  provider,
		}
	}

	return allCircuitBreakers
}

func getCircuitPropsOrDefault(config map[string]int, name string, defaultVal int) int {
	if val, ok := config[name]; ok {
		return val
	}

	return defaultVal
}
