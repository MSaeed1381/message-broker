package utils

import (
	"sync"
)

type IdGenerator struct {
	counter uint64
	mutex   sync.Mutex
}

func NewIdGenerator() *IdGenerator {
	return &IdGenerator{counter: 0}
}

func (g *IdGenerator) Next() uint64 {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	g.counter++
	return g.counter
}
