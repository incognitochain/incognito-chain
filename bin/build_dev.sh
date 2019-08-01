#!/usr/bin/env bash

if [ -f ./incognito ]; then
    rm -rf ./incognito
fi
if [ -f ./bootnode ]; then
    rm -rf ./bootnode
fi
cp ../blockchain/testparams/paramstest.go ../blockchain/params.go
cp ../blockchain/testparams/constantstest.go ../blockchain/constants.go

env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags '-w' -o incognito ../*.go
env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags '-w' -o bootnode ../bootnode/*.go
cp ../keylist.json .
cp ../sample-config.conf .

commit=`git show --summary --oneline | cut -d ' ' -f 1`
docker build --build-arg commit=$commit . -t incognitochain/incognito && docker push incognitochain/incognito && echo "Commit: $commit"
docker rmi -f $(docker images --filter "dangling=true" -q)
