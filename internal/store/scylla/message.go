package scylla

import (
	"context"
	"fmt"
	"github.com/MSaeed1381/message-broker/internal/model"
	"github.com/MSaeed1381/message-broker/internal/store/batch"
	"github.com/MSaeed1381/message-broker/pkg/broker"
	"github.com/gocql/gocql"
	"github.com/google/uuid"
	"time"
)

type MessageInScylla struct {
	scylla       *Scylla
	batchHandler *batch.Handler
}

func NewMessageInScylla(scylla *Scylla) *MessageInScylla {
	m := &MessageInScylla{
		scylla: scylla,
	}

	batchConfig := batch.Config{
		BufferSize:             500,
		FlushDuration:          time.Duration(500) * time.Millisecond,
		MessageResponseTimeout: time.Duration(5) * time.Second,
	}

	m.batchHandler = batch.NewBatchHandler(m.SaveBulkMessage, batchConfig)
	return m
}

func (m *MessageInScylla) Save(ctx context.Context, message *model.Message) (uint64, error) {
	err := m.batchHandler.AddAndWait(ctx, message)
	if err != nil {
		return 0, err
	}

	return message.BrokerMessage.Id, nil
}

func (m *MessageInScylla) GetByID(ctx context.Context, id uint64) (*model.Message, error) {
	query := `SELECT body, expiration, createdAt, subject FROM message_broker.message WHERE id = ?`

	var body string
	var subject string
	var expiration int64
	var createdAt time.Time

	if err := m.scylla.session.Query(query, id).Scan(&body, &expiration, &createdAt, &subject); err != nil {
		return nil, err
	}

	msg := &model.Message{
		BrokerMessage: &broker.Message{
			Id:         id,
			Body:       body,
			Expiration: time.Duration(expiration) * time.Second,
		},
		CreateAt: createdAt,
		Subject:  subject,
	}

	return msg, nil
}

func (m *MessageInScylla) SaveBulkMessage(ctx context.Context, items []*batch.Item) {
	newBatch := m.scylla.session.NewBatch(gocql.UnloggedBatch)

	for _, item := range items {
		newId := uuid.New().ID()
		query := `INSERT INTO message_broker.message (id, body, createdAt, subject, expiration) VALUES (?, ?, ?, ?, ?) USING TTL ?`
		newBatch.Query(query, newId, item.Message.BrokerMessage.Body, item.Message.CreateAt, item.Message.Subject,
			int64(item.Message.BrokerMessage.Expiration.Seconds()), int64(item.Message.BrokerMessage.Expiration.Seconds()))

		item.Message.BrokerMessage.Id = uint64(newId)

		close(item.Done)
	}

	if err := m.scylla.session.ExecuteBatch(newBatch.WithContext(ctx)); err != nil {
		fmt.Println(err)
	}
}
