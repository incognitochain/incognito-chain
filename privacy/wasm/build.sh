#!/usr/bin/env bash

echo "rm -rfv privacy.wasm"
rm -rfv privacy.wasm

echo "GOOS=js GOARCH=wasm go build -o privacy.wasm *.go"
GOOS=js GOARCH=wasm go build -o privacy.wasm *.go
