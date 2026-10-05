// Package app wires shorty together: configuration, routes, middleware and
// the two listeners (public API and private debug).
package app

import (
	"log/slog"
	"strings"
	"time"

	"example.com/ch06-shorty/internal/platform"
)

// Config is everything shorty reads from the environment.
type Config struct {
	Addr            string        // SHORTY_ADDR, public listener
	DebugAddr       string        // SHORTY_DEBUG_ADDR, pprof listener; "off" disables it
	BaseURL         string        // SHORTY_BASE_URL, required: used to build short_url
	LogLevel        slog.Level    // SHORTY_LOG_LEVEL
	ShutdownTimeout time.Duration // SHORTY_SHUTDOWN_TIMEOUT, grace period for requests in flight
	MaxBodyBytes    int64         // SHORTY_MAX_BODY_BYTES, limit for JSON request bodies
}

// LoadConfig reads and validates the configuration. It reports every problem
// at once; main refuses to start when the error is not nil.
func LoadConfig(lookup func(string) (string, bool)) (Config, error) {
	env := platform.NewEnv(lookup)
	cfg := Config{
		Addr:            env.String("SHORTY_ADDR", ":8080"),
		DebugAddr:       env.String("SHORTY_DEBUG_ADDR", "127.0.0.1:6060"),
		BaseURL:         strings.TrimSuffix(env.Required("SHORTY_BASE_URL"), "/"),
		LogLevel:        env.Level("SHORTY_LOG_LEVEL", slog.LevelInfo),
		ShutdownTimeout: env.Duration("SHORTY_SHUTDOWN_TIMEOUT", 10*time.Second),
		MaxBodyBytes:    env.Int64("SHORTY_MAX_BODY_BYTES", 4096),
	}

	// TODO 4 (day 11, section 6.3): validation that needs more than one value,
	// each failure recorded with env.Fail:
	//   - a set BaseURL must parse as an absolute http or https URL;
	//   - ShutdownTimeout must be positive;
	//   - MaxBodyBytes must be between 1 and 1<<20;
	//   - DebugAddr must differ from Addr.

	return cfg, env.Err()
}
