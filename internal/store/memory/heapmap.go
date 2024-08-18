package memory

import (
	"container/heap"
	"github.com/MSaeed1381/message-broker/internal/model"
	"sync"
	"time"
)

// An Item is something we manage in a priority queue.
type Item struct {
	msgId      uint64
	Expiration int64
}

// A PriorityQueue implements heap.Interface and holds Items.
type PriorityQueue []*Item

func (pq PriorityQueue) Len() int {
	return len(pq)
}

func (pq PriorityQueue) Less(i, j int) bool {
	return pq[i].Expiration < pq[j].Expiration
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

func (pq *PriorityQueue) Push(x any) {
	item := x.(*Item)
	*pq = append(*pq, item)
}

func (pq *PriorityQueue) Pop() any {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil // avoid memory leak
	*pq = old[0 : n-1]
	return item
}

// HeapMap wraps sync.Map with TTL support.
type HeapMap struct {
	messages        sync.Map
	pq              PriorityQueue
	pqMutex         sync.Mutex
	cleanupInterval time.Duration
}

// NewHeapMap creates a new SyncMap With TTL.
func NewHeapMap(cleanupInterval time.Duration) *HeapMap {
	smt := &HeapMap{
		pq:              PriorityQueue{},
		cleanupInterval: cleanupInterval,
	}
	heap.Init(&smt.pq)
	go smt.startCleanupRoutine()
	return smt
}

// Set adds a key-value pair to the map with a specified TTL.
func (m *HeapMap) Set(msgId uint64, msg *model.Message, ttl time.Duration) {
	expiration := time.Now().Add(ttl).UnixNano()

	item := &Item{
		msgId:      msgId,
		Expiration: expiration,
	}

	m.messages.Store(msgId, msg) // store message
	m.pqMutex.Lock()
	heap.Push(&m.pq, item)
	m.pqMutex.Unlock()
}

// Get retrieves a value by key. Returns the value and a boolean indicating if the key was found and is still valid.
func (m *HeapMap) Get(msgId uint64) (*model.Message, bool) {
	item, ok := m.messages.Load(msgId)
	if !ok {
		return nil, false
	}

	// item expired
	if item == nil {
		return nil, true
	}

	msg := item.(*model.Message)
	return msg, true
}

// Cleanup removes expired items from the map.
func (m *HeapMap) Cleanup() {
	now := time.Now().UnixNano()

	m.pqMutex.Lock() // Lock the mutex for safe access to the priority queue
	defer m.pqMutex.Unlock()

	for m.pq.Len() > 0 {
		head := m.pq[0]
		if now <= head.Expiration {
			break
		}
		heap.Pop(&m.pq)
		m.messages.Store(head.msgId, nil) // we behave like black-list
	}
}

// startCleanupRoutine starts a background goroutine that periodically cleans up expired items.
func (m *HeapMap) startCleanupRoutine() {
	ticker := time.NewTicker(m.cleanupInterval)

	for {
		select {
		case <-ticker.C:
			m.Cleanup()
		}
	}
}
