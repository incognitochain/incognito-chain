package transaction

import (
	"github.com/internet-cash/prototype/common"
)

const (
	MaxTxInSequenceNum int = 0xffffffff
)

// OutPoint defines a coin data type that is used to track previous
// transaction outputs.
type OutPoint struct {
	Hash common.Hash
	Vout int
}

//func (self OutPoint) MarshalJSON() ([]byte, error) {
//	result, _ := json.Marshal(&struct {
//		Hash string
//		*OutPoint
//	}{Hash: self.Hash.String(), OutPoint: &self})
//	return result, nil
//}
//
//func (self OutPoint) UnmarshalJSON(data []byte) error {
//	json.Unmarshal(data, &struct {
//		*OutPoint
//	}{
//		OutPoint: &self,
//	})
//	return nil
//}

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
