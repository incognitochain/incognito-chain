#!/usr/bin/env bash
echo "Start build Incognito"

git pull

GO111MODULE=on go get -v

APP_NAME="incognito"

GO111MODULE=on go test -cover ./blockchain/committeestate/*.go
GO111MODULE=on go test -cover ./dataaccessobject/statedb/*.go

echo "go build -o $APP_NAME"
GO111MODULE=on go build -o $APP_NAME

echo "cp ./$APP_NAME $GOPATH/bin/$APP_NAME"
mv ./$APP_NAME $GOPATH/bin/$APP_NAME

echo "Build Incognito success!"
