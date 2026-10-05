package api

import (
	"net/http"
)

// authenticate resolves "Authorization: Bearer hk_..." to a tenant and puts it in the context. The key
// is hashed and looked up by its hash: the plaintext is never stored, logged or compared.
func (s *server) authenticate(next http.Handler) http.Handler {
	panic("TODO: see chapter 38")

}

func unauthorized(w http.ResponseWriter, r *http.Request, detail string) {
	panic("TODO: see chapter 38")

}

// getTenant handles GET /v1/tenant: who am I, and my limits.
func (s *server) getTenant(w http.ResponseWriter, r *http.Request) {
	panic("TODO: see chapter 38")

}
