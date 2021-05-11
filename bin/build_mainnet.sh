#!/usr/bin/env bash

docker login

if [ -f ./incognito ]; then
    rm -rf ./incognito
fi
if [ -f ./bootnode ]; then
    rm -rf ./bootnode
fi
env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags '-w' -o incognito ../*.go
env CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -ldflags '-w' -o regressiontest ../testsuite/apps/regressiontest/*.go

cp ../keylist-mainnet.json ./keylist.json
cp ../keylist-mainnet-v2.json ./keylist-v2.json
cp ../sample-config.conf .

commit=`git show --summary --oneline | cut -d ' ' -f 1`
docker build -f Dockerfile-mainnet --build-arg commit=$commit . -t incognitochain/incognito-mainnet:${tag} && docker push incognitochain/incognito-mainnet:${tag} && echo "Commit: $commit"
docker rmi -f $(docker images --filter "dangling=true" -q)
