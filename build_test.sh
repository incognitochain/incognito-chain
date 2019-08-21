#!/usr/bin/env bash
echo "Start build Incognito"
if [ "$1" == "y" ]; then
  rm -rfv data/*
fi
#git pull

echo "Package install"
dep ensure -v

APP_NAME="incognito"

cp blockchain/params.go blockchain/testparams/params
cp blockchain/testparams/paramstest blockchain/params.go
cp blockchain/constants.go blockchain/testparams/constants
cp blockchain/testparams/constantstest blockchain/constants.go

echo "go build -o $APP_NAME"
go build -o $APP_NAME

echo "cp ./$APP_NAME $GOPATH/bin/$APP_NAME"
cp ./$APP_NAME $GOPATH/bin/$APP_NAME

cp blockchain/testparams/params blockchain/params.go
cp blockchain/testparams/constants blockchain/constants.go

rm blockchain/testparams/params
rm blockchain/testparams/constants

echo "Build Incognito success!"