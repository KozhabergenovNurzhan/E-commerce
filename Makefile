.PHONY: run build test lint tidy docker-up docker-down

run:
	go run ./cmd/api

build:
	go build -o bin/ecommerce ./cmd/api

test:
	go test ./...

lint:
	golangci-lint run ./...

tidy:
	go mod tidy

docker-up:
	docker-compose up --build

docker-down:
	docker-compose down
