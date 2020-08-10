#!/usr/bin/env bash

if [ -f ./incognito ]; then
    rm -rf ./incognito
fi

# testnet-2 need to have a prefix on tagging
if [[ $tag != "testnet2-"* ]]; then
  tag="testnet2-${tag}"
fi

echo "Deploy docker tag ${tag}"

env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags '-w' -o incognito ../*.go
env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 ENV=testnet go build -ldflags '-w' -o incognito-test ../tests/*.go

# Genesis committee
cp ../keylist.json .

# Replace committee
cp ../keylist-v2.json .
echo "[]"> ./keylist-v2.json

cp ../sample-config.conf .

# Build docker and push
commit=`git show --summary --oneline | cut -d ' ' -f 1`
docker build --build-arg commit=$commit . -t incognitochain/incognito:${tag}

docker push incognitochain/incognito:${tag} && echo "Commit: $commit"
docker rmi -f $(docker images --filter "dangling=true" -q)
