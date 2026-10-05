// Command projects serves the tenant-scoped projects API behind JWT
// authentication. It is the service the chapter's threat model describes.
//
//	PROJECTS_ADDR       listen address (default 127.0.0.1:8080)
//	PROJECTS_ISSUER     expected "iss" (required)
//	PROJECTS_AUDIENCE   expected "aud" (required)
//	PROJECTS_JWKS_FILE  path to the identity provider's JWKS (required)
package main

import (
	"errors"
	"log/slog"
	"net/http"
	"os"
	"time"

	"example.com/ch28-security/internal/jwtauth"
	"example.com/ch28-security/internal/projects"
	"example.com/ch28-security/internal/redact"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{ReplaceAttr: redact.ReplaceAttr}))
	if err := run(log); err != nil {
		log.Error("exit", slog.Any("err", err))
		os.Exit(1)
	}
}

func run(log *slog.Logger) error {
	issuer, audience, jwksFile := os.Getenv("PROJECTS_ISSUER"), os.Getenv("PROJECTS_AUDIENCE"), os.Getenv("PROJECTS_JWKS_FILE")
	if issuer == "" || audience == "" || jwksFile == "" {
		return errors.New("PROJECTS_ISSUER, PROJECTS_AUDIENCE and PROJECTS_JWKS_FILE are required")
	}
	data, err := os.ReadFile(jwksFile)
	if err != nil {
		return err
	}
	keys, err := jwtauth.ParseJWKS(data)
	if err != nil {
		return err
	}
	verifier, err := jwtauth.NewVerifier(jwtauth.Config{
		Issuer: issuer, Audience: audience, Methods: []string{"RS256", "ES256"},
		Leeway: 30 * time.Second, Keys: keys,
	})
	if err != nil {
		return err
	}

	mux := http.NewServeMux()
	(&projects.Handler{Store: projects.NewMemoryStore(), Log: log}).Register(mux)

	addr := os.Getenv("PROJECTS_ADDR")
	if addr == "" {
		addr = "127.0.0.1:8080"
	}
	srv := &http.Server{
		Addr:              addr,
		Handler:           verifier.Middleware(log)(mux),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    16 << 10,
	}
	log.Info("listening", slog.String("addr", addr))
	return srv.ListenAndServe()
}
