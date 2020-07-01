#!/usr/bin/env bash

docker login

if [ -f ./incognito ]; then
    rm -rf ./incognito
fi
if [ -f ./bootnode ]; then
    rm -rf ./bootnode
fi
env CGO_ENABLED=0 GOOS=linux GOARM=7 GOARCH=arm go build -ldflags '-w' -o incognito ../*.go
env CGO_ENABLED=0 GOOS=linux GOARM=7 GOARCH=arm go build -ldflags '-w' -o bootnode ../bootnode/*.go
env CGO_ENABLED=0 GOOS=linux GOARM=7 GOARCH=arm ENV=testnet go build -ldflags '-w' -o incognito-test ../tests/*.go
cp ../keylist-mainnet.json ./keylist.json
cp ../keylist-mainnet-v2.json ./keylist-v2.json
cp ../sample-config.conf .

commit=`git show --summary --oneline | cut -d ' ' -f 1`
docker buildx build -f Dockerfile-mainnet --platform linux/arm64,linux/amd64 --build-arg commit=$commit --push -t incognitochain/incognito-mainnet-arm:${tag} .
#docker push incognitochain/incognito-mainnet-arm:${tag} && echo "Commit: $commit"
#docker rmi -f $(docker images --filter "dangling=true" -q)
