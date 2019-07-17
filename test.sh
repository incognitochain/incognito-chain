#!/usr/bin/env bash
echo "go test -cover -tags test -test.v"
go test ./... -coverprofile=cover.out -cover -tags test -test.v

go tool cover -html=cover.out
