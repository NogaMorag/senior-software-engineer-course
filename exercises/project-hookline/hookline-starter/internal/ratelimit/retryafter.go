package ratelimit

import (
	"time"
)

// RetryAfterSeconds formats a wait as Retry-After's delta-seconds: whole seconds, rounded up so a
// client never comes back too early, and at least 1.
func RetryAfterSeconds(d time.Duration) string {
	panic("TODO: see chapter 38")

}
