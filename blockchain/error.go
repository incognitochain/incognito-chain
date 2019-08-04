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
	DatabaseError
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
	DuplicateBlockError
	CommitteeOrValidatorError
	ShardBlockSanityError
	StoreIncomingCrossShardError
	DeleteIncomingCrossShardError
)

var ErrCodeMessage = map[int]struct {
	Code    int
	message string
}{
	UnExpectedError:               {-1000, "Unexpected error"},
	UpdateMerkleTreeForBlockError: {-1001, "Update Merkle Commitments Tree For Block is failed"},
	UnmashallJsonBlockError:       {-1002, "Unmarshall json block is failed"},
	CanNotCheckDoubleSpendError:   {-1003, "CanNotCheckDoubleSpend Error"},
	HashError:                     {-1004, "Hash error"},
	VersionError:                  {-1005, "Version error"},
	BlockHeightError:              {-1006, "Block height error"},
	DatabaseError:                 {-1007, "Database Error"},
	EpochError:                    {-1008, "Epoch Error"},
	TimestampError:                {-1009, "Timestamp Error"},
	InstructionHashError:          {-1010, "Instruction Hash Error"},
	ShardStateHashError:           {-1011, "ShardState Hash Error"},
	RandomError:                   {-1012, "Random Number Error"},
	VerificationError:             {-1013, "Verify Block Error"},
	BeaconError:                   {-1014, "Beacon Error"},
	CrossShardBlockError:          {-1015, "CrossShardBlockError"},
	SignatureError:                {-1016, "Signature Error"},
	CandidateError:                {-1017, "Candidate Error"},
	ShardIDError:                  {-1018, "ShardID Error"},
	ProducerError:                 {-1019, "Producer Error"},
	ShardStateError:               {-1020, "Shard State Error"},
	TransactionError:              {-1021, "Transaction invalid"},
	InstructionError:              {-1022, "Instruction Error"},
	SwapError:                     {-1023, "Swap Error"},
	MashallJsonError:              {-1024, "MashallJson Error"},
	DuplicateBlockError:           {-1025, "Duplicate Block Error"},
	CommitteeOrValidatorError:     {-1026, "Committee or Validator Error"},
	ShardBlockSanityError:         {-1027, "Shard Block Sanity Data Error"},
	StoreIncomingCrossShardError:  {-1028, "Store Incoming Cross Shard Block Error"},
	DeleteIncomingCrossShardError: {-1029, "Delete Incoming Cross Shard Block Error"},
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
