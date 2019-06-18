package rpcserver

import (
	"encoding/json"
)

// JsonRequest is a type for raw JSON-RPC 1.0 requests.  The Method field identifies
// the specific command type which in turns leads to different parameters.
// Callers typically will not use this directly since this package provides a
// statically typed command infrastructure which handles creation of these
// requests, however this struct it being exported in case the caller wants to
// construct raw requests for some reason.

// The JSON-RPC 1.0 spec defines that notifications must have their "id"
// set to null and states that notifications do not have a response.
//
// A JSON-RPC 2.0 notification is a request with "json-rpc":"2.0", and
// without an "id" member. The specification states that notifications
// must not be responded to. JSON-RPC 2.0 permits the null value as a
// valid request id, therefore such requests are not notifications.
//
// coin Core serves requests with "id":null or even an absent "id",
// and responds to such requests with "id":null in the response.
//
// Rpc does not respond to any request without and "id" or "id":null,
// regardless the indicated JSON-RPC protocol version unless RPC quirks
// are enabled. With RPC quirks enabled, such requests will be responded
// to if the reqeust does not indicate JSON-RPC version.
//
// RPC quirks can be enabled by the user to avoid compatibility issues
// with software relying on Core's behavior.
type JsonRequest struct {
	Jsonrpc string      `json:"Jsonrpc"`
	Method  string      `json:"Method"`
	Params  interface{} `json:"Params"`
	Id      interface{} `json:"Id"`
}

func parseJsonRequest(rawMessage []byte) (*JsonRequest, error) {
	var request JsonRequest
	err := json.Unmarshal(rawMessage, &request)
	if err != nil {
		return &request, NewRPCError(ErrRPCParse, err)
	} else {
		return &request, nil
	}
}

type SubcriptionRequest struct {
	JsonRequest JsonRequest `json:"Request"`
	Subcription string      `json:"Subcription"`
}

func parseSubcriptionRequest(rawMessage []byte) (*SubcriptionRequest, error) {
	var request SubcriptionRequest
	err := json.Unmarshal(rawMessage, &request)
	if err != nil {
		return &request, NewRPCError(ErrRPCParse, err)
	} else {
		return &request, nil
	}
}
