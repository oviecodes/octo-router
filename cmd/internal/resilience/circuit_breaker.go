package resilience

import (
	"sync"
	"time"
)

type Circuit struct {
	Mu           sync.Mutex
	State        string
	FailureCount int
	Threshold    int
	Timeout      time.Duration
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
	c.State = "CLOSED"
}

func (c *Circuit) scheduleHalfOpen() {
	time.Sleep(c.Timeout)
	c.Mu.Lock()
	defer c.Mu.Unlock()
	c.State = "HALF_OPEN"
}

func NewCircuitBreakers(providers []string, config map[string]int) map[string]*Circuit {

	failureThreshold := getCircuitPropsOrDefault(config, "failureThreshold", 5)
	resetTimeout := getCircuitPropsOrDefault(config, "resetTimeout", 60000)

	allCircuitBreakers := make(map[string]*Circuit)

	for _, provider := range providers {
		allCircuitBreakers[provider] = &Circuit{
			State:     "CLOSED",
			Threshold: failureThreshold,
			Timeout:   time.Duration(resetTimeout) * time.Millisecond,
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
