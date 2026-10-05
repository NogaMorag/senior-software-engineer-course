package wire_test

import (
	"context"
	"testing"

	"example.com/hookline/internal/domain"
	"example.com/hookline/internal/wire"
)

type loader struct{ n int }

func (l *loader) ActiveSubscriptions(context.Context, string) ([]domain.Subscription, error) {
	l.n++
	return []domain.Subscription{{EndpointID: "ep", EventType: "*"}}, nil
}

// Without REDIS_ADDR the cache is disabled: every lookup reads the source, and nothing fails.
func TestSubscriptionsWithoutRedis(t *testing.T) {
	l := &loader{}
	subs := wire.Subscriptions(wire.Redis(""), l)
	for range 2 {
		if ids, err := subs.Endpoints(context.Background(), "t", "a.b"); err != nil || len(ids) != 1 {
			t.Fatalf("Endpoints = %v, %v", ids, err)
		}
	}
	if l.n != 2 {
		t.Fatalf("loads = %d, want 2", l.n)
	}
}

func TestLimiterWithoutRedisIsLocal(t *testing.T) {
	lim := wire.Limiter(wire.Redis(""))
	d, err := lim.Allow(context.Background(), "tenant", 1, 1)
	if err != nil || !d.Allowed {
		t.Fatalf("first request = %+v, %v", d, err)
	}
	if d, _ := lim.Allow(context.Background(), "tenant", 1, 1); d.Allowed {
		t.Fatal("second request within the same second allowed with burst 1")
	}
}
