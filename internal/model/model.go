package model

import (
	"github.com/MSaeed1381/message-broker/pkg/broker"
	"time"
)

type Message struct {
	BrokerMessage *broker.Message
	Subject       string
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

func NewTopicModel(subject string) *Topic {
	return &Topic{Subject: subject}
}

func NewConnectionModel(channel *chan broker.Message) *Connection {
	return &Connection{Channel: channel}
}
