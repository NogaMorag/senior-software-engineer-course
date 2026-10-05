package platform

import (
	"net/http"
	"net/http/pprof"
)

// DebugHandler serves the pprof endpoints under /debug/pprof/. Serve it on a
// separate listener bound to localhost, never on the public port: profiles
// expose internals and a CPU profile costs CPU while it runs.
//
// Importing net/http/pprof also registers these paths on
// http.DefaultServeMux. That is harmless here because no server in this
// service uses DefaultServeMux.
func DebugHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/debug/pprof/", pprof.Index) // index, heap, goroutine, allocs, ...
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile) // CPU, ?seconds=N
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	return mux
}
