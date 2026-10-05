//! The Go tests of chapter 16 (`tests/ratelimit/bucket_test.go`), case for case, plus one
//! test that Go checks with `-race` and Rust checks by compiling: sharing between threads.

use std::sync::{Arc, Mutex};
use std::thread;
use std::time::{Duration, Instant};

use tokenbucket::{Bucket, Clock, Decision, Local};

/// A fake time source: a test of "half a second later" moves it, not the world. It is
/// `Clone`, and every clone shares the same time, so the test keeps one and the bucket
/// gets another.
#[derive(Clone)]
struct FakeClock(Arc<Mutex<Instant>>);

impl FakeClock {
    fn new() -> Self {
        // Start a day in the future, so that moving back an hour never underflows.
        FakeClock(Arc::new(Mutex::new(
            Instant::now() + Duration::from_secs(24 * 3600),
        )))
    }
    fn advance(&self, d: Duration) {
        *self.0.lock().unwrap() += d;
    }
    fn rewind(&self, d: Duration) {
        *self.0.lock().unwrap() -= d;
    }
}

impl Clock for FakeClock {
    fn now(&self) -> Instant {
        *self.0.lock().unwrap()
    }
}

const MS: Duration = Duration::from_millis(1);

struct Step {
    advance: Duration,
    n: u32,
    allowed: bool,
    remaining: u32,
    retry: Duration,
}

fn step(advance: Duration, n: u32, allowed: bool, remaining: u32, retry: Duration) -> Step {
    Step {
        advance,
        n,
        allowed,
        remaining,
        retry,
    }
}

#[test]
fn bucket_table() {
    let zero = Duration::ZERO;
    let cases: Vec<(&str, Vec<Step>)> = vec![
        (
            "burst then empty",
            vec![
                step(zero, 1, true, 4, zero),
                step(zero, 1, true, 3, zero),
                step(zero, 1, true, 2, zero),
                step(zero, 1, true, 1, zero),
                step(zero, 1, true, 0, zero),
                // 2 tokens per second: the next one is 0.5 s away
                step(zero, 1, false, 0, 500 * MS),
            ],
        ),
        (
            "refill",
            vec![
                step(zero, 5, true, 0, zero),
                step(250 * MS, 1, false, 0, 250 * MS), // half a token so far
                step(250 * MS, 1, true, 0, zero),
                step(Duration::from_secs(1), 1, true, 1, zero), // two more, one taken
            ],
        ),
        (
            "refill stops at burst",
            vec![step(Duration::from_secs(3600), 1, true, 4, zero)],
        ),
        (
            "more than burst is never allowed",
            vec![step(Duration::from_secs(3600), 6, false, 5, zero)],
        ),
        (
            "retry-after for several tokens",
            vec![
                step(zero, 5, true, 0, zero),
                step(zero, 3, false, 0, 1500 * MS),
            ],
        ),
    ];
    for (name, steps) in cases {
        let clock = FakeClock::new();
        let bucket = Bucket::new(2.0, 5, clock.clone()); // 2 tokens per second, room for 5
        for (i, s) in steps.iter().enumerate() {
            clock.advance(s.advance);
            let got = bucket.allow_n(s.n);
            let want = Decision {
                allowed: s.allowed,
                remaining: s.remaining,
                retry_after: s.retry,
            };
            assert_eq!(got, want, "{name}: step {i} (+{:?}, n={})", s.advance, s.n);
        }
    }
}

#[test]
fn clock_going_backwards_earns_nothing() {
    let clock = FakeClock::new();
    let bucket = Bucket::new(1.0, 1, clock.clone());
    bucket.allow();
    clock.rewind(Duration::from_secs(3600));
    let d = bucket.allow();
    assert!(
        !d.allowed,
        "after the clock went back an hour: {d:?}, want refused"
    );
}

#[test]
fn local_keys_are_independent() {
    let clock = FakeClock::new();
    let limiter = Local::new(1.0, 2, clock);
    for (i, want) in [true, true, false].into_iter().enumerate() {
        assert_eq!(
            limiter.allow("tenant-a").allowed,
            want,
            "tenant-a request {i}"
        );
    }
    assert!(
        limiter.allow("tenant-b").allowed,
        "tenant-b was refused because tenant-a used its own bucket up"
    );
}

#[test]
fn shared_between_threads_never_over_allows() {
    // Go proves this with `go test -race`; here the compiler refuses to build the test if
    // Bucket cannot be shared, and the count proves the lock is held across read and write.
    let clock = FakeClock::new(); // never advanced: no refill during the test
    let bucket = Arc::new(Bucket::new(1.0, 100, clock));
    let allowed: u32 = thread::scope(|s| {
        let handles: Vec<_> = (0..8)
            .map(|_| {
                let b = Arc::clone(&bucket);
                s.spawn(move || (0..50).filter(|_| b.allow().allowed).count() as u32)
            })
            .collect();
        handles.into_iter().map(|h| h.join().unwrap()).sum()
    });
    assert_eq!(allowed, 100, "400 requests against a burst of 100");
}
