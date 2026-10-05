# Hookline: starter

The full project structure, tests, migrations, contracts, charts, workflows and scripts, with
every Go function body replaced by `panic("TODO: see chapter 38")` and every Python function body by
`raise NotImplementedError`. Chapter 38 walks through the ten milestones; each names the packages to implement
and the tests that must pass. The reference implementation is next to this directory (`../hookline`).

```bash
go build ./...          # compiles from day one
go test ./...           # red at first; green package by package, M1 to M10
```
