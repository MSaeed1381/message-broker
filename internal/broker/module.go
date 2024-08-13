package broker

import (
	"context"
	"errors"
	"github.com/MSaeed1381/message-broker/internal/model"
	"github.com/MSaeed1381/message-broker/internal/store"
	"github.com/MSaeed1381/message-broker/internal/store/cache"
	"github.com/MSaeed1381/message-broker/pkg/broker"
)

type Module struct {
	Topics store.Topic // have msg Store and connection Store
	Cache  cache.Cache // store black-list for message expiration
	Closed bool
}

func NewModule(topic store.Topic, cache cache.Cache) broker.Broker {
	return &Module{
		Topics: topic,
		Closed: false,
		Cache:  cache,
	}
}

// Close TODO implement with Channel
func (m *Module) Close() error {
	m.Closed = true
	return nil
}

func (m *Module) Publish(ctx context.Context, subject string, msg *broker.Message) (uint64, error) {
	if m.Closed {
		return 0, broker.ErrUnavailable
	}

	topic, err := m.Topics.GetBySubject(ctx, subject)

	// TODO synchronize this
	if err != nil && errors.As(err, &store.ErrTopicNotFound{}) {
		// create the new model and saves to data store
		topic = model.NewTopicModel(subject)

		err := m.Topics.Save(ctx, topic)
		if err != nil {
			return 0, err
		}
	}

	msgId, err := m.Topics.SaveMessage(ctx, model.NewMessageModel(subject, msg))
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
		}(connection, msg)
	}

	// set the msg in cache memory
	if err := m.Cache.Set(ctx, msgId, msg.Expiration); err != nil {
		return 0, err
	}

	return msgId, nil
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

func (m *Module) Fetch(ctx context.Context, subject string, id uint64) (broker.Message, error) {
	if m.Closed {
		return broker.Message{}, broker.ErrUnavailable
	}

	// expiration handling section
	expired, err := m.Cache.IsKeyExpired(ctx, id)

	if err != nil {
		if errors.As(err, &store.ErrMessageNotFound{}) {
			return broker.Message{}, broker.ErrInvalidID
		}
		return broker.Message{}, err
	}

	if expired {
		return broker.Message{}, broker.ErrExpiredID
	}

	// get message from data store
	msg, err := m.Topics.GetMessage(ctx, id, subject)

	// handling error occurs in store level
	// TODO (we can check if id is less that current id and if not in map so we can find that was expired)
	if errors.As(err, &store.ErrMessageNotFound{}) {
		return broker.Message{}, broker.ErrInvalidID
	} else if (errors.As(err, &store.ErrMessageExpired{})) {
		return broker.Message{}, broker.ErrExpiredID
	} else if err != nil {
		return broker.Message{}, err
	}

	return *msg.BrokerMessage, nil
}
