# Rust vs Go: one page (chapter 33)

Date: ____  Based on: the token bucket in Go (chapter 16) and my Rust port (this exercise, or "not ported").

Fill every blank from something you did or measured, not from something you read. If a cell is a guess, write
"guess" in it.

## 1. The same component, twice
| Aspect | Go (`ratelimit/bucket.go`) | Rust (`src/lib.rs`) | Which I prefer here, and why |
| --- | --- | --- | --- |
| Lines of code (without tests) | | | |
| Injecting the clock | | | |
| Making it safe to share between threads | | | |
| What the compiler caught that Go found only at run time (or not at all) | | | |
| What was faster to write | | | |
| How a teammate new to the language would read it | | | |

## 2. When I would choose Rust (at least one real case)
| Case | Why Rust | What it costs the team |
| --- | --- | --- |
| | | |

## 3. When I would choose Go (at least one real case)
| Case | Why Go | What it costs |
| --- | --- | --- |
| | | |

## 4. The sentence I will say in an interview
"For ____ I would pick ____, because ____; I would not pick it for ____, because ____."

## 5. Two patterns I translate in my head
| In Rust I would write | In Go I write |
| --- | --- |
| | |
| | |
