package subcache_test

import (
	"context"
	"errors"
	"os"
	"slices"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"

	"example.com/hookline/internal/cache"
	"example.com/hookline/internal/domain"
	"example.com/hookline/internal/subcache"
)

// mapStore is an in-process cache.Store, so the unit tests need no Redis.
type mapStore struct {
	mu sync.Mutex
	m  map[string][]byte
}

func (s *mapStore) Get(_ context.Context, k string) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if v, ok := s.m[k]; ok {
		return v, nil
	}
	return nil, cache.ErrMiss
}

func (s *mapStore) Set(_ context.Context, k string, v []byte, _ time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.m[k] = v
	return nil
}

func (s *mapStore) Del(_ context.Context, ks ...string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, k := range ks {
		delete(s.m, k)
	}
	return nil
}

// loader serves a fixed list and counts reads of the "database".
type loader struct {
	mu    sync.Mutex
	subs  []domain.Subscription
	loads atomic.Int64
	delay time.Duration
	err   error
}

func (l *loader) ActiveSubscriptions(_ context.Context, tenantID string) ([]domain.Subscription, error) {
	l.loads.Add(1)
	time.Sleep(l.delay)
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.err != nil {
		return nil, l.err
	}
	var out []domain.Subscription
	for _, s := range l.subs {
		if s.TenantID == tenantID {
			out = append(out, s)
		}
	}
	return out, nil
}

func (l *loader) set(subs ...domain.Subscription) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.subs = subs
}

func sub(tenant, endpoint, typ string) domain.Subscription {
	return domain.Subscription{ID: endpoint + "/" + typ, TenantID: tenant, EndpointID: endpoint, EventType: typ}
}

var opts = cache.Options{TTL: time.Minute}

func TestEndpointsMatchesExactAndWildcardOnce(t *testing.T) {
	l := &loader{}
	l.set(sub("t1", "ep-a", "invoice.paid"), sub("t1", "ep-a", "*"), sub("t1", "ep-b", "*"),
		sub("t1", "ep-c", "shipment.moved"), sub("t2", "ep-z", "*"))
	c := subcache.New(&mapStore{m: map[string][]byte{}}, l, opts)
	got, err := c.Endpoints(context.Background(), "t1", "invoice.paid")
	if err != nil || !slices.Equal(got, []string{"ep-a", "ep-b"}) {
		t.Fatalf("Endpoints = %v, %v; want [ep-a ep-b] (ep-a once, ep-c not, ep-z is another tenant's)", got, err)
	}
	if none, _ := c.Endpoints(context.Background(), "t3", "invoice.paid"); len(none) != 0 {
		t.Fatalf("a tenant without subscriptions matches %v", none)
	}
}

func TestOneLoadPerTenantUntilInvalidated(t *testing.T) {
	ctx := context.Background()
	l := &loader{}
	l.set(sub("t1", "ep-a", "invoice.paid"))
	c := subcache.New(&mapStore{m: map[string][]byte{}}, l, opts)
	for _, typ := range []string{"invoice.paid", "shipment.moved", "invoice.paid"} {
		if _, err := c.Endpoints(ctx, "t1", typ); err != nil {
			t.Fatal(err)
		}
	}
	if n := l.loads.Load(); n != 1 {
		t.Fatalf("loads = %d for three lookups of one tenant, want 1 (one entry per tenant)", n)
	}
	l.set(sub("t1", "ep-a", "invoice.paid"), sub("t1", "ep-new", "*"))
	if err := c.Invalidate(ctx, "t1"); err != nil {
		t.Fatal(err)
	}
	got, _ := c.Endpoints(ctx, "t1", "invoice.paid")
	if !slices.Equal(got, []string{"ep-a", "ep-new"}) || l.loads.Load() != 2 {
		t.Fatalf("after invalidate = %v with %d loads; want the new subscription and a second load", got, l.loads.Load())
	}
	if s := c.Stats(); s.Hits != 2 || s.Misses != 2 {
		t.Errorf("stats = %+v, want 2 hits and 2 misses", s)
	}
}

// TestStampedeLoadsOnce: a burst of events for one tenant right after an invalidation reads the
// database once, not once per event.
func TestStampedeLoadsOnce(t *testing.T) {
	l := &loader{delay: 50 * time.Millisecond}
	l.set(sub("t1", "ep-a", "*"))
	c := subcache.New(&mapStore{m: map[string][]byte{}}, l, opts)
	var wg sync.WaitGroup
	start := make(chan struct{})
	for range 40 {
		wg.Go(func() {
			<-start
			if got, err := c.Endpoints(context.Background(), "t1", "invoice.paid"); err != nil || len(got) != 1 {
				t.Errorf("Endpoints = %v, %v", got, err)
			}
		})
	}
	close(start)
	wg.Wait()
	if n := l.loads.Load(); n != 1 {
		t.Fatalf("loads = %d for 40 concurrent misses, want 1", n)
	}
}

func TestLoadErrorsPassThroughAndAreNotCached(t *testing.T) {
	boom := errors.New("database down")
	l := &loader{err: boom}
	c := subcache.New(&mapStore{m: map[string][]byte{}}, l, opts)
	if _, err := c.Endpoints(context.Background(), "t1", "a.b"); !errors.Is(err, boom) {
		t.Fatalf("err = %v, want %v", err, boom)
	}
	l.err = nil
	l.set(sub("t1", "ep-a", "*"))
	if got, err := c.Endpoints(context.Background(), "t1", "a.b"); err != nil || len(got) != 1 {
		t.Fatalf("after recovery = %v, %v", got, err)
	}
}

func TestDisabledAlwaysReadsTheDatabase(t *testing.T) {
	l := &loader{}
	l.set(sub("t1", "ep-a", "*"))
	c := subcache.New(subcache.Disabled(), l, opts)
	for range 3 {
		if _, err := c.Endpoints(context.Background(), "t1", "a.b"); err != nil {
			t.Fatal(err)
		}
	}
	if n := l.loads.Load(); n != 3 {
		t.Fatalf("loads = %d, want 3: the disabled store caches nothing", n)
	}
	if err := c.Invalidate(context.Background(), "t1"); err != nil {
		t.Fatal(err)
	}
}

// TestAgainstRedis runs the same flow on the compose Redis (REDIS_ADDR), including the JSON round
// trip of the cached value.
func TestAgainstRedis(t *testing.T) {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		t.Skip("integration test: set REDIS_ADDR to the Hookline Redis (make up)")
	}
	rdb := redis.NewClient(&redis.Options{Addr: addr})
	t.Cleanup(func() { rdb.Close() })
	tenant := domain.NewID(time.Now())
	l := &loader{}
	l.set(sub(tenant, "ep-a", "invoice.paid"), sub(tenant, "ep-b", "*"))
	c := subcache.New(cache.NewRedisStore(rdb), l, cache.Options{TTL: time.Minute, Jitter: time.Second})
	ctx := context.Background()
	t.Cleanup(func() { _ = c.Invalidate(context.Background(), tenant) })
	for range 2 {
		got, err := c.Endpoints(ctx, tenant, "invoice.paid")
		if err != nil || !slices.Equal(got, []string{"ep-a", "ep-b"}) {
			t.Fatalf("Endpoints = %v, %v", got, err)
		}
	}
	if l.loads.Load() != 1 {
		t.Fatalf("loads = %d, want 1: the second lookup must come from Redis", l.loads.Load())
	}
	if ttl, err := rdb.PTTL(ctx, subcache.Key(tenant)).Result(); err != nil || ttl <= 0 || ttl > time.Minute+time.Second {
		t.Fatalf("PTTL = %v, %v; want a TTL between 0 and 61s", ttl, err)
	}
}
