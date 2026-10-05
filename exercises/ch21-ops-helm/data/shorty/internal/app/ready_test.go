package app_test

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"example.com/ch06-shorty/internal/app"
	"example.com/ch06-shorty/internal/links"
)

func TestReadinessSwitch(t *testing.T) {
	var ready app.Readiness
	mux := http.NewServeMux()
	mux.Handle("GET /readyz", &ready)
	mux.HandleFunc("POST /debug/ready", ready.Toggle)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	steps := []struct {
		name   string
		method string
		path   string
		want   int
	}{
		{"ready at start", http.MethodGet, "/readyz", http.StatusOK},
		{"turn off", http.MethodPost, "/debug/ready?state=off", http.StatusServiceUnavailable},
		{"not ready", http.MethodGet, "/readyz", http.StatusServiceUnavailable},
		{"bad state", http.MethodPost, "/debug/ready?state=maybe", http.StatusBadRequest},
		{"turn on", http.MethodPost, "/debug/ready?state=on", http.StatusOK},
		{"ready again", http.MethodGet, "/readyz", http.StatusOK},
	}
	for _, s := range steps {
		req, _ := http.NewRequest(s.method, srv.URL+s.path, nil)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		if resp.StatusCode != s.want {
			t.Errorf("%s: %s %s = %d, want %d", s.name, s.method, s.path, resp.StatusCode, s.want)
		}
	}
}

func TestPublicHandlerServesReadyz(t *testing.T) {
	cfg := app.Config{BaseURL: "http://sho.rt", MaxBodyBytes: 4096}
	srv := httptest.NewServer(app.NewHandler(cfg, links.NewMemoryStore(), slog.New(slog.DiscardHandler)))
	defer srv.Close()
	resp, err := http.Get(srv.URL + "/readyz")
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK || string(body) != "ready\n" {
		t.Errorf("readyz = %d %q", resp.StatusCode, body)
	}
}
