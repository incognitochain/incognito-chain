#!/usr/bin/env bash

# if [ -z "$env" ]; then
#     env="testnet";
# fi

# commit=`git show --summary --oneline | cut -d ' ' -f 1`

# if [[ $env == "testnet" ]]; then
#     docker build --build-arg commit=$commit . -t incognitochaintestnet/incognito:${tag} && docker push incognitochaintestnet/incognito:${tag} && echo "Commit: $commit"
# elif [ $env == "mainnet" ]; then
#     docker build --build-arg commit=$commit . -t incognitochain/incognito-mainnet:${tag} && docker push incognitochain/incognito-mainnet:${tag} && echo "Commit: $commit"
# fi

# docker rmi -f $(docker images --filter "dangling=true" -q)



#!/usr/bin/env bash

if [ -f ./incognito ]; then
    rm -rf ./incognito
fi
if [ -f ./bootnode ]; then
    rm -rf ./bootnode
fi

env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags '-w' -o incognito

echo "build execuable file successfully"

if [ -z "$env" ]; then
    env="testnet";
fi

commit=`git show --summary --oneline | cut -d ' ' -f 1`

if [[ $env == "testnet" ]]; then
    docker build --build-arg commit=$commit . -t incognitochaintestnet/incognito:${tag} && docker push incognitochaintestnet/incognito:${tag} && echo "Commit: $commit"
elif [ $env == "mainnet" ]; then
    docker build --build-arg commit=$commit . -t incognitochain/incognito-mainnet:${tag} && docker push incognitochain/incognito-mainnet:${tag} && echo "Commit: $commit"
elif [ $env == "fixednodes" ]; then
    docker build --build-arg commit=$commit . -t incognitochain/incognito-fixed-nodes:${tag} && docker push incognitochain/incognito-fixed-nodes:${tag} && echo "Commit: $commit"
elif [ $env == "local" ]; then
    DOCKER_BUILDKIT=1 docker build --build-arg commit=$commit . -t incognitochain:local
fi

docker rmi -f $(docker images --filter "dangling=true" -q)

