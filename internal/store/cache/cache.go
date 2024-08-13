package cache

import (
	"context"
	"time"
)

type Cache interface {
	Set(ctx context.Context, msgId uint64, duration time.Duration) error
	IsKeyExpired(ctx context.Context, msgId uint64) (bool, error)
}

type NoImpl struct{}

func NewNoImpl() Cache {
	return &NoImpl{}
}

func (n *NoImpl) Set(_ context.Context, _ uint64, _ time.Duration) error {
	return nil
}

func (n *NoImpl) IsKeyExpired(_ context.Context, _ uint64) (bool, error) {
	return false, nil
}
