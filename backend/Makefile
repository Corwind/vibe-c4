.PHONY: build test run lint clean

BINARY_NAME=vibe-c4-api
BUILD_DIR=bin

build:
	go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/server/

test:
	go test ./... -v -race

run:
	go run ./cmd/server/

lint:
	golangci-lint run ./...

clean:
	rm -rf $(BUILD_DIR)
