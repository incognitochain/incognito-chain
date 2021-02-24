package common

import (
	"fmt"
	"github.com/pkg/errors"
)

const (
	UnexpectedError = iota

	// eth utils
	VerifyProofAndParseReceiptError
)

var ErrCodeMessage = map[int]struct {
	Code    int
	Message string
}{
	UnexpectedError: {-1, "Unexpected error"},

	// eth utils
	VerifyProofAndParseReceiptError: {-16001, "Verify proof and parse receipt eth error"},
}

type PortalCommonError struct {
	Code    int    // The code to send with reject messages
	Message string // Human readable message of the issue
	Err     error
}

// Error satisfies the error interface and prints human-readable errors.
func (e PortalCommonError) Error() string {
	return fmt.Sprintf("%d: %s %+v", e.Code, e.Message, e.Err)
}

func NewPortalCommonError(key int, err error, params ...interface{}) *PortalCommonError {
	return &PortalCommonError{
		Code:    ErrCodeMessage[key].Code,
		Message: fmt.Sprintf(ErrCodeMessage[key].Message, params),
		Err:     errors.Wrap(err, ErrCodeMessage[key].Message),
	}
}
