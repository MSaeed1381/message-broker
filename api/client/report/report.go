package report

import (
	"fmt"
	"sync"
)

type Counter struct {
	mu      sync.Mutex
	counter int
}

func (c *Counter) Inc() {
	c.mu.Lock()
	c.counter++
	c.mu.Unlock()
}

func (c *Counter) Value() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.counter
}

type Report struct {
	FetchFailure   Counter
	FetchSuccess   Counter
	PublishFailure Counter
	PublishSuccess Counter
	Dropped        Counter
}

func NewReport() *Report {
	return &Report{}
}

func (r *Report) String() string {
	return fmt.Sprintf("{Fetch Failure: %d, Success: %d, Publish Failure: %d, Publish Success: %d, Dropped: %d}",
		r.FetchFailure.Value(), r.FetchSuccess.Value(), r.PublishFailure.Value(), r.PublishSuccess.Value(), r.Dropped.Value())
}
