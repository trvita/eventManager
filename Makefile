PCG_NAME=event
APP_NAME=main

BIN_DIR=build/bin
CMD_DIRS := $(wildcard $(CMD_DIR)/*)
API_DIR=api
RPC_DIR=grpc

GO_EXT=go
GO_BUILD=go build

APP_PATH=$(BIN_DIR)/$(APP_NAME)
CMD_DIRS=$(wildcard $(CMD_DIR)/*/)

.PHONY: all module proto build

all: clean module proto build

module:
	go mod tidy

proto:
	protoc --go_out=./$(API_DIR) $(API_DIR)/$(RPC_DIR)/$(PCG_NAME).proto && \
	protoc --go-grpc_out=./$(API_DIR) $(API_DIR)/$(RPC_DIR)/$(PCG_NAME).proto

build: 
	DIRS=$$(find ./cmd -type d | cut -f 3 -d /);for DIR in $$DIRS; do go build -o build/bin/"$$DIR" cmd/"$$DIR"/main.go; done

clean:
	rm -rf $(BIN_DIR)
