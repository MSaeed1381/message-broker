package cache

import (
	"context"
	"github.com/MSaeed1381/message-broker/internal/store"
	"github.com/redis/go-redis/v9"
	"strconv"
	"time"
)

type Redis struct {
	Client *redis.Client
}

// NewRedisClient initializes and returns a Redis client
func NewRedisClient(config Config) *Redis {
	return &Redis{
		Client: redis.NewClient(&redis.Options{
			Addr:     config.Address,
			Password: config.Password,
			DB:       config.DBNumber,
		})}
}

// Set stores the expiration time in nanoseconds for a given key in Redis
func (rdb *Redis) Set(ctx context.Context, msgId uint64, duration time.Duration) error {
	expirationInNano := time.Now().Add(duration).UnixNano()
	return rdb.Client.Set(ctx, strconv.FormatUint(msgId, 10), expirationInNano, 0).Err() // Set with no TTL
}

// IsKeyExpired checks if the key has expired based on the stored expiration time
func (rdb *Redis) IsKeyExpired(ctx context.Context, msgId uint64) (bool, error) {
	val, err := rdb.Client.Get(ctx, strconv.FormatUint(msgId, 10)).Result()
	if err != nil {
		return false, store.ErrMessageNotFound{ID: msgId}
	}

	storedExpirationNano, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return false, err
	}

	return time.Now().UnixNano() > storedExpirationNano, nil
}
