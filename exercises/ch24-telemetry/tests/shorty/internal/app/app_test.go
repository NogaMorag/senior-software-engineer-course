package app_test

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"example.com/ch24-telemetry/internal/app"
	"example.com/ch24-telemetry/internal/links"
)

func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	cfg := app.Config{BaseURL: "http://sho.rt", MaxBodyBytes: 4096}
	srv := httptest.NewServer(app.NewHandler(cfg, links.NewMemoryStore(), slog.New(slog.DiscardHandler), nil))
	t.Cleanup(srv.Close)
	return srv
}

// noRedirect returns a client that reports a 302 instead of following it.
func noRedirect() *http.Client {
	return &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
}

func TestCreateResolveStatsOverHTTP(t *testing.T) {
	srv := newTestServer(t)
	client := noRedirect()

	resp, err := client.Post(srv.URL+"/v1/links", "application/json", strings.NewReader(`{"url": "https://go.dev/", "code": "gohome"}`))
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusCreated || resp.Header.Get("X-Request-ID") == "" {
		t.Fatalf("create = %d, request id %q", resp.StatusCode, resp.Header.Get("X-Request-ID"))
	}

	resp, err = client.Get(srv.URL + "/gohome")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusFound || resp.Header.Get("Location") != "https://go.dev/" {
		t.Fatalf("resolve = %d %q", resp.StatusCode, resp.Header.Get("Location"))
	}

	resp, err = client.Get(srv.URL + "/v1/links/gohome")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var stats struct {
		Hits     int64  `json:"hits"`
		ShortURL string `json:"short_url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil || stats.Hits != 1 || stats.ShortURL != "http://sho.rt/gohome" {
		t.Errorf("stats = %+v, %v", stats, err)
	}
}

func TestHealthzAndNoPprofOnPublicPort(t *testing.T) {
	srv := newTestServer(t)
	resp, err := http.Get(srv.URL + "/healthz")
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK || string(body) != "ok\n" {
		t.Errorf("healthz = %d %q", resp.StatusCode, body)
	}

	resp, err = http.Get(srv.URL + "/debug/pprof/")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("public /debug/pprof/ = %d, want 404: profiles belong on the debug listener only", resp.StatusCode)
	}
}
