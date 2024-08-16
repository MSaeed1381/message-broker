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
	conf   Config
}

func NewModule(topic store.Topic, cache cache.Cache, config Config) broker.Broker {
	return &Module{
		Topics: topic,
		Closed: false,
		Cache:  cache,
		conf:   config,
	}
}

func (m *Module) Close() error {
	m.Closed = true
	return nil
}

func (m *Module) Publish(ctx context.Context, subject string, msg broker.Message) (uint64, error) {
	if m.Closed {
		return 0, broker.ErrUnavailable
	}

	topic, err := m.Topics.GetBySubject(ctx, subject)
	if err != nil {
		if errors.As(err, &store.ErrTopicNotFound{}) {
			// create the new model and saves to data store
			topic = model.NewTopicModel(subject)
			if err := m.Topics.Save(ctx, topic); err != nil {
				return 0, err
			}
		} else {
			return 0, err
		}
	}

	// save the message in its topic and return msgId
	message := model.NewMessageModel(subject, &msg)
	msgId, err := m.Topics.SaveMessage(ctx, message)
	if err != nil {
		return 0, err
	}

	// add message to each channel that subscribe for this topic
	if err := m.Topics.SendMessageToSubscribers(ctx, subject, message); err != nil {
		return 0, err
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
		if errors.As(err, &store.ErrTopicNotFound{}) {
			topic = model.NewTopicModel(subject)
			if err := m.Topics.Save(ctx, topic); err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}
	result := make(chan broker.Message, m.conf.ChannelBufferSize)

	if err = m.Topics.SaveConnection(ctx, subject, model.NewConnectionModel(result)); err != nil {
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
	if err != nil {
		if errors.As(err, &store.ErrMessageNotFound{}) {
			return broker.Message{}, broker.ErrInvalidID
		} else if (errors.As(err, &store.ErrMessageExpired{})) {
			return broker.Message{}, broker.ErrExpiredID
		}
		return broker.Message{}, err
	}

	return *msg.BrokerMessage, nil
}
