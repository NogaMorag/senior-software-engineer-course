package cache

import (
	"time"
)

// Jittered returns base plus a random duration in [0, jitter). Entries written
// together then expire spread out instead of all in the same second.
func Jittered(base, jitter time.Duration) time.Duration {
	// TODO(day 31): return base when jitter <= 0, otherwise base + rand.N(jitter).
	panic("TODO(day 31): Jittered")
}
