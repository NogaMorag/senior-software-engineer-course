package app

import (
	"context"
	"log/slog"
	"net"
	"net/http"

	"golang.org/x/sync/errgroup"

	"example.com/ch06-shorty/internal/links"
	"example.com/ch06-shorty/internal/platform"
)

// NewHandler builds the public handler: the routes behind the middleware
// chain. RequestID is outermost so every later log line carries the ID;
// Recover is innermost so AccessLog still records the 500 it produces.
func NewHandler(cfg Config, store links.Store, log *slog.Logger) http.Handler {
	return newHandler(cfg, store, log, &Readiness{})
}

// newHandler is NewHandler with the Readiness that Run shares with the debug
// listener. /healthz says "the process works"; /readyz says "send me traffic".
func newHandler(cfg Config, store links.Store, log *slog.Logger, ready *Readiness) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok\n"))
	})
	mux.Handle("GET /readyz", ready)
	links.NewHandler(store, cfg.BaseURL, cfg.MaxBodyBytes, log).Register(mux)
	return platform.Chain(mux,
		platform.RequestID,
		platform.AccessLog(log),
		platform.Recover(log),
	)
}

// Run serves the public API and, unless disabled, the debug listener, until
// ctx is cancelled (main cancels it on SIGINT or SIGTERM). Then both servers
// shut down gracefully. If either server fails, the other is stopped too.
func Run(ctx context.Context, cfg Config, log *slog.Logger) error {
	ln, err := net.Listen("tcp", cfg.Addr)
	if err != nil {
		return err
	}
	ready := &Readiness{}
	api := platform.NewServer(newHandler(cfg, links.NewMemoryStore(), log, ready), log)

	var debugLn net.Listener
	if cfg.DebugAddr != "off" {
		if debugLn, err = net.Listen("tcp", cfg.DebugAddr); err != nil {
			_ = ln.Close()
			return err
		}
	}

	g, gctx := errgroup.WithContext(ctx)
	log.Info("listening", slog.String("addr", ln.Addr().String()), slog.String("base_url", cfg.BaseURL))
	g.Go(func() error { return platform.Serve(gctx, api, ln, cfg.ShutdownTimeout) })
	if debugLn != nil {
		log.Info("debug listener", slog.String("addr", debugLn.Addr().String()))
		dmux := http.NewServeMux()
		dmux.Handle("/", platform.DebugHandler())
		dmux.HandleFunc("POST /debug/ready", ready.Toggle)
		debug := platform.NewServer(dmux, log)
		g.Go(func() error { return platform.Serve(gctx, debug, debugLn, cfg.ShutdownTimeout) })
	}
	g.Go(func() error {
		<-gctx.Done()
		log.Info("shutting down", slog.Any("cause", context.Cause(gctx)))
		return nil
	})

	err = g.Wait()
	log.Info("stopped", slog.Any("error", err))
	return err
}
