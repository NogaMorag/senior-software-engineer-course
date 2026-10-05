package ratelimit

import (
	"context"
	_ "embed"

	"github.com/redis/go-redis/v9"
)

//go:embed tokenbucket.lua
var tokenBucketLua string

// tokenBucket is loaded once per process. Run sends EVALSHA with the script's SHA1 and,
// when Redis answers NOSCRIPT (restart, failover, SCRIPT FLUSH), falls back to EVAL,
// which also puts the script back in Redis' cache.
var tokenBucket = redis.NewScript(tokenBucketLua)

// Redis is a token bucket per key stored in Redis, so every instance of a service draws
// from the same bucket. The whole read-refill-take-write runs as one script: Redis runs a
// script atomically, so two instances can never both spend the last token.
type Redis struct {
	Client redis.Scripter
	Prefix string // namespace for the keys, e.g. "hookline:rl:ingest:"
}

// Allow takes one token from key's bucket, refilled at rate up to burst. An error means Redis
// could not be asked; the caller decides whether to fail open (serve) or closed (refuse).
func (r *Redis) Allow(ctx context.Context, key string, rate float64, burst int) (Decision, error) {
	panic("TODO: see chapter 38")

}
