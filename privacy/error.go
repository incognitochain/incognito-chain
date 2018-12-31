package privacy

import (
	"fmt"
	"github.com/pkg/errors"
)

const (
	UnexpectedErr = iota
	InvalidOutputValue
	ProvingErr
	VerificationErr
	MarshalErr
	UnmarshalErr
	SetBytesProofErr
)

var ErrCodeMessage = map[int]struct {
	code    int
	message string
}{
	UnexpectedErr: {-1, "Unexpected error"},

	InvalidOutputValue: {-2, "Invalid output value"},
	ProvingErr : {-3, "Zero knowledge proving error"},
	VerificationErr : {-4, "Zero knowledge verification error"},
	MarshalErr : {-5, "Marshal payment proof error"},
	UnmarshalErr : {-6, "Unmarshal payment proof error"},
	SetBytesProofErr : {-6, "Set bytes payment proof error"},
}

type PrivacyError struct {
	code    int
	message string
	err     error
}

func (e PrivacyError) Error() string {
	return fmt.Sprintf("%+v: %+v %+v", e.code, e.message, e.err)
}

func NewPrivacyErr(key int, err error) *PrivacyError {
	return &PrivacyError{
		err:     errors.Wrap(err, ErrCodeMessage[key].message),
		code:    ErrCodeMessage[key].code,
		message: ErrCodeMessage[key].message,
	}
}
