package app_test

import (
	"log/slog"
	"strings"
	"testing"
	"time"

	"example.com/ch06-shorty/internal/app"
)

func lookupFrom(m map[string]string) func(string) (string, bool) {
	return func(k string) (string, bool) {
		v, ok := m[k]
		return v, ok
	}
}

func TestLoadConfigDefaults(t *testing.T) {
	cfg, err := app.LoadConfig(lookupFrom(map[string]string{"SHORTY_BASE_URL": "http://localhost:8080/"}))
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	want := app.Config{
		Addr:            ":8080",
		DebugAddr:       "127.0.0.1:6060",
		BaseURL:         "http://localhost:8080", // trailing slash removed
		LogLevel:        slog.LevelInfo,
		ShutdownTimeout: 10 * time.Second,
		MaxBodyBytes:    4096,
	}
	if cfg != want {
		t.Errorf("cfg = %+v\nwant  %+v", cfg, want)
	}
}

func TestLoadConfigErrors(t *testing.T) {
	tests := []struct {
		name string
		env  map[string]string
		want []string // every one must appear in the error
	}{
		{"base url missing", map[string]string{}, []string{"SHORTY_BASE_URL: required, not set"}},
		{"base url relative", map[string]string{"SHORTY_BASE_URL": "localhost:8080"}, []string{"SHORTY_BASE_URL", "absolute"}},
		{"several at once", map[string]string{
			"SHORTY_SHUTDOWN_TIMEOUT": "ten",
			"SHORTY_LOG_LEVEL":        "chatty",
			"SHORTY_MAX_BODY_BYTES":   "0",
		}, []string{"SHORTY_BASE_URL", "SHORTY_SHUTDOWN_TIMEOUT", "SHORTY_LOG_LEVEL", "SHORTY_MAX_BODY_BYTES"}},
		{"negative timeout", map[string]string{"SHORTY_BASE_URL": "http://x", "SHORTY_SHUTDOWN_TIMEOUT": "-1s"}, []string{"must be positive"}},
		{"same address twice", map[string]string{"SHORTY_BASE_URL": "http://x", "SHORTY_ADDR": ":9000", "SHORTY_DEBUG_ADDR": ":9000"}, []string{"must differ"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := app.LoadConfig(lookupFrom(tt.env))
			if err == nil {
				t.Fatal("LoadConfig = nil error, want a refusal")
			}
			for _, w := range tt.want {
				if !strings.Contains(err.Error(), w) {
					t.Errorf("error %q does not mention %q", err, w)
				}
			}
		})
	}
}
