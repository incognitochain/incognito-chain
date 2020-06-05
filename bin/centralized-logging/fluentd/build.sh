#!/usr/bin/env bash

if [ -f ./server ]; then
    rm -rf ./server
fi

env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags '-w' -o server *.go
docker build . -t incognitochain/incognito-logger && docker push incognitochain/incognito-logger
docker rmi -f $(docker images --filter "dangling=true" -q)
