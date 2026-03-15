.PHONY: run build docker-up docker-down migrate-up migrate-down lint

include .env
export

run:
	go run cmd/warehouse-control/main.go

build:
	go build -o bin/warehouse-control cmd/warehouse-control/main.go

docker-up:
	docker-compose up -d --build

docker-down:
	docker-compose down

migrate-up:
	goose -dir migrations postgres "postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@localhost:${POSTGRES_PORT}/${POSTGRES_DB}?sslmode=disable" up

migrate-down:
	goose -dir migrations postgres "postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@localhost:${POSTGRES_PORT}/${POSTGRES_DB}?sslmode=disable" down

lint:
	golangci-lint run