#!/usr/bin/env bash

if [ -f ./incognito ]; then
    rm -rf ./incognito
fi
if [ -f ./bootnode ]; then
    rm -rf ./bootnode
fi
env CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -ldflags '-w' -o incognito ../*.go

cp ../keylist.json .
cp ../keylist-v2.json .
cp ../keylist_256.json .
cp ../sample-config.conf .

commit=`git show --summary --oneline | cut -d ' ' -f 1`
docker build --build-arg commit=$commit . -t incognitochainmainnet/incognito:syncmainnet
#docker rmi -f $(docker images --filter "dangling=true" -q)

