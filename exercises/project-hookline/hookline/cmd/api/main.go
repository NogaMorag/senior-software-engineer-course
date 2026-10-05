// Command api serves Hookline's HTTP API. Configuration comes from HOOKLINE_* environment variables and
// DATABASE_URL; see internal/config. Without DATABASE_URL (dev mode only) it runs on the memory store
// and creates one development tenant, whose API key it prints once to stderr.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"time"

	"golang.org/x/sync/errgroup"

	"example.com/hookline/internal/admin"
	"example.com/hookline/internal/api"
	"example.com/hookline/internal/config"
	"example.com/hookline/internal/platform"
	"example.com/hookline/internal/shed"
	"example.com/hookline/internal/store"
	"example.com/hookline/internal/store/memory"
	"example.com/hookline/internal/store/postgres"
	"example.com/hookline/internal/wire"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "api:", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load(os.LookupEnv)
	if err != nil {
		return err // refuse to start with a bad configuration
	}
	log := platform.NewLogger(os.Stdout, cfg.LogLevel)

	ctx, stop := wire.SignalContext()
	defer stop()

	st, closeStore, err := openStore(ctx, cfg)
	if err != nil {
		return err
	}
	defer closeStore()

	rdb := wire.Redis(cfg.RedisAddr)
	if rdb != nil {
		defer rdb.Close()
	}
	opts := api.Options{
		MaxBodyBytes: cfg.MaxBodyBytes, DevMode: cfg.DevMode, Logger: log,
		// The API never reads the subscription cache; it only drops a tenant's entry after each change.
		Subscriptions: wire.Subscriptions(rdb, st),
		Limiter:       wire.Limiter(rdb), // fails open; the log line is the signal until milestone M8's metric
	}
	if cfg.MaxInFlight > 0 {
		opts.Shedder = shed.New(shed.Options{Limit: cfg.MaxInFlight, RetryAfter: time.Second})
	}
	handler := api.New(st, opts)
	ln, err := net.Listen("tcp", cfg.Addr)
	if err != nil {
		return err
	}
	g, gctx := errgroup.WithContext(ctx)
	log.Info("listening", slog.String("addr", ln.Addr().String()), slog.String("config", cfg.String()))
	g.Go(func() error { return platform.Serve(gctx, platform.NewServer(handler, log), ln, cfg.ShutdownTimeout) })
	if cfg.DebugAddr != "off" {
		debugLn, err := net.Listen("tcp", cfg.DebugAddr)
		if err != nil {
			_ = ln.Close()
			return err
		}
		log.Info("debug listener", slog.String("addr", debugLn.Addr().String()))
		g.Go(func() error {
			return platform.Serve(gctx, platform.NewServer(platform.DebugHandler(), log), debugLn, cfg.ShutdownTimeout)
		})
	}
	err = g.Wait()
	log.Info("stopped", slog.Any("error", err))
	return err
}

// openStore returns the PostgreSQL store when DATABASE_URL is set, otherwise the memory store with one
// development tenant (config.Load allows that only in dev mode).
func openStore(ctx context.Context, cfg config.Config) (store.Store, func(), error) {
	if cfg.DatabaseURL == "" {
		st := memory.New()
		t, key, err := admin.CreateTenant(ctx, st, "dev", admin.DefaultRatePerSec, admin.DefaultBurst, time.Now)
		if err != nil {
			return nil, nil, err
		}
		fmt.Fprintf(os.Stderr, "memory store: development tenant %s, API key (shown once): %s\n", t.ID, key)
		return st, func() {}, nil
	}
	pool, err := wire.Pool(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, nil, err
	}
	return postgres.New(pool), pool.Close, nil
}
