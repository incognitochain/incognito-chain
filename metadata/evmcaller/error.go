package evmcaller

import (
	"fmt"

	"github.com/pkg/errors"
)

const (
	UnexpectedError = iota

	GetEVMHeaderByHashError
	GetEVMHeaderByHeightError
	GetEVMBlockHeightError
	GetEVMHeaderResultFromDBError
)

var ErrCodeMessage = map[int]struct {
	Code    int
	Message string
}{
	UnexpectedError: {-16000, "Unexpected error"},

	GetEVMHeaderByHashError:       {-16001, "Get evm header by block hash error"},
	GetEVMHeaderByHeightError:     {-16002, "Get evm header by block height error"},
	GetEVMBlockHeightError:        {-16003, "Get latest evm block height error"},
	GetEVMHeaderResultFromDBError: {-16004, "Get evm header result from DB error"},
}

type EVMCallerError struct {
	Code    int    // The code to send with reject messages
	Message string // Human readable message of the issue
	Err     error
}

// Error satisfies the error interface and prints human-readable errors.
func (e EVMCallerError) Error() string {
	return fmt.Sprintf("%d: %s %+v", e.Code, e.Message, e.Err)
}

func NewEVMCallerError(key int, err error, params ...interface{}) *EVMCallerError {
	return &EVMCallerError{
		Code:    ErrCodeMessage[key].Code,
		Message: fmt.Sprintf(ErrCodeMessage[key].Message, params),
		Err:     errors.Wrap(err, ErrCodeMessage[key].Message),
	}
}
