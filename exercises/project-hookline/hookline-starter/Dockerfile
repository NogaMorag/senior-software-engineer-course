# syntax=docker/dockerfile:1
# One image with every Hookline binary (chapter 19): api, relay, dispatcher, worker and hookline-admin.
# Each Deployment of the Helm chart runs one of them as its command. Build from this directory:
#   docker build -t hookline:dev .

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
    CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o /out/ \
      ./cmd/api ./cmd/relay ./cmd/dispatcher ./cmd/worker ./cmd/hookline-admin

# ---- runtime stage: the binaries and almost nothing else ----
FROM gcr.io/distroless/static-debian13:nonroot
ARG GIT_SHA=unknown
LABEL org.opencontainers.image.source="https://example.com/senior-course/hookline" \
      org.opencontainers.image.revision="${GIT_SHA}"
COPY --from=build /out/ /usr/local/bin/
# Numeric, so Kubernetes can check runAsNonRoot (65532 is distroless "nonroot").
USER 65532:65532
ENV HOOKLINE_ADDR=:8080
EXPOSE 8080
HEALTHCHECK --interval=10s --timeout=3s --retries=3 CMD ["/usr/local/bin/api", "healthcheck"]
# The default command is the API; the chart sets the command of every other Deployment.
ENTRYPOINT ["/usr/local/bin/api"]
