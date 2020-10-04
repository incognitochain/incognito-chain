#!/usr/bin/env bash

echo "rm -rfv ./build/privacy.wasm"
rm -rfv ./build/privacy.wasm

echo "GOOS=js GOARCH=wasm go build -o ./build/privacy.wasm *.go"
GOOS=js GOARCH=wasm go build -o ./build/privacy.wasm *.go