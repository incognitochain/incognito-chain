package bnb

import (
	"fmt"
	"github.com/pkg/errors"
)

const (
	UnexpectedErr = iota
	InvalidBasicSignedHeaderErr
	InvalidSignatureSignedHeaderErr
	InvalidNewHeaderErr
	InvalidBasicHeaderErr
	InvalidTxProofErr
	ParseProofErr
	ExistedNewHeaderErr
	GetBNBDataHashErr

	StoreBNBChainErr
	GetBNBChainErr
	FullOrphanBlockErr
	AddBlockToOrphanBlockErr
	CheckOrphanBlockErr
)

var ErrCodeMessage = map[int]struct {
	Code    int
	Message string
}{
	UnexpectedErr: {-14000, "Unexpected error"},

	InvalidBasicSignedHeaderErr:     {-14001, "Invalid basic signed header error"},
	InvalidSignatureSignedHeaderErr: {-14002, "Invalid signature signed header error"},
	InvalidNewHeaderErr:             {-14003, "Invalid new header"},
	InvalidBasicHeaderErr:           {-14004, "Invalid basic header error"},
	InvalidTxProofErr:               {-14005, "Invalid tx proof error"},
	ParseProofErr:                   {-14006, "Parse proof from json string error"},
	ExistedNewHeaderErr:             {-14007, "New header is existed in list of unconfirmed headers error"},
	GetBNBDataHashErr:               {-14008, "Can not get bnb data hash from db error"},

	StoreBNBChainErr:         {-14009, "Store bnb chain to lvdb error"},
	GetBNBChainErr:           {-14010, "Get latest block to lvdb error"},
	FullOrphanBlockErr:       {-14011, "Full orphan blocks error"},
	AddBlockToOrphanBlockErr: {-14012, "Add block to orphan blocks error"},
	CheckOrphanBlockErr:      {-14013, "Check orphan blocks error"},
}

type BNBRelayingError struct {
	Code    int
	Message string
	err     error
}

func (e BNBRelayingError) Error() string {
	return fmt.Sprintf("%+v: %+v %+v", e.Code, e.Message, e.err)
}

func (e BNBRelayingError) GetCode() int {
	return e.Code
}

func NewBNBRelayingError(key int, err error) *BNBRelayingError {
	return &BNBRelayingError{
		err:     errors.Wrap(err, ErrCodeMessage[key].Message),
		Code:    ErrCodeMessage[key].Code,
		Message: ErrCodeMessage[key].Message,
	}
}
