package rpcserver

import (
	"fmt"
	"github.com/pkg/errors"
)

const (
	ErrUnexpected = iota
	ErrAlreadyStarted
	ErrRPCInvalidRequest
	ErrRPCMethodNotFound
	ErrRPCInvalidParams
	ErrRPCInvalidMethodPermission
	ErrRPCInternal
	ErrRPCParse
	ErrInvalidType
	ErrAuthFail
	ErrInvalidSenderPrivateKey
	ErrInvalidSenderViewingKey
	ErrInvalidReceiverPaymentAddress
	ErrListCustomTokenNotFound
	ErrCanNotSign
	ErrGetOutputCoin
	ErrCreateTxData
	ErrSendTxData
	ErrTxTypeInvalid
	ErrRejectInvalidFee
	ErrTxNotExistedInMemAndBLock
)

// Standard JSON-RPC 2.0 errors.
var ErrCodeMessage = map[int]struct {
	code    int
	message string
}{
	// general
	ErrUnexpected:     {-1, "Unexpected error"},
	ErrAlreadyStarted: {-2, "RPC server is already started"},

	// validate component -1xxx
	ErrRPCInvalidRequest:             {-1001, "Invalid request"},
	ErrRPCMethodNotFound:             {-1002, "Method not found"},
	ErrRPCInvalidParams:              {-1003, "Invalid parameters"},
	ErrRPCInternal:                   {-1004, "Internal error"},
	ErrRPCParse:                      {-1005, "Parse error"},
	ErrInvalidType:                   {-1006, "Invalid type"},
	ErrAuthFail:                      {-1007, "Auth failure"},
	ErrRPCInvalidMethodPermission:    {-1008, "Invalid method permission"},
	ErrInvalidReceiverPaymentAddress: {-1009, "Invalid receiver paymentaddress"},
	ErrListCustomTokenNotFound:       {-1010, "Can not find any custom token"},
	ErrCanNotSign:                    {-1011, "Can not sign with key"},
	ErrInvalidSenderPrivateKey:       {-1012, "Invalid sender's key"},
	ErrGetOutputCoin:                 {-1013, "Can not get output coin"},
	ErrTxTypeInvalid:                 {-1014, "Invalid tx type"},
	ErrInvalidSenderViewingKey:       {-1015, "Invalid viewing key"},
	ErrRejectInvalidFee:              {-1016, "Reject invalid fee"},
	ErrTxNotExistedInMemAndBLock:     {-1017, "Tx is not existed in mem and block"},

	// processing -2xxx
	ErrCreateTxData: {-2001, "Can not create tx"},
	ErrSendTxData:   {-2002, "Can not send tx"},
}

// RPCError represents an error that is used as a part of a JSON-RPC JsonResponse
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
// for use in a JSON-RPC JsonResponse object.
func NewRPCError(key int, err error) *RPCError {
	return &RPCError{
		Code:    ErrCodeMessage[key].code,
		Message: ErrCodeMessage[key].message,
		err:     errors.Wrap(err, ErrCodeMessage[key].message),
	}
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
