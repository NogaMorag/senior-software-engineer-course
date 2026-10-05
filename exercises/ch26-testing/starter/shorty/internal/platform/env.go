// Package platform holds the service plumbing that does not depend on what
// the service does: configuration from the environment, structured logging,
// middleware, JSON helpers, the HTTP server with graceful shutdown, and the
// private debug (pprof) handler.
package platform

import (
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"time"
)

// Env reads typed settings from environment variables and collects every
// problem it finds, so a bad configuration is reported in one go instead of
// one restart per mistake. An empty value counts as not set.
type Env struct {
	lookup func(string) (string, bool)
	errs   []error
}

// NewEnv returns an Env that reads through lookup (os.LookupEnv in main, a
// map in tests).
func NewEnv(lookup func(string) (string, bool)) *Env {
	return &Env{lookup: lookup}
}

func (e *Env) get(key string) (string, bool) {
	v, ok := e.lookup(key)
	return v, ok && v != ""
}

// Fail records a problem with key; LoadConfig functions use it for checks
// that involve more than one value.
func (e *Env) Fail(key, format string, args ...any) {
	e.errs = append(e.errs, fmt.Errorf("%s: %s", key, fmt.Sprintf(format, args...)))
}

// String returns the value of key, or def when it is not set.
func (e *Env) String(key, def string) string {
	if v, ok := e.get(key); ok {
		return v
	}
	return def
}

// Required returns the value of key and records a problem when it is not set.
func (e *Env) Required(key string) string {
	v, ok := e.get(key)
	if !ok {
		e.Fail(key, "required, not set")
	}
	return v
}

// Int64 returns key parsed as a base-10 integer, or def when it is not set.
func (e *Env) Int64(key string, def int64) int64 {
	v, ok := e.get(key)
	if !ok {
		return def
	}
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		e.Fail(key, "%q is not an integer", v)
		return def
	}
	return n
}

// Duration returns key parsed by time.ParseDuration, or def when it is not set.
func (e *Env) Duration(key string, def time.Duration) time.Duration {
	v, ok := e.get(key)
	if !ok {
		return def
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		e.Fail(key, "%q is not a duration (example: 10s)", v)
		return def
	}
	return d
}

// Level returns key parsed as a slog level (debug, info, warn, error), or def.
func (e *Env) Level(key string, def slog.Level) slog.Level {
	v, ok := e.get(key)
	if !ok {
		return def
	}
	var l slog.Level
	if err := l.UnmarshalText([]byte(v)); err != nil {
		e.Fail(key, "%q is not a log level (debug, info, warn, error)", v)
		return def
	}
	return l
}

// Err returns all recorded problems joined into one error, or nil.
func (e *Env) Err() error {
	if len(e.errs) == 0 {
		return nil
	}
	return fmt.Errorf("invalid configuration:\n%w", errors.Join(e.errs...))
}
