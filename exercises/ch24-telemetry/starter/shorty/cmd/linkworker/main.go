// Command linkworker consumes links.created and posts every new link to the
// receiver. Configuration comes from LINKWORKER_* and OTEL_* variables.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"golang.org/x/sync/errgroup"

	"example.com/ch24-telemetry/internal/events"
	"example.com/ch24-telemetry/internal/linkworker"
	"example.com/ch24-telemetry/internal/platform"
	"example.com/ch24-telemetry/internal/telemetry"
)

var version = "dev"

func main() {
	if len(os.Args) > 1 && os.Args[1] == "healthcheck" {
		os.Exit(platform.Healthcheck("http://127.0.0.1:8082/healthz", os.Stderr))
	}
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "linkworker:", err)
		os.Exit(1)
	}
}

func run() error {
	env := platform.NewEnv(os.LookupEnv)
	brokers := env.Required("LINKWORKER_KAFKA_BROKERS")
	topic := env.String("LINKWORKER_TOPIC", "links.created")
	group := env.String("LINKWORKER_GROUP", "linkworker")
	receiverURL := env.Required("LINKWORKER_RECEIVER_URL")
	secret := env.Required("LINKWORKER_SECRET")
	addr := env.String("LINKWORKER_ADDR", ":8082")
	level := env.Level("LINKWORKER_LOG_LEVEL", slog.LevelInfo)
	if err := env.Err(); err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	context.AfterFunc(ctx, stop)

	shutdown, err := telemetry.Setup(ctx, "linkworker", version)
	if err != nil {
		return err
	}
	defer func() {
		flushCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = shutdown(flushCtx)
	}()
	log := platform.NewLogger(os.Stdout, level, otelslog.NewHandler("linkworker"))

	hooks, tracer := events.Hooks()
	client, err := kgo.NewClient(
		kgo.SeedBrokers(strings.Split(brokers, ",")...),
		kgo.ConsumerGroup(group),
		kgo.ConsumeTopics(topic),
		hooks,
	)
	if err != nil {
		return err
	}
	defer client.Close()

	worker, err := linkworker.New(tracer, receiverURL, secret, log)
	if err != nil {
		return err
	}

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte("ok\n")) })
	health := platform.NewServer(mux, log)

	log.Info("consuming", slog.String("topic", topic), slog.String("group", group), slog.String("receiver", receiverURL))
	g, gctx := errgroup.WithContext(ctx)
	g.Go(func() error { return platform.Serve(gctx, health, ln, 5*time.Second) })
	g.Go(func() error { return worker.Run(gctx, client) })
	err = g.Wait()
	log.Info("stopped", slog.Any("error", err))
	return err
}
