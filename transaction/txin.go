package transaction

import "github.com/internet-cash/prototype/common"

const (
	MaxTxInSequenceNum int = 0xffffffff
)

// OutPoint defines a coin data type that is used to track previous
// transaction outputs.
type OutPoint struct {
	Hash common.Hash
	Vout int
}

type TxIn struct {
	PreviousOutPoint OutPoint
	SignatureScript  []byte
	Sequence         int
}

func (self TxIn) NewTxIn(prevOut *OutPoint, signatureScript []byte, witness [][]byte) *TxIn {
	self = TxIn{
		PreviousOutPoint: *prevOut,
		SignatureScript:  signatureScript,
		Sequence:         MaxTxInSequenceNum,
	}
	return &self
}
