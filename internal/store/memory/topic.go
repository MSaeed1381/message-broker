package memory

import (
	"context"
	"github.com/MSaeed1381/message-broker/internal/model"
	"github.com/MSaeed1381/message-broker/internal/store"
	"github.com/MSaeed1381/message-broker/internal/utils"
	"github.com/MSaeed1381/message-broker/pkg/broker"
	"sync"
)

// TopicMemoryWrapper wrap one model.Topic instance
type TopicMemoryWrapper struct {
	Message    store.Message // save messages in it
	Connection *ConnectionInMemory
	topic      *model.Topic
}

type TopicInMemory struct {
	topics   sync.Map // list of all memory.topics (saves TopicMemoryWrapper in it)
	MsgIdGen *utils.IdGenerator
	MsgStore store.Message // nil when storage is in-memory
}

func NewTopicInMemory(msgStore store.Message) *TopicInMemory {
	return &TopicInMemory{
		topics:   sync.Map{},
		MsgIdGen: utils.NewIdGenerator(),
		MsgStore: msgStore,
	}
}

func (t *TopicInMemory) Save(ctx context.Context, topic *model.Topic) error {
	topicWrapper := &TopicMemoryWrapper{
		Message:    t.MsgStore,
		Connection: NewConnectionInMemory(),
		topic:      topic,
	}

	// if msgStore is nil then means that we use message in memory (default)
	if topicWrapper.Message == nil {
		topicWrapper.Message = NewMessageInMemory()
	}

	t.topics.Store(topic.Subject, topicWrapper)
	return nil
}

func (t *TopicInMemory) GetBySubject(ctx context.Context, subject string) (*model.Topic, error) {
	tw, ok := t.topics.Load(subject)
	if !ok {
		return &model.Topic{}, store.ErrTopicNotFound{Subject: subject}
	}
	return tw.(*TopicMemoryWrapper).topic, nil
}

func (t *TopicInMemory) GetOpenConnections(ctx context.Context, subject string) ([]*model.Connection, error) {
	tw, ok := t.topics.Load(subject)
	if !ok {
		return make([]*model.Connection, 0), store.ErrTopicNotFound{Subject: subject}
	}

	connections, err := tw.(*TopicMemoryWrapper).Connection.GetAllConnections(ctx)
	if err != nil {
		return nil, err
	}

	return connections, nil
}

func (t *TopicInMemory) SaveMessage(ctx context.Context, subject string, message *broker.Message) (uint64, error) {
	tw, ok := t.topics.Load(subject)
	if !ok {
		return 0, store.ErrTopicNotFound{Subject: subject}
	}

	msgId, err := tw.(*TopicMemoryWrapper).Message.Save(ctx, message, subject)
	if err != nil {
		return 0, err
	}

	return msgId, nil
}

func (t *TopicInMemory) GetMessage(ctx context.Context, messageId uint64, subject string) (*model.Message, error) {
	tw, ok := t.topics.Load(subject)
	if !ok {
		return nil, store.ErrTopicNotFound{Subject: subject}
	}

	msg, err := tw.(*TopicMemoryWrapper).Message.GetByID(ctx, messageId)
	if err != nil {
		return nil, err
	}

	return msg, nil
}

func (t *TopicInMemory) SaveConnection(ctx context.Context, subject string, connection *model.Connection) error {
	tw, ok := t.topics.Load(subject)
	if !ok {
		return store.ErrTopicNotFound{Subject: subject}
	}

	if err := tw.(*TopicMemoryWrapper).Connection.Save(ctx, connection); err != nil {
		return err
	}
	return nil
}
