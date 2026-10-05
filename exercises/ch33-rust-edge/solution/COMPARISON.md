# Rust vs Go: one page (chapter 33), worked example

This is the author's filled-in page, written while porting the bucket for the book. Yours will differ: the point
is that every cell comes from the two files, not from opinions about the languages.

## 1. The same component, twice
| Aspect | Go (`ratelimit/bucket.go`, `local.go`) | Rust (`src/lib.rs`) | Which I prefer here, and why |
| --- | --- | --- | --- |
| Lines of code (without tests), by `wc -l` | 103 (60 and 43) | 144 | Rust is longer: struct literals, explicit conversions (`f64::from`), and the `Clock` trait |
| Injecting the clock | a field `now func() time.Time` | a generic parameter `C: Clock` | Go: one line, no generics. Rust: the fake clock is part of the type, and calls are static |
| Making it safe to share between threads | `sync.Mutex` next to the fields; correctness checked by `go test -race` at run time | `Mutex<State>` wraps the fields; `Bucket` is `Sync` only if `C` is, checked at compile time | Rust: forgetting the lock is impossible, because the fields are only reachable through it |
| What the compiler caught that Go found only at run time (or not at all) | a clock that goes back: handled by an explicit `elapsed > 0` check | `checked_duration_since` returns `Option`, so "backwards" must be handled to compile | Rust, for this kind of edge |
| What was faster to write | Go | Rust took longer: the borrow on the map entry in `Local::allow` needed a block | Go, for a component this small |
| How a teammate new to the language would read it | straightforward | the `Arc<Bucket<C>>` inside `Mutex<HashMap<..>>` needs explaining | Go |

## 2. When I would choose Rust
| Case | Why Rust | What it costs the team |
| --- | --- | --- |
| A packet-path or proxy component with a tail-latency target where GC pauses show up in p99 | no garbage collector; memory layout under control | slower onboarding, fewer people who can review it, a second toolchain in CI |
| Code that parses untrusted input at a security boundary, or wraps a C library through FFI | memory safety without a runtime; `unsafe` is confined and greppable | the same, plus FFI review discipline |

## 3. When I would choose Go
| Case | Why Go | What it costs |
| --- | --- | --- |
| A Kubernetes operator, a CLI for the platform team, or an HTTP/gRPC service like Hookline's API | the cloud-native ecosystem (client-go, controller-runtime, OpenTelemetry) is Go-first; fast to write and to review | GC and less help from the type system with shared state; `-race` and tests have to cover it |

## 4. The sentence I will say in an interview
"For a Kubernetes controller I would pick Go, because the ecosystem and the team are there; I would not pick it
for a component on a hot path with a strict p99 target, where I would measure GC pauses first and consider Rust."

## 5. Two patterns I translate in my head
| In Rust I would write | In Go I write |
| --- | --- |
| `Result<T, E>` and `?` | `if err != nil { return fmt.Errorf("...: %w", err) }` |
| `Mutex<State>` owning the data | `mu sync.Mutex` beside the fields, by convention, plus `-race` in CI |
