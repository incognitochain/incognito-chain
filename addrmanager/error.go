package addrmanager

import "fmt"

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
	Err     error
}

func (e AddrManagerError) Error() string {
	return fmt.Sprintf("%d: %s", e.Code, e.Message)
}

func NewAddrManagerError(key int, err error) *AddrManagerError {
	return &AddrManagerError{
		Code:    ErrCodeMessage[key].code,
		Message: ErrCodeMessage[key].message,
		Err:     err,
	}
}
