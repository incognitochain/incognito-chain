#!/usr/bin/env bash

if [ -f ./constant ]; then
    rm -rf ./constant
fi
if [ -f ./bootnode ]; then
    rm -rf ./bootnode
fi
env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags '-w' github.com/ninjadotorg/constant
env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags '-w' -o bootnode ../bootnode/*.go

docker build . -t dungvanautonomous/constant
