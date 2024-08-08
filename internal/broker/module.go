package broker

import (
	"context"
	"errors"
	"github.com/MSaeed1381/message-broker/internal/model"
	"github.com/MSaeed1381/message-broker/internal/store"
	"github.com/MSaeed1381/message-broker/pkg/broker"
	"time"
)

type Module struct {
	Topics store.Topic // have msg Store and connection Store
	Closed bool
}

func NewModule(topic store.Topic) broker.Broker {
	return &Module{
		Topics: topic,
		Closed: false,
	}
}

// Close TODO implement with Channel
func (m *Module) Close() error {
	m.Closed = true
	return nil
}

func (m *Module) Publish(ctx context.Context, subject string, msg broker.Message) (int, error) {
	if m.Closed {
		return 0, broker.ErrUnavailable
	}

	topic, err := m.Topics.GetBySubject(ctx, subject)

	// TODO synchronize this
	if errors.As(err, &store.ErrTopicNotFound{}) {
		// create the new model and saves to data store
		topic = model.NewTopicModel(subject)

		err := m.Topics.Save(ctx, topic)
		if err != nil {
			return 0, err
		}
	}

	msgId, err := m.Topics.SaveMessage(ctx, subject, &msg)
	if err != nil {
		return 0, err
	}

	connections, err := m.Topics.GetOpenConnections(ctx, subject)
	if err != nil {
		return 0, err
	}

	// TODO for until connection saved (concurrent go)
	for _, connection := range connections {
		func(c *model.Connection, msg *broker.Message) {
			if c != nil && c.Channel != nil {
				*c.Channel <- *msg
			}
		}(connection, &msg)
	}

	return int(msgId), nil
}

func (m *Module) Subscribe(ctx context.Context, subject string) (<-chan broker.Message, error) {
	if m.Closed {
		return nil, broker.ErrUnavailable
	}

	topic, err := m.Topics.GetBySubject(ctx, subject)
	if err != nil {
		topic = model.NewTopicModel(subject)

		err := m.Topics.Save(ctx, topic)
		if err != nil {
			return nil, err
		}
	}

	// TODO put size as constant in config file
	result := make(chan broker.Message, 1000)

	err = m.Topics.SaveConnection(
		ctx,
		subject,
		model.NewConnectionModel(&result),
	)

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (m *Module) Fetch(ctx context.Context, subject string, id int) (broker.Message, error) {
	if m.Closed {
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
