// Package redislock is the single-instance Redis lock: SET key token NX PX ttl to take it,
// and a compare-and-delete to give it back. It is fine for efficiency (avoid doing the same
// work twice). It is not safe for correctness on its own: the key expires after ttl whether
// or not the holder is done, so a paused holder can wake up after someone else took the
// lock. Correctness needs a fencing token checked by the storage.
package redislock

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

// ErrNotAcquired means someone else holds the lock.
var ErrNotAcquired = errors.New("redislock: held by someone else")

// ErrNotHeld means the lock has expired, and perhaps been taken by someone else.
var ErrNotHeld = errors.New("redislock: not held any more")

// Lock is one acquired lock.
type Lock struct {
	rdb   *redis.Client
	key   string
	token string // unique to this acquisition, so only this holder can release it
}

// Acquire takes the lock on key for ttl, or returns ErrNotAcquired without waiting.
func Acquire(ctx context.Context, rdb *redis.Client, key string, ttl time.Duration) (*Lock, error) {
	panic("TODO(day 52): SET key token NX PX ttl with a random token; see section 18.3")
}

// The scripts compare the stored token with ours and act only if they match. Redis runs a
// script atomically, so no other command lands between the GET and the DEL.
var (
	release = redis.NewScript(`
if redis.call("GET", KEYS[1]) == ARGV[1] then
	return redis.call("DEL", KEYS[1])
end
return 0`)
	extend = redis.NewScript(`
if redis.call("GET", KEYS[1]) == ARGV[1] then
	return redis.call("PEXPIRE", KEYS[1], ARGV[2])
end
return 0`)
)

// Release gives the lock back. A plain DEL would delete whoever's lock is there now; the
// script deletes the key only if it still holds this holder's token.
func (l *Lock) Release(ctx context.Context) error {
	panic("TODO(day 52): run the release script; zero deleted keys means ErrNotHeld; see section 18.3")
}

// Extend resets the lock's expiry to ttl from now, if this holder still has it.
func (l *Lock) Extend(ctx context.Context, ttl time.Duration) error {
	n, err := extend.Run(ctx, l.rdb, []string{l.key}, l.token, ttl.Milliseconds()).Int()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotHeld
	}
	return nil
}
