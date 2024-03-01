PCG_NAME=event
APP_NAME=main

BIN_DIR=build
CMD_DIRS := $(wildcard $(CMD_DIR)/*)
API_DIR=api
RPC_DIR=grpc

GO_EXT=go
GO_BUILD=go build

APP_PATH=$(BIN_DIR)/$(APP_NAME)
CMD_DIRS=$(wildcard $(CMD_DIR)/*/)

.PHONY: all module proto build

all: clean module proto build broker

module:
	go mod tidy

proto:
	protoc --go_out=./$(API_DIR) $(API_DIR)/$(RPC_DIR)/$(PCG_NAME).proto && \
	protoc --go-grpc_out=./$(API_DIR) $(API_DIR)/$(RPC_DIR)/$(PCG_NAME).proto

build:
	go build -o ./$(BIN_DIR)/ ./...

run:
	./$(BIN_DIR)/$(PCG_NAME)_server

broker:
	docker start rabbitmq || docker run -it --rm --detach --name rabbitmq -p 5672:5672 -p 15672:15672 rabbitmq:3.12-management

test:
	gotestsum --format PCG_NAME --raw-command go test -json -cover ./...

clean:
	rm -rf $(BIN_DIR)
