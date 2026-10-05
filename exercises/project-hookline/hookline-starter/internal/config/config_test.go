package config_test

import (
	"strings"
	"testing"
	"time"

	"example.com/hookline/internal/config"
)

func lookup(m map[string]string) func(string) (string, bool) {
	return func(k string) (string, bool) { v, ok := m[k]; return v, ok }
}

func TestLoadDefaultsInDevMode(t *testing.T) {
	cfg, err := config.Load(lookup(map[string]string{"HOOKLINE_DEV_MODE": "true"}))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Addr != ":8080" || cfg.MaxBodyBytes != 64<<10 || cfg.ShutdownTimeout != 10*time.Second || !cfg.DevMode || cfg.DatabaseURL != "" {
		t.Errorf("defaults = %+v", cfg)
	}
}

func TestLoadReportsEveryProblem(t *testing.T) {
	_, err := config.Load(lookup(map[string]string{
		"HOOKLINE_DEV_MODE":       "maybe",
		"HOOKLINE_MAX_BODY_BYTES": "10",
		"HOOKLINE_ADDR":           ":7000",
		"HOOKLINE_DEBUG_ADDR":     ":7000",
	}))
	if err == nil {
		t.Fatal("Load accepted an invalid configuration")
	}
	for _, want := range []string{"HOOKLINE_DEV_MODE", "HOOKLINE_MAX_BODY_BYTES", "HOOKLINE_DEBUG_ADDR", "DATABASE_URL: required"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q does not mention %s", err, want)
		}
	}
}

func TestLoadWithDatabase(t *testing.T) {
	cfg, err := config.Load(lookup(map[string]string{"DATABASE_URL": "postgres://hookline:change-me@localhost:5438/hookline"}))
	if err != nil || cfg.DevMode || cfg.DatabaseURL == "" {
		t.Fatalf("Load = %+v, %v", cfg, err)
	}
}

func TestDeliverySettingsAreChecked(t *testing.T) {
	_, err := config.Load(lookup(map[string]string{
		"HOOKLINE_DEV_MODE":             "true",
		"HOOKLINE_LEASE":                "5s",
		"HOOKLINE_ATTEMPT_TIMEOUT":      "10s",
		"HOOKLINE_MAX_ATTEMPTS":         "0",
		"HOOKLINE_BACKOFF_BASE":         "1m",
		"HOOKLINE_BACKOFF_CAP":          "1s",
		"HOOKLINE_ENDPOINT_CONCURRENCY": "9",
		"HOOKLINE_BREAKER_FAILURES":     "0",
		"HOOKLINE_MAX_IN_FLIGHT":        "-1",
	}))
	for _, want := range []string{"HOOKLINE_LEASE", "HOOKLINE_MAX_ATTEMPTS", "HOOKLINE_BACKOFF_CAP",
		"HOOKLINE_ENDPOINT_CONCURRENCY", "HOOKLINE_BREAKER_FAILURES", "HOOKLINE_MAX_IN_FLIGHT"} {
		if err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("error %v does not mention %s", err, want)
		}
	}
}

func TestRequire(t *testing.T) {
	cfg, err := config.Load(lookup(map[string]string{"DATABASE_URL": "postgres://x@localhost/x"}))
	if err != nil {
		t.Fatal(err)
	}
	if err := cfg.Require("DATABASE_URL"); err != nil {
		t.Errorf("Require(DATABASE_URL) = %v", err)
	}
	if err := cfg.Require("DATABASE_URL", "KAFKA_BROKERS", "REDIS_ADDR"); err == nil || !strings.Contains(err.Error(), "KAFKA_BROKERS, REDIS_ADDR") {
		t.Errorf("Require = %v, want both missing names", err)
	}
	if strings.Contains(cfg.String(), "postgres://") {
		t.Error("String must not print the database URL")
	}
}
