package app

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"strings"

	"golang.org/x/sync/errgroup"

	"example.com/ch24-telemetry/internal/events"
	"example.com/ch24-telemetry/internal/links"
	"example.com/ch24-telemetry/internal/platform"
)

// NewHandler builds the public handler: the routes behind the middleware
// chain. RoutePattern is outermost so every layer sees the matched route;
// otelhttp comes next, so the server span and the request metrics cover
// everything after it and every log line below it carries the trace ID.
// RequestID, AccessLog and Recover follow as in chapter 6. publish may be nil.
func NewHandler(cfg Config, store links.Store, log *slog.Logger, publish links.PublisherFunc) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok\n"))
	})
	links.NewHandler(store, cfg.BaseURL, cfg.MaxBodyBytes, log).WithPublisher(publish).Register(mux)
	return platform.Chain(mux,
		platform.RoutePattern(mux),
		// TODO(chapter 24, section 24.2): add otelhttp.NewMiddleware("shorty", ...)
		// here, with otelhttp.WithFilter that leaves /healthz out.
		platform.RequestID,
		platform.AccessLog(log),
		platform.Recover(log),
	)
}

// Run serves the public API and, unless disabled, the debug listener, until
// ctx is cancelled (main cancels it on SIGINT or SIGTERM). Then both servers
// shut down gracefully. If either server fails, the other is stopped too.
func Run(ctx context.Context, cfg Config, log *slog.Logger) error {
	var publish links.PublisherFunc
	if cfg.KafkaBrokers != "" {
		producer, err := events.NewProducer(strings.Split(cfg.KafkaBrokers, ","), cfg.KafkaTopic, cfg.PublishTimeout)
		if err != nil {
			return err
		}
		defer producer.Close()
		publish = func(ctx context.Context, l links.Link) error {
			return producer.LinkCreated(ctx, events.LinkCreated{Code: l.Code, URL: l.URL, CreatedAt: l.CreatedAt})
		}
	}

	ln, err := net.Listen("tcp", cfg.Addr)
	if err != nil {
		return err
	}
	api := platform.NewServer(NewHandler(cfg, links.NewMemoryStore(), log, publish), log)

	var debugLn net.Listener
	if cfg.DebugAddr != "off" {
		if debugLn, err = net.Listen("tcp", cfg.DebugAddr); err != nil {
			_ = ln.Close()
			return err
		}
	}

	g, gctx := errgroup.WithContext(ctx)
	log.Info("listening", slog.String("addr", ln.Addr().String()), slog.String("base_url", cfg.BaseURL),
		slog.String("kafka_topic", cfg.KafkaTopic), slog.Bool("publishing", publish != nil))
	g.Go(func() error { return platform.Serve(gctx, api, ln, cfg.ShutdownTimeout) })
	if debugLn != nil {
		log.Info("debug listener", slog.String("addr", debugLn.Addr().String()))
		debug := platform.NewServer(platform.DebugHandler(), log)
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
