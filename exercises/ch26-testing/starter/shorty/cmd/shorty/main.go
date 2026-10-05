// Command shorty is a small URL-shortener service. Configuration comes from
// SHORTY_* environment variables; see internal/app/config.go.
// Copied without its tests from exercises/ch06-shorty/solution (imports renamed): the k6 target.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"example.com/ch26-testing/shorty/internal/app"
	"example.com/ch26-testing/shorty/internal/platform"
)

func main() {
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
