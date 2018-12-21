package rpcserver

import (
	"fmt"
	"github.com/pkg/errors"
)

const (
	ErrUnexpected                 = iota
	ErrAlreadyStarted
	ErrRPCInvalidRequest
	ErrRPCMethodNotFound
	ErrRPCInvalidParams
	ErrRPCInvalidMethodPermission
	ErrRPCInternal
	ErrRPCParse
	ErrInvalidType
	ErrAuthFail
)

// Standard JSON-RPC 2.0 errors.
var ErrCodeMessage = map[int]struct {
	code    int
	message string
}{
	// rpc server error
	ErrUnexpected:     {-1, "Unexpected error"},
	ErrAlreadyStarted: {-2, "RPC server is already started"},

	// rpc api error
	ErrRPCInvalidRequest:          {-1001, "Invalid request"},
	ErrRPCMethodNotFound:          {-1002, "Method not found"},
	ErrRPCInvalidParams:           {-1003, "Invalid parameters"},
	ErrRPCInternal:                {-1004, "Internal error"},
	ErrRPCParse:                   {-1005, "Parse error"},
	ErrInvalidType:                {-1006, "Invalid type"},
	ErrAuthFail:                   {-1007, "Auth failure"},
	ErrRPCInvalidMethodPermission: {-1008, "Invalid method permission"},
}

// RPCError represents an error that is used as a part of a JSON-RPC Response
// object.
type RPCError struct {
	Code       int    `json:"Code,omitempty"`
	Message    string `json:"Message,omitempty"`
	err        error  `json:"Err"`
	StackTrace string `json:"StackTrace"`
}

// Guarantee RPCError satisifies the builtin error interface.
var _, _ error = RPCError{}, (*RPCError)(nil)

// Error returns a string describing the RPC error.  This satisifies the
// builtin error interface.
func (e RPCError) Error() string {
	return fmt.Sprintf("%d: %+v", e.Code, e.err)
}

func (e RPCError) GetErr() error {
	return e.err
}

// NewRPCError constructs and returns a new JSON-RPC error that is suitable
// for use in a JSON-RPC Response object.
func NewRPCError(key int, err error) *RPCError {
	return &RPCError{
		Code:    ErrCodeMessage[key].code,
		Message: ErrCodeMessage[key].message,
		err:     errors.Wrap(err, ErrCodeMessage[key].message),
	}
}
