# This Makefile is meant to be used by people that do not usually work
# with Go source code. If you know what GOPATH is then you probably
# don't need to bother with make.

GOBUILD = env GO111MODULE=on go build
DATADIR = ./data
MAINNET = mainnet
TESTNET = testnet
VERSION_1 = 1
VERSION_2 = 2
LOCAL = local
BUILD_FILE_NAME = incognito

build:
	$(GOBUILD) -o $(BUILD_FILE_NAME) -v

local:
	INCOGNITO_NETWORK_KEY=$(LOCAL) ./$(BUILD_FILE_NAME) 2>&1 | tee local.log

testnet-1:
	INCOGNITO_NETWORK_KEY=$(TESTNET) INCOGNITO_NETWORK_VERSION_KEY=$(VERSION_1) ./$(BUILD_FILE_NAME) 2>&1 | tee testnet-1.log

testnet-2:
	INCOGNITO_NETWORK_KEY=$(TESTNET) INCOGNITO_NETWORK_VERSION_KEY=$(VERSION_2) ./$(BUILD_FILE_NAME) 2>&1 | tee testnet-2.log

mainnet:
	INCOGNITO_NETWORK_KEY=$(MAINNET) ./$(BUILD_FILE_NAME) 2>&1 | tee mainnet.log

prune-mainnet:
	INCOGNITO_NETWORK_KEY=$(MAINNET) ./$(BUILD_FILE_NAME) --stateprune

prune-local:
	INCOGNITO_NETWORK_KEY=$(LOCAL) ./$(BUILD_FILE_NAME) --stateprune


test:
	make build
	go test ./.../pdexv3
	go test ./.../pdex
	go test ./.../bridgeagg

clean:
	env GO111MODULE=on go clean -cache
	rm -rf $(DATADIR)
	rm -rf $(BUILD_FILE_NAME)
