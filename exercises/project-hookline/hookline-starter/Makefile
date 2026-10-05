# Makefile for Hookline. Every target is one short, documented command, so the same steps run on your
# machine and, from milestone M7, in CI.

GO           ?= go
COMPOSE      := docker compose -p hookline -f deploy/compose.yaml
DATABASE_URL ?= postgres://hookline:change-me@localhost:5438/hookline
export DATABASE_URL

.PHONY: all build test test-unit lint db-up db-down migrate tenant run run-dev clean

all: lint test-unit build

build:
	$(GO) build -o bin/ ./cmd/...

# Unit tests only: DATABASE_URL is cleared, integration tests skip, the API tests use the memory store.
test-unit:
	DATABASE_URL= $(GO) test -race ./...

# Everything, against the compose database (make db-up first).
test:
	$(GO) test -race -count=1 ./...

lint:
	@out="$$(gofmt -l .)"; if [ -n "$$out" ]; then echo "gofmt needed:"; echo "$$out"; exit 1; fi
	$(GO) vet ./...
	staticcheck ./...

db-up:
	$(COMPOSE) up -d --wait

db-down:
	$(COMPOSE) down -v

migrate:
	$(GO) run ./cmd/hookline-admin migrate up

# make tenant NAME=acme
tenant:
	$(GO) run ./cmd/hookline-admin create-tenant -name "$(NAME)"

run:
	$(GO) run ./cmd/api

# Memory store and dev mode: no database needed; prints a development API key at start.
run-dev:
	DATABASE_URL= HOOKLINE_DEV_MODE=true $(GO) run ./cmd/api

clean:
	rm -rf bin
