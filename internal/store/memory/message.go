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
		return 0, store.ErrMessageAlreadyExists{ID: message.BrokerMessage.Id}
	}

	message.BrokerMessage.Id = m.idGen.Next() // set id to broker message

	m.messages.Store(message.BrokerMessage.Id, message)
	return message.BrokerMessage.Id, nil
}

func (m *MessageInMemory) GetByID(_ context.Context, id uint64) (*model.Message, error) {
	message, ok := m.messages.Load(id)
	if !ok {
		return &model.Message{}, store.ErrMessageNotFound{ID: id}
	}

	return message.(*model.Message), nil
}
