.PHONY: build build-ui build-wrapper build-server install run-server dev test clean

BIN_DIR := bin

build: build-ui build-server

build-ui:
	cd web && npm install && npm run build

build-wrapper:
	go build -o $(BIN_DIR)/angelix ./cmd/angelix

build-server: build-ui
	go build -o $(BIN_DIR)/server ./cmd/server

install:
	go install ./cmd/angelix ./cmd/server

run-server:
	@set -a && . ./.env && set +a && go run ./cmd/server

dev:
	@echo "Start the server:  make run-server"
	@echo "Start the UI:      cd web && npm run dev  (http://localhost:5173)"

test:
	go test ./...

clean:
	rm -rf $(BIN_DIR) web/dist/assets web/node_modules
