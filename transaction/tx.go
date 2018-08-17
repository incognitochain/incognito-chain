package transaction

import (
	"github.com/internet-cash/prototype/common"
	"strconv"
	"fmt"
)

type Tx struct {
	Version  int
	TxIn     []TxIn
	TxOut    []TxOut
	LockTime int
	txHash   *common.Hash
}

func (self *Tx) AddTxIn(ti TxIn) {
	self.TxIn = append(self.TxIn, ti)
}

func (self *Tx) AddTxOut(to TxOut) {
	self.TxOut = append(self.TxOut, to)
}

func (self *Tx) Hash() (*common.Hash) {
	if self.txHash != nil {
		return self.txHash
	}
	record := strconv.Itoa(self.Version) + strconv.Itoa(self.Version)
	for _, txin := range self.TxIn {
		record += fmt.Sprint(txin.Sequence)
		record += string(txin.SignatureScript)
		record += fmt.Sprint(txin.PreviousOutPoint.Vout)
		record += fmt.Sprint(txin.PreviousOutPoint.Hash.String())
	}
	for _, txout := range self.TxOut {
		record += fmt.Sprint(txout.Value)
		record += string(txout.PkScript)
	}
	hash := common.DoubleHashH([]byte(record))
	self.txHash = &hash
	return self.txHash
}
