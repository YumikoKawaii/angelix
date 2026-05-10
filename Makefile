.PHONY: build build-ui build-wrapper build-server install setup-claude run-server dev test clean

BIN_DIR := bin

build: build-ui build-server

build-ui:
	cd web && npm install && npm run build

build-wrapper:
	go build -o $(BIN_DIR)/angelix ./cmd/angelix

build-server: build-ui
	go build -o $(BIN_DIR)/server ./cmd/server

# Download the claude binary into ~/.angelix/bin/claude.
# Pass VERSION=x.y.z to pin a specific release, e.g.:  make setup-claude VERSION=2.1.138
setup-claude:
	./scripts/download-claude.sh $(VERSION)

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
