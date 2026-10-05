// Package idem implements Idempotency-Key handling for unsafe requests (POST), after the IETF httpapi
// draft "The Idempotency-Key HTTP Header Field": the same key with the same request replays the stored
// response; the same key with a different request is 422; a retry while the first is still running is 409.
package idem

import (
	"bytes"
	"crypto/sha256"
	"maps"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Header is the request header that carries the key. ReplayedHeader marks a replayed response; it is
// this API's own convention, not part of the draft.
const (
	Header         = "Idempotency-Key"
	ReplayedHeader = "Idempotent-Replayed"
	maxKeyLen      = 255
)

// Outcome is what Begin decided for a request.
type Outcome int

// The four possible outcomes of Begin.
const (
	Started    Outcome = iota // first time: run the handler, then Complete or Release
	Replay                    // finished before with the same request: send the stored response
	Mismatch                  // finished or running with a different request: 422
	InProgress                // same request still running: 409
)

// Response is a stored response: enough to replay it byte for byte.
type Response struct {
	Status int
	Header http.Header
	Body   []byte
}

type entry struct {
	fingerprint [32]byte
	done        bool
	resp        Response
	expires     time.Time
}

// Store remembers keys for TTL. It is in memory, so it is per process; several replicas need a shared
// store (PostgreSQL in the capstone).
type Store struct {
	mu        sync.Mutex
	entries   map[string]*entry
	ttl       time.Duration
	now       func() time.Time
	lastSweep time.Time
}

// NewStore returns a store whose keys live for ttl. now is injected for tests (nil means time.Now).
func NewStore(ttl time.Duration, now func() time.Time) *Store {
	if now == nil {
		now = time.Now
	}
	return &Store{entries: map[string]*entry{}, ttl: ttl, now: now, lastSweep: now()}
}

// Begin claims key for a request with fingerprint fp, atomically.
func (s *Store) Begin(key string, fp [32]byte) (Outcome, Response) {
	// TODO(day 20): under the lock: sweep; treat an expired entry as absent; then
	//   absent                      -> store a new in-progress entry, return Started
	//   different fingerprint       -> Mismatch
	//   same fingerprint, not done  -> InProgress
	//   same fingerprint, done      -> Replay with the stored response
	panic("TODO(day 20): Store.Begin")
}

// Complete stores the response for a key claimed by Begin.
func (s *Store) Complete(key string, resp Response) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if e, ok := s.entries[key]; ok {
		e.done, e.resp = true, resp
	}
}

// Release forgets a claimed key so that a retry runs the handler again (used after a 5xx or a panic).
func (s *Store) Release(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.entries, key)
}

// Len reports how many keys are stored (for tests and metrics).
func (s *Store) Len() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.entries)
}

// sweep drops expired entries at most once per TTL, so memory stays bounded by the key rate times TTL.
func (s *Store) sweep(now time.Time) {
	if now.Sub(s.lastSweep) < s.ttl {
		return
	}
	for k, e := range s.entries {
		if now.After(e.expires) {
			delete(s.entries, k)
		}
	}
	s.lastSweep = now
}

// Middleware enforces idempotency keys on every request it wraps. Put it inside the body-size limit
// and around handlers of unsafe methods only.
func (s *Store) Middleware(next http.Handler) http.Handler {
	// TODO(day 20): build the middleware described in section 9.5 and the README, in this order:
	//  1. parseKey; a missing or invalid key is an IdempotencyMissing problem.
	//  2. read the whole body (a *http.MaxBytesError is a TooLarge problem) and put it back on r.Body.
	//  3. scope the key by method and path, then Begin with fingerprint(r, body).
	//  4. Replay: writeStored. Mismatch: KeyReused problem. InProgress: Retry-After and InProgress problem.
	//  5. Started: run next into a recorder; on 5xx or panic Release the key, otherwise Complete it;
	//     then copy the recorded response to w.
	return next
}

// parseKey accepts a bare token or the quoted Structured Fields string form the draft specifies.
func parseKey(v string) (string, bool) {
	if len(v) >= 2 && v[0] == '"' && v[len(v)-1] == '"' {
		v = v[1 : len(v)-1]
	}
	if v == "" || len(v) > maxKeyLen {
		return "", false
	}
	for i := 0; i < len(v); i++ {
		if v[i] < 0x21 || v[i] > 0x7e || v[i] == '"' {
			return "", false
		}
	}
	return v, true
}

// fingerprint identifies "the same request": method, path, query, content type and exact body bytes.
func fingerprint(r *http.Request, body []byte) [32]byte {
	h := sha256.New()
	for _, part := range []string{r.Method, r.URL.Path, r.URL.RawQuery, strings.ToLower(r.Header.Get("Content-Type"))} {
		h.Write([]byte(part))
		h.Write([]byte{0})
	}
	h.Write(body)
	var out [32]byte
	copy(out[:], h.Sum(nil))
	return out
}

func writeStored(w http.ResponseWriter, resp Response) {
	copyHeader(w.Header(), resp.Header)
	w.Header().Set(ReplayedHeader, "true")
	w.WriteHeader(resp.Status)
	_, _ = w.Write(resp.Body)
}

func copyHeader(dst, src http.Header) {
	maps.Copy(dst, src.Clone())
}

// recorder buffers a handler's response so it can be stored before it is sent.
type recorder struct {
	header      http.Header
	status      int
	wroteHeader bool
	body        bytes.Buffer
}

func (r *recorder) Header() http.Header { return r.header }

func (r *recorder) WriteHeader(status int) {
	if !r.wroteHeader {
		r.status, r.wroteHeader = status, true
	}
}

func (r *recorder) Write(b []byte) (int, error) {
	r.wroteHeader = true
	return r.body.Write(b)
}
