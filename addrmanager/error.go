package addrmanager

import (
	"fmt"

	"github.com/pkg/errors"
)

const (
	UnexpectedError = iota
	StopError
	CreateDataFileError
	EncodeDataFileError
	OpenDataFileError
	DecodeDataFileError
	WrongVersionError
)

var ErrCodeMessage = map[int]struct {
	Code    int
	Message string
}{
	UnexpectedError:     {-1, "Unexpected error"},
	StopError:           {-2, "Address manager is already in the process of shutting down"},
	CreateDataFileError: {-3, "Error opening file"},
	EncodeDataFileError: {-4, "Failed to encode file"},
	OpenDataFileError:   {-5, "Error opening file"},
	DecodeDataFileError: {-6, "Error to decode file"},
	WrongVersionError:   {-7, "Unknown Version in serialized addrmanager"},
}

type AddrManagerError struct {
	Code    int
	Message string
	err     error
}

func (e AddrManagerError) Error() string {
	return fmt.Sprintf("%d: %s %+v", e.Code, e.Message, e.err)
}

func NewAddrManagerError(key int, err error) *AddrManagerError {
	return &AddrManagerError{
		Code:    ErrCodeMessage[key].Code,
		Message: ErrCodeMessage[key].Message,
		err:     errors.Wrap(err, ErrCodeMessage[key].Message),
	}
}
