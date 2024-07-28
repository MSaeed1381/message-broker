package memory

import (
	"fmt"
	"sync"
	"therealbroker/internal/model"
	"therealbroker/internal/store"
	"therealbroker/internal/utils"
	"therealbroker/pkg/broker"
	"time"
)

type TopicInMemory struct {
	topics   sync.Map
	MsgIdGen *utils.IdGenerator
}

func NewTopicInMemory() *TopicInMemory {
	return &TopicInMemory{
		topics:   sync.Map{},
		MsgIdGen: utils.NewIdGenerator(),
	}
}

func (t *TopicInMemory) Save(topic *model.Topic) error {
	//_, ok := t.topics.Load(topic.Subject)
	//if ok {
	//	return store.ErrTopicAlreadyExists{Subject: topic.Subject}
	//}

	t.topics.Store(topic.Subject, topic)
	return nil
}

func (t *TopicInMemory) GetBySubject(subject string) (*model.Topic, error) {
	topic, ok := t.topics.Load(subject)
	if !ok {
		return &model.Topic{}, store.ErrTopicNotFound{Subject: subject}
	}
	return topic.(*model.Topic), nil
}

func (t *TopicInMemory) GetOpenConnections(subject string) ([]*model.Connection, error) {
	topic, err := t.GetBySubject(subject)
	if err != nil {
		return nil, err
	}

	return topic.Connections, nil
}

func (t *TopicInMemory) SaveMessage(subject string, message *broker.Message) (uint64, error) {
	topic, err := t.GetBySubject(subject)
	if err != nil {
		return 0, err
	}

	msgId := t.MsgIdGen.Next()

	topic.Messages.Store(msgId, &model.Message{BrokerMessage: message, CreateAt: time.Now()})
	return msgId, nil
}

func (t *TopicInMemory) GetMessage(messageId uint64, subject string) (*model.Message, error) {
	topic, err := t.GetBySubject(subject)
	if err != nil {
		return &model.Message{}, err
	}

	msg, ok := topic.Messages.Load(messageId)
	if !ok {
		return &model.Message{}, err
	}

	fmt.Println(msg.(*model.Message))
	return msg.(*model.Message), nil
}

func (t *TopicInMemory) SaveConnection(subject string, connection *model.Connection) error {
	topic, err := t.GetBySubject(subject)
	if err != nil {
		return err
	}
	topic.Connections = append(topic.Connections, connection)
	return nil
}