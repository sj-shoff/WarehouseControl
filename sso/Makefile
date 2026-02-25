.PHONY: run build migrate-up migrate-down docker-up docker-down proto

include .env
export

run:
	go run cmd/sso/main.go

build:
	go build -o bin/sso cmd/sso/main.go

docker-up:
	docker-compose up -d --build

docker-down:
	docker-compose down

migrate-up:
	goose -dir migrations postgres "postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@localhost:${POSTGRES_PORT}/${POSTGRES_DB}?sslmode=disable" up

migrate-down:
	goose -dir migrations postgres "postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@localhost:${POSTGRES_PORT}/${POSTGRES_DB}?sslmode=disable" down

proto:
	protoc --go_out=$(GEN_PATH) \
		--go_opt=paths=source_relative \
		--go-grpc_out=$(GEN_PATH) \
		--go-grpc_opt=paths=source_relative \
		-I=$(PROTO_PATH) \
		$(PROTO_PATH)/sso/v1/auth.proto


