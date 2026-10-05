# Makefile for Hookline. Every target is one short, documented command, so the same steps run on your
# machine and, from milestone M7, in CI.

GO           ?= go
COMPOSE      := docker compose -p hookline -f deploy/compose.yaml
DATABASE_URL  ?= postgres://hookline:change-me@localhost:5438/hookline
REDIS_ADDR    ?= localhost:6438
KAFKA_BROKERS ?= localhost:19438
SINK_URL      ?= http://localhost:8091
export DATABASE_URL REDIS_ADDR KAFKA_BROKERS SINK_URL

.PHONY: all build test test-unit lint up down db-up db-down migrate topic tenant run run-dev relay dispatcher worker sink-test lab-n3 images chart-check kind-smoke kind-down tf-check clean

all: lint test-unit build

build:
	$(GO) build -o bin/ ./cmd/...

# Unit tests only: the connection variables are cleared, integration tests skip, the API tests use the
# memory store.
test-unit:
	DATABASE_URL= REDIS_ADDR= KAFKA_BROKERS= SINK_URL= $(GO) test -race ./...

# Everything, against the compose stack (make up first).
test:
	$(GO) test -race -count=1 ./...

lint:
	@out="$$(gofmt -l .)"; if [ -n "$$out" ]; then echo "gofmt needed:"; echo "$$out"; exit 1; fi
	$(GO) vet ./...
	staticcheck ./...

# The whole stack: PostgreSQL, Redis, Redpanda and its console, the sink (built from sink/).
up:
	$(COMPOSE) up -d --wait --build

down:
	$(COMPOSE) down -v

# Only PostgreSQL (milestone M2).
db-up:
	$(COMPOSE) up -d --wait postgres

db-down: down

migrate:
	$(GO) run ./cmd/hookline-admin migrate up

topic:
	$(GO) run ./cmd/hookline-admin create-topic -partitions 6

# make tenant NAME=acme
tenant:
	$(GO) run ./cmd/hookline-admin create-tenant -name "$(NAME)"

run:
	$(GO) run ./cmd/api

# Memory store and dev mode: no database needed; prints a development API key at start.
run-dev:
	DATABASE_URL= HOOKLINE_DEV_MODE=true $(GO) run ./cmd/api

relay:
	$(GO) run ./cmd/relay

dispatcher:
	$(GO) run ./cmd/dispatcher

worker:
	$(GO) run ./cmd/worker

# Spec N3: one receiver hangs, another is healthy; prints the healthy one's delivery lag.
# GUARDS=off make lab-n3 runs the same without breakers and bulkheads.
lab-n3:
	scripts/lab-n3.sh

# Milestone M7: the two images, the chart checks, the kind smoke test, Terraform without a cluster.
images:
	docker build -t hookline:dev .
	docker build -t hookline-sink:dev sink

chart-check:
	scripts/check-chart.sh

# KEEP=1 make kind-smoke leaves the cluster running (for Terraform, or to look around).
kind-smoke:
	scripts/kind-smoke.sh

kind-down:
	kind delete cluster --name hookline

tf-check:
	cd deploy/terraform && terraform fmt -check -recursive && terraform init -backend=false -input=false && terraform validate

sink-test:
	cd sink && uv sync --frozen && uv run ruff check . && uv run mypy src && uv run pytest -q

clean:
	rm -rf bin
