package memory

import (
	"context"
	"github.com/MSaeed1381/message-broker/internal/model"
	"github.com/MSaeed1381/message-broker/internal/store"
	"github.com/MSaeed1381/message-broker/internal/utils"
	"github.com/MSaeed1381/message-broker/pkg/broker"
	"sync"
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
	message.Id = int(messageId) // set id to broker message

	m.messages.Store(messageId, &model.Message{BrokerMessage: message, CreateAt: time.Now()})
	return messageId, nil
}

func (m *MessageInMemory) GetByID(ctx context.Context, id uint64) (*model.Message, error) {
	message, ok := m.messages.Load(id)
	if !ok {
		return &model.Message{}, store.ErrMessageNotFound{ID: id}
	}

	return message.(*model.Message), nil
}
