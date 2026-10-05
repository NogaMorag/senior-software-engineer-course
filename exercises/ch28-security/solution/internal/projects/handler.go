package projects

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"example.com/ch28-security/internal/jwtauth"
)

// Store is what the handlers need. Every method takes the tenant: there is
// no way to ask for an object without saying whose it must be.
type Store interface {
	Get(ctx context.Context, tenantID, id string) (Project, error)
	List(ctx context.Context, tenantID string) ([]Project, error)
}

// Handler serves the projects API. It must run behind jwtauth.Middleware.
type Handler struct {
	Store Store
	Log   *slog.Logger
}

// Register adds the routes to mux.
func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /v1/projects", h.list)
	mux.HandleFunc("GET /v1/projects/{id}", h.get)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	claims := jwtauth.FromContext(r.Context())
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	// The tenant comes from the verified token, never from the request.
	p, err := h.Store.Get(r.Context(), claims.Tenant, r.PathValue("id"))
	switch {
	case errors.Is(err, ErrNotFound):
		http.Error(w, "not found", http.StatusNotFound) // same answer for "missing" and "not yours"
		return
	case err != nil:
		h.Log.ErrorContext(r.Context(), "get project", slog.Any("err", err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, p)
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	claims := jwtauth.FromContext(r.Context())
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	ps, err := h.Store.List(r.Context(), claims.Tenant)
	if err != nil {
		h.Log.ErrorContext(r.Context(), "list projects", slog.Any("err", err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]any{"items": ps})
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}
