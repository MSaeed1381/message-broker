package broker

import "errors"

var (
	// ErrUnavailable Use this error for the calls that are coming after that
	// server is shutting down
	ErrUnavailable = errors.New("service is unavailable")
	// ErrInvalidID Use this error when the message with provided id is not available
	ErrInvalidID = errors.New("message with id provided is not valid or never published")
	// ErrExpiredID Use this error when message had been published, but it is not
	// available anymore because the expiration time has reached.
	ErrExpiredID = errors.New("message with id provided is expired")
)
