// Package webhook is drill 4: an HTTP handler that receives webhooks from a sender that
// delivers "at least once", and makes processing safe to retry. Deliveries with the same
// event id within a window are processed once, including deliveries that arrive at the
// same moment.
package webhook

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// EventIDHeader carries the sender's id for the event: the same on every retry.
const EventIDHeader = "Hookline-Event-Id"

// MaxBody is the largest request body the handler reads.
const MaxBody = 1 << 20

// Event is what the processor receives.
type Event struct {
	ID      string
	Payload json.RawMessage
}

// Processor does the real work for one event. An error means "not done": the id is
// forgotten, the sender gets a 500, and its retry will be processed again.
type Processor func(ctx context.Context, ev Event) error

// Handler dedupes deliveries by event id. It is an http.Handler and safe for concurrent use.
type Handler struct {
	process Processor
	window  time.Duration
	now     func() time.Time

	mu       sync.Mutex
	inflight map[string]*call     // id -> the delivery that is processing it now
	done     map[string]time.Time // id -> when it was processed
	order    []string             // ids in done, oldest first, for pruning
}

// call is one processing of one id. Duplicates that arrive meanwhile wait on finished.
type call struct {
	finished chan struct{}
	err      error // written before finished is closed, read after
}

// New returns a handler that remembers a processed id for window. now is the clock.
func New(p Processor, window time.Duration, now func() time.Time) *Handler {
	return &Handler{process: p, window: window, now: now, inflight: make(map[string]*call), done: make(map[string]time.Time)}
}

type reply struct {
	Status string `json:"status"`
}

// ServeHTTP answers 200 {"status":"processed"} for the delivery that did the work,
// 200 {"status":"duplicate"} for every other delivery of a processed id, and 500 when
// processing failed (for the delivery that tried and for duplicates that waited on it).
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	id := r.Header.Get(EventIDHeader)
	if id == "" {
		http.Error(w, "missing "+EventIDHeader, http.StatusBadRequest)
		return
	}
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, MaxBody))
	if err != nil {
		var tooBig *http.MaxBytesError
		if errors.As(err, &tooBig) {
			http.Error(w, "body too large", http.StatusRequestEntityTooLarge)
			return
		}
		http.Error(w, "cannot read body", http.StatusBadRequest)
		return
	}
	if !json.Valid(body) {
		http.Error(w, "body is not JSON", http.StatusBadRequest)
		return
	}

	c, leader, dup := h.claim(id)
	switch {
	case dup:
		writeJSON(w, http.StatusOK, "duplicate")
		return
	case leader:
		// Detached from the request: if the sender hangs up mid-way, the work still
		// finishes and is recorded, and the waiting duplicates get its answer.
		c.err = h.run(context.WithoutCancel(r.Context()), Event{ID: id, Payload: body})
		h.complete(id, c)
		if c.err != nil {
			http.Error(w, "processing failed", http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, "processed")
		return
	}
	// Another delivery of the same id is processing it right now: wait for its outcome.
	select {
	case <-c.finished:
	case <-r.Context().Done():
		http.Error(w, "gave up waiting", http.StatusServiceUnavailable)
		return
	}
	if c.err != nil {
		http.Error(w, "processing failed", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, "duplicate")
}

// run calls the processor and turns a panic into an error. Without this, a panic would
// leave the id in inflight forever, and every later delivery of it would wait in vain.
func (h *Handler) run(ctx context.Context, ev Event) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("webhook: processor panicked: %v", r)
		}
	}()
	return h.process(ctx, ev)
}

// claim decides, atomically, what this delivery does: it is a duplicate of a processed id,
// or the leader that processes the id, or a follower that waits for the leader.
func (h *Handler) claim(id string) (c *call, leader, dup bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.prune()
	if _, ok := h.done[id]; ok {
		return nil, false, true
	}
	if c, ok := h.inflight[id]; ok {
		return c, false, false
	}
	c = &call{finished: make(chan struct{})}
	h.inflight[id] = c
	return c, true, false
}

// complete records the outcome and releases the followers. A failure is not remembered,
// so the sender's next retry is processed again.
func (h *Handler) complete(id string, c *call) {
	h.mu.Lock()
	delete(h.inflight, id)
	if c.err == nil {
		h.done[id] = h.now()
		h.order = append(h.order, id)
	}
	h.mu.Unlock()
	close(c.finished)
}

// prune forgets ids processed more than window ago. Ids enter order in completion time,
// so the expired ones are a prefix: the cost is proportional to what is removed, not to
// how much is remembered. Called with h.mu held.
func (h *Handler) prune() {
	cutoff := h.now().Add(-h.window)
	i := 0
	for ; i < len(h.order); i++ {
		if h.done[h.order[i]].After(cutoff) {
			break
		}
		delete(h.done, h.order[i])
	}
	h.order = h.order[i:]
}

// Remembered reports how many processed ids are inside the window right now.
func (h *Handler) Remembered() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.prune()
	return len(h.done)
}

func writeJSON(w http.ResponseWriter, status int, s string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(reply{Status: s})
}
