package delivery

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

// Job is a claimed delivery with what sending it needs: the endpoint and the event.
type Job struct {
	DeliveryID  string
	TenantID    string
	EventID     string
	EndpointID  string
	Attempt     int // the attempts value after this claim: 1, 2, 3, ... (Hookline-Attempt)
	Round       int // attempts in this round, after a replay (compared with the maximum)
	URL         string
	Secret      string
	Enabled     bool
	EventType   string
	EventAt     time.Time
	Payload     json.RawMessage
	LeaseExpiry time.Time
}

// Body is the JSON every receiver gets. It is built from the stored event, so every attempt and every
// endpoint gets the same bytes; only the signature's timestamp changes between attempts.
func (j Job) Body() []byte {
	b, _ := json.Marshal(struct {
		ID        string          `json:"id"`
		Type      string          `json:"type"`
		CreatedAt time.Time       `json:"created_at"`
		Payload   json.RawMessage `json:"payload"`
	}{j.EventID, j.EventType, j.EventAt, j.Payload})
	return b
}

// Result is what one attempt produced.
type Result struct {
	StatusCode int           // 0 when no response arrived
	Err        error         // transport error, timeout, or the classification of a bad status
	Retryable  bool          // false for success and for failures a retry cannot fix
	RetryAfter time.Duration // from a Retry-After header in seconds, when present
	Started    time.Time
	Duration   time.Duration
}

// OK reports whether the receiver accepted the delivery.
func (r Result) OK() bool { return r.Err == nil }

// Sender makes one HTTP attempt.
type Sender struct {
	Client  *http.Client     // nil means NewClient()
	Timeout time.Duration    // per attempt; 0 means 10 seconds
	Now     func() time.Time // nil means time.Now; the signature timestamp comes from it
}

// NewClient returns the client a Sender should use: no redirects (a 3xx is the receiver's answer,
// not an instruction to post the signed body somewhere else) and no global timeout (each attempt has
// its own, from the context).
func NewClient() *http.Client {
	return &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
}

// Send posts j's body to its endpoint, signed, with the four Hookline headers, and classifies the outcome.
func (s Sender) Send(ctx context.Context, j Job) Result {
	now, client, timeout := s.Now, s.Client, s.Timeout
	if now == nil {
		now = time.Now
	}
	if client == nil {
		client = NewClient()
	}
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	body := j.Body()
	started := now()
	res := Result{Started: started}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, j.URL, bytes.NewReader(body))
	if err != nil {
		res.Err = fmt.Errorf("build request: %w", err) // the URL was validated at creation: permanent
		return res
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Hookline/1")
	req.Header.Set(HeaderEventID, j.EventID)
	req.Header.Set(HeaderDeliveryID, j.DeliveryID)
	req.Header.Set(HeaderAttempt, strconv.Itoa(j.Attempt))
	req.Header.Set(HeaderSignature, Sign(j.Secret, started.Unix(), body))

	resp, err := client.Do(req)
	res.Duration = time.Since(started)
	if err != nil {
		res.Err, res.Retryable = err, true // timeout, refused, reset, DNS: the receiver may recover
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			res.Err = fmt.Errorf("no response within %s: %w", timeout, err)
		}
		return res
	}
	// Read a little of the body so the connection can be reused, then drop it: receivers' bodies are
	// not stored.
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 64<<10))
	resp.Body.Close()
	res.StatusCode = resp.StatusCode
	res.Retryable, res.Err = Classify(resp.StatusCode)
	if secs, err := strconv.Atoi(resp.Header.Get("Retry-After")); err == nil && secs > 0 {
		res.RetryAfter = time.Duration(secs) * time.Second
	}
	return res
}

// Classify maps a status code to the retry policy of spec F6: 2xx succeeds; 408, 429 and 5xx are worth
// retrying; anything else (other 4xx, 3xx) is the receiver's final answer.
func Classify(status int) (retryable bool, err error) {
	switch {
	case status >= 200 && status < 300:
		return false, nil
	case status == http.StatusRequestTimeout || status == http.StatusTooManyRequests || status >= 500:
		return true, fmt.Errorf("receiver answered %d", status)
	default:
		return false, fmt.Errorf("receiver answered %d; not retried", status)
	}
}
