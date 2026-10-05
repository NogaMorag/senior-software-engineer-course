# syntax=docker/dockerfile:1
# The chapter 19 Dockerfile for shorty (multi-stage, distroless, non-root).
# In CI the OCI labels (source, revision, version) come from
# docker/metadata-action, so this file sets none of its own.

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
COPY --from=build /out/shorty /shorty
# Numeric, so Kubernetes can check runAsNonRoot (65532 is distroless "nonroot").
USER 65532:65532
ENV SHORTY_ADDR=:8080
EXPOSE 8080
ENTRYPOINT ["/shorty"]
