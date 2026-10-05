package replica

import (
	"sync"
	"time"
)

// Router implements session stickiness, the simplest read-your-writes strategy: for Window
// after a user's last write, that user's reads go to the primary; everyone else, and that
// user afterwards, reads from the replica. Window must exceed the replica lag you expect,
// which is why the lag you measure in this chapter matters.
type Router struct {
	Window time.Duration
	Now    func() time.Time // injected so tests control time

	mu        sync.Mutex
	lastWrite map[string]time.Time
}

// NewRouter returns a Router that uses the real clock.
func NewRouter(window time.Duration) *Router {
	return &Router{Window: window, Now: time.Now}
}

// RecordWrite notes that user has just written.
func (r *Router) RecordWrite(user string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.lastWrite == nil {
		r.lastWrite = make(map[string]time.Time)
	}
	r.lastWrite[user] = r.Now()
}

// ForRead says which server should answer user's next read.
func (r *Router) ForRead(user string) Source {
	r.mu.Lock()
	defer r.mu.Unlock()
	last, ok := r.lastWrite[user]
	if !ok {
		return FromReplica
	}
	if r.Now().Sub(last) < r.Window {
		return FromPrimary
	}
	delete(r.lastWrite, user) // the window has passed; forget the user
	return FromReplica
}
