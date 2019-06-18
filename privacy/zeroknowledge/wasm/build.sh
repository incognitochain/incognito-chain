#!/usr/bin/env bash

echo "rm -rfv privacy.wasm"
rm -rfv privacy.wasm

echo "GOOS=js GOARCH=wasm go build -o privacy.wasm *.go"
GOOS=js GOARCH=wasm go build -o privacy.wasm *.go

ECHO "rm ./wasm_exec.js"
rm ./wasm_exec.js

echo "cp $GOROOT/misc/wasm/wasm_exec.js ./wasm_exec.js"
cp $GOROOT/misc/wasm/wasm_exec.js ./wasm_exec.js
