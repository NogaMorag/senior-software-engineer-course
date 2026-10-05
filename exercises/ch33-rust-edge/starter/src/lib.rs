//! The chapter 16 token bucket (`exercises/ch16-ratelimit/solution/ratelimit/bucket.go` and
//! `local.go`), ported to Rust: one bucket with an injected clock, and a bucket per key.
//! The Redis-backed limiter is not ported; it is the same algorithm in a Lua script.

// The starter leaves fields unread until the two TODOs are done; delete this line then.
#![allow(dead_code)]

use std::collections::HashMap;
use std::sync::{Arc, Mutex};
use std::time::{Duration, Instant};

/// The answer for one request. Go: `ratelimit.Decision`.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub struct Decision {
    pub allowed: bool,
    /// Whole tokens left after this decision.
    pub remaining: u32,
    /// When refused: how long until enough tokens have accrued. Zero when allowed, or
    /// when the request can never be allowed.
    pub retry_after: Duration,
}

/// A source of time. Go passes `now func() time.Time`; Rust uses a trait so that the clock
/// is part of the type and a test can pass a fake one.
pub trait Clock {
    fn now(&self) -> Instant;
}

/// The real clock. `Instant` is monotonic: it never goes backwards.
#[derive(Debug, Clone, Copy, Default)]
pub struct SystemClock;

impl Clock for SystemClock {
    fn now(&self) -> Instant {
        Instant::now()
    }
}

#[derive(Debug)]
struct State {
    tokens: f64,
    last: Instant,
}

/// Holds up to `burst` tokens and gains `rate` tokens per second. A request takes tokens;
/// with too few, it is refused. `Bucket` is `Sync` when its clock is, so it can be shared
/// between threads behind an `Arc`, like the Go version behind a pointer.
#[derive(Debug)]
pub struct Bucket<C: Clock> {
    rate: f64,
    burst: f64,
    clock: C,
    state: Mutex<State>,
}

impl<C: Clock> Bucket<C> {
    /// Returns a full bucket. Go: `NewBucket(rate, burst, now)`.
    pub fn new(rate: f64, burst: u32, clock: C) -> Self {
        let last = clock.now();
        Bucket {
            rate,
            burst: f64::from(burst),
            clock,
            state: Mutex::new(State {
                tokens: f64::from(burst),
                last,
            }),
        }
    }

    /// Asks for one token.
    pub fn allow(&self) -> Decision {
        self.allow_n(1)
    }

    /// Asks for `n` tokens. A request for more than `burst` is never allowed.
    pub fn allow_n(&self, n: u32) -> Decision {
        // TODO(day 92): port AllowN from bucket.go.
        // 1. Lock the state. 2. Read the clock. 3. Refill by the time elapsed since `last`,
        //    capped at `burst`; a clock that went backwards earns nothing
        //    (`Instant::checked_duration_since` returns `None`).
        // 4. More than burst: refused, retry_after zero. 5. Enough tokens: take them.
        // 6. Otherwise: refused, retry_after = the time until enough tokens, rounded up to
        //    a whole nanosecond.
        todo!(
            "day 92: Bucket::allow_n for {n} token(s), {} tokens/s",
            self.rate
        )
    }
}

/// One bucket per key in this process. Go: `ratelimit.Local`. As in Go, buckets are never
/// removed; a real service evicts idle keys.
#[derive(Debug)]
pub struct Local<C: Clock + Clone> {
    rate: f64,
    burst: u32,
    clock: C,
    buckets: Mutex<HashMap<String, Arc<Bucket<C>>>>,
}

impl<C: Clock + Clone> Local<C> {
    pub fn new(rate: f64, burst: u32, clock: C) -> Self {
        Local {
            rate,
            burst,
            clock,
            buckets: Mutex::new(HashMap::new()),
        }
    }

    /// Takes one token from `key`'s bucket. The map lock is held only to find or create the
    /// bucket; the bucket's own lock is taken after the map lock is released.
    pub fn allow(&self, key: &str) -> Decision {
        // TODO(day 92): port Local.Allow from local.go. Find or create the key's bucket
        // while holding the map lock, clone the Arc, release the map lock, then call
        // allow() on the bucket. Hint: HashMap::entry and or_insert_with.
        todo!("day 92: Local::allow for key {key:?}, burst {}", self.burst)
    }
}
