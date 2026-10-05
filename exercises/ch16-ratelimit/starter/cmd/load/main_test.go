package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLoadCountsStatuses(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Tenant") != "a" {
			t.Errorf("header X-Tenant = %q, want a", r.Header.Get("X-Tenant"))
		}
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer srv.Close()
	var out bytes.Buffer
	if code := run([]string{"-url", srv.URL, "-rate", "200", "-for", "200ms", "-H", "X-Tenant: a"}, &out); code != 0 {
		t.Fatalf("exit code %d", code)
	}
	if !strings.Contains(out.String(), "429 ") || !strings.Contains(out.String(), "p99") {
		t.Fatalf("report lacks the 429 row:\n%s", out.String())
	}
}
