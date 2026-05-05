_default:
    @just --list

build:
    go build -o bin/server ./cmd/server
    go build -o bin/health ./cmd/cli

run:
    go run ./cmd/server

cli *ARGS:
    go run ./cmd/cli {{ARGS}}

test:
    go test ./...

lint:
    golangci-lint run

clean:
    rm -rf bin/

# Regenerate OpenAPI spec from swaggo annotations
openapi:
    go run github.com/swaggo/swag/cmd/swag@latest init \
        -g cmd/server/main.go \
        -o internal/spec/ \
        --parseDependency --parseInternal

docker-build:
    docker build -t health-connect .

docker-run:
    docker compose up
