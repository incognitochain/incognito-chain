#!/usr/bin/env bash
echo "Start build Incognito"

git pull

echo "Package install"
dep ensure -v

APP_NAME="incognito"

cp blockchain/params.go blockchain/testparams/params
cp blockchain/testparams/paramstest blockchain/params.go
cp blockchain/constants.go blockchain/testparams/constants
cp blockchain/testparams/constantstest blockchain/constants.go
cp common/constants.go blockchain/testparams/commonconstants
cp blockchain/testparams/commonconstantstest common/constants.go

echo "go build -o $APP_NAME"
go build -o $APP_NAME

echo "cp ./$APP_NAME $GOPATH/bin/$APP_NAME"
cp ./$APP_NAME $GOPATH/bin/$APP_NAME

cp blockchain/testparams/params blockchain/params.go
cp blockchain/testparams/constants blockchain/constants.go
cp blockchain/testparams/commonconstants common/constants.go
rm blockchain/testparams/params
rm blockchain/testparams/constants
rm blockchain/testparams/commonconstants

echo "Build Incognito success!"