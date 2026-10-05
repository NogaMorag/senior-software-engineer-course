package cache

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisStore adapts a go-redis client to Store.
type RedisStore struct {
	rdb redis.UniversalClient
}

// NewRedisStore returns a Store over rdb (a *redis.Client or a cluster client).
func NewRedisStore(rdb redis.UniversalClient) *RedisStore { return &RedisStore{rdb: rdb} }

// Get returns the bytes under key, or ErrMiss. go-redis reports a missing key as
// the error redis.Nil; this is where it becomes our own sentinel.
func (s *RedisStore) Get(ctx context.Context, key string) ([]byte, error) {
	b, err := s.rdb.Get(ctx, key).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, ErrMiss
	}
	return b, err
}

// Set writes value under key with a time to live.
func (s *RedisStore) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return s.rdb.Set(ctx, key, value, ttl).Err()
}

// Del deletes keys; deleting a missing key is not an error.
func (s *RedisStore) Del(ctx context.Context, keys ...string) error {
	return s.rdb.Del(ctx, keys...).Err()
}
