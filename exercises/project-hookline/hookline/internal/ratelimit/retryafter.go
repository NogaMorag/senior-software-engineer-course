package ratelimit

import (
	"math"
	"strconv"
	"time"
)

// RetryAfterSeconds formats a wait as Retry-After's delta-seconds: whole seconds, rounded up so a
// client never comes back too early, and at least 1.
func RetryAfterSeconds(d time.Duration) string {
	return strconv.Itoa(max(1, int(math.Ceil(d.Seconds()))))
}
