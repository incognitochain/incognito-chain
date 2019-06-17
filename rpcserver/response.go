package rpcserver

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
)

// RPCResponse is the general form of a JSON-RPC response.  The type of the Result
// field varies from one command to the next, so it is implemented as an
// interface.  The Id field has to be a pointer for Go to put a null in it when
// empty.
type RPCResponse struct {
	Result json.RawMessage `json:"Result"`
	Error  *RPCError       `json:"Error"`
	Id     *interface{}    `json:"Id"`
}

// NewResponse returns a new JSON-RPC response object given the provided id,
// marshalled result, and RPC error.  This function is only provided in case the
// caller wants to construct raw responses for some reason.
//
// Typically callers will instead want to create the fully marshalled JSON-RPC
// response to send over the wire with the MarshalResponse function.
func NewResponse(id interface{}, marshalledResult []byte, rpcErr *RPCError) (*RPCResponse, error) {
	if !IsValidIDType(id) {
		str := fmt.Sprintf("The id of type '%T' is invalid", id)
		return nil, NewRPCError(ErrInvalidType, errors.New(str))
	}

	pid := &id
	resp := &RPCResponse{
		Result: marshalledResult,
		Error:  rpcErr,
		Id:     pid,
	}
	if resp.Error != nil {
		resp.Error.StackTrace = rpcErr.Error()
	}
	return resp, nil
}

// IsValidIDType checks that the Id field (which can go in any of the JSON-RPC
// requests, responses, or notifications) is valid.  JSON-RPC 1.0 allows any
// valid JSON type.  JSON-RPC 2.0 (which coind follows for some parts) only
// allows string, number, or null, so this function restricts the allowed types
// to that list.  This function is only provided in case the caller is manually
// marshalling for some reason.    The functions which accept an Id in this
// package already call this function to ensure the provided id is valid.
func IsValidIDType(id interface{}) bool {
	switch id.(type) {
	case int, int8, int16, int32, int64,
	uint, uint8, uint16, uint32, uint64,
	float32, float64,
	string,
	nil:
		return true
	default:
		return false
	}
}

// MarshalResponse marshals the passed id, result, and RPCError to a JSON-RPC
// response byte slice that is suitable for transmission to a JSON-RPC client.
func MarshalResponse(id interface{}, result interface{}, rpcErr *RPCError) ([]byte, error) {
	marshalledResult, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}
	response, err := NewResponse(id, marshalledResult, rpcErr)
	if err != nil {
		return nil, err
	}
	resultResp, err := json.MarshalIndent(&response, "", "\t")
	if err != nil {
		return nil, err
	}
	return resultResp, nil
}


// createMarshalledReply returns a new marshalled JSON-RPC response given the
// passed parameters.  It will automatically convert errors that are not of
// the type *btcjson.RPCError to the appropriate type as needed.
func createMarshalledReply(id, result interface{}, replyErr error) ([]byte, error) {
	var jsonErr *RPCError
	if replyErr != nil {
		if jErr, ok := replyErr.(*RPCError); ok {
			jsonErr = jErr
		} else {
			jsonErr = internalRPCError(replyErr.Error(), "")
		}
	}
	return MarshalResponse(id, result, jsonErr)
}

// internalRPCError is a convenience function to convert an internal error to
// an RPC error with the appropriate Code set.  It also logs the error to the
// RPC server subsystem since internal errors really should not occur.  The
// context parameter is only used in the log Message and may be empty if it's
// not needed.
func internalRPCError(errStr, context string) *RPCError {
	logStr := errStr
	if context != "" {
		logStr = context + ": " + errStr
	}
	Logger.log.Info(logStr)
	return NewRPCError(ErrRPCInternal, errors.New(errStr))
}
