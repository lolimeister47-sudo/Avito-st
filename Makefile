APP_NAME=pr-service
CMD_PATH=./cmd/pr-service

ifeq ($(OS),Windows_NT)
	BIN_DIR=bin
	BIN=$(BIN_DIR)/$(APP_NAME).exe
	MKDIR_P=if not exist $(BIN_DIR) mkdir $(BIN_DIR)
else
	BIN_DIR=./bin
	BIN=$(BIN_DIR)/$(APP_NAME)
	MKDIR_P=mkdir -p $(BIN_DIR)
endif

.PHONY: build run test lint generate docker-build up

build:
	$(MKDIR_P)
	go build -o $(BIN) $(CMD_PATH)

run: build
	$(BIN)

test:
	go test ./...

lint:
	golangci-lint run ./...

generate:
	go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest \
		-generate types,chi-server \
		-package api \
		-o internal/adapter/http/api/openapi.gen.go \
		api/openapi.yaml

docker-build:
	docker build -t pr-service:local .

up:
	docker-compose up --build
