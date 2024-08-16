package model

import (
	"github.com/MSaeed1381/message-broker/pkg/broker"
	"sync"
	"time"
)

type Message struct {
	BrokerMessage *broker.Message
	Subject       string
	CreateAt      time.Time
}

func NewMessageModel(subject string, brokerMessage *broker.Message) *Message {
	return &Message{
		BrokerMessage: brokerMessage,
		Subject:       subject,
		CreateAt:      time.Now(),
	}
}

type Topic struct {
	ID      uint64
	Subject string
}

// Connection for store channel (this must be synchronized by mutex)
type Connection struct {
	ID        uint64
	Channel   chan broker.Message
	ChanMutex sync.Mutex
}

func NewTopicModel(subject string) *Topic {
	return &Topic{Subject: subject}
}

func NewConnectionModel(channel chan broker.Message) *Connection {
	return &Connection{Channel: channel}
}
