package postgres

import (
	"context"
	"github.com/MSaeed1381/message-broker/internal/model"
	"github.com/MSaeed1381/message-broker/pkg/broker"
	"time"
)

type MessageInPostgres struct {
	psql Postgres
}

func NewMessageInPostgres(psql Postgres) *MessageInPostgres {
	return &MessageInPostgres{
		psql: psql,
	}
}

func (m *MessageInPostgres) Save(ctx context.Context, message *broker.Message, subject string) (uint64, error) {
	var id uint64
	err := m.psql.db.QueryRow(ctx, "INSERT INTO message(body, expiration, createdAt, subject) VALUES($1, $2, $3, $4) RETURNING id",
		message.Body, message.Expiration.Seconds(), time.Now(), subject).Scan(&id)

	if err != nil {
		return 0, err
	}
	return id, nil
}

func (m *MessageInPostgres) GetByID(ctx context.Context, id uint64) (*model.Message, error) {
	rows, err := m.psql.db.Query(ctx, "SELECT body, expiration, createdAt FROM message WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var body string
	var expiration int64
	var createdAt time.Time

	rows.Next()
	if err := rows.Scan(&body, &expiration, &createdAt); err != nil {
		return nil, err
	}

	msg := &model.Message{
		BrokerMessage: &broker.Message{
			Id:         int(id),
			Body:       body,
			Expiration: time.Duration(expiration) * time.Second,
		},
		CreateAt: createdAt,
	}

	return msg, nil
}
