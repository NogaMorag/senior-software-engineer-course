package cache

import (
	"context"

	"time"

	"github.com/redis/go-redis/v9"
)

// RedisStore adapts a go-redis client to Store.
type RedisStore struct {
	rdb redis.UniversalClient
}

// NewRedisStore returns a Store over rdb (a *redis.Client or a cluster client).
func NewRedisStore(rdb redis.UniversalClient) *RedisStore { panic("TODO: see chapter 38") }

// Get returns the bytes under key, or ErrMiss. go-redis reports a missing key as
// the error redis.Nil; this is where it becomes our own sentinel.
func (s *RedisStore) Get(ctx context.Context, key string) ([]byte, error) {
	panic("TODO: see chapter 38")

}

// Set writes value under key with a time to live.
func (s *RedisStore) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	panic("TODO: see chapter 38")

}

// Del deletes keys; deleting a missing key is not an error.
func (s *RedisStore) Del(ctx context.Context, keys ...string) error {
	panic("TODO: see chapter 38")

}
