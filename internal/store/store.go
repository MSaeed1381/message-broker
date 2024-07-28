package store

import (
	"fmt"
	"therealbroker/internal/model"
	"therealbroker/pkg/broker"
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

// Message Save message With Broker Message and return Broker Message From Data Store
type Message interface {
	Save(message *broker.Message) (uint64, error)
	GetByID(id int) (*model.Message, error)
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
	Save(topic *model.Topic) error
	GetBySubject(subject string) (*model.Topic, error)
	GetOpenConnections(subject string) ([]*model.Connection, error)
	SaveMessage(subject string, message *broker.Message) (uint64, error)
	GetMessage(messageId uint64, subject string) (*model.Message, error)
	SaveConnection(subject string, connection *model.Connection) error
}
