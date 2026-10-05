package delivery

import (
	"context"
	"encoding/json"

	"net/http"

	"time"

	"example.com/hookline/internal/ssrf"
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
	// TraceContext is the ingest request's traceparent (milestone M8), "" without telemetry.
	TraceContext string
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
	Client  *http.Client     // nil means NewClient(Policy)
	Policy  *ssrf.Policy     // the address guard for every connection (milestone M9); nil means none
	Timeout time.Duration    // per attempt; 0 means 10 seconds
	Now     func() time.Time // nil means time.Now; the signature timestamp comes from it
}

// NewClient returns the client a Sender should use: no redirects (a 3xx is the receiver's answer,
// not an instruction to post the signed body somewhere else), no global timeout (each attempt has its
// own, from the context), and otelhttp's transport, which makes a client span per attempt and sends
// traceparent, so a receiver that traces continues Hookline's trace (milestone M8). With a policy,
// every connection goes through its dialer check, after DNS resolution: an endpoint whose name now
// resolves to a private address is refused at send time, whatever it resolved to at creation (M9).
func NewClient(policy *ssrf.Policy) *http.Client {
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
