#!/bin/bash

if [ -f ./i ]; then
    rm -rf ./incognito
fi
if [ -f ./bootnode ]; then
    rm -rf ./bootnode
fi
echo "!23"

env CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -ldflags '-w' -o incognito ../*.go
cp ../keylist.json .
cp ../keylist_256.json .
cp ../sample-config.conf .

commit=`git show --summary --oneline | cut -d ' ' -f 1`
docker build --build-arg commit=$commit . -t incognito && echo "Commit: $commit"
