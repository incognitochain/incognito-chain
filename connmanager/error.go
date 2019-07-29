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
)

var ErrCodeMessage = map[int]struct {
	code    int
	message string
}{
	UnexpectedError: {-1, "Unexpected error"},
	GetPeerIdError:  {-2, "Get peer id fail"},
	ConnectError:    {-3, "Connect error"},
	StartError:      {-4, "Start error"},
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
		Code:    ErrCodeMessage[key].code,
		Message: ErrCodeMessage[key].message,
		err:     errors.Wrap(err, ErrCodeMessage[key].message),
	}
}
