package app

import (
	"net/http"
	"sync/atomic"
)

// Readiness answers GET /readyz. Added for chapters 19 and 20: an orchestrator
// sends traffic only to a copy that says it is ready. This service has no
// dependencies to check, so readiness is a switch that an operator can turn
// off by hand (the readiness drill of chapter 20) on the private debug listener.
type Readiness struct {
	off atomic.Bool // the zero value is ready
}

// ServeHTTP is GET /readyz: 200 while ready, 503 otherwise.
func (r *Readiness) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	if r.off.Load() {
		http.Error(w, "not ready", http.StatusServiceUnavailable)
		return
	}
	_, _ = w.Write([]byte("ready\n"))
}

// Set turns readiness on or off.
func (r *Readiness) Set(ready bool) { r.off.Store(!ready) }

// Toggle is POST /debug/ready?state=on|off on the debug listener.
func (r *Readiness) Toggle(w http.ResponseWriter, req *http.Request) {
	switch req.URL.Query().Get("state") {
	case "on":
		r.Set(true)
	case "off":
		r.Set(false)
	default:
		http.Error(w, "state must be on or off", http.StatusBadRequest)
		return
	}
	r.ServeHTTP(w, req)
}
