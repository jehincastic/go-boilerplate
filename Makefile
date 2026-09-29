.PHONY: help run test vet lint migrate-up migrate-down migrate-create docker-up docker-down

-include .env
export

DB_DSN ?= postgres://app:pa55word@localhost:5432/app?sslmode=disable

## help: print this help message
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

## run: run the API server
run:
	go run ./cmd/api

## test: run unit tests
test:
	go test ./...

## vet: run go vet
vet:
	go vet ./...

## lint: run golangci-lint
lint:
	golangci-lint run

## migrate-up: apply all database migrations
migrate-up:
	migrate -path ./migrations -database "$(DB_DSN)" up

## migrate-down: roll back the latest migration
migrate-down:
	migrate -path ./migrations -database "$(DB_DSN)" down 1

## migrate-create name=<name>: create a new SQL migration
migrate-create:
	@test -n "$(name)" || (echo "usage: make migrate-create name=add_widget" && exit 1)
	migrate create -seq -ext sql -dir ./migrations $(name)

## docker-up: start Postgres and Redis, and apply migrations
docker-up:
	docker compose up -d

## docker-down: stop the compose stack
docker-down:
	docker compose down
