#!/usr/bin/env bash
echo "Start build Incognito Test"

git pull

echo "Package install"
dep ensure -v

APP_NAME="incognito-test"

echo "go build -o $APP_NAME"
go build -o $APP_NAME

echo "cp ./$APP_NAME $GOPATH/bin/$APP_NAME"
mv ./$APP_NAME $GOPATH/bin/$APP_NAME

echo "Build Incognito success!"