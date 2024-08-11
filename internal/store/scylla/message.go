package scylla

import (
	"context"
	"errors"
	"fmt"
	"github.com/MSaeed1381/message-broker/internal/model"
	"github.com/MSaeed1381/message-broker/internal/store/batch"
	"github.com/MSaeed1381/message-broker/internal/utils"
	"github.com/MSaeed1381/message-broker/pkg/broker"
	"github.com/gocql/gocql"
	"time"
)

type MessageInScylla struct {
	scylla       *Scylla
	batchHandler *batch.Handler
	idGenerator  utils.IdGenerator
}

func NewMessageInScylla(scylla *Scylla) *MessageInScylla {
	m := &MessageInScylla{
		scylla: scylla,
	}

	m.batchHandler = batch.NewBatchHandler(m.SaveBulkMessage, 3000)
	return m
}

func (m *MessageInScylla) Save(ctx context.Context, message *model.Message) (uint64, error) {

	start := time.Now()
	m.batchHandler.AddAndWait(ctx, message)

	if message.BrokerMessage.Id == 0 {
		return 0, errors.New("saving message failed")
	}

	fmt.Println(time.Since(start))

	return uint64(message.BrokerMessage.Id), nil
}

func (m *MessageInScylla) GetByID(ctx context.Context, id uint64) (*model.Message, error) {
	query := `SELECT body, expiration, createdAt, subject FROM message_broker.messages WHERE id = ?`

	var body string
	var subject string
	var expiration int64
	var createdAt time.Time

	if err := m.scylla.session.Query(query, id).Scan(&body, &expiration, &createdAt, &subject); err != nil {
		return nil, err
	}

	msg := &model.Message{
		BrokerMessage: &broker.Message{
			Id:         int(id),
			Body:       body,
			Expiration: time.Duration(expiration) * time.Second,
		},
		CreateAt: createdAt,
		Subject:  subject,
	}

	return msg, nil
}

func (m *MessageInScylla) SaveBulkMessage(ctx context.Context, messages []*model.Message) {
	newBatch := m.scylla.session.NewBatch(gocql.UnloggedBatch)

	for _, msg := range messages {
		newId := int(m.idGenerator.Next())
		query := `INSERT INTO message_broker.messages (id, body, createdAt, subject, expiration) VALUES (?, ?, ?, ?, ?)`
		newBatch.Query(query, newId, msg.BrokerMessage.Body, msg.CreateAt, msg.Subject, int64(msg.BrokerMessage.Expiration.Seconds()))

		msg.BrokerMessage.Id = newId
	}

	if err := m.scylla.session.ExecuteBatch(newBatch.WithContext(ctx)); err != nil {
		fmt.Println(err)
	}
}
