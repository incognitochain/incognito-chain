#!/bin/sh
echo "Running linters"
golangci-lint run privacy/... transaction/... || exit 1
echo "Running unit tests"
go test ./privacy/operation || exit 1
go test ./privacy/coin || exit 1
go test ./privacy/key || exit 1
go test ./privacy/privacy_util || exit 1
go test ./privacy/privacy_v2 || exit 1
go test -v ./transaction/tx_ver2 || exit 1
echo "Success"