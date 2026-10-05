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
