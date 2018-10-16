package rpcserver

import "fmt"

const (
	ErrUnexpected        = "ErrUnexpected"
	ErrAlreadyStarted    = "ErrAlreadyStarted"
	ErrRPCInvalidRequest = "ErrRPCInvalidRequest"
	ErrRPCMethodNotFound = "ErrRPCMethodNotFound"
	ErrRPCInvalidParams  = "ErrRPCInvalidParams"
	ErrRPCInternal       = "ErrRPCInternal"
	ErrRPCParse          = "ErrRPCParse"
	ErrInvalidType       = "ErrInvalidType"
)

// Standard JSON-RPC 2.0 errors.
var ErrCodeMessage = map[string]struct {
	code    int
	message string
}{
	// rpc server error
	ErrUnexpected:     {-1, "Unexpected error"},
	ErrAlreadyStarted: {-2, "RPC server is already started"},

	// rpc api error
	ErrRPCInvalidRequest: {-1001, "Invalid request"},
	ErrRPCMethodNotFound: {-1002, "Method not found"},
	ErrRPCInvalidParams:  {-1003, "Invalid parameters"},
	ErrRPCInternal:       {-1004, "Internal error"},
	ErrRPCParse:          {-1005, "Parse error"},
	ErrInvalidType:       {-1006, "Invalid type"},
}

// RpcErrorCode identifies a kind of error.  These error codes are NOT used for
// JSON-RPC response errors.
type RpcErrorCode int

// RPCError represents an error that is used as a part of a JSON-RPC Response
// object.
type RPCError struct {
	code    int    `json:"code,omitempty"`
	message string `json:"message,omitempty"`
	err     error  `json:"err"`
}

// Guarantee RPCError satisifies the builtin error interface.
var _, _ error = RPCError{}, (*RPCError)(nil)

// Error returns a string describing the RPC error.  This satisifies the
// builtin error interface.
func (e RPCError) Error() string {
	return fmt.Sprintf("%d: %s", e.code, e.message)
}

// NewRPCError constructs and returns a new JSON-RPC error that is suitable
// for use in a JSON-RPC Response object.
func NewRPCError(key string, err error) *RPCError {
	return &RPCError{
		code:    ErrCodeMessage[key].code,
		message: ErrCodeMessage[key].message,
		err:     err,
	}
}
