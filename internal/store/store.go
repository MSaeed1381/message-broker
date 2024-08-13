package store

import (
	"context"
	"fmt"
	"github.com/MSaeed1381/message-broker/internal/model"
)

// Message Store Errors

// ErrMessageNotFound (Message Not Found Error)
type ErrMessageNotFound struct {
	ID uint64
}

// override Error method
func (e ErrMessageNotFound) Error() string {
	return fmt.Sprintf("message %d not found.", e.ID)
}

// ErrMessageAlreadyExists (Message Already Exists)
type ErrMessageAlreadyExists struct {
	ID uint64
}

// override Error method
func (e ErrMessageAlreadyExists) Error() string {
	return fmt.Sprintf("message %d already exists.", e.ID)
}

// ErrMessageInvalid (Message Invalid)
type ErrMessageInvalid struct {
	ID uint64
}

// override Error method
func (e ErrMessageInvalid) Error() string {
	return fmt.Sprintf("message %d invalid.", e.ID)
}

type ErrMessageExpired struct {
	ID uint64
}

func (e ErrMessageExpired) Error() string {
	return fmt.Sprintf("message %d expired.", e.ID)
}

// Message Save message With Broker Message and return Broker Message From Data Store
type Message interface {
	Save(ctx context.Context, message *model.Message) (uint64, error)
	GetByID(ctx context.Context, id uint64) (*model.Message, error)
}

// Topic Store Errors

// ErrTopicNotFound not found error
type ErrTopicNotFound struct {
	Subject string
}

func (e ErrTopicNotFound) Error() string {
	return fmt.Sprintf("topic %s not found.", e.Subject)
}

type ErrTopicAlreadyExists struct {
	Subject string
}

func (e ErrTopicAlreadyExists) Error() string {
	return fmt.Sprintf("topic %s already exists.", e.Subject)
}

type Topic interface {
	Save(ctx context.Context, topic *model.Topic) error
	GetBySubject(ctx context.Context, subject string) (*model.Topic, error)
	GetOpenConnections(ctx context.Context, subject string) ([]*model.Connection, error)
	SaveMessage(ctx context.Context, message *model.Message) (uint64, error)
	GetMessage(ctx context.Context, messageId uint64, subject string) (*model.Message, error)
	SaveConnection(ctx context.Context, subject string, connection *model.Connection) error
}

type Connection interface {
	Save(ctx context.Context, connection *model.Connection) error
	GetAllConnections(ctx context.Context) ([]*model.Connection, error)
}
