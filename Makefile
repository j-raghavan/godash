BINARY_NAME=godash

all: build

build:
	go build -o $(BINARY_NAME) ./cmd/godash

run:
	go run ./cmd/godash

fmt:
	go fmt ./...

lint:
	golangci-lint run

test:
	go test ./...

dev: fmt test build
