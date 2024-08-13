package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/MSaeed1381/message-broker/internal/model"
	"github.com/MSaeed1381/message-broker/internal/store/batch"
	"github.com/MSaeed1381/message-broker/pkg/broker"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"log"
	"time"
)

type MessageInPostgres struct {
	psql         Postgres
	batchHandler *batch.Handler
}

func NewMessageInPostgres(psql Postgres) *MessageInPostgres {
	m := &MessageInPostgres{
		psql: psql,
	}

	config := batch.Config{
		BufferSize:             2000,
		FlushDuration:          time.Duration(500) * time.Millisecond,
		MessageResponseTimeout: time.Duration(5) * time.Second,
	}

	m.batchHandler = batch.NewBatchHandler(m.SaveBulkMessage, config)
	return m
}

func (m *MessageInPostgres) Save(ctx context.Context, message *model.Message) (uint64, error) {
	err := m.batchHandler.AddAndWait(ctx, message)
	if err != nil {
		return 0, err
	}

	return message.BrokerMessage.Id, nil
}

func (m *MessageInPostgres) GetByID(ctx context.Context, id uint64) (*model.Message, error) {
	query := `SELECT body, expiration, createdAt, subject FROM message WHERE id = @msgId`
	args := pgx.NamedArgs{
		"msgId": id,
	}

	rows, err := m.psql.db.Query(ctx, query, args)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var body string
	var subject string
	var expiration int64
	var createdAt time.Time

	rows.Next()
	if err := rows.Scan(&body, &expiration, &createdAt, &subject); err != nil {
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

func (m *MessageInPostgres) SaveBulkMessage(ctx context.Context, items []*batch.Item) {
	query := `INSERT INTO message(body, expiration, createdAt, subject) VALUES(@body, @expiration, @createdAt, @subject) RETURNING id`

	newBatch := &pgx.Batch{}
	for _, item := range items {
		args := pgx.NamedArgs{
			"body":       item.Message.BrokerMessage.Body,
			"expiration": item.Message.BrokerMessage.Expiration.Seconds(),
			"subject":    item.Message.Subject,
			"createdAt":  time.Now().UTC(),
		}
		newBatch.Queue(query, args)
	}

	results := m.psql.db.SendBatch(ctx, newBatch)

	defer func(results pgx.BatchResults) {
		err := results.Close()
		if err != nil {
			panic("failed to close batch")
		}
	}(results)

	for _, item := range items {
		var id uint64
		err := results.QueryRow().Scan(&id)
		item.Message.BrokerMessage.Id = id
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) {
				log.Printf("message %d already exists", item.Message.BrokerMessage.Id)
				continue
			}

			item.Message.BrokerMessage.Id = 0 // Error when saving batch

			fmt.Println(err)
			fmt.Printf("unable to insert row %d: %e", item.Message.BrokerMessage.Id, err)
		}

		close(item.Done)
	}

	return
}
