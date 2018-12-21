package addrmanager

import (
	"fmt"
	"github.com/pkg/errors"
)

const (
	UnexpectedError = iota
)

var ErrCodeMessage = map[int]struct {
	code    int
	message string
}{
	UnexpectedError: {-1, "Unexpected error"},
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
		Code:    ErrCodeMessage[key].code,
		Message: ErrCodeMessage[key].message,
		err:     errors.Wrap(err, ErrCodeMessage[key].message),
	}
}
