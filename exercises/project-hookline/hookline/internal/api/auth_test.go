package api_test

import (
	"net/http"
	"testing"
)

func TestHealthNeedsNoKey(t *testing.T) {
	hs := newHarness(t, false)
	rec := hs.do(req{Method: "GET", Path: "/healthz"})
	if rec.Code != http.StatusOK || decode(t, rec)["status"] != "ok" {
		t.Fatalf("GET /healthz = %d %s", rec.Code, rec.Body)
	}
	if rec.Header().Get("X-Request-ID") == "" {
		t.Error("every response carries X-Request-ID")
	}
}

func TestAuthentication(t *testing.T) {
	hs := newHarness(t, false)
	key := hs.tenant("Acme")
	tests := []struct {
		name   string
		header string
		status int
	}{
		{"no header", "", http.StatusUnauthorized},
		{"wrong scheme", "Basic " + key, http.StatusUnauthorized},
		{"not a key", "Bearer hello", http.StatusUnauthorized},
		{"well-formed but unknown", "Bearer hk_AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA", http.StatusUnauthorized},
		{"valid", "Bearer " + key, http.StatusOK},
		{"scheme is case-insensitive", "bearer " + key, http.StatusOK},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := req{Method: "GET", Path: "/v1/tenant"}
			if tt.header != "" {
				r.Header = map[string]string{"Authorization": tt.header}
			}
			rec := hs.do(r)
			if tt.status == http.StatusUnauthorized {
				wantProblem(t, rec, http.StatusUnauthorized, "unauthorized")
				if rec.Header().Get("WWW-Authenticate") == "" {
					t.Error("a 401 must carry WWW-Authenticate")
				}
				return
			}
			if rec.Code != tt.status {
				t.Fatalf("status = %d, want %d: %s", rec.Code, tt.status, rec.Body)
			}
		})
	}
}

func TestGetTenant(t *testing.T) {
	hs := newHarness(t, false)
	key := hs.tenant("Acme")
	body := decode(t, hs.do(req{Method: "GET", Path: "/v1/tenant", Key: key}))
	if body["name"] != "Acme" || body["rate_per_sec"] != float64(50) || body["burst"] != float64(100) || body["id"] == nil {
		t.Fatalf("GET /v1/tenant = %v", body)
	}
	if _, leaked := body["api_key_hash"]; leaked {
		t.Error("the key hash must never leave the server")
	}
}

func TestRoutingErrorsAreProblems(t *testing.T) {
	hs := newHarness(t, false)
	wantProblem(t, hs.do(req{Method: "GET", Path: "/v2/nothing"}), http.StatusNotFound, "not-found")
	rec := hs.do(req{Method: "PUT", Path: "/v1/endpoints"})
	wantProblem(t, rec, http.StatusMethodNotAllowed, "method-not-allowed")
	if rec.Header().Get("Allow") != "GET, POST" {
		t.Errorf("Allow = %q", rec.Header().Get("Allow"))
	}
	wantProblem(t, hs.do(req{Method: "DELETE", Path: "/v1/events/x"}), http.StatusMethodNotAllowed, "method-not-allowed")
}
