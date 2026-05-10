.PHONY: build build-wrapper build-server install run-server test clean

BIN_DIR := bin

build: build-wrapper build-server

build-wrapper:
	go build -o $(BIN_DIR)/angelix ./cmd/angelix

build-server:
	go build -o $(BIN_DIR)/server ./cmd/server

install:
	go install ./cmd/angelix ./cmd/server

run-server:
	@set -a && . ./.env && set +a && go run ./cmd/server

test:
	go test ./...

clean:
	rm -rf $(BIN_DIR)
