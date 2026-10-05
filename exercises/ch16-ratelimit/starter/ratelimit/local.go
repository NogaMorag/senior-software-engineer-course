package ratelimit

import (
	"context"
	"sync"
	"time"
)

// Limiter decides per key: a tenant, an API key, a client address.
type Limiter interface {
	Allow(ctx context.Context, key string) (Decision, error)
}

// Local keeps one Bucket per key in this process. Every replica of a service has its own,
// so with three replicas a client gets three times the limit: use Redis for a shared one.
// Buckets are never removed; a real service evicts idle keys.
type Local struct {
	Rate  float64
	Burst int
	Now   func() time.Time // default time.Now

	mu      sync.Mutex
	buckets map[string]*Bucket
}

// Allow takes one token from key's bucket. It never fails.
func (l *Local) Allow(_ context.Context, key string) (Decision, error) {
	l.mu.Lock()
	b, ok := l.buckets[key]
	if !ok {
		if l.buckets == nil {
			l.buckets = make(map[string]*Bucket)
		}
		now := l.Now
		if now == nil {
			now = time.Now
		}
		b = NewBucket(l.Rate, l.Burst, now)
		l.buckets[key] = b
	}
	l.mu.Unlock()
	return b.Allow(), nil
}
