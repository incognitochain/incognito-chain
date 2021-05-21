# This Makefile is meant to be used by people that do not usually work
# with Go source code. If you know what GOPATH is then you probably
# don't need to bother with make.

GOBUILD = env GO111MODULE=on go build
DATADIR = ./data
MAINNET = mainnet
TESTNET = testnet
TESTNETVERSION2 = 2
LOCAL = local
BUILD_FILE_NAME = incognito
FILE_MODE = file

build:
	$(GOBUILD) -v -o $(BUILD_FILE_NAME)

local:
	INCOGNITO_NETWORK_KEY=$(LOCAL) INCOGNITO_CONFIG_MODE_KEY=$(FILE_MODE)  ./$(BUILD_FILE_NAME)

testnet2:
	INCOGNITO_NETWORK_KEY=$(TESTNET) INCOGNITO_CONFIG_MODE_KEY=$(FILE_MODE) INCOGNITO_NETWORK_VERSION_KEY=$(TESTNETVERSION2) ./$(BUILD_FILE_NAME)

mainnet:
	INCOGNITO_NETWORK_KEY=$(MAINNET) INCOGNITO_CONFIG_MODE_KEY=$(FILE_MODE) ./$(BUILD_FILE_NAME)

test: 
	go test ./...

clean:
	env GO111MODULE=on go clean -cache
	rm -rf $(DATADIR)
