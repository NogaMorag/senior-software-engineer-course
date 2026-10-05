# syntax=docker/dockerfile:1
# Multi-stage build for shorty (copied from chapter 19; chapter 21 adds the
# loadgen binary to the same image, so the load test can run inside the
# cluster). Build from this directory: docker build -t shorty:ch21 .

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
    CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o /out/ ./cmd/shorty ./cmd/loadgen

# ---- runtime stage: the binary and almost nothing else ----
FROM gcr.io/distroless/static-debian13:nonroot
ARG GIT_SHA=unknown
LABEL org.opencontainers.image.source="https://example.com/senior-course/ch21" \
      org.opencontainers.image.revision="${GIT_SHA}"
COPY --from=build /out/shorty /out/loadgen /
# Numeric, so Kubernetes can check runAsNonRoot (65532 is distroless "nonroot").
USER 65532:65532
ENV SHORTY_ADDR=:8080
EXPOSE 8080
ENTRYPOINT ["/shorty"]
