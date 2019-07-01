// Copyright (c) 2014-2016 The thaibaoautonomous developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package btc

import (
	"fmt"

	"github.com/pkg/errors"
)

const (
	UnExpectedError = iota
	APIError
	UnmashallJsonBlockError
	TimestampError
	NonceError
	WrongTypeError
	TimeParseError
)

var ErrCodeMessage = map[int]struct {
	code    int
	message string
}{
	UnExpectedError:         {-1, "Unexpected error"},
	APIError:                {-2, "API Error"},
	TimestampError:          {-3, "Timestamp Error"},
	UnmashallJsonBlockError: {-4, "Unmarshall json block is failed"},
	NonceError:              {-5, "Nonce Error"},
	WrongTypeError:              {-6, "Wrong Type Error"},
	TimeParseError:              {-7, "Time Parse Error"},
}

type BTCAPIError struct {
	Code    int
	Message string
	err     error
}

func (e BTCAPIError) Error() string {
	return fmt.Sprintf("%d: %s \n %+v", e.Code, e.Message, e.err)
}

func NewBTCAPIError(key int, err error) *BTCAPIError {
	return &BTCAPIError{
		Code:    ErrCodeMessage[key].code,
		Message: ErrCodeMessage[key].message,
		err:     errors.Wrap(err, ErrCodeMessage[key].message),
	}
}
