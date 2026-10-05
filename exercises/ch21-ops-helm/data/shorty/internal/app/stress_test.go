package app_test

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"example.com/ch06-shorty/internal/app"
	"example.com/ch06-shorty/internal/links"
)

func TestStressRoutes(t *testing.T) {
	tests := []struct {
		name       string
		stress     bool
		method     string
		path       string
		wantStatus int
		wantBody   string
	}{
		{"memory", true, http.MethodPost, "/stress/memory?mb=1", http.StatusOK, "holding"},
		{"cpu", true, http.MethodGet, "/stress/cpu?work=1", http.StatusOK, "work 1, wall"},
		{"memory out of range", true, http.MethodPost, "/stress/memory?mb=0", http.StatusBadRequest, "between 1 and 1024"},
		{"cpu not a number", true, http.MethodGet, "/stress/cpu?work=lots", http.StatusBadRequest, "between 1 and 1000"},
		{"off by default", false, http.MethodGet, "/stress/cpu?work=1", http.StatusNotFound, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := app.Config{BaseURL: "http://sho.rt", MaxBodyBytes: 4096, Stress: tt.stress}
			srv := httptest.NewServer(app.NewHandler(cfg, links.NewMemoryStore(), slog.New(slog.DiscardHandler)))
			defer srv.Close()
			req, err := http.NewRequest(tt.method, srv.URL+tt.path, nil)
			if err != nil {
				t.Fatal(err)
			}
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			if resp.StatusCode != tt.wantStatus || !strings.Contains(string(body), tt.wantBody) {
				t.Errorf("%s %s = %d %q, want %d containing %q", tt.method, tt.path, resp.StatusCode, body, tt.wantStatus, tt.wantBody)
			}
		})
	}
}
