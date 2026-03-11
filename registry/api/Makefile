.PHONY: dev test sqlc migrate build lint docker

# Dev stack: from repo root run: docker compose -f registry/docker-compose.dev.yml up
dev:
	docker compose -f ../docker-compose.dev.yml up

test:
	go test ./... -race -cover

sqlc:
	go run github.com/sqlc-dev/sqlc/cmd/sqlc@latest generate

migrate:
	migrate -path internal/db/migrations -database "$${DATABASE_URL}" up

build:
	go build -o bin/registry ./cmd/registry

lint:
	golangci-lint run

docker:
	docker build -t hivemind-registry-api .
