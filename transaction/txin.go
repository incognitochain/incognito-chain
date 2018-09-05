package transaction

import (
	"github.com/ninjadotorg/cash-prototype/common"
)

const (
	MaxTxInSequenceNum int = 0xffffffff

	// MaxPrevOutIndex is the maximum index the index field of a previous
	// outpoint can be.
	MaxPrevOutIndex uint32 = 0xffffffff
)

// OutPoint defines a coin data type that is used to track previous
// transaction outputs.
type OutPoint struct {
	Hash common.Hash
	Vout uint32
}

type TxIn struct {
	// PreviousOutPoint contains (hash and index of txout) in prev tx
	PreviousOutPoint OutPoint

	SignatureScript []byte
	Sequence        int
}

func (self TxIn) NewTxIn(prevOut *OutPoint, signatureScript []byte) *TxIn {
	self = TxIn{
		PreviousOutPoint: *prevOut,
		SignatureScript:  signatureScript,
		Sequence:         MaxTxInSequenceNum,
	}
	return &self
}
