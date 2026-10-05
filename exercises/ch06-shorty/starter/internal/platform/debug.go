package platform

import (
	"net/http"
)

// DebugHandler serves the pprof endpoints under /debug/pprof/. Serve it on a
// separate listener bound to localhost, never on the public port: profiles
// expose internals and a CPU profile costs CPU while it runs.
//
// Importing net/http/pprof also registers these paths on
// http.DefaultServeMux. That is harmless here because no server in this
// service uses DefaultServeMux.
func DebugHandler() http.Handler {
	// TODO 9 (day 13, section 6.8): a new ServeMux with the five pprof handlers:
	// /debug/pprof/ (pprof.Index), cmdline, profile, symbol and trace.
	return http.NotFoundHandler() // until then: every debug path is a 404
}
