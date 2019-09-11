#!/usr/bin/env bash

if [ -f ./incognito ]; then
    rm -rf ./incognito
fi
if [ -f ./bootnode ]; then
    rm -rf ./bootnode
fi

cp ../blockchain/params.go ../blockchain/testparams/params
cp ../blockchain/testparams/paramstest ../blockchain/params.go
cp ../blockchain/constants.go ../blockchain/testparams/constants
cp ../blockchain/testparams/constantstest ../blockchain/constants.go

env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags '-w' -o incognito ../*.go
env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags '-w' -o bootnode ../bootnode/*.go
cp ../keylist.json .
cp ../sample-config.conf .

cp ../blockchain/testparams/params ../blockchain/params.go
cp ../blockchain/testparams/constants ../blockchain/constants.go
rm ../blockchain/testparams/params
rm ../blockchain/testparams/constants

commit=`git show --summary --oneline | cut -d ' ' -f 1`
docker build --build-arg commit=$commit . -t hungngoautonomous/incognito && docker push hungngoautonomous/incognito && echo "Commit: $commit"
docker rmi -f $(docker images --filter "dangling=true" -q)