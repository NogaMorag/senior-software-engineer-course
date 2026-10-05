# Makefile for shorty. Every target is a short, documented command, so the
# same steps run on your machine and later in CI.

GO        ?= go
BIN      := bin/shorty
BASE_URL ?= http://localhost:8080

.PHONY: all build test lint run bench profile clean

all: lint test build

build:
	$(GO) build -o $(BIN) ./cmd/shorty

test:
	$(GO) test -race ./...

lint:
	@out="$$(gofmt -l .)"; if [ -n "$$out" ]; then echo "gofmt needed:"; echo "$$out"; exit 1; fi
	$(GO) vet ./...
	staticcheck ./...

run: build
	SHORTY_BASE_URL=$(BASE_URL) ./$(BIN)

bench:
	$(GO) test -run '^$$' -bench Resolve -benchmem ./internal/app/

# Needs a running service (make run) under load (go run ./cmd/loadgen ...).
profile:
	$(GO) tool pprof -top -nodecount 20 'http://127.0.0.1:6060/debug/pprof/profile?seconds=10'

clean:
	rm -rf bin
