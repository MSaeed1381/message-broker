package model

import (
	"github.com/MSaeed1381/message-broker/pkg/broker"
	"sync"
	"time"
)

type Message struct {
	BrokerMessage *broker.Message
	CreateAt      time.Time `json:"create_at"`
}

type Topic struct {
	ID          uint64        `json:"id"`
	Subject     string        `json:"subject"`
	Messages    sync.Map      `json:"messages"` // one-to-many relation - messageInMemory
	Mu          sync.Mutex    // to synchronize Connections
	Connections []*Connection `json:"connections"`
}

// Connection for store channel (this must be synchronized by mutex)
type Connection struct {
	ID      uint64 `json:"id"`
	Mu      sync.Mutex
	Channel *chan broker.Message `json:"channels"`
}
