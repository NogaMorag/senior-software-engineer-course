package delivery

import (
	"context"
	"encoding/json"

	"net/http"

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
	panic("TODO: see chapter 38")

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
func (r Result) OK() bool { panic("TODO: see chapter 38") }

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
	panic("TODO: see chapter 38")

}

// Send posts j's body to its endpoint, signed, with the four Hookline headers, and classifies the outcome.
func (s Sender) Send(ctx context.Context, j Job) Result {
	panic("TODO: see chapter 38")

}

// Classify maps a status code to the retry policy of spec F6: 2xx succeeds; 408, 429 and 5xx are worth
// retrying; anything else (other 4xx, 3xx) is the receiver's final answer.
func Classify(status int) (retryable bool, err error) {
	panic("TODO: see chapter 38")

}
