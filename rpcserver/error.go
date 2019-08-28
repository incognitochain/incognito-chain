package rpcserver

import (
	"fmt"

	"github.com/pkg/errors"
)

const (
	UnexpectedError = iota
	AlreadyStartedError
	RPCInvalidRequestError
	RPCMethodNotFoundError
	RPCInvalidParamsError
	RPCInvalidMethodPermissionError
	RPCInternalError
	RPCParseError
	InvalidTypeError
	AuthFailError
	InvalidSenderPrivateKeyError
	InvalidSenderViewingKeyError
	InvalidReceiverPaymentAddressError
	ListCustomTokenNotFoundError
	CanNotSignError
	GetOutputCoinError
	CreateTxDataError
	SendTxDataError
	TxTypeInvalidError
	RejectInvalidFeeError
	TxNotExistedInMemAndBLockError
	UnsubcribeError
	SubcribeError
	NetworkError
	TokenIsInvalidError
	GetClonedBeaconBestStateError
	GetClonedShardBestStateError
)

// Standard JSON-RPC 2.0 errors.
var ErrCodeMessage = map[int]struct {
	Code    int
	Message string
}{
	// general
	UnexpectedError:     {-1, "Unexpected error"},
	AlreadyStartedError: {-2, "RPC server is already started"},
	NetworkError:        {-3, "Network Error, failed to send request to RPC server"},

	// validate component -1xxx
	RPCInvalidRequestError:             {-1001, "Invalid request"},
	RPCMethodNotFoundError:             {-1002, "Method not found"},
	RPCInvalidParamsError:              {-1003, "Invalid parameters"},
	RPCInternalError:                   {-1004, "Internal error"},
	RPCParseError:                      {-1005, "Parse error"},
	InvalidTypeError:                   {-1006, "Invalid type"},
	AuthFailError:                      {-1007, "Auth failure"},
	RPCInvalidMethodPermissionError:    {-1008, "Invalid method permission"},
	InvalidReceiverPaymentAddressError: {-1009, "Invalid receiver paymentaddress"},
	ListCustomTokenNotFoundError:       {-1010, "Can not find any custom token"},
	CanNotSignError:                    {-1011, "Can not sign with key"},
	InvalidSenderPrivateKeyError:       {-1012, "Invalid sender's key"},
	GetOutputCoinError:                 {-1013, "Can not get output coin"},
	TxTypeInvalidError:                 {-1014, "Invalid tx type"},
	InvalidSenderViewingKeyError:       {-1015, "Invalid viewing key"},
	RejectInvalidFeeError:              {-1016, "Reject invalid fee"},
	TxNotExistedInMemAndBLockError:     {-1017, "Tx is not existed in mem and block"},
	TokenIsInvalidError:                {-1018, "Token is invalid"},
	GetClonedBeaconBestStateError:      {-1019, "Get Cloned Beacon Best State Error"},
	GetClonedShardBestStateError:       {-1020, "Get Cloned Shard Best State Error"},

	// processing -2xxx
	CreateTxDataError: {-2001, "Can not create tx"},
	SendTxDataError:   {-2002, "Can not send tx"},
	// socket/subcribe -3xxx
	SubcribeError:   {-3001, "Failed to subcribe"},
	UnsubcribeError: {-2002, "Failed to unsubcribe"},
}

// RPCError represents an error that is used as a part of a JSON-RPC JsonResponse
// object.
type RPCError struct {
	Code       int    `json:"Code,omitempty"`
	Message    string `json:"Message,omitempty"`
	err        error  `json:"Err"`
	StackTrace string `json:"StackTrace"`
}

func GetErrorCode(err int) int {
	return ErrCodeMessage[err].Code
}

// Guarantee RPCError satisifies the builtin error interface.
var _, _ error = RPCError{}, (*RPCError)(nil)

// Error returns a string describing the RPC error.  This satisifies the
// builtin error interface.
func (e RPCError) Error() string {
	return fmt.Sprintf("%d: %+v %+v", e.Code, e.err, e.StackTrace)
}

func (e RPCError) GetErr() error {
	return e.err
}

// NewRPCError constructs and returns a new JSON-RPC error that is suitable
// for use in a JSON-RPC JsonResponse object.
func NewRPCError(key int, err error, param ...interface{}) *RPCError {
	return &RPCError{
		Code:    ErrCodeMessage[key].Code,
		Message: fmt.Sprintf(ErrCodeMessage[key].Message, param),
		err:     errors.Wrap(err, ErrCodeMessage[key].Message),
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
	return NewRPCError(RPCInternalError, errors.New(errStr))
}
