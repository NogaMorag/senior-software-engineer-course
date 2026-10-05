//! The chapter 16 token bucket (`exercises/ch16-ratelimit/solution/ratelimit/bucket.go` and
//! `local.go`), ported to Rust: one bucket with an injected clock, and a bucket per key.
//! The Redis-backed limiter is not ported; it is the same algorithm in a Lua script.

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
        // A poisoned mutex means another thread panicked while holding it; the state is
        // two numbers that are always consistent, so keep going with it.
        let mut s = self.state.lock().unwrap_or_else(|e| e.into_inner());
        let now = self.clock.now();
        // Refill lazily: no timer, just the tokens earned since the last call. A clock that
        // went backwards earns nothing (checked_duration_since returns None).
        if let Some(elapsed) = now.checked_duration_since(s.last)
            && !elapsed.is_zero()
        {
            s.tokens = self.burst.min(s.tokens + elapsed.as_secs_f64() * self.rate);
            s.last = now;
        }
        let remaining = s.tokens as u32; // truncates toward zero, like Go's int()
        let need = f64::from(n);
        if need > self.burst {
            return Decision {
                allowed: false,
                remaining,
                retry_after: Duration::ZERO,
            };
        }
        if s.tokens >= need {
            s.tokens -= need;
            return Decision {
                allowed: true,
                remaining: s.tokens as u32,
                retry_after: Duration::ZERO,
            };
        }
        let wait_nanos = ((need - s.tokens) / self.rate * 1e9).ceil() as u64;
        Decision {
            allowed: false,
            remaining,
            retry_after: Duration::from_nanos(wait_nanos),
        }
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
        let bucket = {
            let mut map = self.buckets.lock().unwrap_or_else(|e| e.into_inner());
            Arc::clone(map.entry(key.to_owned()).or_insert_with(|| {
                Arc::new(Bucket::new(self.rate, self.burst, self.clock.clone()))
            }))
        };
        bucket.allow()
    }
}
