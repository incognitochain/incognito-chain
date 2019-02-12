#!/usr/bin/env bash

if [ -f ./constant ]; then
    rm -rf ./constant
fi
if [ -f ./bootnode ]; then
    rm -rf ./bootnode
fi
env GOOS=linux GOARCH=amd64 go build github.com/ninjadotorg/constant
env GOOS=linux GOARCH=amd64 go build -o bootnode ../bootnode/*.go

docker build . -t constant
