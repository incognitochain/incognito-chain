// Copyright (c) 2014-2016 The thaibaoautonomous developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package blockchain

import (
	"fmt"
	"github.com/pkg/errors"
)

const (
	UnExpectedError               = iota
	UpdateMerkleTreeForBlockError
	UnmashallJsonBlockError
	CanNotCheckDoubleSpendError
	NotSupportInLightMode
)

var ErrCodeMessage = map[int]struct {
	code    int
	message string
}{
	UnExpectedError:               {-1, "Unexpected error"},
	UpdateMerkleTreeForBlockError: {-2, "Update Merkle Commitments Tree For Block is failed"},
	UnmashallJsonBlockError:       {-3, "Unmarshall json block is failed"},
	CanNotCheckDoubleSpendError:   {-4, "Unmarshall json block is failed"},
	NotSupportInLightMode:         {-5, "This features is not supported in light mode running"},
}

type BlockChainError struct {
	Code    int
	Message string
	err     error
}

func (e BlockChainError) Error() string {
	return fmt.Sprintf("%d: %s \n %+v", e.Code, e.Message, e.err)
}

func NewBlockChainError(key int, err error) *BlockChainError {
	return &BlockChainError{
		Code:    ErrCodeMessage[key].code,
		Message: ErrCodeMessage[key].message,
		err:     errors.Wrap(err, ErrCodeMessage[key].message),
	}
}
