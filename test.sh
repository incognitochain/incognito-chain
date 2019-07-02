#!/usr/bin/env bash
echo "go test -cover -tags test"
go test ./... -cover -tags test
