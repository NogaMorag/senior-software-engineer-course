// Package config reads the configuration of the Hookline binaries from HOOKLINE_* environment variables
// (and DATABASE_URL), with the collect-every-problem Env of chapter 6.
package config

import (
	"log/slog"
	"time"

	"example.com/hookline/internal/platform"
)

// Config is everything cmd/api reads from the environment.
type Config struct {
	Addr            string        // HOOKLINE_ADDR, public listener
	DebugAddr       string        // HOOKLINE_DEBUG_ADDR, pprof listener; "off" disables it
	LogLevel        slog.Level    // HOOKLINE_LOG_LEVEL
	ShutdownTimeout time.Duration // HOOKLINE_SHUTDOWN_TIMEOUT, grace period for requests in flight
	MaxBodyBytes    int64         // HOOKLINE_MAX_BODY_BYTES, limit for JSON request bodies
	DevMode         bool          // HOOKLINE_DEV_MODE: allows http:// endpoint URLs and the in-memory store
	DatabaseURL     string        // DATABASE_URL; empty means the in-memory store (dev mode only)
}

// Load reads and validates the configuration. It reports every problem at once; main refuses to start
// when the error is not nil.
func Load(lookup func(string) (string, bool)) (Config, error) {
	env := platform.NewEnv(lookup)
	cfg := Config{
		Addr:            env.String("HOOKLINE_ADDR", ":8080"),
		DebugAddr:       env.String("HOOKLINE_DEBUG_ADDR", "127.0.0.1:6060"),
		LogLevel:        env.Level("HOOKLINE_LOG_LEVEL", slog.LevelInfo),
		ShutdownTimeout: env.Duration("HOOKLINE_SHUTDOWN_TIMEOUT", 10*time.Second),
		MaxBodyBytes:    env.Int64("HOOKLINE_MAX_BODY_BYTES", 64<<10),
		DatabaseURL:     env.String("DATABASE_URL", ""),
	}
	switch v := env.String("HOOKLINE_DEV_MODE", "false"); v {
	case "true", "1":
		cfg.DevMode = true
	case "false", "0":
	default:
		env.Fail("HOOKLINE_DEV_MODE", "%q is not true or false", v)
	}
	if cfg.ShutdownTimeout <= 0 {
		env.Fail("HOOKLINE_SHUTDOWN_TIMEOUT", "must be positive, got %s", cfg.ShutdownTimeout)
	}
	if cfg.MaxBodyBytes < 1<<10 || cfg.MaxBodyBytes > 1<<20 {
		env.Fail("HOOKLINE_MAX_BODY_BYTES", "must be between 1024 and 1048576, got %d", cfg.MaxBodyBytes)
	}
	if cfg.DebugAddr == cfg.Addr {
		env.Fail("HOOKLINE_DEBUG_ADDR", "must differ from HOOKLINE_ADDR (%s)", cfg.Addr)
	}
	if cfg.DatabaseURL == "" && !cfg.DevMode {
		env.Fail("DATABASE_URL", "required unless HOOKLINE_DEV_MODE=true (the in-memory store is for development)")
	}
	return cfg, env.Err()
}
