# syntax=docker/dockerfile:1
# One image, three programs: shorty, linkworker and loadgen. The multi-stage
# build is chapter 19's; compose picks the program with "entrypoint".

FROM golang:1.27.1 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download
COPY . .
ARG VERSION=dev
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    for cmd in shorty linkworker loadgen; do \
      CGO_ENABLED=0 go build -trimpath -ldflags="-s -w -X main.version=${VERSION}" -o /out/$cmd ./cmd/$cmd || exit 1; \
    done

FROM gcr.io/distroless/static-debian13:nonroot
COPY --from=build /out/ /
USER 65532:65532
ENTRYPOINT ["/shorty"]
