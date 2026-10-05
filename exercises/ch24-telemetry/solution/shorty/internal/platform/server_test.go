package platform_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net"
	"net/http"
	"testing"
	"time"

	"example.com/ch24-telemetry/internal/platform"
)

// slowServer starts Serve with a handler that blocks until release is closed,
// and returns the base URL, a channel that receives once the handler has
// started, the channel to close, the cancel that starts the shutdown, and
// the channel where Serve's result arrives.
func slowServer(t *testing.T, grace time.Duration) (url string, started <-chan struct{}, release chan struct{}, shutdown context.CancelFunc, served <-chan error) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	startedc := make(chan struct{}, 1)
	release = make(chan struct{})
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startedc <- struct{}{}
		<-release
		_, _ = io.WriteString(w, "finished")
	})
	log := slog.New(slog.DiscardHandler)
	ctx, cancel := context.WithCancel(context.Background())
	servedc := make(chan error, 1)
	go func() { servedc <- platform.Serve(ctx, platform.NewServer(h, log), ln, grace) }()
	return "http://" + ln.Addr().String(), startedc, release, cancel, servedc
}

func TestShutdownLetsRequestsInFlightFinish(t *testing.T) {
	url, started, release, shutdown, served := slowServer(t, 5*time.Second)

	type result struct {
		body string
		err  error
	}
	got := make(chan result, 1)
	go func() {
		resp, err := http.Get(url)
		if err != nil {
			got <- result{err: err}
			return
		}
		defer resp.Body.Close()
		b, err := io.ReadAll(resp.Body)
		got <- result{string(b), err}
	}()

	<-started  // the request is in the handler
	shutdown() // what SIGTERM does in main
	select {
	case err := <-served:
		t.Fatalf("Serve returned %v while a request was still in flight", err)
	case <-time.After(100 * time.Millisecond):
	}
	if _, err := net.DialTimeout("tcp", url[len("http://"):], time.Second); err == nil {
		t.Error("new connections are still accepted during shutdown")
	}

	close(release) // the handler finishes
	if r := <-got; r.err != nil || r.body != "finished" {
		t.Errorf("in-flight request got %q, %v; want the full response", r.body, r.err)
	}
	select {
	case err := <-served:
		if err != nil {
			t.Errorf("Serve = %v, want nil after a clean shutdown", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Serve did not return after the shutdown; is ctx ignored?")
	}
}

func TestShutdownGivesUpAfterGrace(t *testing.T) {
	url, started, release, shutdown, served := slowServer(t, 50*time.Millisecond)
	defer close(release)
	go func() {
		if resp, err := http.Get(url); err == nil {
			resp.Body.Close()
		}
	}()
	<-started
	shutdown()
	select {
	case err := <-served:
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Errorf("Serve = %v, want an error wrapping context.DeadlineExceeded", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Serve did not return after the grace period")
	}
}

func TestServeReportsListenerErrors(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	ln.Close() // Serve on a closed listener fails at once
	srv := platform.NewServer(http.NotFoundHandler(), slog.New(slog.DiscardHandler))
	if err := platform.Serve(context.Background(), srv, ln, time.Second); err == nil {
		t.Error("Serve on a closed listener returned nil")
	}
}
