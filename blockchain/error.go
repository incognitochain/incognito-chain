// Copyright (c) 2014-2016 The thaibaoautonomous developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package blockchain

import (
	"fmt"

	"github.com/pkg/errors"
)

const (
	UnExpectedError = iota
	UpdateMerkleTreeForBlockError
	UnmashallJsonBlockError
	CanNotCheckDoubleSpendError
	HashError
	VersionError
	BlockHeightError
	DBError
	EpochError
	TimestampError
	InstructionHashError
	ShardStateHashError
)

var ErrCodeMessage = map[int]struct {
	code    int
	message string
}{
	UnExpectedError:               {-1, "Unexpected error"},
	UpdateMerkleTreeForBlockError: {-2, "Update Merkle Commitments Tree For Block is failed"},
	UnmashallJsonBlockError:       {-3, "Unmarshall json block is failed"},
	CanNotCheckDoubleSpendError:   {-4, "Unmarshall json block is failed"},
	HashError:                     {-5, "Hash error"},
	VersionError:                  {-6, "Version error"},
	BlockHeightError:              {-7, "Block height error"},
	DBError:                       {-8, "Database Error"},
	EpochError:                    {-9, "Epoch Error"},
	TimestampError:                {-10, "Timestamp Error"},
	InstructionHashError:          {-11, "Instruction Hash Error"},
	ShardStateHashError:           {-12, "ShardState Hash Error"},
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
