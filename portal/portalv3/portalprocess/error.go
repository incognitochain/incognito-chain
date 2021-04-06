package portalprocess

import (
	"fmt"
	"github.com/pkg/errors"
)

const (
	UnexpectedError = iota
	HandlePortingRequestError
)

var ErrCodeMessage = map[int]struct {
	Code    int
	Message string
}{
	UnexpectedError: {-1, "Unexpected error"},
	HandlePortingRequestError:          {-13001, "Handle porting request error"},
}

type PortalProcessError struct {
	Code    int    // The code to send with reject messages
	Message string // Human readable message of the issue
	Err     error
}

// Error satisfies the error interface and prints human-readable errors.
func (e PortalProcessError) Error() string {
	return fmt.Sprintf("%d: %s %+v", e.Code, e.Message, e.Err)
}

func NewPortalProcessError(key int, err error, params ...interface{}) *PortalProcessError {
	return &PortalProcessError{
		Code:    ErrCodeMessage[key].Code,
		Message: fmt.Sprintf(ErrCodeMessage[key].Message, params),
		Err:     errors.Wrap(err, ErrCodeMessage[key].Message),
	}
}