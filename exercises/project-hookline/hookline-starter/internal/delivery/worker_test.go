package delivery_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"example.com/hookline/internal/backoff"
	"example.com/hookline/internal/delivery"
	"example.com/hookline/internal/domain"
)

func pool(w *world) *delivery.Pool {
	return &delivery.Pool{
		DB: w.pool, Sender: delivery.Sender{Timeout: 2 * time.Second}, Workers: 4,
		Lease: 5 * time.Second, Poll: 20 * time.Millisecond, MaxAttempts: 4,
		Backoff: backoff.Policy{Base: 10 * time.Millisecond, Cap: 50 * time.Millisecond},
	}
}

// waitFor polls the event's deliveries until done says yes, or fails after 10 seconds.
func waitFor(t *testing.T, w *world, ev domain.Event, done func([]domain.Delivery) bool) []domain.Delivery {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for {
		ds := w.history(ev)
		if done(ds) {
			return ds
		}
		if time.Now().After(deadline) {
			t.Fatalf("deliveries did not settle: %+v", ds)
		}
		time.Sleep(20 * time.Millisecond)
	}
}

func settled(ds []domain.Delivery) bool {
	for _, d := range ds {
		if d.Status != domain.Delivered && d.Status != domain.Dead {
			return false
		}
	}
	return len(ds) > 0
}

// TestPoolRetriesFailsAndDelivers runs the pool against three receivers: one that fails twice and
// then accepts, one that rejects with 400, and a disabled endpoint that must never be called.
func TestPoolRetriesFailsAndDelivers(t *testing.T) {
	w := newWorld(t)
	var flaky, rejected, disabled atomic.Int64
	mux := http.NewServeMux()
	mux.HandleFunc("/flaky", func(rw http.ResponseWriter, r *http.Request) {
		if flaky.Add(1) <= 2 {
			rw.WriteHeader(http.StatusServiceUnavailable)
		}
	})
	mux.HandleFunc("/reject", func(rw http.ResponseWriter, r *http.Request) {
		rejected.Add(1)
		rw.WriteHeader(http.StatusBadRequest)
	})
	mux.HandleFunc("/disabled", func(http.ResponseWriter, *http.Request) { disabled.Add(1) })
	srv := httptest.NewServer(mux)
	defer srv.Close()

	ok, bad, off := w.endpoint(srv.URL+"/flaky"), w.endpoint(srv.URL+"/reject"), w.endpoint(srv.URL+"/disabled")
	if _, err := w.st.DisableEndpoint(context.Background(), w.tenant.ID, off.ID); err != nil {
		t.Fatal(err)
	}
	ev := w.deliveries(ok, bad, off)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- pool(w).Run(ctx) }()
	ds := waitFor(t, w, ev, settled)
	cancel()
	if err := <-done; err != nil {
		t.Fatal(err)
	}

	byEndpoint := map[string]domain.Delivery{}
	for _, d := range ds {
		byEndpoint[d.EndpointID] = d
	}
	if d := byEndpoint[ok.ID]; d.Status != domain.Delivered || d.Attempts != 3 || len(d.History) != 3 {
		t.Errorf("flaky receiver: %+v; want delivered on attempt 3", d)
	}
	if d := byEndpoint[bad.ID]; d.Status != domain.Dead || d.Attempts != 1 || rejected.Load() != 1 {
		t.Errorf("400 receiver: %+v after %d calls; want dead after one", d, rejected.Load())
	}
	if d := byEndpoint[off.ID]; d.Status != domain.Dead || disabled.Load() != 0 {
		t.Errorf("disabled endpoint: %+v after %d calls; want dead without a call", d, disabled.Load())
	}
}

// TestPoolDrainsOnShutdown: cancelling the pool while a request is in flight lets the request finish
// and records its result before Run returns.
func TestPoolDrainsOnShutdown(t *testing.T) {
	w := newWorld(t)
	started, release := make(chan struct{}), make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		close(started)
		<-release
	}))
	defer srv.Close()
	ev := w.deliveries(w.endpoint(srv.URL))

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- pool(w).Run(ctx) }()
	<-started
	cancel() // SIGTERM while the attempt is in flight
	select {
	case <-done:
		t.Fatal("Run returned while an attempt was in flight")
	case <-time.After(100 * time.Millisecond):
	}
	close(release)
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	if d := w.history(ev)[0]; d.Status != domain.Delivered {
		t.Fatalf("after the drain: %+v, want delivered", d)
	}
}
