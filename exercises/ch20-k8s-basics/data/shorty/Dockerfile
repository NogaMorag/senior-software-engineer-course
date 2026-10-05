# syntax=docker/dockerfile:1
# Multi-stage build for shorty: a full Go toolchain to compile, a distroless
# base to run. Build from this directory: docker build -t shorty:dev .

# ---- build stage: compiler, module cache, source ----
FROM golang:1.27.1 AS build
WORKDIR /src

# Dependencies first: this layer stays cached until go.mod or go.sum change.
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Then the source. A change to a .go file rebuilds from here down only.
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o /out/shorty ./cmd/shorty

# ---- runtime stage: the binary and almost nothing else ----
FROM gcr.io/distroless/static-debian13:nonroot
ARG GIT_SHA=unknown
LABEL org.opencontainers.image.source="https://example.com/senior-course/ch19" \
      org.opencontainers.image.revision="${GIT_SHA}"
COPY --from=build /out/shorty /shorty
# Numeric, so Kubernetes can check runAsNonRoot (65532 is distroless "nonroot").
USER 65532:65532
ENV SHORTY_ADDR=:8080
EXPOSE 8080
ENTRYPOINT ["/shorty"]
