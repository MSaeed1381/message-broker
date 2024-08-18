package memory

import (
	"context"
	"github.com/MSaeed1381/message-broker/internal/model"
	"github.com/MSaeed1381/message-broker/internal/store"
	"github.com/MSaeed1381/message-broker/internal/utils"
	"time"
)

type MessageInMemory struct {
	MsgStore *HeapMap
	idGen    utils.IdGenerator
}

func NewMessageInMemory() *MessageInMemory {
	return &MessageInMemory{
		MsgStore: NewHeapMap(time.Second * 2),
		idGen:    utils.IdGenerator{},
	}
}

func (m *MessageInMemory) Save(_ context.Context, message *model.Message) (uint64, error) {
	if _, ok := m.MsgStore.Get(message.BrokerMessage.Id); ok {
		return 0, store.ErrMessageAlreadyExists{ID: message.BrokerMessage.Id}
	}

	newId := m.idGen.Next()
	message.BrokerMessage.Id = newId // set id to broker message
	m.MsgStore.Set(newId, message, message.BrokerMessage.Expiration)

	return message.BrokerMessage.Id, nil
}

func (m *MessageInMemory) GetByID(_ context.Context, id uint64) (*model.Message, error) {
	message, ok := m.MsgStore.Get(id)

	// message is invalid and didn't publish
	if !ok {
		return &model.Message{}, store.ErrMessageNotFound{ID: id}
	} else if message == nil {
		return &model.Message{}, store.ErrMessageExpired{ID: id}
	}

	return message, nil
}

func (m *MessageInMemory) Close() {}
