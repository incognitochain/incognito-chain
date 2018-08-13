package transaction

import "github.com/internet-cash/prototype/common"

const (
	// MaxTxInSequenceNum is the maximum sequence number the sequence field
	// of a transaction input can be.
	MaxTxInSequenceNum uint32 = 0xffffffff
)

// OutPoint defines a bitcoin data type that is used to track previous
// transaction outputs.
type OutPoint struct {
	Hash common.Hash
	Vout uint32
}

type TxIn struct {
	PreviousOutPoint OutPoint
	SignatureScript  []byte
	Sequence         uint32
}

func (self TxIn) NewTxIn(prevOut *OutPoint, signatureScript []byte, witness [][]byte) *TxIn {
	self = TxIn{
		PreviousOutPoint: *prevOut,
		SignatureScript:  signatureScript,
		Sequence:         MaxTxInSequenceNum,
	}
	return &self
}
