package memory

import (
	"context"
	"github.com/MSaeed1381/message-broker/internal/model"
	"github.com/MSaeed1381/message-broker/internal/store"
	"github.com/MSaeed1381/message-broker/internal/utils"
	"sync"
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

func (m *MessageInMemory) Save(_ context.Context, message *model.Message) (uint64, error) {
	_, ok := m.messages.Load(message.BrokerMessage.Id)
	if ok {
		return 0, store.ErrMessageAlreadyExists{ID: uint64(message.BrokerMessage.Id)}
	}

	messageId := m.idGen.Next()
	message.BrokerMessage.Id = int(messageId) // set id to broker message

	m.messages.Store(messageId, message)
	return messageId, nil
}

func (m *MessageInMemory) GetByID(_ context.Context, id uint64) (*model.Message, error) {
	message, ok := m.messages.Load(id)
	if !ok {
		return &model.Message{}, store.ErrMessageNotFound{ID: id}
	}

	return message.(*model.Message), nil
}
