package memory

import (
	"context"
	"github.com/MSaeed1381/message-broker/internal/model"
	"github.com/MSaeed1381/message-broker/internal/store"
	"github.com/MSaeed1381/message-broker/internal/utils"
	"github.com/MSaeed1381/message-broker/pkg/broker"
	"sync"
)

// TopicMemoryWrapper wrap model.Topic instance
type TopicMemoryWrapper struct {
	Message    *MessageInMemory // save messages in it
	Connection *ConnectionInMemory
	topic      *model.Topic
}

type TopicInMemory struct {
	topics   sync.Map // list of all memory.topics
	messages sync.Map //

	MsgIdGen *utils.IdGenerator
	mu       *sync.Mutex
}

func NewTopicInMemory() *TopicInMemory {
	return &TopicInMemory{
		topics:   sync.Map{},
		MsgIdGen: utils.NewIdGenerator(),
		mu:       &sync.Mutex{},
	}
}

func (t *TopicInMemory) Save(ctx context.Context, topic *model.Topic) error {
	//_, ok := t.topics.Load(topic.Topic)
	//if ok {
	//	return store.ErrTopicAlreadyExists{Topic: topic.Topic}
	//}

	t.topics.Store(topic.Subject, TopicMemoryWrapper{
		Message:    &MessageInMemory{},
		Connection: &ConnectionInMemory{},
		topic:      topic,
	})
	return nil
}

func (t *TopicInMemory) GetBySubject(ctx context.Context, subject string) (*model.Topic, error) {
	tw, ok := t.topics.Load(subject)
	if !ok {
		return &model.Topic{}, store.ErrTopicNotFound{Subject: subject}
	}
	return tw.(TopicMemoryWrapper).topic, nil
}

func (t *TopicInMemory) GetOpenConnections(ctx context.Context, subject string) ([]*model.Connection, error) {
	tw, ok := t.topics.Load(subject)
	if !ok {
		return make([]*model.Connection, 0), store.ErrTopicNotFound{Subject: subject}
	}

	connections, err := tw.(TopicMemoryWrapper).Connection.GetAllConnections(ctx)
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

	msgId, err := tw.(TopicMemoryWrapper).Message.Save(ctx, message)
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

	msg, err := tw.(TopicMemoryWrapper).Message.GetByID(ctx, messageId)
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

	if err := tw.(TopicMemoryWrapper).Connection.Save(ctx, connection); err != nil {
		return err
	}
	return nil
}
