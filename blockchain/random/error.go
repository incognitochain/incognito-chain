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
	BlockHashParseError
	GetBlockHashResultError
	GetBlockHeaderResultError
	ParseNonceResultError
	ParseTimestampResultError
)

var ErrCodeMessage = map[int]struct {
	code    int
	message string
}{
	UnExpectedError:           {-1, "Unexpected error"},
	APIError:                  {-2, "API Error"},
	TimestampError:            {-3, "Timestamp Error"},
	UnmashallJsonBlockError:   {-4, "Unmarshall json block is failed"},
	NonceError:                {-5, "Nonce Error"},
	WrongTypeError:            {-6, "Wrong Type Error"},
	TimeParseError:            {-7, "Time Parse Error"},
	BlockHashParseError:       {-8, "Block Hash Parse Error"},
	GetBlockHashResultError:   {-9, "Get Block Hash Result Error"},
	GetBlockHeaderResultError: {-10, "Get Block Header Result Error"},
	ParseNonceResultError:     {-11, "Parse Nonce Result Error"},
	ParseTimestampResultError: {-12, "Parse Timestamp Result Error"},
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
