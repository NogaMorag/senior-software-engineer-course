// Command catalog serves the catalog over HTTP (the public contract) and gRPC (internal callers).
package main

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"

	catalogv1 "example.com/ch09-catalog-api/gen/catalog/v1"
	"example.com/ch09-catalog-api/internal/catalog"
	"example.com/ch09-catalog-api/internal/config"
	"example.com/ch09-catalog-api/internal/grpcapi"
	"example.com/ch09-catalog-api/internal/httpapi"
	"example.com/ch09-catalog-api/internal/idem"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	if err := run(log); err != nil {
		log.Error("exit", "err", err)
		os.Exit(1)
	}
}

func run(log *slog.Logger) error {
	cfg, err := config.Load(nil)
	if err != nil {
		return err
	}
	store := catalog.NewMemory(nil)
	keys := idem.NewStore(cfg.IdempotencyTTL, nil)

	httpSrv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           httpapi.New(store, keys, httpapi.Options{MaxBodyBytes: cfg.MaxBodyBytes, Logger: log}),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		MaxHeaderBytes:    16 << 10,
	}
	grpcSrv := grpc.NewServer(grpc.MaxRecvMsgSize(64 << 10))
	catalogv1.RegisterCatalogServiceServer(grpcSrv, &grpcapi.Server{Store: store})
	lis, err := net.Listen("tcp", cfg.GRPCAddr)
	if err != nil {
		return err
	}

	errc := make(chan error, 2)
	go func() { errc <- httpSrv.ListenAndServe() }()
	go func() { errc <- grpcSrv.Serve(lis) }()
	log.Info("listening", "http", cfg.HTTPAddr, "grpc", cfg.GRPCAddr)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	select {
	case err := <-errc:
		return err
	case <-ctx.Done():
	}
	log.Info("shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	grpcSrv.GracefulStop()
	if err := httpSrv.Shutdown(shutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}
