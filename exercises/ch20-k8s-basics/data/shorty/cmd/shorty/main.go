// Command shorty is a small URL-shortener service. Configuration comes from
// SHORTY_* environment variables; see internal/app/config.go.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"example.com/ch06-shorty/internal/app"
	"example.com/ch06-shorty/internal/platform"
)

func main() {
	// "shorty healthcheck [url]" probes a running shorty and exits 0 or 1.
	if len(os.Args) > 1 && os.Args[1] == "healthcheck" {
		url := defaultHealthURL
		if len(os.Args) > 2 {
			url = os.Args[2]
		}
		os.Exit(healthcheck(url, os.Stderr))
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
	log := platform.NewLogger(os.Stdout, cfg.LogLevel)

	// The first SIGINT or SIGTERM cancels ctx and starts the graceful shutdown.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	// Then restore the default behaviour, so a second Ctrl-C kills at once.
	context.AfterFunc(ctx, stop)

	return app.Run(ctx, cfg, log)
}
