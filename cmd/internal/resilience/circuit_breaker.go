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
	time.Sleep(60 * time.Second)
	c.Mu.Lock()
	defer c.Mu.Unlock()
	c.State = "HALF_OPEN"
}
