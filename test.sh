#!/usr/bin/env bash
echo "go test -cover -tags test -test.v"
go test ./... -cover -tags test -test.v
