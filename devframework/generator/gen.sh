go run localRPC.go
go run remoteRPC.go
cat apispec.go | sed 's/package main/package devframework \/\/This file is auto generated. Please do not change if you dont know what you are doing/' > ../rpc_interfaces.go