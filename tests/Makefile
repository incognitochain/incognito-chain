GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
BINARY_NAME=incognito-test
GOBIN = $(shell pwd)/build/bin
GO ?= latest

test:
	$(GOBUILD) -o $(BINARY_NAME)
	./$(BINARY_NAME) all
test: client
	$(GOBUILD) -o $(BINARY_NAME)
	./$(BINARY_NAME) client