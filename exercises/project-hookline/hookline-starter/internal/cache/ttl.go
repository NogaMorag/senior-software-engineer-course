package cache

import (
	"time"
)

// Jittered returns base plus a random duration in [0, jitter). Entries written
// together then expire spread out instead of all in the same second.
func Jittered(base, jitter time.Duration) time.Duration {
	panic("TODO: see chapter 38")

}
