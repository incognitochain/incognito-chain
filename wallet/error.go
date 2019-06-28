package wallet

import (
	"fmt"
	"github.com/pkg/errors"
)

const (
	InvalidChecksumErr    = iota
	WrongPassphraseErr
	ExistedAccountErr
	ExistedAccountNameErr
	UnexpectedErr
	EmptyWalletNameErr
	NotFoundAccountErr
)

var ErrCodeMessage = map[int]struct {
	code    int
	message string
}{
	UnexpectedErr: {-1, "Unexpected error"},

	InvalidChecksumErr:    {-1000, "Checksum does not match"},
	WrongPassphraseErr:    {-1001, "Wrong passphrase"},
	ExistedAccountErr:     {-1002, "Existed account"},
	ExistedAccountNameErr: {-1002, "Existed account name"},
	EmptyWalletNameErr: {-1003, "Wallet name is empty"},
	NotFoundAccountErr: {-1004, "Account wallet is not found"},
}

type WalletError struct {
	code    int
	message string
	err     error
}

func (e WalletError) Error() string {
	return fmt.Sprintf("%+v: %+v", e.code, e.message)
}

func NewWalletError(key int, err error) *WalletError {
	return &WalletError{
		err:     errors.Wrap(err, ErrCodeMessage[key].message),
		code:    ErrCodeMessage[key].code,
		message: ErrCodeMessage[key].message,
	}
}
