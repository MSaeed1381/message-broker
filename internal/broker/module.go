package broker

import (
	"context"
	"sync"
	"therealbroker/internal/model"
	"therealbroker/internal/store"
	"therealbroker/internal/store/memory"
	"therealbroker/pkg/broker"
	"time"
)

type Module struct {
	Topics   store.Topic
	Messages store.Message
	closed   bool
}

func NewModule() broker.Broker {
	return &Module{ // TODO change in memory to general form
		Topics:   memory.NewTopicInMemory(),
		Messages: memory.NewMessageInMemory(),
		closed:   false,
	}
}

func (m *Module) Close() error {
	m.closed = true
	return nil
}

func (m *Module) Publish(ctx context.Context, subject string, msg broker.Message) (int, error) {
	if m.closed {
		return 0, broker.ErrUnavailable
	}

	topic, err := m.Topics.GetBySubject(ctx, subject)

	// TODO AS function to check error
	if err != nil {
		topic = &model.Topic{
			Subject:     subject,
			Messages:    sync.Map{},
			Connections: make([]*model.Connection, 10),
		}

		err := m.Topics.Save(ctx, topic)
		if err != nil {
			return 0, err
		}
	}

	messageId, err := m.Topics.SaveMessage(ctx, subject, &msg)
	if err != nil {
		return 0, err
	}

	connections, err := m.Topics.GetOpenConnections(ctx, subject)
	if err != nil {
		return 0, err
	}

	// TODO for until connection saved
	for _, connection := range connections {
		if connection != nil && connection.Channel != nil {
			*connection.Channel <- msg
		}
	}

	return int(messageId), nil
}

func (m *Module) Subscribe(ctx context.Context, subject string) (<-chan broker.Message, error) {
	if m.closed {
		return nil, broker.ErrUnavailable
	}

	topic, err := m.Topics.GetBySubject(ctx, subject)
	if err != nil {
		topic = &model.Topic{
			Subject:     subject,
			Messages:    sync.Map{},
			Connections: make([]*model.Connection, 0),
		}

		err := m.Topics.Save(ctx, topic)
		if err != nil {
			return nil, err
		}
	}

	// TODO put size as constant in config file
	result := make(chan broker.Message, 10000)

	err = m.Topics.SaveConnection(
		ctx,
		subject,
		&model.Connection{
			Mu:      sync.Mutex{},
			Channel: &result,
		})

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (m *Module) Fetch(ctx context.Context, subject string, id int) (broker.Message, error) {
	if m.closed {
		return broker.Message{}, broker.ErrUnavailable
	}

	msg, err := m.Topics.GetMessage(ctx, uint64(id), subject)
	if err != nil {
		return broker.Message{}, broker.ErrInvalidID
	}

	// handle expiration time
	if time.Now().Sub(msg.CreateAt) > msg.BrokerMessage.Expiration {
		return broker.Message{}, broker.ErrExpiredID
	}

	return *msg.BrokerMessage, nil
}
