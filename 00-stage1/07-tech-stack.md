# Tech stack, pinned versions and installation

Written 2026-10-05. Every version below was verified on 2026-10-05 against a primary source; the evidence (URL per
row, arch notes, disagreements between sources) is in `versions-verified-2026-10-05.md`. These pins are copied into
`go.mod`, `pyproject.toml`, compose files, workflows and appendix C; prose in the book never hard-codes them (it says
"נכון לאוקטובר 2026" and points to appendix C).

## Author environment (where the code was executed)
The brief names a Linux machine; this rebuild ran on the reader's Mac instead (see `03-decisions-2026-10-04.md` D4):
macOS on Apple silicon (arm64), 12 cores, 24 GB RAM, Docker Desktop with Engine 29.1.3 and an 8 GB VM, zsh/bash.
User-space installs only: Go into `~/.local/go`, single binaries into `~/.local/bin`, Go tools with
`GOBIN=~/.local/bin`. Scratch copies for running suites live under `/tmp/senior-course/`. **The WSL2 instructions in
chapter 2 were written from Microsoft's and Docker's documentation and could not be executed here**; `09-review.md`
records this.

## Core toolchain
| Tool | Pin | Role | WSL2 (Ubuntu) install | Mac install | Cost |
| --- | --- | --- | --- | --- | --- |
| WSL | 3.0.1, default distro Ubuntu 26.04 LTS | Linux on Windows | PowerShell (admin): `wsl --install` | n/a | free |
| Docker Desktop | 4.93.0 (Engine 29.8.1) | containers; WSL2 backend | Windows installer, enable WSL integration | `.dmg` | free for personal use |
| Git | distro package | VCS | `sudo apt install git` | Xcode CLT / Homebrew | free |
| VS Code + WSL extension | current | editor | Windows installer | `.zip` / Homebrew | free |
| Go | 1.27.1 | primary language | official tarball into `/usr/local/go` (or `~/.local/go`) | official `.pkg` or tarball | free |
| uv | 0.12.23 | Python projects | `curl -LsSf https://astral.sh/uv/install.sh \| sh` | same, or Homebrew | free |
| Python | 3.14 (via uv) | second language | `uv python install 3.14` | same | free |
| staticcheck | 2026.2.1 (module v0.8.1) | Go linter | `go install honnef.co/go/tools/cmd/staticcheck@2026.2.1` | same | free |
| golangci-lint | v2.14.0 | Go meta-linter | release binary | release binary / Homebrew | free |
| govulncheck | v1.8.0 | Go vulnerability scan | `go install golang.org/x/vuln/cmd/govulncheck@v1.8.0` | same | free |
| goose | v3.28.0 | migrations | `go install github.com/pressly/goose/v3/cmd/goose@v3.28.0` | same | free |
| buf | v1.73.0 (+ protoc-gen-go v1.36.12, protoc-gen-go-grpc v1.6.2) | protobuf/gRPC | `go install` | same | free |

## Part E tools (installed on day 57, chapter 2 §2.10)
| Tool | Pin | WSL2 install | Mac install |
| --- | --- | --- | --- |
| kind | v0.33.0 (node image `kindest/node:v1.37.0` pinned by digest) | release binary `kind-linux-amd64` into `~/.local/bin` | `kind-darwin-arm64` |
| kubectl | v1.37.1 | `dl.k8s.io/release/v1.37.1/bin/linux/amd64/kubectl` | Homebrew or the darwin binary |
| Helm | v4.3.0 | release tarball | Homebrew or tarball |
| Terraform | 1.16.5 (providers kubernetes v3.3.0, helm v3.3.0) | zip from releases.hashicorp.com | `darwin_arm64` zip |
| k6 | v2.3.0 | release tarball | `macos-arm64` zip |
| actionlint | v1.7.12 | `go install` | same |
| Trivy / Grype / Syft | v0.75.0 / v0.120.0 / v1.54.0 | release binaries or `docker run` | same |

## Go modules (pinned in each `go.mod`)
| Module | Version |
| --- | --- |
| github.com/jackc/pgx/v5 | v5.11.0 |
| github.com/redis/go-redis/v9 | v9.22.0 |
| github.com/twmb/franz-go | v1.22.1 |
| github.com/pressly/goose/v3 | v3.28.0 |
| google.golang.org/grpc / google.golang.org/protobuf | v1.84.0 / v1.36.12 |
| go.opentelemetry.io/otel (+ sdk, exporters) | v1.47.0; semconv package `semconv/v1.43.0` |
| go.opentelemetry.io/contrib/.../otelhttp | v0.72.0 |
| github.com/prometheus/client_golang (mentioned, not imported by an exercise) | v1.24.1 |
| github.com/testcontainers/testcontainers-go | v0.44.0 |
| pgregory.net/rapid | v1.3.0 |
| github.com/sony/gobreaker/v2 | v2.4.0 (shown for comparison; Hookline implements its own breaker) |
| golang.org/x/time | v0.16.0 |
| golang.org/x/sync | from `go get` at write time (singleflight, errgroup) |
| github.com/golang-jwt/jwt/v5 | v5.3.1 |
| github.com/coreos/go-oidc/v3 (mentioned, not imported by an exercise) | v3.21.0 |

## Python packages (pinned in each `pyproject.toml`, locked in `uv.lock`)
fastapi 0.142.2 · uvicorn 0.54.0 · pydantic 2.13.5 · httpx 0.28.1 · pytest 9.1.1 · ruff 0.16.10 · mypy 2.4.0 ·
opentelemetry-sdk 1.45.0. Verified but not used by any exercise (kept for reference): hypothesis 6.168.4,
psycopg 3.3.6, redis 8.1.0.

## Container images (pinned in compose files and charts)
| Image | Tag |
| --- | --- |
| postgres | 18.6 |
| redis | 8.10.2 |
| docker.redpanda.com/redpandadata/redpanda | v26.2.3 |
| docker.redpanda.com/redpandadata/console | v3.12.0 |
| prom/prometheus | v3.15.0 |
| grafana/grafana | 13.2.3 |
| grafana/loki | 3.7.8 |
| grafana/tempo | 3.1.0 |
| otel/opentelemetry-collector-contrib | 0.162.0 (multi-arch on GHCR; Docker Hub `latest` lagged at 0.161.0 on 2026-10-05) |
| gcr.io/distroless/static-debian13 | `nonroot` |
| golang | 1.27.1 |
| python | 3.14.8-slim |
| redocly/cli | 2.57.0 (OpenAPI lint, chapter 9; Docker Hub tag pushed 2026-09-30, amd64+arm64) |

## GitHub Actions majors
actions/checkout@v7 · actions/setup-go@v7 · actions/setup-python@v7 · astral-sh/setup-uv@v10 · actions/cache@v6 ·
actions/upload-artifact@v7 · docker/setup-buildx-action@v4 · docker/build-push-action@v7 · docker/login-action@v4 ·
docker/metadata-action@v6 · helm/kind-action@v1 · azure/setup-helm@v5 · hashicorp/setup-terraform@v4 ·
golangci/golangci-lint-action@v9 · grafana/setup-k6-action@v1. Used by the workflows: checkout, setup-go, setup-uv,
setup-buildx, build-push, login, metadata, kind-action; the others were verified for reference only.

## Cost
Everything is free and local. The optional AWS section of chapter 22 uses free-tier resources and tells the reader
to set a budget alarm first; it is the only place a cloud account appears, and it is optional.

## Pins added by chapter writers (verified by them on 2026-10-05)
| Item | Pin | Chapter |
| --- | --- | --- |
| golang.org/x/sync | v0.23.0 | 5, 12 |
| ghcr.io/yannh/kubeconform | v0.8.0 | 20–21 |
| metrics-server | v0.9.0 (needs `--kubelet-insecure-tls` on kind) | 21 |
| cloud-provider-kind | v0.12.0 (Ingress on kind; ingress-nginx retired March 2026) | 21 |
| busybox | 1.37.0 | 21 |
| hashicorp/aws provider | 6.67.0 (optional section, not applied) | 22 |
| helm/kind-action | v1.15.0 | 23 |
| softprops/action-gh-release | v3 (mentioned; `gh release create` used) | 23 |
| koalaman/shellcheck | stable (0.11.0 on 2026-10-05) | 2, 23 |
| Rust toolchain | 1.99.0 (rustfmt, clippy; channel manifest dated 2026-10-01), installed user-space with rustup `--profile minimal` | 33 (optional) |
