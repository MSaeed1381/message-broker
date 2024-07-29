package memory

import (
	"context"
	"sync"
	"therealbroker/internal/model"
	"therealbroker/internal/store"
	"therealbroker/internal/utils"
	"therealbroker/pkg/broker"
	"time"
)

type MessageInMemory struct {
	messages sync.Map
	idGen    utils.IdGenerator
}

func NewMessageInMemory() *MessageInMemory {
	return &MessageInMemory{
		messages: sync.Map{},
		idGen:    utils.IdGenerator{},
	}
}

func (m *MessageInMemory) Save(ctx context.Context, message *broker.Message) (uint64, error) {
	_, ok := m.messages.Load(message.Id)
	if ok {
		return 0, store.ErrMessageAlreadyExists{ID: uint64(message.Id)}
	}

	messageId := m.idGen.Next()

	m.messages.Store(messageId, model.Message{BrokerMessage: message, CreateAt: time.Now()})
	return messageId, nil
}

func (m *MessageInMemory) GetByID(ctx context.Context, id int) (*model.Message, error) {
	message, ok := m.messages.Load(id)
	if !ok {
		return message.(*model.Message), store.ErrMessageNotFound{ID: uint64(id)}
	}

	return message.(*model.Message), nil
}

func (m *MessageInMemory) GetAll(ctx context.Context) ([]model.Message, error) {
	messages := make([]model.Message, 0)
	for _, msg := range messages {
		messages = append(messages, msg)
	}

	return messages, nil
}
