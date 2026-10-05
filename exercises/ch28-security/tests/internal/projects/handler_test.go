package projects_test

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"example.com/ch28-security/internal/jwtauth"
	"example.com/ch28-security/internal/projects"
)

func newServer() http.Handler {
	store := projects.NewMemoryStore(
		projects.Project{ID: "prj_a1", TenantID: "acme", Name: "Acme roadmap"},
		projects.Project{ID: "prj_a2", TenantID: "acme", Name: "Acme hiring"},
		projects.Project{ID: "prj_g1", TenantID: "globex", Name: "Globex launch plan"},
	)
	mux := http.NewServeMux()
	(&projects.Handler{Store: store, Log: slog.New(slog.DiscardHandler)}).Register(mux)
	return mux
}

// as sends a GET as the given tenant, the way jwtauth.Middleware would after a valid token.
func as(t *testing.T, h http.Handler, tenant, path string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	if tenant != "" {
		req = req.WithContext(jwtauth.WithClaims(req.Context(), &jwtauth.Claims{Tenant: tenant}))
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestOwnerReadsOwnProject(t *testing.T) {
	rec := as(t, newServer(), "acme", "/v1/projects/prj_a1")
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "Acme roadmap") {
		t.Fatalf("status %d, body %q; want 200 with the project", rec.Code, rec.Body.String())
	}
}

// TestTenantCannotReadAnotherTenantsProject is the BOLA test: a valid,
// authenticated caller asks for an id that exists but is not hers.
func TestTenantCannotReadAnotherTenantsProject(t *testing.T) {
	h := newServer()
	rec := as(t, h, "globex", "/v1/projects/prj_a1")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("globex reading acme's project: status %d, body %q; want 404", rec.Code, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), "Acme") {
		t.Fatalf("response leaks the other tenant's data: %q", rec.Body.String())
	}
	// "Not yours" must look exactly like "does not exist", or ids can be enumerated.
	missing := as(t, h, "globex", "/v1/projects/prj_zz")
	if missing.Code != rec.Code || missing.Body.String() != rec.Body.String() {
		t.Fatalf("cross-tenant answer %d %q differs from missing-id answer %d %q",
			rec.Code, rec.Body.String(), missing.Code, missing.Body.String())
	}
}

func TestListIsScopedToTenant(t *testing.T) {
	rec := as(t, newServer(), "globex", "/v1/projects")
	var body struct {
		Items []projects.Project `json:"items"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("status %d, body %q: %v", rec.Code, rec.Body.String(), err)
	}
	if len(body.Items) != 1 || body.Items[0].ID != "prj_g1" {
		t.Fatalf("globex list = %+v; want only prj_g1", body.Items)
	}
}

func TestNoIdentityIsUnauthorized(t *testing.T) {
	if rec := as(t, newServer(), "", "/v1/projects/prj_a1"); rec.Code != http.StatusUnauthorized {
		t.Fatalf("status %d; want 401", rec.Code)
	}
}
