#!/usr/bin/env bash
protoc -I . proxy_register.proto --go_out=plugins=grpc:../.
