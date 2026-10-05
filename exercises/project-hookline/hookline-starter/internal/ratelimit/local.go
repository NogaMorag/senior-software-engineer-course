package ratelimit

import (
	"context"
	"sync"
	"time"
)

// Limiter decides per key: a tenant, an API key, a client address. Hookline passes each tenant's
// own rate and burst with every call (tenants.rate_per_sec and tenants.burst), so one limiter
// serves every tenant.
type Limiter interface {
	Allow(ctx context.Context, key string, rate float64, burst int) (Decision, error)
}

// Local keeps one Bucket per key in this process. Every replica of a service has its own,
// so with three replicas a client gets three times the limit: use Redis for a shared one.
// cmd/api uses it only without REDIS_ADDR. Buckets are never removed, and a bucket keeps the
// rate it was created with; a real service evicts idle keys.
type Local struct {
	Now func() time.Time // default time.Now

	mu      sync.Mutex
	buckets map[string]*Bucket
}

// Allow takes one token from key's bucket, creating it with rate and burst. It never fails.
func (l *Local) Allow(_ context.Context, key string, rate float64, burst int) (Decision, error) {
	panic("TODO: see chapter 38")

}
