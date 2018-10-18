package rpcserver

// RpcRequest is a type for raw JSON-RPC 1.0 requests.  The Method field identifies
// the specific command type which in turns leads to different parameters.
// Callers typically will not use this directly since this package provides a
// statically typed command infrastructure which handles creation of these
// requests, however this struct it being exported in case the caller wants to
// construct raw requests for some reason.

type RpcRequest struct {
	Jsonrpc string      `json:"Jsonrpc"`
	Method  string      `json:"Method"`
	Params  interface{} `json:"Params"`
	Id      interface{} `json:"Id"`
}
