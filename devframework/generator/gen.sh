go run localRPC.go
go run remoteRPC.go
cat apispec.go | sed 's/package main/package devframework/' > ../rpc_interfaces.go