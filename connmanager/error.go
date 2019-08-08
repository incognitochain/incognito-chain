package connmanager

import (
	"fmt"

	"github.com/pkg/errors"
)

const (
	UnexpectedError = iota
	GetPeerIdError
	ConnectError
	StartError
	StopError
	NotAcceptConnectionError
)

var ErrCodeMessage = map[int]struct {
	Code    int
	Message string
}{
	UnexpectedError:          {-1, "Unexpected error"},
	GetPeerIdError:           {-2, "Get peer id fail"},
	ConnectError:             {-3, "Connect error"},
	StartError:               {-4, "Start error"},
	StopError:                {-5, "Stop errior"},
	NotAcceptConnectionError: {-6, "Not accept connection"},
}

type ConnManagerError struct {
	Code    int
	Message string
	err     error
}

func (e ConnManagerError) Error() string {
	return fmt.Sprintf("%d: %s %+v", e.Code, e.Message, e.err)
}

func NewConnManagerError(key int, err error) *ConnManagerError {
	return &ConnManagerError{
		Code:    ErrCodeMessage[key].Code,
		Message: ErrCodeMessage[key].Message,
		err:     errors.Wrap(err, ErrCodeMessage[key].Message),
	}
}
