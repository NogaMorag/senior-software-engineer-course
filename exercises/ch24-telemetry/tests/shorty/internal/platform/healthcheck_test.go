package platform_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"example.com/ch24-telemetry/internal/platform"
)

func TestHealthcheckExitCode(t *testing.T) {
	ok := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {}))
	defer ok.Close()
	failing := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer failing.Close()
	closed := httptest.NewServer(http.NotFoundHandler())
	closed.Close() // nothing listens there any more

	tests := []struct {
		name string
		url  string
		want int
	}{
		{"200 is healthy", ok.URL, 0},
		{"503 is unhealthy", failing.URL, 1},
		{"no listener is unhealthy", closed.URL, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := platform.Healthcheck(tt.url, io.Discard); got != tt.want {
				t.Errorf("healthcheck(%s) = %d, want %d", tt.url, got, tt.want)
			}
		})
	}
}
