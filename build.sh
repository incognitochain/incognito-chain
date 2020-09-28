#!/usr/bin/env bash
echo "Start Install Dependencies Package"
GO111MODULE=on go get -v

echo "Start Unit-Test"
GO111MODULE=on go test -cover ./blockchain/committeestate/*.go
GO111MODULE=on go test -cover ./dataaccessobject/statedb/*.go
GO111MODULE=on go test -cover ./instruction/*.go

echo "Start build Incognito"
APP_NAME="incognito"
echo "go build -o $APP_NAME"
GO111MODULE=on go build -o $APP_NAME

echo "cp ./$APP_NAME $GOPATH/bin/$APP_NAME"
mv ./$APP_NAME $GOPATH/bin/$APP_NAME

echo "Build Incognito success!"
