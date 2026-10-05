package platform

import (
	"context"
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
	// TODO 8 (day 12, section 6.5):
	//   - run srv.Serve(ln) in a goroutine that sends its error on a buffered channel;
	//   - wait for that error (return it wrapped) or for ctx.Done();
	//   - then srv.Shutdown with a context.WithTimeout(context.WithoutCancel(ctx), grace);
	//     if Shutdown fails, srv.Close() and return the wrapped error;
	//   - finally read the Serve error: http.ErrServerClosed means a clean stop (nil).
	return srv.Serve(ln) // until then: ctx is ignored and Ctrl-C kills requests in flight
}
