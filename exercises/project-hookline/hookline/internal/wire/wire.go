// Package wire builds the shared connections every binary needs, so the main functions stay short:
// the PostgreSQL pool, the Redis client, the subscription cache and the ingest limiter over it (or their
// in-process stand-ins without Redis), and the signal context.
package wire

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/contrib/bridges/otelslog"

	"example.com/hookline/internal/cache"
	"example.com/hookline/internal/db"
	"example.com/hookline/internal/platform"
	"example.com/hookline/internal/ratelimit"
	"example.com/hookline/internal/subcache"
	"example.com/hookline/internal/telemetry"
)

// SignalContext is cancelled by the first SIGINT or SIGTERM, which starts the graceful shutdown; a
// second signal kills the process at once (chapter 6, section 6.5).
func SignalContext() (context.Context, context.CancelFunc) {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	context.AfterFunc(ctx, stop)
	return ctx, stop
}

// Pool opens the PostgreSQL pool with the limits of internal/db.
func Pool(ctx context.Context, url string) (*pgxpool.Pool, error) {
	cfg, err := db.Config(url)
	if err != nil {
		return nil, err
	}
	return db.Connect(ctx, cfg)
}

// Redis returns a client for addr, or nil when addr is empty (no Redis configured). The client
// connects lazily; the first command that cannot reach Redis fails, and its caller decides.
func Redis(addr string) *redis.Client {
	if addr == "" {
		return nil
	}
	return redis.NewClient(&redis.Options{Addr: addr})
}

// Subscriptions returns the subscription cache: over rdb, or without a cache when rdb is nil.
func Subscriptions(rdb *redis.Client, loader subcache.Loader) *subcache.Cache {
	opts := cache.Options{TTL: subcache.DefaultTTL, Jitter: subcache.DefaultJitter}
	if rdb == nil {
		return subcache.New(subcache.Disabled(), loader, opts)
	}
	return subcache.New(cache.NewRedisStore(rdb), loader, opts)
}

// Limiter returns the per-tenant ingest limiter: one token bucket per tenant in Redis, shared by every
// API replica; or, without Redis (dev mode), one in this process, which replicas would multiply.
func Limiter(rdb *redis.Client) ratelimit.Limiter {
	if rdb == nil {
		return &ratelimit.Local{}
	}
	return &ratelimit.Redis{Client: rdb, Prefix: "hookline:rl:ingest:"}
}

// Telemetry starts OpenTelemetry for service when a collector is configured (telemetry.Start) and
// returns the logger every binary uses: JSON lines on stdout and, with a collector, the same records
// through the OpenTelemetry bridge to Loki (chapter 24). Call the returned flush on the way out.
func Telemetry(ctx context.Context, service string, level slog.Leveler) (*slog.Logger, func(), error) {
	shutdown, err := telemetry.Start(ctx, service)
	if err != nil {
		return nil, nil, err
	}
	flush := func() {
		flushCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = shutdown(flushCtx)
	}
	if !telemetry.Enabled() {
		return platform.NewLogger(os.Stdout, level), flush, nil
	}
	return platform.NewLogger(os.Stdout, level, otelslog.NewHandler(service)), flush, nil
}

// Health serves GET /healthz on addr until ctx is done, for the probes of the relay, the dispatcher
// and the worker (milestone M8); "off" or "" serves nothing. Liveness only: the process answers. The
// metrics leave by OTLP push, not from a scrape endpoint (chapter 24).
func Health(ctx context.Context, addr string, log *slog.Logger) error {
	if addr == "" || addr == "off" {
		<-ctx.Done()
		return nil
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte("ok\n")) })
	log.Info("health listener", slog.String("addr", ln.Addr().String()))
	return platform.Serve(ctx, platform.NewServer(mux, log), ln, 5*time.Second)
}
