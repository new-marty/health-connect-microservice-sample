.PHONY: build run test lint clean

build:
	go build -o bin/server ./cmd/server
	go build -o bin/health ./cmd/cli

run:
	go run ./cmd/server

cli:
	go run ./cmd/cli $(ARGS)

test:
	go test ./...

lint:
	golangci-lint run

clean:
	rm -rf bin/

docker-build:
	docker build -t health-connect .

docker-run:
	docker-compose up

.DEFAULT_GOAL := build
