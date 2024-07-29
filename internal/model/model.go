package model

import (
	"github.com/MSaeed1381/message-broker/pkg/broker"
	"time"
)

type Message struct {
	BrokerMessage *broker.Message
	CreateAt      time.Time
}

type Topic struct {
	ID      uint64
	Subject string
}

// Connection for store channel (this must be synchronized by mutex)
type Connection struct {
	ID      uint64
	Channel *chan broker.Message
}
