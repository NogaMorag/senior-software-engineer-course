package ttlcache

import (
	"encoding/json"
	"io"
	"net/http"
)

// NewHandler serves the cache over HTTP:
//
//	GET    /cache/{key}  200 with the value, or 404
//	PUT    /cache/{key}  204; the body is the value, at most maxBody bytes
//	DELETE /cache/{key}  204, or 404
//	GET    /stats        200 with Stats as JSON
func NewHandler(c *Cache, maxBody int64) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /cache/{key}", func(w http.ResponseWriter, r *http.Request) {
		v, ok := c.Get(r.PathValue("key"))
		if !ok {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/octet-stream")
		_, _ = w.Write(v)
	})
	mux.HandleFunc("PUT /cache/{key}", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(http.MaxBytesReader(w, r.Body, maxBody))
		c.Set(r.PathValue("key"), body)
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("DELETE /cache/{key}", func(w http.ResponseWriter, r *http.Request) {
		if !c.Delete(r.PathValue("key")) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("GET /stats", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(c.Stats())
	})
	return mux
}
