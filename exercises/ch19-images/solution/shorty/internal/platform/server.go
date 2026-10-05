package platform

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"
)

// NewServer returns an http.Server for h with timeouts set. The zero
// http.Server has none, so one slow client can hold a connection forever.
func NewServer(h http.Handler, log *slog.Logger) *http.Server {
	return &http.Server{
		Handler:           h,
		ReadHeaderTimeout: 5 * time.Second,  // slow-header (Slowloris) clients
		ReadTimeout:       10 * time.Second, // the whole request, body included
		WriteTimeout:      30 * time.Second, // from the end of the headers to the end of the response
		IdleTimeout:       60 * time.Second, // keep-alive connections between requests
		ErrorLog:          slog.NewLogLogger(log.Handler(), slog.LevelWarn),
	}
}

// Serve runs srv on ln until ctx is done, then shuts it down gracefully: it
// stops accepting, lets requests in flight finish for up to grace, and only
// then returns. A nil error means a clean shutdown.
func Serve(ctx context.Context, srv *http.Server, ln net.Listener, grace time.Duration) error {
	errc := make(chan error, 1) // buffered: the goroutine can always send and exit
	go func() { errc <- srv.Serve(ln) }()

	select {
	case err := <-errc:
		return fmt.Errorf("serve %s: %w", ln.Addr(), err) // failed before any shutdown
	case <-ctx.Done():
	}

	// ctx is already cancelled; the grace period needs a context of its own.
	shutCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), grace)
	defer cancel()
	if err := srv.Shutdown(shutCtx); err != nil {
		_ = srv.Close() // grace is over: cut the remaining connections
		return fmt.Errorf("shutdown %s: %w", ln.Addr(), err)
	}
	if err := <-errc; !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("serve %s: %w", ln.Addr(), err)
	}
	return nil
}
