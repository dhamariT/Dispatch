.PHONY: help db-up db-down db-reset migrate-up migrate-down build run tidy

DATABASE_URL ?= postgres://dispatch:dispatch@localhost:5432/dispatch?sslmode=disable

help:
	@echo "Dispatch dev tasks:"
	@echo "  make db-up        Start local Postgres via docker compose"
	@echo "  make db-down      Stop local Postgres"
	@echo "  make db-reset     Drop volume and re-create Postgres"
	@echo "  make migrate-up   Apply all pending migrations"
	@echo "  make migrate-down Roll back the most recent migration"
	@echo "  make build        Build dispatchd"
	@echo "  make run          Run dispatchd against local Postgres"

db-up:
	docker compose up -d postgres

db-down:
	docker compose down

db-reset:
	docker compose down -v
	docker compose up -d postgres

migrate-up:
	migrate -path db/migrations -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path db/migrations -database "$(DATABASE_URL)" down 1

build:
	go build -o bin/dispatchd ./cmd/dispatchd

run: build
	DISPATCH_DATABASE_URL="$(DATABASE_URL)" ./bin/dispatchd

tidy:
	go mod tidy
