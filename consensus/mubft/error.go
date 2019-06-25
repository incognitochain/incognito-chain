package mubft

import (
	"fmt"
	"github.com/pkg/errors"
)

const (
	ErrUnexpected = iota
	ErrBlockSizeExceed
	ErrNotInCommittee
	ErrSigWrongOrNotExits
	ErrChainNotFullySynced
	ErrTxIsWrong
	ErrNotEnoughChainData
	ErrExceedSigWaitTime
	ErrMerkleRootCommitments
	ErrNotEnoughSigs
	ErrExceedBlockRetry
)

var ErrCodeMessage = map[int]struct {
	code    int
	message string
}{
	ErrUnexpected:            {-1, "Unexpected error"},
	ErrBlockSizeExceed:       {-2, "block size is too big"},
	ErrNotInCommittee:        {-3, "user not in committee"},
	ErrSigWrongOrNotExits:    {-4, "signature is wrong or not existed in block header"},
	ErrChainNotFullySynced:   {-5, "chains are not fully synced"},
	ErrTxIsWrong:             {-6, "transaction is wrong"},
	ErrNotEnoughChainData:    {-7, "not enough chain data"},
	ErrExceedSigWaitTime:     {-8, "exceed blocksig wait time"},
	ErrMerkleRootCommitments: {-9, "MerkleRootCommitments is wrong"},
	ErrNotEnoughSigs:         {-10, "not enough signatures"},
	ErrExceedBlockRetry:      {-11, "exceed block retry"},
}

type ConsensusError struct {
	Code    int
	Message string
	Err     error
}

func (e ConsensusError) Error() string {
	return fmt.Sprintf("%d: %s %+v", e.Code, e.Message, e.Err)
}

func NewConsensusError(key int, err error) *ConsensusError {
	return &ConsensusError{
		Code:    ErrCodeMessage[key].code,
		Message: ErrCodeMessage[key].message,
		Err:     errors.Wrap(err, ErrCodeMessage[key].message),
	}
}
