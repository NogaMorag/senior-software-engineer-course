# Tool versions verified 2026-10-05

Every row was checked on **2026-10-05** against a primary source only: an official download page or JSON feed, the GitHub Releases API (`gh api repos/<o>/<r>/releases/latest`), the PyPI JSON API, the Docker Hub tags API, or `proxy.golang.org/<module>/@latest`. Dates in parentheses are the upstream publish dates.

Conventions:
- "GH" means `https://api.github.com/repos/<o>/<r>/releases/latest`.
- "proxy" means `https://proxy.golang.org/<module>/@latest`.
- "PyPI" means `https://pypi.org/pypi/<pkg>/json`.
- "Hub" means `https://hub.docker.com/v2/repositories/<ns>/<repo>/tags?page_size=100&ordering=last_updated`.
- Arch notes come from the release's asset list or the image manifest's architecture list.

## Languages, Kubernetes, IaC

| Tool | Latest stable | Source URL | Date checked | Notes |
|---|---|---|---|---|
| Go | **go1.27.1** (go1.26.8 is the previous supported line) | https://go.dev/dl/?mode=json | 2026-10-05 | Archives for linux/darwin/windows × amd64/arm64. On WSL2, use the linux tarball inside the distro. |
| Kubernetes | **v1.37.1** (2026-09-23) | https://dl.k8s.io/release/stable.txt | 2026-10-05 | GH kubernetes/kubernetes agrees. |
| kubectl | **v1.37.1** | https://dl.k8s.io/release/stable.txt; binary at `https://dl.k8s.io/release/v1.37.1/bin/linux/{amd64,arm64}/kubectl` | 2026-10-05 | The linux/arm64 binary URL returned HTTP 200. |
| kind | **v0.33.0** (2026-08-26) | GH kubernetes-sigs/kind | 2026-10-05 | Assets: kind-linux-amd64, kind-linux-arm64, kind-windows-amd64 (plus darwin builds). |
| kindest/node (kind v0.33.0 default) | **v1.37.0** `@sha256:a1ed56cfb0e7b93589bdf97c8cd566405a265939e3620fc4f5de89adff580ae5` | kind v0.33.0 release notes (GH) | 2026-10-05 | Prebuilt images: v1.37.0, v1.36.4, v1.35.8, v1.34.11. The release notes say you **must** pin by digest. Images are amd64 and arm64, and the image must match the host arch. |
| Helm | **v4.3.0** (2026-09-09) | GH helm/helm; https://get.helm.sh | 2026-10-05 | Helm **4** major. Assets include helm-v4.3.0-linux-amd64 and -linux-arm64 tar.gz files with sha256sum files. |
| Terraform | **1.16.5** (2026-10-02) | https://checkpoint-api.hashicorp.com/v1/check/terraform | 2026-10-05 | `releases.hashicorp.com/terraform/1.16.5/index.json` lists linux/darwin/windows amd64 and arm64 builds, among others. |
| terraform-provider-kubernetes (hashicorp/kubernetes) | **v3.3.0** (2026-10-01) | GH hashicorp/terraform-provider-kubernetes | 2026-10-05 | Provider major **3**. |
| terraform-provider-helm (hashicorp/helm) | **v3.3.0** (2026-09-02) | GH hashicorp/terraform-provider-helm | 2026-10-05 | Provider major **3**. |
| Python (CPython) | **3.14.8** (2026-09-30); 3.15 is still at **3.15.0rc2** | https://www.python.org/api/v2/downloads/release/?is_published=true&pre_release=false | 2026-10-05 | 3.15 is not GA as of the check date. |

## Go toolchain / quality / security CLIs

| Tool | Latest stable | Source URL | Date checked | Notes |
|---|---|---|---|---|
| k6 | **v2.3.0** (2026-09-21) | GH grafana/k6 | 2026-10-05 | k6 **2.x** major. Assets: linux amd64 tar.gz/deb/rpm, linux arm64 tar.gz, macOS amd64/arm64, windows amd64 msi/zip. |
| actionlint | **v1.7.12** (2026-03-30) | GH rhysd/actionlint; proxy `github.com/rhysd/actionlint` = v1.7.12 | 2026-10-05 | linux/windows amd64 and arm64. |
| staticcheck | **2026.2.1** (2026-08-21) | GH dominikh/go-tools | 2026-10-05 | Module `honnef.co/go/tools` resolves to **v0.8.1** (tag 2026.2.1) via proxy. Assets: staticcheck_linux_{amd64,arm64}. |
| golangci-lint | **v2.14.0** (2026-09-24) | GH golangci/golangci-lint; proxy `github.com/golangci/golangci-lint/v2` = v2.14.0 | 2026-10-05 | v2 config format. linux/windows amd64 and arm64. |
| govulncheck (golang.org/x/vuln) | **v1.8.0** (2026-09-08) | proxy `golang.org/x/vuln` | 2026-10-05 | GH `golang/vuln/releases/latest` is stale (it returns v1.1.4 from 2025-01). The proxy and the repo tags (v1.8.0) are authoritative. Install with `go install golang.org/x/vuln/cmd/govulncheck@v1.8.0`. |
| buf CLI | **v1.73.0** (2026-09-11) | GH bufbuild/buf; proxy `github.com/bufbuild/buf` | 2026-10-05 | Assets: buf-Linux-x86_64 and buf-Linux-aarch64. |
| protoc-gen-go | **v1.36.12** (2026-08-10) | proxy `google.golang.org/protobuf` | 2026-10-05 | Install with `go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.12`. Related: `protoc-gen-go-grpc` is **v1.6.2** (proxy `google.golang.org/grpc/cmd/protoc-gen-go-grpc`), and protoc (GH protocolbuffers/protobuf) is **v36.2**. |
| Syft | **v1.54.0** (2026-10-01) | GH anchore/syft | 2026-10-05 | linux/windows amd64 and arm64. |
| Grype | **v0.120.0** (2026-10-02) | GH anchore/grype | 2026-10-05 | linux amd64 and arm64. |
| Trivy | **v0.75.0** (2026-10-01) | GH aquasecurity/trivy | 2026-10-05 | Assets: Linux-64bit and Linux-ARM64. |
| Docker Compose (plugin) | **v5.6.0** (2026-10-02) | GH docker/compose | 2026-10-05 | Extra row; Compose **v5** major. |

## Docker Desktop / WSL2

| Tool | Latest stable | Source URL | Date checked | Notes |
|---|---|---|---|---|
| Docker Desktop | **4.93.0** (2026-09-28) | https://docs.docker.com/desktop/release-notes/ | 2026-10-05 | Bundles Docker Engine v29.8.1. Windows x86_64 is GA, and Windows Arm is **Early Access**. |
| Docker Desktop: WSL2 requirements | n/a | https://docs.docker.com/desktop/setup/install/windows-install/ | 2026-10-05 | **WSL ≥ 2.1.5**. WSL ≥ 2.6 is required for Enhanced Container Isolation (kernel ≥ 6.3; WSL 2.6+ ships 6.6). Windows 10 64-bit 22H2 (build 19045), or Windows 11 64-bit 23H2 (build 22631)+. The page lists "Enterprise, Pro, or Education" editions. The LanmanServer service must be enabled and set to Automatic. Hardware: SLAT CPU, 8 GB RAM, virtualization enabled in BIOS/UEFI. Per-user install (the default) needs no admin and uses the WSL 2 or Docker VMM backend. Hyper-V needs the all-users install. |
| WSL | **3.0.1** (2026-09-29) | GH microsoft/WSL | 2026-10-05 | Install with `wsl --install` from an elevated PowerShell, then reboot. Use `wsl --install -d <Distro>` to pick a distro and `wsl --list --online` to list them (https://learn.microsoft.com/en-us/windows/wsl/install). |
| WSL default Ubuntu | **Ubuntu 26.04.1 LTS** ("resolute") | https://raw.githubusercontent.com/microsoft/WSL/master/distributions/DistributionInfo.json; https://changelogs.ubuntu.com/meta-release-lts | 2026-10-05 | The MS docs say the default distro is "Ubuntu". In the WSL manifest, `Ubuntu` (Default=true) points to `ubuntu-26.04.1-wsl-amd64.wsl` and `-arm64.wsl`. Pinned alternatives: Ubuntu-26.04, -24.04 (24.04.5), -22.04, -20.04. |

## Python packages (PyPI)

| Package | Latest stable | Source URL | Date checked | Notes (upload date, requires_python) |
|---|---|---|---|---|
| uv | **0.12.23** | PyPI uv; GH astral-sh/uv | 2026-10-05 | 2026-10-03. Standalone linux x86_64 and aarch64 builds (gnu and musl) and windows x86_64 and aarch64 builds. |
| ruff | **0.16.10** | PyPI ruff; GH astral-sh/ruff | 2026-10-05 | 2026-10-01 |
| pytest | **9.1.1** | https://pypi.org/pypi/pytest/json | 2026-10-05 | 2026-06-19, >=3.10 |
| pydantic | **2.13.5** | https://pypi.org/pypi/pydantic/json | 2026-10-05 | 2026-08-28, >=3.9 |
| fastapi | **0.142.2** | https://pypi.org/pypi/fastapi/json | 2026-10-05 | 2026-09-30, >=3.10 |
| uvicorn | **0.54.0** | https://pypi.org/pypi/uvicorn/json | 2026-10-05 | 2026-09-25, >=3.10 |
| httpx | **0.28.1** | https://pypi.org/pypi/httpx/json | 2026-10-05 | 2024-12-06 (no newer release), >=3.8 |
| mypy | **2.4.0** | https://pypi.org/pypi/mypy/json | 2026-10-05 | 2026-10-01, >=3.10; mypy **2.x** major |
| hypothesis | **6.168.4** | https://pypi.org/pypi/hypothesis/json | 2026-10-05 | 2026-10-05, >=3.10 (releases very frequently) |
| asyncpg | **0.31.0** | https://pypi.org/pypi/asyncpg/json | 2026-10-05 | 2025-11-24, >=3.9 |
| psycopg (v3) | **3.3.6** | https://pypi.org/pypi/psycopg/json | 2026-10-05 | 2026-09-18, >=3.10 |
| sqlalchemy | **2.1.3** | https://pypi.org/pypi/sqlalchemy/json | 2026-10-05 | 2026-10-03, **>=3.11** |
| alembic | **1.20.0** | https://pypi.org/pypi/alembic/json | 2026-10-05 | 2026-09-11, >=3.10 |
| redis (redis-py) | **8.1.0** | https://pypi.org/pypi/redis/json | 2026-10-05 | 2026-07-30, >=3.10 |
| aiokafka | **0.14.0** | https://pypi.org/pypi/aiokafka/json | 2026-10-05 | 2026-04-29, >=3.10 |
| confluent-kafka | **2.15.1** | https://pypi.org/pypi/confluent-kafka/json | 2026-10-05 | 2026-09-10, >=3.8 |
| opentelemetry-sdk / -api | **1.45.0** | https://pypi.org/pypi/opentelemetry-sdk/json | 2026-10-05 | 2026-09-25, >=3.10; GH open-telemetry/opentelemetry-python v1.45.0 agrees |

## Go modules (proxy.golang.org)

| Module | Latest | Source URL | Date checked | Notes (publish time) |
|---|---|---|---|---|
| go.opentelemetry.io/otel (and /sdk) | **v1.47.0** | proxy | 2026-10-05 | 2026-10-02. contrib `otelhttp` is **v0.72.0**. |
| OTel Go semconv package | **`go.opentelemetry.io/otel/semconv/v1.43.0`** | GH contents `open-telemetry/opentelemetry-go/semconv?ref=v1.47.0`; `sdk/resource/builtin.go` @ v1.47.0 | 2026-10-05 | v1.43.0 is the newest semconv directory in otel v1.47.0, and it is the one the SDK's own `sdk/resource` imports. It is a package inside the `go.opentelemetry.io/otel` module, not a separate module (proxy returns "not found" for `.../semconv`). It also has typed sub-packages such as `semconv/v1.43.0/httpconv`. Keep the semconv version aligned with the SDK's. |
| github.com/prometheus/client_golang | **v1.24.1** | proxy | 2026-10-05 | 2026-07-24 |
| github.com/testcontainers/testcontainers-go | **v0.44.0** | proxy; GH | 2026-10-05 | 2026-08-07 |
| github.com/jackc/pgx/v5 | **v5.11.0** | proxy | 2026-10-05 | 2026-09-07 |
| github.com/redis/go-redis/v9 | **v9.22.0** | proxy | 2026-10-05 | 2026-08-03 |
| github.com/twmb/franz-go | **v1.22.1** | proxy | 2026-10-05 | 2026-09-27. The repo publishes no GitHub Releases (GH returns 404); tags match the proxy. |
| github.com/pressly/goose/v3 | **v3.28.0** | proxy | 2026-10-05 | 2026-09-02 |
| github.com/golang-migrate/migrate/v4 | **v4.20.1** | proxy | 2026-10-05 | 2026-09-09 |
| google.golang.org/grpc | **v1.84.0** | proxy | 2026-10-05 | 2026-09-17 |
| google.golang.org/protobuf | **v1.36.12** | proxy | 2026-10-05 | 2026-08-10 |
| github.com/stretchr/testify | **v1.12.1** | proxy | 2026-10-05 | 2026-08-17 |
| pgregory.net/rapid | **v1.3.0** | proxy | 2026-10-05 | 2026-03-30 (repo flyingmutant/rapid) |
| github.com/sony/gobreaker/v2 | **v2.4.0** | proxy | 2026-10-05 | 2026-01-01. No GitHub Releases (404); tags only. |
| golang.org/x/time (rate) | **v0.16.0** | proxy | 2026-10-05 | 2026-08-19 |
| github.com/golang-jwt/jwt/v5 | **v5.3.1** | proxy | 2026-10-05 | 2026-01-28 |
| github.com/coreos/go-oidc/v3 | **v3.21.0** | proxy | 2026-10-05 | 2026-09-01 |

## Container images

| Image | Latest stable tag | Source URL | Date checked | Notes |
|---|---|---|---|---|
| postgres | **18.6** (major tag `18`; `latest` = 18.6) | Hub library/postgres | 2026-10-05 | 19 is beta only (19beta4). Supported majors: 14.24, 15.19, 16.15, 17.11. Multi-arch, including amd64 and arm64. |
| redis | **8.10.2** (`8`, `latest`) | Hub library/redis | 2026-10-05 | Other maintained lines: 8.8.3, 8.6.7, 8.4.7. amd64 and arm64. |
| redpandadata/redpanda | **v26.2.3** (2026-09-17) | Hub redpandadata/redpanda | 2026-10-05 | `latest` = v26.2.3. GH releases/latest still says v26.2.2 (v26.2.3 has no GH release object), and docs plus Hub use v26.2.3. Maintenance lines: v26.1.18, v25.3.17. amd64 and arm64. |
| redpandadata/console | **v3.12.0** (2026-09-16) | Hub; GH redpanda-data/console | 2026-10-05 | amd64 and arm64 |
| Redpanda quickstart compose | `docker.redpanda.com/redpandadata/redpanda:v26.2.3` + `docker.redpanda.com/redpandadata/console:v3.12.0` | https://docs.redpanda.com/current/get-started/quick-start/ (bundle: https://docs.redpanda.com/redpanda-quickstart.tar.gz) | 2026-10-05 | Image references extracted from the rendered docs page. |
| prom/prometheus | **v3.15.0** (2026-09-25; `latest`, `v3`) | Hub; GH prometheus/prometheus | 2026-10-05 | `-distroless` and `-busybox` variants exist. v3.13.4 (2026-09-29) is a patch on an older line, not the newest. amd64 and arm64. |
| grafana/grafana | **13.2.3** (2026-09-29) | Hub; GH grafana/grafana v13.2.3 | 2026-10-05 | `-slim` and `-distroless` variants exist. Patched older lines: 13.1.7, 13.0.10, 12.4.12. amd64 and arm64. |
| grafana/loki | **3.7.8** (2026-09-17) | GH grafana/loki; Hub tag `3.7.8` | 2026-10-05 | amd64, arm64, arm |
| grafana/alloy | **v1.20.1** (2026-09-28) | Hub; GH grafana/alloy | 2026-10-05 | amd64, arm64, ppc64le, s390x. `-distroless` and `-boringcrypto` variants exist. |
| grafana/tempo | **3.1.0** (2026-09-29) | Hub; GH grafana/tempo v3.1.0 | 2026-10-05 | Tempo **3.x** major. amd64 and arm64. |
| otel/opentelemetry-collector-contrib | **0.162.0** (2026-09-29) | GH open-telemetry/opentelemetry-collector-releases; GHCR manifest | 2026-10-05 | **Caveat:** on Docker Hub, 0.162.0 currently exists **only as per-arch tags** (`0.162.0-amd64`, `-arm64`, …), the multi-arch tag `0.162.0` returns 404, and `latest` was last pushed 2026-09-16 (= 0.161.0). The multi-arch image `ghcr.io/open-telemetry/opentelemetry-collector-releases/opentelemetry-collector-contrib:0.162.0` exists (amd64, arm64, arm, 386, ppc64le, riscv64, s390x). On Docker Hub, the newest multi-arch tag is `0.161.0`. |
| gcr.io/distroless/static-debian13 | `latest` / `nonroot` (also `debug`, `debug-nonroot`) | https://github.com/GoogleContainerTools/distroless README; `gcr.io/v2/distroless/static-debian13/manifests/latest` → 200 | 2026-10-05 | **debian13 is now the documented default.** The README examples use `static-debian13`. Archs: amd64, arm64, arm, s390x, ppc64le, riscv64. `static-debian12` still resolves (HTTP 200). |
| golang | **1.27.1** (`1.27`, `1`, `1.27.1-trixie`, `-bookworm`, `-alpine`, `-alpine3.24`) | Hub library/golang | 2026-10-05 | 1.26.8 is also published. |
| python | **3.14.8** (`3.14`, `3.14-slim`, `3.14-slim-trixie`, `3.14-alpine`, …) | Hub library/python | 2026-10-05 | `3.15-rc` / `3.15.0rc2` is a pre-release. Other maintained lines: 3.13.16, 3.12.15, 3.11.17, 3.10.22. |
| kindest/node | **v1.37.0** (kind v0.33.0 default; digest above) | Hub kindest/node; kind release notes | 2026-10-05 | amd64 and arm64 |

## GitHub Actions (latest release → major to pin)

All from `GH <owner>/<repo>` on 2026-10-05.

| Action | Latest release | Major | Published |
|---|---|---|---|
| actions/checkout | v7.0.1 | **v7** | 2026-07-20 |
| actions/setup-go | v7.0.0 | **v7** | 2026-07-16 |
| actions/setup-python | v7.0.0 | **v7** | 2026-07-20 |
| astral-sh/setup-uv | v10.2.0 | **v10** | 2026-09-21 |
| actions/cache | v6.1.0 | **v6** | 2026-06-26 |
| actions/upload-artifact | v7.0.1 | **v7** | 2026-04-10 |
| docker/setup-buildx-action | v4.4.1 | **v4** | 2026-09-16 |
| docker/build-push-action | v7.4.0 | **v7** | 2026-09-15 |
| docker/login-action | v4.6.0 | **v4** | 2026-07-29 |
| docker/metadata-action | v6.2.0 | **v6** | 2026-07-02 |
| helm/kind-action | v1.15.0 | **v1** | 2026-09-02 |
| azure/setup-helm | v5.0.1 | **v5** | 2026-06-23 |
| hashicorp/setup-terraform | v4.0.1 | **v4** | 2026-05-12 |
| golangci/golangci-lint-action | v9.3.0 | **v9** | 2026-06-29 |
| grafana/setup-k6-action | v1.2.2 | **v1** | 2026-09-22 |

## Source caveats (read before citing)

- **GitHub "latest release" can lag the real latest.** This happens for golang/vuln (stale at v1.1.4 vs module v1.8.0) and redpanda-data/redpanda (v26.2.2 vs the v26.2.3 image and docs). For Go modules, use proxy.golang.org. For images, use the registry.
- **twmb/franz-go and sony/gobreaker publish no GitHub Releases**, so their versions come from proxy.golang.org.
- **The OTel Collector contrib multi-arch tag on Docker Hub lags** (see that row). Prefer GHCR or pin per-arch tags.
- **Every source was reachable.** No row is marked unverified.
