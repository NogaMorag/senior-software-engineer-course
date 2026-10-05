// Command shorty is a small URL-shortener service. Configuration comes from
// SHORTY_* environment variables (see internal/app/config.go) and, for
// telemetry, from the standard OTEL_* variables (see internal/telemetry).
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.opentelemetry.io/contrib/bridges/otelslog"

	"example.com/ch24-telemetry/internal/app"
	"example.com/ch24-telemetry/internal/platform"
	"example.com/ch24-telemetry/internal/telemetry"
)

// version is set at build time: go build -ldflags "-X main.version=..."
var version = "dev"

func main() {
	// "shorty healthcheck [url]" probes a running shorty and exits 0 or 1.
	if len(os.Args) > 1 && os.Args[1] == "healthcheck" {
		url := "http://127.0.0.1:8080/healthz"
		if len(os.Args) > 2 {
			url = os.Args[2]
		}
		os.Exit(platform.Healthcheck(url, os.Stderr))
	}
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "shorty:", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := app.LoadConfig(os.LookupEnv)
	if err != nil {
		return err // refuse to start with a bad configuration
	}

	// The first SIGINT or SIGTERM cancels ctx and starts the graceful shutdown.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	// Then restore the default behaviour, so a second Ctrl-C kills at once.
	context.AfterFunc(ctx, stop)

	// Telemetry first: the logger, otelhttp and kotel all use the globals it sets.
	shutdown, err := telemetry.Setup(ctx, "shorty", version)
	if err != nil {
		return err
	}
	defer func() {
		// ctx is cancelled by now; the flush gets a few seconds of its own.
		flushCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := shutdown(flushCtx); err != nil {
			fmt.Fprintln(os.Stderr, "shorty: telemetry shutdown:", err)
		}
	}()

	// Every log line goes to stdout as JSON and, through the bridge, to the collector.
	log := platform.NewLogger(os.Stdout, cfg.LogLevel, otelslog.NewHandler("shorty"))
	return app.Run(ctx, cfg, log)
}
