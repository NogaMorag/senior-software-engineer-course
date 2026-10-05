# syntax=docker/dockerfile:1
# Multi-stage build for shorty: a full Go toolchain to compile, a distroless
# base to run. Build from this directory: docker build -t shorty:dev .

# ---- build stage ----
FROM golang:1.27.1 AS build
WORKDIR /src

# TODO(day 58): copy go.mod and go.sum only, then download the modules with a
# BuildKit cache mount on /go/pkg/mod. This layer must not depend on .go files.

# TODO(day 58): copy the rest of the source, then build ./cmd/shorty into
# /out/shorty: static (CGO_ENABLED=0), -trimpath, -ldflags='-s -w', with cache
# mounts on /go/pkg/mod and /root/.cache/go-build.

# ---- runtime stage ----
# TODO(day 58): start from gcr.io/distroless/static-debian13:nonroot, accept a
# GIT_SHA build argument and put it in the org.opencontainers.image.revision
# label, copy /out/shorty from the build stage to /shorty, run as the numeric
# user 65532:65532, set SHORTY_ADDR=:8080, expose 8080, and make /shorty the
# entrypoint (exec form).
