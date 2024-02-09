PCG_NAME=event
APP_NAME=main

BIN_DIR=build/bin
CMD_DIR=cmd
INT_DIR=internal
API_DIR=api
RPC_DIR=grpc

GO_EXT=go
GO_BUILD=go build

APP_PATH=$(BIN_DIR)/$(APP_NAME)
APP_DIRS=$(shell find . -type d | grep "./cmd//*" | cut -f 3 -d /)

.PHONY: all
all: module proto compile

module:
	go mod tidy
proto:
	protoc --go_out=./$(API_DIR) $(API_DIR)/$(RPC_DIR)/$(PCG_NAME).proto &&	protoc --go-grpc_out=./$(API_DIR) $(API_DIR)/$(RPC_DIR)/$(PCG_NAME).proto
compile: 
	$(shell for DIR in $(APP_DIRS);	do $(GO_BUILD) -o $(BIN_DIR)/$(APP_NAME)/$DIR cmd/$(DIR)/main.go; done)
