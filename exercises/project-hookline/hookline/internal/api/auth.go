package api

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"example.com/hookline/internal/domain"
	"example.com/hookline/internal/problem"
)

// authenticate resolves "Authorization: Bearer hk_..." to a tenant and puts it in the context. The key
// is hashed and looked up by its hash: the plaintext is never stored, logged or compared.
func (s *server) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		scheme, key, ok := strings.Cut(r.Header.Get("Authorization"), " ")
		if !ok || !strings.EqualFold(scheme, "Bearer") || !domain.LooksLikeAPIKey(key) {
			unauthorized(w, r, "send the header Authorization: Bearer followed by your API key")
			return
		}
		t, err := s.store.TenantByKeyHash(r.Context(), domain.HashAPIKey(key))
		if errors.Is(err, domain.ErrNotFound) {
			unauthorized(w, r, "the API key is not valid")
			return
		}
		if err != nil {
			s.internal(w, r, err)
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), tenantKey{}, t)))
	})
}

func unauthorized(w http.ResponseWriter, r *http.Request, detail string) {
	w.Header().Set("WWW-Authenticate", `Bearer realm="hookline"`)
	problem.Write(w, r, problem.Unauthorized.New(detail))
}

// getTenant handles GET /v1/tenant: who am I, and my limits.
func (s *server) getTenant(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, tenantFrom(r.Context()))
}
