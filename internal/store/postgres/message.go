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

	m.batchHandler = batch.NewBatchHandler(m.SaveBulkMessage, 1000)
	return m
}

func (m *MessageInPostgres) Save(ctx context.Context, message *model.Message) (uint64, error) {

	m.batchHandler.AddAndWait(ctx, message)

	if message.BrokerMessage.Id == 0 {
		return 0, errors.New("saving message failed")
	}

	return uint64(message.BrokerMessage.Id), nil
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
			Id:         int(id),
			Body:       body,
			Expiration: time.Duration(expiration) * time.Second,
		},
		CreateAt: createdAt,
		Subject:  subject,
	}

	return msg, nil
}

func (m *MessageInPostgres) SaveBulkMessage(ctx context.Context, messages []*model.Message) {
	query := `INSERT INTO message(body, expiration, createdAt, subject) VALUES(@body, @expiration, @createdAt, @subject) RETURNING id`

	newBatch := &pgx.Batch{}
	for _, msg := range messages {
		args := pgx.NamedArgs{
			"body":       msg.BrokerMessage.Body,
			"expiration": msg.BrokerMessage.Expiration.Seconds(),
			"subject":    msg.Subject,
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

	for _, msg := range messages {
		var id uint64
		err := results.QueryRow().Scan(&id)
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) {
				log.Printf("message %d already exists", msg.BrokerMessage.Id)
				continue
			}

			msg.BrokerMessage.Id = 0 // Error when saving batch

			fmt.Println(err)
			fmt.Printf("unable to insert row %d: %e", msg.BrokerMessage.Id, err)
		}

		msg.BrokerMessage.Id = int(id)
	}

	return
}
