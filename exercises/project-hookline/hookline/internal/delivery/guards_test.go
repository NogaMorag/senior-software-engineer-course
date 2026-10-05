package delivery_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"example.com/hookline/internal/breaker"
	"example.com/hookline/internal/delivery"
	"example.com/hookline/internal/domain"
)

func guarded(w *world, g *delivery.Guards) *delivery.Pool {
	p := pool(w)
	p.Guards = g
	p.MaxAttempts = 100 // the breaker, not the attempt budget, must be what stops the requests
	return p
}

// TestOpenBreakerStopsRequestsWithoutSpendingAttempts: after FailureThreshold failures the endpoint
// gets no more requests while the breaker is open, and its deliveries keep their attempt count.
func TestOpenBreakerStopsRequestsWithoutSpendingAttempts(t *testing.T) {
	w := newWorld(t)
	var calls atomic.Int64
	srv := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		rw.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()
	ep := w.endpoint(srv.URL)
	var events []domain.Event
	for range 5 {
		events = append(events, w.deliveries(ep))
	}
	g := &delivery.Guards{Breaker: breaker.Settings{FailureThreshold: 3, OpenTimeout: time.Hour}}
	p := guarded(w, g)
	p.Workers = 1 // one request at a time, so exactly FailureThreshold requests reach the endpoint
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- p.Run(ctx) }()
	deadline := time.Now().Add(5 * time.Second)
	for g.State(ep.ID) != breaker.Open && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if g.State(ep.ID) != breaker.Open {
		t.Fatalf("breaker = %v after %d calls, want open", g.State(ep.ID), calls.Load())
	}
	time.Sleep(300 * time.Millisecond) // the pool keeps running: nothing may reach the endpoint now
	cancel()
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	if n := calls.Load(); n != 3 {
		t.Fatalf("the endpoint got %d requests, want exactly 3: an open breaker sends none", n)
	}
	spent := 0
	for _, ev := range events {
		d := w.history(ev)[0]
		spent += d.Attempts
		if d.Status != domain.Pending {
			t.Errorf("delivery %s is %s, want pending (postponed, not dead)", d.ID, d.Status)
		}
	}
	if spent != 3 {
		t.Fatalf("attempts spent = %d, want 3: postponing must not spend attempts", spent)
	}
}

func TestFinalAnswersDoNotOpenTheBreaker(t *testing.T) {
	w := newWorld(t)
	srv := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) { rw.WriteHeader(http.StatusBadRequest) }))
	defer srv.Close()
	ep := w.endpoint(srv.URL)
	var events []domain.Event
	for range 4 {
		events = append(events, w.deliveries(ep))
	}
	g := &delivery.Guards{Breaker: breaker.Settings{FailureThreshold: 2, OpenTimeout: time.Hour}}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- guarded(w, g).Run(ctx) }()
	for _, ev := range events {
		waitFor(t, w, ev, settled)
	}
	cancel()
	<-done
	if s := g.State(ep.ID); s != breaker.Closed {
		t.Fatalf("breaker = %v after four 400s, want closed: a 400 is an answer from a working endpoint", s)
	}
}

// TestBulkheadKeepsASlowEndpointFromStarvingOthers is the N3 property inside one process: a receiver
// that never answers holds at most Concurrency workers, and a healthy receiver's deliveries still go
// out while it hangs.
func TestBulkheadKeepsASlowEndpointFromStarvingOthers(t *testing.T) {
	w := newWorld(t)
	hang := make(chan struct{})
	defer close(hang)
	var slowInFlight, slowMax atomic.Int64
	slow := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		n := slowInFlight.Add(1)
		for m := slowMax.Load(); n > m && !slowMax.CompareAndSwap(m, n); m = slowMax.Load() {
		}
		<-hang
		slowInFlight.Add(-1)
	}))
	defer slow.Close()
	healthy := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	defer healthy.Close()
	slowEP, okEP := w.endpoint(slow.URL), w.endpoint(healthy.URL)
	for range 10 {
		w.deliveries(slowEP)
	}
	var healthyEvents []domain.Event
	for range 10 {
		healthyEvents = append(healthyEvents, w.deliveries(okEP))
	}

	p := guarded(w, &delivery.Guards{Concurrency: 1})
	p.Sender.Timeout = 30 * time.Second // the slow receiver would hold a worker this long
	p.Lease = time.Minute
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- p.Run(ctx) }()
	start := time.Now()
	for _, ev := range healthyEvents {
		waitFor(t, w, ev, func(ds []domain.Delivery) bool { return len(ds) == 1 && ds[0].Status == domain.Delivered })
	}
	t.Logf("10 healthy deliveries in %v while the slow endpoint hung", time.Since(start))
	if m := slowMax.Load(); m != 1 {
		t.Fatalf("the slow endpoint held %d workers at once, want 1 (the bulkhead)", m)
	}
	cancel()
	hang <- struct{}{} // release the one hung request so the drain can finish
	if err := <-done; err != nil {
		t.Fatal(err)
	}
}
