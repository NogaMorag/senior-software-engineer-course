// Package config reads the configuration of the Hookline binaries from HOOKLINE_* environment variables
// (and DATABASE_URL, REDIS_ADDR, KAFKA_BROKERS), with the collect-every-problem Env of chapter 6. One
// Config serves every binary; each main calls Require for the connections it cannot run without.
package config

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"example.com/hookline/internal/platform"
)

// Config is everything the binaries read from the environment.
type Config struct {
	Addr            string        // HOOKLINE_ADDR, public listener
	DebugAddr       string        // HOOKLINE_DEBUG_ADDR, pprof listener; "off" disables it
	LogLevel        slog.Level    // HOOKLINE_LOG_LEVEL
	ShutdownTimeout time.Duration // HOOKLINE_SHUTDOWN_TIMEOUT, grace period for requests in flight
	MaxBodyBytes    int64         // HOOKLINE_MAX_BODY_BYTES, limit for JSON request bodies
	DevMode         bool          // HOOKLINE_DEV_MODE: allows http:// endpoint URLs and the in-memory store
	DatabaseURL     string        // DATABASE_URL; empty means the in-memory store (cmd/api in dev mode only)
	RedisAddr       string        // REDIS_ADDR; empty means no subscription cache (every lookup reads PostgreSQL)
	KafkaBrokers    string        // KAFKA_BROKERS, host:port[,host:port]; the relay and the dispatcher need it

	RelayInterval  time.Duration // HOOKLINE_RELAY_INTERVAL, the relay's idle wait
	ConsumerGroup  string        // HOOKLINE_CONSUMER_GROUP, the dispatchers' consumer group
	Workers        int           // HOOKLINE_WORKERS, delivery goroutines per worker process
	Lease          time.Duration // HOOKLINE_LEASE, how long a claimed delivery stays ours
	AttemptTimeout time.Duration // HOOKLINE_ATTEMPT_TIMEOUT, per HTTP attempt; must be below the lease
	MaxAttempts    int           // HOOKLINE_MAX_ATTEMPTS, per round; then the delivery is dead
	BackoffBase    time.Duration // HOOKLINE_BACKOFF_BASE, ceiling of the first retry's delay
	BackoffCap     time.Duration // HOOKLINE_BACKOFF_CAP, largest delay between attempts
	Poll           time.Duration // HOOKLINE_POLL, a worker's idle wait

	EndpointConcurrency int           // HOOKLINE_ENDPOINT_CONCURRENCY, workers one endpoint may hold (bulkhead)
	BreakerFailures     int           // HOOKLINE_BREAKER_FAILURES, consecutive failures that open an endpoint's breaker
	BreakerOpen         time.Duration // HOOKLINE_BREAKER_OPEN, how long it stays open before a probe
	MaxInFlight         int           // HOOKLINE_MAX_IN_FLIGHT, /v1 requests before shedding; 0 disables
	HealthAddr          string        // HOOKLINE_HEALTH_ADDR, /healthz of relay, dispatcher and worker; "off" disables
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
		RedisAddr:       env.String("REDIS_ADDR", ""),
		KafkaBrokers:    env.String("KAFKA_BROKERS", ""),
		RelayInterval:   env.Duration("HOOKLINE_RELAY_INTERVAL", 200*time.Millisecond),
		ConsumerGroup:   env.String("HOOKLINE_CONSUMER_GROUP", "hookline-dispatcher"),
		Workers:         int(env.Int64("HOOKLINE_WORKERS", 8)),
		Lease:           env.Duration("HOOKLINE_LEASE", 30*time.Second),
		AttemptTimeout:  env.Duration("HOOKLINE_ATTEMPT_TIMEOUT", 10*time.Second),
		MaxAttempts:     int(env.Int64("HOOKLINE_MAX_ATTEMPTS", 10)),
		BackoffBase:     env.Duration("HOOKLINE_BACKOFF_BASE", time.Second),
		BackoffCap:      env.Duration("HOOKLINE_BACKOFF_CAP", 15*time.Minute),
		Poll:            env.Duration("HOOKLINE_POLL", 500*time.Millisecond),

		EndpointConcurrency: int(env.Int64("HOOKLINE_ENDPOINT_CONCURRENCY", 2)),
		BreakerFailures:     int(env.Int64("HOOKLINE_BREAKER_FAILURES", 5)),
		BreakerOpen:         env.Duration("HOOKLINE_BREAKER_OPEN", 30*time.Second),
		MaxInFlight:         int(env.Int64("HOOKLINE_MAX_IN_FLIGHT", 256)),
		HealthAddr:          env.String("HOOKLINE_HEALTH_ADDR", "off"),
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
	if cfg.Workers < 1 || cfg.Workers > 1000 {
		env.Fail("HOOKLINE_WORKERS", "must be between 1 and 1000, got %d", cfg.Workers)
	}
	if cfg.MaxAttempts < 1 {
		env.Fail("HOOKLINE_MAX_ATTEMPTS", "must be at least 1, got %d", cfg.MaxAttempts)
	}
	if cfg.AttemptTimeout <= 0 || cfg.Lease <= cfg.AttemptTimeout {
		env.Fail("HOOKLINE_LEASE", "must be longer than HOOKLINE_ATTEMPT_TIMEOUT (%s), got %s", cfg.AttemptTimeout, cfg.Lease)
	}
	if cfg.BackoffBase <= 0 || cfg.BackoffCap < cfg.BackoffBase {
		env.Fail("HOOKLINE_BACKOFF_CAP", "must be at least HOOKLINE_BACKOFF_BASE (%s), got %s", cfg.BackoffBase, cfg.BackoffCap)
	}
	if cfg.EndpointConcurrency < 1 || cfg.EndpointConcurrency > cfg.Workers {
		env.Fail("HOOKLINE_ENDPOINT_CONCURRENCY", "must be between 1 and HOOKLINE_WORKERS (%d), got %d", cfg.Workers, cfg.EndpointConcurrency)
	}
	if cfg.BreakerFailures < 1 {
		env.Fail("HOOKLINE_BREAKER_FAILURES", "must be at least 1, got %d", cfg.BreakerFailures)
	}
	if cfg.MaxInFlight < 0 {
		env.Fail("HOOKLINE_MAX_IN_FLIGHT", "must be 0 (off) or positive, got %d", cfg.MaxInFlight)
	}
	for key, d := range map[string]time.Duration{"HOOKLINE_RELAY_INTERVAL": cfg.RelayInterval, "HOOKLINE_POLL": cfg.Poll, "HOOKLINE_BREAKER_OPEN": cfg.BreakerOpen} {
		if d <= 0 {
			env.Fail(key, "must be positive, got %s", d)
		}
	}
	if cfg.DatabaseURL == "" && !cfg.DevMode {
		env.Fail("DATABASE_URL", "required unless HOOKLINE_DEV_MODE=true (the in-memory store is for development)")
	}
	return cfg, env.Err()
}

// Require reports which of the named connection settings are empty. The relay, the dispatcher and the
// worker call it, since unlike cmd/api they have no in-memory mode.
func (c Config) Require(names ...string) error {
	values := map[string]string{"DATABASE_URL": c.DatabaseURL, "REDIS_ADDR": c.RedisAddr, "KAFKA_BROKERS": c.KafkaBrokers}
	var missing []string
	for _, n := range names {
		if v, ok := values[n]; !ok || v == "" {
			missing = append(missing, n)
		}
	}
	if len(missing) > 0 {
		return errors.New("required, not set: " + strings.Join(missing, ", "))
	}
	return nil
}

// String summarises the configuration for the start-up log line, without the database password.
func (c Config) String() string {
	return fmt.Sprintf("addr=%s dev_mode=%t redis=%t kafka=%s workers=%d lease=%s timeout=%s max_attempts=%d "+
		"endpoint_concurrency=%d breaker=%d/%s max_in_flight=%d",
		c.Addr, c.DevMode, c.RedisAddr != "", c.KafkaBrokers, c.Workers, c.Lease, c.AttemptTimeout, c.MaxAttempts,
		c.EndpointConcurrency, c.BreakerFailures, c.BreakerOpen, c.MaxInFlight)
}
