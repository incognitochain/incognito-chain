// Copyright (c) 2014-2016 The thaibaoautonomous developers
// Use of this source Code is governed by an ISC
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
	MashallJsonError
	CanNotCheckDoubleSpendError
	HashError
	VersionError
	BlockHeightError
	DBError
	EpochError
	TimestampError
	InstructionHashError
	ShardStateHashError
	RandomError
	VerificationError
	ShardError
	BeaconError
	SignatureError
	CrossShardBlockError
	CandidateError
	ShardIDError
	ProducerError
	ShardStateError
	TransactionError
	InstructionError
	SwapError
	DuplicateBlockErr
)

var ErrCodeMessage = map[int]struct {
	Code    int
	message string
}{
	UnExpectedError:               {-1, "Unexpected error"},
	UpdateMerkleTreeForBlockError: {-2, "Update Merkle Commitments Tree For Block is failed"},
	UnmashallJsonBlockError:       {-3, "Unmarshall json block is failed"},
	CanNotCheckDoubleSpendError:   {-4, "CanNotCheckDoubleSpend Error"},
	HashError:                     {-5, "Hash error"},
	VersionError:                  {-6, "Version error"},
	BlockHeightError:              {-7, "Block height error"},
	DBError:                       {-8, "Database Error"},
	EpochError:                    {-9, "Epoch Error"},
	TimestampError:                {-10, "Timestamp Error"},
	InstructionHashError:          {-11, "Instruction Hash Error"},
	ShardStateHashError:           {-12, "ShardState Hash Error"},
	RandomError:                   {-13, "Random Number Error"},
	VerificationError:             {-14, "Verify Block Error"},
	BeaconError:                   {-15, "Beacon Error"},
	CrossShardBlockError:          {-17, "CrossShardBlockError"},
	SignatureError:                {-16, "Signature Error"},
	CandidateError:                {-17, "Candidate Error"},
	ShardIDError:                  {-18, "ShardID Error"},
	ProducerError:                 {-19, "Producer Error"},
	ShardStateError:               {-21, "Shard State Error"},
	TransactionError:              {-22, "Transaction invalid"},
	InstructionError:              {-23, "Instruction Error"},
	SwapError:                     {-24, "Swap Error"},
	MashallJsonError:              {-25, "MashallJson Error"},
	DuplicateBlockErr:             {-26, "Duplicate Block Error"},
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
		Code:    ErrCodeMessage[key].Code,
		Message: ErrCodeMessage[key].message,
		err:     errors.Wrap(err, ErrCodeMessage[key].message),
	}
}
