package netsync

import (
	"fmt"
	"github.com/pkg/errors"
)

const (
	UnexpectedError = iota
	AlreadyStartError
	AlreadyShutdownError
)

var ErrCodeMessage = map[int]struct {
	Code    int
	Message string
}{
	UnexpectedError:      {-1, "Unexpected error"},
	AlreadyStartError:    {-2, "Already started"},
	AlreadyShutdownError: {-3, "We're shutting down"},
}

type NetSyncError struct {
	Code    int
	Message string
	err     error
}

func (e NetSyncError) Error() string {
	return fmt.Sprintf("%d: %s %+v", e.Code, e.Message, e.err)
}

func NewNetSyncError(key int, err error, params ...interface{}) *NetSyncError {
	return &NetSyncError{
		Code:    ErrCodeMessage[key].Code,
		Message: fmt.Sprintf(ErrCodeMessage[key].Message, params),
		err:     errors.Wrap(err, ErrCodeMessage[key].Message),
	}
}
